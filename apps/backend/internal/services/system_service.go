package services

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// SystemService defines the interface for system settings business operations
type SystemService interface {
	// GetSettings retrieves all system settings
	// Returns settings keyed by setting key
	GetSettings(ctx context.Context) (map[string]string, error)

	// GetPharmacySettings retrieves pharmacy-related settings
	// Returns a PharmacySettings struct with business information
	GetPharmacySettings(ctx context.Context) (*models.PharmacySettings, error)

	// UpdateSettings updates multiple system settings
	// Validates admin permissions, logs changes to audit trail (AC7)
	UpdateSettings(ctx context.Context, settings *models.PharmacySettings, adminID uint, adminUsername, ipAddress string) error

	// GetPublicSettings retrieves settings safe for public display
	// Used for receipts, reports without authentication (AC6)
	GetPublicSettings(ctx context.Context) (*models.PublicSettings, error)

	// GetBusinessName retrieves the pharmacy business name
	GetBusinessName(ctx context.Context) (string, error)

	// GetBusinessAddress retrieves the pharmacy address
	GetBusinessAddress(ctx context.Context) (string, error)

	// GetBusinessPhone retrieves the pharmacy phone number
	GetBusinessPhone(ctx context.Context) (string, error)

	// GetBusinessEmail retrieves the pharmacy email address
	GetBusinessEmail(ctx context.Context) (string, error)
}

// SystemSettingsUpdateRequest represents a request to update system settings
type SystemSettingsUpdateRequest struct {
	BusinessName string `json:"businessName" validate:"required"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Email        string `json:"email" validate:"required,email"`
}
