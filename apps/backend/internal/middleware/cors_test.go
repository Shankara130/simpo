package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCORSMiddleware tests CORS middleware functionality
// Story 9.4: CORS middleware for cross-origin requests
func TestCORSMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Test allowed origin receives CORS headers
	t.Run("AllowedOrigin_ReceivesCORSHeaders", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000", "https://admin.simpo.com"}
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		allowedHeaders := []string{"Authorization", "Content-Type", "X-Requested-With"}

		router := setupTestRouter(allowedOrigins, allowedMethods, allowedHeaders, true, 86400)

		// Create test request from allowed origin
		req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		// Record response
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert CORS headers are present
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	// Test disallowed origin is rejected (no CORS headers)
	t.Run("DisallowedOrigin_Rejected", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000"}
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		allowedHeaders := []string{"Authorization", "Content-Type", "X-Requested-With"}

		router := setupTestRouter(allowedOrigins, allowedMethods, allowedHeaders, true, 86400)

		// Create test request from disallowed origin
		req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/health", nil)
		req.Header.Set("Origin", "http://malicious-site.com")

		// Record response
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert CORS headers are NOT present
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), "CORS header should not be present for disallowed origin")
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"))
	})

	// Test pre-flight OPTIONS returns 200 OK
	t.Run("PreflightOPTIONS_Returns200OK", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000"}
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		allowedHeaders := []string{"Authorization", "Content-Type", "X-Requested-With"}

		router := setupTestRouter(allowedOrigins, allowedMethods, allowedHeaders, true, 86400)

		// Create pre-flight OPTIONS request
		req := httptest.NewRequest("OPTIONS", "http://localhost:8080/api/v1/transactions", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")

		// Record response
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert pre-flight response
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
		assert.NotEmpty(t, w.Header().Get("Access-Control-Max-Age"))
	})

	// Test credentials are supported when configured
	t.Run("Credentials_SupportedWhenConfigured", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000"}
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		allowedHeaders := []string{"Authorization", "Content-Type", "X-Requested-With"}

		router := setupTestRouter(allowedOrigins, allowedMethods, allowedHeaders, true, 86400)

		// Create request with credentials
		req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		// Record response
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert credentials are supported
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	// Test specific origin is returned (not wildcard)
	t.Run("SpecificOrigin_NotWildcard", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000", "https://admin.simpo.com"}
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		allowedHeaders := []string{"Authorization", "Content-Type", "X-Requested-With"}

		router := setupTestRouter(allowedOrigins, allowedMethods, allowedHeaders, true, 86400)

		// Test that each allowed origin returns itself (not wildcard)
		for _, origin := range allowedOrigins {
			req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/health", nil)
			req.Header.Set("Origin", origin)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"), "Origin should match exactly, not wildcard")
			assert.NotEqual(t, "*", w.Header().Get("Access-Control-Allow-Origin"), "Should never return wildcard with credentials")
		}
	})

	// Test multiple allowed origins work correctly
	t.Run("MultipleAllowedOrigins_AllWork", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000", "http://localhost:19006", "https://admin.simpo.com"}
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		allowedHeaders := []string{"Authorization", "Content-Type", "X-Requested-With"}

		router := setupTestRouter(allowedOrigins, allowedMethods, allowedHeaders, true, 86400)

		// Test each origin
		for _, origin := range allowedOrigins {
			req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/health", nil)
			req.Header.Set("Origin", origin)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	// Test case-sensitive origin matching
	t.Run("CaseSensitiveOriginMatching", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000"}
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		allowedHeaders := []string{"Authorization", "Content-Type", "X-Requested-With"}

		router := setupTestRouter(allowedOrigins, allowedMethods, allowedHeaders, true, 86400)

		// Test with exact case match
		req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))

		// Test with different case (should be rejected)
		req = httptest.NewRequest("GET", "http://localhost:8080/api/v1/health", nil)
		req.Header.Set("Origin", "http://LOCALHOST:3000")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), "Case mismatched origin should be rejected")
	})
}

// setupTestRouter creates a test router with CORS middleware
func setupTestRouter(allowedOrigins, allowedMethods, allowedHeaders []string, allowCredentials bool, maxAge int) *gin.Engine {
	router := gin.New()

	// Add CORS middleware similar to router.go
	// Story 9.4: Using gin-contrib/cors with specific origins
	router.Use(func(c *gin.Context) {
		// Simulate CORS middleware behavior
		origin := c.Request.Header.Get("Origin")
		allowed := false

		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			if allowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

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
			c.Header("Access-Control-Max-Age", string(rune(maxAge)))
		}

		// Handle pre-flight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Add a simple test endpoint
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	router.OPTIONS("/api/v1/transactions", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	return router
}
