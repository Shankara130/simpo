package health

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
)

// MetricsCollector collects and aggregates health metrics for the dashboard
// Story 6.2, Task 1: Enhanced health metrics collection
type MetricsCollector struct {
	startTime   time.Time
	version     string
	environment string
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(startTime time.Time, version, environment string) *MetricsCollector {
	return &MetricsCollector{
		startTime:   startTime,
		version:     version,
		environment: environment,
	}
}

// GetUptimePercentage calculates the system uptime as a percentage
// Story 6.2, AC1: System uptime percentage (>99.5% target)
//
// NOTE: This implementation returns 100% as it assumes continuous uptime since service start.
// For production use with actual downtime tracking, implement:
// - Persistent downtime event logging
// - Incident start/end time tracking
// - Calculation: (total_time - downtime) / total_time * 100
func (mc *MetricsCollector) GetUptimePercentage() float64 {
	// For now, return 100% as we track no downtime events
	// This represents uptime since the service was last started
	return 100.0
}

// GetActiveSessions retrieves the count of active user sessions from Redis
// Story 6.2, AC4: Active user sessions count
func (mc *MetricsCollector) GetActiveSessions(ctx context.Context, redisClient *redis.Client) int {
	if redisClient == nil {
		slog.Debug("Redis not available - session count unavailable")
		return 0
	}

	// Count session keys in Redis
	pattern := "session:*"
	keys, err := redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		slog.Error("Failed to count active sessions", "error", err)
		return 0
	}

	return len(keys)
}

// CalculateErrorRate calculates the error rate as a percentage
// Story 6.2, AC5: Error rate calculator (error count / total requests)
func (mc *MetricsCollector) CalculateErrorRate(errorCount, totalRequests int) float64 {
	if totalRequests == 0 {
		return 0.0
	}

	rate := (float64(errorCount) / float64(totalRequests)) * 100
	return rate
}

// CollectMetrics gathers all health metrics from checkers and system state
// Story 6.2, AC1-7: Dashboard metrics collection
func (mc *MetricsCollector) CollectMetrics(
	ctx context.Context,
	checkers []Checker,
	redisClient *redis.Client,
	errorCount, totalRequests int,
) dto.HealthMetrics {
	start := time.Now()

	// Run all checkers to get their results
	checkResults := make(map[string]CheckResult)
	for _, checker := range checkers {
		if checker == nil {
			continue
		}
		result := checker.Check(ctx)
		checkResults[checker.Name()] = result
	}

	// Extract database metrics
	dbMetrics := dto.DatabaseMetrics{
		Status: "disconnected",
	}
	if dbResult, ok := checkResults["database"]; ok {
		if dbResult.Status == CheckPass {
			dbMetrics.Status = "connected"
			dbMetrics.ResponseTime = dbResult.ResponseTime
		}
	}

	// Extract Redis metrics
	redisMetrics := dto.RedisMetrics{
		Status: "disconnected",
	}
	if redisResult, ok := checkResults["redis"]; ok {
		if redisResult.Status == CheckPass {
			redisMetrics.Status = "connected"
			redisMetrics.ResponseTime = redisResult.ResponseTime
		}
	}

	// Get active sessions
	activeSessions := mc.GetActiveSessions(ctx, redisClient)
	sessionsMetrics := dto.SessionsMetrics{
		Active: activeSessions,
	}

	// Calculate API metrics (simplified - would use request tracking in production)
	duration := time.Since(start)
	apiMetrics := dto.APIMetrics{
		AvgResponseTime:   fmt.Sprintf("%dms", duration.Milliseconds()),
		RequestsPerSecond: 0, // Would be calculated from request tracking
	}

	// Calculate error metrics
	errorRate := mc.CalculateErrorRate(errorCount, totalRequests)
	errorsMetrics := dto.ErrorsMetrics{
		Rate:          errorRate,
		Count:         errorCount,
		TotalRequests: totalRequests,
	}

	// Extract disk metrics
	diskMetrics := dto.DiskMetrics{}
	if diskResult, ok := checkResults["disk"]; ok {
		if details, ok := diskResult.Details.(map[string]interface{}); ok {
			if usedGB, ok := details["used_gb"].(float64); ok {
				diskMetrics.UsedGB = usedGB
			}
			if totalGB, ok := details["total_gb"].(float64); ok {
				diskMetrics.TotalGB = totalGB
			}
			if freePercent, ok := details["free_percentage"].(float64); ok {
				diskMetrics.FreePercentage = freePercent
			}
		}
	}

	return dto.HealthMetrics{
		Database: dbMetrics,
		Redis:    redisMetrics,
		Sessions: sessionsMetrics,
		API:      apiMetrics,
		Errors:   errorsMetrics,
		Disk:     diskMetrics,
	}
}

