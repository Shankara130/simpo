package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ============================================================================
// Story 1.6 Integration Tests - End-to-End RBAC Testing
// ============================================================================

// TestRBACIntegration_SuccessfulAccess tests successful access with valid permissions
// Story 1.6, AC8: Test successful access with valid permissions
func TestRBACIntegration_SuccessfulAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-integration"

	// Create test router with full RBAC stack
	router := setupIntegrationRouter(secret)

	testCases := []struct {
		name           string
		role           string
		endpoint       string
		method         string
		expectedStatus int
	}{
		{"SYSTEM_ADMIN access users", RoleSystemAdmin, "/api/v1/users", "GET", http.StatusOK},
		{"OWNER access reports", RoleOwner, "/api/v1/reports", "GET", http.StatusOK},
		{"OWNER access products", RoleOwner, "/api/v1/products", "GET", http.StatusOK},
		{"CASHIER access transactions", RoleCashier, "/api/v1/transactions", "GET", http.StatusOK},
		{"CASHIER access products", RoleCashier, "/api/v1/products", "GET", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := generateIntegrationToken(t, secret, 1, "testuser", "test@example.com", tc.role, nil)

			req := httptest.NewRequest(tc.method, tc.endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d for role %s accessing %s, got %d. Body: %s",
					tc.expectedStatus, tc.role, tc.endpoint, w.Code, w.Body.String())
			}
		})
	}
}

// TestRBACIntegration_AccessDenial tests access denial for insufficient permissions
// Story 1.6, AC8: Test access denial for insufficient permissions
func TestRBACIntegration_AccessDenial(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-integration"

	router := setupIntegrationRouter(secret)

	testCases := []struct {
		name           string
		role           string
		endpoint       string
		method         string
		expectedStatus int
	}{
		{"CASHIER cannot access reports", RoleCashier, "/api/v1/reports", "GET", http.StatusForbidden},
		{"CASHIER cannot access users", RoleCashier, "/api/v1/users", "GET", http.StatusForbidden},
		{"OWNER cannot access admin settings", RoleOwner, "/api/v1/admin/settings", "GET", http.StatusForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := generateIntegrationToken(t, secret, 1, "testuser", "test@example.com", tc.role, nil)

			req := httptest.NewRequest(tc.method, tc.endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d for role %s accessing %s, got %d",
					tc.expectedStatus, tc.role, tc.endpoint, w.Code)
			}

			// Verify RFC 7807 format
			body := w.Body.String()
			if !containsRFC7807Fields(body) {
				t.Errorf("Expected RFC 7807 format, got: %s", body)
			}
		})
	}
}

// TestRBACIntegration_BranchFiltering tests branch-level filtering (cashier vs owner)
// Story 1.6, AC8: Test branch-level filtering (cashier vs owner)
func TestRBACIntegration_BranchFiltering(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-integration"

	router := setupIntegrationRouter(secret)

	// Test branch access info extraction
	testCases := []struct {
		name            string
		role            string
		branchID        *uint
		canAccessAll    bool
		assignedBranch  *uint
	}{
		{
			name:         "SYSTEM_ADMIN - all branches",
			role:         RoleSystemAdmin,
			branchID:     nil,
			canAccessAll: true,
			assignedBranch: nil,
		},
		{
			name:         "OWNER - all branches",
			role:         RoleOwner,
			branchID:     nil,
			canAccessAll: true,
			assignedBranch: nil,
		},
		{
			name:           "CASHIER - assigned branch only",
			role:           RoleCashier,
			branchID:       func() *uint { i := uint(5); return &i }(),
			canAccessAll:   false,
			assignedBranch: func() *uint { i := uint(5); return &i }(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := generateIntegrationToken(t, secret, 1, "testuser", "test@example.com", tc.role, tc.branchID)

			req := httptest.NewRequest("GET", "/api/v1/products", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// For products endpoint, all these roles should have access
			// The branch filtering is done at repository level (not tested here)
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s accessing products, got %d", tc.role, w.Code)
			}
		})
	}
}

// TestRBACIntegration_ExpiredToken tests expired/invalid token handling
// Story 1.6, AC8: Test expired/invalid token handling
func TestRBACIntegration_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-integration"
	expiredTTL := -1 * time.Hour // Expired token

	router := setupIntegrationRouter(secret)

	// Generate expired token
	expiredToken := generateIntegrationTokenWithTTL(t, secret, 1, "testuser", "test@example.com", RoleOwner, nil, expiredTTL)

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 401 for expired token
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for expired token, got %d", w.Code)
	}

	// Verify RFC 7807 format
	body := w.Body.String()
	if !containsRFC7807Fields(body) {
		t.Errorf("Expected RFC 7807 format for expired token, got: %s", body)
	}
}

// TestRBACIntegration_InvalidToken tests invalid token signature
func TestRBACIntegration_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-integration"

	router := setupIntegrationRouter(secret)

	// Generate token with different secret (invalid signature)
	invalidToken := generateIntegrationToken(t, "wrong-secret", 1, "testuser", "test@example.com", RoleOwner, nil)

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+invalidToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 401 for invalid token
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for invalid token, got %d", w.Code)
	}
}

