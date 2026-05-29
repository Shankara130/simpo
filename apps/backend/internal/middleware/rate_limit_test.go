package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// MockStorage is a mock implementation of Storage interface for testing
type MockStorage struct {
	store map[string]*rate.Limiter
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		store: make(map[string]*rate.Limiter),
	}
}

func (m *MockStorage) Add(key string, limiter *rate.Limiter) bool {
	if _, exists := m.store[key]; exists {
		return false
	}
	m.store[key] = limiter
	return true
}

func (m *MockStorage) Get(key string) (*rate.Limiter, bool) {
	limiter, exists := m.store[key]
	return limiter, exists
}

// TestNewRateLimitMiddleware tests the NewRateLimitMiddleware function
func TestNewRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		window       time.Duration
		requests     int
		keyFunc      func(*gin.Context) string
		store        Storage
		testRequests int
		expectBlocks int
		description  string
	}{
		{
			name:     "basic rate limiting with IP-based key",
			window:   time.Second,
			requests: 2,
			keyFunc: func(c *gin.Context) string {
				return c.ClientIP()
			},
			store:        NewMockStorage(),
			testRequests: 5,
			expectBlocks: 3, // First 2 pass, next 3 blocked
			description:  "2 requests per second should allow first 2, block remaining",
		},
		{
			name:     "rate limiting with custom key function",
			window:   2 * time.Second,
			requests: 3,
			keyFunc: func(c *gin.Context) string {
				return c.GetHeader("X-User-ID")
			},
			store:        NewMockStorage(),
			testRequests: 4,
			expectBlocks: 1, // First 3 pass, last 1 blocked
			description:  "3 requests per 2 seconds with user header key",
		},
		{
			name:     "rate limiting with nil store uses default",
			window:   time.Second,
			requests: 1,
			keyFunc: func(c *gin.Context) string {
				return "test-key"
			},
			store:        nil, // Should use default store
			testRequests: 3,
			expectBlocks: 2, // First 1 passes, next 2 blocked
			description:  "nil store should use default LRU store",
		},
		{
			name:     "high rate limit allows many requests",
			window:   time.Second,
			requests: 100,
			keyFunc: func(c *gin.Context) string {
				return c.ClientIP()
			},
			store:        NewMockStorage(),
			testRequests: 10,
			expectBlocks: 0, // All should pass with high limit
			description:  "100 requests per second should allow all 10 test requests",
		},
		{
			name:     "very strict rate limiting",
			window:   10 * time.Second,
			requests: 1,
			keyFunc: func(c *gin.Context) string {
				return "single-key"
			},
			store:        NewMockStorage(),
			testRequests: 2,
			expectBlocks: 1, // Only first request passes
			description:  "1 request per 10 seconds should be very restrictive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRateLimitMiddleware(tt.window, tt.requests, tt.keyFunc, tt.store)
			assert.NotNil(t, middleware, "Middleware should not be nil")

			router := gin.New()
			router.Use(apiErrors.ErrorHandler())
			router.Use(middleware)
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			successCount := 0
			blockedCount := 0

			// Make test requests
			for i := 0; i < tt.testRequests; i++ {
				req := httptest.NewRequest("GET", "/test", nil)

				// Set custom header if key function uses it
				if tt.keyFunc != nil {
					if tt.name == "rate limiting with custom key function" {
						req.Header.Set("X-User-ID", "user123")
					}
				}

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					successCount++
				} else if w.Code == http.StatusTooManyRequests {
					blockedCount++

					// Verify rate limit headers are set
					assert.NotEmpty(t, w.Header().Get("Retry-After"), "Retry-After header should be set")
					assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"), "X-RateLimit-Limit header should be set")
					assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"), "X-RateLimit-Remaining header should be set")
					assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"), "X-RateLimit-Reset header should be set")

					var response map[string]interface{}
					assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

					assert.False(t, response["success"].(bool))
					errorObj := response["error"].(map[string]interface{})
					assert.Equal(t, "Rate Limit Exceeded", errorObj["title"])
					assert.Contains(t, errorObj, "retry_after")
				}
			}

			// Verify expectations
			assert.Equal(t, tt.expectBlocks, blockedCount,
				"Expected %d blocked requests, got %d. %s", tt.expectBlocks, blockedCount, tt.description)
			assert.Equal(t, tt.testRequests-tt.expectBlocks, successCount,
				"Expected %d successful requests, got %d", tt.testRequests-tt.expectBlocks, successCount)
		})
	}
}

