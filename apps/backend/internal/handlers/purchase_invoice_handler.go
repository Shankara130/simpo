package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// PurchaseInvoiceHandler handles purchase invoice management operations
// Story 10.2: Handler for purchase invoice CRUD operations with audit logging
type PurchaseInvoiceHandler struct {
	purchaseInvoiceService services.PurchaseInvoiceService
}

// NewPurchaseInvoiceHandler creates a new purchase invoice handler
// Story 10.2: Factory function with dependency injection
func NewPurchaseInvoiceHandler(purchaseInvoiceService services.PurchaseInvoiceService) *PurchaseInvoiceHandler {
	return &PurchaseInvoiceHandler{
		purchaseInvoiceService: purchaseInvoiceService,
	}
}

// CreatePurchaseInvoice godoc
//
//	@Summary		Create new purchase invoice
//	@Description	Creates a new purchase invoice with validation and audit logging (Story 10.2, AC1)
//	@Tags			Purchase Invoice Management
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.CreatePurchaseInvoiceRequest	true	"Purchase invoice creation request"
//	@Security		BearerAuth
//	@Success		201	{object}	dto.PurchaseInvoiceResponse	"Purchase invoice created successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin only"
//	@Failure		409	{object}	dto.ErrorResponse	"Duplicate invoice number"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/purchase-invoices [post]
func (h *PurchaseInvoiceHandler) CreatePurchaseInvoice(c *gin.Context) {
	// Parse request
	var req dto.CreatePurchaseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request format: " + err.Error(),
			Instance: "/api/v1/purchase-invoices",
		})
		return
	}

	// Extract user context for audit logging
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// Parse invoice date
	invoiceDate, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid invoice date format, expected YYYY-MM-DD: " + err.Error(),
			Instance: "/api/v1/purchase-invoices",
		})
		return
	}

	// Create purchase invoice model from request
	invoice := &models.PurchaseInvoice{
		InvoiceNumber: req.InvoiceNumber,
		InvoiceDate:   invoiceDate,
		SupplierID:    req.SupplierID,
		BranchID:      req.BranchID,
		Notes:         req.Notes,
		DocumentURL:   req.DocumentURL,
	}

	// Convert DTO items to service items
	items := make([]services.CreatePurchaseInvoiceItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = services.CreatePurchaseInvoiceItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitCost:  item.UnitCost,
		}
	}

	// Call service to create purchase invoice
	result, err := h.purchaseInvoiceService.CreatePurchaseInvoice(c.Request.Context(), invoice, items, userCtx.userID, ipAddress)
	if err != nil {
		// Check for duplicate invoice number error
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Type:     "/errors/duplicate-invoice",
				Title:    "Duplicate Invoice",
				Status:   http.StatusConflict,
				Detail:   err.Error(),
				Instance: "/api/v1/purchase-invoices",
			})
			return
		}

		// Check for validation error
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be") || strings.Contains(err.Error(), "cannot be") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/validation-error",
				Title:    "Validation Error",
				Status:   http.StatusBadRequest,
				Detail:   err.Error(),
				Instance: "/api/v1/purchase-invoices",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to create purchase invoice: " + err.Error(),
			Instance: "/api/v1/purchase-invoices",
		})
		return
	}

	// Convert to response DTO
	response := h.toPurchaseInvoiceResponse(result)

	c.JSON(http.StatusCreated, response)
}

// GetPurchaseInvoice godoc
//
//	@Summary		Get purchase invoice by ID
//	@Description	Retrieves a purchase invoice by ID with line items and supplier details (Story 10.2, AC3)
//	@Tags			Purchase Invoice Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Purchase Invoice ID"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.PurchaseInvoiceResponse	"Purchase invoice retrieved successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid purchase invoice ID"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin/Owner only"
//	@Failure		404	{object}	dto.ErrorResponse	"Purchase invoice not found"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/purchase-invoices/{id} [get]
func (h *PurchaseInvoiceHandler) GetPurchaseInvoice(c *gin.Context) {
	// Extract purchase invoice ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid purchase invoice ID",
			Instance: "/api/v1/purchase-invoices/" + idStr,
		})
		return
	}

	// Call service to get purchase invoice
	invoice, err := h.purchaseInvoiceService.GetPurchaseInvoiceByID(c.Request.Context(), uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/purchase-invoice-not-found",
				Title:    "Purchase Invoice Not Found",
				Status:   http.StatusNotFound,
				Detail:   "Purchase invoice not found",
				Instance: "/api/v1/purchase-invoices/" + idStr,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to get purchase invoice: " + err.Error(),
			Instance: "/api/v1/purchase-invoices/" + idStr,
		})
		return
	}

	// PATCH-003: Validate branch access
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}
	if !h.validateBranchAccess(c, userCtx, invoice.BranchID) {
		return
	}

	// Convert to response DTO
	response := h.toPurchaseInvoiceResponse(invoice)

	c.JSON(http.StatusOK, response)
}

