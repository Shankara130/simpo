package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/utils"
)

// Story 10.7: Implement Supplier Transaction Audit Trail

// SupplierAuditHandler handles supplier audit trail API requests
// AC: Queryable audit trail with filters for compliance inspections
type SupplierAuditHandler struct {
	auditService services.SupplierAuditService
}

// NewSupplierAuditHandler creates a new instance of SupplierAuditHandler
func NewSupplierAuditHandler(auditService services.SupplierAuditService) *SupplierAuditHandler {
	return &SupplierAuditHandler{
		auditService: auditService,
	}
}

// QueryAuditTrail handles GET /api/v1/audit/supplier
// AC: Query audit trail with filters for date range, transaction type, entity, user, branch
func (h *SupplierAuditHandler) QueryAuditTrail(c *gin.Context) {
	// Parse query parameters
	var request services.SupplierAuditQueryRequest

	// Date range filters (optional)
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid start_date format. Use YYYY-MM-DD format.",
				"instance": c.Request.URL.Path,
			})
			return
		}
		request.StartDate = &startDate
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid end_date format. Use YYYY-MM-DD format.",
				"instance": c.Request.URL.Path,
			})
			return
		}
		request.EndDate = &endDate
	}

	// Transaction type filter (optional)
	if transactionType := c.Query("transaction_type"); transactionType != "" {
		request.TransactionType = &transactionType
	}

	// Entity type filter (optional)
	if entityType := c.Query("entity_type"); entityType != "" {
		request.EntityType = &entityType
	}

	// Entity ID filter (optional)
	if entityIDStr := c.Query("entity_id"); entityIDStr != "" {
		entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid entity_id format. Must be a positive integer.",
				"instance": c.Request.URL.Path,
			})
			return
		}
		id := uint(entityID)
		request.EntityID = &id
	}

	// User filter (optional)
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid user_id format. Must be a positive integer.",
				"instance": c.Request.URL.Path,
			})
			return
		}
		id := uint(userID)
		request.UserID = &id
	}

	// Branch filter (optional)
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		branchID, err := strconv.ParseUint(branchIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid branch_id format. Must be a positive integer.",
				"instance": c.Request.URL.Path,
			})
			return
		}
		id := uint(branchID)
		request.BranchID = &id
	}

	// Pagination parameters (required)
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": "Invalid page parameter. Must be a positive integer.",
			"instance": c.Request.URL.Path,
		})
		return
	}
	request.Page = &page

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": "Invalid limit parameter. Must be between 1 and 100.",
			"instance": c.Request.URL.Path,
		})
		return
	}
	request.Limit = &limit

	// Call service to query audit trail
	response, err := h.auditService.QueryAuditTrail(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":   "https://api.restful_api.com/problems/internal-server-error",
			"title":  "Internal Server Error",
			"status": http.StatusInternalServerError,
			"detail": "Failed to query audit trail. Please try again later.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAuditByEntity handles GET /api/v1/audit/supplier/entity/:type/:id
// AC: Get audit trail for specific entity (supplier, purchase_invoice, supplier_payment)
func (h *SupplierAuditHandler) GetAuditByEntity(c *gin.Context) {
	// Parse entity type from path parameter
	entityType := c.Param("type")
	if entityType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": "Entity type parameter is required.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	// Validate entity type
	validEntityTypes := map[string]bool{
		"supplier":          true,
		"purchase_invoice":  true,
		"supplier_payment":  true,
		"goods_receipt":     true,
	}
	if !validEntityTypes[entityType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": "Invalid entity type. Must be one of: supplier, purchase_invoice, supplier_payment, goods_receipt",
			"instance": c.Request.URL.Path,
		})
		return
	}

	// Parse entity ID from path parameter
	entityIDStr := c.Param("id")
	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": "Invalid entity ID. Must be a positive integer.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	// Call service to get audit by entity
	audits, err := h.auditService.GetAuditByEntityID(c.Request.Context(), entityType, uint(entityID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":   "https://api.restful_api.com/problems/internal-server-error",
			"title":  "Internal Server Error",
			"status": http.StatusInternalServerError,
			"detail": "Failed to retrieve audit trail. Please try again later.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  audits,
		"count": len(audits),
	})
}

// GetAuditByUser handles GET /api/v1/audit/supplier/user/:id
// AC: Get audit trail for specific user within a date range
func (h *SupplierAuditHandler) GetAuditByUser(c *gin.Context) {
	// Parse user ID from path parameter
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": "Invalid user ID. Must be a positive integer.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	// Parse date range from query parameters (optional)
	var startDate, endDate time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid start_date format. Use YYYY-MM-DD format.",
				"instance": c.Request.URL.Path,
			})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid end_date format. Use YYYY-MM-DD format.",
				"instance": c.Request.URL.Path,
			})
			return
		}
	}

	// Call service to get audit by user
	audits, err := h.auditService.GetAuditByUserID(c.Request.Context(), uint(userID), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":   "https://api.restful_api.com/problems/internal-server-error",
			"title":  "Internal Server Error",
			"status": http.StatusInternalServerError,
			"detail": "Failed to retrieve audit trail. Please try again later.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  audits,
		"count": len(audits),
	})
}