// TestRateLimitMiddleware_DifferentKeys tests that different keys have separate limits
func TestRateLimitMiddleware_DifferentKeys(t *testing.T) {
	keyFunc := func(c *gin.Context) string {
		return c.GetHeader("X-Client-ID")
	}

	middleware := NewRateLimitMiddleware(time.Second, 1, keyFunc, NewMockStorage())

	router := gin.New()
	router.Use(apiErrors.ErrorHandler())
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with client1
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Client-ID", "client1")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code, "First request from client1 should succeed")

	// Test with client2 - should also succeed (different key)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Client-ID", "client2")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code, "First request from client2 should succeed")

	// Test with client1 again - should be blocked
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-Client-ID", "client1")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code, "Second request from client1 should be blocked")
}

// TestRateLimitMiddleware_Headers tests rate limit header values
func TestRateLimitMiddleware_Headers(t *testing.T) {
	middleware := NewRateLimitMiddleware(time.Second, 5, func(c *gin.Context) string {
		return "test"
	}, NewMockStorage())

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First, verify headers on successful responses (200 OK)
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code == http.StatusOK {
		// Verify headers are present on successful responses
		limit := w1.Header().Get("X-RateLimit-Limit")
		assert.NotEmpty(t, limit, "X-RateLimit-Limit should be present on success")

		remaining := w1.Header().Get("X-RateLimit-Remaining")
		assert.NotEmpty(t, remaining, "X-RateLimit-Remaining should be present on success")

		reset := w1.Header().Get("X-RateLimit-Reset")
		assert.NotEmpty(t, reset, "X-RateLimit-Reset should be present on success")
	}

	// Now test headers on blocked responses (429)
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			// Verify header values
			retryAfter := w.Header().Get("Retry-After")
			assert.NotEmpty(t, retryAfter, "Retry-After should not be empty")

			retrySeconds, err := strconv.Atoi(retryAfter)
			assert.NoError(t, err, "Retry-After should be a valid integer")
			assert.Greater(t, retrySeconds, 0, "Retry-After should be positive")

			limit := w.Header().Get("X-RateLimit-Limit")
			assert.Equal(t, "5", limit, "X-RateLimit-Limit should be 5")

			remaining := w.Header().Get("X-RateLimit-Remaining")
			assert.Equal(t, "0", remaining, "X-RateLimit-Remaining should be 0 when blocked")

			reset := w.Header().Get("X-RateLimit-Reset")
			assert.NotEmpty(t, reset, "X-RateLimit-Reset should not be empty")

			resetTime, err := strconv.ParseInt(reset, 10, 64)
			assert.NoError(t, err, "X-RateLimit-Reset should be a valid unix timestamp")
			assert.Greater(t, resetTime, time.Now().Unix(), "Reset time should be in the future")

			break
		}
	}
}

