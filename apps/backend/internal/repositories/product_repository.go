package repositories

import (
	"context"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// ProductRepository defines the interface for product data operations
// AC1: Repository interface with CRUD methods for Product entity
type ProductRepository interface {
	// Create inserts a new product into the database
	Create(ctx context.Context, product *models.Product) error

	// GetByID retrieves a product by its ID
	// Returns ErrNotFound if product doesn't exist
	GetByID(ctx context.Context, id uint) (*models.Product, error)

	// GetBySKU retrieves a product by its SKU within a branch
	// Returns ErrNotFound if product doesn't exist
	GetBySKU(ctx context.Context, branchID uint, sku string) (*models.Product, error)

	// Update modifies an existing product in the database
	Update(ctx context.Context, product *models.Product) error

	// UpdateStock updates the stock quantity for a product
	UpdateStock(ctx context.Context, id uint, quantity int64) error

	// Delete removes a product from the database (soft delete)
	Delete(ctx context.Context, id uint) error

	// List retrieves products with optional filtering and pagination
	// Returns slice of products, total count, and error
	List(ctx context.Context, filter *ProductFilter) ([]*models.Product, int64, error)

	// GetLowStockProducts retrieves products with stock below reorder threshold
	GetLowStockProducts(ctx context.Context, branchID uint) ([]*models.Product, error)

	// GetExpiredProducts retrieves products that have expired
	GetExpiredProducts(ctx context.Context, branchID uint) ([]*models.Product, error)
}

// ProductFilter defines filtering options for product listing
// AC4: Complex query support with filtering, pagination, and sorting
type ProductFilter struct {
	BranchID    *uint     // Filter by branch
	Category    string    // Filter by category
	SearchQuery string    // Search by name or SKU (ILIKE)
	LowStock    bool      // Filter for low stock items
	Expired     bool      // Filter for expired items
	ExpiryBefore *time.Time // Filter for items expiring before date
	Page        int       // Page number (1-indexed)
	Limit       int       // Items per page
	SortBy      string    // Field to sort by (name, price, stock_qty, etc.)
	SortOrder   string    // "asc" or "desc"
}

// ProductSummary represents aggregated product data
type ProductSummary struct {
	TotalProducts   int64 `json:"total_products"`
	LowStockCount   int64 `json:"low_stock_count"`
	ExpiredCount    int64 `json:"expired_count"`
	TotalStockValue int64 `json:"total_stock_value"` // In cents/smallest unit
}
