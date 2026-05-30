package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// supplierPaymentRepository implements SupplierPaymentRepository interface
// Story 10.4: GORM-based concrete implementation
type supplierPaymentRepository struct {
	db *gorm.DB
}

// NewSupplierPaymentRepository creates a new supplier payment repository
// Story 10.4: Factory function for dependency injection
func NewSupplierPaymentRepository(db interface{}) SupplierPaymentRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &supplierPaymentRepository{db: gormDB}
}

// Create inserts a new supplier payment into the database
// Story 10.4, AC1: Record payment made to supplier for a purchase invoice
func (r *supplierPaymentRepository) Create(ctx context.Context, payment *models.SupplierPayment) error {
	if payment == nil {
		return fmt.Errorf("supplier payment cannot be nil")
	}
	if payment.PurchaseInvoiceID == 0 {
		return fmt.Errorf("purchase invoice ID is required")
	}
	if payment.CreatedBy == 0 {
		return fmt.Errorf("createdBy user ID is required")
	}
	if payment.BranchID == 0 {
		return fmt.Errorf("branch ID is required")
	}

	// Validate payment date is not in future (PATCH-018: UTC handling)
	if payment.PaymentDate.After(time.Now().UTC()) {
		return fmt.Errorf("payment date cannot be in the future")
	}

	// Validate payment amount is positive (already handled by binding:gt=0, but double check)
	if payment.PaymentAmount <= 0 {
		return fmt.Errorf("payment amount must be positive")
	}

	// Validate payment method enum value (PATCH-008)
	validMethods := map[string]bool{
		"cash":    true,
		"transfer": true,
		"e-wallet": true,
		"check":    true,
		"other":    true,
	}
	if !validMethods[payment.PaymentMethod] {
		return fmt.Errorf("invalid payment method: %s", payment.PaymentMethod)
	}

	// Set timestamps if not provided
	if payment.CreatedAt.IsZero() {
		payment.CreatedAt = time.Now().UTC()
	}
	if payment.UpdatedAt.IsZero() {
		payment.UpdatedAt = time.Now().UTC()
	}

	// Validate purchase invoice exists and is not deleted
	var invoice models.PurchaseInvoice
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", payment.PurchaseInvoiceID).
		First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("purchase invoice not found")
		}
		return fmt.Errorf("failed to validate purchase invoice: %w", err)
	}

	// Validate invoice is in "received" status (Story 10.4 requirement)
	if invoice.ReceiptStatus != "received" {
		return fmt.Errorf("can only pay for received invoices (current status: %s)", invoice.ReceiptStatus)
	}

	// Validate branch exists
	var branch models.Branch
	if err := r.db.WithContext(ctx).First(&branch, payment.BranchID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("branch not found")
		}
		return fmt.Errorf("failed to validate branch: %w", err)
	}

	// Create the supplier payment
	if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
		return fmt.Errorf("failed to create supplier payment: %w", err)
	}

	return nil
}

// GetByID retrieves a supplier payment by its ID with eager loaded relationships
// Story 10.4: Get payment details for viewing
func (r *supplierPaymentRepository) GetByID(ctx context.Context, id uint) (*models.SupplierPayment, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}

	var payment models.SupplierPayment
	err := r.db.WithContext(ctx).
		Preload("PurchaseInvoice.Supplier").
		Preload("PurchaseInvoice.Branch").
		Preload("Branch").
		First(&payment, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get supplier payment: %w", err)
	}

	return &payment, nil
}

// GetByInvoiceID retrieves all supplier payments for a specific purchase invoice
// Story 10.4: Get payment history for a specific invoice
func (r *supplierPaymentRepository) GetByInvoiceID(ctx context.Context, invoiceID uint) ([]*models.SupplierPayment, error) {
	if invoiceID == 0 {
		return nil, ErrInvalidInput
	}

	var payments []*models.SupplierPayment
	err := r.db.WithContext(ctx).
		Where("purchase_invoice_id = ?", invoiceID).
		Order("payment_date ASC, created_at ASC").
		Find(&payments).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get supplier payments by invoice: %w", err)
	}

	// Return empty slice if no payments found (not an error)
	if payments == nil {
		payments = []*models.SupplierPayment{}
	}

	return payments, nil
}

