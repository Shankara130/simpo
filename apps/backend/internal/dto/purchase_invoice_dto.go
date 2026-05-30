package dto

import "time"

// PurchaseInvoice DTOs for API request/response
// Story 10.2: Data transfer objects for purchase invoice management endpoints

// CreatePurchaseInvoiceRequest represents the request payload for creating a new purchase invoice
// Story 10.2, AC1: Validation tags for required fields
type CreatePurchaseInvoiceRequest struct {
	// InvoiceNumber is the unique invoice identifier (required, max 100 characters)
	// Example: "INV-2023-001"
	InvoiceNumber string `json:"invoiceNumber" binding:"required,max=100" example:"INV-2023-001"`

	// InvoiceDate is the date of the invoice (required, ISO 8601 format)
	// Example: "2023-05-30"
	InvoiceDate string `json:"invoiceDate" binding:"required" example:"2023-05-30"`

	// SupplierID is the ID of the supplier (required)
	// Example: 1
	SupplierID uint `json:"supplierId" binding:"required" example:"1"`

	// BranchID is the ID of the branch (required)
	// Example: 1
	BranchID uint `json:"branchId" binding:"required" example:"1"`

	// Notes is any additional notes about the invoice
	// Example: "Monthly supply order"
	Notes string `json:"notes,omitempty" binding:"omitempty,max=1000" example:"Monthly supply order"`

	// DocumentURL is the URL to the invoice document image
	// Example: "https://storage.example.com/invoices/inv-2023-001.pdf"
	DocumentURL string `json:"documentUrl,omitempty" binding:"omitempty,max=255" example:"https://storage.example.com/invoices/inv-2023-001.pdf"`

	// Items are the line items in the invoice (at least one required)
	Items []CreatePurchaseInvoiceItemRequest `json:"items" binding:"required,min=1"`
}

// CreatePurchaseInvoiceItemRequest represents a line item in a purchase invoice
// Story 10.2, AC1: Line item with product, quantity, unit cost
type CreatePurchaseInvoiceItemRequest struct {
	// ProductID is the ID of the product (required)
	// Example: 10
	ProductID uint `json:"productId" binding:"required" example:"10"`

	// Quantity is the quantity ordered (required, must be positive)
	// Example: 100
	Quantity int `json:"quantity" binding:"required,min=1" example:"100"`

	// UnitCost is the cost per unit (required, non-negative)
	// Example: 15000.00
	UnitCost float64 `json:"unitCost" binding:"required,min=0" example:"15000.00"`
}

// UpdatePurchaseInvoiceRequest represents the request payload for updating an existing purchase invoice
// Story 10.2: All fields optional except reason, which is required for audit trail
type UpdatePurchaseInvoiceRequest struct {
	// InvoiceNumber is the new invoice number
	// Example: "INV-2023-001-UPDATED"
	InvoiceNumber string `json:"invoiceNumber" binding:"omitempty,max=100" example:"INV-2023-001-UPDATED"`

	// InvoiceDate is the new invoice date (ISO 8601 format)
	// Example: "2023-05-31"
	InvoiceDate string `json:"invoiceDate" binding:"omitempty" example:"2023-05-31"`

	// SupplierID is the new supplier ID
	// Example: 2
	SupplierID uint `json:"supplierId" binding:"omitempty" example:"2"`

	// Notes is the new notes
	// Example: "Updated notes for accuracy"
	Notes string `json:"notes,omitempty" binding:"omitempty,max=1000" example:"Updated notes for accuracy"`

	// DocumentURL is the new document URL
	// Example: "https://storage.example.com/invoices/inv-2023-001-updated.pdf"
	DocumentURL string `json:"documentUrl,omitempty" binding:"omitempty,max=255" example:"https://storage.example.com/invoices/inv-2023-001-updated.pdf"`

	// Items are the updated line items (replaces all existing items)
	Items []CreatePurchaseInvoiceItemRequest `json:"items" binding:"required,min=1"`

	// Reason is the reason for the update (required for audit trail)
	// Example: "Correcting invoice amount"
	Reason string `json:"reason" binding:"required,min=5,max=500" example:"Correcting invoice amount"`
}