// ExportAuditTrailCSV handles GET /api/v1/audit/supplier/export/csv
// AC: Export audit trail for compliance inspections in CSV format
func (h *SupplierAuditHandler) ExportAuditTrailCSV(c *gin.Context) {
	// Parse query parameters
	var request services.SupplierAuditExportRequest

	// Start date (required)
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid start_date format. Use YYYY-MM-DD format.",
				"instance": c.Request.URL.Path,
			})
			return
		}
		request.StartDate = startDate
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": "start_date parameter is required for export.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	// End date (required)
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid end_date format. Use YYYY-MM-DD format.",
				"instance": c.Request.URL.Path,
			})
			return
		}
		request.EndDate = endDate
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": "end_date parameter is required for export.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	// Transaction type filter (optional)
	if transactionType := c.Query("transaction_type"); transactionType != "" {
		request.TransactionType = &transactionType
	}

	// Branch filter (optional)
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		branchID, err := strconv.ParseUint(branchIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid branch_id format. Must be a positive integer.",
				"instance": c.Request.URL.Path,
			})
			return
		}
		id := uint(branchID)
		request.BranchID = &id
	}

	// Set format to CSV
	request.Format = "csv"

	// Validate date range (maximum 1 year)
	if err := utils.ValidateExportDateRange(request.StartDate, request.EndDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": err.Error(),
			"instance": c.Request.URL.Path,
		})
		return
	}

	// Call service to get audit trail data
	audits, err := h.auditService.ExportAuditTrail(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":   "https://api.restful_api.com/problems/internal-server-error",
			"title":  "Internal Server Error",
			"status": http.StatusInternalServerError,
			"detail": "Failed to export audit trail. Please try again later.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	// Set response headers for CSV download
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\""+utils.GenerateCSVFileName(request.StartDate, request.EndDate)+"\"")

	// Write CSV directly to response writer
	if err := utils.ExportSupplierAuditTrailToCSV(audits, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":   "https://api.restful_api.com/problems/internal-server-error",
			"title":  "Internal Server Error",
			"status": http.StatusInternalServerError,
			"detail": "Failed to generate CSV file. Please try again later.",
			"instance": c.Request.URL.Path,
		})
		return
	}
}

// ExportAuditTrailPDF handles GET /api/v1/audit/supplier/export/pdf
// AC: Export audit trail for compliance inspections in PDF format
// Note: PDF export functionality will be implemented in Task 9
func (h *SupplierAuditHandler) ExportAuditTrailPDF(c *gin.Context) {
	// Parse query parameters
	var request services.SupplierAuditExportRequest

	// Start date (required)
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid start_date format. Use YYYY-MM-DD format.",
				"instance": c.Request.URL.Path,
			})
			return
		}
		request.StartDate = startDate
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": "start_date parameter is required for export.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	// End date (required)
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid end_date format. Use YYYY-MM-DD format.",
				"instance": c.Request.URL.Path,
			})
			return
		}
		request.EndDate = endDate
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": "end_date parameter is required for export.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	// Transaction type filter (optional)
	if transactionType := c.Query("transaction_type"); transactionType != "" {
		request.TransactionType = &transactionType
	}

	// Branch filter (optional)
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		branchID, err := strconv.ParseUint(branchIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"type":   "https://api.restful_api.com/problems/invalid-request",
				"title":  "Invalid Request",
				"status": http.StatusBadRequest,
				"detail": "Invalid branch_id format. Must be a positive integer.",
				"instance": c.Request.URL.Path,
			})
			return
		}
		id := uint(branchID)
		request.BranchID = &id
	}

	// Set format to PDF
	request.Format = "pdf"

	// Validate date range (maximum 1 year)
	if err := utils.ValidateExportDateRange(request.StartDate, request.EndDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":   "https://api.restful_api.com/problems/invalid-request",
			"title":  "Invalid Request",
			"status": http.StatusBadRequest,
			"detail": err.Error(),
			"instance": c.Request.URL.Path,
		})
		return
	}

	// Call service to get audit trail data
	audits, err := h.auditService.ExportAuditTrail(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":   "https://api.restful_api.com/problems/internal-server-error",
			"title":  "Internal Server Error",
			"status": http.StatusInternalServerError,
			"detail": "Failed to export audit trail. Please try again later.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	// Generate PDF report
	pharmacyName := "Simpo Pharmacy" // Default pharmacy name
	pdf, err := utils.GenerateAuditTrailPDF(audits, request.StartDate, request.EndDate, pharmacyName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":   "https://api.restful_api.com/problems/internal-server-error",
			"title":  "Internal Server Error",
			"status": http.StatusInternalServerError,
			"detail": "Failed to generate PDF file. Please try again later.",
			"instance": c.Request.URL.Path,
		})
		return
	}

	// Set response headers for PDF download
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+utils.GeneratePDFFileName(request.StartDate, request.EndDate)+"\"")

	// Write PDF to response
	err = pdf.Output(c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":   "https://api.restful_api.com/problems/internal-server-error",
			"title":  "Internal Server Error",
			"status": http.StatusInternalServerError,
			"detail": "Failed to write PDF file. Please try again later.",
			"instance": c.Request.URL.Path,
		})
		return
	}
}
