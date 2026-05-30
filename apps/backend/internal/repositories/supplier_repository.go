package repositories

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// SupplierRepository defines the interface for supplier data operations
// Story 10.1: Repository interface with CRUD methods for Supplier entity
type SupplierRepository interface {
	// Create inserts a new supplier into the database
	// createdBy is the user ID who is creating the supplier
	Create(ctx context.Context, supplier *models.Supplier, createdBy uint) error

	// GetByID retrieves a supplier by its ID
	// Returns ErrNotFound if supplier doesn't exist
	GetByID(ctx context.Context, id uint) (*models.Supplier, error)

	// GetByName retrieves a supplier by its name
	// Returns ErrNotFound if supplier doesn't exist
	GetByName(ctx context.Context, name string) (*models.Supplier, error)

	// Update modifies an existing supplier in the database
	// updatedBy is the user ID who is updating the supplier
	Update(ctx context.Context, supplier *models.Supplier, updatedBy uint) error

	// Deactivate soft deletes a supplier (sets deleted_at timestamp)
	// deactivatedBy is the user ID who is deactivating the supplier
	Deactivate(ctx context.Context, id uint, deactivatedBy uint) error

	// List retrieves suppliers with optional filtering and pagination
	// Returns slice of suppliers, total count, and error
	List(ctx context.Context, filter *SupplierFilter) ([]*models.Supplier, int64, error)
}

// SupplierFilter defines filtering options for supplier listing
// Story 10.1: Filter struct for supplier queries with pagination support
type SupplierFilter struct {
	SearchQuery string // Search by name, contact person, or phone
	IsActive     *bool // Filter by active status
	Page         int   // Page number (1-indexed)
	Limit        int   // Items per page
	SortBy       string // Field to sort by (name, created_at)
	SortOrder    string // "asc" or "desc"
}
