package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Should create in-memory database")

	// Create the supplier_audit_trail table
	err = db.AutoMigrate(&models.SupplierAuditTrail{})
	require.NoError(t, err, "Should migrate supplier audit trail model")

	return db
}

// TestNewSupplierAuditService tests the constructor
func TestNewSupplierAuditService(t *testing.T) {
	// Arrange
	db := setupTestDB(t)

	// Act
	service := NewSupplierAuditService(db)

	// Assert
	assert.NotNil(t, service, "Service should be created")
}

// TestSupplierAuditService_LogSupplierOperation_Success tests successful audit logging
func TestSupplierAuditService_LogSupplierOperation_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	auditLog := &models.SupplierAuditTrail{
		TransactionType:     "supplier_operation",
		EntityType:          "supplier",
		EntityID:            1,
		UserID:              10,
		UserRole:            "PharmacyManager",
		ActionType:          "CREATE",
		ActionDescription:    "Created new supplier PT Pharma Indo",
		Reason:              "New supplier onboarding",
		TransactionAmount:   nil,
		AffectedItemsCount:  1,
		IPAddress:           "192.168.1.100",
		UserAgent:           "Mozilla/5.0",
		BranchID:            1,
	}

	// Act
	err := service.LogSupplierOperation(ctx, auditLog)

	// Assert
	assert.NoError(t, err, "Should log supplier operation successfully")
	assert.NotZero(t, auditLog.ID, "Audit log should have ID assigned")
	assert.NotZero(t, auditLog.CreatedAt, "Created timestamp should be set")

	// Verify in database
	var saved models.SupplierAuditTrail
	result := db.First(&saved, auditLog.ID)
	assert.NoError(t, result.Error, "Should retrieve saved audit log")
	assert.Equal(t, "supplier_operation", saved.TransactionType)
	assert.Equal(t, "CREATE", saved.ActionType)
	assert.Equal(t, uint(10), saved.UserID)
}

// TestSupplierAuditService_LogSupplierOperation_WithAmount tests audit logging with transaction amount
func TestSupplierAuditService_LogSupplierOperation_WithAmount(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	amount := 5000000.00
	auditLog := &models.SupplierAuditTrail{
		TransactionType:     "payment",
		EntityType:          "supplier_payment",
		EntityID:            5,
		UserID:              10,
		UserRole:            "FinanceManager",
		ActionType:          "PROCESS",
		ActionDescription:    "Processed payment to supplier",
		TransactionAmount:   &amount,
		AffectedItemsCount:  3,
		BranchID:            1,
	}

	// Act
	err := service.LogSupplierOperation(ctx, auditLog)

	// Assert
	assert.NoError(t, err, "Should log payment with amount successfully")

	var saved models.SupplierAuditTrail
	db.First(&saved, auditLog.ID)
	assert.NotNil(t, saved.TransactionAmount, "Transaction amount should be saved")
	assert.Equal(t, amount, *saved.TransactionAmount)
}

