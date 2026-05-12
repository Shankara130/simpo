package models

import (
	"time"

	"gorm.io/gorm"
)

// Product represents a pharmacy product/item
type Product struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	SKU              string         `gorm:"type:varchar(50);not null;uniqueIndex:idx_products_branch_sku" json:"sku"`
	Name             string         `gorm:"type:varchar(200);not null" json:"name"`
	Description      string         `gorm:"type:text" json:"description,omitempty"`
	StockQty         int64          `gorm:"column:stock_qty;type:bigint;not null;default:0" json:"stockQty"`
	Price            string         `gorm:"type:decimal(15,2);column:price;not null" json:"price"`
	CostPrice        *string        `gorm:"type:decimal(15,2);column:cost_price" json:"cost_price,omitempty"`
	ExpiryDate       *time.Time     `gorm:"column:expiry_date" json:"expiryDate,omitempty"`
	BranchID         uint           `gorm:"column:branch_id;not null;index" json:"branchId"`
	ReorderThreshold int            `gorm:"column:reorder_threshold;default:10" json:"reorderThreshold"`
	Category         string         `gorm:"type:varchar(50)" json:"category,omitempty"`
	CreatedBy        *uint          `gorm:"column:created_by" json:"createdBy,omitempty"`
	UpdatedBy        *uint          `gorm:"column:updated_by" json:"updatedBy,omitempty"`
	Version          int            `gorm:"column:version;not null;default:1" json:"version"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Branch           *Branch        `json:"-" gorm:"foreignKey:BranchID"`
	TransactionItems []TransactionItem `json:"-" gorm:"foreignKey:ProductID"`
}

// TableName specifies the table name for Product model
func (Product) TableName() string {
	return "products"
}
