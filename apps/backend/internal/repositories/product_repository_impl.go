package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// productRepository implements ProductRepository interface
// AC2: GORM-based concrete implementation
type productRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new product repository
// AC5: Factory function for dependency injection
func NewProductRepository(db interface{}) ProductRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &productRepository{db: gormDB}
}

// Create inserts a new product into the database
// AC3: Error handling with wrapping
func (r *productRepository) Create(ctx context.Context, product *models.Product) error {
	if product == nil {
		return fmt.Errorf("product cannot be nil")
	}
	err := r.db.WithContext(ctx).Create(product).Error
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

// GetByID retrieves a product by its ID
// AC3: Distinguish between "not found" and other errors
// P-011: Add zero ID validation
// P-013: Add eager loading for Branch relationship
func (r *productRepository) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	var product models.Product
	err := r.db.WithContext(ctx).Preload("Branch").First(&product, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &product, nil
}

// GetBySKU retrieves a product by its SKU within a branch
// AC3: Descriptive error for business logic consumption
func (r *productRepository) GetBySKU(ctx context.Context, branchID uint, sku string) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).Preload("Branch").Where("branch_id = ? AND sku = ?", branchID, sku).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}
	return &product, nil
}

// Update modifies an existing product in the database
// AC3: Error wrapping for context
func (r *productRepository) Update(ctx context.Context, product *models.Product) error {
	if product == nil {
		return fmt.Errorf("product cannot be nil")
	}
	err := r.db.WithContext(ctx).Save(product).Error
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

// UpdateStock updates the stock quantity for a product
// AC3: Descriptive error message
// P-002: Fix race condition using atomic increment
func (r *productRepository) UpdateStock(ctx context.Context, id uint, delta int64) error {
	// Use atomic increment/decrement with check for negative stock
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("id = ? AND stock_qty + ? >= 0", id, delta).
		Update("stock_qty", gorm.Expr("stock_qty + ?", delta)).Error
	if err != nil {
		return fmt.Errorf("failed to update product stock: %w", err)
	}
	return nil
}

// Delete removes a product from the database (soft delete)
// AC3: Error wrapping with context
// P-004: Check RowsAffected to detect non-existent records
func (r *productRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Product{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves products with optional filtering and pagination
// AC4: Complex query support with filtering, pagination, and sorting
// P-001, P-003, P-005, P-006, P-007, P-008: Security and validation fixes
func (r *productRepository) List(ctx context.Context, filter *ProductFilter) ([]*models.Product, int64, error) {
	// P-005: Check context cancellation
	select {
	case <-ctx.Done():
		return nil, 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// P-007: Handle nil filter
	if filter == nil {
		filter = &ProductFilter{}
	}

	var products []*models.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Product{})

	// Apply filters
	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.LowStock {
		query = query.Where("stock_qty < reorder_threshold")
	}
	if filter.Expired && filter.ExpiryBefore != nil {
		query = query.Where("expiry_date < ?", *filter.ExpiryBefore)
	} else if filter.Expired {
		query = query.Where("expiry_date < ?", time.Now())
	}

	// P-008: Sanitize search query - remove wildcard characters
	if filter.SearchQuery != "" {
		search := strings.ReplaceAll(filter.SearchQuery, "%", "")
		search = strings.ReplaceAll(search, "_", "")
		search = strings.ReplaceAll(search, "\\", "")
		if len(search) > 100 {
			search = search[:100]
		}
		if search != "" {
			query = query.Where("name LIKE ? OR sku LIKE ?",
				"%"+search+"%", "%"+search+"%")
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// P-003, P-006: Apply pagination with bounds checking
	page := filter.Page
	limit := filter.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20 // default
	}
	if limit > 1000 {
		limit = 1000 // maximum to prevent DoS
	}
	// P-006: Check for integer overflow in offset calculation
	if page > 1000000 {
		return nil, 0, fmt.Errorf("page number exceeds maximum allowed")
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// P-001: Whitelist validation for sort fields
	allowedSortFields := map[string]bool{
		"id": true, "name": true, "sku": true, "created_at": true, "updated_at": true,
		"price": true, "stock_qty": true, "category": true,
	}
	sortBy := "created_at"
	if filter.SortBy != "" {
		if allowedSortFields[filter.SortBy] {
			sortBy = filter.SortBy
		}
	}

	allowedSortOrders := map[string]bool{
		"ASC": true, "DESC": true, "asc": true, "desc": true,
	}
	sortOrder := "DESC"
	if filter.SortOrder != "" {
		normalized := strings.ToUpper(filter.SortOrder)
		if allowedSortOrders[normalized] {
			sortOrder = normalized
		}
	}
	query = query.Order(sortBy + " " + sortOrder)

	// Execute query
	if err := query.Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, total, nil
}

// GetLowStockProducts retrieves products with stock below reorder threshold
// AC4: Complex query for low stock items
func (r *productRepository) GetLowStockProducts(ctx context.Context, branchID uint) ([]*models.Product, error) {
	var products []*models.Product
	err := r.db.WithContext(ctx).
		Where("branch_id = ? AND stock_qty < reorder_threshold", branchID).
		Find(&products).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}
	return products, nil
}

// GetExpiredProducts retrieves products that have expired
// AC4: Complex query for expired items
func (r *productRepository) GetExpiredProducts(ctx context.Context, branchID uint) ([]*models.Product, error) {
	var products []*models.Product
	err := r.db.WithContext(ctx).
		Where("branch_id = ? AND expiry_date < ?", branchID, time.Now()).
		Find(&products).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get expired products: %w", err)
	}
	return products, nil
}
