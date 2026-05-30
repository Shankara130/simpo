package repositories

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// GoodsReceiptRepository defines the interface for goods receipt data operations
// Story 10.3: Repository interface with CRUD methods for GoodsReceipt entity
type GoodsReceiptRepository interface {
	// Create inserts a new goods receipt into the database
	// Story 10.3: Create goods receipt record linking to purchase invoice
	Create(ctx context.Context, receipt *models.GoodsReceipt) error

	// GetByID retrieves a goods receipt by its ID with eager loaded relationships
	// Story 10.3: Get goods receipt with purchase invoice details
	// Returns ErrNotFound if receipt doesn't exist
	GetByID(ctx context.Context, id uint) (*models.GoodsReceipt, error)

	// GetByInvoiceID retrieves a goods receipt by its purchase invoice ID
	// Story 10.3: Get goods receipt for a specific invoice (one-to-one relationship)
	// Returns ErrNotFound if receipt doesn't exist for this invoice
	GetByInvoiceID(ctx context.Context, invoiceID uint) (*models.GoodsReceipt, error)

	// List retrieves goods receipts with optional filtering and pagination
	// Story 10.3: List goods receipts with filters (branch, date range, pagination)
	// Returns slice of receipts, total count, and error
	List(ctx context.Context, filter *GoodsReceiptFilter) ([]*models.GoodsReceipt, int64, error)
}

// GoodsReceiptFilter defines filtering options for goods receipt listing
// Story 10.3: Filter struct for goods receipt queries with pagination support
type GoodsReceiptFilter struct {
	BranchID   *uint   // Filter by branch
	StartDate  *string // Filter by received date range start (inclusive)
	EndDate    *string // Filter by received date range end (inclusive)
	ReceivedBy *uint   // Filter by user who processed the receipt
	Page       int     // Page number (1-indexed)
	Limit      int     // Items per page
	SortBy     string  // Field to sort by (received_date, created_at)
	SortOrder  string  // "asc" or "desc"
}
