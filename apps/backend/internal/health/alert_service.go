package health

import (
	"log/slog"
	"sync"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// AlertService evaluates health metrics against thresholds and generates alerts
// Story 6.2, Task 2: Implement alert system with thresholds
type AlertService struct {
	thresholds   dto.AlertThresholdsConfig
	alerts       []dto.Alert
	alertsMutex  sync.RWMutex
	maxAlerts    int // Maximum number of alerts to store in memory
}

// NewAlertService creates a new alert service with the given thresholds
func NewAlertService(thresholds dto.AlertThresholdsConfig) *AlertService {
	return &AlertService{
		thresholds: thresholds,
		alerts:     make([]dto.Alert, 0),
		maxAlerts:  100, // Limit memory usage
	}
}

// EvaluateAllMetrics evaluates all health metrics and generates alerts
// Story 6.2, Task 2: Evaluate health metrics against thresholds
func (as *AlertService) EvaluateAllMetrics(metrics dto.HealthMetrics, checkResults map[string]CheckResult) []dto.Alert {
	var newAlerts []dto.Alert

	// Evaluate error rate
	errorRateAlerts := EvaluateErrorRateAlert(metrics.Errors.Rate, as.thresholds)
	newAlerts = append(newAlerts, errorRateAlerts...)

	// Evaluate disk space
	diskAlerts := EvaluateDiskSpaceAlert(metrics.Disk.FreePercentage, as.thresholds)
	newAlerts = append(newAlerts, diskAlerts...)

	// Evaluate connections
	connectionAlerts := EvaluateConnectionAlerts(checkResults)
	newAlerts = append(newAlerts, connectionAlerts...)

	// Store new alerts
	as.StoreAlerts(newAlerts)

	// Return all alerts
	return as.GetRecentAlerts()
}

// StoreAlerts stores new alerts in memory with automatic cleanup
// Story 6.2, Task 2: Store recent alerts in memory cache for dashboard display
func (as *AlertService) StoreAlerts(alerts []dto.Alert) {
	as.alertsMutex.Lock()
	defer as.alertsMutex.Unlock()

	// Add new alerts
	as.alerts = append(as.alerts, alerts...)

	// Enforce maximum alert limit
	if len(as.alerts) > as.maxAlerts {
		// Remove oldest alerts (from the beginning)
		excess := len(as.alerts) - as.maxAlerts
		as.alerts = as.alerts[excess:]
		slog.Debug("Removed excess alerts from cache", "removed", excess)
	}
}

// GetRecentAlerts returns all stored alerts
func (as *AlertService) GetRecentAlerts() []dto.Alert {
	as.alertsMutex.RLock()
	defer as.alertsMutex.RUnlock()

	// Return a copy to prevent external modification
	alertsCopy := make([]dto.Alert, len(as.alerts))
	copy(alertsCopy, as.alerts)

	return alertsCopy
}

// GetAlertsBySeverity returns alerts filtered by severity level
func (as *AlertService) GetAlertsBySeverity(severity string) []dto.Alert {
	as.alertsMutex.RLock()
	defer as.alertsMutex.RUnlock()

	var filtered []dto.Alert
	for _, alert := range as.alerts {
		if alert.Severity == severity {
			filtered = append(filtered, alert)
		}
	}

	return filtered
}

// GetAlertCount returns the count of alerts by severity
func (as *AlertService) GetAlertCount() (total, critical, warning, info int) {
	as.alertsMutex.RLock()
	defer as.alertsMutex.RUnlock()

	total = len(as.alerts)
	for _, alert := range as.alerts {
		switch alert.Severity {
		case "critical":
			critical++
		case "warning":
			warning++
		case "info":
			info++
		}
	}

	return total, critical, warning, info
}

// ClearAlerts removes all stored alerts
// Useful for testing or manual cleanup
func (as *AlertService) ClearAlerts() {
	as.alertsMutex.Lock()
	defer as.alertsMutex.Unlock()

	as.alerts = make([]dto.Alert, 0)
	slog.Debug("Cleared all alerts from cache")
}

// GetThresholds returns the current alert thresholds
func (as *AlertService) GetThresholds() dto.AlertThresholdsConfig {
	return as.thresholds
}

// UpdateThresholds updates the alert thresholds
func (as *AlertService) UpdateThresholds(newThresholds dto.AlertThresholdsConfig) {
	as.thresholds = newThresholds
	slog.Info("Updated alert thresholds", "thresholds", newThresholds)
}

// CheckThreshold evaluates a single value against a threshold and returns true if it exceeds the threshold
func (as *AlertService) CheckThreshold(value float64, threshold float64) bool {
	return value > threshold
}

// GenerateInfoAlert creates an info-level alert
func (as *AlertService) GenerateInfoAlert(message string) {
	alert := dto.Alert{
		Severity:  "info",
		Message:   message,
		Timestamp: time.Now(),
	}

	as.StoreAlerts([]dto.Alert{alert})
}

// PruneOldAlerts removes alerts older than the specified duration
func (as *AlertService) PruneOldAlerts(maxAge time.Duration) {
	as.alertsMutex.Lock()
	defer as.alertsMutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var remaining []dto.Alert

	for _, alert := range as.alerts {
		if alert.Timestamp.After(cutoff) {
			remaining = append(remaining, alert)
		}
	}

	removed := len(as.alerts) - len(remaining)
	as.alerts = remaining

	if removed > 0 {
		slog.Debug("Pruned old alerts", "removed", removed, "max_age", maxAge)
	}
}
