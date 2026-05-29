package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// systemServiceImpl implements SystemService interface
type systemServiceImpl struct {
	settingRepo  repositories.SystemSettingRepository
	auditService AuditService
}

// NewSystemService creates a new system settings service
func NewSystemService(
	settingRepo repositories.SystemSettingRepository,
	auditService AuditService,
) SystemService {
	return &systemServiceImpl{
		settingRepo:  settingRepo,
		auditService: auditService,
	}
}

// GetSettings retrieves all system settings
func (s *systemServiceImpl) GetSettings(ctx context.Context) (map[string]string, error) {
	settings, err := s.settingRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	result := make(map[string]string)
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}

	return result, nil
}

// GetPharmacySettings retrieves pharmacy-related settings
func (s *systemServiceImpl) GetPharmacySettings(ctx context.Context) (*models.PharmacySettings, error) {
	settings, err := s.settingRepo.GetPharmacySettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pharmacy settings: %w", err)
	}

	// Convert to PharmacySettings struct
	result := &models.PharmacySettings{}
	for _, setting := range settings {
		switch setting.Key {
		case "pharmacy.name":
			result.Name = setting.Value
		case "pharmacy.address":
			result.Address = setting.Value
		case "pharmacy.phone":
			result.Phone = setting.Value
		case "pharmacy.email":
			result.Email = setting.Value
		case "pharmacy.logo_url":
			result.LogoURL = setting.Value
		}
	}

	return result, nil
}

// UpdateSettings updates multiple system settings with audit logging (AC7)
func (s *systemServiceImpl) UpdateSettings(
	ctx context.Context,
	settings *models.PharmacySettings,
	adminID uint,
	adminUsername string,
	ipAddress string,
) error {
	if settings == nil {
		return &InvalidInputError{Field: "settings", Message: "cannot be nil"}
	}

	// Get old settings for audit trail
	oldSettings, err := s.GetPharmacySettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current settings: %w", err)
	}

	// Convert to SystemSetting slice for update
	newSettings := settings.ToSystemSettings(adminID)

	// Update in transaction
	err = s.settingRepo.UpdateMultiple(ctx, newSettings)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	// Log to audit trail (AC7)
	// Story 6.1, AC7: Use dedicated LogSettingsUpdate method for proper audit logging
	changes := s.buildChangesMap(oldSettings, settings)
	changesJSON, _ := json.Marshal(changes)

	if err := s.auditService.LogSettingsUpdate(ctx, adminID, adminUsername, string(changesJSON), ipAddress); err != nil {
		// Log but don't fail the operation
		slog.Warn("Failed to log settings update to audit trail", "error", err)
	}

	return nil
}

// GetPublicSettings retrieves settings safe for public display (AC6)
func (s *systemServiceImpl) GetPublicSettings(ctx context.Context) (*models.PublicSettings, error) {
	// Fetch from repository
	// Note: Caching can be added later as an enhancement
	settings, err := s.settingRepo.GetPublicSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get public settings: %w", err)
	}

	return settings, nil
}

// GetBusinessName retrieves the pharmacy business name
func (s *systemServiceImpl) GetBusinessName(ctx context.Context) (string, error) {
	setting, err := s.settingRepo.GetByKey(ctx, "pharmacy.name")
	if err != nil {
		if err == repositories.ErrNotFound {
			return "Simpo Pharmacy", nil // Default name
		}
		return "", fmt.Errorf("failed to get business name: %w", err)
	}
	return setting.Value, nil
}

// GetBusinessAddress retrieves the pharmacy address
func (s *systemServiceImpl) GetBusinessAddress(ctx context.Context) (string, error) {
	setting, err := s.settingRepo.GetByKey(ctx, "pharmacy.address")
	if err != nil {
		if err == repositories.ErrNotFound {
			return "", nil // Default empty address
		}
		return "", fmt.Errorf("failed to get business address: %w", err)
	}
	return setting.Value, nil
}

// GetBusinessPhone retrieves the pharmacy phone number
func (s *systemServiceImpl) GetBusinessPhone(ctx context.Context) (string, error) {
	setting, err := s.settingRepo.GetByKey(ctx, "pharmacy.phone")
	if err != nil {
		if err == repositories.ErrNotFound {
			return "", nil // Default empty phone
		}
		return "", fmt.Errorf("failed to get business phone: %w", err)
	}
	return setting.Value, nil
}

// GetBusinessEmail retrieves the pharmacy email address
func (s *systemServiceImpl) GetBusinessEmail(ctx context.Context) (string, error) {
	setting, err := s.settingRepo.GetByKey(ctx, "pharmacy.email")
	if err != nil {
		if err == repositories.ErrNotFound {
			return "", nil // Default empty email
		}
		return "", fmt.Errorf("failed to get business email: %w", err)
	}
	return setting.Value, nil
}

// buildChangesMap creates a map of changes for audit logging
func (s *systemServiceImpl) buildChangesMap(old, new *models.PharmacySettings) map[string]interface{} {
	changes := make(map[string]interface{})

	if old.Name != new.Name {
		changes["pharmacy.name"] = map[string]string{
			"old": old.Name,
			"new": new.Name,
		}
	}
	if old.Address != new.Address {
		changes["pharmacy.address"] = map[string]string{
			"old": old.Address,
			"new": new.Address,
		}
	}
	if old.Phone != new.Phone {
		changes["pharmacy.phone"] = map[string]string{
			"old": old.Phone,
			"new": new.Phone,
		}
	}
	if old.Email != new.Email {
		changes["pharmacy.email"] = map[string]string{
			"old": old.Email,
			"new": new.Email,
		}
	}
	if old.LogoURL != new.LogoURL {
		changes["pharmacy.logo_url"] = map[string]string{
			"old": old.LogoURL,
			"new": new.LogoURL,
		}
	}

	return changes
}
