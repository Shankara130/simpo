package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// SupplierProductCatalogHandler handles HTTP requests for supplier product catalog operations
// Story 10.5: HTTP handlers with RBAC, RFC 7807 error responses, and input validation
type SupplierProductCatalogHandler struct {
	catalogService services.SupplierProductCatalogService
}

// NewSupplierProductCatalogHandler creates a new supplier product catalog handler
// Story 10.5: Factory function for dependency injection
func NewSupplierProductCatalogHandler(
	catalogService services.SupplierProductCatalogService,
) *SupplierProductCatalogHandler {
	return &SupplierProductCatalogHandler{
		catalogService: catalogService,
	}
}

// AssociateProduct handles POST /api/v1/supplier-product-catalogs
// Story 10.5, AC1: Associate a product with a supplier and specify purchase price
func (h *SupplierProductCatalogHandler) AssociateProduct(c *gin.Context) {
	var request services.AssociateProductRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request body",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Get authenticated user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "User not authenticated",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// Call service to create catalog entry
	catalog, err := h.catalogService.AssociateProduct(c.Request.Context(), &request, userID.(uint), ipAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/not-found",
				Title:    "Not Found",
				Status:   http.StatusNotFound,
				Detail:   "Supplier, product, or branch not found",
				Instance: c.Request.URL.Path,
			})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Failed to create catalog entry",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Return created catalog entry
	c.JSON(http.StatusCreated, catalog)
}

// GetProductCatalog handles GET /api/v1/supplier-product-catalogs/:id
// Story 10.5: Get catalog entry by ID
func (h *SupplierProductCatalogHandler) GetProductCatalog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid catalog ID",
			Instance: c.Request.URL.Path,
		})
		return
	}

	catalog, err := h.catalogService.GetProductCatalogByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/not-found",
				Title:    "Not Found",
				Status:   http.StatusNotFound,
				Detail:   "Catalog entry not found",
				Instance: c.Request.URL.Path,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-error",
			Title:    "Internal Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to get catalog entry",
			Instance: c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, catalog)
}

// ListProductCatalogs handles GET /api/v1/supplier-product-catalogs
// Story 10.5: List catalog entries with pagination
func (h *SupplierProductCatalogHandler) ListProductCatalogs(c *gin.Context) {
	// Parse query parameters
	filter := &services.SupplierProductCatalogListFilter{}

	if supplierIDStr := c.Query("supplier_id"); supplierIDStr != "" {
		supplierID, err := strconv.ParseUint(supplierIDStr, 10, 64)
		if err == nil {
			sid := uint(supplierID)
			filter.SupplierID = &sid
		}
	}

	if productIDStr := c.Query("product_id"); productIDStr != "" {
		productID, err := strconv.ParseUint(productIDStr, 10, 64)
		if err == nil {
			pid := uint(productID)
			filter.ProductID = &pid
		}
	}

	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		branchID, err := strconv.ParseUint(branchIDStr, 10, 64)
		if err == nil {
			bid := uint(branchID)
			filter.BranchID = &bid
		}
	}

	if isPreferredStr := c.Query("is_preferred"); isPreferredStr != "" {
		isPreferred, err := strconv.ParseBool(isPreferredStr)
		if err == nil {
			filter.IsPreferred = &isPreferred
		}
	}

	// Pagination parameters (PATCH-012: overflow protection)
	filter.Page = 1
	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil && page > 0 {
			filter.Page = page
		}
	}

	filter.Limit = 20
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	filter.SortBy = c.DefaultQuery("sort_by", "price_effective_from")
	filter.SortOrder = c.DefaultQuery("sort_order", "desc")

	catalogs, total, err := h.catalogService.ListProductCatalogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-error",
			Title:    "Internal Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to list catalog entries",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Build pagination response
	response := gin.H{
		"data": catalogs,
		"pagination": gin.H{
			"page":        filter.Page,
			"limit":       filter.Limit,
			"total":       total,
			"total_pages": (total + int64(filter.Limit) - 1) / int64(filter.Limit),
		},
	}

	c.JSON(http.StatusOK, response)
}

// UpdatePurchasePrice handles PUT /api/v1/supplier-product-catalogs/:id/price
// Story 10.5, AC1: Update purchase price with price history tracking
func (h *SupplierProductCatalogHandler) UpdatePurchasePrice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid catalog ID",
			Instance: c.Request.URL.Path,
		})
		return
	}

	var request services.UpdatePriceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request body",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Get authenticated user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "User not authenticated",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// Call service to update price
	if err := h.catalogService.UpdatePurchasePrice(c.Request.Context(), uint(id), &request, userID.(uint), ipAddress); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/not-found",
				Title:    "Not Found",
				Status:   http.StatusNotFound,
				Detail:   "Catalog entry not found",
				Instance: c.Request.URL.Path,
			})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Failed to update purchase price",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Get updated catalog entry
	catalog, err := h.catalogService.GetProductCatalogByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-error",
			Title:    "Internal Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to get updated catalog entry",
			Instance: c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, catalog)
}

