package models

import (
	"time"

	"gorm.io/gorm"
)

// SystemSetting represents a system configuration setting
// Stores pharmacy-wide settings like business name, address, contact info
type SystemSetting struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Key         string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_system_settings_key" json:"key"`
	Value       string         `gorm:"type:text;not null" json:"value"`
	Description string         `gorm:"type:varchar(255)" json:"description,omitempty"`
	CreatedBy   *uint          `gorm:"column:created_by" json:"createdBy,omitempty"`
	UpdatedBy   *uint          `gorm:"column:updated_by" json:"updatedBy,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	// Note: CreatedByUser and UpdatedByUser relationships can be added later
	// The foreign key constraints are handled at the database level
}

// TableName specifies the table name for SystemSetting model
func (SystemSetting) TableName() string {
	return "system_settings"
}

// BeforeCreate is a GORM hook called before creating a system setting
func (s *SystemSetting) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeUpdate is a GORM hook called before updating a system setting
func (s *SystemSetting) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}

// IsPharmacySetting checks if this setting is a pharmacy-related setting
func (s *SystemSetting) IsPharmacySetting() bool {
	prefix := "pharmacy."
	return len(s.Key) > len(prefix) && s.Key[:len(prefix)] == prefix
}

// GetPharmacySettingsKeys returns the list of standard pharmacy setting keys
func GetPharmacySettingsKeys() []string {
	return []string{
		"pharmacy.name",
		"pharmacy.address",
		"pharmacy.phone",
		"pharmacy.email",
		"pharmacy.logo_url",
	}
}

// PharmacySettings represents the collection of pharmacy-related settings
type PharmacySettings struct {
	Name     string `json:"businessName"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	LogoURL  string `json:"logoUrl,omitempty"`
}

// ToSystemSettings converts PharmacySettings to a slice of SystemSetting
func (p *PharmacySettings) ToSystemSettings(updatedBy uint) []SystemSetting {
	now := time.Now()
	return []SystemSetting{
		{
			Key:         "pharmacy.name",
			Value:       p.Name,
			Description: "Pharmacy business name",
			CreatedBy:   &updatedBy,
			UpdatedBy:   &updatedBy,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Key:         "pharmacy.address",
			Value:       p.Address,
			Description: "Pharmacy street address",
			CreatedBy:   &updatedBy,
			UpdatedBy:   &updatedBy,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Key:         "pharmacy.phone",
			Value:       p.Phone,
			Description: "Pharmacy phone number",
			CreatedBy:   &updatedBy,
			UpdatedBy:   &updatedBy,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Key:         "pharmacy.email",
			Value:       p.Email,
			Description: "Pharmacy email address",
			CreatedBy:   &updatedBy,
			UpdatedBy:   &updatedBy,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Key:         "pharmacy.logo_url",
			Value:       p.LogoURL,
			Description: "Pharmacy logo URL",
			CreatedBy:   &updatedBy,
			UpdatedBy:   &updatedBy,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

// PublicSettings represents settings that can be exposed publicly
// (e.g., for receipts, reports) without requiring authentication
type PublicSettings struct {
	BusinessName string `json:"businessName"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
}

// ToPublicSettings converts a slice of SystemSetting to PublicSettings
func ToPublicSettings(settings []SystemSetting) PublicSettings {
	result := PublicSettings{
		BusinessName: "Simpo Pharmacy", // Default fallback
		Address:      "",
		Phone:        "",
		Email:        "",
	}
	for _, s := range settings {
		switch s.Key {
		case "pharmacy.name":
			result.BusinessName = s.Value
		case "pharmacy.address":
			result.Address = s.Value
		case "pharmacy.phone":
			result.Phone = s.Value
		case "pharmacy.email":
			result.Email = s.Value
		}
	}
	return result
}
