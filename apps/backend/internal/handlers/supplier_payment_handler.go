package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// SupplierPaymentHandler handles supplier payment management operations
// Story 10.4: Handler for supplier payment recording and tracking operations
type SupplierPaymentHandler struct {
	supplierPaymentService services.SupplierPaymentService
}

// NewSupplierPaymentHandler creates a new supplier payment handler
// Story 10.4: Factory function with dependency injection
func NewSupplierPaymentHandler(supplierPaymentService services.SupplierPaymentService) *SupplierPaymentHandler {
	return &SupplierPaymentHandler{
		supplierPaymentService: supplierPaymentService,
	}
}

// RecordPayment godoc
//
//	@Summary		Record supplier payment
//	@Description	Records a new supplier payment with validation and audit logging (Story 10.4, AC1)
//	@Tags			Supplier Payment Management
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.RecordPaymentRequest	true	"Payment recording request"
//	@Security		BearerAuth
//	@Success		201	{object}	dto.SupplierPaymentResponse	"Payment recorded successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin or Owner only"
//	@Failure		404	{object}	dto.ErrorResponse	"Purchase invoice not found"
//	@Failure		409	{object}	dto.ErrorResponse	"Payment exceeds balance"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/supplier-payments [post]
func (h *SupplierPaymentHandler) RecordPayment(c *gin.Context) {
	// Parse request
	var req dto.RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request format: " + err.Error(),
			Instance: "/api/v1/supplier-payments",
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

	// Call service to record payment
	payment, err := h.supplierPaymentService.RecordPayment(c.Request.Context(), &services.RecordPaymentRequest{
		PurchaseInvoiceID: req.PurchaseInvoiceID,
		PaymentDate:      req.PaymentDate,
		PaymentAmount:    req.PaymentAmount,
		PaymentMethod:    req.PaymentMethod,
		Notes:            req.Notes,
		ReferenceNumber:  req.ReferenceNumber,
	}, userCtx.userID, ipAddress)
	if err != nil {
		// Check for not found error
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/purchase-invoice-not-found",
				Title:    "Purchase Invoice Not Found",
				Status:   http.StatusNotFound,
				Detail:   err.Error(),
				Instance: "/api/v1/supplier-payments",
			})
			return
		}

		// Check for validation error
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be") || strings.Contains(err.Error(), "cannot be") || strings.Contains(err.Error(), "exceeds") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/validation-error",
				Title:    "Validation Error",
				Status:   http.StatusBadRequest,
				Detail:   err.Error(),
				Instance: "/api/v1/supplier-payments",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to record payment: " + err.Error(),
			Instance: "/api/v1/supplier-payments",
		})
		return
	}

	// Convert to response DTO
	response := h.toSupplierPaymentResponse(payment)

	c.JSON(http.StatusCreated, response)
}

// GetSupplierPayment godoc
//
//	@Summary		Get supplier payment by ID
//	@Description	Retrieves a supplier payment by ID with invoice details (Story 10.4)
//	@Tags			Supplier Payment Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Supplier Payment ID"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.SupplierPaymentResponse	"Payment retrieved successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid payment ID"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin or Owner only"
//	@Failure		404	{object}	dto.ErrorResponse	"Payment not found"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/supplier-payments/{id} [get]
func (h *SupplierPaymentHandler) GetSupplierPayment(c *gin.Context) {
	// Extract payment ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid payment ID",
			Instance: "/api/v1/supplier-payments/" + idStr,
		})
		return
	}

	// Call service to get payment
	payment, err := h.supplierPaymentService.GetSupplierPaymentByID(c.Request.Context(), uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/payment-not-found",
				Title:    "Payment Not Found",
				Status:   http.StatusNotFound,
				Detail:   "Supplier payment not found",
				Instance: "/api/v1/supplier-payments/" + idStr,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to get payment: " + err.Error(),
			Instance: "/api/v1/supplier-payments/" + idStr,
		})
		return
	}

	// PATCH-003: Validate branch access
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}
	if !h.validateBranchAccess(c, userCtx, payment.BranchID) {
		return
	}

	// Convert to response DTO
	response := h.toSupplierPaymentResponse(payment)

	c.JSON(http.StatusOK, response)
}