// TestSupplierAuditService_QueryAuditTrail_AllFilters tests query with all filters
func TestSupplierAuditService_QueryAuditTrail_AllFilters(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	// Create test data
	testLogs := []models.SupplierAuditTrail{
		{
			TransactionType:    "supplier_operation",
			EntityType:        "supplier",
			EntityID:          1,
			UserID:            10,
			UserRole:          "PharmacyManager",
			ActionType:        "CREATE",
			ActionDescription:  "Created supplier 1",
			BranchID:          1,
		},
		{
			TransactionType:    "payment",
			EntityType:        "supplier_payment",
			EntityID:          5,
			UserID:            10,
			UserRole:          "FinanceManager",
			ActionType:        "PROCESS",
			ActionDescription:  "Processed payment",
			BranchID:          1,
		},
		{
			TransactionType:    "supplier_operation",
			EntityType:        "supplier",
			EntityID:          2,
			UserID:            20,
			UserRole:          "PharmacyManager",
			ActionType:        "UPDATE",
			ActionDescription:  "Updated supplier 2",
			BranchID:          2,
		},
	}

	for _, log := range testLogs {
		err := service.LogSupplierOperation(ctx, &log)
		require.NoError(t, err)
	}

	// First verify entries exist in database
	var allCount int64
	db.Model(&models.SupplierAuditTrail{}).Count(&allCount)
	require.Equal(t, int64(3), allCount, "Should have 3 entries in database")

	// Test 1: Query WITHOUT transaction type filter (should return all 3)
	page := 1
	limit := 10
	startDate := time.Now().UTC().Add(-48 * time.Hour)
	endDate := time.Now().UTC().Add(1 * time.Hour)

	requestNoFilter := &SupplierAuditQueryRequest{
		StartDate: &startDate,
		EndDate:   &endDate,
		Page:      &page,
		Limit:     &limit,
	}

	response, err := service.QueryAuditTrail(ctx, requestNoFilter)
	assert.NoError(t, err, "Should query without filters")
	assert.Len(t, response.Data, 3, "Should return all 3 entries without filter")

	// Test 2: Query WITH transaction type filter (should return 2 supplier_operation)
	supplierOpType := "supplier_operation"
	requestWithFilter := &SupplierAuditQueryRequest{
		StartDate:       &startDate,
		EndDate:         &endDate,
		TransactionType: &supplierOpType,
		Page:            &page,
		Limit:           &limit,
	}

	response, err = service.QueryAuditTrail(ctx, requestWithFilter)
	assert.NoError(t, err, "Should query with transaction type filter")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Len(t, response.Data, 2, "Should return 2 supplier_operation entries")
	assert.Equal(t, int64(2), response.Pagination.Total, "Total count should be 2")
	if len(response.Data) > 0 {
		assert.Equal(t, "supplier_operation", response.Data[0].TransactionType)
	}
}

// TestSupplierAuditService_QueryAuditTrail_Pagination tests pagination
func TestSupplierAuditService_QueryAuditTrail_Pagination(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	// Create 25 test entries
	for i := 1; i <= 25; i++ {
		log := &models.SupplierAuditTrail{
			TransactionType:    "supplier_operation",
			EntityType:        "supplier",
			EntityID:          uint(i),
			UserID:            10,
			UserRole:          "PharmacyManager",
			ActionType:        "CREATE",
			ActionDescription:  "Test entry",
			BranchID:          1,
		}
		err := service.LogSupplierOperation(ctx, log)
		require.NoError(t, err)
	}

	// Test first page
	page1 := 1
	limit := 10
	request := &SupplierAuditQueryRequest{
		Page:  &page1,
		Limit: &limit,
	}

	response, err := service.QueryAuditTrail(ctx, request)

	assert.NoError(t, err)
	assert.Len(t, response.Data, 10, "Should return 10 entries on page 1")
	assert.Equal(t, int64(25), response.Pagination.Total, "Total should be 25")
	assert.Equal(t, 3, response.Pagination.TotalPages, "Should have 3 pages")

	// Test second page
	page2 := 2
	request.Page = &page2

	response2, err := service.QueryAuditTrail(ctx, request)

	assert.NoError(t, err)
	assert.Len(t, response2.Data, 10, "Should return 10 entries on page 2")

	// Test third page (partial)
	page3 := 3
	request.Page = &page3

	response3, err := service.QueryAuditTrail(ctx, request)

	assert.NoError(t, err)
	assert.Len(t, response3.Data, 5, "Should return 5 entries on page 3")
}

