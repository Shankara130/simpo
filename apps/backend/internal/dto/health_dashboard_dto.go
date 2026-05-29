package dto

import "time"

// HealthStatus represents the overall health status
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// EnhancedHealthDashboardResponse represents the comprehensive health metrics for admin dashboard
// Story 6.2, AC1-7: Dashboard displays system uptime, DB/Redis status, sessions, errors, response time, disk
type EnhancedHealthDashboardResponse struct {
	Status           HealthStatus  `json:"status"`
	UptimePercentage float64       `json:"uptime_percentage"`
	Uptime           string        `json:"uptime"`
	Version          string        `json:"version"`
	Timestamp        time.Time     `json:"timestamp"`
	Metrics          HealthMetrics `json:"metrics"`
	Alerts           []Alert       `json:"alerts"`
	Environment      string        `json:"environment,omitempty"`
}

// HealthMetrics contains all health monitoring metrics
type HealthMetrics struct {
	Database DatabaseMetrics `json:"database"`
	Redis    RedisMetrics    `json:"redis"`
	Sessions SessionsMetrics `json:"sessions"`
	API      APIMetrics      `json:"api"`
	Errors   ErrorsMetrics   `json:"errors"`
	Disk     DiskMetrics     `json:"disk"`
	Backup   BackupMetrics   `json:"backup"` // Story 6.3: Backup status metrics
}

// DatabaseMetrics represents database health information
type DatabaseMetrics struct {
	Status       string `json:"status"`
	ResponseTime string `json:"response_time,omitempty"`
}

// RedisMetrics represents Redis cache health information
type RedisMetrics struct {
	Status       string `json:"status"`
	ResponseTime string `json:"response_time,omitempty"`
}

// SessionsMetrics represents active session information
type SessionsMetrics struct {
	Active int `json:"active"`
}

// APIMetrics represents API performance information
type APIMetrics struct {
	AvgResponseTime   string  `json:"avg_response_time"`
	RequestsPerSecond float64 `json:"requests_per_second"`
}

// ErrorsMetrics represents error rate information
type ErrorsMetrics struct {
	Rate          float64 `json:"rate"`           // Error rate as percentage
	Count         int     `json:"count"`          // Total error count
	TotalRequests int     `json:"total_requests"` // Total request count for rate calculation
}

// DiskMetrics represents disk usage information
type DiskMetrics struct {
	UsedGB         float64 `json:"used_gb"`
	TotalGB        float64 `json:"total_gb"`
	FreePercentage float64 `json:"free_percentage"`
}

// BackupMetrics represents backup status information
// Story 6.3, Task 6: Backup health monitoring integration
type BackupMetrics struct {
	IsRunning      bool    `json:"is_running"`
	LastBackupTime string  `json:"last_backup_time,omitempty"` // ISO 8601 format
	LastStatus     string  `json:"last_status"`                // success, failed, pending
	NextBackupTime string  `json:"next_backup_time,omitempty"` // ISO 8601 format
	SuccessRate    float64 `json:"success_rate"`               // Percentage (0-100)
	TotalBackups   int     `json:"total_backups"`              // Total number of backups
	TotalSize      int64   `json:"total_size_bytes"`           // Total size of all backups
}

// Alert represents a system health alert
// Story 6.2, AC9-12: Alerts for DB/Redis failures, error rate, disk space
type Alert struct {
	Severity  string    `json:"severity"` // critical, warning, info
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// AlertThresholdsConfig defines threshold values for alert generation
// Story 6.2, Task 2: Alert thresholds configuration
type AlertThresholdsConfig struct {
	ErrorRateMax    float64 `json:"error_rate_max"`    // 0.1% = 0.001
	DiskFreeMin     float64 `json:"disk_free_min"`     // 20% = 0.20
	ResponseTimeMax int     `json:"response_time_max"` // milliseconds
}

// AlertResponse represents the response for alerts endpoint
// Story 6.2, Task 3: GET /api/v1/admin/health/alerts endpoint
type AlertResponse struct {
	Alerts   []Alert `json:"alerts"`
	Total    int     `json:"total"`
	Critical int     `json:"critical"`
	Warning  int     `json:"warning"`
	Info     int     `json:"info"`
}

// MetricsRequest represents query parameters for metrics endpoint
type MetricsRequest struct {
	StartTime  *time.Time `form:"start_time"`
	EndTime    *time.Time `form:"end_time"`
	MetricType string     `form:"metric_type"` // uptime, errors, response_time, disk
	Limit      int        `form:"limit" default:"100"`
}