// TestRateLimitMiddleware_JWTContextExtraction tests JWT context extraction for rate limiting
// Story 9.3, Task 1, AC1: Rate limiter tracks requests by JWT token/user ID
func TestRateLimitMiddleware_JWTContextExtraction(t *testing.T) {
	// Create a key function that mimics what will be in router.go
	// Story 9.3: Extract user ID from JWT context, fallback to IP
	keyFunc := func(c *gin.Context) string {
		// Try to get user ID from JWT context first (auth middleware sets this)
		if userValue, exists := c.Get("user"); exists {
			// Type assertion with safety check - handle nil and invalid types
			if claims, ok := userValue.(*auth.Claims); ok && claims != nil && claims.UserID > 0 {
				// Track by user ID for authenticated requests
				return fmt.Sprintf("user:%d", claims.UserID)
			}
		}
		// Fallback to IP for unauthenticated requests
		ip := c.ClientIP()
		if ip == "" {
			ip = c.GetHeader("X-Forwarded-For")
			if ip == "" {
				ip = c.GetHeader("X-Real-IP")
			}
			if ip == "" {
				ip = "unknown"
			}
		}
		return fmt.Sprintf("ip:%s", ip)
	}

	t.Run("authenticated requests tracked by user ID", func(t *testing.T) {
		// Test key function with user claims set
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("user", &auth.Claims{UserID: 123})
		key := keyFunc(c)
		assert.Equal(t, "user:123", key, "Key should be user:123 when user claims is set in context")
	})

	t.Run("authenticated requests with UserID=0 fall back to IP", func(t *testing.T) {
		// Test that UserID=0 is excluded and falls back to IP
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("user", &auth.Claims{UserID: 0})
		key := keyFunc(c)
		assert.Contains(t, key, "ip:", "Key should be IP-based when UserID is 0")
	})

	t.Run("authenticated requests with nil claims fall back to IP", func(t *testing.T) {
		// Test that nil claims fall back to IP
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("user", (*auth.Claims)(nil)) // Set nil pointer
		key := keyFunc(c)
		assert.Contains(t, key, "ip:", "Key should be IP-based when claims is nil")
	})

	t.Run("authenticated requests with invalid type fall back to IP", func(t *testing.T) {
		// Test that invalid type in context falls back to IP
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("user", "invalid-type") // Set wrong type
		key := keyFunc(c)
		assert.Contains(t, key, "ip:", "Key should be IP-based when context has wrong type")
	})

	t.Run("unauthenticated requests tracked by IP", func(t *testing.T) {
		// Test key function without user set
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/test", nil)
		// Don't set user - simulates unauthenticated request
		key := keyFunc(c)
		assert.Contains(t, key, "ip:", "Key should be IP-based when user is not set")
		// IP address varies by environment, just check it starts with "ip:"
		assert.True(t, len(key) > 4, "IP key should have content after 'ip:' prefix")
	})

	t.Run("different users have separate rate limits", func(t *testing.T) {
		// Create two contexts with different user IDs
		c1, _ := gin.CreateTestContext(httptest.NewRecorder())
		c1.Request = httptest.NewRequest("GET", "/test", nil)
		c1.Set("user", &auth.Claims{UserID: 100})
		key1 := keyFunc(c1)
		assert.Equal(t, "user:100", key1)

		c2, _ := gin.CreateTestContext(httptest.NewRecorder())
		c2.Request = httptest.NewRequest("GET", "/test", nil)
		c2.Set("user", &auth.Claims{UserID: 200})
		key2 := keyFunc(c2)
		assert.Equal(t, "user:200", key2)

		// Different keys should result in separate rate limits
		assert.NotEqual(t, key1, key2, "Different users should have different rate limit keys")
	})
}

// TestRateLimitMiddleware_SameUserSharesLimit tests that same user ID shares rate limit
// Story 9.3, Task 6, AC1: Authenticated users tracked by user ID (same ID shares limit)
func TestRateLimitMiddleware_SameUserSharesLimit(t *testing.T) {
	keyFunc := func(c *gin.Context) string {
		if userValue, exists := c.Get("user"); exists {
			if claims, ok := userValue.(*auth.Claims); ok && claims != nil && claims.UserID > 0 {
				return fmt.Sprintf("user:%d", claims.UserID)
			}
		}
		return "ip:127.0.0.1"
	}

	// Test that requests from same user share the rate limit
	testUserID := uint(123)

	// Create two contexts with same user ID
	c1, _ := gin.CreateTestContext(httptest.NewRecorder())
	c1.Request = httptest.NewRequest("GET", "/test", nil)
	c1.Set("user", &auth.Claims{UserID: testUserID})
	key1 := keyFunc(c1)

	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request = httptest.NewRequest("GET", "/test", nil)
	c2.Set("user", &auth.Claims{UserID: testUserID})
	key2 := keyFunc(c2)

	assert.Equal(t, key1, key2, "Same user should always get same key")
	assert.Equal(t, "user:123", key1, "Key should be user:123 for user ID 123")
}

