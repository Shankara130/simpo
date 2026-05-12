package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// transactionItemRepository implements TransactionItemRepository interface
// AC2: GORM-based concrete implementation
type transactionItemRepository struct {
	db *gorm.DB
}

// NewTransactionItemRepository creates a new transaction item repository
// AC5: Factory function for dependency injection
func NewTransactionItemRepository(db interface{}) TransactionItemRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &transactionItemRepository{db: gormDB}
}

// Create inserts a new transaction item into the database
// AC3: Error handling with wrapping
func (r *transactionItemRepository) Create(ctx context.Context, item *models.TransactionItem) error {
	if item == nil {
		return fmt.Errorf("transaction item cannot be nil")
	}
	err := r.db.WithContext(ctx).Create(item).Error
	if err != nil {
		return fmt.Errorf("failed to create transaction item: %w", err)
	}
	return nil
}

// GetByID retrieves a transaction item by its ID
// AC3: Distinguish between "not found" and other errors
// P-011: Add zero ID validation
func (r *transactionItemRepository) GetByID(ctx context.Context, id uint) (*models.TransactionItem, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	var item models.TransactionItem
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transaction item: %w", err)
	}
	return &item, nil
}

// GetByTransactionID retrieves all items for a specific transaction
// Returns empty slice if no items found (not an error)
func (r *transactionItemRepository) GetByTransactionID(ctx context.Context, transactionID uint) ([]*models.TransactionItem, error) {
	var items []*models.TransactionItem
	err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction items: %w", err)
	}
	return items, nil
}

// Update modifies an existing transaction item in the database
// AC3: Error wrapping for context
func (r *transactionItemRepository) Update(ctx context.Context, item *models.TransactionItem) error {
	if item == nil {
		return fmt.Errorf("transaction item cannot be nil")
	}
	err := r.db.WithContext(ctx).Save(item).Error
	if err != nil {
		return fmt.Errorf("failed to update transaction item: %w", err)
	}
	return nil
}

// Delete removes a transaction item from the database (soft delete)
// AC3: Error wrapping with context
// P-004: Check RowsAffected to detect non-existent records
func (r *transactionItemRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.TransactionItem{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete transaction item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves transaction items with optional filtering and pagination
// AC4: Complex query support with filtering by transaction, product
// P-001, P-003, P-005, P-006, P-007: Security and validation fixes
func (r *transactionItemRepository) List(ctx context.Context, filter *TransactionItemFilter) ([]*models.TransactionItem, int64, error) {
	// P-005: Check context cancellation
	select {
	case <-ctx.Done():
		return nil, 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// P-007: Handle nil filter
	if filter == nil {
		filter = &TransactionItemFilter{}
	}

	var items []*models.TransactionItem
	var total int64

	query := r.db.WithContext(ctx).Model(&models.TransactionItem{})

	// Apply filters
	if filter.TransactionID != nil {
		query = query.Where("transaction_id = ?", *filter.TransactionID)
	}
	if filter.ProductID != nil {
		query = query.Where("product_id = ?", *filter.ProductID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count transaction items: %w", err)
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
		"id": true, "created_at": true, "updated_at": true,
		"quantity": true, "subtotal": true,
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
	if err := query.Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list transaction items: %w", err)
	}

	return items, total, nil
}

// CreateBatch inserts multiple transaction items in a single operation
// Add improved error context
func (r *transactionItemRepository) CreateBatch(ctx context.Context, items []*models.TransactionItem) error {
	if len(items) == 0 {
		return nil
	}

	err := r.db.WithContext(ctx).Create(&items).Error
	if err != nil {
		return fmt.Errorf("failed to create transaction items batch (size=%d): %w", len(items), err)
	}
	return nil
}
