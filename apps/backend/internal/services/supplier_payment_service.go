package services

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// SupplierPaymentService defines the interface for supplier payment business logic operations
// Story 10.4: Service interface with validation, audit logging, and transaction wrapping for SupplierPayment entity
type SupplierPaymentService interface {
	// RecordPayment records a new supplier payment with validation and audit logging
	// Story 10.4, AC1: Validates payment data, wraps transaction, updates invoice status, logs creation
	RecordPayment(ctx context.Context, request *RecordPaymentRequest, createdBy uint, ipAddress string) (*models.SupplierPayment, error)

	// GetSupplierPaymentByID retrieves a supplier payment by ID
	// Story 10.4: Returns payment details with invoice information
	GetSupplierPaymentByID(ctx context.Context, id uint) (*models.SupplierPayment, error)

	// ListSupplierPayments retrieves supplier payments with filtering and pagination
	// Story 10.4: Supports filtering by invoice, date range, payment method, branch
	ListSupplierPayments(ctx context.Context, filter *SupplierPaymentListFilter) ([]*models.SupplierPayment, int64, error)

	// GetPaymentHistoryBySupplier retrieves payment history grouped by supplier
	// Story 10.4, AC2: Returns payments for a supplier with invoice details
	GetPaymentHistoryBySupplier(ctx context.Context, supplierID uint, filter *PaymentHistoryFilter) ([]*PaymentHistoryResponse, error)
}

// RecordPaymentRequest defines the data required to record a supplier payment
// Story 10.4, AC1: Request DTO for recording payment
type RecordPaymentRequest struct {
	PurchaseInvoiceID uint    `json:"purchaseInvoiceId" binding:"required"`
	PaymentDate      string  `json:"paymentDate" binding:"required"` // Format: YYYY-MM-DD
	PaymentAmount    float64 `json:"paymentAmount" binding:"required,gt=0"`
	PaymentMethod    string  `json:"paymentMethod" binding:"required,oneof=cash transfer e-wallet check other"`
	Notes            string  `json:"notes" binding:"omitempty,max=1000"`
	ReferenceNumber  string  `json:"referenceNumber" binding:"omitempty,max=100"`
}

// SupplierPaymentListFilter defines filtering options for supplier payment listing
// Story 10.4: Filter struct for supplier payment queries with pagination support
type SupplierPaymentListFilter struct {
	PurchaseInvoiceID *uint   // Filter by purchase invoice
	StartDate          *string // Filter by payment date range start (inclusive)
	EndDate            *string // Filter by payment date range end (inclusive)
	PaymentMethod     *string // Filter by payment method
	BranchID           *uint   // Filter by branch
	Page               int     // Page number (1-indexed)
	Limit              int     // Items per page
	SortBy             string  // Field to sort by (payment_date, payment_amount)
	SortOrder          string  // "asc" or "desc"
}

// PaymentHistoryFilter defines filtering options for payment history by supplier
// Story 10.4, AC2: Filter struct for payment history with date range support
type PaymentHistoryFilter struct {
	StartDate *string // Filter by payment date range start (inclusive)
	EndDate   *string // Filter by payment date range end (inclusive)
	Page      int     // Page number (1-indexed)
	Limit     int     // Items per page
}

// PaymentHistoryResponse represents a payment in the history with invoice details
// Story 10.4, AC2: Response DTO for payment history with invoice context
type PaymentHistoryResponse struct {
	ID                uint    `json:"id"`
	PaymentDate       string  `json:"paymentDate"`
	PaymentAmount     float64 `json:"paymentAmount"`
	PaymentMethod     string  `json:"paymentMethod"`
	Notes             string  `json:"notes,omitempty"`
	ReferenceNumber   string  `json:"referenceNumber,omitempty"`
	InvoiceNumber     string  `json:"invoiceNumber"`
	InvoiceDate       string  `json:"invoiceDate"`
	InvoiceTotalAmount float64 `json:"invoiceTotalAmount"`
	RemainingBalance  float64 `json:"remainingBalance"`
}
