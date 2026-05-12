package models

import (
	"time"

	"gorm.io/gorm"
)

// Branch represents a pharmacy branch location
type Branch struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Address   string         `gorm:"type:text" json:"address,omitempty"`
	Phone     string         `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Email     string         `gorm:"type:varchar(100);index" json:"email,omitempty"`
	CreatedBy *uint          `gorm:"column:created_by" json:"createdBy,omitempty"`
	UpdatedBy *uint          `gorm:"column:updated_by" json:"updatedBy,omitempty"`
	Version   int            `gorm:"column:version;not null;default:1" json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships for preloading
	Products     []Product     `json:"products,omitempty" gorm:"foreignKey:BranchID"`
	Transactions []Transaction `json:"transactions,omitempty" gorm:"foreignKey:BranchID"`
}

// TableName specifies the table name for Branch model
func (Branch) TableName() string {
	return "branches"
}
