package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

var testSupplierCounter = 0

func setupSupplierTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.Supplier{})
	require.NoError(t, err)

	return db
}

func createTestSupplier(t *testing.T, db *gorm.DB) *models.Supplier {
	testSupplierCounter++
	createdBy := uint(1)
	supplier := &models.Supplier{
		Name:          fmt.Sprintf("Test Supplier %d", testSupplierCounter),
		ContactPerson: "John Doe",
		Phone:         "555-1234",
		Email:         fmt.Sprintf("supplier%d@example.com", testSupplierCounter),
		Address:       "123 Supplier St",
		CreatedBy:     &createdBy,
		UpdatedBy:     &createdBy,
	}
	err := db.Create(supplier).Error
	require.NoError(t, err)
	return supplier
}

// TestSupplierRepository_Create tests creating a new supplier
// Story 10.1: Verify supplier creation with audit fields
func TestSupplierRepository_Create(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	supplier := &models.Supplier{
		Name:          "PT. Pharmasi Jaya",
		ContactPerson: "Budi Santoso",
		Phone:         "+62-21-1234-5678",
		Email:         "contact@pharmasi.com",
		Address:       "Jl. Industri No. 123, Jakarta",
	}

	createdBy := uint(1)
	err := repo.Create(ctx, supplier, createdBy)
	assert.NoError(t, err)
	assert.NotZero(t, supplier.ID)
	assert.Equal(t, &createdBy, supplier.CreatedBy)
	assert.Equal(t, &createdBy, supplier.UpdatedBy)
	assert.True(t, supplier.IsActive)
}

// TestSupplierRepository_CreateNilSupplier tests error handling for nil supplier
func TestSupplierRepository_CreateNilSupplier(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	err := repo.Create(ctx, nil, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSupplierRepository_CreateMissingName tests validation for required name field
func TestSupplierRepository_CreateMissingName(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	supplier := &models.Supplier{
		Phone: "+62-21-1234-5678",
	}

	err := repo.Create(ctx, supplier, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

// TestSupplierRepository_CreateMissingPhone tests validation for required phone field
func TestSupplierRepository_CreateMissingPhone(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	supplier := &models.Supplier{
		Name: "PT. Pharmasi Jaya",
	}

	err := repo.Create(ctx, supplier, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "phone is required")
}

// TestSupplierRepository_GetByID tests retrieving a supplier by ID
// Story 10.1, AC1: Verify supplier retrieval
func TestSupplierRepository_GetByID(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	// Create test supplier
	created := createTestSupplier(t, db)

	// Test GetByID
	found, err := repo.GetByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.Name, found.Name)
	assert.Equal(t, created.Phone, found.Phone)

	// Test not found
	_, err = repo.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

// TestSupplierRepository_GetByIDZeroID tests validation for zero ID
func TestSupplierRepository_GetByIDZeroID(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 0)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
}

// TestSupplierRepository_GetByName tests retrieving a supplier by name
// Story 10.1, AC1: Check for duplicate supplier names
func TestSupplierRepository_GetByName(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	// Create test supplier
	created := createTestSupplier(t, db)

	// Test GetByName with the actual created supplier name
	found, err := repo.GetByName(ctx, created.Name)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)

	// Test not found
	_, err = repo.GetByName(ctx, "Not Exist")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

// TestSupplierRepository_Update tests updating a supplier
// Story 10.1, AC2: Verify supplier update functionality
func TestSupplierRepository_Update(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	// Create test supplier
	supplier := createTestSupplier(t, db)

	// Update supplier
	supplier.Name = "Updated Supplier Name"
	supplier.ContactPerson = "Jane Doe"

	updatedBy := uint(2)
	err := repo.Update(ctx, supplier, updatedBy)
	assert.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, supplier.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Supplier Name", updated.Name)
	assert.Equal(t, "Jane Doe", updated.ContactPerson)
	assert.Equal(t, &updatedBy, updated.UpdatedBy)
}

// TestSupplierRepository_UpdateNilSupplier tests error handling for nil supplier
func TestSupplierRepository_UpdateNilSupplier(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, nil, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSupplierRepository_Deactivate tests soft deleting a supplier
// Story 10.1, AC3: Verify supplier deactivation
func TestSupplierRepository_Deactivate(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	// Create test supplier
	supplier := createTestSupplier(t, db)

	// Deactivate supplier
	deactivatedBy := uint(3)
	err := repo.Deactivate(ctx, supplier.ID, deactivatedBy)
	assert.NoError(t, err)

	// Verify deactivation (should return ErrNotFound)
	_, err = repo.GetByID(ctx, supplier.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

// TestSupplierRepository_DeactivateZeroID tests validation for zero ID
func TestSupplierRepository_DeactivateZeroID(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	err := repo.Deactivate(ctx, 0, 1)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
}

// TestSupplierRepository_List tests listing suppliers with pagination
// Story 10.1, AC2: Verify supplier listing with filters
func TestSupplierRepository_List(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	// Create test suppliers using createTestSupplier for unique names
	for i := 1; i <= 5; i++ {
		createTestSupplier(t, db)
	}

	// Test list all
	suppliers, total, err := repo.List(ctx, &SupplierFilter{
		Page:  1,
		Limit: 10,
	})
	assert.NoError(t, err)
	assert.Len(t, suppliers, 5)
	assert.Equal(t, int64(5), total)

	// Test pagination
	suppliers, total, err = repo.List(ctx, &SupplierFilter{
		Page:  1,
		Limit: 2,
	})
	assert.NoError(t, err)
	assert.Len(t, suppliers, 2)
	assert.Equal(t, int64(5), total)

	// Test search
	suppliers, total, err = repo.List(ctx, &SupplierFilter{
		SearchQuery: "Test",
		Page:        1,
		Limit:       10,
	})
	assert.NoError(t, err)
	assert.Greater(t, len(suppliers), 0)
}

// TestSupplierRepository_ListActiveFilter tests filtering by active status
// Story 10.1, AC2: Verify active status filtering
func TestSupplierRepository_ListActiveFilter(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	// Create test suppliers
	active := createTestSupplier(t, db)
	inactive := createTestSupplier(t, db)

	// Deactivate one supplier
	err := repo.Deactivate(ctx, inactive.ID, 1)
	assert.NoError(t, err)

	// Test active filter
	isActive := true
	suppliers, total, err := repo.List(ctx, &SupplierFilter{
		IsActive: &isActive,
		Page:     1,
		Limit:    10,
	})
	assert.NoError(t, err)
	assert.Greater(t, len(suppliers), 0)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, active.ID, suppliers[0].ID)
}

// TestSupplierRepository_ListNilFilter tests nil filter handling
func TestSupplierRepository_ListNilFilter(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)
	ctx := context.Background()

	// Create test supplier
	createTestSupplier(t, db)

	// Test with nil filter (should use defaults)
	suppliers, total, err := repo.List(ctx, nil)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(suppliers), 1)
	assert.GreaterOrEqual(t, total, int64(1))
}

// TestSupplierRepository_ListCancelledContext tests context cancellation handling
func TestSupplierRepository_ListCancelledContext(t *testing.T) {
	db := setupSupplierTestDB(t)
	repo := NewSupplierRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _, err := repo.List(ctx, &SupplierFilter{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
}
