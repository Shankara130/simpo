package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// MockTransactionService is a mock for testing
type MockTransactionService struct {
	processSaleFunc         func(ctx context.Context, sale *services.SaleRequest, cashierID, branchID uint) (*models.Transaction, error)
	ListTransactionsFunc    func(ctx context.Context, filter *services.TransactionFilter) ([]*models.Transaction, int64, error)
	GetTransactionByIDFunc  func(ctx context.Context, id uint) (*models.Transaction, error)
	GenerateReceiptDataFunc func(ctx context.Context, transactionID uint) (*services.ReceiptData, error)
}

func (m *MockTransactionService) ProcessSale(ctx context.Context, sale *services.SaleRequest, cashierID, branchID uint) (*models.Transaction, error) {
	if m.processSaleFunc != nil {
		return m.processSaleFunc(ctx, sale, cashierID, branchID)
	}
	return nil, errors.New("ProcessSale not implemented")
}

func (m *MockTransactionService) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	return nil
}

func (m *MockTransactionService) CalculateTotal(items []*services.SaleItem) (string, error) {
	return "0", nil
}

func (m *MockTransactionService) GetTransactionByID(ctx context.Context, id uint) (*models.Transaction, error) {
	if m.GetTransactionByIDFunc != nil {
		return m.GetTransactionByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockTransactionService) ListTransactions(ctx context.Context, filter *services.TransactionFilter) ([]*models.Transaction, int64, error) {
	if m.ListTransactionsFunc != nil {
		return m.ListTransactionsFunc(ctx, filter)
	}
	return nil, 0, nil
}

func (m *MockTransactionService) GenerateReceiptData(ctx context.Context, transactionID uint) (*services.ReceiptData, error) {
	if m.GenerateReceiptDataFunc != nil {
		return m.GenerateReceiptDataFunc(ctx, transactionID)
	}
	return nil, nil
}

// SetupTestRouter creates a test router with the transaction handler and optional auth middleware
func setupTestRouter(transactionService services.TransactionService, withAuth bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add auth middleware if needed
	if withAuth {
		router.Use(func(c *gin.Context) {
			c.Set("userID", uint(100))
			c.Set("branchID", uint(1))
			c.Next()
		})
	}

	// Add transaction handler
	handler := NewTransactionHandler(transactionService)
	router.POST("/api/v1/transactions", handler.CreateTransaction)
	// Story 3.7: Transaction history and detail endpoints
	router.GET("/api/v1/transactions", handler.ListTransactions)
	router.GET("/api/v1/transactions/:id", handler.GetTransactionByID)

	return router
}

// TestTransactionHandler_CreateTransaction_Success tests successful transaction creation
func TestTransactionHandler_CreateTransaction_Success(t *testing.T) {
	// Arrange
	expectedTransaction := &models.Transaction{
		ID:                1,
		TransactionNumber: "TRX-20260515-0001",
		CashierID:         100,
		BranchID:          1,
		Total:             "150000.00",
		Status:            models.StatusCompleted,
	}

	mockService := &MockTransactionService{
		processSaleFunc: func(ctx context.Context, sale *services.SaleRequest, cashierID, branchID uint) (*models.Transaction, error) {
			return expectedTransaction, nil
		},
	}

	router := setupTestRouter(mockService, true)

	// Create request body
	requestBody := services.SaleRequest{
		Items: []*services.SaleItem{
			{
				ProductID: 1,
				Quantity:  2,
				UnitPrice: "75000.00",
			},
		},
		PaymentMethod:  "CASH",
		TaxAmount:      "0",
		DiscountAmount: "0",
	}
	bodyJSON, _ := json.Marshal(requestBody)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedTransaction.ID, response.ID)
	assert.Equal(t, expectedTransaction.TransactionNumber, response.TransactionNumber)
	assert.Equal(t, expectedTransaction.CashierID, response.CashierID)
	assert.Equal(t, expectedTransaction.Total, response.Total)
	assert.Equal(t, expectedTransaction.Status, response.Status)
}

