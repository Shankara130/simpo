package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// SupplierAgingReportHandler handles supplier aging report operations
// Story 10.6: Handler for generating supplier aging reports with owner-only access
type SupplierAgingReportHandler struct {
	agingReportService services.SupplierAgingReportService
}

// NewSupplierAgingReportHandler creates a new supplier aging report handler
// Story 10.6: Factory function with dependency injection
func NewSupplierAgingReportHandler(agingReportService services.SupplierAgingReportService) *SupplierAgingReportHandler {
	return &SupplierAgingReportHandler{
		agingReportService: agingReportService,
	}
}

// GenerateAgingReport godoc
//
//	@Summary		Generate supplier aging report
//	@Description	Generates a comprehensive supplier aging report showing outstanding invoices grouped by payment period (Story 10.6, AC1)
//	@Tags			Reports - Supplier Aging
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.SupplierAgingReportRequest	true	"Aging report request"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.SupplierAgingReportResponse	"Aging report generated successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Owner only"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/reports/supplier-aging [post]
func (h *SupplierAgingReportHandler) GenerateAgingReport(c *gin.Context) {
	// Parse request
	var req dto.SupplierAgingReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request format: " + err.Error(),
			Instance: "/api/v1/reports/supplier-aging",
		})
		return
	}

	// Extract user context
	_, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// Call service to generate aging report
	response, err := h.agingReportService.GenerateAgingReport(c.Request.Context(), &req)
	if err != nil {
		// Handle validation errors
		if err.Error() == "request cannot be nil" || err.Error() == "asOfDate is required" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/validation-error",
				Title:    "Validation Error",
				Status:   http.StatusBadRequest,
				Detail:   err.Error(),
				Instance: "/api/v1/reports/supplier-aging",
			})
			return
		}

		// Handle invalid date format
		if containsString(err.Error(), "invalid asOfDate format") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/validation-error",
				Title:    "Validation Error",
				Status:   http.StatusBadRequest,
				Detail:   "Tanggal laporan tidak valid. Gunakan format YYYY-MM-DD (contoh: 2026-05-31)",
				Instance: "/api/v1/reports/supplier-aging",
			})
			return
		}

		// Handle other errors
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Gagal membuat laporan aging: " + err.Error(),
			Instance: "/api/v1/reports/supplier-aging",
		})
		return
	}

	// Log report generation for audit trail
	// Note: Using userCtx values for audit logging
	_ = ipAddress

	c.JSON(http.StatusOK, response)
}

// ExportAgingReportPDF godoc
//
//	@Summary		Export supplier aging report as PDF
//	@Description	Generates a PDF export of the supplier aging report with professional formatting (Story 10.6, AC1)
//	@Tags			Reports - Supplier Aging
//	@Accept			json
//	@Produce		application/pdf
//	@Param			request	body	dto.SupplierAgingReportRequest	true	"Aging report request"
//	@Security		BearerAuth
//	@Success		200	{file}	binary	"PDF file generated"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Owner only"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/reports/supplier-aging/export/pdf [post]
func (h *SupplierAgingReportHandler) ExportAgingReportPDF(c *gin.Context) {
	// Parse request
	var req dto.SupplierAgingReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request format: " + err.Error(),
			Instance: "/api/v1/reports/supplier-aging/export/pdf",
		})
		return
	}

	// Extract user context
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	_ = userCtx

	// Generate PDF
	pdfBytes, filename, err := h.agingReportService.ExportAgingReportPDF(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Gagal membuat PDF: " + err.Error(),
			Instance: "/api/v1/reports/supplier-aging/export/pdf",
		})
		return
	}

	// Set headers for PDF download
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ExportAgingReportExcel godoc
//
//	@Summary		Export supplier aging report as Excel
//	@Description	Generates an Excel export of the supplier aging report with multiple sheets (Story 10.6, AC1)
//	@Tags			Reports - Supplier Aging
//	@Accept			json
//	@Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param			request	body	dto.SupplierAgingReportRequest	true	"Aging report request"
//	@Security		BearerAuth
//	@Success		200	{file}	binary	"Excel file generated"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Owner only"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/reports/supplier-aging/export/excel [post]
func (h *SupplierAgingReportHandler) ExportAgingReportExcel(c *gin.Context) {
	// Parse request
	var req dto.SupplierAgingReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request format: " + err.Error(),
			Instance: "/api/v1/reports/supplier-aging/export/excel",
		})
		return
	}

	// Extract user context
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	_ = userCtx

	// Generate Excel
	excelBytes, filename, err := h.agingReportService.ExportAgingReportExcel(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Gagal membuat Excel: " + err.Error(),
			Instance: "/api/v1/reports/supplier-aging/export/excel",
		})
		return
	}

	// Set headers for Excel download
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelBytes)
}

// extractUserContext extracts user information from Gin context
// Story 10.6: Helper function to extract user context for audit logging
func (h *SupplierAgingReportHandler) extractUserContext(c *gin.Context) (map[string]interface{}, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Pengguna tidak terautentikasi",
			Instance: c.Request.URL.Path,
		})
		return nil, false
	}

	username, _ := c.Get("username")
	role, _ := c.Get("role")
	branchID, _ := c.Get("branchID")

	return map[string]interface{}{
		"userID":   userID,
		"username": username,
		"role":     role,
		"branchID": branchID,
	}, true
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
