package services

import (
	"context"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// StockAdjustmentRequest represents a request to manually adjust stock quantity
// Story 4.3, AC1, AC2, AC3, AC4: Manual stock adjustment with reason logging and audit trail
type StockAdjustmentRequest struct {
	ProductID   uint
	BranchID    uint
	NewStockQty int64
	Reason      string
	ReasonNotes string
}

// StockAdjustmentResult represents the result of a successful stock adjustment
// Story 4.3, AC7: Success confirmation with old/new/changed values
type StockAdjustmentResult struct {
	ProductID  uint
	SKU        string
	Name       string
	OldStockQty int64
	NewStockQty int64
	Change     int64
	Reason     string
	AdjustedBy string
	AdjustedAt time.Time
}

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

	// ManualAdjustStock manually adjusts stock quantity with reason logging
	// Story 4.3, AC1-AC7: Admin-only stock adjustment with audit trail compliance
	// Validates admin permissions, product existence, branch ownership
	// Logs adjustment in append-only audit trail, triggers low stock notifications
	ManualAdjustStock(ctx context.Context, req *StockAdjustmentRequest, adminID uint, adminUsername string) (*StockAdjustmentResult, error)

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
