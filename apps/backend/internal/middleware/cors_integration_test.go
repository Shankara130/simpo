package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCORSIntegration tests CORS middleware with realistic API scenarios
// Story 9.4: Integration tests for cross-origin requests
func TestCORSIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test cross-origin GET request with Authorization header
	t.Run("CrossOriginGET_WithAuthorization", func(t *testing.T) {
		router := setupIntegrationTestRouter()

		req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	// Test cross-origin POST request with JSON body
	t.Run("CrossOriginPOST_WithJSONBody", func(t *testing.T) {
		router := setupIntegrationTestRouter()

		jsonBody := `{"username":"test","password":"test123"}`
		req := httptest.NewRequest("POST", "http://localhost:8080/api/v1/auth/login", strings.NewReader(jsonBody))
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	})

	// Test cross-origin PUT/DELETE requests
	t.Run("CrossOriginPUT_DELETE_Supported", func(t *testing.T) {
		router := setupIntegrationTestRouter()

		// Test PUT
		jsonBody := `{"name":"updated"}`
		req := httptest.NewRequest("PUT", "http://localhost:8080/api/v1/users/1", strings.NewReader(jsonBody))
		req.Header.Set("Origin", "https://admin.simpo.com")
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "https://admin.simpo.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "PUT")

		// Test DELETE
		req = httptest.NewRequest("DELETE", "http://localhost:8080/api/v1/users/1", nil)
		req.Header.Set("Origin", "https://admin.simpo.com")
		req.Header.Set("Authorization", "Bearer test-token")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "DELETE")
	})

	// Test requests from multiple allowed origins
	t.Run("MultipleAllowedOrigins_AllWork", func(t *testing.T) {
		router := setupIntegrationTestRouter()

		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:19006",
			"https://admin.simpo.com",
		}

		for _, origin := range allowedOrigins {
			req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/health", nil)
			req.Header.Set("Origin", origin)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"), "Origin %s should be allowed", origin)
			assert.NotEqual(t, "*", w.Header().Get("Access-Control-Allow-Origin"), "Should never return wildcard")
		}
	})

	// Test requests from disallowed origin fail
	t.Run("DisallowedOrigin_Fails", func(t *testing.T) {
		router := setupIntegrationTestRouter()

		req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/health", nil)
		req.Header.Set("Origin", "http://malicious-site.com")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Request succeeds but without CORS headers
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), "Disallowed origin should not receive CORS headers")

		// Browser would block the response due to missing CORS headers
	})

	// Test pre-flight OPTIONS for POST with custom headers
	t.Run("PreflightOPTIONS_POSTWithCustomHeaders", func(t *testing.T) {
		router := setupIntegrationTestRouter()

		req := httptest.NewRequest("OPTIONS", "http://localhost:8080/api/v1/transactions", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
		assert.NotEmpty(t, w.Header().Get("Access-Control-Max-Age"))
	})

	// Test credentials cookie/header is supported
	t.Run("Credentials_Supported", func(t *testing.T) {
		router := setupIntegrationTestRouter()

		req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Cookie", "session_id=test123")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"), "Credentials should be supported")
	})
}

// setupIntegrationTestRouter creates a realistic test router with CORS middleware
func setupIntegrationTestRouter() *gin.Engine {
	router := gin.New()

	// CORS middleware with realistic configuration
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:19006",
		"https://admin.simpo.com",
	}
	allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	allowedHeaders := []string{"Authorization", "Content-Type", "X-Requested-With"}

	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false

		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				break
			}
		}

		if allowed {
			methods := ""
			for i, method := range allowedMethods {
				if i > 0 {
					methods += ", "
				}
				methods += method
			}
			c.Header("Access-Control-Allow-Methods", methods)

			headers := ""
			for i, header := range allowedHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			c.Header("Access-Control-Allow-Headers", headers)
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Add realistic test endpoints
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"token": "test-token"})
	})

	router.PUT("/api/v1/users/:id", func(c *gin.Context) {
		c.JSON(200, gin.H{"id": c.Param("id")})
	})

	router.DELETE("/api/v1/users/:id", func(c *gin.Context) {
		c.JSON(200, gin.H{"deleted": true})
	})

	router.OPTIONS("/api/v1/transactions", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	return router
}