// ListPurchaseInvoices godoc
//
//	@Summary		List purchase invoices
//	@Description	Retrieves a paginated list of purchase invoices with optional filtering (Story 10.2, AC2)
//	@Tags			Purchase Invoice Management
//	@Accept			json
//	@Produce		json
//	@Param			supplier_id	query	int	false	"Filter by supplier ID"
//	@Param			start_date	query	string	false	"Filter by start date (YYYY-MM-DD)"
//	@Param			end_date	query	string	false	"Filter by end date (YYYY-MM-DD)"
//	@Param			payment_status	query	string	false	"Filter by payment status"
//	@Param			search	query	string	false	"Search by invoice number"
//	@Param			page	query	int	false	"Page number"	default(1)
//	@Param			limit	query	int	false	"Items per page"	default(20)
//	@Param			sort_by	query	string	false	"Sort field"	Enums(invoice_date,total_amount)
//	@Param			sort_order	query	string	false	"Sort order"	Enums(asc,desc)
//	@Security		BearerAuth
//	@Success		200	{object}	dto.PurchaseInvoiceListResponse	"Purchase invoices retrieved successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request parameters"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin/Owner only"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/purchase-invoices [get]
func (h *PurchaseInvoiceHandler) ListPurchaseInvoices(c *gin.Context) {
	// Parse query parameters
	var filter services.PurchaseInvoiceListFilter

	// Parse supplier_id
	if supplierIDStr := c.Query("supplierId"); supplierIDStr != "" {
		supplierID, err := strconv.ParseUint(supplierIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/validation-error",
				Title:    "Validation Error",
				Status:   http.StatusBadRequest,
				Detail:   "Invalid supplier_id parameter",
				Instance: "/api/v1/purchase-invoices",
			})
			return
		}
		supplierIDUint := uint(supplierID)
		filter.SupplierID = &supplierIDUint
	}

	// Parse date range filters
	if startDate := c.Query("start_date"); startDate != "" {
		filter.StartDate = &startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		filter.EndDate = &endDate
	}

	// Parse payment status
	if paymentStatus := c.Query("payment_status"); paymentStatus != "" {
		filter.PaymentStatus = &paymentStatus
	}

	// PATCH-021: Parse and sanitize search query
	searchQuery := c.DefaultQuery("search", "")
	if searchQuery != "" {
		// Trim and validate search query
		searchQuery = strings.TrimSpace(searchQuery)
		// Remove potentially dangerous characters
		searchQuery = strings.ReplaceAll(searchQuery, "%", "")
		searchQuery = strings.ReplaceAll(searchQuery, "_", "")
		// Validate minimum length
		if len(searchQuery) < 2 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/validation-error",
				Title:    "Validation Error",
				Status:   http.StatusBadRequest,
				Detail:   "Search query must be at least 2 characters",
				Instance: "/api/v1/purchase-invoices",
			})
			return
		}
		// Limit maximum search length
		if len(searchQuery) > 100 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/validation-error",
				Title:    "Validation Error",
				Status:   http.StatusBadRequest,
				Detail:   "Search query too long (maximum 100 characters)",
				Instance: "/api/v1/purchase-invoices",
			})
			return
		}
		filter.SearchQuery = searchQuery
	} else {
		filter.SearchQuery = ""
	}

	// Parse pagination
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Parse sorting
	filter.SortBy = c.DefaultQuery("sort_by", "invoice_date")
	filter.SortOrder = c.DefaultQuery("sort_order", "desc")

	// Call service to list purchase invoices
	invoices, total, err := h.purchaseInvoiceService.ListPurchaseInvoices(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to list purchase invoices: " + err.Error(),
			Instance: "/api/v1/purchase-invoices",
		})
		return
	}

	// Convert to response DTO
	response := h.toPurchaseInvoiceListResponse(invoices, total, filter.Page, filter.Limit)

	c.JSON(http.StatusOK, response)
}

