package user

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupVerificationTestDB creates an in-memory SQLite database for testing
func setupVerificationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create table
	err = db.AutoMigrate(&EmailVerificationToken{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// TestEmailVerificationTokenTableName verifies the table name is correct
func TestEmailVerificationTokenTableName(t *testing.T) {
	token := EmailVerificationToken{}
	if token.TableName() != "email_verification_tokens" {
		t.Errorf("Expected table name 'email_verification_tokens', got '%s'", token.TableName())
	}
}

// TestEmailVerificationTokenIsExpired verifies IsExpired method
func TestEmailVerificationTokenIsExpired(t *testing.T) {
	now := time.Now()

	expiredToken := EmailVerificationToken{
		ExpiresAt: now.Add(-1 * time.Hour),
	}

	if !expiredToken.IsExpired() {
		t.Error("Expected token to be expired")
	}

	validToken := EmailVerificationToken{
		ExpiresAt: now.Add(1 * time.Hour),
	}

	if validToken.IsExpired() {
		t.Error("Expected token to be valid")
	}
}

// TestVerificationRepositoryCreateToken verifies CreateToken method
func TestVerificationRepositoryCreateToken(t *testing.T) {
	db := setupVerificationTestDB(t)
	repo := NewVerificationRepository(db)
	ctx := context.Background()

	token := "test-token-123"
	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := repo.CreateToken(ctx, token, email, expiresAt)
	if err != nil {
		t.Errorf("Failed to create token: %v", err)
	}

	// Verify token was created
	found, err := repo.FindByToken(ctx, token)
	if err != nil {
		t.Errorf("Failed to find created token: %v", err)
	}
	if found.Email != email {
		t.Errorf("Expected email '%s', got '%s'", email, found.Email)
	}
}

// TestVerificationRepositoryFindByToken verifies FindByToken method
func TestVerificationRepositoryFindByToken(t *testing.T) {
	db := setupVerificationTestDB(t)
	repo := NewVerificationRepository(db)
	ctx := context.Background()

	token := "test-token-123"
	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create token
	err := repo.CreateToken(ctx, token, email, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Find existing token
	found, err := repo.FindByToken(ctx, token)
	if err != nil {
		t.Errorf("Failed to find token: %v", err)
	}
	if found.Email != email {
		t.Errorf("Expected email '%s', got '%s'", email, found.Email)
	}

	// Find non-existing token
	_, err = repo.FindByToken(ctx, "non-existent-token")
	if err != ErrVerificationTokenNotFound {
		t.Errorf("Expected ErrVerificationTokenNotFound, got: %v", err)
	}
}

// TestVerificationRepositoryDeleteToken verifies DeleteToken method
func TestVerificationRepositoryDeleteToken(t *testing.T) {
	db := setupVerificationTestDB(t)
	repo := NewVerificationRepository(db)
	ctx := context.Background()

	token := "test-token-123"
	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create token
	err := repo.CreateToken(ctx, token, email, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Delete token
	err = repo.DeleteToken(ctx, token)
	if err != nil {
		t.Errorf("Failed to delete token: %v", err)
	}

	// Verify deletion
	_, err = repo.FindByToken(ctx, token)
	if err != ErrVerificationTokenNotFound {
		t.Errorf("Expected ErrVerificationTokenNotFound after deletion, got: %v", err)
	}
}

// TestVerificationRepositoryDuplicateToken verifies duplicate tokens are handled
func TestVerificationRepositoryDuplicateToken(t *testing.T) {
	db := setupVerificationTestDB(t)
	repo := NewVerificationRepository(db)
	ctx := context.Background()

	token := "test-token-123"
	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)

	// First creation should succeed
	err := repo.CreateToken(ctx, token, email, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create first token: %v", err)
	}

	// Second creation with same token should fail (unique constraint)
	err = repo.CreateToken(ctx, token, "another@example.com", expiresAt)
	if err == nil {
		t.Error("Expected error when creating duplicate token")
	}
}