// TestSupplierAuditService_QueryAuditTrail_EntityFilter tests filtering by entity
func TestSupplierAuditService_QueryAuditTrail_EntityFilter(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	// Create entries for different entities
	entries := []models.SupplierAuditTrail{
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 1, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create supplier 1", BranchID: 1},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 1, UserID: 10, UserRole: "Admin", ActionType: "UPDATE", ActionDescription: "Update supplier 1", BranchID: 1},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 2, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create supplier 2", BranchID: 1},
	}

	for _, entry := range entries {
		err := service.LogSupplierOperation(ctx, &entry)
		require.NoError(t, err)
	}

	entityID := uint(1)
	entityType := "supplier"
	page := 1
	limit := 10

	request := &SupplierAuditQueryRequest{
		EntityID:  &entityID,
		EntityType: &entityType,
		Page:      &page,
		Limit:     &limit,
	}

	// Act
	response, err := service.QueryAuditTrail(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, response.Data, 2, "Should return 2 entries for entity ID 1")
	for _, entry := range response.Data {
		assert.Equal(t, uint(1), entry.EntityID, "All entries should have entity ID 1")
	}
}

// TestSupplierAuditService_ExportAuditTrail_CSV tests export functionality
func TestSupplierAuditService_ExportAuditTrail_CSV(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	// Create test data
	testLog := &models.SupplierAuditTrail{
		TransactionType:    "payment",
		EntityType:        "supplier_payment",
		EntityID:          5,
		UserID:            10,
		UserRole:          "FinanceManager",
		ActionType:        "PROCESS",
		ActionDescription:  "Processed payment",
		BranchID:          1,
	}
	err := service.LogSupplierOperation(ctx, testLog)
	require.NoError(t, err)

	now := time.Now().UTC()
	startDate := now.Add(-24 * time.Hour)

	request := &SupplierAuditExportRequest{
		StartDate:       startDate,
		EndDate:         now,
		Format:          "csv",
		TransactionType: strPtr("payment"),
	}

	// Act
	filePath, err := service.ExportAuditTrail(ctx, request)

	// Assert
	assert.NoError(t, err, "Should export audit trail successfully")
	assert.Contains(t, filePath, ".csv", "File path should have .csv extension")
	assert.Contains(t, filePath, "supplier-audit-trail", "File name should contain identifier")
}

// TestSupplierAuditService_GetAuditByEntityID tests retrieval by entity
func TestSupplierAuditService_GetAuditByEntityID(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	// Create test entries
	entries := []models.SupplierAuditTrail{
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 1, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create", BranchID: 1},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 1, UserID: 10, UserRole: "Admin", ActionType: "UPDATE", ActionDescription: "Update", BranchID: 1},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 2, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create", BranchID: 1},
	}

	for _, entry := range entries {
		err := service.LogSupplierOperation(ctx, &entry)
		require.NoError(t, err)
	}

	// Act
	audits, err := service.GetAuditByEntityID(ctx, "supplier", 1)

	// Assert
	assert.NoError(t, err, "Should retrieve audit by entity ID")
	assert.Len(t, audits, 2, "Should return 2 entries for entity ID 1")
	assert.Equal(t, "supplier", audits[0].EntityType)
	assert.Equal(t, uint(1), audits[0].EntityID)
}

// TestSupplierAuditService_GetAuditByUserID tests retrieval by user
func TestSupplierAuditService_GetAuditByUserID(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	// Create entries for different users
	entries := []models.SupplierAuditTrail{
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 1, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create", BranchID: 1},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 2, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create", BranchID: 1},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 3, UserID: 20, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create", BranchID: 1},
	}

	for _, entry := range entries {
		err := service.LogSupplierOperation(ctx, &entry)
		require.NoError(t, err)
	}

	// Query without date filters to get all entries for user 10
	startDate := time.Time{} // Zero time means no start date filter
	endDate := time.Now().UTC().Add(1 * time.Hour) // Future timestamp to ensure we capture all

	// Act
	audits, err := service.GetAuditByUserID(ctx, 10, startDate, endDate)

	// Assert
	assert.NoError(t, err, "Should retrieve audit by user ID")
	assert.Len(t, audits, 2, "Should return 2 entries for user ID 10")
	for _, audit := range audits {
		assert.Equal(t, uint(10), audit.UserID, "All entries should be from user 10")
	}
}

