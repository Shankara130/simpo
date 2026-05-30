package models

import (
	"time"

	"gorm.io/gorm"
)

// GoodsReceipt represents a goods receipt from a supplier
// Story 10.3: Goods Receipt Processing for tracking when purchased goods are received
type GoodsReceipt struct {
	// ID is the unique identifier for the goods receipt
	// Example: 1
	ID uint `gorm:"primaryKey" json:"id" example:"1"`

	// PurchaseInvoiceID is the foreign key to purchase_invoices table (one-to-one relationship)
	// Example: 1
	PurchaseInvoiceID uint `gorm:"column:purchase_invoice_id;not null;uniqueIndex:idx_goods_receipts_purchase_invoice_id" json:"purchaseInvoiceId" binding:"required" example:"1"`

	// ReceivedDate is the date when goods were received from the supplier
	// Example: "2026-05-30"
	ReceivedDate time.Time `gorm:"column:received_date;type:date;not null" json:"receivedDate" binding:"required" example:"2026-05-30"`

	// ReceivedBy is the user who processed the goods receipt
	// Example: 1
	ReceivedBy uint `gorm:"column:received_by;not null" json:"receivedBy" binding:"required" example:"1"`

	// Notes are optional notes about the goods receipt
	// Example: "All items received in good condition"
	Notes string `gorm:"type:text" json:"notes,omitempty" example:"All items received in good condition"`

	// BranchID is the foreign key to branches table for multi-branch support
	// Example: 1
	BranchID uint `gorm:"column:branch_id;not null" json:"branchId" binding:"required" example:"1"`

	// CreatedAt is the timestamp when the goods receipt was created
	// Example: "2026-05-30T10:00:00Z"
	CreatedAt time.Time `json:"createdAt" example:"2026-05-30T10:00:00Z"`

	// UpdatedAt is the timestamp when the goods receipt was last updated
	// Example: "2026-05-30T10:00:00Z"
	UpdatedAt time.Time `json:"updatedAt" example:"2026-05-30T10:00:00Z"`

	// Relationships - eager loaded with Preload
	PurchaseInvoice PurchaseInvoice `gorm:"foreignKey:PurchaseInvoiceID" json:"purchaseInvoice"`
	Branch         Branch           `gorm:"foreignKey:BranchID" json:"branch"`
}

// TableName specifies the table name for GoodsReceipt model
// Story 10.3: Table name follows snake_case plural convention
func (GoodsReceipt) TableName() string {
	return "goods_receipts"
}

// BeforeCreate is a GORM hook called before creating a goods receipt
// Story 10.3: Set default values and perform validations
func (gr *GoodsReceipt) BeforeCreate(tx *gorm.DB) error {
	// Set ReceivedDate to current date if not set
	if gr.ReceivedDate.IsZero() {
		gr.ReceivedDate = time.Now().UTC()
	}
	return nil
}
