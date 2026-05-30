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

// mockPurchaseInvoiceAuditService is a mock implementation of AuditService for testing
type mockPurchaseInvoiceAuditService struct{}

func (m *mockPurchaseInvoiceAuditService) LogLoginAttempt(ctx context.Context, entry AuditLogEntry) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogAuthorizationFailure(ctx context.Context, entry AuditLogEntry) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogWhitelistChange(ctx context.Context, adminID uint, adminUsername string, domain string, action models.AuditAction, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogUserDeactivation(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogStockAdjustment(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogBlockedSaleAttempt(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogReportExport(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogSettingsUpdate(ctx context.Context, adminID uint, adminUsername string, changesJSON string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogBackupCreated(ctx context.Context, adminID uint, adminUsername string, backupFile string, size int64, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogBackupRestored(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogBackupDeleted(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogRoleUpdated(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, oldRole string, newRole string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogPermissionGranted(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogPermissionRevoked(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogBranchCreated(ctx context.Context, adminID uint, adminUsername string, branchName string, branchLocation string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogBranchUpdated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, changes string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogBranchDeactivated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, reason string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogSystemStartup(ctx context.Context, systemID string, serverInfo string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogSystemShutdown(ctx context.Context, systemID string, reason string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogMaintenanceModeEnabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogMaintenanceModeDisabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) LogConflictResolution(ctx context.Context, eventType string, transactionID string, originalError string, resolutionType string, resolvedBy string, resolvedAt time.Time, conflictDetails string, ipAddress string) error {
	return nil
}

func (m *mockPurchaseInvoiceAuditService) ResetMetrics() {}

func (m *mockPurchaseInvoiceAuditService) Shutdown(ctx context.Context) error {
	return nil
}

var testPurchaseInvoiceServiceCounter = 0

func setupPurchaseInvoiceServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.Branch{},
		&models.Supplier{},
		&models.Product{},
		&models.PurchaseInvoice{},
		&models.PurchaseInvoiceItem{},
	)
	require.NoError(t, err)

	return db
}

func createTestBranchForInvoiceService(t *testing.T, db *gorm.DB) *models.Branch {
	testPurchaseInvoiceServiceCounter++
	branch := &models.Branch{
		Name:    fmt.Sprintf("Test Branch %d", testPurchaseInvoiceServiceCounter),
		Address: "123 Test St",
		Phone:   "555-1234",
	}
	err := db.Create(branch).Error
	require.NoError(t, err)
	return branch
}

func createTestSupplierForInvoiceService(t *testing.T, db *gorm.DB) *models.Supplier {
	testPurchaseInvoiceServiceCounter++
	createdBy := uint(1)
	supplier := &models.Supplier{
		Name:          fmt.Sprintf("Invoice Test Supplier %d", testPurchaseInvoiceServiceCounter),
		ContactPerson: "Test Contact",
		Phone:         "555-9999",
		Email:         fmt.Sprintf("invoicetest%d@example.com", testPurchaseInvoiceServiceCounter),
		Address:       "123 Invoice Test St",
		CreatedBy:     &createdBy,
		UpdatedBy:     &createdBy,
	}
	err := db.Create(supplier).Error
	require.NoError(t, err)
	return supplier
}

func createTestProductForInvoiceService(t *testing.T, db *gorm.DB, branchID uint) *models.Product {
	testPurchaseInvoiceServiceCounter++
	expiryDate := time.Now().AddDate(0, 6, 0)
	price := "15000.00"
	product := &models.Product{
		SKU:        fmt.Sprintf("PROD-%d", testPurchaseInvoiceServiceCounter),
		Name:       fmt.Sprintf("Test Product %d", testPurchaseInvoiceServiceCounter),
		StockQty:   100,
		Price:      price,
		ExpiryDate: &expiryDate,
		BranchID:   branchID,
	}
	err := db.Create(product).Error
	require.NoError(t, err)
	return product
}

// TestPurchaseInvoiceService_CreatePurchaseInvoice tests creating a new purchase invoice
// Story 10.2, AC1: Verify invoice creation with validation and audit logging
func TestPurchaseInvoiceService_CreatePurchaseInvoice(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)
	product := createTestProductForInvoiceService(t, db, branch.ID)

	// Use a specific invoice number for predictable testing
	invoice := &models.PurchaseInvoice{
		InvoiceNumber: "INV-001",
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	items := []CreatePurchaseInvoiceItemRequest{
		{
			ProductID: product.ID,
			Quantity:  10,
			UnitCost:  15000.00,
		},
	}

	createdBy := uint(1)
	result, err := service.CreatePurchaseInvoice(ctx, invoice, items, createdBy, "127.0.0.1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotZero(t, result.ID)
	assert.Equal(t, "INV-001", result.InvoiceNumber)
	assert.Equal(t, "unpaid", result.PaymentStatus)
	assert.Equal(t, 150000.00, result.TotalAmount) // 10 * 15000
}

// TestPurchaseInvoiceService_CreatePurchaseInvoiceNilInvoice tests error handling for nil invoice
func TestPurchaseInvoiceService_CreatePurchaseInvoiceNilInvoice(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	items := []CreatePurchaseInvoiceItemRequest{}
	_, err := service.CreatePurchaseInvoice(ctx, nil, items, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestPurchaseInvoiceService_CreatePurchaseInvoiceNoItems tests validation for required items
func TestPurchaseInvoiceService_CreatePurchaseInvoiceNoItems(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: fmt.Sprintf("INV-TEST-%d", testPurchaseInvoiceServiceCounter),
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	_, err := service.CreatePurchaseInvoice(ctx, invoice, []CreatePurchaseInvoiceItemRequest{}, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one line item")
}

// TestPurchaseInvoiceService_CreatePurchaseInvoiceMissingInvoiceNumber tests validation for required invoice number
func TestPurchaseInvoiceService_CreatePurchaseInvoiceMissingInvoiceNumber(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)
	product := createTestProductForInvoiceService(t, db, branch.ID)

	invoice := &models.PurchaseInvoice{
		InvoiceDate: time.Now(),
		SupplierID:  supplier.ID,
		BranchID:    branch.ID,
	}

	items := []CreatePurchaseInvoiceItemRequest{
		{ProductID: product.ID, Quantity: 10, UnitCost: 15000.00},
	}

	_, err := service.CreatePurchaseInvoice(ctx, invoice, items, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invoice number is required")
}

// TestPurchaseInvoiceService_CreatePurchaseInvoiceFutureDate tests validation for future invoice date
func TestPurchaseInvoiceService_CreatePurchaseInvoiceFutureDate(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)
	product := createTestProductForInvoiceService(t, db, branch.ID)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: "INV-TEST-001",
		InvoiceDate:   time.Now().AddDate(0, 0, 1), // Tomorrow
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	items := []CreatePurchaseInvoiceItemRequest{
		{ProductID: product.ID, Quantity: 10, UnitCost: 15000.00},
	}

	_, err := service.CreatePurchaseInvoice(ctx, invoice, items, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be in the future")
}

// TestPurchaseInvoiceService_CreatePurchaseInvoiceInvalidSupplier tests validation for supplier existence
func TestPurchaseInvoiceService_CreatePurchaseInvoiceInvalidSupplier(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	branch := createTestBranchForInvoiceService(t, db)
	product := createTestProductForInvoiceService(t, db, branch.ID)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: "INV-TEST-001",
		InvoiceDate:   time.Now(),
		SupplierID:    999, // Non-existent supplier
		BranchID:      branch.ID,
	}

	items := []CreatePurchaseInvoiceItemRequest{
		{ProductID: product.ID, Quantity: 10, UnitCost: 15000.00},
	}

	_, err := service.CreatePurchaseInvoice(ctx, invoice, items, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "supplier not found")
}

// TestPurchaseInvoiceService_CreatePurchaseInvoiceInvalidProduct tests validation for product existence
func TestPurchaseInvoiceService_CreatePurchaseInvoiceInvalidProduct(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: fmt.Sprintf("INV-TEST-%d", testPurchaseInvoiceServiceCounter),
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	items := []CreatePurchaseInvoiceItemRequest{
		{ProductID: 999, Quantity: 10, UnitCost: 15000.00}, // Non-existent product
	}

	_, err := service.CreatePurchaseInvoice(ctx, invoice, items, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product not found")
}

// TestPurchaseInvoiceService_CreatePurchaseInvoiceInvalidQuantity tests validation for quantity
func TestPurchaseInvoiceService_CreatePurchaseInvoiceInvalidQuantity(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)
	product := createTestProductForInvoiceService(t, db, branch.ID)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: fmt.Sprintf("INV-TEST-%d", testPurchaseInvoiceServiceCounter),
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	items := []CreatePurchaseInvoiceItemRequest{
		{ProductID: product.ID, Quantity: 0, UnitCost: 15000.00}, // Invalid quantity
	}

	_, err := service.CreatePurchaseInvoice(ctx, invoice, items, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be greater than 0")
}

// TestPurchaseInvoiceService_CreatePurchaseInvoiceNegativeUnitCost tests validation for negative unit cost
func TestPurchaseInvoiceService_CreatePurchaseInvoiceNegativeUnitCost(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)
	product := createTestProductForInvoiceService(t, db, branch.ID)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: fmt.Sprintf("INV-TEST-%d", testPurchaseInvoiceServiceCounter),
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	items := []CreatePurchaseInvoiceItemRequest{
		{ProductID: product.ID, Quantity: 10, UnitCost: -100.00}, // Negative cost
	}

	_, err := service.CreatePurchaseInvoice(ctx, invoice, items, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unit cost cannot be negative")
}

// TestPurchaseInvoiceService_GetPurchaseInvoiceByID tests retrieving an invoice
// Story 10.2, AC3: Verify invoice retrieval
func TestPurchaseInvoiceService_GetPurchaseInvoiceByID(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)
	product := createTestProductForInvoiceService(t, db, branch.ID)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: fmt.Sprintf("INV-TEST-%d", testPurchaseInvoiceServiceCounter),
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	items := []CreatePurchaseInvoiceItemRequest{
		{ProductID: product.ID, Quantity: 10, UnitCost: 15000.00},
	}

	created, err := service.CreatePurchaseInvoice(ctx, invoice, items, 1, "127.0.0.1")
	require.NoError(t, err)

	// Test get by ID
	found, err := service.GetPurchaseInvoiceByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.InvoiceNumber, found.InvoiceNumber)

	// Test not found
	_, err = service.GetPurchaseInvoiceByID(ctx, 999)
	assert.Error(t, err)
}

// TestPurchaseInvoiceService_ListPurchaseInvoices tests listing invoices
// Story 10.2, AC2: Verify invoice listing with filters
func TestPurchaseInvoiceService_ListPurchaseInvoices(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)
	product := createTestProductForInvoiceService(t, db, branch.ID)

	// Create multiple invoices
	for i := 0; i < 3; i++ {
		invoice := &models.PurchaseInvoice{
			InvoiceNumber: fmt.Sprintf("INV-%03d", i+1),
			InvoiceDate:   time.Now(),
			SupplierID:    supplier.ID,
			BranchID:      branch.ID,
		}
		items := []CreatePurchaseInvoiceItemRequest{
			{ProductID: product.ID, Quantity: 10, UnitCost: 15000.00},
		}
		service.CreatePurchaseInvoice(ctx, invoice, items, 1, "127.0.0.1")
	}

	// Test list all
	invoices, total, err := service.ListPurchaseInvoices(ctx, &PurchaseInvoiceListFilter{
		Page:  1,
		Limit: 10,
	})
	assert.NoError(t, err)
	assert.Len(t, invoices, 3)
	assert.Equal(t, int64(3), total)

	// Note: Search query test skipped - ILIKE is PostgreSQL-specific
	// and doesn't work with SQLite. Search functionality is tested
	// at the integration level with PostgreSQL.
}

// TestPurchaseInvoiceService_UpdatePurchaseInvoice tests updating an invoice
// Story 10.2: Verify invoice update with validation
func TestPurchaseInvoiceService_UpdatePurchaseInvoice(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)
	product := createTestProductForInvoiceService(t, db, branch.ID)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: fmt.Sprintf("INV-TEST-%d", testPurchaseInvoiceServiceCounter),
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	items := []CreatePurchaseInvoiceItemRequest{
		{ProductID: product.ID, Quantity: 10, UnitCost: 15000.00},
	}

	created, err := service.CreatePurchaseInvoice(ctx, invoice, items, 1, "127.0.0.1")
	require.NoError(t, err)

	// Update invoice
	updates := &UpdatePurchaseInvoiceRequest{
		InvoiceNumber: "INV-TEST-001-UPDATED",
		InvoiceDate:   time.Now().Format(time.RFC3339),
		SupplierID:    supplier.ID,
		Notes:         "Updated notes",
		Items: []CreatePurchaseInvoiceItemRequest{
			{ProductID: product.ID, Quantity: 20, UnitCost: 15000.00},
		},
		Reason: "Updating invoice for correction",
	}

	updatedBy := uint(2)
	result, err := service.UpdatePurchaseInvoice(ctx, created.ID, updates, updatedBy, "127.0.0.1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "INV-TEST-001-UPDATED", result.InvoiceNumber)
	assert.Equal(t, 300000.00, result.TotalAmount) // 20 * 15000
}

// TestPurchaseInvoiceService_UpdatePurchaseInvoiceMissingReason tests validation for required reason
func TestPurchaseInvoiceService_UpdatePurchaseInvoiceMissingReason(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)
	product := createTestProductForInvoiceService(t, db, branch.ID)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: fmt.Sprintf("INV-TEST-%d", testPurchaseInvoiceServiceCounter),
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	items := []CreatePurchaseInvoiceItemRequest{
		{ProductID: product.ID, Quantity: 10, UnitCost: 15000.00},
	}

	created, err := service.CreatePurchaseInvoice(ctx, invoice, items, 1, "127.0.0.1")
	require.NoError(t, err)

	// Update without reason
	updates := &UpdatePurchaseInvoiceRequest{
		InvoiceNumber: "INV-TEST-001-UPDATED",
		InvoiceDate:   time.Now().Format(time.RFC3339),
		SupplierID:    supplier.ID,
		Items: []CreatePurchaseInvoiceItemRequest{
			{ProductID: product.ID, Quantity: 10, UnitCost: 15000.00},
		},
		Reason: "",
	}

	_, err = service.UpdatePurchaseInvoice(ctx, created.ID, updates, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reason is required")
}

// TestPurchaseInvoiceService_DeletePurchaseInvoice tests deleting an invoice
// Story 10.2: Verify invoice deletion with audit logging
func TestPurchaseInvoiceService_DeletePurchaseInvoice(t *testing.T) {
	db := setupPurchaseInvoiceServiceTestDB(t)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	auditSvc := &mockPurchaseInvoiceAuditService{}
	service := NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditSvc)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForInvoiceService(t, db)
	supplier := createTestSupplierForInvoiceService(t, db)
	product := createTestProductForInvoiceService(t, db, branch.ID)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: fmt.Sprintf("INV-TEST-%d", testPurchaseInvoiceServiceCounter),
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	items := []CreatePurchaseInvoiceItemRequest{
		{ProductID: product.ID, Quantity: 10, UnitCost: 15000.00},
	}

	created, err := service.CreatePurchaseInvoice(ctx, invoice, items, 1, "127.0.0.1")
	require.NoError(t, err)

	// Delete invoice
	err = service.DeletePurchaseInvoice(ctx, created.ID, 1, "127.0.0.1")
	assert.NoError(t, err)

	// Verify deletion (should return not found)
	_, err = service.GetPurchaseInvoiceByID(ctx, created.ID)
	assert.Error(t, err)
}
