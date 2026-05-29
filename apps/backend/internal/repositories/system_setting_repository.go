package repositories

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// SystemSettingRepository defines the interface for system setting data operations
type SystemSettingRepository interface {
	// GetByKey retrieves a system setting by its key
	// Returns ErrNotFound if setting doesn't exist
	GetByKey(ctx context.Context, key string) (*models.SystemSetting, error)

	// GetAll retrieves all system settings
	GetAll(ctx context.Context) ([]*models.SystemSetting, error)

	// GetPharmacySettings retrieves all pharmacy-related settings
	GetPharmacySettings(ctx context.Context) ([]*models.SystemSetting, error)

	// SetValue creates or updates a system setting
	SetValue(ctx context.Context, key, value string, updatedBy uint) error

	// UpdateValue updates an existing system setting's value
	UpdateValue(ctx context.Context, key, value string, updatedBy uint) error

	// UpdateMultiple updates multiple settings in a single transaction
	UpdateMultiple(ctx context.Context, settings []models.SystemSetting) error

	// GetPublicSettings retrieves settings safe for public display (receipts, reports)
	GetPublicSettings(ctx context.Context) (*models.PublicSettings, error)
}