// ListSupplierPayments godoc
//
//	@Summary		List supplier payments
//	@Description	Retrieves a paginated list of supplier payments with optional filtering (Story 10.4)
//	@Tags			Supplier Payment Management
//	@Accept			json
//	@Produce		json
//	@Param			purchase_invoice_id	query	int	false	"Filter by purchase invoice ID"
//	@Param			start_date	query	string	false	"Filter by payment date range start (YYYY-MM-DD)"
//	@Param			end_date	query	string	false	"Filter by payment date range end (YYYY-MM-DD)"
//	@Param			payment_method	query	string	false	"Filter by payment method"
//	@Param			branch_id	query	int	false	"Filter by branch ID"
//	@Param			page	query	int	false	"Page number"	default(1)
//	@Param			limit	query	int	false	"Items per page"	default(20)
//	@Param			sort_by	query	string	false	"Sort field"	Enums(payment_date,payment_amount)
//	@Param			sort_order	query	string	false	"Sort order"	Enums(asc,desc)
//	@Security		BearerAuth
//	@Success		200	{object}	dto.SupplierPaymentListResponse	"Payments retrieved successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request parameters"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin or Owner only"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/supplier-payments [get]
func (h *SupplierPaymentHandler) ListSupplierPayments(c *gin.Context) {
	// Parse query parameters
	var filter services.SupplierPaymentListFilter

	// Parse purchase_invoice_id
	if purchaseInvoiceIDStr := c.Query("purchase_invoice_id"); purchaseInvoiceIDStr != "" {
		purchaseInvoiceID, err := strconv.ParseUint(purchaseInvoiceIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/validation-error",
				Title:    "Validation Error",
				Status:   http.StatusBadRequest,
				Detail:   "Invalid purchase_invoice_id parameter",
				Instance: "/api/v1/supplier-payments",
			})
			return
		}
		purchaseInvoiceIDUint := uint(purchaseInvoiceID)
		filter.PurchaseInvoiceID = &purchaseInvoiceIDUint
	}

	// Parse date range filters
	if startDate := c.Query("start_date"); startDate != "" {
		filter.StartDate = &startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		filter.EndDate = &endDate
	}

	// Parse payment method
	if paymentMethod := c.Query("payment_method"); paymentMethod != "" {
		filter.PaymentMethod = &paymentMethod
	}

	// Parse branch_id
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		branchID, err := strconv.ParseUint(branchIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/validation-error",
				Title:    "Validation Error",
				Status:   http.StatusBadRequest,
				Detail:   "Invalid branch_id parameter",
				Instance: "/api/v1/supplier-payments",
			})
			return
		}
		branchIDUint := uint(branchID)
		filter.BranchID = &branchIDUint
	}

	// PATCH-019: Parse pagination with defaults
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Parse sorting
	filter.SortBy = c.DefaultQuery("sort_by", "payment_date")
	filter.SortOrder = c.DefaultQuery("sort_order", "desc")

	// Call service to list payments
	payments, total, err := h.supplierPaymentService.ListSupplierPayments(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to list payments: " + err.Error(),
			Instance: "/api/v1/supplier-payments",
		})
		return
	}

	// Convert to response DTO
	response := h.toSupplierPaymentListResponse(payments, total, filter.Page, filter.Limit)

	c.JSON(http.StatusOK, response)
}

