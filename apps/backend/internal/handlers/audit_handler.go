package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// MED-003: Constants for audit handler configuration
const (
	// Maximum date range for audit queries (prevents DoS)
	auditMaxDateRangeDays = 365
	// Default pagination limit
	auditDefaultLimit = 20
	// Maximum pagination limit
	auditMaxLimit = 100
	// Maximum export limit (records)
	auditMaxExportLimit = 10000
	// Export timeout (prevents hanging on large exports)
	auditExportTimeout = 5 * time.Minute
	// Query timeout (prevents hanging on slow queries)
	auditQueryTimeout = 10 * time.Second
	// Cleanup timeout (prevents hanging on large cleanup operations)
	auditCleanupTimeout = 30 * time.Second
)

// AuditHandler handles audit log-related HTTP requests
// Story 5.4, Task 5: Handler for audit log query and export APIs
type AuditHandler struct {
	auditRepo    repositories.AuditRepository
	auditService services.AuditService
}

// NewAuditHandler creates a new audit handler instance
// Story 5.4, Task 5.1: Constructor with dependency injection
func NewAuditHandler(auditRepo repositories.AuditRepository, auditService services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditRepo:    auditRepo,
		auditService: auditService,
	}
}

// hasAuditAccess checks if the user role has permission to access audit logs
// Story 5.4, Task 5.3: RBAC validation (Admin, Owner, SystemAdmin only)
func hasAuditAccess(userRole string) bool {
	return userRole == user.RoleAdmin || userRole == user.RoleOwner || userRole == user.RoleSystemAdmin
}

// hasCleanupAccess checks if the user has permission to perform retention cleanup
// Story 5.4, Task 7.2: RBAC validation - SystemAdmin only
func hasCleanupAccess(userRole string) bool {
	return userRole == user.RoleSystemAdmin
}

// validateAuditDateRange validates that date range is not excessive for audit queries
// Story 5.4, Task 6.4: Add date range validation for exports (max 1 year)
func validateAuditDateRange(startDate, endDate time.Time) error {
	if startDate.After(endDate) {
		return &services.InvalidInputError{
			Field:   "date_range",
			Message: "start_date must be before or equal to end_date.",
		}
	}

	// Validate that date range is not excessive (prevent DoS)
	duration := endDate.Sub(startDate)
	maxDuration := time.Duration(auditMaxDateRangeDays) * 24 * time.Hour
	if duration > maxDuration {
		return &services.InvalidInputError{
			Field:   "date_range",
			Message: "Date range cannot exceed 1 year. Please select a shorter date range.",
		}
	}

	return nil
}

// validateAuditDate validates and parses a date string for audit queries
func validateAuditDate(dateStr string) (time.Time, error) {
	// Sanitize input - trim whitespace and remove any null characters
	sanitized := strings.TrimSpace(strings.ReplaceAll(dateStr, "\x00", ""))

	// Parse date
	parsedDate, err := time.Parse("2006-01-02", sanitized)
	if err != nil {
		return time.Time{}, &services.InvalidInputError{
			Field:   "date",
			Message: "Invalid date format. Use YYYY-MM-DD format.",
		}
	}

	// Validate date is not in the future
	now := time.Now()
	if parsedDate.After(now) {
		return time.Time{}, &services.InvalidInputError{
			Field:   "date",
			Message: "Date cannot be in the future.",
		}
	}

	return parsedDate, nil
}