// TestTransactionHandler_CreateTransaction_EmptyCart tests validation error for empty cart
func TestTransactionHandler_CreateTransaction_EmptyCart(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{
		processSaleFunc: func(ctx context.Context, sale *services.SaleRequest, cashierID, branchID uint) (*models.Transaction, error) {
			t.Fatal("ProcessSale should not be called for empty cart")
			return nil, nil
		},
	}

	router := setupTestRouter(mockService, true)

	// Create request body with empty items
	requestBody := services.SaleRequest{
		Items:          []*services.SaleItem{},
		PaymentMethod:  "CASH",
		TaxAmount:      "0",
		DiscountAmount: "0",
	}
	bodyJSON, _ := json.Marshal(requestBody)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/empty-cart", response.Type)
	assert.Equal(t, "Cart cannot be empty", response.Title)
	assert.Equal(t, http.StatusBadRequest, response.Status)
	assert.Equal(t, "/api/v1/transactions", response.Instance)
}

// TestTransactionHandler_CreateTransaction_InsufficientStock tests insufficient stock error
func TestTransactionHandler_CreateTransaction_InsufficientStock(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{
		processSaleFunc: func(ctx context.Context, sale *services.SaleRequest, cashierID, branchID uint) (*models.Transaction, error) {
			return nil, &services.InsufficientStockError{
				ProductID:    1,
				ProductName:  "Paracetamol 500mg",
				RequestedQty: 10,
				AvailableQty: 5,
			}
		},
	}

	router := setupTestRouter(mockService, true)

	// Create request body
	requestBody := services.SaleRequest{
		Items: []*services.SaleItem{
			{
				ProductID:  1,
				Quantity:   10,
				UnitPrice:  "75000.00",
			},
		},
		PaymentMethod:  "CASH",
		TaxAmount:      "0",
		DiscountAmount: "0",
	}
	bodyJSON, _ := json.Marshal(requestBody)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/transaction-failed", response.Type)
	assert.Contains(t, response.Detail, "Stok tidak mencukupi")
	assert.Equal(t, http.StatusBadRequest, response.Status)
}

// TestTransactionHandler_CreateTransaction_Unauthorized tests missing user context
func TestTransactionHandler_CreateTransaction_Unauthorized(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{}

	router := setupTestRouter(mockService, false) // No auth middleware

	// Create request body
	requestBody := services.SaleRequest{
		Items: []*services.SaleItem{
			{
				ProductID:  1,
				Quantity:   1,
				UnitPrice:  "75000.00",
			},
		},
		PaymentMethod:  "CASH",
		TaxAmount:      "0",
		DiscountAmount: "0",
	}
	bodyJSON, _ := json.Marshal(requestBody)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/unauthorized", response.Type)
	assert.Equal(t, "Unauthorized", response.Title)
}

// TestTransactionHandler_CreateTransaction_InvalidJSON tests invalid JSON
func TestTransactionHandler_CreateTransaction_InvalidJSON(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{}

	router := setupTestRouter(mockService, true)

	// Create invalid JSON
	invalidJSON := []byte(`{"invalid": json}`)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/invalid-request", response.Type)
	assert.Equal(t, "Invalid request", response.Title)
}

// TestTransactionHandler_CreateTransaction_MissingBranchID tests missing branch ID
func TestTransactionHandler_CreateTransaction_MissingBranchID(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add auth middleware with only userID (no branchID)
	router.Use(func(c *gin.Context) {
		c.Set("userID", uint(100))
		// Note: branchID not set
		c.Next()
	})

	// Add transaction handler
	handler := NewTransactionHandler(mockService)
	router.POST("/api/v1/transactions", handler.CreateTransaction)

	// Create request body
	requestBody := services.SaleRequest{
		Items: []*services.SaleItem{
			{
				ProductID:  1,
				Quantity:   1,
				UnitPrice:  "75000.00",
			},
		},
		PaymentMethod:  "CASH",
		TaxAmount:      "0",
		DiscountAmount: "0",
	}
	bodyJSON, _ := json.Marshal(requestBody)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/missing-branch", response.Type)
	assert.Equal(t, "Branch ID required", response.Title)
}

