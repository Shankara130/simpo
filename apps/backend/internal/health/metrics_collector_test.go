package health

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

func TestMetricsCollector(t *testing.T) {
	t.Run("Calculate uptime percentage", func(t *testing.T) {
		startTime := time.Now().Add(-24 * time.Hour) // 1 day ago
		collector := NewMetricsCollector(startTime, "1.0.0", "test")

		uptimePercentage := collector.GetUptimePercentage()

		// After 1 day, uptime should be close to 100%
		if uptimePercentage < 99.0 {
			t.Errorf("Expected uptime percentage > 99%% after 1 day, got %.2f%%", uptimePercentage)
		}

		if uptimePercentage > 100.0 {
			t.Errorf("Uptime percentage cannot exceed 100%%, got %.2f%%", uptimePercentage)
		}
	})

	t.Run("Calculate uptime percentage for various durations", func(t *testing.T) {
		testCases := []struct {
			name              string
			duration          time.Duration
			minExpectedUptime float64
		}{
			{"1 hour", 1 * time.Hour, 99.9},
			{"1 day", 24 * time.Hour, 99.0},
			{"1 week", 7 * 24 * time.Hour, 95.0},
			{"1 month", 30 * 24 * time.Hour, 80.0},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				startTime := time.Now().Add(-tc.duration)
				collector := NewMetricsCollector(startTime, "1.0.0", "test")

				uptimePercentage := collector.GetUptimePercentage()

				if uptimePercentage < tc.minExpectedUptime {
					t.Errorf("Expected uptime percentage > %.1f%% for %s, got %.2f%%",
						tc.minExpectedUptime, tc.name, uptimePercentage)
				}
			})
		}
	})

	t.Run("Get active sessions from Redis", func(t *testing.T) {
		// Create a test Redis client (nil means Redis not available)
		var redisClient *redis.Client
		collector := NewMetricsCollector(time.Now(), "1.0.0", "test")

		ctx := context.Background()
		activeSessions := collector.GetActiveSessions(ctx, redisClient)

		// Should return 0 if Redis not available
		if activeSessions < 0 {
			t.Errorf("Active sessions cannot be negative, got %d", activeSessions)
		}
	})

	t.Run("Calculate error rate", func(t *testing.T) {
		collector := NewMetricsCollector(time.Now(), "1.0.0", "test")

		// Test error rate calculation
		errorCount := 50
		totalRequests := 100000

		errorRate := collector.CalculateErrorRate(errorCount, totalRequests)

		expectedRate := 0.05 // 50/100000 * 100 = 0.05%
		if errorRate != expectedRate {
			t.Errorf("Expected error rate %.2f%%, got %.2f%%", expectedRate, errorRate)
		}

		// Test edge cases
		t.Run("Zero total requests", func(t *testing.T) {
			rate := collector.CalculateErrorRate(10, 0)
			if rate != 0 {
				t.Errorf("Expected 0%% error rate when total requests is 0, got %.2f%%", rate)
			}
		})

		t.Run("Zero errors", func(t *testing.T) {
			rate := collector.CalculateErrorRate(0, 1000)
			if rate != 0 {
				t.Errorf("Expected 0%% error rate when errors is 0, got %.2f%%", rate)
			}
		})
	})

	t.Run("Collect health metrics", func(t *testing.T) {
		// Use existing mockChecker from service_test.go
		checkers := []Checker{
			&mockChecker{name: "database", result: CheckResult{Status: CheckPass, Message: "Connected"}},
			&mockChecker{name: "redis", result: CheckResult{Status: CheckPass, Message: "Connected"}},
			NewDiskChecker("/"),
		}

		collector := NewMetricsCollector(time.Now(), "1.0.0", "test")
		ctx := context.Background()

		metrics := collector.CollectMetrics(ctx, checkers, nil, 0, 0)

		// Verify metrics structure
		if metrics.Database.Status != "connected" {
			t.Errorf("Expected database status 'connected', got '%s'", metrics.Database.Status)
		}

		if metrics.Redis.Status != "connected" {
			t.Errorf("Expected redis status 'connected', got '%s'", metrics.Redis.Status)
		}

		// Disk metrics should be populated
		if metrics.Disk.TotalGB == 0 {
			t.Error("Expected disk total GB to be populated")
		}
	})
}

func TestAlertEvaluation(t *testing.T) {
	t.Run("Evaluate error rate against threshold", func(t *testing.T) {
		thresholds := dto.AlertThresholdsConfig{
			ErrorRateMax: 0.1, // 0.1%
		}

		// Test below threshold
		alerts := EvaluateErrorRateAlert(0.05, thresholds)
		if len(alerts) > 0 {
			t.Errorf("Expected no alerts when error rate below threshold, got %d", len(alerts))
		}

		// Test at threshold
		alerts = EvaluateErrorRateAlert(0.1, thresholds)
		if len(alerts) > 0 {
			t.Error("Expected alert when error rate at threshold")
		}

		// Test above threshold
		alerts = EvaluateErrorRateAlert(0.15, thresholds)
		if len(alerts) == 0 {
			t.Error("Expected alert when error rate above threshold")
		}

		if len(alerts) > 0 && alerts[0].Severity != "warning" {
			t.Errorf("Expected warning severity for high error rate, got %s", alerts[0].Severity)
		}
	})

	t.Run("Evaluate disk space against threshold", func(t *testing.T) {
		thresholds := dto.AlertThresholdsConfig{
			DiskFreeMin: 20.0, // 20%
		}

		// Test above threshold (healthy)
		alerts := EvaluateDiskSpaceAlert(50.0, thresholds)
		if len(alerts) > 0 {
			t.Errorf("Expected no alerts when disk space above threshold, got %d", len(alerts))
		}

		// Test at warning threshold
		alerts = EvaluateDiskSpaceAlert(20.0, thresholds)
		if len(alerts) > 0 {
			t.Error("Expected no alert when disk space at threshold (only when below)")
		}

		// Test below threshold (critical)
		alerts = EvaluateDiskSpaceAlert(8.0, thresholds)
		if len(alerts) == 0 {
			t.Error("Expected critical alert when disk space critically low")
		}

		if len(alerts) > 0 && alerts[0].Severity != "critical" {
			t.Errorf("Expected critical severity for low disk space, got %s", alerts[0].Severity)
		}
	})
}
