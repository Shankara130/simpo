package repositories

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
)

var testPurchaseInvoiceCounter = 0

func setupPurchaseInvoiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate all required models
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

func createTestBranchForPurchaseInvoice(t *testing.T, db *gorm.DB) *models.Branch {
	testPurchaseInvoiceCounter++
	branch := &models.Branch{
		Name:    fmt.Sprintf("Test Branch %d", testPurchaseInvoiceCounter),
		Address: "123 Test St",
		Phone:   "555-1234",
	}
	err := db.Create(branch).Error
	require.NoError(t, err)
	return branch
}

func createTestSupplierForPurchaseInvoice(t *testing.T, db *gorm.DB) *models.Supplier {
	testPurchaseInvoiceCounter++
	createdBy := uint(1)
	supplier := &models.Supplier{
		Name:          fmt.Sprintf("Test Supplier %d", testPurchaseInvoiceCounter),
		ContactPerson: "John Doe",
		Phone:         "555-1234",
		Email:         fmt.Sprintf("supplier%d@example.com", testPurchaseInvoiceCounter),
		Address:       "123 Supplier St",
		CreatedBy:     &createdBy,
		UpdatedBy:     &createdBy,
	}
	err := db.Create(supplier).Error
	require.NoError(t, err)
	return supplier
}

func createTestProductForPurchaseInvoice(t *testing.T, db *gorm.DB, branchID uint) *models.Product {
	testPurchaseInvoiceCounter++
	expiryDate := time.Now().AddDate(0, 6, 0)
	price := "15000.00"
	product := &models.Product{
		SKU:        fmt.Sprintf("PROD-%d", testPurchaseInvoiceCounter),
		Name:       fmt.Sprintf("Test Product %d", testPurchaseInvoiceCounter),
		StockQty:   100,
		Price:      price,
		ExpiryDate: &expiryDate,
		BranchID:   branchID,
	}
	err := db.Create(product).Error
	require.NoError(t, err)
	return product
}

func createTestPurchaseInvoice(t *testing.T, db *gorm.DB, supplierID, branchID uint) *models.PurchaseInvoice {
	testPurchaseInvoiceCounter++
	createdBy := uint(1)
	invoice := &models.PurchaseInvoice{
		InvoiceNumber: fmt.Sprintf("INV-%d", testPurchaseInvoiceCounter),
		InvoiceDate:   time.Now(),
		SupplierID:    supplierID,
		BranchID:      branchID,
		TotalAmount:   0,
		PaymentStatus: "unpaid",
		CreatedBy:     &createdBy,
		UpdatedBy:     &createdBy,
	}
	err := db.Create(invoice).Error
	require.NoError(t, err)
	return invoice
}

// TestPurchaseInvoiceRepository_Create tests creating a new purchase invoice
// Story 10.2: Verify invoice creation with audit fields
func TestPurchaseInvoiceRepository_Create(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForPurchaseInvoice(t, db)
	supplier := createTestSupplierForPurchaseInvoice(t, db)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: "INV-TEST-001",
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
		TotalAmount:   1500000.00,
		PaymentStatus: "unpaid",
		Notes:         "Test invoice",
	}

	createdBy := uint(1)
	err := repo.Create(ctx, invoice, createdBy, []models.PurchaseInvoiceItem{})
	assert.NoError(t, err)
	assert.NotZero(t, invoice.ID)
	assert.Equal(t, &createdBy, invoice.CreatedBy)
	assert.Equal(t, &createdBy, invoice.UpdatedBy)
	assert.Equal(t, "unpaid", invoice.PaymentStatus)
}