// GetPaymentHistoryBySupplier godoc
//
//	@Summary		Get payment history by supplier
//	@Description	Retrieves payment history grouped by supplier with invoice details (Story 10.4, AC2)
//	@Tags			Supplier Payment Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Supplier ID"
//	@Param			start_date	query	string	false	"Filter by payment date range start (YYYY-MM-DD)"
//	@Param			end_date	query	string	false	"Filter by payment date range end (YYYY-MM-DD)"
//	@Param			page	query	int	false	"Page number"	default(1)
//	@Param			limit	query	int	false	"Items per page"	default(20)
//	@Security		BearerAuth
//	@Success		200	{object}	dto.PaymentHistoryListResponse	"Payment history retrieved successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request parameters"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin or Owner only"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/suppliers/{id}/payment-history [get]
func (h *SupplierPaymentHandler) GetPaymentHistoryBySupplier(c *gin.Context) {
	// Extract supplier ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid supplier ID",
			Instance: "/api/v1/suppliers/" + idStr + "/payment-history",
		})
		return
	}

	// Parse query parameters
	var filter services.PaymentHistoryFilter

	// Parse date range filters
	if startDate := c.Query("start_date"); startDate != "" {
		filter.StartDate = &startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		filter.EndDate = &endDate
	}

	// PATCH-019: Parse pagination with defaults
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Call service to get payment history
	history, err := h.supplierPaymentService.GetPaymentHistoryBySupplier(c.Request.Context(), uint(id), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to get payment history: " + err.Error(),
			Instance: "/api/v1/suppliers/" + idStr + "/payment-history",
		})
		return
	}

	// Convert service response to DTO response
	data := make([]*dto.PaymentHistoryResponse, len(history))
	for i, h := range history {
		data[i] = &dto.PaymentHistoryResponse{
			ID:                h.ID,
			PaymentDate:       h.PaymentDate,
			PaymentAmount:     h.PaymentAmount,
			PaymentMethod:     h.PaymentMethod,
			Notes:             h.Notes,
			ReferenceNumber:   h.ReferenceNumber,
			InvoiceNumber:     h.InvoiceNumber,
			InvoiceDate:       h.InvoiceDate,
			InvoiceTotalAmount: h.InvoiceTotalAmount,
			RemainingBalance:  h.RemainingBalance,
		}
	}

	// Convert to response DTO
	response := dto.PaymentHistoryListResponse{
		Data: data,
	}

	c.JSON(http.StatusOK, response)
}

// supplierPaymentUserContext holds validated user context information
type supplierPaymentUserContext struct {
	userID   uint
	username string
	branchID uint
	userRole string
}

// extractUserContext safely extracts and validates user context from Gin context
// Handles multiple possible types for user_id from different auth middleware implementations
// PATCH-003: Now also extracts branch_id and user_role for access control
func (h *SupplierPaymentHandler) extractUserContext(c *gin.Context) (supplierPaymentUserContext, bool) {
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
		return supplierPaymentUserContext{}, false
	}

	// Robust type conversion for userID - handles uint, int, int64, float64, and string
	var paymentUserID uint
	switch v := userID.(type) {
	case uint:
		paymentUserID = v
	case int:
		paymentUserID = uint(v)
	case int64:
		paymentUserID = uint(v)
	case float64:
		paymentUserID = uint(v)
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
			return supplierPaymentUserContext{}, false
		}
		paymentUserID = uint(parsed)
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Invalid user context type",
			Instance: c.Request.URL.Path,
		})
		return supplierPaymentUserContext{}, false
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
		return supplierPaymentUserContext{}, false
	}

	paymentUsername, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Invalid username context type",
			Instance: c.Request.URL.Path,
		})
		return supplierPaymentUserContext{}, false
	}

	// PATCH-003: Extract branch_id with type safety
	var paymentBranchID uint
	branchID, exists := c.Get("branch_id")
	if exists {
		switch v := branchID.(type) {
		case uint:
			paymentBranchID = v
		case int:
			paymentBranchID = uint(v)
		case int64:
			paymentBranchID = uint(v)
		case float64:
			paymentBranchID = uint(v)
		default:
			// If branch_id exists but invalid type, default to 0
			paymentBranchID = 0
		}
	}

	// PATCH-003: Extract user_role for access control
	var paymentUserRole string
	userRole, exists := c.Get("user_role")
	if exists {
		paymentUserRole, _ = userRole.(string)
	}

	// Validate user ID is not zero
	if paymentUserID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Invalid user ID",
			Instance: c.Request.URL.Path,
		})
		return supplierPaymentUserContext{}, false
	}

	// Validate username is not empty
	if paymentUsername == "" {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Invalid username",
			Instance: c.Request.URL.Path,
		})
		return supplierPaymentUserContext{}, false
	}

	return supplierPaymentUserContext{
		userID:   paymentUserID,
		username: paymentUsername,
		branchID: paymentBranchID,
		userRole: paymentUserRole,
	}, true
}

