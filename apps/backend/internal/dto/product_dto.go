package dto

import "time"

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
