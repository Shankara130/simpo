package handlers

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
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// TestGoodsReceipt_ProcessGoodsReceipt_BasicFlow tests the basic goods receipt processing flow
// Story 10.3: Integration test for goods receipt processing with database
func TestGoodsReceipt_ProcessGoodsReceipt_BasicFlow(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all required models
	err = db.AutoMigrate(
		&models.Branch{},
		&models.Supplier{},
		&models.Product{},
		&models.PurchaseInvoice{},
		&models.PurchaseInvoiceItem{},
		&models.GoodsReceipt{},
	)
	require.NoError(t, err)

	// Create repositories
	branchRepo := repositories.NewBranchRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	productRepo := repositories.NewProductRepository(db)
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(db)
	goodsReceiptRepo := repositories.NewGoodsReceiptRepository(db)

	// Create services with minimal dependencies
	purchaseInvoiceService := services.NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, nil)
	goodsReceiptService := services.NewGoodsReceiptService(db, goodsReceiptRepo, purchaseInvoiceRepo, productRepo, nil, nil, nil)

	// Setup test data
	ctx := context.Background()

	// Create branch
	branch := &models.Branch{
		Name:    "Test Branch",
		Address: "123 Test St",
	}
	err = branchRepo.Create(ctx, branch)
	require.NoError(t, err)

	// Create supplier
	createdBy := uint(1)
	supplier := &models.Supplier{
		Name:          "Test Supplier",
		ContactPerson: "John Doe",
		Phone:         "555-1234",
		Email:         "supplier@test.com",
		Address:       "123 Supplier St",
		CreatedBy:     &createdBy,
		UpdatedBy:     &createdBy,
	}
	err = supplierRepo.Create(ctx, supplier, createdBy)
	require.NoError(t, err)

	// Create product with initial stock
	product := &models.Product{
		SKU:             "TEST-001",
		Name:            "Test Product",
		Description:     "Test product description",
		StockQty:        50,
		Price:           "15000.00",
		CostPrice:       stringPtr("12000.00"),
		ReorderThreshold: 10,
		BranchID:        branch.ID,
	}
	err = productRepo.Create(ctx, product)
	require.NoError(t, err)

	// Create purchase invoice with items
	invoice := &models.PurchaseInvoice{
		InvoiceNumber: "INV-TEST-001",
		InvoiceDate:   time.Now().UTC(),
		SupplierID:    supplier.ID,
		TotalAmount:   1000000.00,
		PaymentStatus: "unpaid",
		ReceiptStatus:  "pending",
		BranchID:      branch.ID,
		CreatedBy:     &createdBy,
		UpdatedBy:     &createdBy,
	}

	// Create invoice items directly (for this test, we'll use repository pattern)
	items := []services.CreatePurchaseInvoiceItemRequest{
		{
			ProductID: product.ID,
			Quantity:  100,
			UnitCost:  15000.00,
		},
	}

	createdInvoice, err := purchaseInvoiceService.CreatePurchaseInvoice(ctx, invoice, items, createdBy, "127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, createdInvoice)

	// Process goods receipt
	receivedBy := uint(1)
	receipt, err := goodsReceiptService.ProcessGoodsReceipt(ctx, createdInvoice.ID, receivedBy, "All items received in good condition", branch.ID)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)
	assert.NotZero(t, receipt.ID)
	assert.Equal(t, createdInvoice.ID, receipt.PurchaseInvoiceID)
	assert.Equal(t, "All items received in good condition", receipt.Notes)

	// Verify stock was updated - original stock was 50, added 100, so new stock should be 150
	updatedProduct, err := productRepo.GetByID(ctx, product.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(150), updatedProduct.StockQty)

	// Verify cost price was updated to latest purchase cost (15000.00)
	assert.NotNil(t, updatedProduct.CostPrice)
	assert.Equal(t, "15000.00", *updatedProduct.CostPrice)

	// Try to process goods receipt again - should fail (already received)
	_, err = goodsReceiptService.ProcessGoodsReceipt(ctx, createdInvoice.ID, receivedBy, "Duplicate receipt", branch.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already been received")
}

