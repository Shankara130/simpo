package services

import (
	"context"
	"fmt"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// userService implements UserService interface
// AC2: Services use repository interfaces (not concrete implementations)
type userService struct {
	userRepo     repositories.UserRepository
	auditService AuditService
}

// NewUserService creates a new user service with dependency injection
// AC2: Services accept repository interfaces via constructor injection
func NewUserService(userRepo repositories.UserRepository, auditService AuditService) UserService {
	// Fail fast on nil dependencies
	if userRepo == nil {
		panic("userService: userRepo cannot be nil")
	}
	if auditService == nil {
		panic("userService: auditService cannot be nil")
	}

	return &userService{
		userRepo:     userRepo,
		auditService: auditService,
	}
}

// CreateUser creates a new user with business validation
// AC3: Business Logic Encapsulation
// Validates: username uniqueness, email uniqueness, required fields
func (s *userService) CreateUser(ctx context.Context, u *user.User, ipAddress string) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate required fields
	if u.Username == "" {
		return &InvalidInputError{Field: "username", Message: "username is required"}
	}
	if u.Email == "" {
		return &InvalidInputError{Field: "email", Message: "email is required"}
	}
	if u.Role == "" {
		return &InvalidInputError{Field: "role", Message: "role is required"}
	}
	if u.PasswordHash == "" {
		return &InvalidInputError{Field: "password", Message: "password is required"}
	}

	// Check username uniqueness (AC3, Business Rules)
	exists, err := s.userRepo.ExistsByUsername(ctx, u.Username)
	if err != nil {
		return &ServiceError{Op: "check username existence", Err: err}
	}
	if exists {
		return &DuplicateUsernameError{Username: u.Username}
	}

	// Check email uniqueness (AC3, Business Rules)
	exists, err = s.userRepo.ExistsByEmail(ctx, u.Email)
	if err != nil {
		return &ServiceError{Op: "check email existence", Err: err}
	}
	if exists {
		return &InvalidInputError{Field: "email", Message: "email already exists"}
	}

	// Create user via repository
	if err := s.userRepo.Create(ctx, u); err != nil {
		return &ServiceError{Op: "create user", Err: err}
	}

	// Log user creation (audit)
	_ = s.auditService.LogUserCreation(ctx, 0, u.ID, "system", u.Username, ipAddress)

	return nil
}

// UpdateUser modifies an existing user with business rules
// AC3: Business rules: role changes require admin role, cannot change own role
func (s *userService) UpdateUser(ctx context.Context, userID uint, u *user.User, adminID uint, ipAddress string) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate ID
	if userID == 0 {
		return &InvalidInputError{Field: "id", Message: "user ID is required"}
	}

	// Get existing user
	existing, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return &UserNotFoundError{UserID: userID}
	}

	// PATCH: Validate email uniqueness if email is being changed
	if u.Email != "" && u.Email != existing.Email {
		// Check if email is already in use by another user
		_, err := s.userRepo.GetByEmail(ctx, u.Email)
		if err == nil {
			return &InvalidInputError{
				Field:   "email",
				Message: "email is already in use by another user",
			}
		}
	}

	// Business rule: cannot change own role (AC3)
	if adminID == userID && u.Role != "" && u.Role != existing.Role {
		return &UnauthorizedError{
			Action: "update own role",
			Reason: "users cannot change their own role",
		}
	}

	// Preserve fields that should not be updated
	u.ID = userID
	u.Username = existing.Username
	u.CreatedAt = existing.CreatedAt

	// PATCH: Preserve required fields if empty (prevent clearing)
	if u.Email == "" {
		u.Email = existing.Email
	}
	if u.Role == "" {
		u.Role = existing.Role
	}
	if u.BranchID == nil && existing.BranchID != nil {
		u.BranchID = existing.BranchID
	}

	// Update user via repository
	if err := s.userRepo.Update(ctx, u); err != nil {
		return &ServiceError{Op: "update user", Err: err}
	}

	return nil
}

