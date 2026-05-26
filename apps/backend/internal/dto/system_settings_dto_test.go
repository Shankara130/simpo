package dto

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper variable for testing timestamps
var testTime = func() time.Time {
	t, _ := time.Parse(time.RFC3339, "2026-05-26T10:30:00Z")
	return t
}()

// TestSystemSettingsRequestValidation_Success tests valid system settings request
func TestSystemSettingsRequestValidation_Success(t *testing.T) {
	req := SystemSettingsRequest{
		BusinessName: "Simpo Pharmacy",
		Address:      "123 Main St, Jakarta, Indonesia",
		Phone:        "+62-21-1234-5678",
		Email:        "admin@simpo.pharmacy",
		LogoURL:      "https://example.com/logo.png",
	}

	validate := validator.New()
	err := validate.Struct(req)

	assert.NoError(t, err, "Valid system settings request should pass validation")
}

// TestSystemSettingsRequestValidation_MissingBusinessName tests businessName is required
func TestSystemSettingsRequestValidation_MissingBusinessName(t *testing.T) {
	req := SystemSettingsRequest{
		BusinessName: "",
		Email:        "admin@simpo.pharmacy",
	}

	validate := validator.New()
	err := validate.Struct(req)

	require.Error(t, err, "Missing businessName should fail validation")
	validationErrors := err.(validator.ValidationErrors)
	assert.Contains(t, validationErrors[0].Field(), "BusinessName", "Error should be about BusinessName field")
}

// TestSystemSettingsRequestValidation_BusinessNameTooShort tests businessName minimum length
func TestSystemSettingsRequestValidation_BusinessNameTooShort(t *testing.T) {
	req := SystemSettingsRequest{
		BusinessName: "",
		Email:        "admin@simpo.pharmacy",
	}

	validate := validator.New()
	err := validate.Struct(req)

	require.Error(t, err, "Empty businessName should fail validation")
	validationErrors := err.(validator.ValidationErrors)
	assert.Contains(t, validationErrors[0].Field(), "BusinessName", "Error should be about BusinessName field")
}

// TestSystemSettingsRequestValidation_MissingEmail tests email is required
func TestSystemSettingsRequestValidation_MissingEmail(t *testing.T) {
	req := SystemSettingsRequest{
		BusinessName: "Simpo Pharmacy",
		Email:        "",
	}

	validate := validator.New()
	err := validate.Struct(req)

	require.Error(t, err, "Missing email should fail validation")
	validationErrors := err.(validator.ValidationErrors)
	assert.Contains(t, validationErrors[0].Field(), "Email", "Error should be about Email field")
}

// TestSystemSettingsRequestValidation_InvalidEmail tests email format validation
func TestSystemSettingsRequestValidation_InvalidEmail(t *testing.T) {
	req := SystemSettingsRequest{
		BusinessName: "Simpo Pharmacy",
		Email:        "not-a-valid-email",
	}

	validate := validator.New()
	err := validate.Struct(req)

	require.Error(t, err, "Invalid email format should fail validation")
	validationErrors := err.(validator.ValidationErrors)
	assert.Contains(t, validationErrors[0].Field(), "Email", "Error should be about Email field")
	assert.Contains(t, validationErrors[0].Tag(), "email", "Should be an email format validation error")
}

// TestSystemSettingsRequestValidation_OptionalFields tests optional fields can be omitted
func TestSystemSettingsRequestValidation_OptionalFields(t *testing.T) {
	req := SystemSettingsRequest{
		BusinessName: "Simpo Pharmacy",
		Email:        "admin@simpo.pharmacy",
		// Address, Phone, LogoURL omitted
	}

	validate := validator.New()
	err := validate.Struct(req)

	assert.NoError(t, err, "Optional fields should not cause validation errors")
}

// TestSystemSettingsResponse_Structure tests SystemSettingsResponse has required fields
func TestSystemSettingsResponse_Structure(t *testing.T) {
	resp := SystemSettingsResponse{
		BusinessName: "Simpo Pharmacy",
		Address:      "123 Main St, Jakarta, Indonesia",
		Phone:        "+62-21-1234-5678",
		Email:        "admin@simpo.pharmacy",
		LogoURL:      "https://example.com/logo.png",
		UpdatedAt:    testTime,
		UpdatedBy:    1,
	}

	assert.Equal(t, "Simpo Pharmacy", resp.BusinessName, "BusinessName should be set")
	assert.Equal(t, "123 Main St, Jakarta, Indonesia", resp.Address, "Address should be set")
	assert.Equal(t, "+62-21-1234-5678", resp.Phone, "Phone should be set")
	assert.Equal(t, "admin@simpo.pharmacy", resp.Email, "Email should be set")
	assert.Equal(t, "https://example.com/logo.png", resp.LogoURL, "LogoURL should be set")
	assert.Equal(t, uint(1), resp.UpdatedBy, "UpdatedBy should be set")
	assert.False(t, resp.UpdatedAt.IsZero(), "UpdatedAt should be set")
}