// PurchaseInvoiceResponse represents the response payload for purchase invoice operations
// Story 10.2, AC3: Response DTO with complete invoice information
type PurchaseInvoiceResponse struct {
	// ID is the unique identifier for the invoice
	// Example: 1
	ID uint `json:"id" example:"1"`

	// InvoiceNumber is the invoice number
	// Example: "INV-2023-001"
	InvoiceNumber string `json:"invoiceNumber" example:"INV-2023-001"`

	// InvoiceDate is the date of the invoice (YYYY-MM-DD format)
	// Example: "2023-05-30"
	InvoiceDate string `json:"invoiceDate" example:"2023-05-30"`

	// SupplierID is the ID of the supplier
	// Example: 1
	SupplierID uint `json:"supplierId" example:"1"`

	// SupplierName is the name of the supplier (from relationship)
	// Example: "PT. Pharmasi Jaya"
	SupplierName string `json:"supplierName,omitempty" example:"PT. Pharmasi Jaya"`

	// SupplierContactPerson is the contact person at the supplier (from relationship)
	// Example: "John Doe"
	SupplierContactPerson string `json:"supplierContactPerson,omitempty" example:"John Doe"`

	// SupplierPhone is the phone number of the supplier (from relationship)
	// Example: "+62-21-555-1234"
	SupplierPhone string `json:"supplierPhone,omitempty" example:"+62-21-555-1234"`

	// SupplierEmail is the email of the supplier (from relationship)
	// Example: "orders@pharmasi-jaya.co.id"
	SupplierEmail string `json:"supplierEmail,omitempty" example:"orders@pharmasi-jaya.co.id"`

	// SupplierAddress is the address of the supplier (from relationship)
	// Example: "Jl. Industri No. 123, Jakarta"
	SupplierAddress string `json:"supplierAddress,omitempty" example:"Jl. Industri No. 123, Jakarta"`

	// BranchID is the ID of the branch
	// Example: 1
	BranchID uint `json:"branchId" example:"1"`

	// BranchName is the name of the branch (from relationship)
	// Example: "Jakarta Pusat"
	BranchName string `json:"branchName,omitempty" example:"Jakarta Pusat"`

	// TotalAmount is the total invoice amount
	// Example: 1500000.00
	TotalAmount float64 `json:"totalAmount" example:"1500000.00"`

	// PaymentStatus is the payment status
	// Example: "unpaid"
	PaymentStatus string `json:"paymentStatus" example:"unpaid"`

	// Notes is any additional notes
	// Example: "Monthly supply order"
	Notes string `json:"notes,omitempty" example:"Monthly supply order"`

	// DocumentURL is the URL to the invoice document
	// Example: "https://storage.example.com/invoices/inv-2023-001.pdf"
	DocumentURL string `json:"documentUrl,omitempty" example:"https://storage.example.com/invoices/inv-2023-001.pdf"`

	// Items are the line items in the invoice
	Items []PurchaseInvoiceItemResponse `json:"items"`

	// CreatedAt is the timestamp when the invoice was created
	// Example: "2023-05-30T10:00:00Z"
	CreatedAt string `json:"createdAt" example:"2023-05-30T10:00:00Z"`

	// UpdatedAt is the timestamp when the invoice was last updated
	// Example: "2023-05-30T10:00:00Z"
	UpdatedAt string `json:"updatedAt" example:"2023-05-30T10:00:00Z"`
}

// PurchaseInvoiceItemResponse represents a line item in the purchase invoice response
// Story 10.2, AC3: Line item response with product details
type PurchaseInvoiceItemResponse struct {
	// ID is the unique identifier for the line item
	// Example: 1
	ID uint `json:"id" example:"1"`

	// ProductID is the ID of the product
	// Example: 10
	ProductID uint `json:"productId" example:"10"`

	// ProductName is the name of the product (from relationship)
	// Example: "Paracetamol 500mg"
	ProductName string `json:"productName,omitempty" example:"Paracetamol 500mg"`

	// ProductSKU is the SKU of the product (from relationship)
	// Example: "PAR-500-100"
	ProductSKU string `json:"productSku,omitempty" example:"PAR-500-100"`

	// Quantity is the quantity ordered
	// Example: 100
	Quantity int `json:"quantity" example:"100"`

	// UnitCost is the cost per unit
	// Example: 15000.00
	UnitCost float64 `json:"unitCost" example:"15000.00"`

	// Subtotal is the subtotal for this line item (quantity * unitCost)
	// Example: 1500000.00
	Subtotal float64 `json:"subtotal" example:"1500000.00"`
}

// PurchaseInvoiceListResponse represents the paginated response for purchase invoice listing
// Story 10.2, AC2: Response DTO with pagination metadata
type PurchaseInvoiceListResponse struct {
	// Data is the list of purchase invoices
	Data []PurchaseInvoiceResponse `json:"data"`

	// Pagination contains pagination metadata
	Pagination PaginationResponse `json:"pagination"`
}

// PurchaseInvoiceListFilter represents the query parameters for filtering purchase invoices
// Story 10.2, AC2: Filter parameters for list endpoint
type PurchaseInvoiceListFilter struct {
	// SupplierID filters by supplier
	// Example: 1
	SupplierID *uint `form:"supplierId" example:"1"`

	// StartDate filters invoices on or after this date (inclusive, ISO 8601)
	// Example: "2023-05-01"
	StartDate *string `form:"startDate" example:"2023-05-01"`

	// EndDate filters invoices on or before this date (inclusive, ISO 8601)
	// Example: "2023-05-31"
	EndDate *string `form:"endDate" example:"2023-05-31"`

	// PaymentStatus filters by payment status
	// Example: "unpaid"
	PaymentStatus *string `form:"paymentStatus" example:"unpaid"`

	// Search searches by invoice number (contains)
	// Example: "INV-2023"
	Search string `form:"search" example:"INV-2023"`

	// Page is the page number (1-indexed, default 1)
	// Example: 1
	Page int `form:"page,default=1" example:"1"`

	// Limit is the number of items per page (default 20, max 1000)
	// Example: 20
	Limit int `form:"limit,default=20" example:"20"`

	// SortBy is the field to sort by (default "invoice_date")
	// Example: "invoice_date"
	SortBy string `form:"sortBy,default=invoice_date" example:"invoice_date"`

	// SortOrder is the sort direction ("asc" or "desc", default "desc")
	// Example: "desc"
	SortOrder string `form:"sortOrder,default=desc" example:"desc"`
}

// DeactivatePurchaseInvoiceRequest represents the request payload for deleting (soft-deleting) a purchase invoice
// Story 10.2: Reason required for audit trail
type DeactivatePurchaseInvoiceRequest struct {
	// Reason is the reason for deletion (required for audit trail)
	// Example: "Duplicate invoice entered by mistake"
	Reason string `json:"reason" binding:"required,min=5,max=500" example:"Duplicate invoice entered by mistake"`
}

// Helper function to parse date string
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
