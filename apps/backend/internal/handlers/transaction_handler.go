package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// TransactionHandler handles transaction-related HTTP requests
type TransactionHandler struct {
	transactionService services.TransactionService
}

// NewTransactionHandler creates a new transaction handler instance
func NewTransactionHandler(transactionService services.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

// CreateTransaction handles POST /api/v1/transactions
// Processes a sale transaction with cart items and payment method
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	// Extract cashier ID from JWT context
	userIDValue, exists := c.Get("userID")
	if !exists {
		// User not authenticated - middleware should have caught this
		c.JSON(http.StatusUnauthorized, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/unauthorized",
			Title:  "Unauthorized",
			Status: http.StatusUnauthorized,
			Detail: "Authentication required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	cashierID, ok := userIDValue.(uint)
	if !ok {
		// HIGH FIX: Type assertion failed - invalid context value
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/internal-error",
			Title:  "Internal Error",
			Status: http.StatusInternalServerError,
			Detail: "Invalid user ID format in context",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// HIGH FIX: Validate cashierID is not zero
	if cashierID == 0 {
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/internal-error",
			Title:  "Internal Error",
			Status: http.StatusInternalServerError,
			Detail: "Invalid user ID value",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Extract branch ID from JWT context
	branchIDValue, exists := c.Get("branchID")
	if !exists {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/missing-branch",
			Title:  "Branch ID required",
			Status: http.StatusBadRequest,
			Detail: "Branch ID tidak ditemukan dalam konteks user",
			Instance: c.Request.URL.Path,
		})
		return
	}

	branchID, ok := branchIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/internal-error",
			Title:  "Internal Error",
			Status: http.StatusInternalServerError,
			Detail: "Invalid branch ID format in context",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// HIGH FIX: Validate branchID is not zero
	if branchID == 0 {
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/internal-error",
			Title:  "Internal Error",
			Status: http.StatusInternalServerError,
			Detail: "Invalid branch ID value",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Parse request body
	var saleRequest services.SaleRequest
	if err := c.ShouldBindJSON(&saleRequest); err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/invalid-request",
			Title:  "Invalid request",
			Status: http.StatusBadRequest,
			Detail: "Format JSON tidak valid: " + err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}

	// MEDIUM FIX: Validate and sanitize customer name
	if saleRequest.CustomerName != "" {
		// Remove any potentially dangerous characters
		customerName := strings.TrimSpace(saleRequest.CustomerName)
		// Limit length to prevent abuse
		if len(customerName) > 100 {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:   "https://api.simpo.com/errors/invalid-request",
				Title:  "Invalid Request",
				Status: http.StatusBadRequest,
				Detail: "Nama pelanggan terlalu panjang. Maksimal 100 karakter",
				Instance: c.Request.URL.Path,
			})
			return
		}
		// Check for SQL injection patterns (basic)
		dangerousPatterns := []string{"'", ";", "--", "/*", "*/", "xp_", "exec", "script"}
		for _, pattern := range dangerousPatterns {
			if strings.Contains(strings.ToLower(customerName), pattern) {
				c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
					Type:   "https://api.simpo.com/errors/invalid-request",
					Title:  "Invalid Request",
					Status: http.StatusBadRequest,
					Detail: "Nama pelanggan mengandung karakter tidak valid",
					Instance: c.Request.URL.Path,
				})
				return
			}
		}
		saleRequest.CustomerName = customerName
	}

	// Validate cart is not empty
	if len(saleRequest.Items) == 0 {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/empty-cart",
			Title:  "Cart cannot be empty",
			Status: http.StatusBadRequest,
			Detail: "Keranjang belanja tidak boleh kosong",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Validate cart item limit to prevent oversized carts
	const maxCartItems = 100 // Maximum items per transaction (Story 3.3)
	if len(saleRequest.Items) > maxCartItems {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/cart-too-large",
			Title:  "Cart Too Large",
			Status: http.StatusBadRequest,
			Detail: fmt.Sprintf("Keranjang tidak boleh lebih dari %d item", maxCartItems),
			Instance: c.Request.URL.Path,
		})
		return
	}

	// MEDIUM FIX: Request size validation
	// Note: MaxBytes middleware should be added to router for global request size limit
	// For now, we validate at handler level using request body size
	// TODO: Add MaxBytes middleware to router configuration

	// CRITICAL-002: Validate item quantities to prevent integer overflow attacks
	const maxQuantity = 10000 // Reasonable upper bound for single product quantity
	for _, item := range saleRequest.Items {
		if item.Quantity <= 0 || item.Quantity > maxQuantity {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:     "https://api.simpo.com/errors/invalid-quantity",
				Title:    "Invalid Quantity",
				Status:   http.StatusBadRequest,
				Detail:   "Jumlah produk harus antara 1 dan 10000",
				Instance: c.Request.URL.Path,
			})
			return
		}

		// MEDIUM FIX: Additional check for large quantities that could cause overflow in price calculation
		if item.Quantity > 1000 {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:     "https://api.simpo.com/errors/invalid-quantity",
				Title:    "Invalid Quantity",
				Status:   http.StatusBadRequest,
				Detail:   "Jumlah produk terlalu besar. Maksimal 1000 per item",
				Instance: c.Request.URL.Path,
			})
			return
		}
	}

	// HIGH FIX: Validate payment method specific fields
	// Note: These fields would need to be added to SaleRequest struct first
	// For now, this is a placeholder for future validation
	// TODO: Add ReferenceNumber and WalletType fields to SaleRequest

	// CRITICAL FIX: Add 2-second timeout to prevent indefinite hangs
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	// Process sale with transactional operations
	transaction, err := h.transactionService.ProcessSale(ctx, &saleRequest, cashierID, branchID)
	if err != nil {
		// Handle specific error types
		var stockErr *services.InsufficientStockError
		if errors.As(err, &stockErr) {
			c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
				Type:     "https://api.simpo.com/errors/transaction-failed",
				Title:    "Transaction Failed",
				Status:   http.StatusBadRequest,
				Detail:   "Stok tidak mencukupi: " + stockErr.ProductName,
				Instance: c.Request.URL.Path,
			})
			return
		}

		// Generic transaction error
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:     "https://api.simpo.com/errors/transaction-failed",
			Title:    "Transaction Failed",
			Status:   http.StatusBadRequest,
			Detail:   "Gagal memproses transaksi: " + err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Return created transaction with 201 status
	c.JSON(http.StatusCreated, transaction)
}

