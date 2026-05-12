package whitelist

import (
	"context"
	"testing"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupServiceTestDB creates an in-memory SQLite database for testing
func setupServiceTestDB(t *testing.T) (*gorm.DB, Repository, Service) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create table
	err = db.AutoMigrate(&WhitelistEntry{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	repo := NewRepository(db)
	service := NewService(repo)

	return db, repo, service
}

// TestServiceAddDomain verifies AddDomain method
func TestServiceAddDomain(t *testing.T) {
	_, _, service := setupServiceTestDB(t)
	ctx := context.Background()

	req := AddWhitelistEntryRequest{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
		Description: "Test pharmacy domain",
	}

	entry, err := service.AddDomain(ctx, req)
	if err != nil {
		t.Errorf("Failed to add domain: %v", err)
	}

	if entry.Domain != "test.pharmacy" {
		t.Errorf("Expected domain 'test.pharmacy', got '%s'", entry.Domain)
	}
	if entry.DefaultRole != "CASHIER" {
		t.Errorf("Expected default role 'CASHIER', got '%s'", entry.DefaultRole)
	}
}

// TestServiceAddDomainEmptyDomain verifies empty domain is rejected
func TestServiceAddDomainEmptyDomain(t *testing.T) {
	_, _, service := setupServiceTestDB(t)
	ctx := context.Background()

	req := AddWhitelistEntryRequest{
		Domain:      "",
		DefaultRole: "CASHIER",
	}

	_, err := service.AddDomain(ctx, req)
	if err != ErrDomainRequired {
		t.Errorf("Expected ErrDomainRequired, got: %v", err)
	}
}

// TestServiceAddDomainInvalidRole verifies invalid role is rejected
func TestServiceAddDomainInvalidRole(t *testing.T) {
	_, _, service := setupServiceTestDB(t)
	ctx := context.Background()

	req := AddWhitelistEntryRequest{
		Domain:      "test.pharmacy",
		DefaultRole: "INVALID_ROLE",
	}

	_, err := service.AddDomain(ctx, req)
	if err != ErrInvalidRole {
		t.Errorf("Expected ErrInvalidRole, got: %v", err)
	}
}

// TestServiceAddDuplicateDomain verifies duplicate domain is rejected
func TestServiceAddDuplicateDomain(t *testing.T) {
	_, _, service := setupServiceTestDB(t)
	ctx := context.Background()

	req := AddWhitelistEntryRequest{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
	}

	// First addition should succeed
	_, err := service.AddDomain(ctx, req)
	if err != nil {
		t.Fatalf("Failed to add domain: %v", err)
	}

	// Second addition should fail
	_, err = service.AddDomain(ctx, req)
	if err != ErrDomainAlreadyExists {
		t.Errorf("Expected ErrDomainAlreadyExists, got: %v", err)
	}
}

// TestServiceGetDomain verifies GetDomain method
func TestServiceGetDomain(t *testing.T) {
	_, _, service := setupServiceTestDB(t)
	ctx := context.Background()

	req := AddWhitelistEntryRequest{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
	}

	created, err := service.AddDomain(ctx, req)
	if err != nil {
		t.Fatalf("Failed to add domain: %v", err)
	}

	// Get existing domain
	found, err := service.GetDomain(ctx, created.ID)
	if err != nil {
		t.Errorf("Failed to get domain: %v", err)
	}
	if found.Domain != "test.pharmacy" {
		t.Errorf("Expected domain 'test.pharmacy', got '%s'", found.Domain)
	}

	// Get non-existing domain
	_, err = service.GetDomain(ctx, 999)
	if err == nil {
		t.Error("Expected error for non-existing domain")
	}
}

// TestServiceListDomains verifies ListDomains method
func TestServiceListDomains(t *testing.T) {
	_, _, service := setupServiceTestDB(t)
	ctx := context.Background()

	// Add multiple domains
	domains := []AddWhitelistEntryRequest{
		{Domain: "test1.pharmacy", DefaultRole: "CASHIER"},
		{Domain: "test2.pharmacy", DefaultRole: "OWNER"},
		{Domain: "test3.pharmacy", DefaultRole: "CASHIER"},
	}

	for _, req := range domains {
		_, err := service.AddDomain(ctx, req)
		if err != nil {
			t.Fatalf("Failed to add domain: %v", err)
		}
	}

	// List all domains
	entries, err := service.ListDomains(ctx)
	if err != nil {
		t.Errorf("Failed to list domains: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 domains, got %d", len(entries))
	}
}

// TestServiceUpdateDomain verifies UpdateDomain method
func TestServiceUpdateDomain(t *testing.T) {
	_, _, service := setupServiceTestDB(t)
	ctx := context.Background()

	// Create domain
	createReq := AddWhitelistEntryRequest{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
		Description: "Original description",
	}

	created, err := service.AddDomain(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to add domain: %v", err)
	}

	// Update domain
	updateReq := UpdateWhitelistEntryRequest{
		DefaultRole: "OWNER",
		Description: "Updated description",
	}

	updated, err := service.UpdateDomain(ctx, created.ID, updateReq)
	if err != nil {
		t.Errorf("Failed to update domain: %v", err)
	}

	if updated.DefaultRole != "OWNER" {
		t.Errorf("Expected default role 'OWNER', got '%s'", updated.DefaultRole)
	}
	if updated.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got '%s'", updated.Description)
	}

	// Update non-existing domain
	_, err = service.UpdateDomain(ctx, 999, updateReq)
	if err == nil {
		t.Error("Expected error for non-existing domain")
	}
}

// TestServiceUpdateDomainInvalidRole verifies invalid role is rejected on update
func TestServiceUpdateDomainInvalidRole(t *testing.T) {
	_, _, service := setupServiceTestDB(t)
	ctx := context.Background()

	// Create domain
	createReq := AddWhitelistEntryRequest{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
	}

	created, err := service.AddDomain(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to add domain: %v", err)
	}

	// Update with invalid role
	updateReq := UpdateWhitelistEntryRequest{
		DefaultRole: "INVALID_ROLE",
	}

	_, err = service.UpdateDomain(ctx, created.ID, updateReq)
	if err != ErrInvalidRole {
		t.Errorf("Expected ErrInvalidRole, got: %v", err)
	}
}

// TestServiceDeleteDomain verifies DeleteDomain method
func TestServiceDeleteDomain(t *testing.T) {
	_, _, service := setupServiceTestDB(t)
	ctx := context.Background()

	// Create domain
	createReq := AddWhitelistEntryRequest{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
	}

	created, err := service.AddDomain(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to add domain: %v", err)
	}

	// Delete domain
	err = service.DeleteDomain(ctx, created.ID)
	if err != nil {
		t.Errorf("Failed to delete domain: %v", err)
	}

	// Verify deletion
	_, err = service.GetDomain(ctx, created.ID)
	if err == nil {
		t.Error("Expected error after deletion")
	}

	// Delete non-existing domain
	err = service.DeleteDomain(ctx, 999)
	if err == nil {
		t.Error("Expected error for non-existing domain")
	}
}

// TestServiceValidateDomainWhitelisted verifies ValidateDomainWhitelisted method
func TestServiceValidateDomainWhitelisted(t *testing.T) {
	_, _, service := setupServiceTestDB(t)
	ctx := context.Background()

	// Validate non-whitelisted domain
	_, err := service.ValidateDomainWhitelisted(ctx, "test.pharmacy")
	if err == nil {
		t.Error("Expected error for non-whitelisted domain")
	}

	// Add domain to whitelist
	createReq := AddWhitelistEntryRequest{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
	}

	_, err = service.AddDomain(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to add domain: %v", err)
	}

	// Validate whitelisted domain
	entry, err := service.ValidateDomainWhitelisted(ctx, "test.pharmacy")
	if err != nil {
		t.Errorf("Failed to validate whitelisted domain: %v", err)
	}
	if entry.DefaultRole != "CASHIER" {
		t.Errorf("Expected default role 'CASHIER', got '%s'", entry.DefaultRole)
	}

	// Validate empty domain
	_, err = service.ValidateDomainWhitelisted(ctx, "")
	if err != ErrDomainRequired {
		t.Errorf("Expected ErrDomainRequired, got: %v", err)
	}
}

// TestServiceAddDomainValidRoles verifies all valid roles can be added
func TestServiceAddDomainValidRoles(t *testing.T) {
	_, _, service := setupServiceTestDB(t)
	ctx := context.Background()

	// Map roles to valid domain names (without underscores)
	validRoles := map[string]string{
		"SYSTEM_ADMIN": "admin.pharmacy",
		"OWNER":        "owner.pharmacy",
		"CASHIER":      "cashier.pharmacy",
	}

	for role, domain := range validRoles {
		req := AddWhitelistEntryRequest{
			Domain:      domain,
			DefaultRole: role,
		}

		_, err := service.AddDomain(ctx, req)
		if err != nil {
			t.Errorf("Failed to add domain with role %s: %v", role, err)
		}
	}
}

// TestIsValidRoleForCreateIntegration tests integration with user.IsValidRoleForCreate
func TestIsValidRoleForCreateIntegration(t *testing.T) {
	// Test valid roles
	validRoles := []string{"SYSTEM_ADMIN", "OWNER", "CASHIER"}
	for _, role := range validRoles {
		if !user.IsValidRoleForCreate(role) {
			t.Errorf("Expected role '%s' to be valid", role)
		}
	}

	// Test invalid roles
	invalidRoles := []string{"ADMIN", "USER", "MANAGER", ""}
	for _, role := range invalidRoles {
		if user.IsValidRoleForCreate(role) {
			t.Errorf("Expected role '%s' to be invalid", role)
		}
	}
}

// mockWhitelistRepository is a mock for testing error conditions
type mockWhitelistRepository struct {
	findByDomainFunc func(ctx context.Context, domain string) (*WhitelistEntry, error)
}

func (m *mockWhitelistRepository) Create(ctx context.Context, entry *WhitelistEntry) error {
	return nil
}

func (m *mockWhitelistRepository) FindByID(ctx context.Context, id uint) (*WhitelistEntry, error) {
	return nil, ErrWhitelistEntryNotFound
}

func (m *mockWhitelistRepository) FindByDomain(ctx context.Context, domain string) (*WhitelistEntry, error) {
	if m.findByDomainFunc != nil {
		return m.findByDomainFunc(ctx, domain)
	}
	return nil, ErrWhitelistEntryNotFound
}

func (m *mockWhitelistRepository) List(ctx context.Context) ([]WhitelistEntry, error) {
	return []WhitelistEntry{}, nil
}

func (m *mockWhitelistRepository) Update(ctx context.Context, entry *WhitelistEntry) error {
	return nil
}

func (m *mockWhitelistRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockWhitelistRepository) Exists(ctx context.Context, domain string) (bool, error) {
	return false, nil
}

// TestServiceAddDomainRepositoryError tests repository error handling
func TestServiceAddDomainRepositoryError(t *testing.T) {
	ctx := context.Background()
	mockRepo := &mockWhitelistRepository{
		findByDomainFunc: func(ctx context.Context, domain string) (*WhitelistEntry, error) {
			return &WhitelistEntry{}, nil // Simulate existing domain
		},
	}
	service := NewService(mockRepo)

	req := AddWhitelistEntryRequest{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
	}

	_, err := service.AddDomain(ctx, req)
	if err != ErrDomainAlreadyExists {
		t.Errorf("Expected ErrDomainAlreadyExists, got: %v", err)
	}
}
