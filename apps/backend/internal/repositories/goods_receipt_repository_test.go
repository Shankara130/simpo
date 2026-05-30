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

var testGoodsReceiptCounter = 0

func setupGoodsReceiptTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.GoodsReceipt{}, &models.PurchaseInvoice{}, &models.Supplier{}, &models.Branch{})
	require.NoError(t, err)

	return db
}

func createTestPurchaseInvoiceForGoodsReceipt(t *testing.T, db *gorm.DB) *models.PurchaseInvoice {
	testGoodsReceiptCounter++
	createdBy := uint(1)
	branchID := uint(1)

	supplier := &models.Supplier{
		Name:          fmt.Sprintf("Test Supplier %d", testGoodsReceiptCounter),
		ContactPerson: "John Doe",
		Phone:         "555-1234",
		Email:         fmt.Sprintf("supplier%d@example.com", testGoodsReceiptCounter),
		Address:       "123 Supplier St",
		CreatedBy:     &createdBy,
		UpdatedBy:     &createdBy,
	}
	err := db.Create(supplier).Error
	require.NoError(t, err)

	invoice := &models.PurchaseInvoice{
		InvoiceNumber: fmt.Sprintf("INV-%d", testGoodsReceiptCounter),
		InvoiceDate:   time.Now().UTC(),
		SupplierID:    supplier.ID,
		TotalAmount:   1000000.00,
		PaymentStatus: "unpaid",
		BranchID:      branchID,
		CreatedBy:     &createdBy,
		UpdatedBy:     &createdBy,
	}
	err = db.Create(invoice).Error
	require.NoError(t, err)

	return invoice
}

// TestGoodsReceiptRepository_Create tests creating a new goods receipt
// Story 10.3: Verify goods receipt creation
func TestGoodsReceiptRepository_Create(t *testing.T) {
	db := setupGoodsReceiptTestDB(t)
	repo := NewGoodsReceiptRepository(db)
	ctx := context.Background()

	invoice := createTestPurchaseInvoiceForGoodsReceipt(t, db)

	receipt := &models.GoodsReceipt{
		PurchaseInvoiceID: invoice.ID,
		ReceivedDate:      time.Now().UTC(),
		ReceivedBy:        1,
		Notes:             "All items received in good condition",
		BranchID:          1,
	}

	err := repo.Create(ctx, receipt)
	assert.NoError(t, err)
	assert.NotZero(t, receipt.ID)
	assert.NotZero(t, receipt.CreatedAt)
}

// TestGoodsReceiptRepository_CreateNilReceipt tests error handling for nil receipt
func TestGoodsReceiptRepository_CreateNilReceipt(t *testing.T) {
	db := setupGoodsReceiptTestDB(t)
	repo := NewGoodsReceiptRepository(db)
	ctx := context.Background()

	err := repo.Create(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestGoodsReceiptRepository_CreateMissingInvoiceID tests validation for required purchase invoice ID
func TestGoodsReceiptRepository_CreateMissingInvoiceID(t *testing.T) {
	db := setupGoodsReceiptTestDB(t)
	repo := NewGoodsReceiptRepository(db)
	ctx := context.Background()

	receipt := &models.GoodsReceipt{
		ReceivedBy: 1,
		BranchID:   1,
	}

	err := repo.Create(ctx, receipt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "purchase invoice ID is required")
}

// TestGoodsReceiptRepository_CreateMissingReceivedBy tests validation for required received by
func TestGoodsReceiptRepository_CreateMissingReceivedBy(t *testing.T) {
	db := setupGoodsReceiptTestDB(t)
	repo := NewGoodsReceiptRepository(db)
	ctx := context.Background()

	invoice := createTestPurchaseInvoiceForGoodsReceipt(t, db)

	receipt := &models.GoodsReceipt{
		PurchaseInvoiceID: invoice.ID,
		BranchID:          1,
	}

	err := repo.Create(ctx, receipt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "received by user ID is required")
}

// TestGoodsReceiptRepository_GetByID tests retrieving a goods receipt by ID
func TestGoodsReceiptRepository_GetByID(t *testing.T) {
	db := setupGoodsReceiptTestDB(t)
	repo := NewGoodsReceiptRepository(db)
	ctx := context.Background()

	invoice := createTestPurchaseInvoiceForGoodsReceipt(t, db)

	receipt := &models.GoodsReceipt{
		PurchaseInvoiceID: invoice.ID,
		ReceivedDate:      time.Now().UTC(),
		ReceivedBy:        1,
		Notes:             "Test receipt",
		BranchID:          1,
	}
	err := repo.Create(ctx, receipt)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, receipt.ID)
	assert.NoError(t, err)
	assert.Equal(t, receipt.ID, found.ID)
	assert.Equal(t, receipt.PurchaseInvoiceID, found.PurchaseInvoiceID)
	assert.Equal(t, receipt.Notes, found.Notes)
}

// TestGoodsReceiptRepository_GetByIDNotFound tests error handling when receipt not found
func TestGoodsReceiptRepository_GetByIDNotFound(t *testing.T) {
	db := setupGoodsReceiptTestDB(t)
	repo := NewGoodsReceiptRepository(db)
	ctx := context.Background()

	found, err := repo.GetByID(ctx, 99999)
	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "not found")
}

