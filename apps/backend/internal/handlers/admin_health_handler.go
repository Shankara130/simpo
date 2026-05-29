package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/health"
)

// AdminHealthHandler handles admin health monitoring endpoints
// Story 6.2, Task 3: Create Admin Health Monitoring API Endpoints
type AdminHealthHandler struct {
	healthService health.Service
	collector     *health.MetricsCollector
	alertService  *health.AlertService
	checkers      []health.Checker // Store checkers for metrics collection
}

// NewAdminHealthHandler creates a new admin health handler
func NewAdminHealthHandler(
	healthService health.Service,
	collector *health.MetricsCollector,
	alertService *health.AlertService,
	checkers []health.Checker,
) *AdminHealthHandler {
	return &AdminHealthHandler{
		healthService: healthService,
		collector:     collector,
		alertService:  alertService,
		checkers:      checkers,
	}
}

// GetDashboard godoc
//
//	@Summary		Get comprehensive health dashboard metrics
//	@Description	Returns system health metrics for admin dashboard (Story 6.2, AC1-7)
//	@Tags			Admin Health
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.EnhancedHealthDashboardResponse	"Health metrics retrieved successfully"
//	@Failure		401	{object}	map[string]string					"Unauthorized"
//	@Failure		403	{object}	map[string]string					"Forbidden - Admin only"
//	@Failure		500	{object}	map[string]string					"Internal server error"
//	@Router			/api/v1/admin/health/dashboard [get]
func (h *AdminHealthHandler) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current health status
	healthResponse := h.healthService.GetHealth(ctx)

	// Collect enhanced metrics
	checkers := h.getCheckers() // Will be populated when we have access to them
	metrics := h.collector.CollectMetrics(ctx, checkers, nil, 0, 0)

	// Evaluate alerts based on metrics
	checkResults := h.getCheckResults(ctx)
	alerts := h.alertService.EvaluateAllMetrics(metrics, checkResults)

	// Build enhanced response
	uptimePercentage := h.collector.GetUptimePercentage()
	uptime := healthResponse.Uptime

	// Convert health status string to dto.HealthStatus
	var status dto.HealthStatus
	statusStr := string(healthResponse.Status)
	switch statusStr {
	case "healthy":
		status = dto.StatusHealthy
	case "degraded":
		status = dto.StatusDegraded
	case "unhealthy":
		status = dto.StatusUnhealthy
	default:
		status = dto.StatusHealthy
	}

	response := dto.EnhancedHealthDashboardResponse{
		Status:           status,
		UptimePercentage: uptimePercentage,
		Uptime:           uptime,
		Version:          healthResponse.Version,
		Timestamp:        healthResponse.Timestamp,
		Metrics:          metrics,
		Alerts:           alerts,
		Environment:      healthResponse.Environment,
	}

	c.JSON(http.StatusOK, response)
}

// GetAlerts godoc
//
//	@Summary		Get active health alerts
//	@Description	Returns current health alerts grouped by severity (Story 6.2, Task 3)
//	@Tags			Admin Health
//	@Accept			json
//	@Produce		json
//	@Param			severity	query	string	false	"none"	"Filter by severity: critical, warning, info"
//	@Param			limit		query	int		false	"100"	"Maximum number of alerts to return"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.AlertResponse	"Alerts retrieved successfully"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		403	{object}	map[string]string	"Forbidden - Admin only"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Router			/api/v1/admin/health/alerts [get]
func (h *AdminHealthHandler) GetAlerts(c *gin.Context) {
	// Parse query parameters
	severity := c.Query("severity")
	limit := 100 // default limit
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := parseLimit(limitParam, 100, 1000); err == nil {
			limit = parsedLimit
		} else {
			slog.Warn("Invalid limit parameter", "value", limitParam, "error", err)
		}
	}

	var alerts []dto.Alert

	// Filter by severity if specified
	if severity != "" && severity != "none" {
		alerts = h.alertService.GetAlertsBySeverity(severity)
	} else {
		alerts = h.alertService.GetRecentAlerts()
	}

	// Apply limit
	if len(alerts) > limit {
		alerts = alerts[:limit]
	}

	// Get counts
	total, critical, warning, info := h.alertService.GetAlertCount()

	response := dto.AlertResponse{
		Alerts:   alerts,
		Total:    total,
		Critical: critical,
		Warning:  warning,
		Info:     info,
	}

	c.JSON(http.StatusOK, response)
}

// GetMetrics godoc
//
//	@Summary		Get historical health metrics
//	@Description	Returns historical health metrics data (Story 6.2, Task 3 - optional)
//	@Tags			Admin Health
//	@Accept			json
//	@Produce		json
//	@Param			start_time	query	string	false	"none"	"Start time (RFC3339 format)"
//	@Param			end_time	query	string	false	"none"	"End time (RFC3339 format)"
//	@Param			metric_type	query	string	false	"none"	"Metric type: uptime, errors, response_time, disk"
//	@Param			limit		query	int		false	"100"	"Maximum number of data points to return"
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Historical metrics retrieved"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Failure		403	{object}	map[string]string		"Forbidden - Admin only"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/api/v1/admin/health/metrics [get]
func (h *AdminHealthHandler) GetMetrics(c *gin.Context) {
	// Parse query parameters
	var req dto.MetricsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For now, return empty historical data
	// In production, this would query a metrics storage system
	c.JSON(http.StatusOK, gin.H{
		"message": "Historical metrics not yet implemented",
		"request": req,
	})
}

// getCheckers returns the list of health checkers
func (h *AdminHealthHandler) getCheckers() []health.Checker {
	return h.checkers
}

// getCheckResults returns the current check results
// This is a placeholder - in production, this would run the actual checks
func (h *AdminHealthHandler) getCheckResults(ctx context.Context) map[string]health.CheckResult {
	// For now, run a fresh health check to get current results
	healthResponse := h.healthService.GetHealth(ctx)

	// Convert health response to check results format
	checkResults := make(map[string]health.CheckResult)

	// Extract from health response checks
	for name, result := range healthResponse.Checks {
		checkResults[name] = result
	}

	return checkResults
}

// parseLimit parses and validates the limit parameter
func parseLimit(value string, min, max int) (int, error) {
	var limit int
	if _, err := fmt.Sscanf(value, "%d", &limit); err != nil {
		return 0, err
	}

	if limit < min {
		limit = min
	}
	if limit > max {
		limit = max
	}

	return limit, nil
}