// UpdatePurchaseInvoice godoc
//
//	@Summary		Update purchase invoice
//	@Description	Updates an existing purchase invoice with validation and audit logging (Story 10.2)
//	@Tags			Purchase Invoice Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Purchase Invoice ID"
//	@Param			request	body	dto.UpdatePurchaseInvoiceRequest	true	"Purchase invoice update request"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.PurchaseInvoiceResponse	"Purchase invoice updated successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin only"
//	@Failure		404	{object}	dto.ErrorResponse	"Purchase invoice not found"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/purchase-invoices/{id} [put]
func (h *PurchaseInvoiceHandler) UpdatePurchaseInvoice(c *gin.Context) {
	// Extract purchase invoice ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid purchase invoice ID",
			Instance: "/api/v1/purchase-invoices/" + idStr,
		})
		return
	}

	// Parse request
	var req dto.UpdatePurchaseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request format: " + err.Error(),
			Instance: "/api/v1/purchase-invoices/" + idStr,
		})
		return
	}

	// Extract user context for audit logging
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// Convert DTO items to service items
	items := make([]services.CreatePurchaseInvoiceItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = services.CreatePurchaseInvoiceItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitCost:  item.UnitCost,
		}
	}

	// Create update request
	updateReq := &services.UpdatePurchaseInvoiceRequest{
		InvoiceNumber: req.InvoiceNumber,
		InvoiceDate:   req.InvoiceDate,
		SupplierID:    req.SupplierID,
		Notes:         req.Notes,
		DocumentURL:   req.DocumentURL,
		Items:         items,
		Reason:        sanitizePurchaseInvoiceReason(req.Reason),
	}

	// Call service to update purchase invoice
	result, err := h.purchaseInvoiceService.UpdatePurchaseInvoice(c.Request.Context(), uint(id), updateReq, userCtx.userID, ipAddress)
	if err != nil {
		// Check for not found error
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/purchase-invoice-not-found",
				Title:    "Purchase Invoice Not Found",
				Status:   http.StatusNotFound,
				Detail:   "Purchase invoice not found",
				Instance: "/api/v1/purchase-invoices/" + idStr,
			})
			return
		}

		// Check for validation error
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "reason") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/validation-error",
				Title:    "Validation Error",
				Status:   http.StatusBadRequest,
				Detail:   err.Error(),
				Instance: "/api/v1/purchase-invoices/" + idStr,
			})
			return
		}

		// Check for duplicate error
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Type:     "/errors/duplicate-invoice",
				Title:    "Duplicate Invoice",
				Status:   http.StatusConflict,
				Detail:   err.Error(),
				Instance: "/api/v1/purchase-invoices/" + idStr,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to update purchase invoice: " + err.Error(),
			Instance: "/api/v1/purchase-invoices/" + idStr,
		})
		return
	}

	// Convert to response DTO
	response := h.toPurchaseInvoiceResponse(result)

	c.JSON(http.StatusOK, response)
}

// DeletePurchaseInvoice godoc
//
//	@Summary		Delete purchase invoice
//	@Description	Deletes (soft-deletes) a purchase invoice with audit logging (Story 10.2)
//	@Tags			Purchase Invoice Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Purchase Invoice ID"
//	@Param			request	body	dto.DeactivatePurchaseInvoiceRequest	true	"Deletion request"
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Purchase invoice deleted successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin only"
//	@Failure		404	{object}	dto.ErrorResponse	"Purchase invoice not found"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/purchase-invoices/{id} [delete]
func (h *PurchaseInvoiceHandler) DeletePurchaseInvoice(c *gin.Context) {
	// Extract purchase invoice ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid purchase invoice ID",
			Instance: "/api/v1/purchase-invoices/" + idStr,
		})
		return
	}

	// Parse request
	var req dto.DeactivatePurchaseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request format: " + err.Error(),
			Instance: "/api/v1/purchase-invoices/" + idStr,
		})
		return
	}

	// Extract user context for audit logging
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// Call service to delete purchase invoice
	err = h.purchaseInvoiceService.DeletePurchaseInvoice(c.Request.Context(), uint(id), userCtx.userID, ipAddress)
	if err != nil {
		// Check for not found error
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/purchase-invoice-not-found",
				Title:    "Purchase Invoice Not Found",
				Status:   http.StatusNotFound,
				Detail:   "Purchase invoice not found",
				Instance: "/api/v1/purchase-invoices/" + idStr,
			})
			return
		}

		// Check for validation error
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "reason") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/validation-error",
				Title:    "Validation Error",
				Status:   http.StatusBadRequest,
				Detail:   err.Error(),
				Instance: "/api/v1/purchase-invoices/" + idStr,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to delete purchase invoice: " + err.Error(),
			Instance: "/api/v1/purchase-invoices/" + idStr,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          uint(id),
		"message":     "Purchase invoice deleted",
		"deletedAt":   time.Now().Format(time.RFC3339),
		"reason":      sanitizePurchaseInvoiceReason(req.Reason),
	})
}