// SetPreferredSupplier handles PUT /api/v1/supplier-product-catalogs/:id/preferred
// Story 10.5, AC1: Set or unset preferred supplier for a product
func (h *SupplierProductCatalogHandler) SetPreferredSupplier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid catalog ID",
			Instance: c.Request.URL.Path,
		})
		return
	}

	var request services.SetPreferredRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request body",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Get authenticated user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "User not authenticated",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// Call service to set preferred supplier
	if err := h.catalogService.SetPreferredSupplier(c.Request.Context(), uint(id), &request, userID.(uint), ipAddress); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/not-found",
				Title:    "Not Found",
				Status:   http.StatusNotFound,
				Detail:   "Catalog entry not found",
				Instance: c.Request.URL.Path,
			})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Failed to set preferred supplier",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Get updated catalog entry
	catalog, err := h.catalogService.GetProductCatalogByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-error",
			Title:    "Internal Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to get updated catalog entry",
			Instance: c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, catalog)
}

// GetPriceHistory handles GET /api/v1/products/:id/price-history
// Story 10.5, AC1: Get price history for a product
func (h *SupplierProductCatalogHandler) GetPriceHistory(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid product ID",
			Instance: c.Request.URL.Path,
		})
		return
	}

	filter := &services.PriceHistoryFilter{}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		filter.StartDate = &startDateStr
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		filter.EndDate = &endDateStr
	}

	if supplierIDStr := c.Query("supplier_id"); supplierIDStr != "" {
		supplierID, err := strconv.ParseUint(supplierIDStr, 10, 64)
		if err == nil {
			sid := uint(supplierID)
			filter.SupplierID = &sid
		}
	}

	// Pagination parameters
	filter.Page = 1
	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil && page > 0 {
			filter.Page = page
		}
	}

	filter.Limit = 20
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	history, err := h.catalogService.GetPriceHistory(c.Request.Context(), uint(productID), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-error",
			Title:    "Internal Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to get price history",
			Instance: c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": history})
}

// GetPreferredSupplier handles GET /api/v1/products/:id/preferred-supplier
// Story 10.5: Get preferred supplier for a product
func (h *SupplierProductCatalogHandler) GetPreferredSupplier(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid product ID",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Get branch ID from query parameter or user context
	var branchID uint
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		branchIDVal, err := strconv.ParseUint(branchIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/invalid-branch-id",
				Title:    "Invalid Branch ID",
				Status:   http.StatusBadRequest,
				Detail:   "Invalid branch ID",
				Instance: c.Request.URL.Path,
			})
			return
		}
		branchID = uint(branchIDVal)
	} else {
		// Get user's branch from JWT context
		if branchIDFromCtx, exists := c.Get("branch_id"); exists {
			switch v := branchIDFromCtx.(type) {
			case uint:
				branchID = v
			case int:
				branchID = uint(v)
			case int64:
				branchID = uint(v)
			case float64:
				branchID = uint(v)
			}
		}
	}

	if branchID == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/missing-branch",
			Title:    "Missing Branch",
			Status:   http.StatusBadRequest,
			Detail:   "Branch ID is required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	catalog, err := h.catalogService.GetPreferredSupplier(c.Request.Context(), uint(productID), branchID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Type:     "/errors/not-found",
				Title:    "Not Found",
				Status:   http.StatusNotFound,
				Detail:   "No preferred supplier found for this product",
				Instance: c.Request.URL.Path,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-error",
			Title:    "Internal Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to get preferred supplier",
			Instance: c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, catalog)
}

// GetSupplierCatalog handles GET /api/v1/suppliers/:id/product-catalog
// Story 10.5: Get supplier's product catalog
func (h *SupplierProductCatalogHandler) GetSupplierCatalog(c *gin.Context) {
	supplierIDStr := c.Param("id")
	supplierID, err := strconv.ParseUint(supplierIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/invalid-id",
			Title:    "Invalid ID",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid supplier ID",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Get branch ID from query parameter or user context
	var branchID uint
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		branchIDVal, err := strconv.ParseUint(branchIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Type:     "/errors/invalid-branch-id",
				Title:    "Invalid Branch ID",
				Status:   http.StatusBadRequest,
				Detail:   "Invalid branch ID",
				Instance: c.Request.URL.Path,
			})
			return
		}
		branchID = uint(branchIDVal)
	} else {
		// Get user's branch from JWT context
		if branchIDFromCtx, exists := c.Get("branch_id"); exists {
			switch v := branchIDFromCtx.(type) {
			case uint:
				branchID = v
			case int:
				branchID = uint(v)
			case int64:
				branchID = uint(v)
			case float64:
				branchID = uint(v)
			}
		}
	}

	if branchID == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/missing-branch",
			Title:    "Missing Branch",
			Status:   http.StatusBadRequest,
			Detail:   "Branch ID is required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	catalogs, err := h.catalogService.GetCatalogBySupplier(c.Request.Context(), uint(supplierID), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-error",
			Title:    "Internal Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to get supplier catalog",
			Instance: c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": catalogs})
}
