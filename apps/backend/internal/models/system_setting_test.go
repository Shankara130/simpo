package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSystemSetting_TableName(t *testing.T) {
	setting := SystemSetting{}
	assert.Equal(t, "system_settings", setting.TableName())
}

func TestSystemSetting_IsPharmacySetting(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "pharmacy name setting",
			key:      "pharmacy.name",
			expected: true,
		},
		{
			name:     "pharmacy address setting",
			key:      "pharmacy.address",
			expected: true,
		},
		{
			name:     "non-pharmacy setting",
			key:      "system.timezone",
			expected: false,
		},
		{
			name:     "pharmacy prefix but not pharmacy setting",
			key:      "pharmacy_automation.enabled",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := SystemSetting{Key: tt.key}
			assert.Equal(t, tt.expected, setting.IsPharmacySetting())
		})
	}
}

func TestGetPharmacySettingsKeys(t *testing.T) {
	keys := GetPharmacySettingsKeys()

	assert.Len(t, keys, 5, "Should have 5 pharmacy settings keys")

	expectedKeys := []string{
		"pharmacy.name",
		"pharmacy.address",
		"pharmacy.phone",
		"pharmacy.email",
		"pharmacy.logo_url",
	}

	for _, expected := range expectedKeys {
		assert.Contains(t, keys, expected, "Should contain key: %s", expected)
	}
}

func TestPharmacySettings_ToSystemSettings(t *testing.T) {
	updatedBy := uint(42)
	settings := PharmacySettings{
		Name:    "Test Pharmacy",
		Address: "123 Test St",
		Phone:   "555-1234",
		Email:   "test@example.com",
		LogoURL: "https://example.com/logo.png",
	}

	result := settings.ToSystemSettings(updatedBy)

	assert.Len(t, result, 5, "Should convert to 5 system settings")

	// Verify each setting
	keyToValue := make(map[string]string)
	for _, s := range result {
		keyToValue[s.Key] = s.Value
		assert.Equal(t, &updatedBy, s.UpdatedBy, "UpdatedBy should be set")
		assert.Equal(t, &updatedBy, s.CreatedBy, "CreatedBy should be set")
		assert.NotZero(t, s.CreatedAt, "CreatedAt should be set")
		assert.NotZero(t, s.UpdatedAt, "UpdatedAt should be set")
	}

	assert.Equal(t, "Test Pharmacy", keyToValue["pharmacy.name"])
	assert.Equal(t, "123 Test St", keyToValue["pharmacy.address"])
	assert.Equal(t, "555-1234", keyToValue["pharmacy.phone"])
	assert.Equal(t, "test@example.com", keyToValue["pharmacy.email"])
	assert.Equal(t, "https://example.com/logo.png", keyToValue["pharmacy.logo_url"])
}

func TestToPublicSettings(t *testing.T) {
	now := time.Now()
	settings := []SystemSetting{
		{Key: "pharmacy.name", Value: "Test Pharmacy", CreatedAt: now},
		{Key: "pharmacy.address", Value: "123 Test St", CreatedAt: now},
		{Key: "pharmacy.phone", Value: "555-1234", CreatedAt: now},
		{Key: "pharmacy.email", Value: "test@example.com", CreatedAt: now},
		{Key: "pharmacy.logo_url", Value: "https://example.com/logo.png", CreatedAt: now},
		{Key: "system.debug", Value: "false", CreatedAt: now}, // Should be ignored
	}

	result := ToPublicSettings(settings)

	assert.Equal(t, "Test Pharmacy", result.BusinessName)
	assert.Equal(t, "123 Test St", result.Address)
	assert.Equal(t, "555-1234", result.Phone)
	assert.Equal(t, "test@example.com", result.Email)
	// LogoURL should not be in public settings
}

func TestToPublicSettings_EmptySettings(t *testing.T) {
	settings := []SystemSetting{}

	result := ToPublicSettings(settings)

	assert.Equal(t, "", result.BusinessName)
	assert.Equal(t, "", result.Address)
	assert.Equal(t, "", result.Phone)
	assert.Equal(t, "", result.Email)
}

func TestPharmacySettings_ToSystemSettings_EmptyValues(t *testing.T) {
	updatedBy := uint(1)
	settings := PharmacySettings{
		Name:    "",
		Address: "",
		Phone:   "",
		Email:   "",
		LogoURL: "",
	}

	result := settings.ToSystemSettings(updatedBy)

	assert.Len(t, result, 5)

	for _, s := range result {
		assert.Empty(t, s.Value, "Value should be empty for key: %s", s.Key)
	}
}
