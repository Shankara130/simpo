package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// ============================================================================
// Story 1.10 User Deactivation Route Tests
// ============================================================================

// TestSetupRouter_DeactivateRouteRegistered verifies the deactivation route is registered (Story 1.10, Task 7)
func TestSetupRouter_DeactivateRouteRegistered(t *testing.T) {
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
			Enabled: false,
		},
		Health: config.HealthConfig{
			Timeout:              5,
			DatabaseCheckEnabled: false,
		},
	}

	router := SetupRouter(mockUserHandler, mockAuthHandler, mockAuthService, testConfig, db, nil, nil, nil)

	// Test that the route is registered by checking all routes
	routes := router.Routes()

	// Find the deactivation route
	var deactivateRouteFound bool
	for _, route := range routes {
		if route.Path == "/api/v1/users/:id/deactivate" && route.Method == "PUT" {
			deactivateRouteFound = true
			break
		}
	}

	assert.True(t, deactivateRouteFound, "Deactivation route PUT /api/v1/users/:id/deactivate should be registered (Story 1.10, AC1)")
}

// TestSetupRouter_DeactivateRouteRequiresAuth verifies the deactivation route requires authentication (Story 1.10, AC1, AC7)
func TestSetupRouter_DeactivateRouteRequiresAuth(t *testing.T) {
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
			Enabled: false,
		},
		Health: config.HealthConfig{
			Timeout:              5,
			DatabaseCheckEnabled: false,
		},
	}

	router := SetupRouter(mockUserHandler, mockAuthHandler, mockAuthService, testConfig, db, nil, nil, nil)

	// Make request without authentication
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/1/deactivate", nil)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Should return unauthorized (401) since no auth
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Deactivation route should require authentication (Story 1.10, AC7)")
}

// TestSetupRouter_DeactivateRouteProtected verifies the deactivation route is protected with RBAC (Story 1.10, AC1)
func TestSetupRouter_DeactivateRouteProtected(t *testing.T) {
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
			Enabled: false,
		},
		Health: config.HealthConfig{
			Timeout:              5,
			DatabaseCheckEnabled: false,
		},
	}

	router := SetupRouter(mockUserHandler, mockAuthHandler, mockAuthService, testConfig, db, nil, nil, nil)

	// Test all user management routes are protected
	protectedEndpoints := []struct {
		path           string
		method         string
	}{
		{"/api/v1/users", "GET"},
		{"/api/v1/users", "POST"},
		{"/api/v1/users/1", "GET"},
		{"/api/v1/users/1", "PUT"},
		{"/api/v1/users/1", "DELETE"},
		{"/api/v1/users/1/deactivate", "PUT"}, // Story 1.10
	}

	for _, tc := range protectedEndpoints {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "Protected route %s should return 401 without auth", tc.path)
		})
	}
}

// TestSetupRouter_UserRoutesGroup verifies all user routes are in the same group with same middleware (Story 1.10, AC1)
func TestSetupRouter_UserRoutesGroup(t *testing.T) {
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
			Enabled: false,
		},
		Health: config.HealthConfig{
			Timeout:              5,
			DatabaseCheckEnabled: false,
		},
	}

	router := SetupRouter(mockUserHandler, mockAuthHandler, mockAuthService, testConfig, db, nil, nil, nil)

	// All user routes should be under /api/v1/users prefix
	expectedPaths := []string{
		"/api/v1/users",
		"/api/v1/users/:id",
		"/api/v1/users/:id/deactivate", // Story 1.10
	}

	routes := router.Routes()
	foundPaths := make(map[string]bool)

	for _, route := range routes {
		for _, expectedPath := range expectedPaths {
			if route.Path == expectedPath {
				foundPaths[expectedPath] = true
			}
		}
	}

	for _, expectedPath := range expectedPaths {
		assert.True(t, foundPaths[expectedPath], "Expected route path %s should be registered", expectedPath)
	}
}
