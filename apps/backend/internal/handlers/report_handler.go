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
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// ReportHandler handles report-related HTTP requests
// Story 5.1, Task 4: Handler for daily sales summary reports
type ReportHandler struct {
	reportService services.ReportService
	exportService services.ExportService
	auditService  services.AuditService
}

// NewReportHandler creates a new report handler instance
// Story 5.1, Task 4.1: Constructor with dependency injection
// Story 5.3, Task 4.3: Inject ExportService for export functionality
// Story 5.3, Task 4.7: Inject AuditService for export event logging
func NewReportHandler(reportService services.ReportService, exportService services.ExportService, auditService services.AuditService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		exportService: exportService,
		auditService:  auditService,
	}
}

// handleReportError is a centralized error handler for report responses
// Code review fix: LOW-002 - Centralized error handling to reduce duplication
func (h *ReportHandler) handleReportError(c *gin.Context, err error) {
	// Handle service errors
	if invalidInputErr, ok := err.(*services.InvalidInputError); ok {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   invalidInputErr.Message,
			Instance: c.Request.URL.Path,
		})
		return
	}

	if serviceErr, ok := err.(*services.ServiceError); ok {
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/service-error",
			Title:    "Report Generation Failed",
			Status:   http.StatusInternalServerError,
			Detail:   serviceErr.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Generic error response
	c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
		Type:     "https://api.simpo.com/errors/internal-error",
		Title:    "Internal Error",
		Status:   http.StatusInternalServerError,
		Detail:   "Failed to generate report. Please try again later.",
		Instance: c.Request.URL.Path,
	})
}

// hasReportAccess checks if the user role has permission to access financial reports
// Code review fix: CRITICAL-008 - Extracted RBAC validation to reduce duplication
func hasReportAccess(userRole string) bool {
	return userRole == user.RoleOwner || userRole == user.RoleAdmin || userRole == user.RoleSystemAdmin
}

// validateAndParseDate validates and parses a date string
// Code review fix: CRITICAL-009 - Extracted date validation to reduce duplication
// Code review fix: CRITICAL-010 - Added input sanitization for security
func validateAndParseDate(dateStr string) (time.Time, error) {
	// Sanitize input - trim whitespace and remove any null characters
	// Code review fix: CRITICAL-010 - Input sanitization to prevent injection
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

	// Validate date is not too far in the past (limit to 1 year)
	oneYearAgo := now.AddDate(-1, 0, 0)
	if parsedDate.Before(oneYearAgo) {
		return time.Time{}, &services.InvalidInputError{
			Field:   "date",
			Message: "Date cannot be more than 1 year in the past.",
		}
	}

	return parsedDate, nil
}

// validateDateRange validates that start_date <= end_date and range is not excessive
// Code review fix: HIGH-002 - Added date range validation
// Code review fix: HIGH-008 - Added maximum duration validation (1 year)
func validateDateRange(startDate, endDate time.Time) error {
	if startDate.After(endDate) {
		return &services.InvalidInputError{
			Field:   "date_range",
			Message: "start_date must be before or equal to end_date.",
		}
	}
	// Validate that date range is not excessive (prevent DoS)
	duration := endDate.Sub(startDate)
	maxDuration := 365 * 24 * time.Hour // 1 year
	if duration > maxDuration {
		return &services.InvalidInputError{
			Field:   "date_range",
			Message: "Date range cannot exceed 1 year. Please select a shorter date range.",
		}
	}
	return nil
}

// validateBreakdownBy validates that breakdown_by parameter is one of the allowed values
// Code review fix: CRITICAL-002 (Round 4) - Add whitelist validation for breakdown_by parameter
func validateBreakdownBy(breakdownBy string) error {
	if breakdownBy == "" {
		return nil // Empty is allowed, will use default
	}
	allowedValues := map[string]bool{
		"category":       true,
		"branch":         true,
		"payment_method": true,
	}
	if !allowedValues[breakdownBy] {
		return &services.InvalidInputError{
			Field:   "breakdown_by",
			Message: "breakdown_by must be one of: category, branch, payment_method",
		}
	}
	return nil
}

