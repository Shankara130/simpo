package dto

import "time"

// SystemSettingsRequest represents a request to update system settings
// Story 6.1, AC1-AC4: System Administrator can configure business name, address, phone, and email
type SystemSettingsRequest struct {
	BusinessName string `json:"businessName" binding:"required" validate:"required,min=1" example:"Simpo Pharmacy"`
	Address      string `json:"address" example:"123 Main St, Jakarta, Indonesia"`
	Phone        string `json:"phone" example:"+62-21-1234-5678"`
	Email        string `json:"email" binding:"required,email" validate:"required,email" example:"admin@simpo.pharmacy"`
	LogoURL      string `json:"logoUrl,omitempty" example:"https://example.com/logo.png"`
}

// SystemSettingsResponse represents the system settings response
// Story 6.1, AC1-AC4: Returns all pharmacy settings with metadata
type SystemSettingsResponse struct {
	BusinessName string    `json:"businessName" example:"Simpo Pharmacy"`
	Address      string    `json:"address" example:"123 Main St, Jakarta, Indonesia"`
	Phone        string    `json:"phone" example:"+62-21-1234-5678"`
	Email        string    `json:"email" example:"admin@simpo.pharmacy"`
	LogoURL      string    `json:"logoUrl,omitempty" example:"https://example.com/logo.png"`
	UpdatedAt    time.Time `json:"updatedAt" example:"2026-05-26T10:30:00Z"`
	UpdatedBy    uint      `json:"updatedBy" example:"1"`
}

// PublicSettingsResponse represents public settings safe for external display
// Story 6.1, AC6: Used for receipts, reports without authentication
type PublicSettingsResponse struct {
	BusinessName string `json:"businessName" example:"Simpo Pharmacy"`
	Address      string `json:"address" example:"123 Main St, Jakarta, Indonesia"`
	Phone        string `json:"phone" example:"+62-21-1234-5678"`
	Email        string `json:"email" example:"admin@simpo.pharmacy"`
}

// SystemSettingValue represents a single key-value system setting
// Used for generic settings operations
type SystemSettingValue struct {
	Key         string `json:"key" example:"pharmacy.name"`
	Value       string `json:"value" example:"Simpo Pharmacy"`
	Description string `json:"description,omitempty" example:"Pharmacy business name"`
}

// SystemSettingsListResponse represents a list of all system settings
// Story 6.1, AC5: Returns all settings with user identification
type SystemSettingsListResponse struct {
	Settings []SystemSettingValue `json:"settings"`
	Count    int                  `json:"count"`
}

// SettingsUpdateResponse represents the result of a settings update operation
// Story 6.1, AC5, AC7: Confirmation of settings change with audit trail reference
type SettingsUpdateResponse struct {
	Message   string    `json:"message" example:"Settings updated successfully"`
	UpdatedAt time.Time `json:"updatedAt" example:"2026-05-26T10:30:00Z"`
	UpdatedBy string    `json:"updatedBy" example:"admin"`
}