// TestRateLimitMiddleware_IPFallbackForUnauthenticated tests IP-based fallback
// Story 9.3, Task 1, AC1: Fallback to IP address for unauthenticated requests
func TestRateLimitMiddleware_IPFallbackForUnauthenticated(t *testing.T) {
	keyFunc := func(c *gin.Context) string {
		if userValue, exists := c.Get("user"); exists {
			if claims, ok := userValue.(*auth.Claims); ok && claims != nil && claims.UserID > 0 {
				return fmt.Sprintf("user:%d", claims.UserID)
			}
		}
		ip := c.ClientIP()
		if ip == "" {
			ip = c.GetHeader("X-Forwarded-For")
			if ip == "" {
				ip = c.GetHeader("X-Real-IP")
			}
			if ip == "" {
				ip = "unknown"
			}
		}
		return fmt.Sprintf("ip:%s", ip)
	}

	// Test without user set (unauthenticated)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/test", nil)
	key := keyFunc(c)
	assert.Contains(t, key, "ip:", "Unauthenticated request should use IP-based key")
	assert.NotContains(t, key, "user:", "Unauthenticated request should not use user-based key")
}

// TestRateLimitMiddleware_WindowRecovery tests that rate limit recovers after window expires
// Story 9.3, Task 7, AC2: Rate limit recovery after window expires
func TestRateLimitMiddleware_WindowRecovery(t *testing.T) {
	// Use a very short window for testing
	window := 100 * time.Millisecond
	middleware := NewRateLimitMiddleware(window, 2, func(c *gin.Context) string {
		return "test-key"
	}, NewMockStorage())

	router := gin.New()
	router.Use(apiErrors.ErrorHandler())
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 3 requests - first 2 should succeed, third should be blocked
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code, "First request should succeed")

	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code, "Second request should succeed")

	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code, "Third request should be blocked")

	// Wait for window to expire
	time.Sleep(window + 50*time.Millisecond)

	// Make another request - should succeed now
	req4 := httptest.NewRequest("GET", "/test", nil)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code, "Request after window expires should succeed")
}

// TestRateLimitMiddleware_ConcurrentRequests tests concurrent request handling
// Story 9.3, Task 7, AC3: Concurrent requests are handled correctly
func TestRateLimitMiddleware_ConcurrentRequests(t *testing.T) {
	middleware := NewRateLimitMiddleware(time.Second, 10, func(c *gin.Context) string {
		return "test-key"
	}, NewMockStorage())

	router := gin.New()
	router.Use(apiErrors.ErrorHandler())
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Launch 15 concurrent requests (limit is 10)
	// Use atomic counters to prevent data races
	var successCount int64
	var blockedCount int64
	done := make(chan bool, 15)

	for i := 0; i < 15; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				atomic.AddInt64(&successCount, 1)
			} else if w.Code == http.StatusTooManyRequests {
				atomic.AddInt64(&blockedCount, 1)
			}
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 15; i++ {
		<-done
	}

	// At most 10 should succeed, at least 5 should be blocked
	assert.LessOrEqual(t, int(successCount), 10, "At most 10 requests should succeed")
	assert.GreaterOrEqual(t, int(blockedCount), 5, "At least 5 requests should be blocked")
	assert.Equal(t, int64(15), successCount+blockedCount, "All requests should be accounted for")
}

// TestRateLimitMiddleware_MultipleEndpoints tests rate limiting across different endpoints
// Story 9.3, Task 7, AC1: Test rate limiting across multiple endpoints
func TestRateLimitMiddleware_MultipleEndpoints(t *testing.T) {
	middleware := NewRateLimitMiddleware(time.Second, 3, func(c *gin.Context) string {
		return "test-user"
	}, NewMockStorage())

	router := gin.New()
	router.Use(apiErrors.ErrorHandler())
	router.Use(middleware)
	router.GET("/endpoint1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"endpoint": "1"})
	})
	router.GET("/endpoint2", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"endpoint": "2"})
	})
	router.POST("/endpoint3", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"endpoint": "3"})
	})

	successCount := 0
	blockedCount := 0

	// Make 5 requests across different endpoints
	requests := []struct {
		method string
		path   string
	}{
		{"GET", "/endpoint1"},
		{"GET", "/endpoint2"},
		{"POST", "/endpoint3"},
		{"GET", "/endpoint1"},
		{"GET", "/endpoint2"},
	}
	for _, reqData := range requests {
		req := httptest.NewRequest(reqData.method, reqData.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			successCount++
		} else if w.Code == http.StatusTooManyRequests {
			blockedCount++
		}
	}

	// First 3 should succeed (regardless of endpoint), next 2 should be blocked
	assert.Equal(t, 3, successCount, "First 3 requests should succeed")
	assert.Equal(t, 2, blockedCount, "Remaining requests should be blocked")
}
