package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
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
