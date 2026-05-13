package services

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// UserService defines the interface for user business operations
// AC1: Service interface for user domain with clear business method signatures
type UserService interface {
	// CreateUser creates a new user with business validation
	// Validates: username uniqueness, email uniqueness, required fields
	CreateUser(ctx context.Context, user *user.User, ipAddress string) error

	// UpdateUser modifies an existing user with business rules
	// Business rules: role changes require admin role, cannot change own role
	UpdateUser(ctx context.Context, userID uint, user *user.User, adminID uint, ipAddress string) error

	// DeactivateUser deactivates a user account with audit logging
	DeactivateUser(ctx context.Context, userID uint, reason string, deactivatedBy uint, ipAddress string) error

	// ListUsers retrieves users with filtering and pagination
	// Filters: role, branch, status
	ListUsers(ctx context.Context, filter *UserFilter) ([]*user.User, int64, error)

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id uint) (*user.User, error)
}

// UserFilter defines filtering options for user listing
type UserFilter struct {
	BranchID    *uint
	Role        string
	Status      string
	SearchQuery string
	Page        int
	Limit       int
	SortBy      string
	SortOrder   string
}
