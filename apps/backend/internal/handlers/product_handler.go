package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// ProductHandler defines product handler interface
// Story 4.1, Task 1: Handler interface for product operations
type ProductHandler interface {
	ListProducts(c *gin.Context)
}

// productHandler implements ProductHandler
type productHandler struct {
	productService services.ProductService
}

// NewProductHandler creates a new product handler
// Story 4.1, Task 1: Constructor with service dependency injection
func NewProductHandler(productService services.ProductService) ProductHandler {
	if productService == nil {
		panic("productService cannot be nil")
	}
	return &productHandler{
		productService: productService,
	}
}

// ListProducts handles product listing with search, filters, and pagination
// Story 4.1, AC1, AC2, AC3, AC4, AC7: Product list with search, filters, and pagination
//
//	@Summary		List products
//	@Description	Get products with search, filters, and pagination. Owners can filter by branch/category. Cashiers see only their branch.
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			search	query		string	false	"Search by name or SKU"
//	@Param			category	query		string	false	"Filter by category"
//	@Param			branch_id	query		int		false	"Filter by branch (Owner only)"
//	@Param			low_stock	query		bool	false	"Filter for low stock items"
//	@Param			expired	query		bool	false	"Filter for expired items"
//	@Param			page		query		int		false	"Page number (default 1)"
//	@Param			limit		query		int		false	"Items per page (default 20, max 1000)"
//	@Param			sort_by		query		string	false	"Field to sort by"	Enums(id, name, sku, price, stock_qty, category, created_at)
//	@Param			sort_order	query		string	false	"Sort order"	Enums(asc, desc)
//	@Success		200			{object}	apiErrors.Response{success=bool,data=dto.ProductListResponse}	"Success response with product list"
//	@Failure		400			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Validation error - invalid input parameters"
//	@Failure		401			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Unauthorized - authentication required"
//	@Failure		403			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Forbidden - insufficient permissions"
//	@Failure		500			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Server error"
//	@Router			/api/v1/products [get]
func (h *productHandler) ListProducts(c *gin.Context) {
	// Story 4.1, Task 1.3: Bind and validate query parameters
	var req dto.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(errors.FromGinValidation(err))
		c.Status(http.StatusBadRequest)
		return
	}

	// Apply defaults for pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20 // Default page size
	}
	if req.Limit > 1000 {
		req.Limit = 1000 // Maximum to prevent DoS
	}

	// Story 4.1, Task 5.1, 5.2, 5.3: Extract user context for RBAC
	userRole, exists := c.Get("user_role")
	if !exists {
		_ = c.Error(errors.Unauthorized("User role not found"))
		c.Status(http.StatusUnauthorized)
		return
	}

	role, ok := userRole.(string)
	if !ok {
		_ = c.Error(errors.InternalServerError(fmt.Errorf("invalid user role type")))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Story 4.1, AC3: Apply branch access control
	// Owners can filter by branch, Cashiers restricted to their assigned branch
	var userBranchID *uint
	branchIDValue, branchExists := c.Get("branch_id")
	if branchExists {
		if bid, ok := branchIDValue.(uint); ok {
			userBranchID = &bid
		}
	}

	// Story 4.1, Task 5.2, 5.3: RBAC - Branch access control
	// Owners can view all branches or filter by branch_id parameter
	// Cashiers can only view their assigned branch
	if role == user.RoleCashier {
		// Story 4.1 Code Review (2026-05-18): Ensure cashier has branch assignment
		if userBranchID == nil {
			_ = c.Error(errors.Forbidden("Cashier must have a branch assignment"))
			c.Status(http.StatusForbidden)
			return
		}
		// Cashier: Override branch_id filter with their assigned branch
		if req.BranchID != nil && *req.BranchID != *userBranchID {
			_ = c.Error(errors.Forbidden("Cashiers can only view products from their assigned branch"))
			c.Status(http.StatusForbidden)
			return
		}
		req.BranchID = userBranchID
	}
	// For Owner: use the branch_id from query parameter if provided, otherwise nil (all branches)

	// Story 4.1 Code Review (2026-05-18): Validate SortBy field to prevent SQL injection
	allowedSortFields := map[string]bool{
		"id": true, "name": true, "sku": true, "price": true,
		"stock_qty": true, "category": true, "created_at": true,
	}
	if req.SortBy != "" && !allowedSortFields[req.SortBy] {
		req.SortBy = "created_at" // Default fallback
	}

	// Build service filter from request
	filter := &services.ProductFilter{
		BranchID:     req.BranchID,
		Category:     req.Category,
		SearchQuery:  req.Search,
		LowStock:     req.LowStock != nil && *req.LowStock,
		Expired:      req.Expired != nil && *req.Expired,
		Page:         req.Page,
		Limit:        req.Limit,
		SortBy:       req.SortBy,
		SortOrder:    req.SortOrder,
	}

	// Call service layer
	products, total, err := h.productService.ListProducts(c.Request.Context(), filter)
	if err != nil {
		_ = c.Error(errors.InternalServerError(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Story 4.1, Task 4: Transform to DTO with indicators
	// Story 4.1, AC5: Calculate low stock indicator
	// Story 4.1, AC6: Calculate expired indicator
	productItems := make([]dto.ProductListItem, 0, len(products))
	now := time.Now()
	for _, p := range products {
		isLowStock := p.StockQty < int64(p.ReorderThreshold)
		isExpired := p.ExpiryDate != nil && p.ExpiryDate.Before(now)

		productItems = append(productItems, dto.ProductListItem{
			ID:               p.ID,
			SKU:              p.SKU,
			Name:             p.Name,
			Description:      p.Description,
			StockQty:         p.StockQty,
			Price:            p.Price,
			ExpiryDate:       p.ExpiryDate,
			BranchID:         p.BranchID,
			Category:         p.Category,
			ReorderThreshold: p.ReorderThreshold,
			IsLowStock:       isLowStock,
			IsExpired:        isExpired,
			CreatedAt:        p.CreatedAt,
			UpdatedAt:        p.UpdatedAt,
		})
	}

	// Calculate total pages
	totalPages := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPages++
	}

	// Build response
	response := dto.ProductListResponse{
		Data: productItems,
		Pagination: dto.PaginationMetadata{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, errors.Success(response))
}
