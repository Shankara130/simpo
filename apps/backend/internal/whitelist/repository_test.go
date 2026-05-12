package whitelist

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create table
	err = db.AutoMigrate(&WhitelistEntry{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// TestRepositoryCreate verifies Create method
func TestRepositoryCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	entry := &WhitelistEntry{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
		Description: "Test pharmacy domain",
	}

	err := repo.Create(ctx, entry)
	if err != nil {
		t.Errorf("Failed to create entry: %v", err)
	}

	if entry.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}
}

// TestRepositoryCreateDuplicate verifies duplicate domain is rejected
func TestRepositoryCreateDuplicate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	entry1 := &WhitelistEntry{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
		Description: "First entry",
	}

	entry2 := &WhitelistEntry{
		Domain:      "test.pharmacy",
		DefaultRole: "OWNER",
		Description: "Duplicate entry",
	}

	// First creation should succeed
	err := repo.Create(ctx, entry1)
	if err != nil {
		t.Fatalf("Failed to create first entry: %v", err)
	}

	// Second creation with same domain should fail
	err = repo.Create(ctx, entry2)
	if err != ErrDomainAlreadyExists {
		t.Errorf("Expected ErrDomainAlreadyExists, got: %v", err)
	}
}

// TestRepositoryFindByID verifies FindByID method
func TestRepositoryFindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	created := &WhitelistEntry{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
		Description: "Test entry",
	}

	err := repo.Create(ctx, created)
	if err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}

	// Find existing entry
	found, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Errorf("Failed to find entry by ID: %v", err)
	}
	if found.Domain != "test.pharmacy" {
		t.Errorf("Expected domain 'test.pharmacy', got '%s'", found.Domain)
	}

	// Find non-existing entry
	_, err = repo.FindByID(ctx, 999)
	if err != ErrWhitelistEntryNotFound {
		t.Errorf("Expected ErrWhitelistEntryNotFound, got: %v", err)
	}
}

// TestRepositoryFindByDomain verifies FindByDomain method
func TestRepositoryFindByDomain(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	created := &WhitelistEntry{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
		Description: "Test entry",
	}

	err := repo.Create(ctx, created)
	if err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}

	// Find existing domain
	found, err := repo.FindByDomain(ctx, "test.pharmacy")
	if err != nil {
		t.Errorf("Failed to find entry by domain: %v", err)
	}
	if found.DefaultRole != "CASHIER" {
		t.Errorf("Expected default role 'CASHIER', got '%s'", found.DefaultRole)
	}

	// Find non-existing domain
	_, err = repo.FindByDomain(ctx, "nonexistent.pharmacy")
	if err != ErrWhitelistEntryNotFound {
		t.Errorf("Expected ErrWhitelistEntryNotFound, got: %v", err)
	}
}

// TestRepositoryList verifies List method
func TestRepositoryList(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Create multiple entries
	entries := []*WhitelistEntry{
		{Domain: "test1.pharmacy", DefaultRole: "CASHIER"},
		{Domain: "test2.pharmacy", DefaultRole: "OWNER"},
		{Domain: "test3.pharmacy", DefaultRole: "CASHIER"},
	}

	for _, entry := range entries {
		err := repo.Create(ctx, entry)
		if err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}
	}

	// List all entries
	found, err := repo.List(ctx)
	if err != nil {
		t.Errorf("Failed to list entries: %v", err)
	}
	if len(found) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(found))
	}
}

// TestRepositoryUpdate verifies Update method
func TestRepositoryUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	created := &WhitelistEntry{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
		Description: "Original description",
	}

	err := repo.Create(ctx, created)
	if err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}

	// Update entry
	created.DefaultRole = "OWNER"
	created.Description = "Updated description"

	err = repo.Update(ctx, created)
	if err != nil {
		t.Errorf("Failed to update entry: %v", err)
	}

	// Verify update
	found, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Errorf("Failed to find updated entry: %v", err)
	}
	if found.DefaultRole != "OWNER" {
		t.Errorf("Expected default role 'OWNER', got '%s'", found.DefaultRole)
	}
	if found.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got '%s'", found.Description)
	}
}

// TestRepositoryDelete verifies Delete method
func TestRepositoryDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	created := &WhitelistEntry{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
		Description: "Test entry",
	}

	err := repo.Create(ctx, created)
	if err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}

	// Delete entry
	err = repo.Delete(ctx, created.ID)
	if err != nil {
		t.Errorf("Failed to delete entry: %v", err)
	}

	// Verify deletion
	_, err = repo.FindByID(ctx, created.ID)
	if err != ErrWhitelistEntryNotFound {
		t.Errorf("Expected ErrWhitelistEntryNotFound after deletion, got: %v", err)
	}

	// Delete non-existing entry
	err = repo.Delete(ctx, 999)
	if err != ErrWhitelistEntryNotFound {
		t.Errorf("Expected ErrWhitelistEntryNotFound, got: %v", err)
	}
}

// TestRepositoryExists verifies Exists method
func TestRepositoryExists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Check non-existing domain
	exists, err := repo.Exists(ctx, "test.pharmacy")
	if err != nil {
		t.Errorf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Expected domain to not exist")
	}

	// Create entry
	entry := &WhitelistEntry{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
	}

	err = repo.Create(ctx, entry)
	if err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}

	// Check existing domain
	exists, err = repo.Exists(ctx, "test.pharmacy")
	if err != nil {
		t.Errorf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected domain to exist")
	}
}
