package dto

// ProcessGoodsReceiptRequest represents a request to process goods receipt for an invoice
// Story 10.3: Request DTO for processing goods receipt
type ProcessGoodsReceiptRequest struct {
	// InvoiceID is the purchase invoice ID to process goods receipt for
	// Example: 1
	InvoiceID uint `json:"invoiceId" binding:"required,min=1" example:"1"`

	// Notes are optional notes about the goods receipt
	// Example: "All items received in good condition"
	Notes string `json:"notes,omitempty" example:"All items received in good condition"`
}

// GoodsReceiptResponse represents a goods receipt response
// Story 10.3: Response DTO for goods receipt operations
type GoodsReceiptResponse struct {
	// ID is the unique identifier for the goods receipt
	// Example: 1
	ID uint `json:"id" example:"1"`

	// PurchaseInvoiceID is the purchase invoice ID
	// Example: 1
	PurchaseInvoiceID uint `json:"purchaseInvoiceId" example:"1"`

	// ReceivedDate is the date when goods were received
	// Example: "2026-05-30"
	ReceivedDate string `json:"receivedDate" example:"2026-05-30"`

	// ReceivedBy is the user who processed the goods receipt
	// Example: 1
	ReceivedBy uint `json:"receivedBy" example:"1"`

	// Notes are notes about the goods receipt
	// Example: "All items received in good condition"
	Notes string `json:"notes,omitempty" example:"All items received in good condition"`

	// BranchID is the branch where goods were received
	// Example: 1
	BranchID uint `json:"branchId" example:"1"`

	// CreatedAt is the timestamp when the goods receipt was created
	// Example: "2026-05-30T10:00:00Z"
	CreatedAt string `json:"createdAt" example:"2026-05-30T10:00:00Z"`

	// UpdatedAt is the timestamp when the goods receipt was last updated
	// Example: "2026-05-30T10:00:00Z"
	UpdatedAt string `json:"updatedAt" example:"2026-05-30T10:00:00Z"`

	// PurchaseInvoice contains the associated purchase invoice details
	// Story 10.3: Include invoice details in response
	PurchaseInvoice *PurchaseInvoiceSummary `json:"purchaseInvoice,omitempty"`
}

// PurchaseInvoiceSummary represents a summary of purchase invoice in goods receipt response
// Story 10.3: Simplified invoice details for goods receipt response
type PurchaseInvoiceSummary struct {
	ID             uint                   `json:"id" example:"1"`
	InvoiceNumber  string                 `json:"invoiceNumber" example:"INV-2026-001"`
	InvoiceDate    string                 `json:"invoiceDate" example:"2026-05-30"`
	SupplierID     uint                   `json:"supplierId" example:"1"`
	SupplierName   string                 `json:"supplierName" example:"PT. Pharmasi Jaya"`
	TotalAmount    float64                `json:"totalAmount" example:"1500000.00"`
	PaymentStatus  string                 `json:"paymentStatus" example:"unpaid"`
	ReceiptStatus  string                 `json:"receiptStatus" example:"received"`
	Items          []PurchaseInvoiceItemSummary `json:"items"`
}

// PurchaseInvoiceItemSummary represents a line item summary in goods receipt response
// Story 10.3: Simplified line item details for goods receipt response
type PurchaseInvoiceItemSummary struct {
	ID        uint    `json:"id" example:"1"`
	ProductID uint    `json:"productId" example:"5"`
	ProductName string `json:"productName" example:"Amoxicillin 500mg"`
	SKU       string  `json:"sku" example:"AMOX-500"`
	Quantity  int     `json:"quantity" example:"100"`
	UnitCost  float64 `json:"unitCost" example:"15000.00"`
	Subtotal  float64 `json:"subtotal" example:"1500000.00"`
}

// GoodsReceiptListResponse represents a paginated list of goods receipts
// Story 10.3: Response DTO for goods receipt listing with pagination
type GoodsReceiptListResponse struct {
	// Data is the list of goods receipts
	Data []*GoodsReceiptResponse `json:"data"`

	// Meta contains pagination metadata
	Meta PaginationMeta `json:"meta"`
}

// GoodsReceiptFilter represents filter parameters for goods receipt listing
// Story 10.3: Filter parameters for querying goods receipts
type GoodsReceiptFilter struct {
	// BranchID filters by branch (optional)
	// Example: 1
	BranchID *uint `form:"branchId,omitempty" example:"1"`

	// StartDate filters by received date range start (inclusive, optional)
	// Example: "2026-05-01"
	StartDate *string `form:"startDate,omitempty" example:"2026-05-01"`

	// EndDate filters by received date range end (inclusive, optional)
	// Example: "2026-05-31"
	EndDate *string `form:"endDate,omitempty" example:"2026-05-31"`

	// ReceivedBy filters by user who processed the receipt (optional)
	// Example: 1
	ReceivedBy *uint `form:"receivedBy,omitempty" example:"1"`

	// Page is the page number (1-indexed, optional)
	// Example: 1
	Page int `form:"page,omitempty" example:"1"`

	// Limit is the number of items per page (optional)
	// Example: 20
	Limit int `form:"limit,omitempty" example:"20"`

	// SortBy is the field to sort by (optional)
	// Example: "received_date"
	SortBy string `form:"sortBy,omitempty" example:"received_date"`

	// SortOrder is the sort order (asc or desc, optional)
	// Example: "desc"
	SortOrder string `form:"sortOrder,omitempty" example:"desc"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	// Page is the current page number
	Page int `json:"page" example:"1"`

	// Limit is the number of items per page
	Limit int `json:"limit" example:"20"`

	// Total is the total number of items
	Total int64 `json:"total" example:"100"`

	// TotalPages is the total number of pages
	TotalPages int `json:"totalPages" example:"5"`
}
