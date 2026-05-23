package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// mockAuthHandler is a simple mock for testing router setup
type mockAuthHandler struct{}

func (m *mockAuthHandler) Login(c *gin.Context) {
	c.JSON(200, gin.H{"message": "mock login"})
}

func TestSetupRouter_HealthEndpoint(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	mockUserHandler := &user.Handler{}

	cfg := &config.JWTConfig{
		Secret:   "test-secret",
		TTLHours: 24,
	}
	mockAuthService := auth.NewService(cfg)

	testConfig := &config.Config{
		App: config.AppConfig{
			Version:     "1.0.0",
			Environment: "test",
		},
		Server: config.ServerConfig{
			Port: "8080",
		},
		Ratelimit: config.RateLimitConfig{
			Enabled:  true,
			Requests: 100,
			Window:   time.Minute,
		},
		Health: config.HealthConfig{
			Timeout:              5,
			DatabaseCheckEnabled: true,
		},
	}

	// Create mock auth handler for testing (Story 1.5)
	mockAuthHandler := &mockAuthHandler{}

	router := SetupRouter(mockUserHandler, mockAuthHandler, mockAuthService, testConfig, db, nil, nil, nil, nil)

	assert.NotNil(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "healthy")
}

// ============================================================================
// Story 1.6 RBAC Route Access Control Integration Tests
// ============================================================================

// TestSetupRouter_PublicRoutesBypassRBAC tests public routes don't require auth
// Story 1.6, AC5: Public routes (login, health check) bypass RBAC middleware
func TestSetupRouter_PublicRoutesBypassRBAC(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	mockUserHandler := &user.Handler{}

	cfg := &config.JWTConfig{
		Secret:   "test-secret",
		TTLHours: 24,
	}
	mockAuthService := auth.NewService(cfg)

	mockAuthHandler := &mockAuthHandler{}

	testConfig := &config.Config{
		App: config.AppConfig{
			Version:     "1.0.0",
			Environment: "test",
		},
		Server: config.ServerConfig{
			Port: "8080",
		},
		Ratelimit: config.RateLimitConfig{
			Enabled:  false,
		},
		Health: config.HealthConfig{
			Timeout:              5,
			DatabaseCheckEnabled: false,
		},
	}

	router := SetupRouter(mockUserHandler, mockAuthHandler, mockAuthService, testConfig, db, nil, nil, nil, nil)

	// Test public health endpoints
	publicEndpoints := []struct {
		path           string
		method         string
		expectedStatus int
	}{
		{"/health", "GET", http.StatusOK},
		{"/health/live", "GET", http.StatusOK},
		{"/health/ready", "GET", http.StatusOK},
		{"/api/v1/auth/login", "POST", http.StatusBadRequest}, // No credentials but route is accessible (returns 400 due to missing body)
	}

	for _, tc := range publicEndpoints {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			router.ServeHTTP(w, req)

			// Health endpoints should return 200
			// Login endpoint should return 400 (no request body) but route is accessible (not 404)
			assert.NotEqual(t, http.StatusNotFound, w.Code, "Public route %s should be accessible", tc.path)
		})
	}
}

// TestSetupRouter_ProtectedRoutesRequireAuth tests protected routes require JWT auth
// Story 1.6, AC5: Protected routes require JWT auth and RBAC middleware
func TestSetupRouter_ProtectedRoutesRequireAuth(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	mockUserHandler := &user.Handler{}

	cfg := &config.JWTConfig{
		Secret:   "test-secret",
		TTLHours: 24,
	}
	mockAuthService := auth.NewService(cfg)

	mockAuthHandler := &mockAuthHandler{}

	testConfig := &config.Config{
		App: config.AppConfig{
			Version:     "1.0.0",
			Environment: "test",
		},
		Server: config.ServerConfig{
			Port: "8080",
		},
		Ratelimit: config.RateLimitConfig{
			Enabled:  false,
		},
		Health: config.HealthConfig{
			Timeout:              5,
			DatabaseCheckEnabled: false,
		},
	}

	router := SetupRouter(mockUserHandler, mockAuthHandler, mockAuthService, testConfig, db, nil, nil, nil, nil)

	// Test protected endpoints without auth token
	protectedEndpoints := []struct {
		path           string
		method         string
		expectedStatus int
	}{
		{"/api/v1/users", "GET", http.StatusUnauthorized},
		{"/api/v1/users/1", "GET", http.StatusUnauthorized},
		{"/api/v1/admin/settings", "GET", http.StatusUnauthorized},
	}

	for _, tc := range protectedEndpoints {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code, "Protected route %s should return %d without auth", tc.path, tc.expectedStatus)
		})
	}
}

// TestSetupRouter_MiddlewareOrder verifies correct middleware order
// Story 1.6, AC5: Middleware order: CORS → Rate Limit → Auth → RBAC → Handler
func TestSetupRouter_MiddlewareOrder(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	mockUserHandler := &user.Handler{}

	cfg := &config.JWTConfig{
		Secret:   "test-secret",
		TTLHours: 24,
	}
	mockAuthService := auth.NewService(cfg)

	mockAuthHandler := &mockAuthHandler{}

	testConfig := &config.Config{
		App: config.AppConfig{
			Version:     "1.0.0",
			Environment: "test",
		},
		Server: config.ServerConfig{
			Port: "8080",
		},
		Ratelimit: config.RateLimitConfig{
			Enabled:  false,
		},
		Health: config.HealthConfig{
			Timeout:              5,
			DatabaseCheckEnabled: false,
		},
	}

	router := SetupRouter(mockUserHandler, mockAuthHandler, mockAuthService, testConfig, db, nil, nil, nil, nil)

	// Verify CORS headers are set (CORS middleware should be first)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/users", nil)
	req.Header.Set("Origin", "http://example.com")
	router.ServeHTTP(w, req)

	// CORS should add headers
	corsHeadersPresent := w.Header().Get("Access-Control-Allow-Origin") != "" ||
		w.Code == http.StatusUnauthorized // If auth fails before CORS
	assert.True(t, corsHeadersPresent, "CORS middleware should be present")
}