// TestRBACIntegration_PublicRoutesBypassRBAC tests public routes bypass RBAC
// Story 1.6, AC8: Test public routes bypass RBAC
func TestRBACIntegration_PublicRoutesBypassRBAC(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-integration"

	router := setupIntegrationRouter(secret)

	testCases := []struct {
		name           string
		endpoint       string
		method         string
		expectedStatus int
	}{
		{"Health endpoint", "/health", "GET", http.StatusOK},
		{"Login endpoint", "/api/v1/auth/login", "POST", http.StatusBadRequest}, // 400 due to missing body, but route is accessible
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.endpoint, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Public routes should be accessible (not 404)
			if w.Code == http.StatusNotFound {
				t.Errorf("Public route %s should be accessible (not 404), got %d", tc.endpoint, w.Code)
			}
		})
	}
}

// TestRBACIntegration_MissingAuthorizationHeader tests missing auth header
func TestRBACIntegration_MissingAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-integration"

	router := setupIntegrationRouter(secret)

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	// No Authorization header set
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 401 for missing auth header
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for missing auth header, got %d", w.Code)
	}

	// Verify RFC 7807 format
	body := w.Body.String()
	if !containsRFC7807Fields(body) {
		t.Errorf("Expected RFC 7807 format, got: %s", body)
	}
}

// TestRBACIntegration_AllRolePermissions tests all roles have correct permissions
// Story 1.6, AC8: Comprehensive permission testing
func TestRBACIntegration_AllRolePermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-integration"

	router := setupIntegrationRouter(secret)

	// Define permission matrix
	permissionMatrix := map[string]map[string]bool{
		RoleSystemAdmin: {
			"/api/v1/users":       true,
			"/api/v1/products":    true,
			"/api/v1/reports":     true,
			"/api/v1/transactions": true,
			"/api/v1/admin/settings": true,
		},
		RoleOwner: {
			"/api/v1/users":       true,
			"/api/v1/products":    true,
			"/api/v1/reports":     true,
			"/api/v1/transactions": true,
			"/api/v1/admin/settings": false,
		},
		RoleCashier: {
			"/api/v1/users":       false,
			"/api/v1/products":    true,
			"/api/v1/reports":     false,
			"/api/v1/transactions": true,
			"/api/v1/admin/settings": false,
		},
	}

	for role, permissions := range permissionMatrix {
		for endpoint, allowed := range permissions {
			t.Run(role+"_"+endpoint, func(t *testing.T) {
				token := generateIntegrationToken(t, secret, 1, "testuser", "test@example.com", role, nil)

				req := httptest.NewRequest("GET", endpoint, nil)
				req.Header.Set("Authorization", "Bearer "+token)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				if allowed {
					if w.Code != http.StatusOK {
						t.Errorf("Expected %s to access %s (status 200), got %d", role, endpoint, w.Code)
					}
				} else {
					if w.Code != http.StatusForbidden {
						t.Errorf("Expected %s to be denied access to %s (status 403), got %d", role, endpoint, w.Code)
					}
				}
			})
		}
	}
}

// ============================================================================
// Integration Test Helpers
// ============================================================================

// setupIntegrationRouter creates a router with full RBAC middleware stack
func setupIntegrationRouter(secret string) *gin.Engine {
	router := gin.New()

	// Add mock JWT auth middleware
	mockValidator := &mockTokenValidator{
		secret: secret,
		ttl:    8 * time.Hour,
	}

	// Public routes (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing body"})
	})

	// Protected routes with full RBAC stack
	protectedGroup := router.Group("/api/v1")
	protectedGroup.Use(JWTAuthMiddleware(mockValidator))
	protectedGroup.Use(RBACMiddleware())
	{
		protectedGroup.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "users list"})
		})
		protectedGroup.GET("/products", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "products list"})
		})
		protectedGroup.GET("/reports", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "reports list"})
		})
		protectedGroup.GET("/transactions", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "transactions list"})
		})
		protectedGroup.GET("/admin/settings", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin settings"})
		})
	}

	return router
}

// generateIntegrationToken generates a JWT token for integration testing
func generateIntegrationToken(t *testing.T, secret string, userID uint, username, email, role string, branchID *uint) string {
	ttl := 8 * time.Hour
	expirationTime := time.Now().Add(ttl)

	claims := jwt.MapClaims{
		"user_id":  float64(userID),
		"username": username,
		"email":    email,
		"role":     role,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	if branchID != nil {
		claims["branch_id"] = float64(*branchID)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to generate integration test token: %v", err)
	}

	return tokenString
}

// generateIntegrationTokenWithTTL generates a JWT token with custom TTL for testing
func generateIntegrationTokenWithTTL(t *testing.T, secret string, userID uint, username, email, role string, branchID *uint, ttl time.Duration) string {
	expirationTime := time.Now().Add(ttl)

	claims := jwt.MapClaims{
		"user_id":  float64(userID),
		"username": username,
		"email":    email,
		"role":     role,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	if branchID != nil {
		claims["branch_id"] = float64(*branchID)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to generate integration test token: %v", err)
	}

	return tokenString
}
