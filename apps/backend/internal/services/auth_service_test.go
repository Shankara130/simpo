package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// MockUserRepository is a mock for user repository
type MockUserRepository struct {
	FindByUsernameFunc func(ctx context.Context, username string) (*user.User, error)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	if m.FindByUsernameFunc != nil {
		return m.FindByUsernameFunc(ctx, username)
	}
	return nil, gorm.ErrRecordNotFound
}

// TestAuthService_Login_Success tests successful login flow (Story 1.5, AC1, AC3)
func TestAuthService_Login_Success(t *testing.T) {
	// Setup
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), 12)
	require.NoError(t, err, "Failed to generate bcrypt hash")

	testUser := &user.User{
		ID:           1,
		Username:     "admin",
		Email:        "admin@simpo.pharmacy",
		Name:         "System Admin",
		PasswordHash: string(hashedPassword),
		Status:       user.UserStatusActive,
		Role:         user.RoleSystemAdmin,
		BranchID:     nil,
	}

	mockRepo := &MockUserRepository{
		FindByUsernameFunc: func(ctx context.Context, username string) (*user.User, error) {
			return testUser, nil
		},
	}

	mockAudit := &MockAuditService{}

	cfg := &config.JWTConfig{
		Secret:          "test-secret-key-for-jwt-signing-min-32-chars",
		AccessTokenTTL:  8 * time.Hour, // 8 hours (Story 1.5, NFR-SEC-002)
		RefreshTokenTTL: 168 * time.Hour,
	}

	authService := NewAuthService(cfg, mockRepo, mockAudit)

	// Execute
	result, err := authService.Login(context.Background(), "admin", "SecurePassword123!", "")

	// Assert
	require.NoError(t, err, "Login should succeed with correct credentials")
	require.NotNil(t, result, "Result should not be nil")
	assert.NotNil(t, result.User, "User should be returned")
	assert.Equal(t, uint(1), result.User.ID, "User ID should match")
	assert.Equal(t, "admin", result.User.Username, "Username should match")
	assert.NotEmpty(t, result.AccessToken, "JWT token should be generated")
	assert.Equal(t, 28800, result.ExpiresIn, "Token should expire in 8 hours")
	assert.Equal(t, 1, mockAudit.LogCount, "Audit log should be called once for successful login")
}

// TestAuthService_Login_InvalidUsername tests login with non-existent username (Story 1.5, AC3)
func TestAuthService_Login_InvalidUsername(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{
		FindByUsernameFunc: func(ctx context.Context, username string) (*user.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	mockAudit := &MockAuditService{}

	cfg := &config.JWTConfig{
		Secret:         "test-secret-key-for-jwt-signing-min-32-chars",
		AccessTokenTTL: 8 * time.Hour,
	}

	authService := NewAuthService(cfg, mockRepo, mockAudit)

	// Execute
	result, err := authService.Login(context.Background(), "nonexistent", "password", "")

	// Assert
	assert.Error(t, err, "Login should fail with non-existent username")
	assert.Nil(t, result, "Result should be nil on failure")
	assert.Equal(t, 1, mockAudit.LogCount, "Audit log should be called once for failed login")
}

// TestAuthService_Login_InvalidPassword tests login with incorrect password (Story 1.5, AC3)
func TestAuthService_Login_InvalidPassword(t *testing.T) {
	// Setup
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("CorrectPassword123!"), 12)
	require.NoError(t, err)

	testUser := &user.User{
		ID:           1,
		Username:     "admin",
		PasswordHash: string(hashedPassword),
		Status:       user.UserStatusActive,
	}

	mockRepo := &MockUserRepository{
		FindByUsernameFunc: func(ctx context.Context, username string) (*user.User, error) {
			return testUser, nil
		},
	}

	mockAudit := &MockAuditService{}

	cfg := &config.JWTConfig{
		Secret:         "test-secret-key-for-jwt-signing-min-32-chars",
		AccessTokenTTL: 8 * time.Hour,
	}

	authService := NewAuthService(cfg, mockRepo, mockAudit)

	// Execute
	result, err := authService.Login(context.Background(), "admin", "WrongPassword123!", "")

	// Assert
	assert.Error(t, err, "Login should fail with incorrect password")
	assert.Nil(t, result, "Result should be nil on wrong password")
	assert.Equal(t, 1, mockAudit.LogCount, "Audit log should be called once for failed password")
}

// TestAuthService_Login_InactiveUser tests login with inactive user (Story 1.5, AC6)
func TestAuthService_Login_InactiveUser(t *testing.T) {
	// Setup
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), 12)
	require.NoError(t, err)

	testUser := &user.User{
		ID:           1,
		Username:     "admin",
		PasswordHash: string(hashedPassword),
		Status:       user.UserStatusInactive, // Inactive user
	}

	mockRepo := &MockUserRepository{
		FindByUsernameFunc: func(ctx context.Context, username string) (*user.User, error) {
			return testUser, nil
		},
	}

	mockAudit := &MockAuditService{}

	cfg := &config.JWTConfig{
		Secret:         "test-secret-key-for-jwt-signing-min-32-chars",
		AccessTokenTTL: 8 * time.Hour,
	}

	authService := NewAuthService(cfg, mockRepo, mockAudit)

	// Execute
	result, err := authService.Login(context.Background(), "admin", "Password123!", "")

	// Assert
	assert.Error(t, err, "Login should fail for inactive users")
	assert.Nil(t, result, "Result should be nil for inactive user")
	assert.Equal(t, 1, mockAudit.LogCount, "Audit log should be called once for inactive user")
}

// TestAuthService_Login_EmptyUsername tests login with empty username
func TestAuthService_Login_EmptyUsername(t *testing.T) {
	// Setup
	mockAudit := &MockAuditService{}

	cfg := &config.JWTConfig{
		Secret:         "test-secret-key-for-jwt-signing-min-32-chars",
		AccessTokenTTL: 8 * time.Hour,
	}

	authService := NewAuthService(cfg, &MockUserRepository{}, mockAudit)

	// Execute
	result, err := authService.Login(context.Background(), "", "Password123!", "")

	// Assert
	assert.Error(t, err, "Login should fail with empty username")
	assert.Nil(t, result, "Result should be nil")
	assert.Equal(t, 1, mockAudit.LogCount, "Audit log should be called once for empty username")
}

// TestAuthService_Login_EmptyPassword tests login with empty password
func TestAuthService_Login_EmptyPassword(t *testing.T) {
	// Setup
	mockAudit := &MockAuditService{}

	cfg := &config.JWTConfig{
		Secret:         "test-secret-key-for-jwt-signing-min-32-chars",
		AccessTokenTTL: 8 * time.Hour,
	}

	authService := NewAuthService(cfg, &MockUserRepository{}, mockAudit)

	// Execute
	result, err := authService.Login(context.Background(), "admin", "", "")

	// Assert
	assert.Error(t, err, "Login should fail with empty password")
	assert.Nil(t, result, "Result should be nil")
	assert.Equal(t, 1, mockAudit.LogCount, "Audit log should be called once for empty password")
}
