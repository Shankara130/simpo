package services

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// PurchaseInvoiceService defines the interface for purchase invoice business logic operations
// Story 10.2: Service interface with validation and audit logging for PurchaseInvoice entity
type PurchaseInvoiceService interface {
	// CreatePurchaseInvoice creates a new purchase invoice with validation and audit logging
	// Story 10.2, AC1: Validates invoice data, calculates total, logs creation
	CreatePurchaseInvoice(ctx context.Context, invoice *models.PurchaseInvoice, items []CreatePurchaseInvoiceItemRequest, createdBy uint, ipAddress string) (*models.PurchaseInvoice, error)

	// GetPurchaseInvoiceByID retrieves a purchase invoice by ID
	// Story 10.2, AC3: Returns invoice details with line items and supplier
	GetPurchaseInvoiceByID(ctx context.Context, id uint) (*models.PurchaseInvoice, error)

	// ListPurchaseInvoices retrieves purchase invoices with filtering and pagination
	// Story 10.2, AC2: Supports filtering by supplier, date range, payment status
	ListPurchaseInvoices(ctx context.Context, filter *PurchaseInvoiceListFilter) ([]*models.PurchaseInvoice, int64, error)

	// UpdatePurchaseInvoice updates an existing purchase invoice with validation and audit logging
	// Story 10.2: Validates changes, recalculates total, logs update
	UpdatePurchaseInvoice(ctx context.Context, id uint, updates *UpdatePurchaseInvoiceRequest, updatedBy uint, ipAddress string) (*models.PurchaseInvoice, error)

	// DeletePurchaseInvoice deletes (soft deletes) a purchase invoice with audit logging
	// Story 10.2: Soft deletes invoice and logs deletion
	DeletePurchaseInvoice(ctx context.Context, id uint, deletedBy uint, ipAddress string) error
}

// PurchaseInvoiceListFilter defines filtering options for purchase invoice listing
// Story 10.2: Filter struct for purchase invoice queries with pagination support
type PurchaseInvoiceListFilter struct {
	SupplierID    *uint   // Filter by supplier
	StartDate      *string // Filter by invoice date range start (inclusive)
	EndDate        *string // Filter by invoice date range end (inclusive)
	PaymentStatus *string // Filter by payment status
	SearchQuery    string  // Search by invoice number
	Page           int     // Page number (1-indexed)
	Limit          int     // Items per page
	SortBy         string  // Field to sort by (invoice_date, total_amount)
	SortOrder      string  // "asc" or "desc"
}

// CreatePurchaseInvoiceItemRequest defines a line item in a purchase invoice
// Story 10.2, AC1: Line item with product, quantity, unit cost
type CreatePurchaseInvoiceItemRequest struct {
	ProductID uint    `json:"productId" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	UnitCost  float64 `json:"unitCost" binding:"required,min=0"`
}

// UpdatePurchaseInvoiceRequest defines the fields that can be updated on a purchase invoice
// Story 10.2: Request DTO for purchase invoice updates
type UpdatePurchaseInvoiceRequest struct {
	InvoiceNumber    string                        `json:"invoiceNumber" binding:"required,max=100"`
	InvoiceDate      string                        `json:"invoiceDate" binding:"required"`
	SupplierID       uint                          `json:"supplierId" binding:"required"`
	Notes            string                        `json:"notes" binding:"omitempty,max=1000"`
	DocumentURL      string                        `json:"documentUrl" binding:"omitempty,max=255"`
	Items            []CreatePurchaseInvoiceItemRequest `json:"items" binding:"required,min=1"`
	Reason           string                        `json:"reason" binding:"required,min=5,max=500"` // Reason for update (required)
}
