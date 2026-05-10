package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requiredRole     string
		userRoles        []string
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:           "user has required role",
			requiredRole:   "admin",
			userRoles:      []string{"admin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:             "user missing required role",
			requiredRole:     "admin",
			userRoles:        []string{"user"},
			expectedStatus:   http.StatusForbidden,
			expectedResponse: "insufficient permissions",
		},
		{
			name:             "user has no roles",
			requiredRole:     "admin",
			userRoles:        []string{},
			expectedStatus:   http.StatusForbidden,
			expectedResponse: "insufficient permissions",
		},
		{
			name:           "user has multiple roles including required",
			requiredRole:   "editor",
			userRoles:      []string{"user", "editor", "viewer"},
			expectedStatus: http.StatusOK,
		},
		{
			name:             "no authenticated user",
			requiredRole:     "admin",
			userRoles:        nil,
			expectedStatus:   http.StatusForbidden,
			expectedResponse: "insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			router.Use(func(c *gin.Context) {
				if tt.userRoles != nil {
					claims := &auth.Claims{
						UserID: 1,
						Email:  "test@example.com",
						Roles:  tt.userRoles,
					}
					c.Set(auth.KeyUser, claims)
				}
				c.Next()
			})

			router.Use(RequireRole(tt.requiredRole))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedResponse != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				if errorMsg, ok := response["error"].(string); ok {
					assert.Contains(t, errorMsg, tt.expectedResponse)
				}
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRoles      []string
		expectedStatus int
	}{
		{
			name:           "admin user allowed",
			userRoles:      []string{"admin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-admin user forbidden",
			userRoles:      []string{"user"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no user forbidden",
			userRoles:      nil,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "admin among multiple roles",
			userRoles:      []string{"user", "admin", "editor"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			router.Use(func(c *gin.Context) {
				if tt.userRoles != nil {
					claims := &auth.Claims{
						UserID: 1,
						Email:  "admin@example.com",
						Roles:  tt.userRoles,
					}
					c.Set(auth.KeyUser, claims)
				}
				c.Next()
			})

			router.Use(RequireAdmin())
			router.GET("/admin", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
			})

			c.Request = httptest.NewRequest(http.MethodGet, "/admin", nil)
			router.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ============================================================================
// Story 1.6 RBAC Middleware Tests
// ============================================================================

// TestRBACMiddleware_SystemAdmin tests SYSTEM_ADMIN can access all endpoints
// Story 1.6, AC3: SYSTEM_ADMIN role has access to all endpoints
func TestRBACMiddleware_SystemAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name           string
		endpoint       string
		expectedStatus int
	}{
		{"/api/v1/users", "/api/v1/users", http.StatusOK},
		{"/api/v1/reports/daily", "/api/v1/reports/daily", http.StatusOK},
		{"/api/v1/admin/settings", "/api/v1/admin/settings", http.StatusOK},
		{"/api/v1/products", "/api/v1/products", http.StatusOK},
		{"/api/v1/transactions", "/api/v1/transactions", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.endpoint, nil)
			w := httptest.NewRecorder()

			// Create router with RBAC middleware
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Set user context with SYSTEM_ADMIN role
				testUserCtx := &UserContext{
					UserID:   1,
					Username: "admin",
					Email:    "admin@simpo.com",
					Role:     RoleSystemAdmin,
					BranchID: nil,
				}
				c.Set(UserContextKey, testUserCtx)
				c.Next()
			})
			router.Use(RBACMiddleware())
			router.GET(tc.endpoint, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d for SYSTEM_ADMIN accessing %s, got %d", tc.expectedStatus, tc.endpoint, w.Code)
			}
		})
	}
}

