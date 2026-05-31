package services

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// SupplierProductCatalogService defines the interface for supplier product catalog business logic operations
// Story 10.5: Service interface with validation, audit logging, and transaction wrapping for SupplierProductCatalog entity
type SupplierProductCatalogService interface {
	// AssociateProduct associates a product with a supplier and specifies purchase price
	// Story 10.5, AC1: Validates supplier and product exist, checks for duplicates, logs creation
	AssociateProduct(ctx context.Context, request *AssociateProductRequest, createdBy uint, ipAddress string) (*models.SupplierProductCatalog, error)

	// GetProductCatalogByID retrieves a catalog entry by ID
	// Story 10.5: Returns catalog entry details with supplier and product information
	GetProductCatalogByID(ctx context.Context, id uint) (*models.SupplierProductCatalog, error)

	// ListProductCatalogs retrieves catalog entries with filtering and pagination
	// Story 10.5: Supports filtering by supplier, product, branch, preferred status
	ListProductCatalogs(ctx context.Context, filter *SupplierProductCatalogListFilter) ([]*models.SupplierProductCatalog, int64, error)

	// UpdatePurchasePrice updates the purchase price with price history tracking
	// Story 10.5, AC1: Archives old price, creates new entry, logs change
	UpdatePurchasePrice(ctx context.Context, catalogID uint, request *UpdatePriceRequest, updatedBy uint, ipAddress string) error

	// SetPreferredSupplier sets or unsets the preferred supplier flag for a product
	// Story 10.5, AC1: Ensures only one preferred supplier per product per branch, logs change
	SetPreferredSupplier(ctx context.Context, catalogID uint, request *SetPreferredRequest, updatedBy uint, ipAddress string) error

	// GetPreferredSupplier retrieves the preferred supplier for a product
	// Story 10.5: Returns preferred supplier catalog entry or error if none set
	GetPreferredSupplier(ctx context.Context, productID uint, branchID uint) (*models.SupplierProductCatalog, error)

	// GetPriceHistory retrieves price history for a product with optional date range filtering
	// Story 10.5, AC1: Returns historical prices grouped by supplier with date ranges
	GetPriceHistory(ctx context.Context, productID uint, filter *PriceHistoryFilter) ([]*PriceHistoryEntry, error)

	// GetCatalogBySupplier retrieves a supplier's product catalog with current prices
	// Story 10.5: Returns all current catalog entries for a supplier
	GetCatalogBySupplier(ctx context.Context, supplierID uint, branchID uint) ([]*models.SupplierProductCatalog, error)
}

// AssociateProductRequest defines the data required to associate a product with a supplier
// Story 10.5, AC1: Request DTO for product-supplier association
type AssociateProductRequest struct {
	SupplierID          uint    `json:"supplierId" binding:"required"`
	ProductID           uint    `json:"productId" binding:"required"`
	BranchID            uint    `json:"branchId" binding:"required"`
	PurchasePrice       float64 `json:"purchasePrice" binding:"required,gt=0"`
	IsPreferred         bool    `json:"isPreferred"`
	SKUCode             string  `json:"skuCode" binding:"omitempty,max=50"`
	MinimumOrderQuantity int    `json:"minimumOrderQuantity" binding:"required,min=1"`
	LeadTimeDays        *int    `json:"leadTimeDays" binding:"omitempty,min=0"`
}

// UpdatePriceRequest defines the data required to update a purchase price
// Story 10.5, AC1: Request DTO for price update
type UpdatePriceRequest struct {
	NewPrice float64 `json:"newPrice" binding:"required,gt=0"`
}

// SetPreferredRequest defines the data required to set preferred supplier
// Story 10.5, AC1: Request DTO for setting preferred supplier
type SetPreferredRequest struct {
	IsPreferred bool `json:"isPreferred"`
}

// SupplierProductCatalogListFilter defines filtering options for catalog listing
// Story 10.5: Filter struct for catalog queries with pagination support
type SupplierProductCatalogListFilter struct {
	SupplierID  *uint  // Filter by supplier
	ProductID   *uint  // Filter by product
	BranchID    *uint  // Filter by branch
	IsPreferred *bool  // Filter by preferred status only
	Page        int    // Page number (1-indexed)
	Limit       int    // Items per page
	SortBy      string // Field to sort by
	SortOrder   string // "asc" or "desc"
}

// PriceHistoryFilter defines filtering options for price history queries
// Story 10.5, AC1: Filter struct for price history with date range support
type PriceHistoryFilter struct {
	StartDate  *string // Filter by price effective date range start (inclusive)
	EndDate    *string // Filter by price effective date range end (inclusive)
	SupplierID *uint  // Optional filter by supplier
	Page       int    // Page number (1-indexed)
	Limit      int    // Items per page
}

// PriceHistoryEntry represents a price in the history with supplier information
// Story 10.5, AC1: Response DTO for price history with supplier context
type PriceHistoryEntry struct {
	ID             uint    `json:"id"`
	SupplierID     uint    `json:"supplierId"`
	SupplierName   string  `json:"supplierName"`
	ProductID      uint    `json:"productId"`
	ProductName    string  `json:"productName"`
	PurchasePrice  float64 `json:"purchasePrice"`
	EffectiveFrom  string  `json:"effectiveFrom"`
	EffectiveTo    string  `json:"effectiveTo,omitempty"`
	IsCurrent      bool    `json:"isCurrent"`
	IsPreferred    bool    `json:"isPreferred"`
	CreatedBy      uint    `json:"createdBy"`
	CreatedAt      string  `json:"createdAt"`
}

// SupplierProductCatalogResponse represents a catalog entry with full details
// Story 10.5: Response DTO for catalog operations
type SupplierProductCatalogResponse struct {
	ID                   uint    `json:"id"`
	SupplierID           uint    `json:"supplierId"`
	SupplierName         string  `json:"supplierName,omitempty"`
	ProductID            uint    `json:"productId"`
	ProductName          string  `json:"productName,omitempty"`
	PurchasePrice        float64 `json:"purchasePrice"`
	IsPreferred          bool    `json:"isPreferred"`
	SKUCode              string  `json:"skuCode,omitempty"`
	MinimumOrderQuantity int    `json:"minimumOrderQuantity"`
	LeadTimeDays         *int    `json:"leadTimeDays,omitempty"`
	BranchID             uint    `json:"branchId"`
	CreatedBy            uint    `json:"createdBy"`
	CreatedByName        string  `json:"createdByName,omitempty"`
	UpdatedBy            *uint   `json:"updatedBy,omitempty"`
	UpdatedByName        string  `json:"updatedByName,omitempty"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
	PriceEffectiveFrom   string  `json:"priceEffectiveFrom"`
	PriceEffectiveTo     string  `json:"priceEffectiveTo,omitempty"`
	IsCurrentPrice       bool    `json:"isCurrentPrice"`
}
