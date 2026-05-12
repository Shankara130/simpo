package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// branchRepository implements BranchRepository interface
// AC2: GORM-based concrete implementation
type branchRepository struct {
	db *gorm.DB
}

// NewBranchRepository creates a new branch repository
// AC5: Factory function for dependency injection
func NewBranchRepository(db interface{}) BranchRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &branchRepository{db: gormDB}
}

// Create inserts a new branch into the database
// AC3: Error handling with wrapping
// P-012: Add nil pointer and validation checks
func (r *branchRepository) Create(ctx context.Context, branch *models.Branch) error {
	if branch == nil {
		return fmt.Errorf("branch cannot be nil")
	}
	if branch.Name == "" {
		return fmt.Errorf("branch name is required")
	}
	err := r.db.WithContext(ctx).Create(branch).Error
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

// GetByID retrieves a branch by its ID
// AC3: Distinguish between "not found" and other errors
// P-011: Add zero ID validation
func (r *branchRepository) GetByID(ctx context.Context, id uint) (*models.Branch, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	var branch models.Branch
	err := r.db.WithContext(ctx).First(&branch, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}
	return &branch, nil
}

// GetByName retrieves a branch by its name
// AC3: Descriptive error for business logic consumption
func (r *branchRepository) GetByName(ctx context.Context, name string) (*models.Branch, error) {
	var branch models.Branch
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&branch).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get branch by name: %w", err)
	}
	return &branch, nil
}

// Update modifies an existing branch in the database
// AC3: Error wrapping for context
func (r *branchRepository) Update(ctx context.Context, branch *models.Branch) error {
	if branch == nil {
		return fmt.Errorf("branch cannot be nil")
	}
	err := r.db.WithContext(ctx).Save(branch).Error
	if err != nil {
		return fmt.Errorf("failed to update branch: %w", err)
	}
	return nil
}

// Delete removes a branch from the database (soft delete)
// AC3: Descriptive error message
// P-004: Check RowsAffected to detect non-existent records
func (r *branchRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Branch{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete branch: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves branches with optional filtering and pagination
// AC4: Complex query support with filtering and pagination
// P-001, P-003, P-005, P-006, P-007, P-008: Security and validation fixes
func (r *branchRepository) List(ctx context.Context, filter *BranchFilter) ([]*models.Branch, int64, error) {
	// P-005: Check context cancellation
	select {
	case <-ctx.Done():
		return nil, 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// P-007: Handle nil filter
	if filter == nil {
		filter = &BranchFilter{}
	}

	var branches []*models.Branch
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Branch{})

	// P-008: Sanitize search query - remove wildcard characters
	if filter.SearchQuery != "" {
		search := strings.ReplaceAll(filter.SearchQuery, "%", "")
		search = strings.ReplaceAll(search, "_", "")
		search = strings.ReplaceAll(search, "\\", "")
		if len(search) > 100 {
			search = search[:100]
		}
		if search != "" {
			query = query.Where("name LIKE ? OR address LIKE ?",
				"%"+search+"%", "%"+search+"%")
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count branches: %w", err)
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
	if err := query.Find(&branches).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list branches: %w", err)
	}

	return branches, total, nil
}
