package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

func setupSystemSettingTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables
	err = db.AutoMigrate(&models.SystemSetting{})
	require.NoError(t, err)

	return db
}

func TestNewSystemSettingRepository(t *testing.T) {
	db := setupSystemSettingTestDB(t)

	repo := NewSystemSettingRepository(db)
	assert.NotNil(t, repo)
}

func TestSystemSettingRepository_GetByKey(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()

	// Test not found
	_, err := repo.GetByKey(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)

	// Test invalid input
	_, err = repo.GetByKey(ctx, "")
	assert.ErrorIs(t, err, ErrInvalidInput)

	// Create test setting
	setting := &models.SystemSetting{
		Key:   "test.setting",
		Value: "test value",
	}
	require.NoError(t, db.Create(setting).Error)

	// Test found
	result, err := repo.GetByKey(ctx, "test.setting")
	require.NoError(t, err)
	assert.Equal(t, "test.setting", result.Key)
	assert.Equal(t, "test value", result.Value)
}

func TestSystemSettingRepository_GetAll(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()

	// Create test settings
	settings := []models.SystemSetting{
		{Key: "setting1", Value: "value1"},
		{Key: "setting2", Value: "value2"},
		{Key: "pharmacy.name", Value: "Test Pharmacy"},
	}
	for _, s := range settings {
		require.NoError(t, db.Create(&s).Error)
	}

	// Get all
	result, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, result, 3)

	// Check ordering (should be ordered by key)
	assert.Equal(t, "pharmacy.name", result[0].Key)
	assert.Equal(t, "setting1", result[1].Key)
	assert.Equal(t, "setting2", result[2].Key)
}

func TestSystemSettingRepository_GetPharmacySettings(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()

	// Create test settings including pharmacy and non-pharmacy
	settings := []models.SystemSetting{
		{Key: "pharmacy.name", Value: "Test Pharmacy"},
		{Key: "pharmacy.address", Value: "123 Test St"},
		{Key: "system.timezone", Value: "UTC"}, // Should not be included
	}
	for _, s := range settings {
		require.NoError(t, db.Create(&s).Error)
	}

	// Get pharmacy settings
	result, err := repo.GetPharmacySettings(ctx)
	require.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify only pharmacy settings are returned
	keys := make([]string, len(result))
	for i, s := range result {
		keys[i] = s.Key
	}
	assert.Contains(t, keys, "pharmacy.name")
	assert.Contains(t, keys, "pharmacy.address")
	assert.NotContains(t, keys, "system.timezone")
}

func TestSystemSettingRepository_SetValue_Create(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()
	updatedBy := uint(1)

	// Create new setting
	err := repo.SetValue(ctx, "new.setting", "new value", updatedBy)
	require.NoError(t, err)

	// Verify it was created
	var setting models.SystemSetting
	err = db.Where("key = ?", "new.setting").First(&setting).Error
	require.NoError(t, err)
	assert.Equal(t, "new.setting", setting.Key)
	assert.Equal(t, "new value", setting.Value)
	assert.Equal(t, &updatedBy, setting.UpdatedBy)
}

func TestSystemSettingRepository_SetValue_Update(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()
	updatedBy := uint(1)

	// Create initial setting
	setting := &models.SystemSetting{
		Key:   "test.setting",
		Value: "initial value",
	}
	require.NoError(t, db.Create(setting).Error)

	// Update using SetValue
	err := repo.SetValue(ctx, "test.setting", "updated value", updatedBy+1)
	require.NoError(t, err)

	// Verify it was updated
	var result models.SystemSetting
	err = db.Where("key = ?", "test.setting").First(&result).Error
	require.NoError(t, err)
	assert.Equal(t, "updated value", result.Value)
	assert.Equal(t, uint(2), *result.UpdatedBy)
}