// TestSupplierAuditService_QueryAuditTrail_UserFilter tests filtering by user ID
func TestSupplierAuditService_QueryAuditTrail_UserFilter(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	// Create entries for different users
	entries := []models.SupplierAuditTrail{
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 1, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create", BranchID: 1},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 2, UserID: 20, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create", BranchID: 1},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 3, UserID: 10, UserRole: "Admin", ActionType: "UPDATE", ActionDescription: "Update", BranchID: 1},
	}

	for _, entry := range entries {
		err := service.LogSupplierOperation(ctx, &entry)
		require.NoError(t, err)
	}

	userID := uint(10)
	page := 1
	limit := 10

	request := &SupplierAuditQueryRequest{
		UserID: &userID,
		Page:   &page,
		Limit:  &limit,
	}

	// Act
	response, err := service.QueryAuditTrail(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, response.Data, 2, "Should return 2 entries for user ID 10")
	for _, entry := range response.Data {
		assert.Equal(t, uint(10), entry.UserID, "All entries should be from user 10")
	}
}

// TestSupplierAuditService_QueryAuditTrail_BranchFilter tests filtering by branch
func TestSupplierAuditService_QueryAuditTrail_BranchFilter(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	// Create entries for different branches
	entries := []models.SupplierAuditTrail{
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 1, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create", BranchID: 1},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 2, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create", BranchID: 2},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 3, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Create", BranchID: 1},
	}

	for _, entry := range entries {
		err := service.LogSupplierOperation(ctx, &entry)
		require.NoError(t, err)
	}

	branchID := uint(1)
	page := 1
	limit := 10

	request := &SupplierAuditQueryRequest{
		BranchID: &branchID,
		Page:     &page,
		Limit:    &limit,
	}

	// Act
	response, err := service.QueryAuditTrail(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, response.Data, 2, "Should return 2 entries for branch ID 1")
	for _, entry := range response.Data {
		assert.Equal(t, uint(1), entry.BranchID, "All entries should be from branch 1")
	}
}

// TestSupplierAuditService_QueryAuditTrail_EmptyResults tests query with no matching results
func TestSupplierAuditService_QueryAuditTrail_EmptyResults(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	// Don't create any test data - database is empty

	page := 1
	limit := 10
	transactionType := "payment"

	request := &SupplierAuditQueryRequest{
		TransactionType: &transactionType,
		Page:            &page,
		Limit:           &limit,
	}

	// Act
	response, err := service.QueryAuditTrail(ctx, request)

	// Assert
	assert.NoError(t, err, "Should not error on empty results")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Len(t, response.Data, 0, "Should return empty data array")
	assert.Equal(t, int64(0), response.Pagination.Total, "Total should be 0")
}

// TestSupplierAuditService_ExportAuditTrail_PDF tests PDF format export
func TestSupplierAuditService_ExportAuditTrail_PDF(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	now := time.Now().UTC()
	startDate := now.Add(-24 * time.Hour)

	request := &SupplierAuditExportRequest{
		StartDate: startDate,
		EndDate:   now,
		Format:    "pdf",
	}

	// Act
	filePath, err := service.ExportAuditTrail(ctx, request)

	// Assert
	assert.NoError(t, err, "Should export audit trail in PDF format")
	assert.Contains(t, filePath, ".pdf", "File path should have .pdf extension")
}