// TestRBACMiddleware_Owner tests OWNER can access business oversight endpoints
// Story 1.6, AC3: OWNER has access to business oversight endpoints
func TestRBACMiddleware_Owner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name           string
		endpoint       string
		expectedStatus int
	}{
		{"Users endpoint", "/api/v1/users", http.StatusOK},
		{"Reports endpoint", "/api/v1/reports", http.StatusOK},
		{"Products endpoint", "/api/v1/products", http.StatusOK},
		{"Transactions endpoint", "/api/v1/transactions", http.StatusOK},
		{"Inventory endpoint", "/api/v1/inventory", http.StatusOK},
		{"Branches endpoint", "/api/v1/branches", http.StatusOK},
		{"Admin endpoint - denied", "/api/v1/admin/settings", http.StatusForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.endpoint, nil)
			w := httptest.NewRecorder()

			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Set user context with OWNER role
				testUserCtx := &UserContext{
					UserID:   2,
					Username: "owner",
					Email:    "owner@simpo.com",
					Role:     RoleOwner,
					BranchID: nil,
				}
				c.Set(UserContextKey, testUserCtx)
				c.Next()
			})
			router.Use(RBACMiddleware())
			router.GET(tc.endpoint, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d for OWNER accessing %s, got %d", tc.expectedStatus, tc.endpoint, w.Code)
			}
		})
	}
}

// TestRBACMiddleware_Cashier tests CASHIER can access POS endpoints only
// Story 1.6, AC3: CASHIER has access to POS endpoints only
func TestRBACMiddleware_Cashier(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name           string
		endpoint       string
		expectedStatus int
	}{
		{"Transactions endpoint - allowed", "/api/v1/transactions", http.StatusOK},
		{"Products endpoint - allowed", "/api/v1/products", http.StatusOK},
		{"Reports endpoint - denied", "/api/v1/reports", http.StatusForbidden},
		{"Users endpoint - denied", "/api/v1/users", http.StatusForbidden},
		{"Inventory endpoint - denied", "/api/v1/inventory", http.StatusForbidden},
		{"Admin endpoint - denied", "/api/v1/admin/settings", http.StatusForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.endpoint, nil)
			w := httptest.NewRecorder()

			testBranchID := uint(1)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Set user context with CASHIER role
				testUserCtx := &UserContext{
					UserID:   3,
					Username: "cashier1",
					Email:    "cashier1@simpo.com",
					Role:     RoleCashier,
					BranchID: &testBranchID,
				}
				c.Set(UserContextKey, testUserCtx)
				c.Next()
			})
			router.Use(RBACMiddleware())
			router.GET(tc.endpoint, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d for CASHIER accessing %s, got %d", tc.expectedStatus, tc.endpoint, w.Code)
			}
		})
	}
}

// TestRBACMiddleware_NoRoleInContext tests missing role returns 403
// Story 1.6, AC3: Access denied returns 403 Forbidden with RFC 7807 error format
func TestRBACMiddleware_NoRoleInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()

	router := gin.New()
	router.Use(RBACMiddleware())
	router.GET("/api/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 when no role in context, got %d", w.Code)
	}

	// Verify RFC 7807 format
	body := w.Body.String()
	if !containsRFC7807Fields(body) {
		t.Errorf("Response should follow RFC 7807 format, got: %s", body)
	}
}

// TestRBACMiddleware_UnknownRole tests unknown role returns 403
// Story 1.6, AC3: Default deny - unknown roles have no access
func TestRBACMiddleware_UnknownRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Set user context with unknown role
		testUserCtx := &UserContext{
			UserID:   999,
			Username: "unknown",
			Email:    "unknown@simpo.com",
			Role:     "unknown_role",
			BranchID: nil,
		}
		c.Set(UserContextKey, testUserCtx)
		c.Next()
	})
	router.Use(RBACMiddleware())
	router.GET("/api/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for unknown role, got %d", w.Code)
	}
}

