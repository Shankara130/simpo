package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// supplierRepository implements SupplierRepository interface
// Story 10.1: GORM-based concrete implementation
type supplierRepository struct {
	db *gorm.DB
}

// NewSupplierRepository creates a new supplier repository
// Story 10.1: Factory function for dependency injection
func NewSupplierRepository(db interface{}) SupplierRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &supplierRepository{db: gormDB}
}

// Create inserts a new supplier into the database
// Story 10.1, AC1: Error handling with wrapping
func (r *supplierRepository) Create(ctx context.Context, supplier *models.Supplier, createdBy uint) error {
	if supplier == nil {
		return fmt.Errorf("supplier cannot be nil")
	}
	if supplier.Name == "" {
		return fmt.Errorf("supplier name is required")
	}
	if supplier.Phone == "" {
		return fmt.Errorf("supplier phone is required")
	}
	if createdBy == 0 {
		return fmt.Errorf("createdBy user ID is required")
	}

	supplier.CreatedBy = &createdBy
	supplier.UpdatedBy = &createdBy

	err := r.db.WithContext(ctx).Create(supplier).Error
	if err != nil {
		return fmt.Errorf("failed to create supplier: %w", err)
	}
	return nil
}

// GetByID retrieves a supplier by its ID
// Story 10.1, AC1: Distinguish between "not found" and other errors
func (r *supplierRepository) GetByID(ctx context.Context, id uint) (*models.Supplier, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	var supplier models.Supplier
	err := r.db.WithContext(ctx).First(&supplier, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}
	return &supplier, nil
}

// GetByName retrieves a supplier by its name
// Story 10.1, AC1: Check for duplicate supplier names with Unicode normalization
func (r *supplierRepository) GetByName(ctx context.Context, name string) (*models.Supplier, error) {
	var supplier models.Supplier
	// Normalize name for consistent comparison (handles Unicode equivalence)
	normalizedName := norm.NFKC.String(strings.TrimSpace(name))
	err := r.db.WithContext(ctx).Where("name = ?", normalizedName).First(&supplier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get supplier by name: %w", err)
	}
	return &supplier, nil
}

// Update modifies an existing supplier in the database
// Story 10.1, AC2: Error wrapping for context, optimistic locking with version check
func (r *supplierRepository) Update(ctx context.Context, supplier *models.Supplier, updatedBy uint) error {
	if supplier == nil {
		return fmt.Errorf("supplier cannot be nil")
	}
	if updatedBy == 0 {
		return fmt.Errorf("updatedBy user ID is required")
	}

	// Check if supplier is deactivated (soft deleted)
	if supplier.DeletedAt.Valid {
		return fmt.Errorf("cannot update deactivated supplier")
	}

	// Optimistic locking: check version hasn't changed
	var existing models.Supplier
	err := r.db.WithContext(ctx).Select("version").Where("id = ?", supplier.ID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to check supplier version: %w", err)
	}
	if existing.Version != supplier.Version {
		return fmt.Errorf("supplier was modified by another user (current version: %d, your version: %d)", existing.Version, supplier.Version)
	}

	supplier.UpdatedBy = &updatedBy

	err = r.db.WithContext(ctx).Save(supplier).Error
	if err != nil {
		return fmt.Errorf("failed to update supplier: %w", err)
	}
	return nil
}

// Deactivate soft deletes a supplier (sets deleted_at and deleted_by timestamps)
// Story 10.1, AC3: Check RowsAffected to detect non-existent records
func (r *supplierRepository) Deactivate(ctx context.Context, id uint, deactivatedBy uint) error {
	if id == 0 {
		return ErrInvalidInput
	}
	if deactivatedBy == 0 {
		return fmt.Errorf("deactivatedBy user ID is required")
	}

	// Explicitly set deleted_at and deleted_by using SQL update
	result := r.db.WithContext(ctx).
		Model(&models.Supplier{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
			"deleted_by": deactivatedBy,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate supplier: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves suppliers with optional filtering and pagination
// Story 10.1, AC2: Complex query support with filtering and pagination
func (r *supplierRepository) List(ctx context.Context, filter *SupplierFilter) ([]*models.Supplier, int64, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// Handle nil filter
	if filter == nil {
		filter = &SupplierFilter{}
	}

	var suppliers []*models.Supplier
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Supplier{})

	// Filter by active status if specified
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	// Sanitize search query - remove wildcard characters
	if filter.SearchQuery != "" {
		search := strings.TrimSpace(filter.SearchQuery)
		// Remove wildcard characters to prevent SQL injection
		search = strings.ReplaceAll(search, "%", "")
		search = strings.ReplaceAll(search, "_", "")
		search = strings.ReplaceAll(search, "\\", "")
		// Require minimum search length and limit max length
		if len(search) >= 2 && len(search) <= 100 {
			query = query.Where("name ILIKE ? OR contact_person ILIKE ? OR phone ILIKE ?",
				"%"+search+"%", "%"+search+"%", "%"+search+"%")
		}
		// If search is too short after sanitization, ignore it (returns all active suppliers)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count suppliers: %w", err)
	}

	// Apply pagination with bounds checking
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
	// Check for integer overflow in offset calculation
	if page > 1000000 {
		return nil, 0, fmt.Errorf("page number exceeds maximum allowed")
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// Whitelist validation for sort fields
	allowedSortFields := map[string]bool{
		"id": true, "name": true, "created_at": true, "updated_at": true,
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
	if err := query.Find(&suppliers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list suppliers: %w", err)
	}

	return suppliers, total, nil
}
