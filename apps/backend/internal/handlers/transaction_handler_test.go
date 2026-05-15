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
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// MockTransactionService is a mock for testing
type MockTransactionService struct {
	processSaleFunc func(ctx context.Context, sale *services.SaleRequest, cashierID, branchID uint) (*models.Transaction, error)
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

func (m *MockTransactionService) GenerateReceiptData(ctx context.Context, transactionID uint) (*services.ReceiptData, error) {
	return nil, nil
}

func (m *MockTransactionService) GetTransactionByID(ctx context.Context, id uint) (*models.Transaction, error) {
	return nil, nil
}

func (m *MockTransactionService) ListTransactions(ctx context.Context, filter *services.TransactionFilter) ([]*models.Transaction, int64, error) {
	return nil, 0, nil
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
