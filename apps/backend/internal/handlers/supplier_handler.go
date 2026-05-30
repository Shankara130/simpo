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

// SupplierHandler handles supplier management operations
// Story 10.1: Handler for supplier CRUD operations with audit logging
type SupplierHandler struct {
	supplierService services.SupplierService
}

// NewSupplierHandler creates a new supplier handler
// Story 10.1: Factory function with dependency injection
func NewSupplierHandler(supplierService services.SupplierService) *SupplierHandler {
	return &SupplierHandler{
		supplierService: supplierService,
	}
}

// CreateSupplier godoc
//
//	@Summary		Create new supplier
//	@Description	Creates a new supplier with validation and audit logging (Story 10.1, AC1)
//	@Tags			Supplier Management
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.CreateSupplierRequest	true	"Supplier creation request"
//	@Security		BearerAuth
//	@Success		201	{object}	dto.SupplierResponse	"Supplier created successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin only"
//	@Failure		409	{object}	dto.ErrorResponse	"Supplier name already exists"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/suppliers [post]
func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	// Parse request
	var req dto.CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:   "/errors/validation-error",
			Title:  "Validation Error",
			Status: http.StatusBadRequest,
			Detail: "Invalid request format: " + err.Error(),
			Instance: "/api/v1/suppliers",
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

	// Create supplier model from request
	supplier := &models.Supplier{
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Phone:         req.Phone,
		Email:         req.Email,
		Address:       req.Address,
	}

	// Call service to create supplier
	result, err := h.supplierService.CreateSupplier(c.Request.Context(), supplier, userCtx.userID, ipAddress)
	if err != nil {
		// Check for duplicate name error
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Type:     "/errors/duplicate-supplier",
				Title:    "Duplicate Supplier",
				Status:   http.StatusConflict,
				Detail:   err.Error(),
				Instance: "/api/v1/suppliers",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to create supplier: " + err.Error(),
			Instance: "/api/v1/suppliers",
		})
		return
	}

	// Convert to response DTO
	response := dto.ToSupplierResponse(result)

	c.JSON(http.StatusCreated, response)
}

// GetSupplier godoc
//
//	@Summary		Get supplier by ID
//	@Description	Retrieves a supplier by ID (Story 10.1, AC1)
//	@Tags			Supplier Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Supplier ID"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.SupplierResponse	"Supplier retrieved successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid supplier ID"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin/Owner only"
//	@Failure		404	{object}	dto.ErrorResponse	"Supplier not found"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/suppliers/{id} [get]
func (h *SupplierHandler) GetSupplier(c *gin.Context) {
	// Extract supplier ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid supplier ID",
			Instance: "/api/v1/suppliers/" + idStr,
		})
		return
	}

	// Call service to get supplier
	supplier, err := h.supplierService.GetSupplierByID(c.Request.Context(), uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "get supplier") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/supplier-not-found",
				Title:    "Supplier Not Found",
				Status:   http.StatusNotFound,
				Detail:   "Supplier not found",
				Instance: "/api/v1/suppliers/" + idStr,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to get supplier: " + err.Error(),
			Instance: "/api/v1/suppliers/" + idStr,
		})
		return
	}

	// Convert to response DTO
	response := dto.ToSupplierResponse(supplier)

	c.JSON(http.StatusOK, response)
}

// ListSuppliers godoc
//
//	@Summary		List suppliers
//	@Description	Retrieves a paginated list of suppliers with optional filtering (Story 10.1, AC2)
//	@Tags			Supplier Management
//	@Accept			json
//	@Produce		json
//	@Param			page	query	int	false	"Page number"	default(1)
//	@Param			limit	query	int	false	"Items per page"	default(20)
//	@Param			search	query	string	false	"Search by name, contact person, or phone"
//	@Param			is_active	query	bool	false	"Filter by active status"
//	@Param			sort_by	query	string	false	"Sort field"	Enums(name,created_at)
//	@Param			sort_order	query	string	false	"Sort order"	Enums(asc,desc)
//	@Security		BearerAuth
//	@Success		200	{object}	dto.SupplierListResponse	"Suppliers retrieved successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request parameters"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin/Owner only"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/suppliers [get]
func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	isActiveStr := c.Query("is_active")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// Parse is_active filter
	var isActive *bool
	if isActiveStr != "" {
		isActiveBool := isActiveStr == "true"
		isActive = &isActiveBool
	}

	// Create filter
	filter := &services.SupplierListFilter{
		SearchQuery: search,
		IsActive:     isActive,
		Page:         page,
		Limit:        limit,
		SortBy:       sortBy,
		SortOrder:    sortOrder,
	}

	// Call service to list suppliers
	suppliers, total, err := h.supplierService.ListSuppliers(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to list suppliers: " + err.Error(),
			Instance: "/api/v1/suppliers",
		})
		return
	}

	// Convert to response DTO
	response := dto.ToSupplierListResponse(suppliers, total, page, limit)

	c.JSON(http.StatusOK, response)
}

