package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// userRepository implements UserRepository interface
// AC2: GORM-based concrete implementation
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
// AC5: Factory function for dependency injection
func NewUserRepository(db interface{}) UserRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &userRepository{db: gormDB}
}

// Create inserts a new user into the database
// AC3: Error handling with wrapping
func (r *userRepository) Create(ctx context.Context, userModel *user.User) error {
	if userModel == nil {
		return fmt.Errorf("user cannot be nil")
	}
	err := r.db.WithContext(ctx).Create(userModel).Error
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by its ID
// AC3: Distinguish between "not found" and other errors
// P-011: Add zero ID validation
func (r *userRepository) GetByID(ctx context.Context, id uint) (*user.User, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	var userModel user.User
	err := r.db.WithContext(ctx).First(&userModel, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &userModel, nil
}

// GetByUsername retrieves a user by their username
// AC3: Descriptive error for business logic consumption
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var userModel user.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&userModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &userModel, nil
}

// GetByEmail retrieves a user by their email address
// AC3: Descriptive error for business logic consumption
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var userModel user.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&userModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &userModel, nil
}

// Update modifies an existing user in the database
// AC3: Error wrapping for context
func (r *userRepository) Update(ctx context.Context, userModel *user.User) error {
	if userModel == nil {
		return fmt.Errorf("user cannot be nil")
	}
	err := r.db.WithContext(ctx).Save(userModel).Error
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete removes a user from the database (soft delete)
// AC3: Error wrapping with context
// P-004: Check RowsAffected to detect non-existent records
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&user.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Deactivate marks a user as inactive
// P-012: Check RowsAffected to detect non-existent user
func (r *userRepository) Deactivate(ctx context.Context, id uint, reason string, deactivatedBy uint) error {
	if reason == "" {
		return fmt.Errorf("deactivation reason is required")
	}

	now := time.Now()
	result := r.db.WithContext(ctx).Model(&user.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":              user.UserStatusInactive,
			"deactivated_at":      &now,
			"deactivated_by":      &deactivatedBy,
			"deactivation_reason": reason,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves users with optional filtering and pagination
// AC4: Complex query support with filtering by role, branch, status
// P-001, P-003, P-005, P-006, P-007: Security and validation fixes
func (r *userRepository) List(ctx context.Context, filter *UserFilter) ([]*user.User, int64, error) {
	// P-005: Check context cancellation
	select {
	case <-ctx.Done():
		return nil, 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// P-007: Handle nil filter
	if filter == nil {
		filter = &UserFilter{}
	}

	var users []*user.User
	var total int64

	query := r.db.WithContext(ctx).Model(&user.User{})

	// Apply filters
	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
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
			query = query.Where("name LIKE ? OR username LIKE ? OR email LIKE ?",
				"%"+search+"%", "%"+search+"%", "%"+search+"%")
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
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
		"id": true, "username": true, "name": true, "email": true,
		"created_at": true, "updated_at": true, "role": true, "status": true,
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
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// GetByBranch retrieves all users assigned to a specific branch
func (r *userRepository) GetByBranch(ctx context.Context, branchID uint) ([]*user.User, error) {
	var users []*user.User
	err := r.db.WithContext(ctx).Where("branch_id = ?", branchID).Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get users by branch: %w", err)
	}
	return users, nil
}

// ExistsByUsername checks if a username exists
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&user.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return count > 0, nil
}

// ExistsByEmail checks if an email exists
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&user.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return count > 0, nil
}
