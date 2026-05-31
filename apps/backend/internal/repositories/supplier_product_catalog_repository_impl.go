package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// supplierProductCatalogRepository implements SupplierProductCatalogRepository interface
// Story 10.5: GORM-based concrete implementation
type supplierProductCatalogRepository struct {
	db *gorm.DB
}

// NewSupplierProductCatalogRepository creates a new supplier product catalog repository
// Story 10.5: Factory function for dependency injection
func NewSupplierProductCatalogRepository(db interface{}) SupplierProductCatalogRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &supplierProductCatalogRepository{db: gormDB}
}

// Create inserts a new supplier product catalog entry into the database
// Story 10.5, AC1: Associate a product with a supplier and specify purchase price
func (r *supplierProductCatalogRepository) Create(ctx context.Context, catalog *models.SupplierProductCatalog) error {
	if catalog == nil {
		return fmt.Errorf("catalog entry cannot be nil")
	}
	if catalog.SupplierID == 0 {
		return fmt.Errorf("supplier ID is required")
	}
	if catalog.ProductID == 0 {
		return fmt.Errorf("product ID is required")
	}
	if catalog.BranchID == 0 {
		return fmt.Errorf("branch ID is required")
	}
	if catalog.CreatedBy == 0 {
		return fmt.Errorf("createdBy user ID is required")
	}

	// Validate purchase price is positive (PATCH-009: negative value validation)
	if catalog.PurchasePrice <= 0 {
		return fmt.Errorf("purchase price must be positive")
	}

	// Validate minimum order quantity (PATCH-009: negative value validation)
	if catalog.MinimumOrderQuantity < 1 {
		return fmt.Errorf("minimum order quantity must be at least 1")
	}

	// Validate lead time days if provided (PATCH-009)
	if catalog.LeadTimeDays != nil && *catalog.LeadTimeDays < 0 {
		return fmt.Errorf("lead time days must be non-negative")
	}

	// Set timestamps if not provided
	if catalog.CreatedAt.IsZero() {
		catalog.CreatedAt = time.Now().UTC()
	}
	if catalog.UpdatedAt.IsZero() {
		catalog.UpdatedAt = time.Now().UTC()
	}
	if catalog.PriceEffectiveFrom.IsZero() {
		catalog.PriceEffectiveFrom = time.Now().UTC()
	}

	// Validate supplier exists and is active (not deleted)
	var supplier models.Supplier
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", catalog.SupplierID).
		First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("supplier not found")
		}
		return fmt.Errorf("failed to validate supplier: %w", err)
	}

	// Validate product exists and is active (not deleted)
	var product models.Product
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", catalog.ProductID).
		First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("product not found")
		}
		return fmt.Errorf("failed to validate product: %w", err)
	}

	// Validate branch exists
	var branch models.Branch
	if err := r.db.WithContext(ctx).First(&branch, catalog.BranchID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("branch not found")
		}
		return fmt.Errorf("failed to validate branch: %w", err)
	}

	// Check for duplicate current entry (CRITICAL: unique association validation)
	// Story 10.5: Only one current price per supplier-product combination
	var existingCount int64
	if err := r.db.WithContext(ctx).
		Model(&models.SupplierProductCatalog{}).
		Where("supplier_id = ? AND product_id = ? AND branch_id = ? AND price_effective_to IS NULL",
			catalog.SupplierID, catalog.ProductID, catalog.BranchID).
		Count(&existingCount).Error; err != nil {
		return fmt.Errorf("failed to check for existing catalog entry: %w", err)
	}
	if existingCount > 0 {
		return fmt.Errorf("catalog entry already exists for this supplier-product combination")
	}

	// Create the catalog entry
	if err := r.db.WithContext(ctx).Create(catalog).Error; err != nil {
		return fmt.Errorf("failed to create catalog entry: %w", err)
	}

	return nil
}

// GetByID retrieves a catalog entry by its ID with eager loaded relationships
// Story 10.5: Get catalog entry details for viewing
func (r *supplierProductCatalogRepository) GetByID(ctx context.Context, id uint) (*models.SupplierProductCatalog, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid catalog ID")
	}

	var catalog models.SupplierProductCatalog
	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Preload("Product").
		Preload("Branch").
		Where("id = ?", id).
		First(&catalog).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("catalog entry not found")
		}
		return nil, fmt.Errorf("failed to get catalog entry: %w", err)
	}

	return &catalog, nil
}