// TestGoodsReceipt_GetById tests retrieving a goods receipt
func TestGoodsReceipt_GetById(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(&models.GoodsReceipt{}, &models.Branch{}, &models.Supplier{}, &models.PurchaseInvoice{}, &models.PurchaseInvoiceItem{})
	require.NoError(t, err)

	// Create repository and service
	goodsReceiptRepo := repositories.NewGoodsReceiptRepository(db)
	goodsReceiptService := services.NewGoodsReceiptService(db, goodsReceiptRepo, nil, nil, nil, nil, nil)

	// Setup test data
	ctx := context.Background()

	// Create branch and supplier
	branch := &models.Branch{Name: "Test Branch", Address: "123 Test St"}
	db.Create(branch)

	createdBy := uint(1)
	supplier := &models.Supplier{Name: "Test Supplier", ContactPerson: "John Doe", Phone: "555-1234", Email: "supplier@test.com", Address: "123 Supplier St", CreatedBy: &createdBy, UpdatedBy: &createdBy}
	db.Create(supplier)

	// Create product and invoice
	product := &models.Product{SKU: "TEST-001", Name: "Test Product", StockQty: 50, Price: "15000.00", ReorderThreshold: 10, BranchID: branch.ID}
	db.Create(product)

	invoice := &models.PurchaseInvoice{InvoiceNumber: "INV-TEST-002", InvoiceDate: time.Now().UTC(), SupplierID: supplier.ID, TotalAmount: 1000000.00, PaymentStatus: "unpaid", ReceiptStatus: "pending", BranchID: branch.ID, CreatedBy: &createdBy, UpdatedBy: &createdBy}
	db.Create(invoice)

	// Create goods receipt
	receipt := &models.GoodsReceipt{
		PurchaseInvoiceID: invoice.ID,
		ReceivedDate:      time.Now().UTC(),
		ReceivedBy:        1,
		Notes:             "Test receipt",
		BranchID:          branch.ID,
	}
	err = goodsReceiptRepo.Create(ctx, receipt)
	require.NoError(t, err)

	// Get goods receipt
	found, err := goodsReceiptService.GetByID(ctx, receipt.ID)
	assert.NoError(t, err)
	assert.Equal(t, receipt.ID, found.ID)
	assert.Equal(t, receipt.Notes, found.Notes)
}

// TestGoodsReceipt_List tests listing goods receipts with pagination
func TestGoodsReceipt_List(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(&models.GoodsReceipt{}, &models.Branch{}, &models.Supplier{}, &models.PurchaseInvoice{})
	require.NoError(t, err)

	// Create repository and service
	goodsReceiptRepo := repositories.NewGoodsReceiptRepository(db)
	goodsReceiptService := services.NewGoodsReceiptService(db, goodsReceiptRepo, nil, nil, nil, nil, nil)

	// Setup test data
	ctx := context.Background()

	// Create branch
	branch := &models.Branch{Name: "Test Branch", Address: "123 Test St"}
	db.Create(branch)

	// Create supplier and invoice
	createdBy := uint(1)
	supplier := &models.Supplier{Name: "Test Supplier", ContactPerson: "John Doe", Phone: "555-1234", Email: "supplier@test.com", Address: "123 Supplier St", CreatedBy: &createdBy, UpdatedBy: &createdBy}
	db.Create(supplier)

	invoice := &models.PurchaseInvoice{InvoiceNumber: "INV-TEST-003", InvoiceDate: time.Now().UTC(), SupplierID: supplier.ID, TotalAmount: 1000000.00, PaymentStatus: "unpaid", BranchID: branch.ID, CreatedBy: &createdBy, UpdatedBy: &createdBy}
	db.Create(invoice)

	// Create multiple goods receipts
	for i := 1; i <= 3; i++ {
		receipt := &models.GoodsReceipt{
			PurchaseInvoiceID: invoice.ID,
			ReceivedDate:      time.Now().UTC(),
			ReceivedBy:        1,
			Notes:             fmt.Sprintf("Test receipt %d", i),
			BranchID:          branch.ID,
		}
		err = goodsReceiptRepo.Create(ctx, receipt)
		require.NoError(t, err)
	}

	// List goods receipts
	filter := &services.GoodsReceiptFilter{
		Page:  1,
		Limit: 10,
	}

	receipts, total, err := goodsReceiptService.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, receipts, 3)
}

// Helper function for string pointer
func stringPtr(s string) *string {
	return &s
}
