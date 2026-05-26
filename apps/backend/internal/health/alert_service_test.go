package health

import (
	"fmt"
	"testing"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

func TestAlertService(t *testing.T) {
	t.Run("Create alert service with thresholds", func(t *testing.T) {
		thresholds := dto.AlertThresholdsConfig{
			ErrorRateMax:    0.1,
			DiskFreeMin:     20.0,
			ResponseTimeMax: 500,
		}

		service := NewAlertService(thresholds)

		if service == nil {
			t.Fatal("Expected non-nil alert service")
		}
	})

	t.Run("Evaluate all metrics and generate alerts", func(t *testing.T) {
		thresholds := dto.AlertThresholdsConfig{
			ErrorRateMax: 0.1,
			DiskFreeMin:  20.0,
		}

		service := NewAlertService(thresholds)

		// Create metrics that would trigger alerts
		metrics := dto.HealthMetrics{
			Errors: dto.ErrorsMetrics{
				Rate:          0.15, // Above threshold
				Count:         150,
				TotalRequests: 100000,
			},
			Disk: dto.DiskMetrics{
				FreePercentage: 15.0, // Below threshold
				UsedGB:        85.0,
				TotalGB:       100.0,
			},
			Database: dto.DatabaseMetrics{
				Status: "connected",
			},
			Redis: dto.RedisMetrics{
				Status: "disconnected", // Trigger critical alert
			},
		}

		checkResults := map[string]CheckResult{
			"database": {Status: CheckPass, Message: "Connected"},
			"redis":    {Status: CheckFail, Message: "Disconnected"},
			"disk":     {Status: CheckWarn, Message: "Low disk space"},
		}

		alerts := service.EvaluateAllMetrics(metrics, checkResults)

		// Should have alerts for: high error rate, low disk space, Redis failure
		if len(alerts) < 3 {
			t.Errorf("Expected at least 3 alerts, got %d", len(alerts))
		}

		// Verify alert types
		hasErrorAlert := false
		hasDiskAlert := false
		hasRedisAlert := false

		for _, alert := range alerts {
			if alert.Severity == "warning" && alert.Message != "" {
				if contains(alert.Message, "Error rate") {
					hasErrorAlert = true
				}
				if contains(alert.Message, "Disk space") {
					hasDiskAlert = true
				}
			}
			if alert.Severity == "critical" && contains(alert.Message, "Redis") {
				hasRedisAlert = true
			}
		}

		if !hasErrorAlert {
			t.Error("Expected error rate alert")
		}
		if !hasDiskAlert {
			t.Error("Expected disk space alert")
		}
		if !hasRedisAlert {
			t.Error("Expected Redis connection alert")
		}
	})

	t.Run("No alerts when all metrics healthy", func(t *testing.T) {
		thresholds := dto.AlertThresholdsConfig{
			ErrorRateMax: 0.1,
			DiskFreeMin:  20.0,
		}

		service := NewAlertService(thresholds)

		metrics := dto.HealthMetrics{
			Errors: dto.ErrorsMetrics{
				Rate:          0.05, // Below threshold
				Count:         50,
				TotalRequests: 100000,
			},
			Disk: dto.DiskMetrics{
				FreePercentage: 50.0, // Above threshold
				UsedGB:        50.0,
				TotalGB:       100.0,
			},
			Database: dto.DatabaseMetrics{
				Status: "connected",
			},
			Redis: dto.RedisMetrics{
				Status: "connected",
			},
		}

		checkResults := map[string]CheckResult{
			"database": {Status: CheckPass, Message: "Connected"},
			"redis":    {Status: CheckPass, Message: "Connected"},
			"disk":     {Status: CheckPass, Message: "Healthy"},
		}

		alerts := service.EvaluateAllMetrics(metrics, checkResults)

		if len(alerts) != 0 {
			t.Errorf("Expected no alerts when all metrics healthy, got %d", len(alerts))
		}
	})

	t.Run("Store and retrieve recent alerts", func(t *testing.T) {
		service := NewAlertService(dto.AlertThresholdsConfig{})

		// Add some alerts
		alerts := []dto.Alert{
			{Severity: "warning", Message: "Test alert 1", Timestamp: time.Now()},
			{Severity: "critical", Message: "Test alert 2", Timestamp: time.Now()},
		}

		service.StoreAlerts(alerts)

		// Retrieve stored alerts
		storedAlerts := service.GetRecentAlerts()

		if len(storedAlerts) != len(alerts) {
			t.Errorf("Expected %d stored alerts, got %d", len(alerts), len(storedAlerts))
		}
	})

	t.Run("Limit stored alerts to prevent memory issues", func(t *testing.T) {
		service := NewAlertService(dto.AlertThresholdsConfig{})

		// Add many alerts (more than the limit)
		for i := 0; i < 150; i++ {
			alerts := []dto.Alert{
				{Severity: "info", Message: fmt.Sprintf("Alert %d", i), Timestamp: time.Now()},
			}
			service.StoreAlerts(alerts)
		}

		storedAlerts := service.GetRecentAlerts()

		// Should be limited to prevent memory issues
		if len(storedAlerts) > 100 {
			t.Errorf("Expected stored alerts to be limited to 100, got %d", len(storedAlerts))
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && s[:len(substr)] == substr) ||
		(len(s) > len(substr) && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