// GetDailySalesReport handles GET /api/v1/reports/daily
// Story 5.1, Task 4.1-4.6: Generate daily sales summary report
func (h *ReportHandler) GetDailySalesReport(c *gin.Context) {
	// Story 5.1, Task 4.5: Extract user role for RBAC validation
	userRoleValue, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Authentication required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	userRole, ok := userRoleValue.(string)
	if !ok {
		// Code review fix: CRITICAL-009 - Return 401 Unauthorized instead of 500 for security
		c.JSON(http.StatusUnauthorized, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Authentication required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Story 5.1, Task 4.5: RBAC validation - Only Owner and Admin can access financial reports
	// Code review fix: Support both legacy RoleAdmin and new RoleSystemAdmin for backward compatibility
	if userRole != user.RoleOwner && userRole != user.RoleAdmin && userRole != user.RoleSystemAdmin {
		c.JSON(http.StatusForbidden, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/forbidden",
			Title:    "Access Denied",
			Status:   http.StatusForbidden,
			Detail:   "You do not have permission to access financial reports. Only Owners and Administrators can view sales reports.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Story 5.1, Task 4.2: Extract and validate date parameter
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "Date parameter is required. Use YYYY-MM-DD format.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Story 5.1, Task 4.3: Validate date format (YYYY-MM-DD)
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid date format. Use YYYY-MM-DD format (e.g., 2026-05-23).",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Story 5.1, Task 4.2: Extract optional branch_id parameter
	var branchID *uint
	branchIDStr := c.Query("branch_id")
	if branchIDStr != "" {
		// Parse branch_id from string to uint
		branchIDUint, err := strconv.ParseUint(branchIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:     "https://api.simpo.com/errors/validation-failed",
				Title:    "Validation Failed",
				Status:   http.StatusBadRequest,
				Detail:   "Invalid branch_id format. Must be a positive integer.",
				Instance: c.Request.URL.Path,
			})
			return
		}
		bid := uint(branchIDUint)
		// Code review fix: HIGH-006 - Validate range to prevent integer overflow
		branchID = &bid
	}

	// Story 5.1, Task 4.4: Create request DTO and call service
	req := &dto.DailySalesRequest{
		Date:     date,
		BranchID: branchID,
	}

	// Story 5.1, Task 4.6: Call ReportService with context timeout
	// Code review fix: LOW-002 - Use centralized error handler
	// Code review fix: CRITICAL-001 - Add timeout to prevent hanging operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	summary, err := h.reportService.GenerateDailySalesSummary(ctx, req)
	if err != nil {
		h.handleReportError(c, err)
		return
	}

	// Story 5.1, Task 4.6: Return 200 OK with report data
	c.JSON(http.StatusOK, summary)
}

// GetProfitLossReport handles GET /api/v1/reports/profit-loss
// Story 5.2, Task 4.1-4.6: Generate profit/loss summary report
func (h *ReportHandler) GetProfitLossReport(c *gin.Context) {
	// Story 5.2, Task 4.5: Extract user role for RBAC validation
	userRoleValue, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Authentication required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	userRole, ok := userRoleValue.(string)
	if !ok {
		// Code review fix: CRITICAL-009 - Return 401 Unauthorized instead of 500 for security
		c.JSON(http.StatusUnauthorized, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Authentication required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Story 5.2, Task 4.5: RBAC validation - Only Owner and Admin can access financial reports
	// Code review fix: Support both legacy RoleAdmin and new RoleSystemAdmin for backward compatibility
	if userRole != user.RoleOwner && userRole != user.RoleAdmin && userRole != user.RoleSystemAdmin {
		c.JSON(http.StatusForbidden, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/forbidden",
			Title:    "Access Denied",
			Status:   http.StatusForbidden,
			Detail:   "You do not have permission to access financial reports. Only Owners and Administrators can view profit/loss reports.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Story 5.2, Task 4.2: Extract and validate required query parameters
	startDate := c.Query("start_date")
	if startDate == "" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "start_date parameter is required. Use YYYY-MM-DD format.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	endDate := c.Query("end_date")
	if endDate == "" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "end_date parameter is required. Use YYYY-MM-DD format.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Story 5.2, Task 4.3: Validate date format (YYYY-MM-DD)
	parsedStartDate, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid start_date format. Use YYYY-MM-DD format (e.g., 2026-05-01).",
			Instance: c.Request.URL.Path,
		})
		return
	}

	parsedEndDate, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid end_date format. Use YYYY-MM-DD format (e.g., 2026-05-23).",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Code review fix: HIGH-002 - Validate date range (start_date <= end_date)
	if err := validateDateRange(parsedStartDate, parsedEndDate); err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}
	// Story 5.2, Task 4.2: Extract optional query parameters
	breakdownBy := c.Query("breakdown_by")
	// Code review fix: CRITICAL-002 (Round 4) - Validate breakdown_by parameter
	if err := validateBreakdownBy(breakdownBy); err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}
	var branchID *uint
	branchIDStr := c.Query("branch_id")
	if branchIDStr != "" {
		// Parse branch_id from string to uint
		branchIDUint, err := strconv.ParseUint(branchIDStr, 10, 32)
		// Code review fix: HIGH-006 - Validate range to prevent integer overflow
		if err != nil {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:     "https://api.simpo.com/errors/validation-failed",
				Title:    "Validation Failed",
				Status:   http.StatusBadRequest,
				Detail:   "Invalid branch_id format. Must be a positive integer.",
				Instance: c.Request.URL.Path,
			})
			return
		}
		bid := uint(branchIDUint)
		branchID = &bid
	}

	// Story 5.2, Task 4.4: Create request DTO and call service
	req := &dto.ProfitLossRequest{
		StartDate:   startDate,
		EndDate:     endDate,
		BreakdownBy: breakdownBy,
		BranchID:    branchID,
	}

	// Story 5.2, Task 4.6: Call ReportService with context timeout
	// Code review fix: LOW-002 - Use centralized error handler
	// Code review fix: CRITICAL-001 - Add timeout to prevent hanging operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	summary, err := h.reportService.GenerateProfitLossSummary(ctx, req)
	if err != nil {
		h.handleReportError(c, err)
		return
	}

	// Story 5.2, Task 4.6: Return 200 OK with report data
	c.JSON(http.StatusOK, summary)
}

