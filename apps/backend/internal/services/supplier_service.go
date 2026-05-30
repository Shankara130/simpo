package services

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// SupplierService defines the interface for supplier business logic operations
// Story 10.1: Service interface with validation and audit logging for Supplier entity
type SupplierService interface {
	// CreateSupplier creates a new supplier with validation and audit logging
	// Story 10.1, AC1: Validates required fields (name, phone) and logs creation
	CreateSupplier(ctx context.Context, supplier *models.Supplier, createdBy uint, ipAddress string) (*models.Supplier, error)

	// GetSupplierByID retrieves a supplier by ID
	// Story 10.1, AC1: Returns supplier details or error if not found
	GetSupplierByID(ctx context.Context, id uint) (*models.Supplier, error)

	// ListSuppliers retrieves suppliers with filtering and pagination
	// Story 10.1, AC2: Supports search, active status filter, and pagination
	ListSuppliers(ctx context.Context, filter *SupplierListFilter) ([]*models.Supplier, int64, error)

	// UpdateSupplier updates an existing supplier with validation and audit logging
	// Story 10.1, AC2: Validates changes and logs update with reason
	UpdateSupplier(ctx context.Context, id uint, updates *UpdateSupplierRequest, updatedBy uint, ipAddress string) (*models.Supplier, error)

	// DeactivateSupplier deactivates a supplier with audit logging
	// Story 10.1, AC3: Soft deletes supplier and logs deactivation with reason
	DeactivateSupplier(ctx context.Context, id uint, reason string, deactivatedBy uint, ipAddress string) error
}

// SupplierListFilter defines filtering options for supplier listing
// Story 10.1: Filter struct for supplier queries with pagination support
type SupplierListFilter struct {
	SearchQuery string // Search by name, contact person, or phone
	IsActive     *bool // Filter by active status
	Page         int   // Page number (1-indexed)
	Limit        int   // Items per page
	SortBy       string // Field to sort by (name, created_at)
	SortOrder    string // "asc" or "desc"
}

// UpdateSupplierRequest defines the fields that can be updated on a supplier
// Story 10.1: Request DTO for supplier updates
type UpdateSupplierRequest struct {
	Name          string `json:"name"`
	ContactPerson string `json:"contact_person"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Reason        string `json:"reason" binding:"required,min=5,max=500"` // Reason for update (required)
}