// TestNewTransactionHandler tests constructor
func TestNewTransactionHandler(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{}

	// Act
	handler := NewTransactionHandler(mockService)

	// Assert
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.transactionService)
}

// TestTransactionHandler_CreateTransaction_ServiceError tests generic service error
func TestTransactionHandler_CreateTransaction_ServiceError(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{
		processSaleFunc: func(ctx context.Context, sale *services.SaleRequest, cashierID, branchID uint) (*models.Transaction, error) {
			return nil, fmt.Errorf("database connection failed")
		},
	}

	router := setupTestRouter(mockService, true)

	// Create request body
	requestBody := services.SaleRequest{
		Items: []*services.SaleItem{
			{
				ProductID:  1,
				Quantity:   1,
				UnitPrice:  "75000.00",
			},
		},
		PaymentMethod:  "CASH",
		TaxAmount:      "0",
		DiscountAmount: "0",
	}
	bodyJSON, _ := json.Marshal(requestBody)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/transaction-failed", response.Type)
	assert.Contains(t, response.Detail, "Gagal memproses transaksi")
}

// Story 3.7: Transaction History Tests

// TestTransactionHandler_ListTransactions_Success tests successful transaction list retrieval
func TestTransactionHandler_ListTransactions_Success(t *testing.T) {
	// Arrange
	expectedTransactions := []*models.Transaction{
		{
			ID:                1,
			TransactionNumber: "TRX-20260515-0001",
			CashierID:         100,
			BranchID:          1,
			Total:             "150000.00",
			Status:            models.StatusCompleted,
		},
		{
			ID:                2,
			TransactionNumber: "TRX-20260515-0002",
			CashierID:         100,
			BranchID:          1,
			Total:             "75000.00",
			Status:            models.StatusCompleted,
		},
	}

	mockService := &MockTransactionService{
		ListTransactionsFunc: func(ctx context.Context, filter *services.TransactionFilter) ([]*models.Transaction, int64, error) {
			return expectedTransactions, 2, nil
		},
	}

	router := setupTestRouter(mockService, true)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/transactions?page=1&limit=20", nil)
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check pagination metadata
	pagination, ok := response["pagination"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(2), pagination["total"])
	assert.Equal(t, float64(1), pagination["currentPage"])

	// Check data array
	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)
}

