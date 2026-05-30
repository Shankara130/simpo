package models

import (
	"time"

	"gorm.io/gorm"
)

// Supplier represents a supplier master data entity
// Story 10.1: Supplier Master Data Management
type Supplier struct {
	// ID is the unique identifier for the supplier
	// Example: 1
	ID uint `gorm:"primaryKey" json:"id" example:"1"`

	// Name is the supplier name (unique, max 200 characters)
	// Example: "PT. Pharmasi Jaya"
	Name string `gorm:"type:varchar(200);uniqueIndex;not null" json:"name" binding:"required,max=200" example:"PT. Pharmasi Jaya"`

	// ContactPerson is the primary contact person name
	// Example: "Budi Santoso"
	ContactPerson string `gorm:"type:varchar(100)" json:"contactPerson,omitempty" binding:"omitempty,max=100" example:"Budi Santoso"`

	// Phone is the contact phone number (required)
	// Example: "+62-21-1234-5678"
	Phone string `gorm:"type:varchar(20);not null" json:"phone" binding:"required" example:"+62-21-1234-5678"`

	// Email is the contact email address
	// Example: "contact@pharmasi.com"
	Email string `gorm:"type:varchar(100)" json:"email,omitempty" binding:"omitempty,email,max=100" example:"contact@pharmasi.com"`

	// Address is the physical address of the supplier
	// Example: "Jl. Industri No. 123, Jakarta"
	Address string `gorm:"type:varchar(500)" json:"address,omitempty" binding:"omitempty,max=500" example:"Jl. Industri No. 123, Jakarta"`

	// IsActive indicates if the supplier is active (soft delete via deleted_at)
	// Example: true
	IsActive bool `gorm:"column:is_active;not null;default:true" json:"isActive" example:"true"`

	// CreatedBy is the user who created the supplier
	// Example: 1
	CreatedBy *uint `gorm:"column:created_by" json:"createdBy,omitempty"`

	// UpdatedBy is the user who last updated the supplier
	// Example: 1
	UpdatedBy *uint `gorm:"column:updated_by" json:"updatedBy,omitempty"`

	// DeletedBy is the user who deactivated the supplier
	// Example: 1
	DeletedBy *uint `gorm:"column:deleted_by" json:"deletedBy,omitempty"`

	// Version is the optimistic locking version
	// Example: 1
	Version int `gorm:"column:version;not null;default:1" json:"version" example:"1"`

	// CreatedAt is the timestamp when the supplier was created
	// Example: "2026-05-30T10:00:00Z"
	CreatedAt time.Time `json:"createdAt" example:"2026-05-30T10:00:00Z"`

	// UpdatedAt is the timestamp when the supplier was last updated
	// Example: "2026-05-30T10:00:00Z"
	UpdatedAt time.Time `json:"updatedAt" example:"2026-05-30T10:00:00Z"`

	// DeletedAt is the soft delete timestamp (NULL for active suppliers)
	// Example: null
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Supplier model
// Story 10.1: Table name follows snake_case plural convention
func (Supplier) TableName() string {
	return "suppliers"
}
