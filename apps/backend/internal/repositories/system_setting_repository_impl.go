package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// systemSettingRepository implements SystemSettingRepository interface
type systemSettingRepository struct {
	db *gorm.DB
}

// NewSystemSettingRepository creates a new system setting repository
func NewSystemSettingRepository(db interface{}) SystemSettingRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &systemSettingRepository{db: gormDB}
}

// GetByKey retrieves a system setting by its key
func (r *systemSettingRepository) GetByKey(ctx context.Context, key string) (*models.SystemSetting, error) {
	if key == "" {
		return nil, ErrInvalidInput
	}
	var setting models.SystemSetting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get system setting: %w", err)
	}
	return &setting, nil
}

// GetAll retrieves all system settings
func (r *systemSettingRepository) GetAll(ctx context.Context) ([]*models.SystemSetting, error) {
	var settings []*models.SystemSetting
	err := r.db.WithContext(ctx).Order("key").Find(&settings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all system settings: %w", err)
	}
	return settings, nil
}

// GetPharmacySettings retrieves all pharmacy-related settings
func (r *systemSettingRepository) GetPharmacySettings(ctx context.Context) ([]*models.SystemSetting, error) {
	keys := models.GetPharmacySettingsKeys()
	var settings []*models.SystemSetting
	err := r.db.WithContext(ctx).Where("key IN ?", keys).Order("key").Find(&settings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get pharmacy settings: %w", err)
	}
	return settings, nil
}

// SetValue creates or updates a system setting
func (r *systemSettingRepository) SetValue(ctx context.Context, key, value string, updatedBy uint) error {
	if key == "" {
		return ErrInvalidInput
	}

	// Use ON CONFLICT for upsert (PostgreSQL-specific)
	query := `
		INSERT INTO system_settings (key, value, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value,
		    updated_by = EXCLUDED.updated_by,
		    updated_at = CURRENT_TIMESTAMP
	`
	return r.db.WithContext(ctx).Exec(query, key, value, updatedBy).Error
}

// UpdateValue updates an existing system setting's value
func (r *systemSettingRepository) UpdateValue(ctx context.Context, key, value string, updatedBy uint) error {
	if key == "" {
		return ErrInvalidInput
	}

	result := r.db.WithContext(ctx).Model(&models.SystemSetting{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{
			"value":     value,
			"updated_by": updatedBy,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update system setting: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// UpdateMultiple updates multiple settings in a single transaction
func (r *systemSettingRepository) UpdateMultiple(ctx context.Context, settings []models.SystemSetting) error {
	if len(settings) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, setting := range settings {
			if setting.Key == "" {
				continue
			}

			// Use ON CONFLICT for upsert
			query := `
				INSERT INTO system_settings (key, value, description, updated_by, created_at, updated_at)
				VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				ON CONFLICT (key) DO UPDATE
				SET value = EXCLUDED.value,
				    description = EXCLUDED.description,
				    updated_by = EXCLUDED.updated_by,
				    updated_at = CURRENT_TIMESTAMP
			`
			err := tx.Exec(query, setting.Key, setting.Value, setting.Description, setting.UpdatedBy).Error
			if err != nil {
				return fmt.Errorf("failed to update setting %s: %w", setting.Key, err)
			}
		}
		return nil
	})
}

// GetPublicSettings retrieves settings safe for public display (receipts, reports)
func (r *systemSettingRepository) GetPublicSettings(ctx context.Context) (*models.PublicSettings, error) {
	settings, err := r.GetPharmacySettings(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to slice for the helper function
	settingSlice := make([]models.SystemSetting, len(settings))
	for i, s := range settings {
		settingSlice[i] = *s
	}

	publicSettings := models.ToPublicSettings(settingSlice)
	return &publicSettings, nil
}