// TestSupplierAuditService_AuditTrailOrdering tests that results are ordered by created_at DESC
func TestSupplierAuditService_AuditTrailOrdering(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	// Create entries with specific timestamps
	now := time.Now().UTC()
	older := now.Add(-2 * time.Hour)
	oldest := now.Add(-4 * time.Hour)

	entries := []models.SupplierAuditTrail{
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 1, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "First", BranchID: 1, CreatedAt: oldest},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 2, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Second", BranchID: 1, CreatedAt: older},
		{TransactionType: "supplier_operation", EntityType: "supplier", EntityID: 3, UserID: 10, UserRole: "Admin", ActionType: "CREATE", ActionDescription: "Third", BranchID: 1, CreatedAt: now},
	}

	for _, entry := range entries {
		err := service.LogSupplierOperation(ctx, &entry)
		require.NoError(t, err)
	}

	page := 1
	limit := 10
	request := &SupplierAuditQueryRequest{
		Page:  &page,
		Limit: &limit,
	}

	// Act
	response, err := service.QueryAuditTrail(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, response.Data, 3, "Should return all 3 entries")
	// Verify ordering (newest first)
	assert.True(t, response.Data[0].CreatedAt.After(response.Data[1].CreatedAt), "First entry should be newer than second")
	assert.True(t, response.Data[1].CreatedAt.After(response.Data[2].CreatedAt), "Second entry should be newer than third")
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}

// TestSupplierAuditService_AppendOnlyCompliance tests that audit trail is append-only
// This ensures Badan POM compliance - no UPDATE or DELETE operations on audit logs
func TestSupplierAuditService_AppendOnlyCompliance(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	service := NewSupplierAuditService(db)
	ctx := context.Background()

	// Create initial audit log
	auditLog := &models.SupplierAuditTrail{
		TransactionType:    "supplier_operation",
		EntityType:        "supplier",
		EntityID:          1,
		UserID:            10,
		UserRole:          "PharmacyManager",
		ActionType:        "CREATE",
		ActionDescription:  "Initial creation",
		BranchID:          1,
	}

	err := service.LogSupplierOperation(ctx, auditLog)
	require.NoError(t, err)

	originalID := auditLog.ID
	originalDescription := auditLog.ActionDescription

	// Verify the service interface doesn't expose Update/Delete methods
	// This is enforced at compile time - the SupplierAuditService interface only has:
	// - LogSupplierOperation (Create only)
	// - QueryAuditTrail (Read only)
	// - ExportAuditTrail (Read only)
	// - GetAuditByEntityID (Read only)
	// - GetAuditByUserID (Read only)
	// No Update or Delete methods exist

	// For append-only compliance, modifications should create NEW audit entries
	// instead of updating existing ones
	modificationLog := &models.SupplierAuditTrail{
		TransactionType:    "supplier_operation",
		EntityType:        "supplier",
		EntityID:          1,
		UserID:            10,
		UserRole:          "PharmacyManager",
		ActionType:        "UPDATE",
		ActionDescription:  "Modified supplier information",
		Reason:            "Correction of supplier details",
		BranchID:          1,
	}

	err = service.LogSupplierOperation(ctx, modificationLog)
	assert.NoError(t, err, "Should create new audit entry for modifications")

	// Verify we now have 2 entries (append-only behavior)
	var count int64
	db.Model(&models.SupplierAuditTrail{}).Count(&count)
	assert.Equal(t, int64(2), count, "Should have 2 audit entries (original + modification)")

	// Verify original entry still exists and is unchanged
	var originalEntry models.SupplierAuditTrail
	err = db.First(&originalEntry, originalID).Error
	require.NoError(t, err, "Original audit record should still exist")
	assert.Equal(t, originalDescription, originalEntry.ActionDescription, "Original entry should remain unchanged")

	// Verify the modification entry exists as a separate record
	var allEntries []models.SupplierAuditTrail
	db.Order("created_at ASC").Find(&allEntries)
	assert.Len(t, allEntries, 2, "Should have exactly 2 entries")
	assert.Equal(t, "CREATE", allEntries[0].ActionType, "First entry should be CREATE")
	assert.Equal(t, "UPDATE", allEntries[1].ActionType, "Second entry should be UPDATE")
}
