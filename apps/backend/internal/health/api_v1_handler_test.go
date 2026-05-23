package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ApiV1Health_ResponseFormat(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	checkers := []Checker{
		&mockChecker{name: "database", result: CheckResult{Status: CheckPass, Message: "connected"}},
		&mockChecker{name: "redis", result: CheckResult{Status: CheckPass, Message: "connected"}},
	}
	service := NewService(checkers, "1.0.0", "test")
	handler := NewHandler(service)

	router := gin.New()
	router.GET("/api/v1/health", handler.Health)

	// Act
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Assert - verify response format matches AC2
	assert.Equal(t, http.StatusOK, resp.Code)

	// Parse JSON response
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify all required fields from AC2
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "connected", response["database"])
	assert.Equal(t, "connected", response["redis"])
	assert.NotEmpty(t, response["uptime"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.NotEmpty(t, response["timestamp"])
}

func TestHandler_ApiV1Health_DatabaseDisconnected_503(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	checkers := []Checker{
		&mockChecker{name: "database", result: CheckResult{Status: CheckFail, Message: "disconnected"}},
		&mockChecker{name: "redis", result: CheckResult{Status: CheckPass, Message: "connected"}},
	}
	service := NewService(checkers, "1.0.0", "test")
	handler := NewHandler(service)

	router := gin.New()
	router.GET("/api/v1/health", handler.Health)

	// Act
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(t, http.StatusServiceUnavailable, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "unhealthy", response["status"])
	assert.Equal(t, "disconnected", response["database"])
	assert.Equal(t, "connected", response["redis"])
}

func TestHandler_ApiV1Health_RedisDisconnected_Degraded200(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	checkers := []Checker{
		&mockChecker{name: "database", result: CheckResult{Status: CheckPass, Message: "connected"}},
		&mockChecker{name: "redis", result: CheckResult{Status: CheckFail, Message: "disconnected"}},
	}
	service := NewService(checkers, "1.0.0", "test")
	handler := NewHandler(service)

	router := gin.New()
	router.GET("/api/v1/health", handler.Health)

	// Act
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code) // Degraded still returns 200

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "degraded", response["status"])
	assert.Equal(t, "connected", response["database"])
	assert.Equal(t, "disconnected", response["redis"])
}

func TestHandler_ApiV1Health_ResponseTime(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	checkers := []Checker{
		&mockChecker{name: "database", result: CheckResult{Status: CheckPass, Message: "connected"}},
		&mockChecker{name: "redis", result: CheckResult{Status: CheckPass, Message: "connected"}},
	}
	service := NewService(checkers, "1.0.0", "test")
	handler := NewHandler(service)

	router := gin.New()
	router.GET("/api/v1/health", handler.Health)

	// Act - measure response time
	start := time.Now()
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	duration := time.Since(start)

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Less(t, duration.Milliseconds(), int64(500), "Health check should respond within 500ms (AC1)")
}
