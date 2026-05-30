package models

import (
	"time"
)

// PurchaseInvoiceItem represents a line item in a purchase invoice
// Story 10.2: Line items with product, quantity, unit cost, subtotal
type PurchaseInvoiceItem struct {
	// ID is the unique identifier for the line item
	// Example: 1
	ID uint `gorm:"primaryKey" json:"id" example:"1"`

	// PurchaseInvoiceID is the foreign key to purchase_invoices table
	// Example: 1
	PurchaseInvoiceID uint `gorm:"column:purchase_invoice_id;not null" json:"purchaseInvoiceId" example:"1"`

	// ProductID is the foreign key to products table
	// Example: 5
	ProductID uint `gorm:"column:product_id;not null" json:"productId" example:"5"`

	// Quantity is the quantity purchased
	// Example: 100
	Quantity int `gorm:"not null" json:"quantity" binding:"required,min=1" example:"100"`

	// UnitCost is the cost per unit
	// Example: 15000.00
	UnitCost float64 `gorm:"type:decimal(15,2);not null" json:"unitCost" binding:"required,min=0" example:"15000.00"`

	// Subtotal is the line item total (quantity * unit_cost)
	// Example: 1500000.00
	Subtotal float64 `gorm:"type:decimal(15,2);not null" json:"subtotal" example:"1500000.00"`

	// CreatedAt is the timestamp when the line item was created
	// Example: "2026-05-30T10:00:00Z"
	CreatedAt time.Time `json:"createdAt" example:"2026-05-30T10:00:00Z"`

	// UpdatedAt is the timestamp when the line item was last updated
	// Example: "2026-05-30T10:00:00Z"
	UpdatedAt time.Time `json:"updatedAt" example:"2026-05-30T10:00:00Z"`

	// Relationships - eager loaded with Preload
	Product Product `gorm:"foreignKey:ProductID" json:"product"`
}

// TableName specifies the table name for PurchaseInvoiceItem model
// Story 10.2: Table name follows snake_case plural convention
func (PurchaseInvoiceItem) TableName() string {
	return "purchase_invoice_items"
}