// GetAuditLogs handles GET /api/v1/audit/logs
// Story 5.4, Task 5.1: Query audit logs with filters and pagination
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	// Extract user role from context (set by JWT middleware)
	userRoleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusForbidden, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/forbidden",
			Title:    "Access Denied",
			Status:   http.StatusForbidden,
			Detail:   "User role not found in request context.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	userRole, ok := userRoleValue.(string)
	if !ok || !hasAuditAccess(userRole) {
		c.JSON(http.StatusForbidden, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/forbidden",
			Title:    "Access Denied",
			Status:   http.StatusForbidden,
			Detail:   "You do not have permission to view audit logs.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Parse query parameters
	filter := &repositories.AuditLogFilter{
		Limit:  20, // Default limit
		Offset: 0,  // Default offset
	}

	// Parse user_id filter (optional)
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:     "https://api.simpo.com/errors/validation-failed",
				Title:    "Invalid user_id",
				Status:   http.StatusBadRequest,
				Detail:   "user_id must be a valid integer.",
				Instance: c.Request.URL.Path,
			})
			return
		}
		uid := uint(userID)
		filter.UserID = &uid
	}

	// Parse action filter (optional)
	if action := c.Query("action"); action != "" {
		// Validate action is a valid AuditAction
		validActions := map[models.AuditAction]bool{
			models.AuditActionLoginSuccess:           true,
			models.AuditActionLoginFailure:           true,
			models.AuditActionLogout:                 true,
			models.AuditActionPasswordReset:          true,
			models.AuditActionAuthFailure:            true,
			models.AuditActionForbiddenAccess:        true,
			models.AuditActionUserCreated:            true,
			models.AuditActionUserDeactivated:        true,
			models.AuditActionSelfRegistration:       true,
			models.AuditActionEmailVerified:          true,
			models.AuditActionWhitelistDomainAdded:   true,
			models.AuditActionWhitelistDomainUpdated: true,
			models.AuditActionWhitelistDomainDeleted: true,
			models.AuditActionStockAdjustment:        true,
			models.AuditActionBlockedSaleAttempt:     true,
			models.AuditActionExportReport:           true,
		}
		actionEnum := models.AuditAction(action)
		if !validActions[actionEnum] {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:     "https://api.simpo.com/errors/validation-failed",
				Title:    "Invalid action",
				Status:   http.StatusBadRequest,
				Detail:   "action must be a valid audit action.",
				Instance: c.Request.URL.Path,
			})
			return
		}
		actionStr := string(actionEnum)
		filter.Action = &actionStr
	}

	// Parse start_date (required)
	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Missing start_date",
			Status:   http.StatusBadRequest,
			Detail:   "start_date parameter is required (format: YYYY-MM-DD).",
			Instance: c.Request.URL.Path,
		})
		return
	}
	startDate, err := validateAuditDate(startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Invalid start_date",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}
	startDateFormatted := startDate.Format("2006-01-02")
	filter.StartDate = &startDateFormatted

	// Parse end_date (required)
	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Missing end_date",
			Status:   http.StatusBadRequest,
			Detail:   "end_date parameter is required (format: YYYY-MM-DD).",
			Instance: c.Request.URL.Path,
		})
		return
	}
	endDate, err := validateAuditDate(endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Invalid end_date",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Validate date range
	if err := validateAuditDateRange(startDate, endDate); err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Invalid Date Range",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}

	endDateFormatted := endDate.Format("2006-01-02")
	filter.EndDate = &endDateFormatted

	// Parse limit (optional)
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:     "https://api.simpo.com/errors/validation-failed",
				Title:    "Invalid limit",
				Status:   http.StatusBadRequest,
				Detail:   "limit must be a positive integer.",
				Instance: c.Request.URL.Path,
			})
			return
		}
		filter.Limit = limit
	}

	// Parse offset (optional)
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:     "https://api.simpo.com/errors/validation-failed",
				Title:    "Invalid offset",
				Status:   http.StatusBadRequest,
				Detail:   "offset must be a non-negative integer.",
				Instance: c.Request.URL.Path,
			})
			return
		}
		filter.Offset = offset
	}

	// Query audit logs from repository
	ctx, cancel := context.WithTimeout(context.Background(), auditQueryTimeout)
	defer cancel()

	auditLogs, total, err := h.auditRepo.Query(ctx, filter)
	if err != nil {
		slog.Error("Failed to query audit logs", "error", err)
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/internal-error",
			Title:    "Query Failed",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to retrieve audit logs. Please try again later.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Calculate pagination metadata
	var totalPages int
	if filter.Limit > 0 {
		totalPages = int(total) / filter.Limit
		if int(total)%filter.Limit > 0 {
			totalPages++
		}
	} else {
		totalPages = 1 // Default to 1 page if limit is invalid
	}

	// Build response with pagination
	type AuditLogResponse struct {
		ID        uint   `json:"id"`
		UserID    uint   `json:"user_id"`
		Username  string `json:"username"`
		Action    string `json:"action"`
		IPAddress string `json:"ip_address,omitempty"`
		Outcome   string `json:"outcome"`
		Reason    string `json:"reason,omitempty"`
		Timestamp string `json:"timestamp"`
	}

	data := make([]AuditLogResponse, len(auditLogs))
	for i, log := range auditLogs {
		data[i] = AuditLogResponse{
			ID:        log.ID,
			UserID:    log.UserID,
			Username:  log.Username,
			Action:    string(log.Action),
			IPAddress: log.IPAddress,
			Outcome:   log.Outcome,
			Reason:    log.Reason,
			Timestamp: log.Timestamp.Format(time.RFC3339),
		}
	}

	response := gin.H{
		"data": data,
		"pagination": gin.H{
			"total":       total,
			"limit":       filter.Limit,
			"offset":      filter.Offset,
			"total_pages": totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetAuditLogsExport handles GET /api/v1/audit/logs/export
// Story 5.4, Task 6.1: Export audit logs in CSV or JSON format
func (h *AuditHandler) GetAuditLogsExport(c *gin.Context) {
	// Extract user role from context (set by JWT middleware)
	userRoleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusForbidden, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/forbidden",
			Title:    "Access Denied",
			Status:   http.StatusForbidden,
			Detail:   "User role not found in request context.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	userRole, ok := userRoleValue.(string)
	if !ok || !hasAuditAccess(userRole) {
		c.JSON(http.StatusForbidden, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/forbidden",
			Title:    "Access Denied",
			Status:   http.StatusForbidden,
			Detail:   "You do not have permission to export audit logs.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Parse query parameters (same validation as GetAuditLogs)
	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Missing start_date",
			Status:   http.StatusBadRequest,
			Detail:   "start_date parameter is required (format: YYYY-MM-DD).",
			Instance: c.Request.URL.Path,
		})
		return
	}
	startDate, err := validateAuditDate(startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Invalid start_date",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}
	startDateFormatted := startDate.Format("2006-01-02")

	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Missing end_date",
			Status:   http.StatusBadRequest,
			Detail:   "end_date parameter is required (format: YYYY-MM-DD).",
			Instance: c.Request.URL.Path,
		})
		return
	}
	endDate, err := validateAuditDate(endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Invalid end_date",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Validate date range (max 1 year for export)
	if err := validateAuditDateRange(startDate, endDate); err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Invalid Date Range",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}
	endDateFormatted := endDate.Format("2006-01-02")

	// Parse format (default: csv)
	format := c.Query("format")
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "json" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Invalid format",
			Status:   http.StatusBadRequest,
			Detail:   "format must be 'csv' or 'json'.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Create filter for export (no pagination, max 10000 records)
	filter := &repositories.AuditLogFilter{
		StartDate: &startDateFormatted,
		EndDate:   &endDateFormatted,
		Limit:     10000, // Export limit
		Offset:    0,
	}

	// Set content type and filename based on format
	var contentType string
	var filename string
	if format == "csv" {
		contentType = "text/csv"
		filename = fmt.Sprintf("AuditLogs_%s_to_%s.csv", startDateFormatted, endDateFormatted)
	} else {
		contentType = "application/json"
		filename = fmt.Sprintf("AuditLogs_%s_to_%s.json", startDateFormatted, endDateFormatted)
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// HIGH-007: Add timeout for export operations (5 minutes max for large exports)
	ctx, cancel := context.WithTimeout(context.Background(), auditExportTimeout)
	defer cancel()

	// Export directly to response writer
	if err := h.auditRepo.Export(ctx, filter, format, c.Writer); err != nil {
		slog.Error("Failed to export audit logs", "error", err)
		// Note: Can't send JSON response after headers are written
		// The error will be visible in logs, and client will receive partial/empty file
		return
	}

	// Log export to audit trail
	// Story 5.4, Task 6.8: Log audit log export events
	userIDValue, _ := c.Get("user_id")
	usernameValue, _ := c.Get("username")
	userID, _ := userIDValue.(uint)
	username, _ := usernameValue.(string)

	_ = h.auditService.LogReportExport(context.Background(), userID, username, "audit_logs", format, fmt.Sprintf("%s_to_%s", startDateFormatted, endDateFormatted), "success", c.ClientIP())
}

// CleanupAuditLogs handles POST /api/v1/audit/cleanup
// Story 5.4, Task 7.1: Manual trigger for 5-year retention cleanup
func (h *AuditHandler) CleanupAuditLogs(c *gin.Context) {
	// Extract user role from context (set by JWT middleware)
	userRoleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusForbidden, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/forbidden",
			Title:    "Access Denied",
			Status:   http.StatusForbidden,
			Detail:   "User role not found in request context.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	userRole, ok := userRoleValue.(string)
	if !ok || !hasCleanupAccess(userRole) {
		c.JSON(http.StatusForbidden, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/forbidden",
			Title:    "Access Denied",
			Status:   http.StatusForbidden,
			Detail:   "Only SystemAdmin can perform retention cleanup.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Require confirmation parameter
	// Story 5.4, Task 7.4: Safety confirmation parameter required
	confirm := c.Query("confirm")
	if confirm != "true" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Confirmation Required",
			Status:   http.StatusBadRequest,
			Detail:   "This operation will permanently delete audit logs older than 5 years. Add ?confirm=true to proceed.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Calculate cutoff date (5 years ago)
	cutoffDate := time.Now().AddDate(-5, 0, 0)
	cutoffDateFormatted := cutoffDate.Format("2006-01-02")

	// Log cleanup operation to audit trail BEFORE execution
	// Story 5.4, Task 7.6: Log cleanup operation to audit trail before execution
	userIDValue, _ := c.Get("user_id")
	usernameValue, _ := c.Get("username")
	userID, _ := userIDValue.(uint)
	username, _ := usernameValue.(string)

	reason := fmt.Sprintf("Retention cleanup: Deleting audit logs older than %s (5-year retention policy)", cutoffDateFormatted)
	entry := services.AuditLogEntry{
		UserID:    &userID,
		Username:  username,
		Action:    models.AuditAction("RETENTION_CLEANUP"),
		IPAddress: c.ClientIP(),
		Outcome:   "pending",
		Reason:    reason,
		Timestamp: time.Now(),
	}
	// HIGH-002: Use LogAuthorizationFailure for cleanup logging (more appropriate than LogLoginAttempt)
	// TODO: Add dedicated LogRetentionCleanup method to AuditService interface
	_ = h.auditService.LogAuthorizationFailure(context.Background(), entry)

	// Perform cleanup
	ctx, cancel := context.WithTimeout(context.Background(), auditCleanupTimeout)
	defer cancel()

	deletedCount, err := h.auditRepo.RetentionCleanup(ctx)
	if err != nil {
		slog.Error("Failed to perform retention cleanup", "error", err)
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/internal-error",
			Title:    "Cleanup Failed",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to perform retention cleanup. Please try again later.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Return summary
	response := gin.H{
		"message":          fmt.Sprintf("Successfully deleted %d audit log entries older than %s", deletedCount, cutoffDateFormatted),
		"deleted_count":    deletedCount,
		"cutoff_date":      cutoffDateFormatted,
		"retention_policy": "5 years",
		"performed_by":     username,
		"performed_at":     time.Now().Format(time.RFC3339),
	}

	slog.Info("Audit log retention cleanup completed",
		"deleted_count", deletedCount,
		"cutoff_date", cutoffDateFormatted,
		"performed_by", username,
	)

	c.JSON(http.StatusOK, response)
}
