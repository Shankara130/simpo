package models

import (
	"time"

	"gorm.io/gorm"
)

// Transaction represents a sales transaction
type Transaction struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	TransactionNumber string         `gorm:"type:varchar(50);uniqueIndex:idx_transactions_number;not null;column:transaction_number" json:"transactionNumber"`
	CashierID         uint           `gorm:"column:cashier_id;not null;index" json:"cashierId"`
	BranchID          uint           `gorm:"column:branch_id;not null;index" json:"branchId"`
	Total             string         `gorm:"type:decimal(12,2);column:total;not null" json:"total"`
	Subtotal          string         `gorm:"type:decimal(12,2);column:subtotal;not null" json:"subtotal"`
	Tax               string         `gorm:"type:decimal(12,2);column:tax;default:0" json:"tax"`
	Discount          string         `gorm:"type:decimal(12,2);column:discount;default:0" json:"discount"`
	PaymentMethod     string         `gorm:"type:varchar(20);column:payment_method;not null" json:"paymentMethod"`
	IdempotencyKey    string         `gorm:"column:idempotency_key;uniqueIndex;size:255" json:"idempotencyKey,omitempty"`
	Status            string         `gorm:"type:varchar(20);column:status;not null;default:COMPLETED" json:"status"`
	CustomerName      *string        `gorm:"type:varchar(100);column:customer_name" json:"customerName,omitempty"`
	Notes             *string        `gorm:"type:text;column:notes" json:"notes,omitempty"`
	CreatedBy         *uint          `gorm:"column:created_by" json:"createdBy,omitempty"`
	UpdatedBy         *uint          `gorm:"column:updated_by" json:"updatedBy,omitempty"`
	Version           int            `gorm:"column:version;not null;default:1" json:"version"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships - Note: User relationship will be handled separately (User is in internal/user package)
	Branch           *Branch           `json:"-" gorm:"foreignKey:BranchID"`
	TransactionItems []TransactionItem `json:"transactionItems,omitempty" gorm:"foreignKey:TransactionID"`
}

// TableName specifies the table name for Transaction model
func (Transaction) TableName() string {
	return "transactions"
}

// Constants for payment methods
const (
	PaymentMethodCash     = "CASH"
	PaymentMethodTransfer = "TRANSFER"
	PaymentMethodEWallet  = "E-WALLET"
	PaymentMethodCard     = "CARD"
	PaymentMethodQRIS     = "QRIS"
)

// Constants for transaction status
const (
	StatusPending   = "PENDING"
	StatusCompleted = "COMPLETED"
	StatusCancelled = "CANCELLED"
	StatusRefunded  = "REFUNDED"
)
