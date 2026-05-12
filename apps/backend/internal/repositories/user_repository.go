package repositories

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// UserRepository defines the interface for user data operations
// AC1: Repository interface with CRUD methods for User entity
// References: internal/user/domain (from Epic 1)
type UserRepository interface {
	// Create inserts a new user into the database
	Create(ctx context.Context, user *user.User) error

	// GetByID retrieves a user by its ID
	// Returns ErrNotFound if user doesn't exist
	GetByID(ctx context.Context, id uint) (*user.User, error)

	// GetByUsername retrieves a user by their username
	// Returns ErrNotFound if user doesn't exist
	GetByUsername(ctx context.Context, username string) (*user.User, error)

	// GetByEmail retrieves a user by their email address
	// Returns ErrNotFound if user doesn't exist
	GetByEmail(ctx context.Context, email string) (*user.User, error)

	// Update modifies an existing user in the database
	Update(ctx context.Context, user *user.User) error

	// Delete removes a user from the database (soft delete)
	Delete(ctx context.Context, id uint) error

	// Deactivate marks a user as inactive
	Deactivate(ctx context.Context, id uint, reason string, deactivatedBy uint) error

	// List retrieves users with optional filtering and pagination
	// Returns slice of users, total count, and error
	List(ctx context.Context, filter *UserFilter) ([]*user.User, int64, error)

	// GetByBranch retrieves all users assigned to a specific branch
	GetByBranch(ctx context.Context, branchID uint) ([]*user.User, error)

	// ExistsByUsername checks if a username exists
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// ExistsByEmail checks if an email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// UserFilter defines filtering options for user listing
// AC4: Complex query support with filtering by role, branch, status
type UserFilter struct {
	BranchID  *uint   // Filter by branch assignment
	Role      string  // Filter by role (SYSTEM_ADMIN, OWNER, CASHIER)
	Status    string  // Filter by status (ACTIVE, INACTIVE, PENDING)
	SearchQuery string // Search by name, username, or email (ILIKE)
	Page      int     // Page number (1-indexed)
	Limit     int     // Items per page
	SortBy    string  // Field to sort by
	SortOrder string  // "asc" or "desc"
}

// UserSummary represents aggregated user data
type UserSummary struct {
	TotalUsers    int64 `json:"total_users"`
	ActiveUsers   int64 `json:"active_users"`
	InactiveUsers int64 `json:"inactive_users"`
	PendingUsers  int64 `json:"pending_users"` // Self-registered, awaiting approval
}
