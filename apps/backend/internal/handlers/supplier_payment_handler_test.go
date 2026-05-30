package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// setupSupplierPaymentTestRouter creates a test router with the supplier payment handler and auth context
// Story 10.4: Test helper for supplier payment handler testing
func setupSupplierPaymentTestRouter(supplierPaymentService services.SupplierPaymentService, userRole string, branchID uint, userID uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add auth context middleware to simulate authenticated user
	// Keys must match what the handler expects: "user_id", "username", "user_role", "branch_id"
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("username", "testuser")
		c.Set("user_role", userRole)
		c.Set("branch_id", branchID)
		c.Next()
	})

	// Add supplier payment handler
	handler := NewSupplierPaymentHandler(supplierPaymentService)
	router.POST("/api/v1/supplier-payments", handler.RecordPayment)
	router.GET("/api/v1/supplier-payments/:id", handler.GetSupplierPayment)
	router.GET("/api/v1/supplier-payments", handler.ListSupplierPayments)
	router.GET("/api/v1/suppliers/:id/payment-history", handler.GetPaymentHistoryBySupplier)

	return router
}

// MockSupplierPaymentService is a mock for testing
type MockSupplierPaymentService struct {
	recordPaymentFunc            func(ctx context.Context, request *services.RecordPaymentRequest, createdBy uint, ipAddress string) (*models.SupplierPayment, error)
	getByIDFunc                   func(ctx context.Context, id uint) (*models.SupplierPayment, error)
	listFunc                     func(ctx context.Context, filter *services.SupplierPaymentListFilter) ([]*models.SupplierPayment, int64, error)
	getPaymentHistoryBySupplierFunc func(ctx context.Context, supplierID uint, filter *services.PaymentHistoryFilter) ([]*services.PaymentHistoryResponse, error)
}

func (m *MockSupplierPaymentService) RecordPayment(ctx context.Context, request *services.RecordPaymentRequest, createdBy uint, ipAddress string) (*models.SupplierPayment, error) {
	if m.recordPaymentFunc != nil {
		return m.recordPaymentFunc(ctx, request, createdBy, ipAddress)
	}
	return &models.SupplierPayment{}, nil
}

func (m *MockSupplierPaymentService) GetSupplierPaymentByID(ctx context.Context, id uint) (*models.SupplierPayment, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &models.SupplierPayment{}, nil
}

func (m *MockSupplierPaymentService) ListSupplierPayments(ctx context.Context, filter *services.SupplierPaymentListFilter) ([]*models.SupplierPayment, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return []*models.SupplierPayment{}, 0, nil
}

func (m *MockSupplierPaymentService) GetPaymentHistoryBySupplier(ctx context.Context, supplierID uint, filter *services.PaymentHistoryFilter) ([]*services.PaymentHistoryResponse, error) {
	if m.getPaymentHistoryBySupplierFunc != nil {
		return m.getPaymentHistoryBySupplierFunc(ctx, supplierID, filter)
	}
	return []*services.PaymentHistoryResponse{}, nil
}