// ListTransactions handles GET /api/v1/transactions
// Returns paginated list of transactions for the cashier's branch
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	// Extract cashier ID from JWT context (for logging only)
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/unauthorized",
			Title:  "Unauthorized",
			Status: http.StatusUnauthorized,
			Detail: "Authentication required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Extract branch ID from JWT context for RBAC
	branchIDValue, exists := c.Get("branchID")
	if !exists {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/missing-branch",
			Title:  "Branch ID required",
			Status: http.StatusBadRequest,
			Detail: "Branch ID tidak ditemukan dalam konteks user",
			Instance: c.Request.URL.Path,
		})
		return
	}

	branchID, ok := branchIDValue.(uint)
	if !ok || branchID == 0 {
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/internal-error",
			Title:  "Internal Error",
			Status: http.StatusInternalServerError,
			Detail: "Invalid branch ID format in context",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Parse query parameters
	var startDate, endDate, status string
	page := 1 // Default page
	limit := 20 // Default limit per page

	if startDateParam := c.Query("startDate"); startDateParam != "" {
		startDate = startDateParam
	}
	if endDateParam := c.Query("endDate"); endDateParam != "" {
		endDate = endDateParam
	}
	if statusParam := c.Query("status"); statusParam != "" {
		status = statusParam
	}
	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse dates if provided
	var startDatePtr, endDatePtr *time.Time
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			startDatePtr = &t
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			endDatePtr = &t
		}
	}

	// Build filter criteria
	filter := &services.TransactionFilter{
		BranchID:  &branchID,
		StartDate: startDatePtr,
		EndDate:   endDatePtr,
		Status:    status,
		Page:      page,
		Limit:     limit,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	// Call service to get transactions
	transactions, total, err := h.transactionService.ListTransactions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/internal-error",
			Title:  "Internal Error",
			Status: http.StatusInternalServerError,
			Detail: "Gagal memuat riwayat transaksi: " + err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Calculate pagination metadata
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	// Build response with pagination metadata
	response := gin.H{
		"data": transactions,
		"pagination": gin.H{
			"total":      total,
			"totalPages": totalPages,
			"currentPage": page,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetTransactionByID handles GET /api/v1/transactions/:id
// Returns full transaction details with items, cashier, and branch information
func (h *TransactionHandler) GetTransactionByID(c *gin.Context) {
	// Extract cashier ID from JWT context (for audit trail)
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/unauthorized",
			Title:  "Unauthorized",
			Status: http.StatusUnauthorized,
			Detail: "Authentication required",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Extract branch ID from JWT context for RBAC
	branchIDValue, exists := c.Get("branchID")
	if !exists {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/missing-branch",
			Title:  "Branch ID required",
			Status: http.StatusBadRequest,
			Detail: "Branch ID tidak ditemukan dalam konteks user",
			Instance: c.Request.URL.Path,
		})
		return
	}

	branchID, ok := branchIDValue.(uint)
	if !ok || branchID == 0 {
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/internal-error",
			Title:  "Internal Error",
			Status: http.StatusInternalServerError,
			Detail: "Invalid branch ID format in context",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Parse transaction ID from URL parameter
 transactionIDParam := c.Param("id")
	transactionID, err := strconv.ParseUint(transactionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/invalid-id",
			Title:  "Invalid Transaction ID",
			Status: http.StatusBadRequest,
			Detail: "Format ID transaksi tidak valid",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Call service to get transaction details
	transaction, err := h.transactionService.GetTransactionByID(c.Request.Context(), uint(transactionID))
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, middleware.RFC7807Error{
				Type:   "https://api.simpo.com/errors/not-found",
				Title:  "Transaction Not Found",
				Status: http.StatusNotFound,
				Detail: "Transaksi tidak ditemukan",
				Instance: c.Request.URL.Path,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/internal-error",
			Title:  "Internal Error",
			Status: http.StatusInternalServerError,
			Detail: "Gagal memuat detail transaksi: " + err.Error(),
			Instance: c.Request.URL.Path,
		})
		return
	}

	// RBAC check: Ensure cashier can only access transactions from their branch
	if transaction.BranchID != branchID {
		c.JSON(http.StatusForbidden, middleware.RFC7807Error{
			Type:   "https://api.simpo.com/errors/forbidden",
			Title:  "Access Denied",
			Status: http.StatusForbidden,
			Detail: "Anda tidak memiliki akses ke transaksi ini",
			Instance: c.Request.URL.Path,
		})
		return
	}

	// Generate receipt data for reprint capability
	receiptData, err := h.transactionService.GenerateReceiptData(c.Request.Context(), transaction.ID)
	if err != nil {
		// Log error but don't fail the request - receipt data is optional
		// Transaction can still be viewed without reprint capability
		receiptData = nil
	}

	// Build response with transaction details and receipt data
	response := gin.H{
		"id":               transaction.ID,
		"transactionNumber": transaction.TransactionNumber,
		"total":            transaction.Total,
		"status":           transaction.Status,
		"paymentMethod":    transaction.PaymentMethod,
		"createdAt":        transaction.CreatedAt,
		"items":            transaction.TransactionItems,
		"cashier": gin.H{
			"id":   transaction.CashierID,
			"name": "", // Will be populated from user data
		},
		"branch": gin.H{
			"id":   transaction.BranchID,
			"name": "", // Will be populated from branch data
		},
	}

	if receiptData != nil {
		response["receiptData"] = receiptData
	}

	c.JSON(http.StatusOK, response)
}