// PATCH-003: validateBranchAccess checks if user has access to the specified branch
// Owner and System Admin roles can access all branches, others can only access their assigned branch
func (h *SupplierPaymentHandler) validateBranchAccess(c *gin.Context, userCtx supplierPaymentUserContext, paymentBranchID uint) bool {
	// Owners and System Admins can access all branches
	if userCtx.userRole == "OWNER" || userCtx.userRole == "SYSTEM_ADMIN" {
		return true
	}

	// For other roles, check if user's branch matches payment's branch
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

	if userCtx.branchID != paymentBranchID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Type:     "/errors/forbidden",
			Title:    "Forbidden",
			Status:   http.StatusForbidden,
			Detail:   "You do not have permission to access payments from this branch",
			Instance: c.Request.URL.Path,
		})
		return false
	}

	return true
}

// toSupplierPaymentResponse converts a SupplierPayment model to SupplierPaymentResponse DTO
// Story 10.4: Helper function for response conversion
func (h *SupplierPaymentHandler) toSupplierPaymentResponse(payment *models.SupplierPayment) dto.SupplierPaymentResponse {
	response := dto.SupplierPaymentResponse{
		ID:                payment.ID,
		PurchaseInvoiceID: payment.PurchaseInvoiceID,
		PaymentDate:       payment.PaymentDate.Format("2006-01-02"),
		PaymentAmount:     payment.PaymentAmount,
		PaymentMethod:     payment.PaymentMethod,
		Notes:             payment.Notes,
		ReferenceNumber:   payment.ReferenceNumber,
		BranchID:          payment.BranchID,
		CreatedBy:         payment.CreatedBy,
		CreatedAt:         payment.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         payment.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Add invoice details if available (safe nil check)
	if payment.PurchaseInvoice.ID != 0 {
		response.InvoiceNumber = payment.PurchaseInvoice.InvoiceNumber
		response.InvoiceDate = payment.PurchaseInvoice.InvoiceDate.Format("2006-01-02")
		response.PaymentStatus = payment.PurchaseInvoice.PaymentStatus

		// Calculate remaining balance
		// This is calculated in the service layer via eager loading
		// For now, we'll set it to 0 and let the service layer populate it
		response.RemainingBalance = 0
	}

	return response
}

// toSupplierPaymentListResponse converts a list of SupplierPayment models to SupplierPaymentListResponse DTO
// Story 10.4: Helper function for paginated list response conversion
// PATCH-019: Handle edge case where total is 0 or page exceeds available pages
func (h *SupplierPaymentHandler) toSupplierPaymentListResponse(payments []*models.SupplierPayment, total int64, page, limit int) dto.SupplierPaymentListResponse {
	// PATCH-019: Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	data := make([]*dto.SupplierPaymentResponse, len(payments))
	for i, payment := range payments {
		paymentResponse := h.toSupplierPaymentResponse(payment)

		// Calculate remaining balance for each payment
		// We need to get total paid for this invoice
		if payment.PurchaseInvoice.ID != 0 {
			totalPaid := payment.PurchaseInvoice.TotalAmount - payment.PurchaseInvoice.TotalAmount
			// This should be populated by the service layer
			remainingBalance := payment.PurchaseInvoice.TotalAmount - totalPaid
			paymentResponse.RemainingBalance = remainingBalance
		}

		data[i] = &paymentResponse
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

	return dto.SupplierPaymentListResponse{
		Data: data,
		Pagination: dto.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
