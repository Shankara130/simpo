package models

import (
	"time"

	"gorm.io/gorm"
)

// PurchaseInvoice represents a purchase invoice from a supplier
// Story 10.2: Purchase Invoice Recording for tracking supplier purchases
type PurchaseInvoice struct {
	// ID is the unique identifier for the purchase invoice
	// Example: 1
	ID uint `gorm:"primaryKey" json:"id" example:"1"`

	// InvoiceNumber is the unique invoice number from supplier
	// Example: "INV-2026-001"
	InvoiceNumber string `gorm:"type:varchar(100);uniqueIndex;not null" json:"invoiceNumber" binding:"required,max=100" example:"INV-2026-001"`

	// InvoiceDate is the date on the invoice
	// Example: "2026-05-30"
	InvoiceDate time.Time `gorm:"type:date;not null" json:"invoiceDate" binding:"required" example:"2026-05-30"`

	// SupplierID is the foreign key to suppliers table
	// Example: 1
	SupplierID uint `gorm:"column:supplier_id;not null" json:"supplierId" binding:"required" example:"1"`

	// TotalAmount is the sum of all line item subtotals
	// Example: 1500000.00
	TotalAmount float64 `gorm:"type:decimal(15,2);not null;default:0" json:"totalAmount" example:"1500000.00"`

	// PaymentStatus indicates payment status: unpaid, partial, paid
	// Example: "unpaid"
	PaymentStatus string `gorm:"column:payment_status;type:varchar(20);not null;default:'unpaid'" json:"paymentStatus" example:"unpaid"`

	// ReceiptStatus indicates goods receipt status: pending, received, partial
	// Story 10.3: Track whether goods have been received for this invoice
	// Example: "pending"
	ReceiptStatus string `gorm:"column:receipt_status;type:varchar(20);not null;default:'pending'" json:"receiptStatus" example:"pending"`

	// GoodsReceiptID is the foreign key to goods_receipts table when invoice has been received
	// Story 10.3: Link to goods receipt record when goods are received
	// Example: 1
	GoodsReceiptID *uint `gorm:"column:goods_receipt_id" json:"goodsReceiptId,omitempty" example:"1"`

	// Notes are optional notes about the invoice
	// Example: "Emergency order for antibiotic stock"
	Notes string `gorm:"type:text" json:"notes,omitempty" example:"Emergency order for antibiotic stock"`

	// DocumentURL is optional link to invoice document image
	// Example: "https://storage.example.com/invoices/inv-2026-001.pdf"
	DocumentURL string `gorm:"column:document_url;type:varchar(255)" json:"documentUrl,omitempty" example:"https://storage.example.com/invoices/inv-2026-001.pdf"`

	// BranchID is the foreign key to branches table for multi-branch support
	// Example: 1
	BranchID uint `gorm:"column:branch_id;not null" json:"branchId" binding:"required" example:"1"`

	// CreatedBy is the user who created the invoice
	// Example: 1
	CreatedBy *uint `gorm:"column:created_by" json:"createdBy,omitempty"`

	// UpdatedBy is the user who last updated the invoice
	// Example: 1
	UpdatedBy *uint `gorm:"column:updated_by" json:"updatedBy,omitempty"`

	// Version is the optimistic locking version
	// Example: 1
	Version int `gorm:"column:version;not null;default:1" json:"version" example:"1"`

	// CreatedAt is the timestamp when the invoice was created
	// Example: "2026-05-30T10:00:00Z"
	CreatedAt time.Time `json:"createdAt" example:"2026-05-30T10:00:00Z"`

	// UpdatedAt is the timestamp when the invoice was last updated
	// Example: "2026-05-30T10:00:00Z"
	UpdatedAt time.Time `json:"updatedAt" example:"2026-05-30T10:00:00Z"`

	// DeletedAt is the soft delete timestamp (NULL for active invoices)
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships - eager loaded with Preload
	Supplier Supplier               `gorm:"foreignKey:SupplierID" json:"supplier"`
	Branch   Branch                 `gorm:"foreignKey:BranchID" json:"branch"`
	Items    []PurchaseInvoiceItem  `gorm:"foreignKey:PurchaseInvoiceID" json:"items"`
}

// TableName specifies the table name for PurchaseInvoice model
// Story 10.2: Table name follows snake_case plural convention
func (PurchaseInvoice) TableName() string {
	return "purchase_invoices"
}
