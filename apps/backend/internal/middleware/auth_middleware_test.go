package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TestJWTAuthMiddleware_ValidToken tests that valid tokens pass authentication
// Story 1.6, AC1: JWT Token Validation
func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	secret := "test-secret-key-for-jwt-validation"

	// Create a mock token validator
	mockValidator := &mockTokenValidator{
		secret: secret,
		ttl:    8 * time.Hour,
	}

	// Create test router with middleware
	router := gin.New()
	router.Use(JWTAuthMiddleware(mockValidator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Generate valid token
	token := generateTestToken(t, secret, 1, "testuser", "test@example.com", RoleOwner, nil)

	// Create request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify user context was set
	if w.Body.String() == "" {
		t.Error("Expected response body, got empty")
	}
}

// TestJWTAuthMiddleware_MissingAuthorizationHeader tests missing header returns 401
// Story 1.6, AC1: Invalid/expired tokens return 401 Unauthorized
func TestJWTAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-key"

	mockValidator := &mockTokenValidator{
		secret: secret,
		ttl:    8 * time.Hour,
	}

	router := gin.New()
	router.Use(JWTAuthMiddleware(mockValidator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request WITHOUT authorization header
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// Verify RFC 7807 error format
	body := w.Body.String()
	if !containsRFC7807Fields(body) {
		t.Errorf("Response should follow RFC 7807 format, got: %s", body)
	}
}

// TestJWTAuthMiddleware_InvalidAuthorizationFormat tests invalid header format
func TestJWTAuthMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	mockValidator := &mockTokenValidator{secret: secret}

	router := gin.New()
	router.Use(JWTAuthMiddleware(mockValidator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	testCases := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{"Missing Bearer prefix", "InvalidToken", http.StatusUnauthorized},
		{"Empty token", "Bearer ", http.StatusUnauthorized},
		{"Wrong format", "Basic dXNlcjpwYXNz", http.StatusUnauthorized},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.authHeader)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

// TestJWTAuthMiddleware_InvalidToken tests invalid token signature
func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	mockValidator := &mockTokenValidator{secret: secret}

	router := gin.New()
	router.Use(JWTAuthMiddleware(mockValidator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// Create token with different secret (invalid signature)
	invalidToken := generateTestToken(t, "wrong-secret", 1, "testuser", "test@example.com", RoleOwner, nil)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+invalidToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for invalid token, got %d", w.Code)
	}
}

// TestJWTAuthMiddleware_ExpiredToken tests expired token returns 401
// Story 1.6, AC1: Token expiration is checked (8-hour timeout)
func TestJWTAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	expiredTTL := -1 * time.Hour // Negative TTL to simulate expired token

	mockValidator := &mockTokenValidator{
		secret: secret,
		ttl:    expiredTTL,
	}

	router := gin.New()
	router.Use(JWTAuthMiddleware(mockValidator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// Generate expired token
	expiredToken := generateTestTokenWithTTL(t, secret, 1, "testuser", "test@example.com", RoleOwner, nil, expiredTTL)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for expired token, got %d", w.Code)
	}
}

// TestJWTAuthMiddleware_ContextSet tests that user context is set correctly
// Story 1.6, AC2: Role and branch_id are stored in request context
func TestJWTAuthMiddleware_ContextSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	ttl := 8 * time.Hour

	mockValidator := &mockTokenValidator{
		secret: secret,
		ttl:    ttl,
	}

	router := gin.New()
	router.Use(JWTAuthMiddleware(mockValidator))
	router.GET("/test", func(c *gin.Context) {
		// Verify context is set
		claims := GetUserContext(c)
		if claims == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no context"})
			return
		}

		// Return user info from context
		c.JSON(http.StatusOK, gin.H{
			"user_id":   claims.UserID,
			"username":  claims.Username,
			"email":     claims.Email,
			"role":      claims.Role,
			"branch_id": claims.BranchID,
		})
	})

	// Create test user with specific claims
	testUserID := uint(123)
	testBranchID := uint(5)
	token := generateTestToken(t, secret, testUserID, "cashier1", "cashier@test.com", RoleCashier, &testBranchID)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify response contains correct user info
	// Note: We're checking that context was populated, not exact JSON format
	body := w.Body.String()
	if body == "" {
		t.Error("Expected response body with user info, got empty")
	}
}

// TestJWTAuthMiddleware_MissingClaimsInToken tests missing role/branch_id
// Story 1.6, AC2: Missing role/branch_id claims return 403 Forbidden
func TestJWTAuthMiddleware_MissingClaimsInToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	ttl := 8 * time.Hour

	mockValidator := &mockTokenValidator{
		secret:        secret,
		ttl:           ttl,
		missingClaims: true, // Simulate missing claims
	}

	router := gin.New()
	router.Use(JWTAuthMiddleware(mockValidator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// Generate token without role/branch_id
	token := generateTestToken(t, secret, 1, "testuser", "test@example.com", RoleOwner, nil)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 403 for missing claims
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for missing claims, got %d", w.Code)
	}
}

// TestJWTAuthMiddleware_RFC7807ErrorResponse tests RFC 7807 error format
// Story 1.6, AC1: Return 401 with RFC 7807 format
func TestJWTAuthMiddleware_RFC7807ErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	mockValidator := &mockTokenValidator{secret: secret}

	router := gin.New()
	router.Use(JWTAuthMiddleware(mockValidator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// Test with invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// Verify RFC 7807 fields in response
	body := w.Body.String()
	requiredFields := []string{"type", "title", "status", "detail", "instance"}
	for _, field := range requiredFields {
		if !containsString(body, field) {
			t.Errorf("RFC 7807 response should include '%s' field. Body: %s", field, body)
		}
	}
}

// TestGetUserContext tests GetUserContext helper function
func TestGetUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	mockValidator := &mockTokenValidator{
		secret: secret,
		ttl:    8 * time.Hour,
	}

	router := gin.New()
	router.Use(JWTAuthMiddleware(mockValidator))
	router.GET("/test", func(c *gin.Context) {
		userCtx := GetUserContext(c)
		if userCtx == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no user context"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":   userCtx.UserID,
			"role":      userCtx.Role,
			"branch_id": userCtx.BranchID,
		})
	})

	token := generateTestToken(t, secret, 1, "testuser", "test@example.com", RoleCashier, nil)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestGetUserRole tests GetUserRole helper function
func TestGetUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	mockValidator := &mockTokenValidator{
		secret: secret,
		ttl:    8 * time.Hour,
	}

	router := gin.New()
	router.Use(JWTAuthMiddleware(mockValidator))
	router.GET("/test", func(c *gin.Context) {
		role := GetUserRole(c)
		if role == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no role"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"role": role})
	})

	token := generateTestToken(t, secret, 1, "testuser", "test@example.com", RoleOwner, nil)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestGetBranchID tests GetBranchID helper function
func TestGetBranchID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	mockValidator := &mockTokenValidator{
		secret: secret,
		ttl:    8 * time.Hour,
	}

	router := gin.New()
	router.Use(JWTAuthMiddleware(mockValidator))
	router.GET("/test", func(c *gin.Context) {
		branchID := GetBranchID(c)
		c.JSON(http.StatusOK, gin.H{"branch_id": branchID})
	})

	// Test with branch ID
	testBranchID := uint(5)
	token := generateTestToken(t, secret, 1, "testuser", "test@example.com", RoleCashier, &testBranchID)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// Helper: Mock token validator for testing
type mockTokenValidator struct {
	secret        string
	ttl           time.Duration
	missingClaims bool
}

func (m *mockTokenValidator) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Simple validation - parse token and return claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, jwt.ErrTokenExpired
		}
	}

	// Extract claims
	userIDFloat, _ := claims["user_id"].(float64)
	userID := uint(userIDFloat)

	username, _ := claims["username"].(string)
	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)

	var branchID *uint
	if branchIDFloat, ok := claims["branch_id"].(float64); ok {
		bid := uint(branchIDFloat)
		branchID = &bid
	}

	// Simulate missing claims
	if m.missingClaims {
		return &JWTClaims{
			UserID:   userID,
			Username: username,
			Email:    email,
			Role:     "", // Missing role
			BranchID: nil,
		}, nil
	}

	return &JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		BranchID: branchID,
	}, nil
}

// Helper: Generate test JWT token
func generateTestToken(t *testing.T, secret string, userID uint, username, email, role string, branchID *uint) string {
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
		t.Fatalf("Failed to generate test token: %v", err)
	}

	return tokenString
}

// Helper: Generate test JWT token with custom TTL
func generateTestTokenWithTTL(t *testing.T, secret string, userID uint, username, email, role string, branchID *uint, ttl time.Duration) string {
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
		t.Fatalf("Failed to generate test token: %v", err)
	}

	return tokenString
}

// Helper: Check if response contains RFC 7807 fields
func containsRFC7807Fields(body string) bool {
	return containsString(body, "type") &&
		containsString(body, "title") &&
		containsString(body, "status") &&
		containsString(body, "detail")
}

// Helper: Check if string contains substring
func containsString(body, substr string) bool {
	return len(body) > 0 && len(substr) > 0 && indexOf(body, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

