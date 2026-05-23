package handlers

import (
	"context"
	"net/http"
	"strconv"
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
}

// NewReportHandler creates a new report handler instance
// Story 5.1, Task 4.1: Constructor with dependency injection
func NewReportHandler(reportService services.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
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
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/internal-error",
			Title:    "Internal Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Invalid user role format in context",
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
		branchID = &bid
	}

	// Story 5.1, Task 4.4: Create request DTO and call service
	req := &dto.DailySalesRequest{
		Date:     date,
		BranchID: branchID,
	}

	// Story 5.1, Task 4.6: Call ReportService with context timeout
	// Code review fix: LOW-002 - Use centralized error handler
	summary, err := h.reportService.GenerateDailySalesSummary(context.Background(), req)
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
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/internal-error",
			Title:    "Internal Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Invalid user role format in context",
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
	_, err := time.Parse("2006-01-02", startDate)
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

	_, err = time.Parse("2006-01-02", endDate)
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

	// Story 5.2, Task 4.2: Extract optional query parameters
	breakdownBy := c.Query("breakdown_by")
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
	summary, err := h.reportService.GenerateProfitLossSummary(context.Background(), req)
	if err != nil {
		h.handleReportError(c, err)
		return
	}

	// Story 5.2, Task 4.6: Return 200 OK with report data
	c.JSON(http.StatusOK, summary)
}
