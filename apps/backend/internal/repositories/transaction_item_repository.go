package repositories

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// TransactionItemRepository defines the interface for transaction item data operations
// AC1: Repository interface with CRUD methods for TransactionItem entity
type TransactionItemRepository interface {
	// Create inserts a new transaction item into the database
	Create(ctx context.Context, item *models.TransactionItem) error

	// GetByID retrieves a transaction item by its ID
	// Returns ErrNotFound if item doesn't exist
	GetByID(ctx context.Context, id uint) (*models.TransactionItem, error)

	// GetByTransactionID retrieves all items for a specific transaction
	// Returns empty slice if no items found (not an error)
	GetByTransactionID(ctx context.Context, transactionID uint) ([]*models.TransactionItem, error)

	// Update modifies an existing transaction item in the database
	Update(ctx context.Context, item *models.TransactionItem) error

	// Delete removes a transaction item from the database (soft delete)
	Delete(ctx context.Context, id uint) error

	// List retrieves transaction items with optional filtering and pagination
	// Returns slice of items, total count, and error
	List(ctx context.Context, filter *TransactionItemFilter) ([]*models.TransactionItem, int64, error)

	// CreateBatch inserts multiple transaction items in a single operation
	CreateBatch(ctx context.Context, items []*models.TransactionItem) error
}

// TransactionItemFilter defines filtering options for transaction item listing
// AC4: Complex query support with filtering by transaction, product
type TransactionItemFilter struct {
	TransactionID *uint  // Filter by transaction
	ProductID     *uint  // Filter by product
	Page          int    // Page number (1-indexed)
	Limit         int    // Items per page
	SortBy        string // Field to sort by
	SortOrder     string // "asc" or "desc"
}

// TransactionItemSummary represents aggregated transaction item data
type TransactionItemSummary struct {
	TotalItems       int64  `json:"total_items"`
	TotalQuantity    int64  `json:"total_quantity"`
	TotalSubtotal    string `json:"total_subtotal"` // Decimal string for precision
}