// TestTransactionHandler_ListTransactions_WithFilters tests transaction list with date and status filters
func TestTransactionHandler_ListTransactions_WithFilters(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{
		ListTransactionsFunc: func(ctx context.Context, filter *services.TransactionFilter) ([]*models.Transaction, int64, error) {
			// Verify filter parameters
			assert.Equal(t, uint(1), *filter.BranchID)
			assert.Equal(t, "COMPLETED", filter.Status)
			assert.NotNil(t, filter.StartDate)
			assert.NotNil(t, filter.EndDate)
			assert.Equal(t, 1, filter.Page)
			assert.Equal(t, 20, filter.Limit)

			return []*models.Transaction{}, 0, nil
		},
	}

	router := setupTestRouter(mockService, true)

	// Create request with filters
	req, _ := http.NewRequest("GET", "/api/v1/transactions?startDate=2026-05-01&endDate=2026-05-15&status=COMPLETED&page=1&limit=20", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestTransactionHandler_ListTransactions_InvalidPage tests invalid page parameter
func TestTransactionHandler_ListTransactions_InvalidPage(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{
		ListTransactionsFunc: func(ctx context.Context, filter *services.TransactionFilter) ([]*models.Transaction, int64, error) {
			// Default to page 1 when invalid page provided
			assert.Equal(t, 1, filter.Page)
			return []*models.Transaction{}, 0, nil
		},
	}

	router := setupTestRouter(mockService, true)

	// Create request with invalid page
	req, _ := http.NewRequest("GET", "/api/v1/transactions?page=invalid", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestTransactionHandler_ListTransactions_Unauthorized tests missing authentication
func TestTransactionHandler_ListTransactions_Unauthorized(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{}
	router := setupTestRouter(mockService, false) // No auth middleware

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/unauthorized", response.Type)
}

// TestTransactionHandler_ListTransactions_ServiceError tests service error handling
func TestTransactionHandler_ListTransactions_ServiceError(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{
		ListTransactionsFunc: func(ctx context.Context, filter *services.TransactionFilter) ([]*models.Transaction, int64, error) {
			return nil, 0, fmt.Errorf("database connection failed")
		},
	}

	router := setupTestRouter(mockService, true)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/internal-error", response.Type)
	assert.Contains(t, response.Detail, "Gagal memuat riwayat transaksi")
}

// TestTransactionHandler_GetTransactionByID_Success tests successful transaction detail retrieval
func TestTransactionHandler_GetTransactionByID_Success(t *testing.T) {
	// Arrange
	expectedTransaction := &models.Transaction{
		ID:                1,
		TransactionNumber: "TRX-20260515-0001",
		CashierID:         100,
		BranchID:          1, // Same as authenticated user's branch
		Total:             "150000.00",
		Status:            models.StatusCompleted,
		TransactionItems: []models.TransactionItem{
			{
				ID:            1,
				TransactionID: 1,
				ProductID:     1,
				ProductName:   "Paracetamol 500mg",
				Quantity:      2,
				UnitPrice:     "75000.00",
				Subtotal:      "150000.00",
			},
		},
	}

	mockService := &MockTransactionService{
		GetTransactionByIDFunc: func(ctx context.Context, id uint) (*models.Transaction, error) {
			return expectedTransaction, nil
		},
		GenerateReceiptDataFunc: func(ctx context.Context, transactionID uint) (*services.ReceiptData, error) {
			return &services.ReceiptData{}, nil
		},
	}

	router := setupTestRouter(mockService, true)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/transactions/1", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(1), response["id"])
	assert.Equal(t, "TRX-20260515-0001", response["transactionNumber"])
}

// TestTransactionHandler_GetTransactionByID_NotFound tests transaction not found error
func TestTransactionHandler_GetTransactionByID_NotFound(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{
		GetTransactionByIDFunc: func(ctx context.Context, id uint) (*models.Transaction, error) {
			return nil, repositories.ErrNotFound
		},
	}

	router := setupTestRouter(mockService, true)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/transactions/999", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/not-found", response.Type)
	assert.Equal(t, "Transaction Not Found", response.Title)
}

// TestTransactionHandler_GetTransactionByID_Forbidden tests RBAC enforcement
func TestTransactionHandler_GetTransactionByID_Forbidden(t *testing.T) {
	// Arrange
	otherBranchTransaction := &models.Transaction{
		ID:                1,
		TransactionNumber: "TRX-20260515-0001",
		CashierID:         100,
		BranchID:          2, // Different branch from authenticated user (branch 1)
		Total:             "150000.00",
		Status:            models.StatusCompleted,
	}

	mockService := &MockTransactionService{
		GetTransactionByIDFunc: func(ctx context.Context, id uint) (*models.Transaction, error) {
			return otherBranchTransaction, nil
		},
	}

	router := setupTestRouter(mockService, true)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/transactions/1", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/forbidden", response.Type)
	assert.Equal(t, "Access Denied", response.Title)
	assert.Contains(t, response.Detail, "Anda tidak memiliki akses")
}

// TestTransactionHandler_GetTransactionByID_InvalidID tests invalid transaction ID
func TestTransactionHandler_GetTransactionByID_InvalidID(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{}
	router := setupTestRouter(mockService, true)

	// Create request with invalid ID
	req, _ := http.NewRequest("GET", "/api/v1/transactions/invalid", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/invalid-id", response.Type)
	assert.Equal(t, "Invalid Transaction ID", response.Title)
}

// TestTransactionHandler_GetTransactionByID_Unauthorized tests missing authentication
func TestTransactionHandler_GetTransactionByID_Unauthorized(t *testing.T) {
	// Arrange
	mockService := &MockTransactionService{}
	router := setupTestRouter(mockService, false) // No auth middleware

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/transactions/1", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response middleware.RFC7807Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "https://api.simpo.com/errors/unauthorized", response.Type)
}
