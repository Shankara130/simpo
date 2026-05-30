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

// GoodsReceiptHandler handles goods receipt management operations
// Story 10.3: Handler for goods receipt processing operations
type GoodsReceiptHandler struct {
	goodsReceiptService services.GoodsReceiptService
}

// NewGoodsReceiptHandler creates a new goods receipt handler
// Story 10.3: Factory function with dependency injection
func NewGoodsReceiptHandler(goodsReceiptService services.GoodsReceiptService) *GoodsReceiptHandler {
	return &GoodsReceiptHandler{
		goodsReceiptService: goodsReceiptService,
	}
}

// ProcessGoodsReceipt godoc
//
//	@Summary		Process goods receipt for a purchase invoice
//	@Description	Processes goods receipt for a purchase invoice, updates stock quantities and cost prices (Story 10.3, AC1)
//	@Tags			Goods Receipt Management
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.ProcessGoodsReceiptRequest	true	"Goods receipt processing request"
//	@Security		BearerAuth
//	@Success		201	{object}	dto.GoodsReceiptResponse	"Goods receipt processed successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin or Owner only"
//	@Failure		404	{object}	dto.ErrorResponse	"Invoice not found"
//	@Failure		409	{object}	dto.ErrorResponse	"Invoice already received"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/goods-receipts/process [post]
func (h *GoodsReceiptHandler) ProcessGoodsReceipt(c *gin.Context) {
	// Parse request
	var req dto.ProcessGoodsReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid request format: " + err.Error(),
			Instance: "/api/v1/goods-receipts/process",
		})
		return
	}

	// Extract user context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "User ID not found in context",
			Instance: "/api/v1/goods-receipts/process",
		})
		return
	}

	receivedBy, ok := userID.(uint)
	if !ok || receivedBy == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Invalid user ID in context",
			Instance: "/api/v1/goods-receipts/process",
		})
		return
	}

	// Extract branch ID from context
	branchID, exists := c.Get("branchID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Branch ID not found in context",
			Instance: "/api/v1/goods-receipts/process",
		})
		return
	}

	userBranchID, ok := branchID.(uint)
	// Code review fix: MEDIUM-004 - Add branch ID validation (check > 0)
	if !ok || userBranchID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Type:     "/errors/unauthorized",
			Title:    "Unauthorized",
			Status:   http.StatusUnauthorized,
			Detail:   "Invalid branch ID in context",
			Instance: "/api/v1/goods-receipts/process",
		})
		return
	}

	// Process goods receipt
	goodsReceipt, err := h.goodsReceiptService.ProcessGoodsReceipt(c.Request.Context(), req.InvoiceID, receivedBy, req.Notes, userBranchID)
	if err != nil {
		// Determine appropriate status code based on error
		statusCode := http.StatusInternalServerError
		errorType := "/errors/internal-server-error"

		// Code review fix: LOW-001, LOW-003 - Use strings.Contains/HasPrefix instead of fragile string parsing
		// Check for specific error types
		if strings.Contains(err.Error(), "invoice not found") {
			statusCode = http.StatusNotFound
			errorType = "/errors/not-found"
		} else if strings.HasPrefix(err.Error(), "invoice has already been received") {
			statusCode = http.StatusConflict
			errorType = "/errors/conflict"
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Type:     errorType,
			Title:    http.StatusText(statusCode),
			Status:   statusCode,
			Detail:   err.Error(),
			Instance: "/api/v1/goods-receipts/process",
		})
		return
	}

	// Convert model to response DTO
	response := h.convertToGoodsReceiptResponse(goodsReceipt)

	c.JSON(http.StatusCreated, response)
}

// GetGoodsReceipt godoc
//
//	@Summary		Get goods receipt by ID
//	@Description	Retrieves a goods receipt by its ID with full details (Story 10.3)
//	@Tags			Goods Receipt Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Goods Receipt ID"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GoodsReceiptResponse	"Goods receipt retrieved successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid ID"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin or Owner only"
//	@Failure		404	{object}	dto.ErrorResponse	"Goods receipt not found"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/goods-receipts/{id} [get]
func (h *GoodsReceiptHandler) GetGoodsReceipt(c *gin.Context) {
	// Parse ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid ID format: " + err.Error(),
			Instance: "/api/v1/goods-receipts/" + idParam,
		})
		return
	}

	// Get goods receipt
	goodsReceipt, err := h.goodsReceiptService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Type:     "/errors/not-found",
			Title:    "Not Found",
			Status:   http.StatusNotFound,
			Detail:   "Goods receipt not found",
			Instance: "/api/v1/goods-receipts/" + idParam,
		})
		return
	}

	// Convert model to response DTO
	response := h.convertToGoodsReceiptResponse(goodsReceipt)

	c.JSON(http.StatusOK, response)
}

