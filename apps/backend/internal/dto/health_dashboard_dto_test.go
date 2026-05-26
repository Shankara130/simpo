package dto

import (
	"testing"
	"time"
)

func TestEnhancedHealthDashboardResponse(t *testing.T) {
	t.Run("Create valid enhanced health dashboard response", func(t *testing.T) {
		now := time.Now()
		response := EnhancedHealthDashboardResponse{
			Status:          StatusHealthy,
			UptimePercentage: 99.8,
			Uptime:          "15d 4h 32m",
			Version:         "1.0.0",
			Timestamp:       now,
			Metrics: HealthMetrics{
				Database: DatabaseMetrics{
					Status:       "connected",
					ResponseTime: "5ms",
				},
				Redis: RedisMetrics{
					Status:       "connected",
					ResponseTime: "2ms",
				},
				Sessions: SessionsMetrics{
					Active: 15,
				},
				API: APIMetrics{
					AvgResponseTime:   "45ms",
					RequestsPerSecond: 12.5,
				},
				Errors: ErrorsMetrics{
					Rate:         0.05,
					Count:        23,
					TotalRequests: 46000,
				},
				Disk: DiskMetrics{
					UsedGB:         45.2,
					TotalGB:        100,
					FreePercentage: 54.8,
				},
			},
			Alerts: []Alert{
				{
					Severity: "warning",
					Message:  "Disk space below 20%",
					Timestamp: now.Add(-5 * time.Minute),
				},
			},
		}

		// Validate status
		if response.Status != StatusHealthy {
			t.Errorf("Expected status %s, got %s", StatusHealthy, response.Status)
		}

		// Validate uptime percentage
		if response.UptimePercentage < 99.5 {
			t.Errorf("Uptime percentage %.2f below 99.5%% threshold", response.UptimePercentage)
		}

		// Validate metrics structure
		if response.Metrics.Database.Status != "connected" {
			t.Errorf("Expected database status 'connected', got '%s'", response.Metrics.Database.Status)
		}

		if response.Metrics.Sessions.Active != 15 {
			t.Errorf("Expected 15 active sessions, got %d", response.Metrics.Sessions.Active)
		}

		if len(response.Alerts) != 1 {
			t.Errorf("Expected 1 alert, got %d", len(response.Alerts))
		}
	})

	t.Run("Validate alert severity levels", func(t *testing.T) {
		validSeverities := map[string]bool{
			"critical": true,
			"warning":  true,
			"info":     true,
		}

		alert := Alert{
			Severity: "warning",
			Message:  "Test alert",
			Timestamp: time.Now(),
		}

		if !validSeverities[alert.Severity] {
			t.Errorf("Invalid severity level: %s", alert.Severity)
		}
	})

	t.Run("Calculate error rate correctly", func(t *testing.T) {
		// Error rate calculation happens in service layer, DTO just holds the values
		// Test that DTO can hold the calculated rate correctly
		count := 50
		totalRequests := 100000
		calculatedRate := (float64(count) / float64(totalRequests)) * 100 // 0.05%

		metrics := ErrorsMetrics{
			Count:         count,
			TotalRequests: totalRequests,
			Rate:         calculatedRate,
		}

		expectedRate := 0.05 // 50/100000 * 100 = 0.05%
		if metrics.Rate != expectedRate {
			t.Errorf("Expected error rate %.2f, got %.2f", expectedRate, metrics.Rate)
		}
	})
}

func TestAlertThresholdsConfig(t *testing.T) {
	t.Run("Create alert thresholds config", func(t *testing.T) {
		config := AlertThresholdsConfig{
			ErrorRateMax:    0.1, // 0.1%
			DiskFreeMin:     20.0, // 20%
			ResponseTimeMax: 500,  // 500ms
		}

		if config.ErrorRateMax != 0.1 {
			t.Errorf("Expected error rate max 0.1, got %.2f", config.ErrorRateMax)
		}

		if config.DiskFreeMin != 20.0 {
			t.Errorf("Expected disk free min 20.0, got %.2f", config.DiskFreeMin)
		}

		if config.ResponseTimeMax != 500 {
			t.Errorf("Expected response time max 500ms, got %d", config.ResponseTimeMax)
		}
	})
}

func TestAlertResponse(t *testing.T) {
	t.Run("Create alert response with filtering", func(t *testing.T) {
		now := time.Now()
		response := AlertResponse{
			Alerts: []Alert{
				{
					Severity: "critical",
					Message:  "Database disconnected",
					Timestamp: now,
				},
				{
					Severity: "warning",
					Message:  "Disk space below 20%",
					Timestamp: now.Add(-1 * time.Hour),
				},
				{
					Severity: "info",
					Message:  "System startup",
					Timestamp: now.Add(-2 * time.Hour),
				},
			},
			Total:   3,
			Critical: 1,
			Warning:  1,
			Info:     1,
		}

		if response.Total != 3 {
			t.Errorf("Expected 3 total alerts, got %d", response.Total)
		}

		if response.Critical != 1 {
			t.Errorf("Expected 1 critical alert, got %d", response.Critical)
		}

		if len(response.Alerts) != 3 {
			t.Errorf("Expected 3 alerts in response, got %d", len(response.Alerts))
		}
	})
}