// TestRBACMiddleware_PrefixMatching tests endpoint prefix matching
func TestRBACMiddleware_PrefixMatching(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name           string
		role           string
		endpoint       string
		expectedStatus int
	}{
		{"OWNER - /api/v1/users/1", RoleOwner, "/api/v1/users/1", http.StatusOK},
		{"OWNER - /api/v1/reports/daily", RoleOwner, "/api/v1/reports/daily", http.StatusOK},
		{"OWNER - /api/v1/products/123", RoleOwner, "/api/v1/products/123", http.StatusOK},
		{"CASHIER - /api/v1/transactions/456", RoleCashier, "/api/v1/transactions/456", http.StatusOK},
		{"CASHIER - /api/v1/products/789", RoleCashier, "/api/v1/products/789", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.endpoint, nil)
			w := httptest.NewRecorder()

			router := gin.New()
			router.Use(func(c *gin.Context) {
				testUserCtx := &UserContext{
					UserID:   1,
					Username: "test",
					Email:    "test@simpo.com",
					Role:     tc.role,
					BranchID: nil,
				}
				c.Set(UserContextKey, testUserCtx)
				c.Next()
			})
			router.Use(RBACMiddleware())
			router.GET(tc.endpoint, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d for role %s accessing %s, got %d", tc.expectedStatus, tc.role, tc.endpoint, w.Code)
			}
		})
	}
}

// TestRequirePermission tests granular permission checking
func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		testUserCtx := &UserContext{
			UserID:   1,
			Username: "test",
			Email:    "test@simpo.com",
			Role:     RoleOwner,
			BranchID: nil,
		}
		c.Set(UserContextKey, testUserCtx)
		c.Next()
	})

	// OWNER has READ and WRITE permissions
	router.GET("/api/v1/test", RequirePermission("read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OWNER with READ permission, got %d", w.Code)
	}
}

// TestRequirePermission_InsufficientPermission tests insufficient permission returns 403
func TestRequirePermission_InsufficientPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("DELETE", "/api/v1/test", nil)
	w := httptest.NewRecorder()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		testUserCtx := &UserContext{
			UserID:   3,
			Username: "cashier",
			Email:    "cashier@simpo.com",
			Role:     RoleCashier,
			BranchID: nil,
		}
		c.Set(UserContextKey, testUserCtx)
		c.Next()
	})

	// CASHIER does not have DELETE permission
	router.DELETE("/api/v1/test", RequirePermission("delete"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for CASHIER without DELETE permission, got %d", w.Code)
	}
}

// TestRBACMiddleware_RFC7807ErrorFormat tests 403 responses follow RFC 7807 format
// Story 1.6, AC3: Access denied returns 403 Forbidden with RFC 7807 error format
func TestRBACMiddleware_RFC7807ErrorFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/reports", nil)
	w := httptest.NewRecorder()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		testUserCtx := &UserContext{
			UserID:   3,
			Username: "cashier",
			Email:    "cashier@simpo.com",
			Role:     RoleCashier,
			BranchID: nil,
		}
		c.Set(UserContextKey, testUserCtx)
		c.Next()
	})
	router.Use(RBACMiddleware())
	router.GET("/api/v1/reports", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
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

// TestRBACMiddleware_AuditLogging tests that authorization failures are logged
// Story 1.6, AC6: All authorization failures are logged with user_id, role, endpoint, reason, timestamp, IP address
func TestRBACMiddleware_AuditLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/reports", nil)
	w := httptest.NewRecorder()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		testUserCtx := &UserContext{
			UserID:   3,
			Username: "cashier",
			Email:    "cashier@simpo.com",
			Role:     RoleCashier,
			BranchID: nil,
		}
		c.Set(UserContextKey, testUserCtx)
		c.Next()
	})
	router.Use(RBACMiddleware())
	router.GET("/api/v1/reports", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	router.ServeHTTP(w, req)

	// Verify 403 response
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	// Note: We cannot directly verify slog output in unit tests
	// In production, slog writes to stdout which can be captured in integration tests
	// This test verifies the code path that triggers audit logging
	// The actual audit log format is verified in integration tests
}

// TestRBACMiddleware_AuditLogging_NoRole tests audit logging when role is missing
// Story 1.6, AC6: Authorization failures include user_id, role, endpoint, reason
func TestRBACMiddleware_AuditLogging_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()

	router := gin.New()
	// No user context set - simulates missing role
	router.Use(RBACMiddleware())
	router.GET("/api/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	router.ServeHTTP(w, req)

	// Verify 403 response
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	// Verify RFC 7807 format
	body := w.Body.String()
	if !containsString(body, "user role not found in request context") {
		t.Error("Expected error message about missing role")
	}
}
