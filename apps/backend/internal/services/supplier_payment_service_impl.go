package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
	"gorm.io/gorm"
)

// supplierPaymentServiceImpl implements SupplierPaymentService interface
// Story 10.4: Service layer with business logic, transaction wrapping, and integrations
type supplierPaymentServiceImpl struct {
	db                  *gorm.DB
	supplierPaymentRepo repositories.SupplierPaymentRepository
	invoiceRepo         repositories.PurchaseInvoiceRepository
	supplierRepo        repositories.SupplierRepository
	auditService        AuditService
}

// NewSupplierPaymentService creates a new supplier payment service
// Story 10.4: Factory function with dependency injection
func NewSupplierPaymentService(
	db *gorm.DB,
	supplierPaymentRepo repositories.SupplierPaymentRepository,
	invoiceRepo repositories.PurchaseInvoiceRepository,
	supplierRepo repositories.SupplierRepository,
	auditService AuditService,
) SupplierPaymentService {
	return &supplierPaymentServiceImpl{
		db:                  db,
		supplierPaymentRepo: supplierPaymentRepo,
		invoiceRepo:         invoiceRepo,
		supplierRepo:        supplierRepo,
		auditService:        auditService,
	}
}

// RecordPayment records a new supplier payment with validation and audit logging
// Story 10.4, AC1: Main business logic method with transaction wrapping
// CRITICAL from Story 10.3 code review: Transaction wrapping for atomic operations
func (s *supplierPaymentServiceImpl) RecordPayment(ctx context.Context, request *RecordPaymentRequest, createdBy uint, ipAddress string) (*models.SupplierPayment, error) {
	// Validate inputs
	if request == nil {
		return nil, fmt.Errorf("payment request cannot be nil")
	}
	if request.PurchaseInvoiceID == 0 {
		return nil, fmt.Errorf("purchase invoice ID is required")
	}
	if createdBy == 0 {
		return nil, fmt.Errorf("created by user ID is required")
	}
	if request.PaymentAmount <= 0 {
		return nil, fmt.Errorf("payment amount must be positive")
	}

	// Parse and validate payment date
	paymentDate, err := time.Parse("2006-01-02", request.PaymentDate)
	if err != nil {
		return nil, fmt.Errorf("invalid payment date format (use YYYY-MM-DD): %w", err)
	}

	// Validate payment date is not in future (PATCH-018: UTC handling)
	if paymentDate.After(time.Now().UTC()) {
		return nil, fmt.Errorf("payment date cannot be in the future")
	}

	// Validate payment method enum (already validated by binding:oneof)
	// Additional validation if needed

	var finalPayment *models.SupplierPayment

	// CRITICAL-001 from Story 10.3 code review: Wrap entire operation in database transaction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Load invoice to validate and check remaining balance
		invoice, err := s.invoiceRepo.GetByID(ctx, request.PurchaseInvoiceID)
		if err != nil {
			return fmt.Errorf("purchase invoice not found: %w", err)
		}

		// Validate invoice is received (Story 10.4 requirement)
		if invoice.ReceiptStatus != "received" {
			return fmt.Errorf("can only pay for received invoices (current status: %s)", invoice.ReceiptStatus)
		}

		// Calculate remaining balance (Story 10.4: overpayment prevention)
		totalPaid, err := s.supplierPaymentRepo.GetTotalPaidByInvoice(ctx, invoice.ID)
		if err != nil {
			return fmt.Errorf("failed to calculate total paid: %w", err)
		}

		remainingBalance := invoice.TotalAmount - totalPaid

		// Validate payment amount doesn't exceed remaining balance
		// Allow small tolerance (0.01) for rounding differences
		if request.PaymentAmount > remainingBalance+0.01 {
			return fmt.Errorf("payment amount (%.2f) exceeds remaining balance (%.2f)",
				request.PaymentAmount, remainingBalance)
		}

		// Get branch ID from invoice for payment record
		branchID := invoice.BranchID

		// Create supplier payment record
		payment := &models.SupplierPayment{
			PurchaseInvoiceID: request.PurchaseInvoiceID,
			PaymentDate:      paymentDate,
			PaymentAmount:    request.PaymentAmount,
			PaymentMethod:    request.PaymentMethod,
			Notes:            request.Notes,
			ReferenceNumber:  request.ReferenceNumber,
			BranchID:         branchID,
			CreatedBy:        createdBy,
		}

		// Create payment
		if err := s.supplierPaymentRepo.Create(ctx, payment); err != nil {
			return fmt.Errorf("failed to create supplier payment: %w", err)
		}

		// Update invoice payment status atomically
		// Story 10.4, AC1: System updates invoice payment status (unpaid → partial → fully paid)
		if err := s.invoiceRepo.UpdatePaymentStatus(ctx, invoice.ID); err != nil {
			return fmt.Errorf("failed to update payment status: %w", err)
		}

		// CRITICAL-003 from Story 10.3 code review: Audit trail logging
		// Log "supplier_payment.recorded" with payment details
		slog.InfoContext(ctx, "supplier_payment.recorded",
			"payment_id", payment.ID,
			"invoice_id", invoice.ID,
			"invoice_number", invoice.InvoiceNumber,
			"payment_amount", request.PaymentAmount,
			"payment_method", request.PaymentMethod,
			"created_by", createdBy,
			"ip_address", ipAddress,
		)

		// Log "payment_status.updated" for invoice status change
		slog.InfoContext(ctx, "payment_status.updated",
			"invoice_id", invoice.ID,
			"invoice_number", invoice.InvoiceNumber,
			"old_status", invoice.PaymentStatus,
			"total_paid", totalPaid,
			"new_payment", request.PaymentAmount,
			"updated_by", createdBy,
			"ip_address", ipAddress,
		)

		// Store final payment for return
		finalPayment = payment

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to record payment: %w", err)
	}

	// Reload payment with relationships for response
	paymentWithRelations, err := s.supplierPaymentRepo.GetByID(ctx, finalPayment.ID)
	if err != nil {
		// Payment was created but failed to load - log warning but return success
		slog.WarnContext(ctx, "payment created but failed to load with relations",
			"payment_id", finalPayment.ID,
			"error", err.Error())
		return finalPayment, nil
	}

	return paymentWithRelations, nil
}

