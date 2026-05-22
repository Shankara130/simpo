package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// setupTestDBForExpiredFilter creates an in-memory SQLite database for testing
func setupTestDBForExpiredFilter(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Migrate tables
	err = db.AutoMigrate(&models.Product{}, &models.Branch{})
	require.NoError(t, err)

	return db
}

// seedTestProducts populates the database with test products
func seedTestProducts(t *testing.T, db *gorm.DB) uint {
	// Clean up any existing data first
	db.Exec("DELETE FROM products WHERE 1=1")
	db.Exec("DELETE FROM branches WHERE 1=1")

	// Create a branch (auto-generate ID)
	branch := &models.Branch{
		Name:    "Test Branch",
		Address: "123 Test St",
	}
	require.NoError(t, db.Create(branch).Error)
	branchID := branch.ID

	// Create test products with different expiry statuses
	now := time.Now()
	pastDate := now.Add(-24 * time.Hour)      // Expired yesterday
	futureDate := now.Add(30 * 24 * time.Hour) // Expires in 30 days
	products := []*models.Product{
		{
			ID:         1,
			SKU:        "EXP001",
			Name:       "Expired Product",
			StockQty:   10,
			Price:      "10000.00",
			BranchID:   branchID,
			ExpiryDate: &pastDate,
		},
		{
			ID:         2,
			SKU:        "VAL001",
			Name:       "Valid Product",
			StockQty:   20,
			Price:      "20000.00",
			BranchID:   branchID,
			ExpiryDate: &futureDate,
		},
		{
			ID:         3,
			SKU:        "NOP001",
			Name:       "No Expiry Product",
			StockQty:   30,
			Price:      "30000.00",
			BranchID:   branchID,
			ExpiryDate: nil, // No expiry date
		},
	}

	for _, p := range products {
		require.NoError(t, db.Create(p).Error)
	}

	return branchID
}

func TestProductRepository_List_WithExpiredFilter(t *testing.T) {
	db := setupTestDBForExpiredFilter(t)
	branchID := seedTestProducts(t, db)
	repo := NewProductRepository(db)
	ctx := context.Background()

	t.Run("Filter expired products only", func(t *testing.T) {
		filter := &ProductFilter{
			BranchID: &branchID,
			Expired:  true,
			Page:     1,
			Limit:    20,
		}

		products, total, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total, "Should find 1 expired product")
		assert.Len(t, products, 1, "Should return 1 expired product")
		assert.Equal(t, "EXP001", products[0].SKU, "Should be the expired product")
		assert.True(t, products[0].IsExpired(), "Product should be marked as expired")
	})

	t.Run("Filter non-expired products only", func(t *testing.T) {
		filter := &ProductFilter{
			BranchID: &branchID,
			Expired:  false,
			Page:     1,
			Limit:    20,
		}

		products, total, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total, "Should find 2 non-expired products")
		assert.Len(t, products, 2, "Should return 2 non-expired products")

		// Verify neither product is expired
		for _, p := range products {
			assert.False(t, p.IsExpired(), "Product should not be expired")
		}
	})

	t.Run("No expired filter returns non-expired products by default", func(t *testing.T) {
		// Story 4.6, Task 2.3: Ensure expired products are excluded from "available for sale" queries by default
		filter := &ProductFilter{
			BranchID: &branchID,
			Page:     1,
			Limit:    20,
		}

		products, total, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total, "Should find 2 non-expired products (default behavior)")
		assert.Len(t, products, 2, "Should return 2 non-expired products")

		// Verify all returned products are not expired
		for _, p := range products {
			assert.False(t, p.IsExpired(), "All products should not be expired by default")
		}
	})
}

func TestProductRepository_List_ExcludedExpiredByDefaultForAvailable(t *testing.T) {
	db := setupTestDBForExpiredFilter(t)
	branchID := seedTestProducts(t, db)
	repo := NewProductRepository(db)
	ctx := context.Background()

	// Test that default listing (no expired filter) still includes all products
	// But we can filter them out by setting Expired=false
	t.Run("Available for sale (non-expired only)", func(t *testing.T) {
		filter := &ProductFilter{
			BranchID: &branchID,
			Expired:  false, // Explicitly exclude expired
			Page:     1,
			Limit:    20,
		}

		products, total, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total, "Should find 2 available (non-expired) products")

		// Verify all returned products are not expired
		for _, p := range products {
			assert.False(t, p.IsExpired(), "All available products should not be expired")
		}
	})
}

func TestProductRepository_GetExpiredProducts(t *testing.T) {
	db := setupTestDBForExpiredFilter(t)
	branchID := seedTestProducts(t, db)
	repo := NewProductRepository(db)
	ctx := context.Background()

	expired, err := repo.GetExpiredProducts(ctx, branchID)
	require.NoError(t, err)
	assert.Len(t, expired, 1, "Should find 1 expired product")
	assert.Equal(t, "EXP001", expired[0].SKU, "Should be the expired product")
	assert.True(t, expired[0].IsExpired(), "Product should be marked as expired")
}