func TestSystemSettingRepository_SetValue_InvalidInput(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()

	// Test empty key
	err := repo.SetValue(ctx, "", "value", 1)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestSystemSettingRepository_UpdateValue(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()
	updatedBy := uint(1)

	// Create initial setting
	setting := &models.SystemSetting{
		Key:   "test.setting",
		Value: "initial value",
	}
	require.NoError(t, db.Create(setting).Error)

	// Update
	err := repo.UpdateValue(ctx, "test.setting", "updated value", updatedBy)
	require.NoError(t, err)

	// Verify
	var result models.SystemSetting
	err = db.Where("key = ?", "test.setting").First(&result).Error
	require.NoError(t, err)
	assert.Equal(t, "updated value", result.Value)
	assert.Equal(t, &updatedBy, result.UpdatedBy)
}

func TestSystemSettingRepository_UpdateValue_NotFound(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()

	// Update non-existent setting
	err := repo.UpdateValue(ctx, "nonexistent", "value", 1)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestSystemSettingRepository_UpdateValue_InvalidInput(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()

	// Test empty key
	err := repo.UpdateValue(ctx, "", "value", 1)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestSystemSettingRepository_UpdateMultiple(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()
	updatedBy := uint(1)

	// Create initial settings
	settings := []models.SystemSetting{
		{Key: "pharmacy.name", Value: "Old Name"},
		{Key: "pharmacy.address", Value: "Old Address"},
	}
	for _, s := range settings {
		require.NoError(t, db.Create(&s).Error)
	}

	// Update multiple
	newSettings := []models.SystemSetting{
		{Key: "pharmacy.name", Value: "New Name", Description: "Pharmacy name", UpdatedBy: &updatedBy},
		{Key: "pharmacy.address", Value: "New Address", Description: "Pharmacy address", UpdatedBy: &updatedBy},
		{Key: "pharmacy.phone", Value: "555-1234", Description: "Phone", UpdatedBy: &updatedBy}, // New setting
	}

	err := repo.UpdateMultiple(ctx, newSettings)
	require.NoError(t, err)

	// Verify all settings
	var results []models.SystemSetting
	err = db.Where("key IN ?", []string{"pharmacy.name", "pharmacy.address", "pharmacy.phone"}).Find(&results).Error
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Create a map for easier verification
	resultMap := make(map[string]models.SystemSetting)
	for _, r := range results {
		resultMap[r.Key] = r
	}

	assert.Equal(t, "New Name", resultMap["pharmacy.name"].Value)
	assert.Equal(t, "New Address", resultMap["pharmacy.address"].Value)
	assert.Equal(t, "555-1234", resultMap["pharmacy.phone"].Value)
}

func TestSystemSettingRepository_UpdateMultiple_Empty(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()

	// Test empty slice
	err := repo.UpdateMultiple(ctx, []models.SystemSetting{})
	assert.NoError(t, err)
}

func TestSystemSettingRepository_GetPublicSettings(t *testing.T) {
	db := setupSystemSettingTestDB(t)
	repo := NewSystemSettingRepository(db)
	ctx := context.Background()

	// Create test settings
	settings := []models.SystemSetting{
		{Key: "pharmacy.name", Value: "Test Pharmacy"},
		{Key: "pharmacy.address", Value: "123 Test St"},
		{Key: "pharmacy.phone", Value: "555-1234"},
		{Key: "pharmacy.email", Value: "test@example.com"},
		{Key: "pharmacy.logo_url", Value: "https://example.com/logo.png"}, // Should not be in public
		{Key: "system.secret", Value: "secret123"}, // Should not be included
	}
	for _, s := range settings {
		require.NoError(t, db.Create(&s).Error)
	}

	// Get public settings
	result, err := repo.GetPublicSettings(ctx)
	require.NoError(t, err)

	assert.Equal(t, "Test Pharmacy", result.BusinessName)
	assert.Equal(t, "123 Test St", result.Address)
	assert.Equal(t, "555-1234", result.Phone)
	assert.Equal(t, "test@example.com", result.Email)
	// LogoURL should not be present in PublicSettings
}
