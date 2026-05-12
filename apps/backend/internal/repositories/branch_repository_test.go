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

var testBranchCounter = 0

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.Branch{})
	require.NoError(t, err)

	return db
}

func createTestBranch(t *testing.T, db *gorm.DB) *models.Branch {
	testBranchCounter++
	branch := &models.Branch{
		Name:    fmt.Sprintf("Test Branch %d", testBranchCounter),
		Address: "123 Test Street",
		Phone:   "555-1234",
		Email:   fmt.Sprintf("test%d@example.com", testBranchCounter),
	}
	err := db.Create(branch).Error
	require.NoError(t, err)
	return branch
}

// TestBranchRepository_Create tests creating a new branch
func TestBranchRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBranchRepository(db)
	ctx := context.Background()

	branch := &models.Branch{
		Name:    "Main Branch",
		Address: "123 Main Street",
		Phone:   "555-0100",
		Email:   "main@example.com",
	}

	err := repo.Create(ctx, branch)
	assert.NoError(t, err)
	assert.NotZero(t, branch.ID)
}

// TestBranchRepository_GetByID tests retrieving a branch by ID
func TestBranchRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBranchRepository(db)
	ctx := context.Background()

	// Create test branch
	created := createTestBranch(t, db)

	// Test GetByID
	found, err := repo.GetByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.Name, found.Name)
	assert.Equal(t, created.Address, found.Address)

	// Test not found
	_, err = repo.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

// TestBranchRepository_GetByName tests retrieving a branch by name
func TestBranchRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBranchRepository(db)
	ctx := context.Background()

	// Create test branch
	created := createTestBranch(t, db)

	// Test GetByName with the actual created branch name
	found, err := repo.GetByName(ctx, created.Name)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)

	// Test not found
	_, err = repo.GetByName(ctx, "Not Exist")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

// TestBranchRepository_Update tests updating a branch
func TestBranchRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBranchRepository(db)
	ctx := context.Background()

	// Create test branch
	branch := createTestBranch(t, db)

	// Update branch
	branch.Name = "Updated Branch"
	branch.Address = "456 New Address"

	err := repo.Update(ctx, branch)
	assert.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, branch.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Branch", updated.Name)
	assert.Equal(t, "456 New Address", updated.Address)
}

// TestBranchRepository_Delete tests soft deleting a branch
func TestBranchRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBranchRepository(db)
	ctx := context.Background()

	// Create test branch
	branch := createTestBranch(t, db)

	// Delete branch
	err := repo.Delete(ctx, branch.ID)
	assert.NoError(t, err)

	// Verify deletion (should return ErrNotFound)
	_, err = repo.GetByID(ctx, branch.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

// TestBranchRepository_List tests listing branches with pagination
func TestBranchRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBranchRepository(db)
	ctx := context.Background()

	// Create test branches using createTestBranch for unique names
	for i := 1; i <= 5; i++ {
		createTestBranch(t, db)
	}

	// Test list all
	filter := &BranchFilter{
		Page:  1,
		Limit: 10,
	}
	branches, total, err := repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, branches, 5)

	// Test pagination
	filter = &BranchFilter{
		Page:  1,
		Limit: 2,
	}
	branches, total, err = repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, branches, 2)
}

// TestBranchRepository_List_Search tests searching branches by name or address
func TestBranchRepository_List_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBranchRepository(db)
	ctx := context.Background()

	// Create test branches
	branch1 := &models.Branch{Name: "Jakarta Branch", Address: "Jl. Sudirman No. 1"}
	branch2 := &models.Branch{Name: "Surabaya Branch", Address: "Jl. Tunjungan No. 2"}
	branch3 := &models.Branch{Name: "Bandung Branch", Address: "Jl. Asia Afrika No. 3"}

	db.Create(branch1)
	db.Create(branch2)
	db.Create(branch3)

	// Test search by name
	filter := &BranchFilter{
		SearchQuery: "Jakarta",
		Page:        1,
		Limit:       10,
	}
	branches, total, err := repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Jakarta Branch", branches[0].Name)

	// Test search by address
	filter = &BranchFilter{
		SearchQuery: "Tunjungan",
		Page:        1,
		Limit:       10,
	}
	branches, total, err = repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Surabaya Branch", branches[0].Name)
}
