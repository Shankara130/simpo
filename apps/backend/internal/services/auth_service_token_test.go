package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// TestAuthService_GenerateToken_Success tests JWT token generation (Story 1.5, AC2)
func TestAuthService_GenerateToken_Success(t *testing.T) {
	// Setup
	cfg := &config.JWTConfig{
		Secret:         "test-secret-key-for-jwt-signing-min-32-chars",
		AccessTokenTTL: 8 * time.Hour,
	}

	authService := NewAuthService(cfg, &MockUserRepository{}, &MockAuditService{})

	testUser := &user.User{
		ID:       1,
		Username: "admin",
		Email:    "admin@simpo.pharmacy",
		Role:     user.RoleSystemAdmin,
		BranchID: nil,
	}

	// Execute
	token, err := authService.generateToken(testUser)

	// Assert
	require.NoError(t, err, "Token generation should succeed")
	assert.NotEmpty(t, token, "Token should not be empty")

	// Verify token structure (Story 1.5, AC2)
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	require.NoError(t, err, "Token should be valid JWT")

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok, "Token should have valid claims")

	// Verify user claims (Story 1.5, AC2)
	assert.Equal(t, float64(1), claims["user_id"], "Token should contain user_id")
	assert.Equal(t, "admin", claims["username"], "Token should contain username")
	assert.Equal(t, "admin@simpo.pharmacy", claims["email"], "Token should contain email")
	assert.Equal(t, "SYSTEM_ADMIN", claims["role"], "Token should contain role")
	assert.Nil(t, claims["branch_id"], "System admin should have nil branch_id")

	// Verify expiration
	exp, ok := claims["exp"].(float64)
	require.True(t, ok, "Token should have exp claim")
	expTime := time.Unix(int64(exp), 0)
	expectedExp := time.Now().Add(8 * time.Hour)
	assert.WithinDuration(t, expectedExp, expTime, time.Second, "Token should expire in 8 hours")
}

// TestAuthService_GenerateToken_WithBranchID tests token with branch_id claim
func TestAuthService_GenerateToken_WithBranchID(t *testing.T) {
	// Setup
	cfg := &config.JWTConfig{
		Secret:         "test-secret-key-for-jwt-signing-min-32-chars",
		AccessTokenTTL: 8 * time.Hour,
	}

	authService := NewAuthService(cfg, &MockUserRepository{}, &MockAuditService{})

	branchID := uint(5)
	testUser := &user.User{
		ID:       2,
		Username: "cashier1",
		Email:    "cashier1@simpo.pharmacy",
		Role:     user.RoleCashier,
		BranchID: &branchID,
	}

	// Execute
	token, err := authService.generateToken(testUser)

	// Assert
	require.NoError(t, err, "Token generation should succeed")

	// Verify token claims
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, float64(2), claims["user_id"])
	assert.Equal(t, "cashier1", claims["username"])
	assert.Equal(t, "CASHIER", claims["role"])
	assert.Equal(t, float64(5), claims["branch_id"], "Token should contain branch_id for cashier")
}

// TestAuthService_GenerateToken_AllRoles tests token generation for all role types
func TestAuthService_GenerateToken_AllRoles(t *testing.T) {
	roles := []struct {
		role         string
		branchID     *uint
		expectedRole string
	}{
		{user.RoleSystemAdmin, nil, "SYSTEM_ADMIN"},
		{user.RoleOwner, nil, "OWNER"},
		{user.RoleCashier, nil, "CASHIER"},
	}

	for _, tc := range roles {
		t.Run(tc.role, func(t *testing.T) {
			cfg := &config.JWTConfig{
				Secret:         "test-secret-key-for-jwt-signing-min-32-chars",
				AccessTokenTTL: 8 * time.Hour,
			}

			authService := NewAuthService(cfg, &MockUserRepository{}, &MockAuditService{})

			testUser := &user.User{
				ID:       1,
				Username: "testuser",
				Email:    "test@simpo.pharmacy",
				Role:     tc.role,
				BranchID: tc.branchID,
			}

			token, err := authService.generateToken(testUser)
			require.NoError(t, err)

			parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.Secret), nil
			})
			claims, _ := parsedToken.Claims.(jwt.MapClaims)
			assert.Equal(t, tc.expectedRole, claims["role"], "Token should contain correct role")
		})
	}
}

// TestAuthService_GenerateToken_Expiration tests 8-hour token expiration (Story 1.5, NFR-SEC-002)
func TestAuthService_GenerateToken_Expiration(t *testing.T) {
	// Setup
	cfg := &config.JWTConfig{
		Secret:         "test-secret-key-for-jwt-signing-min-32-chars",
		AccessTokenTTL: 8 * time.Hour, // Story 1.5, NFR-SEC-002
	}

	authService := NewAuthService(cfg, &MockUserRepository{}, &MockAuditService{})

	testUser := &user.User{
		ID:       1,
		Username: "admin",
		Email:    "admin@simpo.pharmacy",
		Role:     user.RoleSystemAdmin,
	}

	// Execute
	token, err := authService.generateToken(testUser)
	require.NoError(t, err)

	// Verify expiration time
	parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	claims, _ := parsedToken.Claims.(jwt.MapClaims)

	exp, ok := claims["exp"].(float64)
	require.True(t, ok)
	expTime := time.Unix(int64(exp), 0)
	expectedExp := time.Now().Add(8 * time.Hour)

	assert.WithinDuration(t, expectedExp, expTime, time.Second, "Token should expire in 8 hours (28800 seconds)")
	assert.Equal(t, int64(28800), int64(authService.accessTokenTTL.Seconds()), "TTL should be 28800 seconds")
}

// TestAuthService_GenerateToken_Issuer tests token issuer claim
func TestAuthService_GenerateToken_Issuer(t *testing.T) {
	// Setup
	cfg := &config.JWTConfig{
		Secret:         "test-secret-key-for-jwt-signing-min-32-chars",
		AccessTokenTTL: 8 * time.Hour,
	}

	authService := NewAuthService(cfg, &MockUserRepository{}, &MockAuditService{})

	testUser := &user.User{
		ID:       1,
		Username: "admin",
		Email:    "admin@simpo.pharmacy",
		Role:     user.RoleSystemAdmin,
	}

	// Execute
	token, err := authService.generateToken(testUser)
	require.NoError(t, err)

	// Verify issuer
	parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	claims, _ := parsedToken.Claims.(jwt.MapClaims)

	assert.Equal(t, "simpo-api", claims["iss"], "Token issuer should be simpo-api")
}