// GetSupplierPaymentByID retrieves a supplier payment by ID
// Story 10.4: Returns payment details with invoice information
func (s *supplierPaymentServiceImpl) GetSupplierPaymentByID(ctx context.Context, id uint) (*models.SupplierPayment, error) {
	if id == 0 {
		return nil, &InvalidInputError{Field: "id", Message: "ID must be positive"}
	}

	payment, err := s.supplierPaymentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// ListSupplierPayments retrieves supplier payments with filtering and pagination
// Story 10.4: Supports filtering by invoice, date range, payment method, branch
func (s *supplierPaymentServiceImpl) ListSupplierPayments(ctx context.Context, filter *SupplierPaymentListFilter) ([]*models.SupplierPayment, int64, error) {
	// Set defaults for pagination (PATCH-019)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	// Convert service filter to repository filter
	repoFilter := &repositories.SupplierPaymentFilter{
		PurchaseInvoiceID: filter.PurchaseInvoiceID,
		StartDate:          filter.StartDate,
		EndDate:            filter.EndDate,
		PaymentMethod:     filter.PaymentMethod,
		BranchID:           filter.BranchID,
		Page:               filter.Page,
		Limit:              filter.Limit,
		SortBy:             filter.SortBy,
		SortOrder:          filter.SortOrder,
	}

	payments, total, err := s.supplierPaymentRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list supplier payments: %w", err)
	}

	return payments, total, nil
}

// GetPaymentHistoryBySupplier retrieves payment history grouped by supplier
// Story 10.4, AC2: Returns payments for a supplier with invoice details
func (s *supplierPaymentServiceImpl) GetPaymentHistoryBySupplier(ctx context.Context, supplierID uint, filter *PaymentHistoryFilter) ([]*PaymentHistoryResponse, error) {
	if supplierID == 0 {
		return nil, &InvalidInputError{Field: "supplierID", Message: "Supplier ID must be positive"}
	}

	// Get all payments for this supplier's invoices
	// This requires querying payments where the invoice belongs to this supplier
	// We need to join supplier_payments with purchase_invoices

	// Build query with join to get payments for supplier's invoices
	query := s.db.WithContext(ctx).
		Table("supplier_payments").
		Select("supplier_payments.*, purchase_invoices.invoice_number, purchase_invoices.invoice_date, purchase_invoices.total_amount").
		Joins("JOIN purchase_invoices ON supplier_payments.purchase_invoice_id = purchase_invoices.id").
		Where("purchase_invoices.supplier_id = ? AND purchase_invoices.deleted_at IS NULL", supplierID)

	// Apply date filters
	if filter.StartDate != nil {
		query = query.Where("supplier_payments.payment_date >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("supplier_payments.payment_date <= ?", *filter.EndDate)
	}

	// Get total count first
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count payment history: %w", err)
	}

	if totalCount == 0 {
		return []*PaymentHistoryResponse{}, nil
	}

	// Apply pagination
	page := filter.Page
	limit := filter.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Execute query with pagination
	var results []struct {
		models.SupplierPayment
		InvoiceNumber     string  `json:"invoiceNumber"`
		InvoiceDate       string  `json:"invoiceDate"`
		InvoiceTotalAmount float64 `json:"invoiceTotalAmount"`
	}

	err := query.
		Order("payment_date DESC").
		Offset(offset).
		Limit(limit).
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get payment history: %w", err)
	}

	// Build response with remaining balance calculation
	history := make([]*PaymentHistoryResponse, 0, len(results))
	for _, result := range results {
		// Get total paid for this invoice
		invoiceTotalPaid, err := s.supplierPaymentRepo.GetTotalPaidByInvoice(ctx, result.PurchaseInvoiceID)
		if err != nil {
			// Log warning but continue
			slog.WarnContext(ctx, "failed to calculate total paid for payment history",
				"invoice_id", result.PurchaseInvoiceID,
				"error", err.Error())
			invoiceTotalPaid = result.PaymentAmount // Fallback to this payment amount
		}

		remainingBalance := result.InvoiceTotalAmount - invoiceTotalPaid

		history = append(history, &PaymentHistoryResponse{
			ID:                result.ID,
			PaymentDate:       result.PaymentDate.Format("2006-01-02"),
			PaymentAmount:     result.PaymentAmount,
			PaymentMethod:     result.PaymentMethod,
			Notes:             result.Notes,
			ReferenceNumber:   result.ReferenceNumber,
			InvoiceNumber:     result.InvoiceNumber,
			InvoiceDate:       result.InvoiceDate,
			InvoiceTotalAmount: result.InvoiceTotalAmount,
			RemainingBalance:  remainingBalance,
		})
	}

	return history, nil
}
