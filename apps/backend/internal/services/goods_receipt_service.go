package services

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// GoodsReceiptService defines the interface for goods receipt business operations
// Story 10.3: Service interface with business logic methods for goods receipt processing
type GoodsReceiptService interface {
	// ProcessGoodsReceipt processes a goods receipt for a purchase invoice
	// Story 10.3: Main business logic method that:
	// - Validates invoice can be received (exists, not already received, has items)
	// - Creates goods receipt record
	// - Updates stock quantities for all invoice items
	// - Updates cost prices to latest purchase cost
	// - Updates invoice receipt_status to "received"
	// - Logs audit trail entries
	// - Triggers low stock alerts
	// - Publishes stock update events
	// All wrapped in a database transaction for atomicity
	ProcessGoodsReceipt(ctx context.Context, invoiceID uint, receivedBy uint, notes string, branchID uint) (*models.GoodsReceipt, error)

	// GetByID retrieves a goods receipt by its ID
	// Story 10.3: Get goods receipt with full details
	GetByID(ctx context.Context, id uint) (*models.GoodsReceipt, error)

	// List retrieves goods receipts with optional filtering and pagination
	// Story 10.3: List goods receipts with filters
	List(ctx context.Context, filter *GoodsReceiptFilter) ([]*models.GoodsReceipt, int64, error)
}

// GoodsReceiptFilter defines filtering options for goods receipt listing
// Story 10.3: Filter struct for goods receipt queries
type GoodsReceiptFilter struct {
	BranchID   *uint   // Filter by branch
	StartDate  *string // Filter by received date range start (inclusive)
	EndDate    *string // Filter by received date range end (inclusive)
	ReceivedBy *uint   // Filter by user who processed the receipt
	Page       int     // Page number (1-indexed)
	Limit      int     // Items per page
	SortBy     string  // Field to sort by (received_date, created_at)
	SortOrder  string  // "asc" or "desc"
}