// List retrieves supplier payments with optional filtering and pagination
// Story 10.4: List payments for payment history views
func (r *supplierPaymentRepository) List(ctx context.Context, filter *SupplierPaymentFilter) ([]*models.SupplierPayment, int64, error) {
	// Set defaults for pagination (PATCH-019)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	// Build query with filters
	query := r.db.WithContext(ctx).Model(&models.SupplierPayment{})

	// Apply filters
	if filter.PurchaseInvoiceID != nil {
		query = query.Where("purchase_invoice_id = ?", *filter.PurchaseInvoiceID)
	}

	if filter.StartDate != nil {
		query = query.Where("payment_date >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("payment_date <= ?", *filter.EndDate)
	}

	if filter.PaymentMethod != nil {
		query = query.Where("payment_method = ?", *filter.PaymentMethod)
	}

	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}

	// Validate date range (PATCH-005)
	if filter.StartDate != nil && filter.EndDate != nil {
		startDate, err := time.Parse("2006-01-02", *filter.StartDate)
		endDate, err2 := time.Parse("2006-01-02", *filter.EndDate)
		if err == nil && err2 == nil && startDate.After(endDate) {
			return nil, 0, fmt.Errorf("start date cannot be after end date")
		}
	}

	// Get total count for pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count supplier payments: %w", err)
	}

	// Handle empty results (PATCH-019)
	if total == 0 {
		emptyPayments := []*models.SupplierPayment{}
		return emptyPayments, 0, nil
	}

	// Apply sorting (PATCH-001: SQL injection prevention)
	// Use whitelist for sort_by field
	validSortFields := map[string]bool{
		"payment_date":   true,
		"payment_amount": true,
		"created_at":     true,
	}

	sortField := "payment_date" // default
	if filter.SortBy != "" && validSortFields[filter.SortBy] {
		sortField = filter.SortBy
	}

	// Use whitelist for sort_order
	validSortOrders := map[string]bool{
		"asc":  true,
		"desc": true,
	}

	sortOrder := "desc" // default
	if filter.SortOrder != "" && validSortOrders[filter.SortOrder] {
		sortOrder = filter.SortOrder
	}

	// Apply sorting with constant strings (no interpolation)
	if sortField == "payment_date" {
		if sortOrder == "asc" {
			query = query.Order("payment_date ASC")
		} else {
			query = query.Order("payment_date DESC")
		}
	} else if sortField == "payment_amount" {
		if sortOrder == "asc" {
			query = query.Order("payment_amount ASC")
		} else {
			query = query.Order("payment_amount DESC")
		}
	} else {
		// created_at
		if sortOrder == "asc" {
			query = query.Order("created_at ASC")
		} else {
			query = query.Order("created_at DESC")
		}
	}

	// Apply pagination with overflow check (PATCH-012)
	offset := (filter.Page - 1) * filter.Limit
	if offset < 0 {
		offset = 0
	}

	var payments []*models.SupplierPayment
	err := query.
		Offset(offset).
		Limit(filter.Limit).
		Find(&payments).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list supplier payments: %w", err)
	}

	// Handle empty slice
	if payments == nil {
		payments = []*models.SupplierPayment{}
	}

	return payments, total, nil
}

// GetTotalPaidByInvoice calculates the total amount paid for a specific purchase invoice
// Story 10.4: Used for payment status calculation (unpaid/partial/fully paid)
func (r *supplierPaymentRepository) GetTotalPaidByInvoice(ctx context.Context, invoiceID uint) (float64, error) {
	if invoiceID == 0 {
		return 0, ErrInvalidInput
	}

	var totalPaid float64
	err := r.db.WithContext(ctx).
		Model(&models.SupplierPayment{}).
		Where("purchase_invoice_id = ?", invoiceID).
		Select("COALESCE(SUM(payment_amount), 0)").
		Scan(&totalPaid).Error

	if err != nil {
		return 0, fmt.Errorf("failed to calculate total paid: %w", err)
	}

	return totalPaid, nil
}
