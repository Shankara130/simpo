package repositories

import (
	"context"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// SupplierProductCatalogRepository defines the interface for supplier product catalog data operations
// Story 10.5: Repository interface with CRUD methods for SupplierProductCatalog entity
type SupplierProductCatalogRepository interface {
	// Create inserts a new supplier product catalog entry into the database
	// Story 10.5: Associate a product with a supplier and specify purchase price
	Create(ctx context.Context, catalog *models.SupplierProductCatalog) error

	// GetByID retrieves a catalog entry by its ID with eager loaded relationships
	// Story 10.5: Get catalog entry details for viewing
	// Returns ErrNotFound if catalog entry doesn't exist
	GetByID(ctx context.Context, id uint) (*models.SupplierProductCatalog, error)

	// GetBySupplierAndProduct retrieves all catalog entries (current and historical) for a specific supplier-product combination
	// Story 10.5: Get price history for a specific supplier-product pair
	// Returns empty slice if no entries exist (not an error)
	GetBySupplierAndProduct(ctx context.Context, supplierID, productID uint) ([]*models.SupplierProductCatalog, error)

	// GetCurrentPrice retrieves the current active price entry for a supplier-product combination
	// Story 10.5: Get current price (price_effective_to IS NULL)
	// Returns ErrNotFound if no current price exists
	GetCurrentPrice(ctx context.Context, supplierID, productID uint) (*models.SupplierProductCatalog, error)

	// List retrieves catalog entries with optional filtering and pagination
	// Story 10.5: List catalog entries for supplier product catalog views
	// Returns slice of catalog entries, total count, and error
	List(ctx context.Context, filter *SupplierProductCatalogFilter) ([]*models.SupplierProductCatalog, int64, error)

	// GetPriceHistory retrieves price history for a product within an optional date range
	// Story 10.5: Track cost changes over time with date range filtering
	// Returns empty slice if no history exists (not an error)
	GetPriceHistory(ctx context.Context, productID uint, startDate, endDate *time.Time) ([]*models.SupplierProductCatalog, error)

	// UpdatePrice updates the purchase price for a catalog entry with price history tracking
	// Story 10.5: Archive old price (set price_effective_to) and create new entry with new price
	// This is a transactional operation that preserves price history
	UpdatePrice(ctx context.Context, catalogID uint, newPrice float64, updatedBy uint) error

	// SetPreferredSupplier sets or unsets the preferred supplier flag for a product
	// Story 10.5: Mark preferred supplier; unset previous preferred if setting new one
	// Returns ErrNotFound if catalog entry doesn't exist
	SetPreferredSupplier(ctx context.Context, supplierID, productID uint, isPreferred bool, branchID uint) error

	// GetPreferredSupplier retrieves the preferred supplier catalog entry for a product
	// Story 10.5: Get preferred supplier for product (one per product per branch)
	// Returns ErrNotFound if no preferred supplier exists
	GetPreferredSupplier(ctx context.Context, productID uint, branchID uint) (*models.SupplierProductCatalog, error)

	// GetCatalogBySupplier retrieves all current catalog entries for a specific supplier
	// Story 10.5: Get supplier's product catalog with current prices
	// Returns empty slice if supplier has no catalog entries (not an error)
	GetCatalogBySupplier(ctx context.Context, supplierID uint, branchID uint) ([]*models.SupplierProductCatalog, error)
}

// SupplierProductCatalogFilter defines filtering options for supplier product catalog listing
// Story 10.5: Filter struct for catalog queries with pagination support
type SupplierProductCatalogFilter struct {
	SupplierID  *uint // Filter by supplier
	ProductID   *uint // Filter by product
	BranchID    *uint // Filter by branch
	IsPreferred *bool // Filter by preferred status only
	Page        int   // Page number (1-indexed)
	Limit       int   // Items per page
	SortBy      string // Field to sort by (price_effective_from, purchase_price, product_name)
	SortOrder   string // "asc" or "desc"
}