// GetBySupplierAndProduct retrieves all catalog entries (current and historical) for a supplier-product combination
// Story 10.5: Get price history for a specific supplier-product pair
func (r *supplierProductCatalogRepository) GetBySupplierAndProduct(ctx context.Context, supplierID, productID uint) ([]*models.SupplierProductCatalog, error) {
	if supplierID == 0 || productID == 0 {
		return nil, fmt.Errorf("supplier ID and product ID are required")
	}

	var catalogs []*models.SupplierProductCatalog
	err := r.db.WithContext(ctx).
		Where("supplier_id = ? AND product_id = ?", supplierID, productID).
		Order("price_effective_from DESC").
		Find(&catalogs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get catalog entries: %w", err)
	}

	return catalogs, nil
}

// GetCurrentPrice retrieves the current active price entry for a supplier-product combination
// Story 10.5: Get current price (price_effective_to IS NULL)
func (r *supplierProductCatalogRepository) GetCurrentPrice(ctx context.Context, supplierID, productID uint) (*models.SupplierProductCatalog, error) {
	if supplierID == 0 || productID == 0 {
		return nil, fmt.Errorf("supplier ID and product ID are required")
	}

	var catalog models.SupplierProductCatalog
	err := r.db.WithContext(ctx).
		Where("supplier_id = ? AND product_id = ? AND price_effective_to IS NULL", supplierID, productID).
		First(&catalog).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no current price found for this supplier-product combination")
		}
		return nil, fmt.Errorf("failed to get current price: %w", err)
	}

	return &catalog, nil
}

// List retrieves catalog entries with optional filtering and pagination
// Story 10.5: List catalog entries for supplier product catalog views
func (r *supplierProductCatalogRepository) List(ctx context.Context, filter *SupplierProductCatalogFilter) ([]*models.SupplierProductCatalog, int64, error) {
	if filter == nil {
		filter = &SupplierProductCatalogFilter{}
	}

	// Set default pagination values to prevent division by zero (PATCH-012)
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}

	// Build query
	query := r.db.WithContext(ctx).Model(&models.SupplierProductCatalog{})

	// Apply filters
	if filter.SupplierID != nil {
		query = query.Where("supplier_id = ?", *filter.SupplierID)
	}
	if filter.ProductID != nil {
		query = query.Where("product_id = ?", *filter.ProductID)
	}
	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.IsPreferred != nil {
		query = query.Where("is_preferred = ?", *filter.IsPreferred)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count catalog entries: %w", err)
	}

	// Apply sorting
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "price_effective_from"
	}
	sortOrder := filter.SortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit)

	// Execute query with eager loading
	var catalogs []*models.SupplierProductCatalog
	if err := query.
		Preload("Supplier").
		Preload("Product").
		Preload("Branch").
		Find(&catalogs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list catalog entries: %w", err)
	}

	return catalogs, total, nil
}

// GetPriceHistory retrieves price history for a product within an optional date range
// Story 10.5: Track cost changes over time with date range filtering
func (r *supplierProductCatalogRepository) GetPriceHistory(ctx context.Context, productID uint, startDate, endDate *time.Time) ([]*models.SupplierProductCatalog, error) {
	if productID == 0 {
		return nil, fmt.Errorf("product ID is required")
	}

	query := r.db.WithContext(ctx).
		Model(&models.SupplierProductCatalog{}).
		Where("product_id = ?", productID).
		Preload("Supplier")

	// Apply date range filter if provided
	if startDate != nil {
		query = query.Where("price_effective_from >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("(price_effective_to <= ? OR price_effective_to IS NULL)", *endDate)
	}

	// Order by date descending (most recent first)
	query = query.Order("price_effective_from DESC")

	var catalogs []*models.SupplierProductCatalog
	if err := query.Find(&catalogs).Error; err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}

	return catalogs, nil
}

