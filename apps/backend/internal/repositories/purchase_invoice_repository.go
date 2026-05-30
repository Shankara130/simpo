package repositories

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// PurchaseInvoiceRepository defines the interface for purchase invoice data operations
// Story 10.2: Repository interface with CRUD methods for PurchaseInvoice entity
type PurchaseInvoiceRepository interface {
	// Create inserts a new purchase invoice into the database with optional line items
	// createdBy is the user ID who is creating the invoice
	// items is the slice of line items to create (can be empty for legacy calls)
	// DN-001: Updated to support creating line items within the same transaction
	Create(ctx context.Context, invoice *models.PurchaseInvoice, createdBy uint, items []models.PurchaseInvoiceItem) error

	// GetByID retrieves a purchase invoice by its ID with eager loaded relationships
	// Returns ErrNotFound if invoice doesn't exist
	GetByID(ctx context.Context, id uint) (*models.PurchaseInvoice, error)

	// GetByInvoiceNumber retrieves a purchase invoice by its invoice number
	// Returns ErrNotFound if invoice doesn't exist
	GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*models.PurchaseInvoice, error)

	// Update modifies an existing purchase invoice in the database
	// updatedBy is the user ID who is updating the invoice
	Update(ctx context.Context, invoice *models.PurchaseInvoice, updatedBy uint) error

	// Delete soft deletes a purchase invoice (sets deleted_at timestamp)
	// deletedBy is the user ID who is deleting the invoice
	Delete(ctx context.Context, id uint, deletedBy uint) error

	// List retrieves purchase invoices with optional filtering and pagination
	// Returns slice of invoices, total count, and error
	List(ctx context.Context, filter *PurchaseInvoiceFilter) ([]*models.PurchaseInvoice, int64, error)

	// UpdatePaymentStatus updates the payment status of a purchase invoice based on total payments
	// Story 10.4: Calculate and update payment status (unpaid/partial/fully paid) based on payments
	// Returns error if invoice not found or update fails
	UpdatePaymentStatus(ctx context.Context, invoiceID uint) error
}

// PurchaseInvoiceFilter defines filtering options for purchase invoice listing
// Story 10.2: Filter struct for purchase invoice queries with pagination support
type PurchaseInvoiceFilter struct {
	SupplierID     *uint   // Filter by supplier
	StartDate       *string // Filter by invoice date range start (inclusive)
	EndDate         *string // Filter by invoice date range end (inclusive)
	PaymentStatus  *string // Filter by payment status
	SearchQuery     string  // Search by invoice number
	Page            int     // Page number (1-indexed)
	Limit           int     // Items per page
	SortBy          string  // Field to sort by (invoice_date, total_amount)
	SortOrder       string  // "asc" or "desc"
}