// ExportDailySalesReport handles GET /api/v1/reports/daily/export
// Story 5.3, Task 4.1-4.8: Export daily sales report as PDF or Excel
func (h *ReportHandler) ExportDailySalesReport(c *gin.Context) {
	// Note: Rate limiting should be applied at middleware level for all expensive operations (deferred - global concern)
	// Extract user role for RBAC validation
	userRoleValue, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Authentication required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	userRole, ok := userRoleValue.(string)
	if !ok {
		// Code review fix: CRITICAL-009 - Return 401 Unauthorized instead of 500 for security
		c.JSON(http.StatusUnauthorized, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Authentication required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Story 5.3, Task 4.6: RBAC validation - Only Owner and Admin can export

	// Code review fix: CRITICAL-008 - Use helper function to reduce duplication
	if !hasReportAccess(userRole) {
		c.JSON(http.StatusForbidden, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/forbidden",
			Title:    "Access Denied",
			Status:   http.StatusForbidden,
			Detail:   "You do not have permission to export financial reports. Only Owners and Administrators can export reports.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Story 5.3, Task 4.1: Extract and validate date parameter
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "Date parameter is required. Use YYYY-MM-DD format.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Validate date format
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid date format. Use YYYY-MM-DD format (e.g., 2026-05-23).",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Story 5.3, Task 4.1: Extract and validate format parameter
	format := c.Query("format")
	if format == "" {
		format = "pdf" // Default to PDF
	}

	if format != "pdf" && format != "xlsx" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Invalid Format Parameter",
			Status:   http.StatusBadRequest,
			Detail:   "Format must be 'pdf' or 'xlsx'.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Extract optional branch_id parameter
	var branchID *uint
	branchIDStr := c.Query("branch_id")
	if branchIDStr != "" {
		branchIDUint, err := strconv.ParseUint(branchIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:     "https://api.simpo.com/errors/validation-failed",
				Title:    "Validation Failed",
				Status:   http.StatusBadRequest,
				Detail:   "Invalid branch_id format. Must be a positive integer.",
				Instance: c.Request.URL.Path,
			})
			return
		}
		bid := uint(branchIDUint)
		branchID = &bid
	}

	// Extract user ID for audit trail
	// Code review fix: HIGH-014 - Use pointer to distinguish "not authenticated" from "user ID 0"
	userIDValue, exists := c.Get("user_id")
	var userID *uint
	if exists {
		if uid, ok := userIDValue.(uint); ok {
			userID = &uid
		}
	}

	// Extract username for audit trail
	usernameValue, exists := c.Get("username")
	var username string
	if exists {
		if uname, ok := usernameValue.(string); ok {
			username = uname
		}
	}

	// Create export request
	// Code review fix: HIGH-014 - Unwrap userID pointer for ExportRequest
	exportUserID := uint(0)
	if userID != nil {
		exportUserID = *userID
	}
	req := &dto.ExportRequest{
		ReportType: dto.ReportTypeDailySales,
		Format:     format,
		Date:       &date,
		BranchID:   branchID,
		UserID:     exportUserID,
		UserRole:   userRole,
	}

	// Call export service based on format
	var response *dto.ExportResponse
	if format == "pdf" {
		response, err = h.exportService.ExportDailySalesToPDF(c.Request.Context(), req)
	} else {
		response, err = h.exportService.ExportDailySalesToExcel(c.Request.Context(), req)
	}

	if err != nil {
		h.handleReportError(c, err)
		return
	}

	// Code review fix: CRITICAL-002 - Audit trail logging for regulatory compliance
	// Log the export event for Badan POM requirements
	dateRange := date
	// Code review fix: HIGH-014 - Handle nil userID pointer properly
	auditUserID := uint(0)
	if userID != nil {
		auditUserID = *userID
	}
	if err := h.auditService.LogReportExport(context.Background(), auditUserID, username, "daily_sales", format, dateRange, "success", c.ClientIP()); err != nil {
		// CRITICAL FIX: Log audit failures for regulatory compliance
		// Code review fix: MEDIUM-018 (Round 4) - Use proper logging instead of fmt.Printf
		slog.Error("Audit log failed for daily_sales export", "user_id", userID, "error", err)
		// Continue with export to avoid blocking user operations
	}

	// Story 5.3, Task 4.4-4.5: Set headers and return file

	// Story 5.3, Task 4.4-4.5: Set headers and return file
	contentType := "application/pdf"
	if format == "xlsx" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", response.FileName))
	// Code review fix: LOW-001 (Round 5) - Add security headers
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-Download-Options", "noopen")
	c.Data(http.StatusOK, contentType, response.FileData)
}

// ExportProfitLossReport handles GET /api/v1/reports/profit-loss/export
// Story 5.3, Task 4.1-4.8: Export profit/loss report as PDF or Excel
func (h *ReportHandler) ExportProfitLossReport(c *gin.Context) {
	// Extract user role for RBAC validation
	userRoleValue, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Authentication required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	userRole, ok := userRoleValue.(string)
	if !ok {
		// Code review fix: CRITICAL-009 - Return 401 Unauthorized instead of 500 for security
		c.JSON(http.StatusUnauthorized, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Authentication required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Story 5.3, Task 4.6: RBAC validation - Only Owner and Admin can export

	// Code review fix: CRITICAL-008 - Use helper function to reduce duplication
	if !hasReportAccess(userRole) {
		c.JSON(http.StatusForbidden, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/forbidden",
			Title:    "Access Denied",
			Status:   http.StatusForbidden,
			Detail:   "You do not have permission to export financial reports. Only Owners and Administrators can export reports.",
			Instance: c.Request.URL.Path,
		})
		return
	}
	// Story 5.3, Task 4.1: Extract and validate date parameters
	startDate := c.Query("start_date")

	if startDate == "" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "start_date parameter is required. Use YYYY-MM-DD format.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	endDate := c.Query("end_date")
	if endDate == "" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "end_date parameter is required. Use YYYY-MM-DD format.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Validate date formats
	parsedStartDate, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid start_date format. Use YYYY-MM-DD format (e.g., 2026-05-23).",
			Instance: c.Request.URL.Path,
		})
		return
	}

	parsedEndDate, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid end_date format. Use YYYY-MM-DD format (e.g., 2026-05-23).",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Code review fix: HIGH-002 - Validate date range (start_date <= end_date)
	if err := validateDateRange(parsedStartDate, parsedEndDate); err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}
	format := c.Query("format")

	if format == "" {
		format = "pdf" // Default to PDF
	}

	if format != "pdf" && format != "xlsx" {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Invalid Format Parameter",
			Status:   http.StatusBadRequest,
			Detail:   "Format must be 'pdf' or 'xlsx'.",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Extract optional query parameters
	breakdownBy := c.Query("breakdown_by")
	// Code review fix: CRITICAL-002 (Round 4) - Validate breakdown_by parameter
	if err := validateBreakdownBy(breakdownBy); err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/validation-failed",
			Title:    "Validation Failed",
			Status:   http.StatusBadRequest,
			Detail:   err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}
	var branchID *uint
	branchIDStr := c.Query("branch_id")
	if branchIDStr != "" {
		branchIDUint, err := strconv.ParseUint(branchIDStr, 10, 32)
		// Code review fix: HIGH-006 - Validate range to prevent integer overflow

		if err != nil {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:     "https://api.simpo.com/errors/validation-failed",
				Title:    "Validation Failed",
				Status:   http.StatusBadRequest,
				Detail:   "Invalid branch_id format. Must be a positive integer.",
				Instance: c.Request.URL.Path,
			})
			return
		}
		bid := uint(branchIDUint)
		branchID = &bid
	}

	// Extract user ID for audit trail
	// Code review fix: HIGH-014 - Use pointer to distinguish "not authenticated" from "user ID 0"
	userIDValue, exists := c.Get("user_id")
	var userID *uint
	if exists {
		if uid, ok := userIDValue.(uint); ok {
			userID = &uid
		}
	}

	// Extract username for audit trail
	usernameValue, exists := c.Get("username")
	var username string
	if exists {
		if uname, ok := usernameValue.(string); ok {
			username = uname
		}
	}

	// Create export request
	// Code review fix: HIGH-014 - Unwrap userID pointer for ExportRequest
	exportUserID := uint(0)
	if userID != nil {
		exportUserID = *userID
	}
	req := &dto.ExportRequest{
		ReportType:  dto.ReportTypeProfitLoss,
		Format:      format,
		StartDate:   &startDate,
		EndDate:     &endDate,
		BreakdownBy: &breakdownBy,
		BranchID:    branchID,
		UserID:      exportUserID,
		UserRole:    userRole,
	}

	// Code review fix: HIGH-013 - Handle empty breakdown_by correctly
	// If breakdownBy is empty string, set BreakdownBy to nil instead of pointer to empty string
	if breakdownBy == "" {
		req.BreakdownBy = nil
	}

	// Call export service based on format
	var response *dto.ExportResponse
	if format == "pdf" {
		response, err = h.exportService.ExportProfitLossToPDF(c.Request.Context(), req)
	} else {
		response, err = h.exportService.ExportProfitLossToExcel(c.Request.Context(), req)
	}

	if err != nil {
		h.handleReportError(c, err)
		return
	}

	// Code review fix: CRITICAL-002 - Audit trail logging for regulatory compliance
	// Log the export event for Badan POM requirements
	dateRange := fmt.Sprintf("%s to %s", startDate, endDate)
	// Code review fix: HIGH-014 - Handle nil userID pointer properly
	auditUserID := uint(0)
	if userID != nil {
		auditUserID = *userID
	}
	if err := h.auditService.LogReportExport(context.Background(), auditUserID, username, "profit_loss", format, dateRange, "success", c.ClientIP()); err != nil {
		// CRITICAL FIX: Log audit failures for regulatory compliance
		// Code review fix: MEDIUM-018 (Round 4) - Use proper logging instead of fmt.Printf
		slog.Error("Audit log failed for profit_loss export", "user_id", userID, "error", err)
		// Continue with export to avoid blocking user operations
	}

	// Story 5.3, Task 4.4-4.5: Set headers and return file
	contentType := "application/pdf"
	if format == "xlsx" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", response.FileName))
	// Code review fix: LOW-001 (Round 5) - Add security headers
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-Download-Options", "noopen")
	c.Data(http.StatusOK, contentType, response.FileData)
}
