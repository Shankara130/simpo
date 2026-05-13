package services

import (
	"context"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// ProductService defines the interface for product business operations
// AC1: Service interface for product domain with clear business method signatures
type ProductService interface {
	// CreateProduct creates a new product with business validation
	// Validates: SKU uniqueness within branch, required fields
	CreateProduct(ctx context.Context, product *models.Product) error

	// UpdateProduct modifies an existing product with business rules
	// Business rules: cannot update SKU, preserves created_at
	UpdateProduct(ctx context.Context, id uint, product *models.Product) error

	// UpdateStock updates product stock quantity atomically
	// Uses atomic increment to prevent race conditions (Epic 2 retro)
	UpdateStock(ctx context.Context, id uint, quantity int64) error

	// CheckAvailability checks if sufficient stock is available
	// Returns available quantity (min of stock_qty and requested_qty)
	CheckAvailability(ctx context.Context, id uint, requestedQty int64) (int64, error)

	// ListProducts retrieves products with filtering and pagination
	// Delegates to repository with security considerations
	ListProducts(ctx context.Context, filter *ProductFilter) ([]*models.Product, int64, error)

	// GetProductByID retrieves a product by ID with relationships
	GetProductByID(ctx context.Context, id uint) (*models.Product, error)

	// GetLowStockProducts retrieves products with stock below reorder threshold
	GetLowStockProducts(ctx context.Context, branchID uint) ([]*models.Product, error)
}

// ProductFilter defines filtering options for product listing
type ProductFilter struct {
	BranchID     *uint
	Category     string
	SearchQuery  string
	LowStock     bool
	Expired      bool
	ExpiryBefore *time.Time
	Page         int
	Limit        int
	SortBy       string
	SortOrder    string
}