// DeactivateUser deactivates a user account with audit logging
func (s *userService) DeactivateUser(ctx context.Context, userID uint, reason string, deactivatedBy uint, ipAddress string) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate ID
	if userID == 0 {
		return &InvalidInputError{Field: "id", Message: "user ID is required"}
	}

	// Get user to deactivate
	targetUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return &UserNotFoundError{UserID: userID}
	}

	// Business rule: cannot deactivate yourself (AC3)
	if deactivatedBy == userID {
		return &UnauthorizedError{
			Action: "deactivate self",
			Reason: "users cannot deactivate their own account",
		}
	}

	// PATCH: Check if this is the last admin for the branch
	if targetUser.Role == "Admin" && targetUser.BranchID != nil {
		// Count other admins in the same branch
		admins, _, err := s.userRepo.List(ctx, &repositories.UserFilter{
			BranchID: targetUser.BranchID,
			Role:     "Admin",
			Status:   "ACTIVE",
			Limit:    10, // We only need to know if there's more than 1
		})
		if err != nil {
			return &ServiceError{Op: "check branch admins", Err: err}
		}
		// If this is the only active admin, prevent deactivation
		if len(admins) <= 1 {
			return &UnauthorizedError{
				Action: "deactivate last admin",
				Reason: "cannot deactivate the last admin for a branch",
			}
		}
	}

	// Get admin user
	adminUser, err := s.userRepo.GetByID(ctx, deactivatedBy)
	if err != nil {
		return &UserNotFoundError{UserID: deactivatedBy}
	}

	// Deactivate via repository
	if err := s.userRepo.Deactivate(ctx, userID, reason, deactivatedBy); err != nil {
		return &ServiceError{Op: "deactivate user", Err: err}
	}

	// Log deactivation (audit)
	_ = s.auditService.LogUserDeactivation(ctx, deactivatedBy, userID, adminUser.Username, targetUser.Username, reason, ipAddress)

	return nil
}

// ListUsers retrieves users with filtering and pagination
// AC3: Filters: role, branch, status
func (s *userService) ListUsers(ctx context.Context, filter *UserFilter) ([]*user.User, int64, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("operation cancelled: %w", err)
	}

	// Default filter if nil
	if filter == nil {
		filter = &UserFilter{
			Page:  1,
			Limit: 20,
		}
	}

	// Validate pagination (Epic 2 retro: bounds checking)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}

	// PATCH: Validate search query length (prevent performance issues)
	if filter.SearchQuery != "" && len(filter.SearchQuery) > 200 {
		return nil, 0, &InvalidInputError{
			Field:   "search_query",
			Message: "search query cannot exceed 200 characters",
		}
	}

	// PATCH: Validate SortBy and SortOrder (prevent SQL injection via ORDER BY)
	allowedSortFields := map[string]bool{
		"id": true, "username": true, "email": true, "role": true,
		"branch_id": true, "status": true, "created_at": true, "updated_at": true,
	}
	if filter.SortBy != "" && !allowedSortFields[filter.SortBy] {
		return nil, 0, &InvalidInputError{
			Field:   "sort_by",
			Message: "invalid sort field",
		}
	}
	if filter.SortOrder != "" && filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		return nil, 0, &InvalidInputError{
			Field:   "sort_order",
			Message: "sort order must be 'asc' or 'desc'",
		}
	}

	// Convert to repository filter
	repoFilter := &repositories.UserFilter{
		BranchID:     filter.BranchID,
		Role:         filter.Role,
		Status:       filter.Status,
		SearchQuery:  filter.SearchQuery,
		Page:         filter.Page,
		Limit:        filter.Limit,
		SortBy:       filter.SortBy,
		SortOrder:    filter.SortOrder,
	}

	// Sanitize search input (Epic 2 retro: remove wildcard characters)
	if repoFilter.SearchQuery != "" {
		cleaned := sanitizeSearchInput(repoFilter.SearchQuery)
		repoFilter.SearchQuery = cleaned
	}

	// List users via repository
	users, total, err := s.userRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, 0, &ServiceError{Op: "list users", Err: err}
	}

	return users, total, nil
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(ctx context.Context, id uint) (*user.User, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate ID (Epic 2 retro: zero ID validation)
	if id == 0 {
		return nil, &InvalidInputError{Field: "id", Message: "user ID cannot be zero"}
	}

	// Get user via repository
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, &UserNotFoundError{UserID: id}
	}

	return u, nil
}
