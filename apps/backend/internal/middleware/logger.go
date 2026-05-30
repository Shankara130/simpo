package middleware

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/utils"
)

// LoggerConfig defines the configuration for the logger middleware
type LoggerConfig struct {
	// SkipPaths is a list of paths that should not be logged
	SkipPaths []string
	// Logger is the slog logger instance to use
	Logger *slog.Logger
	// IncludeCaller adds caller information (function, file, line) to logs
	IncludeCaller bool
	// RedactEnabled enables sensitive data redaction
	RedactEnabled bool
	// RedactPatterns defines which field names should be redacted
	RedactPatterns []string
}

// DefaultLoggerConfig returns a default configuration for the logger middleware
func DefaultLoggerConfig() *LoggerConfig {
	// Create a JSON logger that writes to stdout
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &LoggerConfig{
		SkipPaths: []string{"/health"},
		Logger:    logger,
	}
}

// NewLoggerConfig creates a logger configuration from logging config
func NewLoggerConfig(logLevel slog.Level, skipPaths []string, includeCaller, redactEnabled bool, redactPatterns []string) *LoggerConfig {
	// Create a JSON logger that writes to stdout
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	return &LoggerConfig{
		SkipPaths:      skipPaths,
		Logger:         logger,
		IncludeCaller:  includeCaller,
		RedactEnabled:  redactEnabled,
		RedactPatterns: redactPatterns,
	}
}

// Logger returns a Gin middleware for structured request logging
func Logger(config *LoggerConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	// Build a map for fast path lookup
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Process request
		c.Next()

		// Skip logging for specified paths
		if skipPaths[path] {
			return
		}

		// Calculate request duration
		duration := time.Since(start)

		// Get response status
		statusCode := c.Writer.Status()

		// Determine log level based on status code
		level := slog.LevelInfo
		if statusCode >= 500 {
			level = slog.LevelError
		} else if statusCode >= 400 {
			level = slog.LevelWarn
		}

		// Get redactor if enabled
		redactor := getRedactor(config)

		// Add query string to path if present (redact if enabled)
		if raw != "" {
			redactedQuery := raw
			if redactor != nil {
				redactedQuery = redactor.RedactQueryString(raw)
			}
			path = path + "?" + redactedQuery
		}

		// Log structured data
		logArgs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.Duration("duration", duration),
			slog.String("duration_ms", formatDuration(duration)),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.Int("response_size", c.Writer.Size()),
		}

		// Add important headers if redaction is enabled
		if redactor != nil {
			// Log Authorization header (redacted)
			if auth := c.GetHeader("Authorization"); auth != "" {
				logArgs = append(logArgs, slog.String("authorization", redactHeader("Authorization", auth, redactor)))
			}
			// Log Cookie header (redacted)
			if cookie := c.GetHeader("Cookie"); cookie != "" {
				logArgs = append(logArgs, slog.String("cookie", redactHeader("Cookie", cookie, redactor)))
			}
		}

		// Add redacted query parameters if present
		if raw != "" && redactor != nil {
			redactedParams := redactQueryParams(raw, redactor)
			if len(redactedParams) > 0 {
				// Add query params as a nested map
				logArgs = append(logArgs, slog.Any("query_params", redactedParams))
			}
		}

		// Add caller information if enabled
		if config.IncludeCaller {
			logArgs = append(logArgs, slog.String("caller", getCallerInfo(2)))
		}

		logger.Log(c.Request.Context(), level, "HTTP Request", logArgs...)

		// Log error if present
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("Request error",
					slog.String("request_id", requestID),
					slog.String("error", e.Error()),
				)
			}
		}
	}
}

// formatDuration formats duration to milliseconds string
func formatDuration(d time.Duration) string {
	return d.Round(time.Millisecond).String()
}

// formatCaller formats caller information from runtime.Caller
func formatCaller(file string, line int) string {
	// Get the base filename for cleaner logs
	filename := filepath.Base(file)
	return fmt.Sprintf("%s:%d", filename, line)
}

// getCallerInfo retrieves caller information using runtime.Caller
// skip is the number of stack frames to skip
func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return formatCaller(file, line)
}

// getRedactor returns a Redactor instance if redaction is enabled
func getRedactor(config *LoggerConfig) *utils.Redactor {
	if config.RedactEnabled {
		redactor := utils.NewRedactor()
		// Add custom patterns if specified
		for _, pattern := range config.RedactPatterns {
			redactor.AddPattern(pattern)
		}
		return redactor
	}
	return nil
}

// redactHeader returns redacted header value if the header should be redacted
func redactHeader(name, value string, redactor *utils.Redactor) string {
	if redactor != nil {
		return redactor.RedactHeader(name, value)
	}
	return value
}

// redactQueryParams returns a map of query parameters with sensitive values redacted
func redactQueryParams(query string, redactor *utils.Redactor) map[string]string {
	result := make(map[string]string)
	if query == "" {
		return result
	}

	// Use Redactor's RedactQueryString to get redacted query string
	redactedQuery := query
	if redactor != nil {
		redactedQuery = redactor.RedactQueryString(query)
	}

	// Parse the redacted query string for logging
	params, err := url.ParseQuery(redactedQuery)
	if err != nil {
		// If parsing fails, return empty map (safety fallback)
		return result
	}

	for key, values := range params {
		if len(values) == 0 {
			result[key] = ""
			continue
		}
		// Use the redacted value(s)
		if len(values) == 1 {
			result[key] = values[0]
		} else {
			result[key] = strings.Join(values, ", ")
		}
	}

	return result
}

// LoggerWithConfig returns a Gin middleware for structured request logging with custom configuration
func LoggerWithConfig(skipPaths []string, logLevel slog.Level) gin.HandlerFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	config := &LoggerConfig{
		SkipPaths: skipPaths,
		Logger:    logger,
	}

	return Logger(config)
}