// UpdatePrice updates the purchase price for a catalog entry with price history tracking
// Story 10.5: Archive old price (set price_effective_to) and create new entry with new price
// CRITICAL: Transaction wrapping for atomic price history tracking
func (r *supplierProductCatalogRepository) UpdatePrice(ctx context.Context, catalogID uint, newPrice float64, updatedBy uint) error {
	if catalogID == 0 {
		return fmt.Errorf("invalid catalog ID")
	}
	if updatedBy == 0 {
		return fmt.Errorf("updatedBy user ID is required")
	}

	// Validate new price is positive (PATCH-009)
	if newPrice <= 0 {
		return fmt.Errorf("new price must be positive")
	}

	// Use transaction for atomic price update (CRITICAL from Story 10-2, 10-3)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get current catalog entry
		var currentCatalog models.SupplierProductCatalog
		if err := tx.Where("id = ?", catalogID).First(&currentCatalog).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("catalog entry not found")
			}
			return fmt.Errorf("failed to get catalog entry: %w", err)
		}

		// Archive old price: set price_effective_to to current date
		now := time.Now().UTC()
		if err := tx.Model(&currentCatalog).
			Updates(map[string]interface{}{
				"price_effective_to": now,
				"updated_by":         updatedBy,
				"updated_at":         now,
			}).Error; err != nil {
			return fmt.Errorf("failed to archive old price: %w", err)
		}

		// Create new entry with new price
		newCatalog := models.SupplierProductCatalog{
			SupplierID:          currentCatalog.SupplierID,
			ProductID:           currentCatalog.ProductID,
			PurchasePrice:       newPrice,
			IsPreferred:         currentCatalog.IsPreferred,
			SKUCode:             currentCatalog.SKUCode,
			MinimumOrderQuantity: currentCatalog.MinimumOrderQuantity,
			LeadTimeDays:        currentCatalog.LeadTimeDays,
			BranchID:            currentCatalog.BranchID,
			CreatedBy:           updatedBy,
			PriceEffectiveFrom: now,
		}

		if err := tx.Create(&newCatalog).Error; err != nil {
			return fmt.Errorf("failed to create new price entry: %w", err)
		}

		return nil
	})
}

// SetPreferredSupplier sets or unsets the preferred supplier flag for a product
// Story 10.5: Mark preferred supplier; unset previous preferred if setting new one
func (r *supplierProductCatalogRepository) SetPreferredSupplier(ctx context.Context, supplierID, productID uint, isPreferred bool, branchID uint) error {
	if supplierID == 0 || productID == 0 {
		return fmt.Errorf("supplier ID and product ID are required")
	}

	// Use transaction for atomic preferred supplier update (CRITICAL)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if isPreferred {
			// Unset previous preferred supplier for this product-branch combination
			if err := tx.Model(&models.SupplierProductCatalog{}).
				Where("product_id = ? AND branch_id = ? AND is_preferred = ?", productID, branchID, true).
				Update("is_preferred", false).Error; err != nil {
				return fmt.Errorf("failed to unset previous preferred supplier: %w", err)
			}
		}

		// Set or unset the preferred flag for the specified supplier-product combination
		result := tx.Model(&models.SupplierProductCatalog{}).
			Where("supplier_id = ? AND product_id = ? AND branch_id = ? AND price_effective_to IS NULL",
				supplierID, productID, branchID).
			Update("is_preferred", isPreferred)

		if result.Error != nil {
			return fmt.Errorf("failed to update preferred supplier: %w", result.Error)
		}

		// Check if any row was affected
		if result.RowsAffected == 0 {
			return fmt.Errorf("catalog entry not found for this supplier-product combination")
		}

		return nil
	})
}

// GetPreferredSupplier retrieves the preferred supplier catalog entry for a product
// Story 10.5: Get preferred supplier for product (one per product per branch)
func (r *supplierProductCatalogRepository) GetPreferredSupplier(ctx context.Context, productID uint, branchID uint) (*models.SupplierProductCatalog, error) {
	if productID == 0 || branchID == 0 {
		return nil, fmt.Errorf("product ID and branch ID are required")
	}

	var catalog models.SupplierProductCatalog
	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Where("product_id = ? AND branch_id = ? AND is_preferred = ? AND price_effective_to IS NULL",
			productID, branchID, true).
		First(&catalog).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no preferred supplier found for this product")
		}
		return nil, fmt.Errorf("failed to get preferred supplier: %w", err)
	}

	return &catalog, nil
}

// GetCatalogBySupplier retrieves all current catalog entries for a specific supplier
// Story 10.5: Get supplier's product catalog with current prices
func (r *supplierProductCatalogRepository) GetCatalogBySupplier(ctx context.Context, supplierID uint, branchID uint) ([]*models.SupplierProductCatalog, error) {
	if supplierID == 0 || branchID == 0 {
		return nil, fmt.Errorf("supplier ID and branch ID are required")
	}

	var catalogs []*models.SupplierProductCatalog
	err := r.db.WithContext(ctx).
		Preload("Product").
		Where("supplier_id = ? AND branch_id = ? AND price_effective_to IS NULL", supplierID, branchID).
		Order("product_id ASC").
		Find(&catalogs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get supplier catalog: %w", err)
	}

	return catalogs, nil
}