// TestGoodsReceiptRepository_GetByInvoiceID tests retrieving a goods receipt by invoice ID
func TestGoodsReceiptRepository_GetByInvoiceID(t *testing.T) {
	db := setupGoodsReceiptTestDB(t)
	repo := NewGoodsReceiptRepository(db)
	ctx := context.Background()

	invoice := createTestPurchaseInvoiceForGoodsReceipt(t, db)

	receipt := &models.GoodsReceipt{
		PurchaseInvoiceID: invoice.ID,
		ReceivedDate:      time.Now().UTC(),
		ReceivedBy:        1,
		Notes:             "Test receipt",
		BranchID:          1,
	}
	err := repo.Create(ctx, receipt)
	require.NoError(t, err)

	found, err := repo.GetByInvoiceID(ctx, invoice.ID)
	assert.NoError(t, err)
	assert.Equal(t, receipt.ID, found.ID)
	assert.Equal(t, receipt.PurchaseInvoiceID, found.PurchaseInvoiceID)
}

// TestGoodsReceiptRepository_GetByInvoiceIDNotFound tests error handling when receipt not found for invoice
func TestGoodsReceiptRepository_GetByInvoiceIDNotFound(t *testing.T) {
	db := setupGoodsReceiptTestDB(t)
	repo := NewGoodsReceiptRepository(db)
	ctx := context.Background()

	found, err := repo.GetByInvoiceID(ctx, 99999)
	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "not found")
}

// TestGoodsReceiptRepository_List tests listing goods receipts with pagination
func TestGoodsReceiptRepository_List(t *testing.T) {
	db := setupGoodsReceiptTestDB(t)
	repo := NewGoodsReceiptRepository(db)
	ctx := context.Background()

	// Create test receipts
	for i := 0; i < 5; i++ {
		invoice := createTestPurchaseInvoiceForGoodsReceipt(t, db)
		receipt := &models.GoodsReceipt{
			PurchaseInvoiceID: invoice.ID,
			ReceivedDate:      time.Now().UTC(),
			ReceivedBy:        1,
			BranchID:          1,
		}
		err := repo.Create(ctx, receipt)
		require.NoError(t, err)
	}

	// Test list without filters
	filter := &GoodsReceiptFilter{
		Page:  1,
		Limit: 10,
	}

	receipts, total, err := repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, receipts, 5)
}

// TestGoodsReceiptRepository_ListWithBranchFilter tests listing goods receipts filtered by branch
func TestGoodsReceiptRepository_ListWithBranchFilter(t *testing.T) {
	db := setupGoodsReceiptTestDB(t)
	repo := NewGoodsReceiptRepository(db)
	ctx := context.Background()

	// Create test receipts for different branches
	for i := 1; i <= 3; i++ {
		invoice := createTestPurchaseInvoiceForGoodsReceipt(t, db)
		receipt := &models.GoodsReceipt{
			PurchaseInvoiceID: invoice.ID,
			ReceivedDate:      time.Now().UTC(),
			ReceivedBy:        1,
			BranchID:          uint(i),
		}
		err := repo.Create(ctx, receipt)
		require.NoError(t, err)
	}

	// Filter by branch 1
	branchID := uint(1)
	filter := &GoodsReceiptFilter{
		BranchID: &branchID,
		Page:     1,
		Limit:    10,
	}

	receipts, total, err := repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, receipts, 1)
	assert.Equal(t, uint(1), receipts[0].BranchID)
}