// TestSupplierPaymentHandler_RecordPayment_Success tests successful payment recording
// Story 10.4, AC1: Admin/Owner can record supplier payments
func TestSupplierPaymentHandler_RecordPayment_Success(t *testing.T) {
	paymentDate := time.Now().Format("2006-01-02")

	createdPayment := &models.SupplierPayment{
		ID:                1,
		PurchaseInvoiceID: 1,
		PaymentDate:       time.Now(),
		PaymentAmount:     1500000.00,
		PaymentMethod:     "transfer",
		Notes:             "Payment for May 2026 invoice",
		ReferenceNumber:   "TRX-20260531-12345",
		BranchID:          1,
		CreatedBy:         1,
		PurchaseInvoice: models.PurchaseInvoice{
			ID:            1,
			InvoiceNumber: "INV-2026-001",
			InvoiceDate:   time.Now(),
			PaymentStatus: "partial",
			TotalAmount:   2000000.00,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService := &MockSupplierPaymentService{
		recordPaymentFunc: func(ctx context.Context, request *services.RecordPaymentRequest, createdBy uint, ipAddress string) (*models.SupplierPayment, error) {
			assert.Equal(t, uint(1), request.PurchaseInvoiceID)
			assert.Equal(t, paymentDate, request.PaymentDate)
			assert.Equal(t, 1500000.00, request.PaymentAmount)
			assert.Equal(t, "transfer", request.PaymentMethod)
			assert.Equal(t, "Payment for May 2026 invoice", request.Notes)
			assert.Equal(t, "TRX-20260531-12345", request.ReferenceNumber)
			assert.Equal(t, uint(1), createdBy)
			assert.Equal(t, "127.0.0.1", ipAddress)
			return createdPayment, nil
		},
	}

	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	requestBody := map[string]interface{}{
		"purchaseInvoiceId": 1,
		"paymentDate":       paymentDate,
		"paymentAmount":     1500000.00,
		"paymentMethod":     "transfer",
		"notes":             "Payment for May 2026 invoice",
		"referenceNumber":   "TRX-20260531-12345",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/supplier-payments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response dto.SupplierPaymentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, "INV-2026-001", response.InvoiceNumber)
	assert.Equal(t, "partial", response.PaymentStatus)
	assert.Equal(t, 1500000.00, response.PaymentAmount)
}

// TestSupplierPaymentHandler_RecordPayment_InvalidJSON tests invalid JSON handling
// Story 10.4: Handler validates request format
func TestSupplierPaymentHandler_RecordPayment_InvalidJSON(t *testing.T) {
	mockService := &MockSupplierPaymentService{}
	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	req, _ := http.NewRequest("POST", "/api/v1/supplier-payments", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/validation-error", errorResp.Type)
}

// TestSupplierPaymentHandler_RecordPayment_InvalidPaymentMethod tests invalid payment method
// Story 10.4: Handler validates payment method enum
func TestSupplierPaymentHandler_RecordPayment_InvalidPaymentMethod(t *testing.T) {
	mockService := &MockSupplierPaymentService{}
	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	requestBody := map[string]interface{}{
		"purchaseInvoiceId": 1,
		"paymentDate":       time.Now().Format("2006-01-02"),
		"paymentAmount":     1500000.00,
		"paymentMethod":     "invalid_method", // Invalid enum value
		"notes":             "Test",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/supplier-payments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Gin validation will reject invalid enum values
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSupplierPaymentHandler_RecordPayment_Overpayment tests overpayment prevention
// Story 10.4: Handler prevents payment exceeding remaining balance
func TestSupplierPaymentHandler_RecordPayment_Overpayment(t *testing.T) {
	mockService := &MockSupplierPaymentService{
		recordPaymentFunc: func(ctx context.Context, request *services.RecordPaymentRequest, createdBy uint, ipAddress string) (*models.SupplierPayment, error) {
			return nil, fmt.Errorf("payment amount (2000000.00) exceeds remaining balance (500000.00)")
		},
	}

	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	requestBody := map[string]interface{}{
		"purchaseInvoiceId": 1,
		"paymentDate":       time.Now().Format("2006-01-02"),
		"paymentAmount":     2000000.00,
		"paymentMethod":     "transfer",
		"notes":             "Test",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/supplier-payments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp.Detail, "exceeds remaining balance")
}

// TestSupplierPaymentHandler_RecordPayment_InvoiceNotFound tests invoice not found handling
// Story 10.4: Handler returns 404 for non-existent invoices
func TestSupplierPaymentHandler_RecordPayment_InvoiceNotFound(t *testing.T) {
	mockService := &MockSupplierPaymentService{
		recordPaymentFunc: func(ctx context.Context, request *services.RecordPaymentRequest, createdBy uint, ipAddress string) (*models.SupplierPayment, error) {
			return nil, fmt.Errorf("purchase invoice not found")
		},
	}

	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	requestBody := map[string]interface{}{
		"purchaseInvoiceId": 999,
		"paymentDate":       time.Now().Format("2006-01-02"),
		"paymentAmount":     1500000.00,
		"paymentMethod":     "transfer",
		"notes":             "Test",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/supplier-payments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/purchase-invoice-not-found", errorResp.Type)
}

// TestSupplierPaymentHandler_GetSupplierPayment_Success tests successful payment retrieval
// Story 10.4: Admin/Owner can view payment details
func TestSupplierPaymentHandler_GetSupplierPayment_Success(t *testing.T) {
	payment := &models.SupplierPayment{
		ID:                1,
		PurchaseInvoiceID: 1,
		PaymentDate:       time.Now(),
		PaymentAmount:     1500000.00,
		PaymentMethod:     "transfer",
		Notes:             "Payment for May 2026 invoice",
		ReferenceNumber:   "TRX-20260531-12345",
		BranchID:          1,
		CreatedBy:         1,
		PurchaseInvoice: models.PurchaseInvoice{
			ID:            1,
			InvoiceNumber: "INV-2026-001",
			InvoiceDate:   time.Now(),
			PaymentStatus: "partial",
			TotalAmount:   2000000.00,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService := &MockSupplierPaymentService{
		getByIDFunc: func(ctx context.Context, id uint) (*models.SupplierPayment, error) {
			assert.Equal(t, uint(1), id)
			return payment, nil
		},
	}

	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/supplier-payments/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SupplierPaymentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, "INV-2026-001", response.InvoiceNumber)
	assert.Equal(t, 1500000.00, response.PaymentAmount)
}

// TestSupplierPaymentHandler_GetSupplierPayment_InvalidID tests invalid ID handling
// Story 10.4: Handler validates ID parameter
func TestSupplierPaymentHandler_GetSupplierPayment_InvalidID(t *testing.T) {
	mockService := &MockSupplierPaymentService{}
	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/supplier-payments/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/invalid-id", errorResp.Type)
}

// TestSupplierPaymentHandler_GetSupplierPayment_NotFound tests not found handling
// Story 10.4: Handler returns 404 for non-existent payments
func TestSupplierPaymentHandler_GetSupplierPayment_NotFound(t *testing.T) {
	mockService := &MockSupplierPaymentService{
		getByIDFunc: func(ctx context.Context, id uint) (*models.SupplierPayment, error) {
			return nil, fmt.Errorf("supplier payment not found")
		},
	}

	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/supplier-payments/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/payment-not-found", errorResp.Type)
}

// TestSupplierPaymentHandler_GetSupplierPayment_BranchAccessDenied tests branch access validation
// Story 10.4, PATCH-003: Handler validates branch access
func TestSupplierPaymentHandler_GetSupplierPayment_BranchAccessDenied(t *testing.T) {
	payment := &models.SupplierPayment{
		ID:                1,
		PurchaseInvoiceID: 1,
		PaymentDate:       time.Now(),
		PaymentAmount:     1500000.00,
		PaymentMethod:     "transfer",
		BranchID:          2, // Different branch
		CreatedBy:         1,
		PurchaseInvoice: models.PurchaseInvoice{
			ID:            1,
			InvoiceNumber: "INV-2026-001",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService := &MockSupplierPaymentService{
		getByIDFunc: func(ctx context.Context, id uint) (*models.SupplierPayment, error) {
			return payment, nil
		},
	}

	// User from branch 1, payment from branch 2
	router := setupSupplierPaymentTestRouter(mockService, "CASHIER", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/supplier-payments/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/forbidden", errorResp.Type)
}

// TestSupplierPaymentHandler_ListSupplierPayments_Success tests successful listing with pagination
// Story 10.4: Admin/Owner can list payments with pagination
func TestSupplierPaymentHandler_ListSupplierPayments_Success(t *testing.T) {
	payments := []*models.SupplierPayment{
		{
			ID:                1,
			PurchaseInvoiceID: 1,
			PaymentDate:       time.Now(),
			PaymentAmount:     1500000.00,
			PaymentMethod:     "transfer",
			BranchID:          1,
			CreatedBy:         1,
			PurchaseInvoice: models.PurchaseInvoice{
				ID:            1,
				InvoiceNumber: "INV-2026-001",
				PaymentStatus: "partial",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:                2,
			PurchaseInvoiceID: 2,
			PaymentDate:       time.Now(),
			PaymentAmount:     2000000.00,
			PaymentMethod:     "cash",
			BranchID:          1,
			CreatedBy:         1,
			PurchaseInvoice: models.PurchaseInvoice{
				ID:            2,
				InvoiceNumber: "INV-2026-002",
				PaymentStatus: "paid",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockService := &MockSupplierPaymentService{
		listFunc: func(ctx context.Context, filter *services.SupplierPaymentListFilter) ([]*models.SupplierPayment, int64, error) {
			assert.Equal(t, 1, filter.Page)
			assert.Equal(t, 20, filter.Limit)
			return payments, int64(2), nil
		},
	}

	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/supplier-payments?page=1&limit=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SupplierPaymentListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, int64(2), response.Pagination.Total)
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 20, response.Pagination.Limit)
}

// TestSupplierPaymentHandler_ListSupplierPayments_WithFilters tests filtering functionality
// Story 10.4: Payments can be filtered by invoice, date range, payment method, branch
func TestSupplierPaymentHandler_ListSupplierPayments_WithFilters(t *testing.T) {
	mockService := &MockSupplierPaymentService{
		listFunc: func(ctx context.Context, filter *services.SupplierPaymentListFilter) ([]*models.SupplierPayment, int64, error) {
			// Verify filters are passed correctly
			assert.NotNil(t, filter.PurchaseInvoiceID)
			assert.Equal(t, uint(1), *filter.PurchaseInvoiceID)
			assert.NotNil(t, filter.StartDate)
			assert.Equal(t, "2026-05-01", *filter.StartDate)
			assert.NotNil(t, filter.EndDate)
			assert.Equal(t, "2026-05-31", *filter.EndDate)
			assert.NotNil(t, filter.PaymentMethod)
			assert.Equal(t, "transfer", *filter.PaymentMethod)
			assert.NotNil(t, filter.BranchID)
			assert.Equal(t, uint(1), *filter.BranchID)
			return []*models.SupplierPayment{}, 0, nil
		},
	}

	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/supplier-payments?purchase_invoice_id=1&start_date=2026-05-01&end_date=2026-05-31&payment_method=transfer&branch_id=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSupplierPaymentHandler_ListSupplierPayments_InvalidInvoiceID tests invalid invoice ID
// Story 10.4: Handler validates purchase_invoice_id parameter
func TestSupplierPaymentHandler_ListSupplierPayments_InvalidInvoiceID(t *testing.T) {
	mockService := &MockSupplierPaymentService{}
	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/supplier-payments?purchase_invoice_id=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/validation-error", errorResp.Type)
}

// TestSupplierPaymentHandler_GetPaymentHistoryBySupplier_Success tests successful payment history retrieval
// Story 10.4, AC2: Admin/Owner can view payment history by supplier
func TestSupplierPaymentHandler_GetPaymentHistoryBySupplier_Success(t *testing.T) {
	history := []*services.PaymentHistoryResponse{
		{
			ID:                1,
			PaymentDate:       "2026-05-31",
			PaymentAmount:     1500000.00,
			PaymentMethod:     "transfer",
			Notes:             "Payment for May 2026",
			ReferenceNumber:   "TRX-20260531-12345",
			InvoiceNumber:     "INV-2026-001",
			InvoiceDate:       "2026-05-30",
			InvoiceTotalAmount: 2000000.00,
			RemainingBalance:  500000.00,
		},
		{
			ID:                2,
			PaymentDate:       "2026-05-15",
			PaymentAmount:     1000000.00,
			PaymentMethod:     "cash",
			Notes:             "Partial payment",
			InvoiceNumber:     "INV-2026-002",
			InvoiceDate:       "2026-05-10",
			InvoiceTotalAmount: 1500000.00,
			RemainingBalance:  500000.00,
		},
	}

	mockService := &MockSupplierPaymentService{
		getPaymentHistoryBySupplierFunc: func(ctx context.Context, supplierID uint, filter *services.PaymentHistoryFilter) ([]*services.PaymentHistoryResponse, error) {
			assert.Equal(t, uint(1), supplierID)
			assert.Equal(t, 1, filter.Page)
			assert.Equal(t, 20, filter.Limit)
			return history, nil
		},
	}

	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/suppliers/1/payment-history?page=1&limit=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PaymentHistoryListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, "2026-05-31", response.Data[0].PaymentDate)
	assert.Equal(t, "INV-2026-001", response.Data[0].InvoiceNumber)
}

// TestSupplierPaymentHandler_GetPaymentHistoryBySupplier_WithDateFilters tests date range filtering
// Story 10.4, AC2: Payment history can be filtered by date range
func TestSupplierPaymentHandler_GetPaymentHistoryBySupplier_WithDateFilters(t *testing.T) {
	mockService := &MockSupplierPaymentService{
		getPaymentHistoryBySupplierFunc: func(ctx context.Context, supplierID uint, filter *services.PaymentHistoryFilter) ([]*services.PaymentHistoryResponse, error) {
			assert.NotNil(t, filter.StartDate)
			assert.Equal(t, "2026-05-01", *filter.StartDate)
			assert.NotNil(t, filter.EndDate)
			assert.Equal(t, "2026-05-31", *filter.EndDate)
			return []*services.PaymentHistoryResponse{}, nil
		},
	}

	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/suppliers/1/payment-history?start_date=2026-05-01&end_date=2026-05-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSupplierPaymentHandler_GetPaymentHistoryBySupplier_InvalidSupplierID tests invalid supplier ID
// Story 10.4: Handler validates supplier ID parameter
func TestSupplierPaymentHandler_GetPaymentHistoryBySupplier_InvalidSupplierID(t *testing.T) {
	mockService := &MockSupplierPaymentService{}
	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/suppliers/invalid/payment-history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/invalid-id", errorResp.Type)
}

// TestSupplierPaymentHandler_extractUserContext_TypeSafety tests user context extraction with type safety
// Story 10.4: Handler safely extracts user context with type checking
func TestSupplierPaymentHandler_extractUserContext_TypeSafety(t *testing.T) {
	testCases := []struct {
		name         string
		setupContext  func(*gin.Context)
		expectStatus int
	}{
		{
			name: "Valid uint user_id",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uint(1))
				c.Set("username", "testuser")
				c.Set("user_role", "OWNER")
				c.Set("branch_id", uint(1))
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "Valid int user_id",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("username", "testuser")
				c.Set("user_role", "ADMIN")
				c.Set("branch_id", 1)
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "Missing user_id",
			setupContext: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectStatus: http.StatusUnauthorized,
		},
		{
			name: "Missing username",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uint(1))
			},
			expectStatus: http.StatusUnauthorized,
		},
		{
			name: "Zero user_id",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uint(0))
				c.Set("username", "testuser")
			},
			expectStatus: http.StatusUnauthorized,
		},
		{
			name: "Empty username",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uint(1))
				c.Set("username", "")
			},
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			// Add custom middleware to set up context
			router.Use(func(c *gin.Context) {
				tc.setupContext(c)
				c.Next()
			})

			// Add a simple test handler that tries to extract user context
			var extractedUserCtx supplierPaymentUserContext
			var extractionOk bool
			router.GET("/test", func(c *gin.Context) {
				handler := NewSupplierPaymentHandler(&MockSupplierPaymentService{})
				extractedUserCtx, extractionOk = handler.extractUserContext(c)
				c.Status(http.StatusOK)
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tc.expectStatus == http.StatusOK {
				assert.True(t, extractionOk, "Expected successful extraction")
				assert.Equal(t, uint(1), extractedUserCtx.userID)
				assert.Equal(t, "testuser", extractedUserCtx.username)
			} else {
				assert.False(t, extractionOk, "Expected extraction failure")
			}
		})
	}
}

// TestSupplierPaymentHandler_validateBranchAccess_OwnerBypass tests Owner role bypass
// Story 10.4, PATCH-003: Owner can access all branches
func TestSupplierPaymentHandler_validateBranchAccess_OwnerBypass(t *testing.T) {
	// Test via actual HTTP call - Owner should access any branch
	mockService := &MockSupplierPaymentService{
		getByIDFunc: func(ctx context.Context, id uint) (*models.SupplierPayment, error) {
			return &models.SupplierPayment{
				ID:       1,
				BranchID: 999, // Different branch
				PurchaseInvoice: models.PurchaseInvoice{
					ID:            1,
					InvoiceNumber: "INV-001",
				},
			}, nil
		},
	}

	router := setupSupplierPaymentTestRouter(mockService, "OWNER", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/supplier-payments/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Owner should be able to access payments from any branch
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSupplierPaymentHandler_validateBranchAccess_AdminBypass tests Admin role bypass
// Story 10.4, PATCH-003: Admin can access all branches
func TestSupplierPaymentHandler_validateBranchAccess_AdminBypass(t *testing.T) {
	// Test via actual HTTP call - Admin should access any branch
	mockService := &MockSupplierPaymentService{
		getByIDFunc: func(ctx context.Context, id uint) (*models.SupplierPayment, error) {
			return &models.SupplierPayment{
				ID:       1,
				BranchID: 999, // Different branch
				PurchaseInvoice: models.PurchaseInvoice{
					ID:            1,
					InvoiceNumber: "INV-001",
				},
			}, nil
		},
	}

	router := setupSupplierPaymentTestRouter(mockService, "SYSTEM_ADMIN", 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/supplier-payments/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Admin should be able to access payments from any branch
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSupplierPaymentHandler_validateBranchAccess_Denied tests branch access denial
// Story 10.4, PATCH-003: Non-admin roles must match branch
func TestSupplierPaymentHandler_validateBranchAccess_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test router to simulate the scenario
	mockService := &MockSupplierPaymentService{}
	router := setupSupplierPaymentTestRouter(mockService, "CASHIER", 1, 1)

	// Create a payment from branch 2
	payment := &models.SupplierPayment{
		ID:                1,
		PurchaseInvoiceID: 1,
		PaymentDate:       time.Now(),
		PaymentAmount:     1500000.00,
		PaymentMethod:     "transfer",
		BranchID:          2, // Different branch
		CreatedBy:         1,
	}

	mockService.getByIDFunc = func(ctx context.Context, id uint) (*models.SupplierPayment, error) {
		return payment, nil
	}

	// Cashier from branch 1 should not access payment from branch 2
	req, _ := http.NewRequest("GET", "/api/v1/supplier-payments/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/forbidden", errorResp.Type)
}
