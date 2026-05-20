package dto

import "time"

// StockAdjustmentRequest represents a request to manually adjust stock quantity
// Story 4.3, AC1, AC2, AC3: Manual stock adjustment with reason logging
type StockAdjustmentRequest struct {
	ProductID  uint   `json:"productId" binding:"required,min=1"`           // AC1: Product selection required
	BranchID   uint   `json:"branchId" binding:"required,min=1"`            // AC1: Branch location required
	NewStockQty int64  `json:"newStockQty" binding:"required,min=0"`             // AC2: New quantity (not increment)
	Reason     string `json:"reason" binding:"required,min=1"`                   // AC3: Reason required
	ReasonNotes string `json:"reasonNotes,omitempty"`                            // AC3: Optional details for "Other" reason
}

// StockAdjustmentResult represents the result of a successful stock adjustment
// Story 4.3, AC7: Success confirmation with old/new/changed values
type StockAdjustmentResult struct {
	ProductID   uint      `json:"productId"`
	SKU          string    `json:"sku"`
	Name         string    `json:"name"`
	OldStockQty  int64     `json:"oldStockQty"`
	NewStockQty  int64     `json:"newStockQty"`
	Change       int64     `json:"change"`       // Calculated delta (new - old)
	Reason       string    `json:"reason"`
	AdjustedBy   string    `json:"adjustedBy"`
	AdjustedAt   time.Time `json:"adjustedAt"`
}

// StockAdjustmentReason defines the allowed reasons for stock adjustment
// Story 4.3, AC3: Predefined reason categories for audit trail
type StockAdjustmentReason string

const (
	// StockAdjustmentDamage indicates products were damaged and removed from stock
	StockAdjustmentDamage StockAdjustmentReason = "Damage"
	// StockAdjustmentExpiration indicates products expired and were disposed
	StockAdjustmentExpiration StockAdjustmentReason = "Expiration"
	// StockAdjustmentDeliveryReceipt indicates new stock received from supplier
	StockAdjustmentDeliveryReceipt StockAdjustmentReason = "DeliveryReceipt"
	// StockAdjustmentPhysicalCount indicates adjustment based on physical inventory count
	StockAdjustmentPhysicalCount StockAdjustmentReason = "PhysicalCount"
	// StockAdjustmentTheftLoss indicates stock lost due to theft or loss
	StockAdjustmentTheftLoss StockAdjustmentReason = "TheftLoss"
	// StockAdjustmentOther indicates other reasons with additional notes
	StockAdjustmentOther StockAdjustmentReason = "Other"
)

// ValidStockAdjustmentReasons returns a list of valid stock adjustment reasons
// Useful for frontend dropdown population and validation
func ValidStockAdjustmentReasons() []StockAdjustmentReason {
	return []StockAdjustmentReason{
		StockAdjustmentDamage,
		StockAdjustmentExpiration,
		StockAdjustmentDeliveryReceipt,
		StockAdjustmentPhysicalCount,
		StockAdjustmentTheftLoss,
		StockAdjustmentOther,
	}
}

// LowStockNotificationEvent represents a low stock notification event
// Story 4.4, AC2, AC4: Event structure for low stock notifications
type LowStockNotificationEvent struct {
	EventID   string              `json:"eventId"`   // UUID for event tracking
	EventType string              `json:"eventType"` // "stock.low"
	Timestamp string              `json:"timestamp"` // ISO 8601 timestamp
	Data      ProductLowStockData `json:"data"`      // Low stock details
}

// ProductLowStockData contains product information for low stock notifications
// Story 4.4, AC4: Product details for notification payload
type ProductLowStockData struct {
	ProductID         uint   `json:"productId"`
	SKU               string `json:"sku"`
	ProductName       string `json:"productName"`
	CurrentStock      int    `json:"currentStock"`
	ReorderThreshold  int    `json:"reorderThreshold"`
	SuggestedOrderQty int    `json:"suggestedOrderQty"` // threshold - current + buffer
	BranchID          uint   `json:"branchId"`
	BranchName        string `json:"branchName"`
}

// LowStockProductResponse represents a low stock product in API responses
// Story 4.4, Task 5: API endpoint response structure
type LowStockProductResponse struct {
	ProductID         uint   `json:"productId"`
	SKU                string `json:"sku"`
	Name               string `json:"name"`
	CurrentStock       int64  `json:"currentStock"`
	ReorderThreshold   int    `json:"reorderThreshold"`
	SuggestedOrderQty  int    `json:"suggestedOrderQty"`
	BranchID           uint   `json:"branchId"`
	BranchName         string `json:"branchName"`
}

// ProductListRequest represents query parameters for product listing
// Story 4.1, AC2, AC3, AC7: Search, filter, and pagination parameters
type ProductListRequest struct {
	Search     string `form:"search"`      // Search by name or SKU
	Category   string `form:"category"`    // Filter by category
	BranchID   *uint  `form:"branch_id"`   // Filter by branch (Owner only)
	LowStock   *bool  `form:"low_stock"`   // Filter for low stock items
	Expired    *bool  `form:"expired"`     // Filter for expired items
	Page       int    `form:"page"`        // Page number (default 1)
	Limit      int    `form:"limit"`       // Items per page (default 20, max 1000)
	SortBy     string `form:"sort_by"`     // Field to sort by
	SortOrder  string `form:"sort_order"`  // "asc" or "desc"
}

// ProductListItem represents a product in the list view
// Story 4.1, AC4: Display fields for product list
type ProductListItem struct {
	ID               uint       `json:"id"`
	SKU              string     `json:"sku"`
	Name             string     `json:"name"`
	Description      string     `json:"description,omitempty"`
	StockQty         int64      `json:"stockQty"`
	Price            string     `json:"price"`
	ExpiryDate       *time.Time `json:"expiryDate,omitempty"`
	BranchID         uint       `json:"branchId"`
	Category         string     `json:"category,omitempty"`
	ReorderThreshold int        `json:"reorderThreshold"`
	IsLowStock       bool       `json:"isLowStock"`       // Story 4.1, AC5: Low stock indicator
	IsExpired        bool       `json:"isExpired"`        // Story 4.1, AC6: Expired indicator
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// ProductListResponse represents the paginated product list response
// Story 4.1, AC7: Pagination support for large catalogs
type ProductListResponse struct {
	Data       []ProductListItem    `json:"data"`
	Pagination PaginationMetadata   `json:"pagination"`
}

// PaginationMetadata represents pagination information
type PaginationMetadata struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int  `json:"totalPages"`
}
