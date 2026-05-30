package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// mockSupplierAuditService is a mock implementation of AuditService for testing
type mockSupplierAuditService struct{}

func (m *mockSupplierAuditService) LogLoginAttempt(ctx context.Context, entry AuditLogEntry) error {
	return nil
}

func (m *mockSupplierAuditService) LogAuthorizationFailure(ctx context.Context, entry AuditLogEntry) error {
	return nil
}

func (m *mockSupplierAuditService) LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogWhitelistChange(ctx context.Context, adminID uint, adminUsername string, domain string, action models.AuditAction, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogUserDeactivation(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogStockAdjustment(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogBlockedSaleAttempt(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogReportExport(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogSettingsUpdate(ctx context.Context, adminID uint, adminUsername string, changesJSON string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogBackupCreated(ctx context.Context, adminID uint, adminUsername string, backupFile string, size int64, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogBackupRestored(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogBackupDeleted(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogRoleUpdated(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, oldRole string, newRole string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogPermissionGranted(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogPermissionRevoked(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogBranchCreated(ctx context.Context, adminID uint, adminUsername string, branchName string, branchLocation string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogBranchUpdated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, changes string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogBranchDeactivated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, reason string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogSystemStartup(ctx context.Context, systemID string, serverInfo string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogSystemShutdown(ctx context.Context, systemID string, reason string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogMaintenanceModeEnabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogMaintenanceModeDisabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) LogConflictResolution(ctx context.Context, eventType string, transactionID string, originalError string, resolutionType string, resolvedBy string, resolvedAt time.Time, conflictDetails string, ipAddress string) error {
	return nil
}

func (m *mockSupplierAuditService) ResetMetrics() {}

func (m *mockSupplierAuditService) Shutdown(ctx context.Context) error {
	return nil
}

var testSupplierServiceCounter = 0

func setupSupplierServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.Supplier{})
	require.NoError(t, err)

	return db
}

func createTestSupplierForService(t *testing.T, db *gorm.DB) *models.Supplier {
	testSupplierServiceCounter++
	createdBy := uint(1)
	supplier := &models.Supplier{
		Name:          fmt.Sprintf("Service Test Supplier %d", testSupplierServiceCounter),
		ContactPerson: "Service Test Person",
		Phone:         "555-9999",
		Email:         fmt.Sprintf("servicetest%d@example.com", testSupplierServiceCounter),
		Address:       "123 Service Test St",
		CreatedBy:     &createdBy,
		UpdatedBy:     &createdBy,
	}
	err := db.Create(supplier).Error
	require.NoError(t, err)
	return supplier
}

// TestSupplierService_CreateSupplier tests creating a new supplier
// Story 10.1, AC1: Verify supplier creation with validation
func TestSupplierService_CreateSupplier(t *testing.T) {
	db := setupSupplierServiceTestDB(t)
	repo := repositories.NewSupplierRepository(db)
	auditSvc := &mockSupplierAuditService{}
	service := NewSupplierService(repo, auditSvc)
	ctx := context.Background()

	supplier := &models.Supplier{
		Name:          "PT. Test Supplier",
		ContactPerson: "Test Contact",
		Phone:         "+62-21-9876-5432",
		Email:         "test@supplier.com",
		Address:       "Test Address",
	}

	createdBy := uint(1)
	result, err := service.CreateSupplier(ctx, supplier, createdBy, "127.0.0.1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotZero(t, result.ID)
	assert.Equal(t, "PT. Test Supplier", result.Name)
	assert.True(t, result.IsActive)
}

// TestSupplierService_CreateSupplierNil tests error handling for nil supplier
func TestSupplierService_CreateSupplierNil(t *testing.T) {
	db := setupSupplierServiceTestDB(t)
	repo := repositories.NewSupplierRepository(db)
	auditSvc := &mockSupplierAuditService{}
	service := NewSupplierService(repo, auditSvc)
	ctx := context.Background()

	_, err := service.CreateSupplier(ctx, nil, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSupplierService_CreateSupplierMissingName tests validation for required name
func TestSupplierService_CreateSupplierMissingName(t *testing.T) {
	db := setupSupplierServiceTestDB(t)
	repo := repositories.NewSupplierRepository(db)
	auditSvc := &mockSupplierAuditService{}
	service := NewSupplierService(repo, auditSvc)
	ctx := context.Background()

	supplier := &models.Supplier{
		Phone: "+62-21-9876-5432",
	}

	_, err := service.CreateSupplier(ctx, supplier, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

// TestSupplierService_CreateSupplierDuplicateName tests duplicate name validation
func TestSupplierService_CreateSupplierDuplicateName(t *testing.T) {
	db := setupSupplierServiceTestDB(t)
	repo := repositories.NewSupplierRepository(db)
	auditSvc := &mockSupplierAuditService{}
	service := NewSupplierService(repo, auditSvc)
	ctx := context.Background()

	// Create first supplier
	supplier1 := &models.Supplier{
		Name: "Duplicate Name",
		Phone: "+62-21-1111-2222",
	}
	_, err := service.CreateSupplier(ctx, supplier1, 1, "127.0.0.1")
	assert.NoError(t, err)

	// Try to create duplicate
	supplier2 := &models.Supplier{
		Name: "Duplicate Name",
		Phone: "+62-21-3333-4444",
	}
	_, err = service.CreateSupplier(ctx, supplier2, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// TestSupplierService_GetSupplierByID tests retrieving a supplier
// Story 10.1, AC1: Verify supplier retrieval
func TestSupplierService_GetSupplierByID(t *testing.T) {
	db := setupSupplierServiceTestDB(t)
	repo := repositories.NewSupplierRepository(db)
	auditSvc := &mockSupplierAuditService{}
	service := NewSupplierService(repo, auditSvc)
	ctx := context.Background()

	// Create test supplier
	created := createTestSupplierForService(t, db)

	// Test get by ID
	found, err := service.GetSupplierByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.Name, found.Name)

	// Test not found
	_, err = service.GetSupplierByID(ctx, 999)
	assert.Error(t, err)
}

// TestSupplierService_ListSuppliers tests listing suppliers
// Story 10.1, AC2: Verify supplier listing with filters
func TestSupplierService_ListSuppliers(t *testing.T) {
	db := setupSupplierServiceTestDB(t)
	repo := repositories.NewSupplierRepository(db)
	auditSvc := &mockSupplierAuditService{}
	service := NewSupplierService(repo, auditSvc)
	ctx := context.Background()

	// Create test suppliers
	for i := 1; i <= 3; i++ {
		createTestSupplierForService(t, db)
	}

	// Test list all
	suppliers, total, err := service.ListSuppliers(ctx, &SupplierListFilter{
		Page:  1,
		Limit: 10,
	})
	assert.NoError(t, err)
	assert.Len(t, suppliers, 3)
	assert.Equal(t, int64(3), total)

	// Test search
	suppliers, total, err = service.ListSuppliers(ctx, &SupplierListFilter{
		SearchQuery: "Service",
		Page:        1,
		Limit:       10,
	})
	assert.NoError(t, err)
	assert.Greater(t, len(suppliers), 0)
}

// TestSupplierService_UpdateSupplier tests updating a supplier
// Story 10.1, AC2: Verify supplier update with validation
func TestSupplierService_UpdateSupplier(t *testing.T) {
	db := setupSupplierServiceTestDB(t)
	repo := repositories.NewSupplierRepository(db)
	auditSvc := &mockSupplierAuditService{}
	service := NewSupplierService(repo, auditSvc)
	ctx := context.Background()

	// Create test supplier
	created := createTestSupplierForService(t, db)

	// Update supplier
	updates := &UpdateSupplierRequest{
		Name:          "Updated Supplier Name",
		ContactPerson: "Updated Contact",
		Phone:         "+62-21-5555-6666",
		Email:         "updated@supplier.com",
		Address:       "Updated Address",
		Reason:        "Updating supplier information for accuracy",
	}

	updatedBy := uint(2)
	result, err := service.UpdateSupplier(ctx, created.ID, updates, updatedBy, "127.0.0.1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Supplier Name", result.Name)
	assert.Equal(t, "Updated Contact", result.ContactPerson)
}

// TestSupplierService_UpdateSupplierMissingReason tests validation for required reason
func TestSupplierService_UpdateSupplierMissingReason(t *testing.T) {
	db := setupSupplierServiceTestDB(t)
	repo := repositories.NewSupplierRepository(db)
	auditSvc := &mockSupplierAuditService{}
	service := NewSupplierService(repo, auditSvc)
	ctx := context.Background()

	// Create test supplier
	created := createTestSupplierForService(t, db)

	// Update without reason
	updates := &UpdateSupplierRequest{
		Name:   "Updated Name",
		Phone:  "+62-21-5555-6666",
		Reason: "",
	}

	_, err := service.UpdateSupplier(ctx, created.ID, updates, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reason is required")
}

// TestSupplierService_DeactivateSupplier tests deactivating a supplier
// Story 10.1, AC3: Verify supplier deactivation
func TestSupplierService_DeactivateSupplier(t *testing.T) {
	db := setupSupplierServiceTestDB(t)
	repo := repositories.NewSupplierRepository(db)
	auditSvc := &mockSupplierAuditService{}
	service := NewSupplierService(repo, auditSvc)
	ctx := context.Background()

	// Create test supplier
	created := createTestSupplierForService(t, db)

	// Deactivate supplier
	err := service.DeactivateSupplier(ctx, created.ID, "Supplier no longer in business", 1, "127.0.0.1")
	assert.NoError(t, err)

	// Verify deactivation (should return not found)
	_, err = service.GetSupplierByID(ctx, created.ID)
	assert.Error(t, err)
}

// TestSupplierService_DeactivateSupplierMissingReason tests validation for required reason
func TestSupplierService_DeactivateSupplierMissingReason(t *testing.T) {
	db := setupSupplierServiceTestDB(t)
	repo := repositories.NewSupplierRepository(db)
	auditSvc := &mockSupplierAuditService{}
	service := NewSupplierService(repo, auditSvc)
	ctx := context.Background()

	// Create test supplier
	created := createTestSupplierForService(t, db)

	// Deactivate without reason
	err := service.DeactivateSupplier(ctx, created.ID, "", 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reason is required")
}