// purchaseInvoiceUserContext holds validated user context information
type purchaseInvoiceUserContext struct {
	userID   uint
	username  string
	branchID  uint
	userRole  string
}

// extractUserContext safely extracts and validates user context from Gin context
// Handles multiple possible types for user_id from different auth middleware implementations
// PATCH-003: Now also extracts branch_id and user_role for access control
func (h *PurchaseInvoiceHandler) extractUserContext(c *gin.Context) (purchaseInvoiceUserContext, bool) {
	// Extract user ID with type safety check - handle multiple possible types
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "User not authenticated",
			Instance: c.Request.URL.Path,
		})
		return purchaseInvoiceUserContext{}, false
	}

	// Robust type conversion for userID - handles uint, int, int64, float64, and string
	var invoiceUserID uint
	switch v := userID.(type) {
	case uint:
		invoiceUserID = v
	case int:
		invoiceUserID = uint(v)
	case int64:
		invoiceUserID = uint(v)
	case float64:
		invoiceUserID = uint(v)
	case string:
		// Try to parse string as uint
		parsed, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Type:     "/errors/unauthorized",
				Title:    "Unauthorized",
				Status:   http.StatusUnauthorized,
				Detail:   "Invalid user ID format",
				Instance: c.Request.URL.Path,
			})
			return purchaseInvoiceUserContext{}, false
		}
		invoiceUserID = uint(parsed)
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Invalid user context type",
			Instance: c.Request.URL.Path,
		})
		return purchaseInvoiceUserContext{}, false
	}

	// Extract username with type safety check
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Username not found",
			Instance: c.Request.URL.Path,
		})
		return purchaseInvoiceUserContext{}, false
	}

	invoiceUsername, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Invalid username context type",
			Instance: c.Request.URL.Path,
		})
		return purchaseInvoiceUserContext{}, false
	}

	// PATCH-003: Extract branch_id with type safety
	var invoiceBranchID uint
	branchID, exists := c.Get("branch_id")
	if exists {
		switch v := branchID.(type) {
		case uint:
			invoiceBranchID = v
		case int:
			invoiceBranchID = uint(v)
		case int64:
			invoiceBranchID = uint(v)
		case float64:
			invoiceBranchID = uint(v)
		default:
			// If branch_id exists but invalid type, default to 0
			invoiceBranchID = 0
		}
	}

	// PATCH-003: Extract user_role for access control
	var invoiceUserRole string
	userRole, exists := c.Get("user_role")
	if exists {
		invoiceUserRole, _ = userRole.(string)
	}

	// Validate user ID is not zero
	if invoiceUserID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Invalid user ID",
			Instance: c.Request.URL.Path,
		})
		return purchaseInvoiceUserContext{}, false
	}

	// Validate username is not empty
	if invoiceUsername == "" {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Invalid username",
			Instance: c.Request.URL.Path,
		})
		return purchaseInvoiceUserContext{}, false
	}

	return purchaseInvoiceUserContext{
		userID:   invoiceUserID,
		username:  invoiceUsername,
		branchID:  invoiceBranchID,
		userRole:  invoiceUserRole,
	}, true
}