// EvaluateErrorRateAlert evaluates error rate against thresholds and generates alerts
// Story 6.2, Task 2: Alert generation for high error rate
func EvaluateErrorRateAlert(errorRate float64, thresholds dto.AlertThresholdsConfig) []dto.Alert {
	var alerts []dto.Alert

	if errorRate > thresholds.ErrorRateMax {
		alerts = append(alerts, dto.Alert{
			Severity:  "warning",
			Message:   fmt.Sprintf("Error rate exceeds threshold: %.2f%% (max: %.2f%%)", errorRate, thresholds.ErrorRateMax),
			Timestamp: time.Now(),
		})
	}

	return alerts
}

// EvaluateDiskSpaceAlert evaluates disk space against thresholds and generates alerts
// Story 6.2, Task 2: Alert generation for low disk space
func EvaluateDiskSpaceAlert(freePercentage float64, thresholds dto.AlertThresholdsConfig) []dto.Alert {
	var alerts []dto.Alert

	if freePercentage < 10.0 {
		alerts = append(alerts, dto.Alert{
			Severity:  "critical",
			Message:   fmt.Sprintf("Critical disk space: %.2f%% free (minimum: 10%%)", freePercentage),
			Timestamp: time.Now(),
		})
	} else if freePercentage < thresholds.DiskFreeMin {
		alerts = append(alerts, dto.Alert{
			Severity:  "warning",
			Message:   fmt.Sprintf("Disk space below threshold: %.2f%% free (minimum: %.2f%%)", freePercentage, thresholds.DiskFreeMin),
			Timestamp: time.Now(),
		})
	}

	return alerts
}

// EvaluateConnectionAlerts generates alerts for database and Redis connection failures
// Story 6.2, Task 2: Alert generation for DB/Redis failures
func EvaluateConnectionAlerts(checkResults map[string]CheckResult) []dto.Alert {
	var alerts []dto.Alert

	// Check database connection
	if dbResult, ok := checkResults["database"]; ok {
		if dbResult.Status == CheckFail {
			alerts = append(alerts, dto.Alert{
				Severity:  "critical",
				Message:   "Database connection failed",
				Timestamp: time.Now(),
			})
		}
	}

	// Check Redis connection
	if redisResult, ok := checkResults["redis"]; ok {
		if redisResult.Status == CheckFail {
			alerts = append(alerts, dto.Alert{
				Severity:  "critical",
				Message:   "Redis connection failed",
				Timestamp: time.Now(),
			})
		}
	}

	return alerts
}

// GetSessionCountFromSessionManager retrieves session count using SessionManager
// Useful when SessionManager is available instead of raw Redis client
func (mc *MetricsCollector) GetSessionCountFromSessionManager(ctx context.Context, sessionManager *middleware.SessionManager) int {
	if sessionManager == nil {
		return 0
	}

	// Scan for session keys using pattern matching
	// Note: This requires exposing a method to count sessions from SessionManager
	// For now, return 0 as placeholder
	return 0
}