// UpdateSupplier godoc
//
//	@Summary		Update supplier
//	@Description	Updates an existing supplier with validation and audit logging (Story 10.1, AC2)
//	@Tags			Supplier Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Supplier ID"
//	@Param			request	body	dto.UpdateSupplierRequest	true	"Supplier update request"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.SupplierResponse	"Supplier updated successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin only"
//	@Failure		404	{object}	dto.ErrorResponse	"Supplier not found"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/suppliers/{id} [put]
func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	// Extract supplier ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid supplier ID",
			Instance: "/api/v1/suppliers/" + idStr,
		})
		return
	}

	// Parse request
	var req dto.UpdateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request format: " + err.Error(),
			Instance: "/api/v1/suppliers/" + idStr,
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

	// Create update request
	updateReq := &services.UpdateSupplierRequest{
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Phone:         req.Phone,
		Email:         req.Email,
		Address:       req.Address,
		Reason:        sanitizeSupplierReason(req.Reason),
	}

	// Call service to update supplier
	result, err := h.supplierService.UpdateSupplier(c.Request.Context(), uint(id), updateReq, userCtx.userID, ipAddress)
	if err != nil {
		// Check for not found error
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/supplier-not-found",
				Title:    "Supplier Not Found",
				Status:   http.StatusNotFound,
				Detail:   "Supplier not found",
				Instance: "/api/v1/suppliers/" + idStr,
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
				Instance: "/api/v1/suppliers/" + idStr,
			})
			return
		}

		// Check for duplicate error
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Type:     "/errors/duplicate-supplier",
				Title:    "Duplicate Supplier",
				Status:   http.StatusConflict,
				Detail:   err.Error(),
				Instance: "/api/v1/suppliers/" + idStr,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to update supplier: " + err.Error(),
			Instance: "/api/v1/suppliers/" + idStr,
		})
		return
	}

	// Convert to response DTO
	response := dto.ToSupplierResponse(result)

	c.JSON(http.StatusOK, response)
}

// DeactivateSupplier godoc
//
//	@Summary		Deactivate supplier
//	@Description	Deactivates a supplier with audit logging (Story 10.1, AC3)
//	@Tags			Supplier Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Supplier ID"
//	@Param			request	body	dto.DeactivateSupplierRequest	true	"Deactivation request"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.SupplierResponse	"Supplier deactivated successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin only"
//	@Failure		404	{object}	dto.ErrorResponse	"Supplier not found"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/suppliers/{id} [delete]
func (h *SupplierHandler) DeactivateSupplier(c *gin.Context) {
	// Extract supplier ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid supplier ID",
			Instance: "/api/v1/suppliers/" + idStr,
		})
		return
	}

	// Parse request
	var req dto.DeactivateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request format: " + err.Error(),
			Instance: "/api/v1/suppliers/" + idStr,
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

	// Call service to deactivate supplier
	err = h.supplierService.DeactivateSupplier(c.Request.Context(), uint(id), sanitizeSupplierReason(req.Reason), userCtx.userID, ipAddress)
	if err != nil {
		// Check for not found error
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/supplier-not-found",
				Title:    "Supplier Not Found",
				Status:   http.StatusNotFound,
				Detail:   "Supplier not found",
				Instance: "/api/v1/suppliers/" + idStr,
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
				Instance: "/api/v1/suppliers/" + idStr,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to deactivate supplier: " + err.Error(),
			Instance: "/api/v1/suppliers/" + idStr,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            uint(id),
		"name":          "Supplier deactivated",
		"deactivatedAt": time.Now().Format(time.RFC3339),
		"reason":        sanitizeSupplierReason(req.Reason),
	})
}

// supplierUserContext holds validated user context information
type supplierUserContext struct {
	userID   uint
	username string
}

// extractUserContext safely extracts and validates user context from Gin context
// Handles multiple possible types for user_id from different auth middleware implementations
func (h *SupplierHandler) extractUserContext(c *gin.Context) (supplierUserContext, bool) {
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
		return supplierUserContext{}, false
	}

	// Robust type conversion for userID - handles uint, int, int64, float64, and string
	var supplierUserID uint
	switch v := userID.(type) {
	case uint:
		supplierUserID = v
	case int:
		supplierUserID = uint(v)
	case int64:
		supplierUserID = uint(v)
	case float64:
		supplierUserID = uint(v)
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
			return supplierUserContext{}, false
		}
		supplierUserID = uint(parsed)
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Invalid user context type",
			Instance: c.Request.URL.Path,
		})
		return supplierUserContext{}, false
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
		return supplierUserContext{}, false
	}

	supplierUsername, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Invalid username context type",
			Instance: c.Request.URL.Path,
		})
		return supplierUserContext{}, false
	}

	// Validate user ID is not zero
	if supplierUserID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Invalid user ID",
			Instance: c.Request.URL.Path,
		})
		return supplierUserContext{}, false
	}

	// Validate username is not empty
	if supplierUsername == "" {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Invalid username",
			Instance: c.Request.URL.Path,
		})
		return supplierUserContext{}, false
	}

	return supplierUserContext{
		userID:   supplierUserID,
		username: supplierUsername,
	}, true
}

// sanitizeSupplierReason sanitizes user-provided reason text to prevent injection attacks
func sanitizeSupplierReason(reason string) string {
	// Trim whitespace
	reason = strings.TrimSpace(reason)

	// Remove any null bytes
	reason = strings.ReplaceAll(reason, "\x00", "")

	// Limit length to prevent abuse
	const maxLength = 500
	if len(reason) > maxLength {
		reason = reason[:maxLength]
	}

	return reason
}
