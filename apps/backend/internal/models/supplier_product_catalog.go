package models

import (
	"time"

	"gorm.io/gorm"
)

// SupplierProductCatalog represents an association between a supplier and a product with purchase price
// Story 10.5: Supplier Product Catalog for maintaining supplier product associations and purchase prices
// This model stores both current and historical price data using effective date ranges
type SupplierProductCatalog struct {
	// ID is the unique identifier for the catalog entry
	// Example: 1
	ID uint `gorm:"primaryKey" json:"id" example:"1"`

	// SupplierID is the foreign key to suppliers table
	// Example: 1
	SupplierID uint `gorm:"column:supplier_id;not null;index:idx_supplier_product_catalog_current,priority:1;index:idx_supplier_product_catalog_current,priority:1" json:"supplierId" binding:"required" example:"1"`

	// ProductID is the foreign key to products table
	// Example: 1
	ProductID uint `gorm:"column:product_id;not null;index:idx_supplier_product_catalog_current,priority:2;index:idx_supplier_product_catalog_current,priority:2" json:"productId" binding:"required" example:"1"`

	// PurchasePrice is the current purchase price from this supplier
	// Example: 45000.00
	PurchasePrice float64 `gorm:"type:decimal(15,2);not null" json:"purchasePrice" binding:"required,gt=0" example:"45000.00"`

	// IsPreferred indicates if this is the preferred supplier for this product
	// Only one supplier can be preferred per product per branch
	// Example: true
	IsPreferred bool `gorm:"column:is_preferred;not null;default:false" json:"isPreferred" example:"false"`

	// SKUCode is the supplier's internal SKU code for this product
	// Example: "PARA-500-100"
	SKUCode string `gorm:"column:sku_code;type:varchar(50)" json:"skuCode,omitempty" binding:"omitempty,max=50" example:"PARA-500-100"`

	// MinimumOrderQuantity is the minimum quantity that can be ordered from this supplier
	// Example: 10
	MinimumOrderQuantity int `gorm:"column:minimum_order_quantity;not null;default:1" json:"minimumOrderQuantity" binding:"required,min=1" example:"10"`

	// LeadTimeDays is the average lead time in days for delivery from this supplier
	// Example: 3
	LeadTimeDays *int `gorm:"column:lead_time_days" json:"leadTimeDays,omitempty" binding:"omitempty,min=0" example:"3"`

	// BranchID is the foreign key to branches table for multi-branch support
	// Example: 1
	BranchID uint `gorm:"column:branch_id;not null;index:idx_supplier_product_catalog_branch" json:"branchId" binding:"required" example:"1"`

	// CreatedBy is the user who created this catalog entry
	// Example: 1
	CreatedBy uint `gorm:"column:created_by;not null" json:"createdBy" binding:"required" example:"1"`

	// UpdatedBy is the user who last updated this catalog entry
	// Example: 1
	UpdatedBy *uint `gorm:"column:updated_by" json:"updatedBy,omitempty" example:"1"`

	// CreatedAt is the timestamp when the catalog entry was created
	// Example: "2026-05-31T10:00:00Z"
	CreatedAt time.Time `json:"createdAt" example:"2026-05-31T10:00:00Z"`

	// UpdatedAt is the timestamp when the catalog entry was last updated
	// Example: "2026-05-31T10:00:00Z"
	UpdatedAt time.Time `json:"updatedAt" example:"2026-05-31T10:00:00Z"`

	// PriceEffectiveFrom is the date when this price becomes effective
	// Example: "2026-05-31"
	PriceEffectiveFrom time.Time `gorm:"column:price_effective_from;type:date;not null;default:CURRENT_DATE;index:idx_supplier_product_catalog_current,priority:3;index:idx_supplier_product_catalog_dates,priority:1" json:"priceEffectiveFrom" example:"2026-05-31"`

	// PriceEffectiveTo is the date when this price ends (NULL means this is the current price)
	// Example: null (current price) or "2026-06-15" (historical price)
	// Story 10.5: Price history tracking - NULL = current price, set date = historical price
	PriceEffectiveTo *time.Time `gorm:"column:price_effective_to;type:date;index:idx_supplier_product_catalog_current,where:price_effective_to IS NULL;index:idx_supplier_product_catalog_dates,priority:2" json:"priceEffectiveTo,omitempty" example:""`

	// Relationships - eager loaded with Preload
	Supplier Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Product  Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Branch   Branch   `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

// TableName specifies the table name for SupplierProductCatalog model
// Story 10.5: Table name follows snake_case plural convention
func (SupplierProductCatalog) TableName() string {
	return "supplier_product_catalogs"
}

// IsCurrentPrice checks if this catalog entry represents the current active price
// A price is current if price_effective_to is NULL
// Story 10.5: Price history tracking method
func (spc *SupplierProductCatalog) IsCurrentPrice() bool {
	return spc.PriceEffectiveTo == nil
}

// BeforeCreate is a GORM hook called before creating a catalog entry
// Story 10.5: Set default values and perform validations
func (spc *SupplierProductCatalog) BeforeCreate(tx *gorm.DB) error {
	// Set PriceEffectiveFrom to current date if not set
	if spc.PriceEffectiveFrom.IsZero() {
		spc.PriceEffectiveFrom = time.Now().UTC()
	}
	return nil
}

// BeforeUpdate is a GORM hook called before updating a catalog entry
// Story 10.5: Prevent direct updates to price (use UpdatePrice method instead)
func (spc *SupplierProductCatalog) BeforeUpdate(tx *gorm.DB) error {
	// Prevent direct updates to PurchasePrice - should use UpdatePrice service method
	// This ensures price history is properly tracked
	if tx.Statement.Changed("purchase_price") {
		// Allow updates only if PriceEffectiveTo is being set (archiving old price)
		if spc.PriceEffectiveTo == nil {
			return gorm.ErrInvalidTransaction
		}
	}
	return nil
}