// TestPurchaseInvoiceRepository_CreateNilInvoice tests error handling for nil invoice
func TestPurchaseInvoiceRepository_CreateNilInvoice(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	err := repo.Create(ctx, nil, 1, []models.PurchaseInvoiceItem{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestPurchaseInvoiceRepository_CreateMissingInvoiceNumber tests validation for required invoice number
func TestPurchaseInvoiceRepository_CreateMissingInvoiceNumber(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	branch := createTestBranchForPurchaseInvoice(t, db)
	supplier := createTestSupplierForPurchaseInvoice(t, db)

	invoice := &models.PurchaseInvoice{
		SupplierID:  supplier.ID,
		BranchID:    branch.ID,
		InvoiceDate: time.Now(),
	}

	err := repo.Create(ctx, invoice, 1, []models.PurchaseInvoiceItem{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invoice number is required")
}

// TestPurchaseInvoiceRepository_CreateDuplicateInvoiceNumber tests duplicate invoice number detection
// Story 10.2, AC1: Invoice number must be unique
func TestPurchaseInvoiceRepository_CreateDuplicateInvoiceNumber(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	branch := createTestBranchForPurchaseInvoice(t, db)
	supplier := createTestSupplierForPurchaseInvoice(t, db)

	// Create first invoice
	invoice1 := &models.PurchaseInvoice{
		InvoiceNumber: "INV-DUP-001",
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	err := repo.Create(ctx, invoice1, 1, []models.PurchaseInvoiceItem{})
	assert.NoError(t, err)

	// Try to create duplicate
	invoice2 := &models.PurchaseInvoice{
		InvoiceNumber: "INV-DUP-001",
		InvoiceDate:   time.Now(),
		SupplierID:    supplier.ID,
		BranchID:      branch.ID,
	}

	err = repo.Create(ctx, invoice2, 1, []models.PurchaseInvoiceItem{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// TestPurchaseInvoiceRepository_CreateInvalidSupplier tests validation for supplier existence
func TestPurchaseInvoiceRepository_CreateInvalidSupplier(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	branch := createTestBranchForPurchaseInvoice(t, db)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: "INV-TEST-001",
		InvoiceDate:   time.Now(),
		SupplierID:    999, // Non-existent supplier
		BranchID:      branch.ID,
	}

	err := repo.Create(ctx, invoice, 1, []models.PurchaseInvoiceItem{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "supplier not found")
}

// TestPurchaseInvoiceRepository_GetByID tests retrieving an invoice by ID
// Story 10.2, AC3: Verify invoice retrieval with relationships
func TestPurchaseInvoiceRepository_GetByID(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForPurchaseInvoice(t, db)
	supplier := createTestSupplierForPurchaseInvoice(t, db)
	created := createTestPurchaseInvoice(t, db, supplier.ID, branch.ID)

	// Test GetByID
	found, err := repo.GetByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.InvoiceNumber, found.InvoiceNumber)
	assert.Equal(t, supplier.ID, found.SupplierID)

	// Test not found
	_, err = repo.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

// TestPurchaseInvoiceRepository_GetByIDZeroID tests validation for zero ID
func TestPurchaseInvoiceRepository_GetByIDZeroID(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 0)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
}

// TestPurchaseInvoiceRepository_Update tests updating an existing invoice
// Story 10.2: Verify update with optimistic locking
func TestPurchaseInvoiceRepository_Update(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForPurchaseInvoice(t, db)
	supplier := createTestSupplierForPurchaseInvoice(t, db)
	created := createTestPurchaseInvoice(t, db, supplier.ID, branch.ID)

	// Update invoice
	created.Notes = "Updated notes"
	created.TotalAmount = 2000000.00

	updatedBy := uint(2)
	err := repo.Update(ctx, created, updatedBy)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated notes", found.Notes)
	assert.Equal(t, 2000000.00, found.TotalAmount)
}

// TestPurchaseInvoiceRepository_UpdateOptimisticLocking tests version conflict detection
// Story 10.2: Apply learnings from Story 10-1 code review patch #5
func TestPurchaseInvoiceRepository_UpdateOptimisticLocking(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForPurchaseInvoice(t, db)
	supplier := createTestSupplierForPurchaseInvoice(t, db)
	created := createTestPurchaseInvoice(t, db, supplier.ID, branch.ID)

	// Modify version in DB directly
	updatedBy := uint(2)
	created.Version = 99 // Wrong version

	err := repo.Update(ctx, created, updatedBy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "was modified by another user")
}

// TestPurchaseInvoiceRepository_Delete tests soft deleting an invoice
// Story 10.2: Verify soft delete with explicit check
func TestPurchaseInvoiceRepository_Delete(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForPurchaseInvoice(t, db)
	supplier := createTestSupplierForPurchaseInvoice(t, db)
	created := createTestPurchaseInvoice(t, db, supplier.ID, branch.ID)

	// Delete invoice
	deletedBy := uint(3)
	err := repo.Delete(ctx, created.ID, deletedBy)
	assert.NoError(t, err)

	// Verify soft delete - should not be found with normal query
	_, err = repo.GetByID(ctx, created.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

// TestPurchaseInvoiceRepository_List tests listing invoices with pagination
// Story 10.2, AC2: Verify filtering and pagination
func TestPurchaseInvoiceRepository_List(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForPurchaseInvoice(t, db)
	supplier := createTestSupplierForPurchaseInvoice(t, db)

	// Create multiple invoices
	for i := 0; i < 5; i++ {
		createTestPurchaseInvoice(t, db, supplier.ID, branch.ID)
	}

	// Test list with pagination
	filter := &PurchaseInvoiceFilter{
		Page:  1,
		Limit: 3,
	}

	invoices, total, err := repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, invoices, 3)
}

// TestPurchaseInvoiceRepository_ListWithSupplierFilter tests filtering by supplier
func TestPurchaseInvoiceRepository_ListWithSupplierFilter(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForPurchaseInvoice(t, db)
	supplier1 := createTestSupplierForPurchaseInvoice(t, db)
	supplier2 := createTestSupplierForPurchaseInvoice(t, db)

	// Create invoices for different suppliers
	createTestPurchaseInvoice(t, db, supplier1.ID, branch.ID)
	createTestPurchaseInvoice(t, db, supplier2.ID, branch.ID)

	// Filter by supplier1
	supplierID := supplier1.ID
	filter := &PurchaseInvoiceFilter{
		SupplierID: &supplierID,
		Page:       1,
		Limit:      10,
	}

	invoices, total, err := repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, invoices, 1)
	assert.Equal(t, supplier1.ID, invoices[0].SupplierID)
}

// TestPurchaseInvoiceRepository_ListWithPaymentStatusFilter tests filtering by payment status
func TestPurchaseInvoiceRepository_ListWithPaymentStatusFilter(t *testing.T) {
	db := setupPurchaseInvoiceTestDB(t)
	repo := NewPurchaseInvoiceRepository(db)
	ctx := context.Background()

	// Setup test data
	branch := createTestBranchForPurchaseInvoice(t, db)
	supplier := createTestSupplierForPurchaseInvoice(t, db)

	// Create invoices with different statuses
	invoice1 := createTestPurchaseInvoice(t, db, supplier.ID, branch.ID)
	invoice1.PaymentStatus = "unpaid"
	db.Save(invoice1)

	invoice2 := createTestPurchaseInvoice(t, db, supplier.ID, branch.ID)
	invoice2.PaymentStatus = "paid"
	db.Save(invoice2)

	// Filter by payment status
	paymentStatus := "unpaid"
	filter := &PurchaseInvoiceFilter{
		PaymentStatus: &paymentStatus,
		Page:          1,
		Limit:         10,
	}

	invoices, total, err := repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, invoices, 1)
	assert.Equal(t, "unpaid", invoices[0].PaymentStatus)
}
