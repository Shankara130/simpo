package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// mockAuditService is a mock audit service for testing
type mockAuditService struct{}

func (m *mockAuditService) LogLoginAttempt(ctx context.Context, entry AuditLogEntry) error {
	return nil
}

func (m *mockAuditService) LogAuthorizationFailure(ctx context.Context, entry AuditLogEntry) error {
	return nil
}

func (m *mockAuditService) LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogWhitelistChange(ctx context.Context, adminID uint, adminUsername string, domain string, action models.AuditAction, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogUserDeactivation(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogStockAdjustment(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogBlockedSaleAttempt(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogReportExport(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string, ipAddress string) error {
	return nil
}

// Story 6.4: Backup audit methods
func (m *mockAuditService) LogSettingsUpdate(ctx context.Context, adminID uint, adminUsername string, changesJSON string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogBackupCreated(ctx context.Context, adminID uint, adminUsername string, backupFile string, size int64, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogBackupRestored(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogBackupDeleted(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogRoleUpdated(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, oldRole string, newRole string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogPermissionGranted(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogPermissionRevoked(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogBranchCreated(ctx context.Context, adminID uint, adminUsername string, branchName string, branchLocation string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogBranchUpdated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, changes string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogBranchDeactivated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, reason string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogSystemStartup(ctx context.Context, systemID string, serverInfo string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogSystemShutdown(ctx context.Context, systemID string, reason string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogMaintenanceModeEnabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	return nil
}

func (m *mockAuditService) LogMaintenanceModeDisabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	return nil
}

func setupSystemServiceTestDB(t *testing.T) (*gorm.DB, SystemService) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Create tables
	err = db.AutoMigrate(&models.SystemSetting{})
	require.NoError(t, err)

	settingRepo := repositories.NewSystemSettingRepository(db)
	auditSvc := &mockAuditService{}

	service := NewSystemService(settingRepo, auditSvc)

	return db, service
}

func TestSystemService_GetSettings(t *testing.T) {
	db, service := setupSystemServiceTestDB(t)
	ctx := context.Background()

	// Create test settings
	settings := []models.SystemSetting{
		{Key: "pharmacy.name", Value: "Test Pharmacy"},
		{Key: "system.timezone", Value: "UTC"},
	}
	for _, s := range settings {
		require.NoError(t, db.Create(&s).Error)
	}

	// Get all settings
	result, err := service.GetSettings(ctx)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Test Pharmacy", result["pharmacy.name"])
	assert.Equal(t, "UTC", result["system.timezone"])
}

func TestSystemService_GetPharmacySettings(t *testing.T) {
	db, service := setupSystemServiceTestDB(t)
	ctx := context.Background()

	// Insert default pharmacy name from migration
	defaultSetting := &models.SystemSetting{
		Key:   "pharmacy.name",
		Value: "Simpo Pharmacy",
	}
	require.NoError(t, db.Create(defaultSetting).Error)

	result, err := service.GetPharmacySettings(ctx)
	require.NoError(t, err)

	// Should have default values from migration
	assert.NotEmpty(t, result.Name) // pharmacy.name has default "Simpo Pharmacy"
	assert.Empty(t, result.Address)  // pharmacy.address is empty by default
}

func TestSystemService_UpdateSettings(t *testing.T) {
	_, service := setupSystemServiceTestDB(t)
	ctx := context.Background()
	adminID := uint(1)
	adminUsername := "admin"
	ipAddress := "127.0.0.1"

	// Initial settings
	initialSettings := &models.PharmacySettings{
		Name:    "Initial Pharmacy",
		Address:  "123 Initial St",
		Phone:    "555-0000",
		Email:    "initial@example.com",
		LogoURL:  "",
	}
	err := service.UpdateSettings(ctx, initialSettings, adminID, adminUsername, ipAddress)
	require.NoError(t, err)

	// Verify settings were updated
	result, err := service.GetPharmacySettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Initial Pharmacy", result.Name)
	assert.Equal(t, "123 Initial St", result.Address)
	assert.Equal(t, "555-0000", result.Phone)
	assert.Equal(t, "initial@example.com", result.Email)

	// Update with new values
	newSettings := &models.PharmacySettings{
		Name:    "Updated Pharmacy",
		Address:  "456 Updated St",
		Phone:    "555-9999",
		Email:    "updated@example.com",
		LogoURL:  "",
	}
	err = service.UpdateSettings(ctx, newSettings, adminID, adminUsername, ipAddress)
	require.NoError(t, err)

	// Verify updates
	result, err = service.GetPharmacySettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Updated Pharmacy", result.Name)
	assert.Equal(t, "456 Updated St", result.Address)
	assert.Equal(t, "555-9999", result.Phone)
	assert.Equal(t, "updated@example.com", result.Email)
}

func TestSystemService_UpdateSettings_InvalidInput(t *testing.T) {
	_, service := setupSystemServiceTestDB(t)
	ctx := context.Background()

	err := service.UpdateSettings(ctx, nil, 1, "admin", "127.0.0.1")
	assert.Error(t, err)
	assert.IsType(t, &InvalidInputError{}, err)
}

func TestSystemService_GetPublicSettings(t *testing.T) {
	_, service := setupSystemServiceTestDB(t)
	ctx := context.Background()

	// Set up pharmacy settings
	settings := &models.PharmacySettings{
		Name:    "Public Test Pharmacy",
		Address:  "789 Public St",
		Phone:    "555-7777",
		Email:    "public@example.com",
		LogoURL:  "",
	}
	err := service.UpdateSettings(ctx, settings, 1, "admin", "127.0.0.1")
	require.NoError(t, err)

	// Get public settings
	result, err := service.GetPublicSettings(ctx)
	require.NoError(t, err)

	assert.Equal(t, "Public Test Pharmacy", result.BusinessName)
	assert.Equal(t, "789 Public St", result.Address)
	assert.Equal(t, "555-7777", result.Phone)
	assert.Equal(t, "public@example.com", result.Email)
}

func TestSystemService_GetBusinessName_Default(t *testing.T) {
	_, service := setupSystemServiceTestDB(t)
	ctx := context.Background()

	// Should return default name when not set
	name, err := service.GetBusinessName(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Simpo Pharmacy", name)
}

func TestSystemService_GetBusinessName_Custom(t *testing.T) {
	_, service := setupSystemServiceTestDB(t)
	ctx := context.Background()

	// Set custom name
	err := service.UpdateSettings(ctx, &models.PharmacySettings{
		Name:   "Custom Pharmacy",
		Email:  "test@example.com",
	}, 1, "admin", "127.0.0.1")
	require.NoError(t, err)

	// Get business name
	name, err := service.GetBusinessName(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Custom Pharmacy", name)
}

func TestSystemService_GetBusinessAddress_Empty(t *testing.T) {
	_, service := setupSystemServiceTestDB(t)
	ctx := context.Background()

	// Should return empty string when not set
	address, err := service.GetBusinessAddress(ctx)
	require.NoError(t, err)
	assert.Empty(t, address)
}

func TestSystemService_GetBusinessAddress_Custom(t *testing.T) {
	_, service := setupSystemServiceTestDB(t)
	ctx := context.Background()

	// Set custom address
	err := service.UpdateSettings(ctx, &models.PharmacySettings{
		Name:    "Test Pharmacy",
		Address: "123 Custom St",
		Email:   "test@example.com",
	}, 1, "admin", "127.0.0.1")
	require.NoError(t, err)

	// Get business address
	address, err := service.GetBusinessAddress(ctx)
	require.NoError(t, err)
	assert.Equal(t, "123 Custom St", address)
}