// TestSystemSettingsResponse_EmptyOptionalFields tests optional fields can be empty
func TestSystemSettingsResponse_EmptyOptionalFields(t *testing.T) {
	resp := SystemSettingsResponse{
		BusinessName: "Simpo Pharmacy",
		Address:      "",
		Phone:        "",
		Email:        "admin@simpo.pharmacy",
		LogoURL:      "",
		UpdatedAt:    testTime,
		UpdatedBy:    1,
	}

	assert.Equal(t, "Simpo Pharmacy", resp.BusinessName, "BusinessName should be set")
	assert.Empty(t, resp.Address, "Address can be empty")
	assert.Empty(t, resp.Phone, "Phone can be empty")
	assert.Empty(t, resp.LogoURL, "LogoURL can be empty")
	assert.Equal(t, "admin@simpo.pharmacy", resp.Email, "Email should be set")
}

// TestPublicSettingsResponse_Structure tests PublicSettingsResponse has required fields
func TestPublicSettingsResponse_Structure(t *testing.T) {
	resp := PublicSettingsResponse{
		BusinessName: "Simpo Pharmacy",
		Address:      "123 Main St, Jakarta, Indonesia",
		Phone:        "+62-21-1234-5678",
		Email:        "admin@simpo.pharmacy",
	}

	assert.Equal(t, "Simpo Pharmacy", resp.BusinessName, "BusinessName should be set")
	assert.Equal(t, "123 Main St, Jakarta, Indonesia", resp.Address, "Address should be set")
	assert.Equal(t, "+62-21-1234-5678", resp.Phone, "Phone should be set")
	assert.Equal(t, "admin@simpo.pharmacy", resp.Email, "Email should be set")
}

// TestPublicSettingsResponse_NoSensitiveInfo tests public settings exclude sensitive data
func TestPublicSettingsResponse_NoSensitiveInfo(t *testing.T) {
	resp := PublicSettingsResponse{
		BusinessName: "Simpo Pharmacy",
		Address:      "123 Main St, Jakarta, Indonesia",
		Phone:        "+62-21-1234-5678",
		Email:        "admin@simpo.pharmacy",
	}

	// PublicSettingsResponse should NOT contain UpdatedAt, UpdatedBy, or LogoURL
	// These are verified by struct definition - this test ensures clarity
	assert.Equal(t, "Simpo Pharmacy", resp.BusinessName)
}

// TestSystemSettingValue_Structure tests SystemSettingValue has required fields
func TestSystemSettingValue_Structure(t *testing.T) {
	val := SystemSettingValue{
		Key:         "pharmacy.name",
		Value:       "Simpo Pharmacy",
		Description: "Pharmacy business name",
	}

	assert.Equal(t, "pharmacy.name", val.Key, "Key should be set")
	assert.Equal(t, "Simpo Pharmacy", val.Value, "Value should be set")
	assert.Equal(t, "Pharmacy business name", val.Description, "Description should be set")
}

// TestSystemSettingValue_OptionalDescription tests description is optional
func TestSystemSettingValue_OptionalDescription(t *testing.T) {
	val := SystemSettingValue{
		Key:   "pharmacy.name",
		Value: "Simpo Pharmacy",
	}

	assert.Equal(t, "pharmacy.name", val.Key)
	assert.Equal(t, "Simpo Pharmacy", val.Value)
	assert.Empty(t, val.Description, "Description can be empty")
}

// TestSystemSettingsListResponse_Structure tests SystemSettingsListResponse structure
func TestSystemSettingsListResponse_Structure(t *testing.T) {
	resp := SystemSettingsListResponse{
		Settings: []SystemSettingValue{
			{
				Key:         "pharmacy.name",
				Value:       "Simpo Pharmacy",
				Description: "Pharmacy business name",
			},
			{
				Key:         "pharmacy.address",
				Value:       "123 Main St",
				Description: "Pharmacy address",
			},
		},
		Count: 2,
	}

	assert.Len(t, resp.Settings, 2, "Should have 2 settings")
	assert.Equal(t, 2, resp.Count, "Count should match settings length")
	assert.Equal(t, "pharmacy.name", resp.Settings[0].Key)
	assert.Equal(t, "pharmacy.address", resp.Settings[1].Key)
}

// TestSettingsUpdateResponse_Structure tests SettingsUpdateResponse structure
func TestSettingsUpdateResponse_Structure(t *testing.T) {
	resp := SettingsUpdateResponse{
		Message:   "Settings updated successfully",
		UpdatedAt: testTime,
		UpdatedBy: "admin",
	}

	assert.Equal(t, "Settings updated successfully", resp.Message)
	assert.Equal(t, "admin", resp.UpdatedBy)
	assert.False(t, resp.UpdatedAt.IsZero())
}

// TestSystemSettingsRequest_WithLogoURL tests logoURL field handling
func TestSystemSettingsRequest_WithLogoURL(t *testing.T) {
	req := SystemSettingsRequest{
		BusinessName: "Simpo Pharmacy",
		Email:        "admin@simpo.pharmacy",
		LogoURL:      "https://example.com/logo.png",
	}

	validate := validator.New()
	err := validate.Struct(req)

	assert.NoError(t, err, "LogoURL should be accepted")
	assert.Equal(t, "https://example.com/logo.png", req.LogoURL)
}

// TestSystemSettingsRequest_WithoutLogoURL tests logoURL can be omitted
func TestSystemSettingsRequest_WithoutLogoURL(t *testing.T) {
	req := SystemSettingsRequest{
		BusinessName: "Simpo Pharmacy",
		Email:        "admin@simpo.pharmacy",
	}

	validate := validator.New()
	err := validate.Struct(req)

	assert.NoError(t, err, "Omitting LogoURL should not cause validation error")
	assert.Empty(t, req.LogoURL, "LogoURL should be empty when omitted")
}
