package repositories

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// SupplierPaymentRepository defines the interface for supplier payment data operations
// Story 10.4: Repository interface with CRUD methods for SupplierPayment entity
type SupplierPaymentRepository interface {
	// Create inserts a new supplier payment into the database
	// Story 10.4: Record payment made to supplier for a purchase invoice
	Create(ctx context.Context, payment *models.SupplierPayment) error

	// GetByID retrieves a supplier payment by its ID with eager loaded relationships
	// Story 10.4: Get payment details for viewing
	// Returns ErrNotFound if payment doesn't exist
	GetByID(ctx context.Context, id uint) (*models.SupplierPayment, error)

	// GetByInvoiceID retrieves all supplier payments for a specific purchase invoice
	// Story 10.4: Get payment history for a specific invoice
	// Returns empty slice if no payments exist (not an error)
	GetByInvoiceID(ctx context.Context, invoiceID uint) ([]*models.SupplierPayment, error)

	// List retrieves supplier payments with optional filtering and pagination
	// Story 10.4: List payments for payment history views
	// Returns slice of payments, total count, and error
	List(ctx context.Context, filter *SupplierPaymentFilter) ([]*models.SupplierPayment, int64, error)

	// GetTotalPaidByInvoice calculates the total amount paid for a specific purchase invoice
	// Story 10.4: Used for payment status calculation (unpaid/partial/fully paid)
	// Returns 0 if no payments exist (not an error)
	GetTotalPaidByInvoice(ctx context.Context, invoiceID uint) (float64, error)
}

// SupplierPaymentFilter defines filtering options for supplier payment listing
// Story 10.4: Filter struct for supplier payment queries with pagination support
type SupplierPaymentFilter struct {
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