// PATCH-003: validateBranchAccess checks if user has access to the specified branch
// Owner and System Admin roles can access all branches, others can only access their assigned branch
func (h *PurchaseInvoiceHandler) validateBranchAccess(c *gin.Context, userCtx purchaseInvoiceUserContext, invoiceBranchID uint) bool {
	// Owners and System Admins can access all branches
	if userCtx.userRole == "OWNER" || userCtx.userRole == "SYSTEM_ADMIN" {
		return true
	}

	// For other roles, check if user's branch matches invoice's branch
	if userCtx.branchID == 0 {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Type:     "/errors/forbidden",
			Title:    "Forbidden",
			Status:   http.StatusForbidden,
			Detail:   "User branch assignment not found",
			Instance: c.Request.URL.Path,
		})
		return false
	}

	if userCtx.branchID != invoiceBranchID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Type:     "/errors/forbidden",
			Title:    "Forbidden",
			Status:   http.StatusForbidden,
			Detail:   "You do not have permission to access invoices from this branch",
			Instance: c.Request.URL.Path,
		})
		return false
	}

	return true
}

// toPurchaseInvoiceResponse converts a PurchaseInvoice model to PurchaseInvoiceResponse DTO
// Story 10.2: Helper function for response conversion
func (h *PurchaseInvoiceHandler) toPurchaseInvoiceResponse(invoice *models.PurchaseInvoice) dto.PurchaseInvoiceResponse {
	response := dto.PurchaseInvoiceResponse{
		ID:            invoice.ID,
		InvoiceNumber: invoice.InvoiceNumber,
		InvoiceDate:   invoice.InvoiceDate.Format("2006-01-02"), // DN-008: Use YYYY-MM-DD format for consistency
		SupplierID:    invoice.SupplierID,
		BranchID:      invoice.BranchID,
		TotalAmount:   invoice.TotalAmount,
		PaymentStatus: invoice.PaymentStatus,
		Notes:         invoice.Notes,
		DocumentURL:   invoice.DocumentURL,
		Items:         make([]dto.PurchaseInvoiceItemResponse, 0),
		CreatedAt:     invoice.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     invoice.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// DN-002: Add supplier contact fields if available (safe nil check)
	if invoice.Supplier.ID != 0 && invoice.Supplier.Name != "" {
		response.SupplierName = invoice.Supplier.Name
		response.SupplierContactPerson = invoice.Supplier.ContactPerson
		response.SupplierPhone = invoice.Supplier.Phone
		response.SupplierEmail = invoice.Supplier.Email
		response.SupplierAddress = invoice.Supplier.Address
	}

	// DN-004: Add branch name if available (safe nil check)
	if invoice.Branch.ID != 0 && invoice.Branch.Name != "" {
		response.BranchName = invoice.Branch.Name
	}

	// Convert line items (PATCH-016: safe nil check for items slice)
	if len(invoice.Items) > 0 {
		for _, item := range invoice.Items {
			itemResponse := dto.PurchaseInvoiceItemResponse{
				ID:        item.ID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				UnitCost:  item.UnitCost,
				Subtotal:  item.Subtotal,
			}

			// Add product details if available (safe nil check)
			if item.Product.ID != 0 && item.Product.Name != "" {
				itemResponse.ProductName = item.Product.Name
				itemResponse.ProductSKU = item.Product.SKU
			}

			response.Items = append(response.Items, itemResponse)
		}
	}

	return response
}

// toPurchaseInvoiceListResponse converts a list of PurchaseInvoice models to PurchaseInvoiceListResponse DTO
// Story 10.2: Helper function for paginated list response conversion
// PATCH-019: Handle edge case where total is 0 or page exceeds available pages
func (h *PurchaseInvoiceHandler) toPurchaseInvoiceListResponse(invoices []*models.PurchaseInvoice, total int64, page, limit int) dto.PurchaseInvoiceListResponse {
	// PATCH-019: Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	data := make([]dto.PurchaseInvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		data[i] = h.toPurchaseInvoiceResponse(invoice)
	}

	// PATCH-019: Calculate total pages safely
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	// PATCH-019: If requested page exceeds available pages, cap it
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	return dto.PurchaseInvoiceListResponse{
		Data: data,
		Pagination: dto.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

// sanitizePurchaseInvoiceReason sanitizes user-provided reason text to prevent injection attacks
func sanitizePurchaseInvoiceReason(reason string) string {
	// Trim whitespace
	reason = strings.TrimSpace(reason)

	// Remove any null bytes
	reason = strings.ReplaceAll(reason, "\x00", "")

	// PATCH-022: Limit length to prevent abuse with indication
	const maxLength = 500
	if len(reason) > maxLength {
		// Truncate and add indicator that truncation occurred
		reason = reason[:maxLength-3] + "..."
	}

	return reason
}
