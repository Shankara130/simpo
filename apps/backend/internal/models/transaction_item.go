package models

import (
	"time"

	"gorm.io/gorm"
)

// TransactionItem represents a line item in a transaction
type TransactionItem struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TransactionID uint           `gorm:"column:transaction_id;not null;index" json:"transactionId"`
	ProductID     uint           `gorm:"column:product_id;not null;index" json:"productId"`
	Quantity      int64          `gorm:"column:quantity;type:bigint;not null" json:"quantity"`
	UnitPrice     string         `gorm:"type:decimal(12,2);column:unit_price;not null" json:"unitPrice"`
	Subtotal      string         `gorm:"type:decimal(12,2);column:subtotal;not null" json:"subtotal"`
	CostPrice     *string        `gorm:"type:decimal(12,2);column:cost_price" json:"costPrice,omitempty"`
	ProductName   string         `gorm:"type:varchar(200);column:product_name;not null" json:"productName"`
	ProductSKU    string         `gorm:"type:varchar(50);column:product_sku;not null" json:"productSku"`
	CreatedBy     *uint          `gorm:"column:created_by" json:"createdBy,omitempty"`
	Version       int            `gorm:"column:version;not null;default:1" json:"version"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Transaction *Transaction `json:"-" gorm:"foreignKey:TransactionID"`
	Product     *Product     `json:"-" gorm:"foreignKey:ProductID"`
}

// TableName specifies the table name for TransactionItem model
func (TransactionItem) TableName() string {
	return "transaction_items"
}
