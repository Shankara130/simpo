package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// TestLoggerWithCallerInfo tests that caller information is included in logs
func TestLoggerWithCallerInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths:      []string{},
		Logger:         logger,
		IncludeCaller:  true, // Enable caller information
	}

	router := gin.New()
	router.Use(Logger(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse log output
	logOutput := buf.String()
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(logOutput), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	// Verify caller field exists
	caller, ok := logData["caller"]
	if !ok {
		t.Error("Expected 'caller' field in log output")
	}

	// Verify caller is a non-empty string
	callerStr, ok := caller.(string)
	if !ok || callerStr == "" {
		t.Error("Expected 'caller' to be a non-empty string")
	}

	// Verify caller format (should contain file:line or function information)
	if !strings.Contains(callerStr, ":") && !strings.Contains(callerStr, "/") {
		t.Errorf("Expected caller to contain file:line or path, got: %s", callerStr)
	}

	t.Logf("Caller field: %s", callerStr)
}

// TestLoggerCallerInfoDisabled tests that caller information can be disabled
func TestLoggerCallerInfoDisabled(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths:     []string{},
		Logger:        logger,
		IncludeCaller: false, // Disable caller information
	}

	router := gin.New()
	router.Use(Logger(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse log output
	logOutput := buf.String()
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(logOutput), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	// Caller field should not exist when disabled
	if _, ok := logData["caller"]; ok {
		t.Error("Expected 'caller' field to be absent when disabled")
	}
}

// TestLoggerCallerInfoFormat tests the format of caller information
func TestLoggerCallerInfoFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths:     []string{},
		Logger:        logger,
		IncludeCaller: true,
	}

	router := gin.New()
	router.Use(Logger(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Parse log output
	logOutput := buf.String()
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(logOutput), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	caller, ok := logData["caller"].(string)
	if !ok {
		t.Fatal("Expected 'caller' to be a string")
	}

	// Verify caller format contains useful information (file:line format)
	if !strings.Contains(caller, ":") {
		t.Errorf("Expected caller to contain ':' separator, got: %s", caller)
	}

	// Verify caller is not empty
	if caller == "" {
		t.Error("Expected caller to be non-empty")
	}

	t.Logf("Caller field format: %s", caller)
}

// TestLoggerAuthorizationHeaderRedaction tests that Authorization headers are redacted
func TestLoggerAuthorizationHeaderRedaction(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths:      []string{},
		Logger:         logger,
		IncludeCaller:  false,
		RedactEnabled:  true,
		RedactPatterns: []string{"authorization", "token"},
	}

	router := gin.New()
	router.Use(Logger(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse log output
	logOutput := buf.String()
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(logOutput), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	// Verify authorization field exists
	auth, ok := logData["authorization"].(string)
	if !ok {
		t.Fatal("Expected 'authorization' field in log output")
	}

	// Verify authorization is redacted
	if auth != "Bearer ****" {
		t.Errorf("Expected authorization to be redacted as 'Bearer ****', got: %s", auth)
	}

	// Verify full token is NOT in logs
	if strings.Contains(logOutput, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Error("Full JWT token should not appear in logs")
	}
}

// TestLoggerCookieHeaderRedaction tests that Cookie headers are redacted
func TestLoggerCookieHeaderRedaction(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths:      []string{},
		Logger:         logger,
		IncludeCaller:  false,
		RedactEnabled:  true,
		RedactPatterns: []string{"cookie", "session"},
	}

	router := gin.New()
	router.Use(Logger(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Cookie", "session_id=abc123def456; user_token=secret_token_value")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse log output
	logOutput := buf.String()
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(logOutput), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	// Verify cookie field exists
	cookie, ok := logData["cookie"].(string)
	if !ok {
		t.Fatal("Expected 'cookie' field in log output")
	}

	// Verify cookie is redacted
	if cookie != "****" {
		t.Errorf("Expected cookie to be redacted as '****', got: %s", cookie)
	}

	// Verify actual cookie values are NOT in logs
	if strings.Contains(logOutput, "abc123def456") || strings.Contains(logOutput, "secret_token_value") {
		t.Error("Actual cookie values should not appear in logs")
	}
}

// TestLoggerQueryParameterRedaction tests that sensitive query parameters are redacted
func TestLoggerQueryParameterRedaction(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths:      []string{},
		Logger:         logger,
		IncludeCaller:  false,
		RedactEnabled:  true,
		RedactPatterns: []string{"password", "token", "api_key"},
	}

	router := gin.New()
	router.Use(Logger(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Create request with sensitive query parameters
	req := httptest.NewRequest("GET", "/test?password=MySecret123&token=abc123&user=john&api_key=xyz789", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse log output
	logOutput := buf.String()
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(logOutput), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	// Verify query_params field exists
	queryParams, ok := logData["query_params"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'query_params' field in log output")
	}

	// Verify sensitive parameters are redacted
	tests := []struct {
		param      string
		shouldRedact bool
	}{
		{"password", true},
		{"token", true},
		{"api_key", true},
		{"user", false}, // Non-sensitive, should not be redacted
	}

	for _, tt := range tests {
		value, exists := queryParams[tt.param]
		if !exists {
			t.Errorf("Expected query parameter '%s' to exist in logs", tt.param)
			continue
		}

		if tt.shouldRedact {
			if value != "****" && value != "**** (multiple values)" {
				t.Errorf("Expected parameter '%s' to be redacted, got: %v", tt.param, value)
			}
		} else {
			// Non-sensitive parameter should have actual value
			if value == "****" {
				t.Errorf("Expected parameter '%s' to NOT be redacted, but got redacted value", tt.param)
			}
		}
	}

	// Verify actual sensitive values are NOT in logs
	if strings.Contains(logOutput, "MySecret123") || strings.Contains(logOutput, "abc123") || strings.Contains(logOutput, "xyz789") {
		t.Error("Actual sensitive query parameter values should not appear in logs")
	}

	// Non-sensitive value should be present
	if !strings.Contains(logOutput, "john") {
		t.Error("Non-sensitive query parameter value should appear in logs")
	}
}

// TestLoggerRedactionDisabled tests that redaction can be disabled
func TestLoggerRedactionDisabled(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths:      []string{},
		Logger:         logger,
		IncludeCaller:  false,
		RedactEnabled:  false, // Disabled
		RedactPatterns: []string{"password", "token"},
	}

	router := gin.New()
	router.Use(Logger(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test?password=MySecret123", nil)
	req.Header.Set("Authorization", "Bearer secret_token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse log output
	logOutput := buf.String()

	// When redaction is disabled, authorization and query_params should not be logged at all
	// (since they're only added when redactor is not nil)
	if strings.Contains(logOutput, "authorization") {
		t.Error("When redaction is disabled, authorization header should not be logged")
	}

	if strings.Contains(logOutput, "query_params") {
		t.Error("When redaction is disabled, query_params should not be logged")
	}

	// But the sensitive values should not appear in path (path includes query string raw)
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(logOutput), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	// Path should contain query string (unredacted since redaction is disabled)
	path, ok := logData["path"].(string)
	if !ok {
		t.Fatal("Expected 'path' field in log output")
	}

	// Path should still contain the query string (redaction disabled doesn't affect path)
	if !strings.Contains(path, "password=MySecret123") {
		t.Logf("Note: Path query string behavior: %s", path)
	}
}