// ListGoodsReceipts godoc
//
//	@Summary		List goods receipts
//	@Description	Lists goods receipts with optional filtering and pagination (Story 10.3)
//	@Tags			Goods Receipt Management
//	@Accept			json
//	@Produce		json
//	@Param			branchId	query	int	false	"Filter by branch"
//	@Param			startDate	query	string	false	"Filter by received date range start (YYYY-MM-DD)"
//	@Param			endDate	query	string	false	"Filter by received date range end (YYYY-MM-DD)"
//	@Param			receivedBy	query	int	false	"Filter by user who processed the receipt"
//	@Param			page	query	int	false	"Page number (default: 1)"
//	@Param			limit	query	int	false	"Items per page (default: 20)"
//	@Param			sortBy	query	string	false	"Field to sort by (default: received_date)"
//	@Param			sortOrder	query	string	false	"Sort order (asc or desc, default: desc)"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GoodsReceiptListResponse	"Goods receipts listed successfully"
//	@Failure		400	{object}	dto.ErrorResponse	"Invalid filter parameters"
//	@Failure		401	{object}	dto.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	dto.ErrorResponse	"Forbidden - Admin or Owner only"
//	@Failure		500	{object}	dto.ErrorResponse	"Internal server error"
//	@Router			/api/v1/goods-receipts [get]
func (h *GoodsReceiptHandler) ListGoodsReceipts(c *gin.Context) {
	// Parse filter parameters
	var filter dto.GoodsReceiptFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Type:     "/errors/validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid filter parameters: " + err.Error(),
			Instance: "/api/v1/goods-receipts",
		})
		return
	}

	// Code review fix: HIGH-003 - Set pagination defaults to prevent division by zero
	// Convert DTO filter to service filter
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	serviceFilter := &services.GoodsReceiptFilter{
		BranchID:   filter.BranchID,
		StartDate:  filter.StartDate,
		EndDate:    filter.EndDate,
		ReceivedBy: filter.ReceivedBy,
		Page:       filter.Page,
		Limit:      filter.Limit,
		SortBy:     filter.SortBy,
		SortOrder:  filter.SortOrder,
	}

	// List goods receipts
	receipts, total, err := h.goodsReceiptService.List(c.Request.Context(), serviceFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Type:     "/errors/internal-server-error",
			Title:    "Internal Server Error",
			Status:   http.StatusInternalServerError,
			Detail:   "Failed to list goods receipts: " + err.Error(),
			Instance: "/api/v1/goods-receipts",
		})
		return
	}

	// Convert models to response DTOs
	data := make([]*dto.GoodsReceiptResponse, len(receipts))
	for i, receipt := range receipts {
		data[i] = h.convertToGoodsReceiptResponse(receipt)
	}

	// Calculate total pages
	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}

	response := dto.GoodsReceiptListResponse{
		Data: data,
		Meta: dto.PaginationMeta{
			Page:       filter.Page,
			Limit:      filter.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// convertToGoodsReceiptResponse converts a goods receipt model to response DTO
func (h *GoodsReceiptHandler) convertToGoodsReceiptResponse(receipt *models.GoodsReceipt) *dto.GoodsReceiptResponse {
	response := &dto.GoodsReceiptResponse{
		ID:                receipt.ID,
		PurchaseInvoiceID: receipt.PurchaseInvoiceID,
		ReceivedDate:      receipt.ReceivedDate.Format("2006-01-02"),
		ReceivedBy:        receipt.ReceivedBy,
		Notes:             receipt.Notes,
		BranchID:          receipt.BranchID,
		CreatedAt:         receipt.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         receipt.UpdatedAt.Format(time.RFC3339),
	}

	// Convert purchase invoice if available
	// Code review fix: MEDIUM-003 - Check if PurchaseInvoice was eager loaded (ID non-zero)
	// Note: PurchaseInvoice is a struct, not a pointer, so we check ID field directly
	if receipt.PurchaseInvoice.ID != 0 {
		response.PurchaseInvoice = h.convertToPurchaseInvoiceSummary(&receipt.PurchaseInvoice)
	}

	return response
}

// convertToPurchaseInvoiceSummary converts a purchase invoice model to summary DTO
func (h *GoodsReceiptHandler) convertToPurchaseInvoiceSummary(invoice *models.PurchaseInvoice) *dto.PurchaseInvoiceSummary {
	summary := &dto.PurchaseInvoiceSummary{
		ID:            invoice.ID,
		InvoiceNumber: invoice.InvoiceNumber,
		InvoiceDate:   invoice.InvoiceDate.Format("2006-01-02"),
		SupplierID:    invoice.SupplierID,
		TotalAmount:   invoice.TotalAmount,
		PaymentStatus: invoice.PaymentStatus,
		ReceiptStatus: invoice.ReceiptStatus,
	}

	// Add supplier name if available
	if invoice.Supplier.ID != 0 {
		summary.SupplierName = invoice.Supplier.Name
	}

	// Convert items
	summary.Items = make([]dto.PurchaseInvoiceItemSummary, len(invoice.Items))
	for i, item := range invoice.Items {
		summary.Items[i] = dto.PurchaseInvoiceItemSummary{
			ID:     item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitCost:  item.UnitCost,
			Subtotal: item.Subtotal,
		}

		// Add product details if available
		if item.Product.ID != 0 {
			summary.Items[i].ProductName = item.Product.Name
			summary.Items[i].SKU = item.Product.SKU
		}
	}

	return summary
}
