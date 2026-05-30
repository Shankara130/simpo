package models

import (
	"time"

	"gorm.io/gorm"
)

// SupplierPayment represents a payment made to a supplier for a purchase invoice
// Story 10.4: Supplier Payment Tracking for managing cash flow and payment deadlines
type SupplierPayment struct {
	// ID is the unique identifier for the supplier payment
	// Example: 1
	ID uint `gorm:"primaryKey" json:"id" example:"1"`

	// PurchaseInvoiceID is the foreign key to purchase_invoices table
	// Example: 1
	PurchaseInvoiceID uint `gorm:"column:purchase_invoice_id;not null;index:idx_supplier_payments_invoice_id" json:"purchaseInvoiceId" binding:"required" example:"1"`

	// PaymentDate is the date when the payment was made
	// Example: "2026-05-31"
	PaymentDate time.Time `gorm:"column:payment_date;type:date;not null;index:idx_supplier_payments_payment_date" json:"paymentDate" binding:"required" example:"2026-05-31"`

	// PaymentAmount is the amount paid to the supplier
	// Example: 1500000.00
	PaymentAmount float64 `gorm:"type:decimal(15,2);not null" json:"paymentAmount" binding:"required,gt=0" example:"1500000.00"`

	// PaymentMethod indicates how the payment was made: cash, transfer, e-wallet, check, other
	// Example: "transfer"
	PaymentMethod string `gorm:"column:payment_method;type:varchar(50);not null" json:"paymentMethod" binding:"required,oneof=cash transfer e-wallet check other" example:"transfer"`

	// Notes are optional notes about the payment
	// Example: "Payment for May 2026 invoice - antibiotics batch"
	Notes string `gorm:"type:text" json:"notes,omitempty" example:"Payment for May 2026 invoice - antibiotics batch"`

	// ReferenceNumber is the transaction reference number for transfer/e-wallet payments
	// Example: "TRX-20260531-12345"
	ReferenceNumber string `gorm:"column:reference_number;type:varchar(100)" json:"referenceNumber,omitempty" example:"TRX-20260531-12345"`

	// BranchID is the foreign key to branches table for multi-branch support
	// Example: 1
	BranchID uint `gorm:"column:branch_id;not null;index:idx_supplier_payments_branch_id" json:"branchId" binding:"required" example:"1"`

	// CreatedBy is the user who recorded the payment
	// Example: 1
	CreatedBy uint `gorm:"column:created_by;not null" json:"createdBy" binding:"required" example:"1"`

	// CreatedAt is the timestamp when the payment was recorded
	// Example: "2026-05-31T10:00:00Z"
	CreatedAt time.Time `json:"createdAt" example:"2026-05-31T10:00:00Z"`

	// UpdatedAt is the timestamp when the payment was last updated
	// Example: "2026-05-31T10:00:00Z"
	UpdatedAt time.Time `json:"updatedAt" example:"2026-05-31T10:00:00Z"`

	// Relationships - eager loaded with Preload
	PurchaseInvoice PurchaseInvoice `gorm:"foreignKey:PurchaseInvoiceID" json:"purchaseInvoice"`
	Branch           Branch           `gorm:"foreignKey:BranchID" json:"branch"`
}

// TableName specifies the table name for SupplierPayment model
// Story 10.4: Table name follows snake_case plural convention
func (SupplierPayment) TableName() string {
	return "supplier_payments"
}

// BeforeCreate is a GORM hook called before creating a supplier payment
// Story 10.4: Set default values and perform validations
func (sp *SupplierPayment) BeforeCreate(tx *gorm.DB) error {
	// Set PaymentDate to current date if not set
	if sp.PaymentDate.IsZero() {
		sp.PaymentDate = time.Now().UTC()
	}
	return nil
}
