package repositories

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// BranchRepository defines the interface for branch data operations
// AC1: Repository interface with CRUD methods for Branch entity
type BranchRepository interface {
	// Create inserts a new branch into the database
	Create(ctx context.Context, branch *models.Branch) error

	// GetByID retrieves a branch by its ID
	// Returns ErrNotFound if branch doesn't exist
	GetByID(ctx context.Context, id uint) (*models.Branch, error)

	// GetByName retrieves a branch by its name
	// Returns ErrNotFound if branch doesn't exist
	GetByName(ctx context.Context, name string) (*models.Branch, error)

	// Update modifies an existing branch in the database
	Update(ctx context.Context, branch *models.Branch) error

	// Delete removes a branch from the database (soft delete)
	Delete(ctx context.Context, id uint) error

	// List retrieves branches with optional filtering and pagination
	// Returns slice of branches, total count, and error
	List(ctx context.Context, filter *BranchFilter) ([]*models.Branch, int64, error)
}

// BranchFilter defines filtering options for branch listing
type BranchFilter struct {
	SearchQuery string // Search by name or address
	Page        int    // Page number (1-indexed)
	Limit       int    // Items per page
	SortBy      string // Field to sort by
	SortOrder   string // "asc" or "desc"
}
