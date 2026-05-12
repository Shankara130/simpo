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

var testProductCounter = 0

func setupProductTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.Branch{}, &models.Product{})
	require.NoError(t, err)

	return db
}

func createTestProductBranch(t *testing.T, db *gorm.DB) *models.Branch {
	branch := &models.Branch{
		Name:    fmt.Sprintf("Product Test Branch %d", testProductCounter),
		Address: "123 Product Test Street",
	}
	err := db.Create(branch).Error
	require.NoError(t, err)
	return branch
}

func createTestProduct(t *testing.T, db *gorm.DB, branch *models.Branch) *models.Product {
	testProductCounter++
	product := &models.Product{
		SKU:             fmt.Sprintf("TEST-PROD-%d", testProductCounter),
		Name:            fmt.Sprintf("Test Product %d", testProductCounter),
		Description:     "Test Description",
		StockQty:        100,
		Price:           "50000.00",
		ExpiryDate:      &time.Time{},
		BranchID:        branch.ID,
		ReorderThreshold: 10,
		Category:        "Test Category",
	}
	err := db.Create(product).Error
	require.NoError(t, err)
	return product
}

// TestProductRepository_Create tests creating a new product
func TestProductRepository_Create(t *testing.T) {
	db := setupProductTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	branch := createTestProductBranch(t, db)
	product := &models.Product{
		SKU:             "NEW-PROD-001",
		Name:            "New Product",
		StockQty:        50,
		Price:           "75000.00",
		BranchID:        branch.ID,
		ReorderThreshold: 5,
	}

	err := repo.Create(ctx, product)
	assert.NoError(t, err)
	assert.NotZero(t, product.ID)
}

// TestProductRepository_GetBySKU tests retrieving a product by SKU
func TestProductRepository_GetBySKU(t *testing.T) {
	db := setupProductTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	branch := createTestProductBranch(t, db)
	created := createTestProduct(t, db, branch)

	// Test GetBySKU
	found, err := repo.GetBySKU(ctx, branch.ID, created.SKU)
	assert.NoError(t, err)
	assert.Equal(t, created.Name, found.Name)

	// Test not found
	_, err = repo.GetBySKU(ctx, branch.ID, "NOT-EXIST")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

// TestProductRepository_List_Filtering tests filtering products
func TestProductRepository_List_Filtering(t *testing.T) {
	db := setupProductTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	branch := createTestProductBranch(t, db)

	// Create products with different categories
	repo.Create(ctx, &models.Product{SKU: "P1", Name: "Paracetamol", Category: "Obat", BranchID: branch.ID, StockQty: 5, Price: "10000", ReorderThreshold: 10})
	repo.Create(ctx, &models.Product{SKU: "P2", Name: "Vitamin C", Category: "Vitamin", BranchID: branch.ID, StockQty: 50, Price: "20000", ReorderThreshold: 10})

	// Test filter by category
	filter := &ProductFilter{
		BranchID: &branch.ID,
		Category: "Obat",
		Page:     1,
		Limit:    10,
	}
	products, total, err := repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Paracetamol", products[0].Name)

	// Test low stock filter
	filter.LowStock = true
	products, total, err = repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total) // Only Paracetamol has stock < 10
}

// TestProductRepository_GetLowStockProducts tests low stock query
func TestProductRepository_GetLowStockProducts(t *testing.T) {
	db := setupProductTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	branch := createTestProductBranch(t, db)

	// Create products
	repo.Create(ctx, &models.Product{SKU: "LOW-1", Name: "Low Stock Item", BranchID: branch.ID, StockQty: 5, Price: "10000", ReorderThreshold: 10})
	repo.Create(ctx, &models.Product{SKU: "OK-1", Name: "Normal Stock", BranchID: branch.ID, StockQty: 20, Price: "20000", ReorderThreshold: 10})

	products, err := repo.GetLowStockProducts(ctx, branch.ID)
	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, "Low Stock Item", products[0].Name)
}
