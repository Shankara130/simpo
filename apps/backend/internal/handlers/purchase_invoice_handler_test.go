package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// setupPurchaseInvoiceTestRouter creates a test router with the purchase invoice handler and auth context
// Story 10.2: Test helper for purchase invoice handler testing
func setupPurchaseInvoiceTestRouter(purchaseInvoiceService services.PurchaseInvoiceService, userRole string, branchID uint, userID uint) *gin.Engine {
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

	// Add purchase invoice handler
	handler := NewPurchaseInvoiceHandler(purchaseInvoiceService)
	router.POST("/api/v1/purchase-invoices", handler.CreatePurchaseInvoice)
	router.GET("/api/v1/purchase-invoices", handler.ListPurchaseInvoices)
	router.GET("/api/v1/purchase-invoices/:id", handler.GetPurchaseInvoice)
	router.PUT("/api/v1/purchase-invoices/:id", handler.UpdatePurchaseInvoice)
	router.DELETE("/api/v1/purchase-invoices/:id", handler.DeletePurchaseInvoice)

	return router
}

// MockPurchaseInvoiceService is a mock for testing
type MockPurchaseInvoiceService struct {
	createFunc func(ctx context.Context, invoice *models.PurchaseInvoice, items []services.CreatePurchaseInvoiceItemRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error)
	getByIDFunc func(ctx context.Context, id uint) (*models.PurchaseInvoice, error)
	listFunc func(ctx context.Context, filter *services.PurchaseInvoiceListFilter) ([]*models.PurchaseInvoice, int64, error)
	updateFunc func(ctx context.Context, id uint, req *services.UpdatePurchaseInvoiceRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error)
	deleteFunc func(ctx context.Context, id uint, userID uint, ipAddress string) error
}

func (m *MockPurchaseInvoiceService) CreatePurchaseInvoice(ctx context.Context, invoice *models.PurchaseInvoice, items []services.CreatePurchaseInvoiceItemRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, invoice, items, userID, ipAddress)
	}
	return &models.PurchaseInvoice{}, nil
}

func (m *MockPurchaseInvoiceService) GetPurchaseInvoiceByID(ctx context.Context, id uint) (*models.PurchaseInvoice, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &models.PurchaseInvoice{}, nil
}

func (m *MockPurchaseInvoiceService) ListPurchaseInvoices(ctx context.Context, filter *services.PurchaseInvoiceListFilter) ([]*models.PurchaseInvoice, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return []*models.PurchaseInvoice{}, 0, nil
}

func (m *MockPurchaseInvoiceService) UpdatePurchaseInvoice(ctx context.Context, id uint, req *services.UpdatePurchaseInvoiceRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, req, userID, ipAddress)
	}
	return &models.PurchaseInvoice{}, nil
}

func (m *MockPurchaseInvoiceService) DeletePurchaseInvoice(ctx context.Context, id uint, userID uint, ipAddress string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id, userID, ipAddress)
	}
	return nil
}

// GetSuggestedPrice retrieves suggested price from supplier catalog
// Story 10.5, AC1: Mock implementation for testing
func (m *MockPurchaseInvoiceService) GetSuggestedPrice(ctx context.Context, supplierID uint, productID uint, branchID uint) (float64, error) {
	return 0, fmt.Errorf("catalog price not found")
}

// TestPurchaseInvoiceHandler_CreatePurchaseInvoice_Success tests successful purchase invoice creation
// Story 10.2, AC1: Admin/Owner can create purchase invoices with items
func TestPurchaseInvoiceHandler_CreatePurchaseInvoice_Success(t *testing.T) {
	invoiceDate := time.Now().Format("2006-01-02")

	// Setup mock service
	createdInvoice := &models.PurchaseInvoice{
		ID:            1,
		InvoiceNumber: "INV-2025-001",
		InvoiceDate:   time.Now(),
		SupplierID:    1,
		BranchID:      1,
		TotalAmount:   150000.00,
		PaymentStatus: "unpaid",
		Supplier: models.Supplier{
			ID:   1,
			Name: "PT Pharma Jaya",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService := &MockPurchaseInvoiceService{
		createFunc: func(ctx context.Context, invoice *models.PurchaseInvoice, items []services.CreatePurchaseInvoiceItemRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
			assert.Equal(t, "INV-2025-001", invoice.InvoiceNumber)
			assert.Equal(t, uint(1), invoice.SupplierID)
			assert.Equal(t, uint(1), invoice.BranchID)
			assert.Len(t, items, 2)
			assert.Equal(t, uint(1), userID)
			return createdInvoice, nil
		},
	}

	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	// Create request body
	requestBody := map[string]interface{}{
		"invoiceNumber": "INV-2025-001",
		"invoiceDate":   invoiceDate,
		"supplierId":     1,
		"branchId":       1,
		"notes":          "Test invoice",
		"items": []map[string]interface{}{
			{
				"productId": 1,
				"quantity":  10,
				"unitCost":  "10000.00",
			},
			{
				"productId": 2,
				"quantity":  5,
				"unitCost":  "10000.00",
			},
		},
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/purchase-invoices", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response dto.PurchaseInvoiceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "INV-2025-001", response.InvoiceNumber)
	assert.Equal(t, "PT Pharma Jaya", response.SupplierName)
}

// TestPurchaseInvoiceHandler_CreatePurchaseInvoice_InvalidJSON tests invalid JSON handling
// Story 10.2: Handler validates request format
func TestPurchaseInvoiceHandler_CreatePurchaseInvoice_InvalidJSON(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{}
	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	req, _ := http.NewRequest("POST", "/api/v1/purchase-invoices", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/validation-error", errorResp.Type)
}

// TestPurchaseInvoiceHandler_CreatePurchaseInvoice_InvalidDateFormat tests invalid date format
// Story 10.2: Handler validates date format
func TestPurchaseInvoiceHandler_CreatePurchaseInvoice_InvalidDateFormat(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{}
	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	requestBody := map[string]interface{}{
		"invoiceNumber": "INV-2025-001",
		"invoiceDate":   "invalid-date",
		"supplierId":     1,
		"branchId":       1,
		"items":          []map[string]interface{}{},
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/purchase-invoices", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp.Detail, "Invalid invoice date format")
}

// TestPurchaseInvoiceHandler_CreatePurchaseInvoice_DuplicateInvoice tests duplicate invoice handling
// Story 10.2: Handler returns 409 Conflict for duplicate invoice numbers
func TestPurchaseInvoiceHandler_CreatePurchaseInvoice_DuplicateInvoice(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{
		createFunc: func(ctx context.Context, invoice *models.PurchaseInvoice, items []services.CreatePurchaseInvoiceItemRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
			return nil, &services.DuplicateInvoiceError{InvoiceNumber: invoice.InvoiceNumber}
		},
	}

	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	requestBody := map[string]interface{}{
		"invoiceNumber": "INV-2025-001",
		"invoiceDate":   time.Now().Format("2006-01-02"),
		"supplierId":     1,
		"branchId":       1,
		"items":          []map[string]interface{}{},
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/purchase-invoices", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/duplicate-invoice", errorResp.Type)
}

// TestPurchaseInvoiceHandler_GetPurchaseInvoice_Success tests successful retrieval
// Story 10.2, AC3: Admin/Owner can view purchase invoice details
func TestPurchaseInvoiceHandler_GetPurchaseInvoice_Success(t *testing.T) {
	invoice := &models.PurchaseInvoice{
		ID:            1,
		InvoiceNumber: "INV-2025-001",
		InvoiceDate:   time.Now(),
		SupplierID:    1,
		BranchID:      1,
		TotalAmount:   150000.00,
		PaymentStatus: "unpaid",
		Supplier: models.Supplier{
			ID:   1,
			Name: "PT Pharma Jaya",
		},
		Items: []models.PurchaseInvoiceItem{
			{
				ID:        1,
				ProductID: 1,
				Quantity:  10,
				UnitCost:  10000.00,
				Subtotal:  100000.00,
				Product: models.Product{
					ID:   1,
					Name: "Paracetamol 500mg",
					SKU:  "PARA-001",
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService := &MockPurchaseInvoiceService{
		getByIDFunc: func(ctx context.Context, id uint) (*models.PurchaseInvoice, error) {
			assert.Equal(t, uint(1), id)
			return invoice, nil
		},
	}

	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/purchase-invoices/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PurchaseInvoiceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "INV-2025-001", response.InvoiceNumber)
	assert.Equal(t, "PT Pharma Jaya", response.SupplierName)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, "Paracetamol 500mg", response.Items[0].ProductName)
}

// TestPurchaseInvoiceHandler_GetPurchaseInvoice_InvalidID tests invalid ID handling
// Story 10.2: Handler validates ID parameter
func TestPurchaseInvoiceHandler_GetPurchaseInvoice_InvalidID(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{}
	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/purchase-invoices/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/invalid-id", errorResp.Type)
}

// TestPurchaseInvoiceHandler_GetPurchaseInvoice_NotFound tests not found handling
// Story 10.2: Handler returns 404 for non-existent invoices
func TestPurchaseInvoiceHandler_GetPurchaseInvoice_NotFound(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{
		getByIDFunc: func(ctx context.Context, id uint) (*models.PurchaseInvoice, error) {
			return nil, &services.InvoiceNotFoundError{ID: id}
		},
	}

	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/purchase-invoices/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/purchase-invoice-not-found", errorResp.Type)
}

// TestPurchaseInvoiceHandler_ListPurchaseInvoices_Success tests successful listing with pagination
// Story 10.2, AC2: Admin/Owner can list invoices with pagination
func TestPurchaseInvoiceHandler_ListPurchaseInvoices_Success(t *testing.T) {
	invoices := []*models.PurchaseInvoice{
		{
			ID:            1,
			InvoiceNumber: "INV-2025-001",
			InvoiceDate:   time.Now(),
			SupplierID:    1,
			BranchID:      1,
			TotalAmount:   150000.00,
			PaymentStatus: "unpaid",
			Supplier: models.Supplier{
				ID:   1,
				Name: "PT Pharma Jaya",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:            2,
			InvoiceNumber: "INV-2025-002",
			InvoiceDate:   time.Now(),
			SupplierID:    2,
			BranchID:      1,
			TotalAmount:   200000.00,
			PaymentStatus: "paid",
			Supplier: models.Supplier{
				ID:   2,
				Name: "CV Medica Sehat",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockService := &MockPurchaseInvoiceService{
		listFunc: func(ctx context.Context, filter *services.PurchaseInvoiceListFilter) ([]*models.PurchaseInvoice, int64, error) {
			assert.Equal(t, 1, filter.Page)
			assert.Equal(t, 20, filter.Limit)
			return invoices, int64(2), nil
		},
	}

	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleOwner, 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/purchase-invoices?page=1&limit=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PurchaseInvoiceListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, int64(2), response.Pagination.Total)
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 20, response.Pagination.Limit)
}

// TestPurchaseInvoiceHandler_ListPurchaseInvoices_WithFilters tests filtering functionality
// Story 10.2, AC2: Invoices can be filtered by supplier, date range, payment status
func TestPurchaseInvoiceHandler_ListPurchaseInvoices_WithFilters(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{
		listFunc: func(ctx context.Context, filter *services.PurchaseInvoiceListFilter) ([]*models.PurchaseInvoice, int64, error) {
			// Verify filters are passed correctly
			assert.NotNil(t, filter.SupplierID)
			assert.Equal(t, uint(1), *filter.SupplierID)
			assert.NotNil(t, filter.StartDate)
			assert.Equal(t, "2025-01-01", *filter.StartDate)
			assert.NotNil(t, filter.EndDate)
			assert.Equal(t, "2025-12-31", *filter.EndDate)
			assert.NotNil(t, filter.PaymentStatus)
			assert.Equal(t, "unpaid", *filter.PaymentStatus)
			return []*models.PurchaseInvoice{}, 0, nil
		},
	}

	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleOwner, 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/purchase-invoices?supplierId=1&start_date=2025-01-01&end_date=2025-12-31&payment_status=unpaid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestPurchaseInvoiceHandler_ListPurchaseInvoices_InvalidSupplierID tests invalid supplier ID
// Story 10.2: Handler validates supplier_id parameter
func TestPurchaseInvoiceHandler_ListPurchaseInvoices_InvalidSupplierID(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{}
	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleOwner, 1, 1)

	req, _ := http.NewRequest("GET", "/api/v1/purchase-invoices?supplierId=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "/errors/validation-error", errorResp.Type)
}

// TestPurchaseInvoiceHandler_UpdatePurchaseInvoice_Success tests successful update
// Story 10.2: Admin can update purchase invoices
func TestPurchaseInvoiceHandler_UpdatePurchaseInvoice_Success(t *testing.T) {
	updatedInvoice := &models.PurchaseInvoice{
		ID:            1,
		InvoiceNumber: "INV-2025-001-UPDATED",
		InvoiceDate:   time.Now(),
		SupplierID:    1,
		BranchID:      1,
		TotalAmount:   150000.00,
		PaymentStatus: "unpaid",
		Notes:         "Updated notes",
		Supplier: models.Supplier{
			ID:   1,
			Name: "PT Pharma Jaya",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService := &MockPurchaseInvoiceService{
		updateFunc: func(ctx context.Context, id uint, req *services.UpdatePurchaseInvoiceRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
			assert.Equal(t, uint(1), id)
			assert.Equal(t, "Updated notes", req.Reason)
			assert.Equal(t, uint(1), userID)
			return updatedInvoice, nil
		},
	}

	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	requestBody := map[string]interface{}{
		"invoiceNumber": "INV-2025-001-UPDATED",
		"invoiceDate":   time.Now().Format("2006-01-02"),
		"supplierId":     1,
		"notes":          "Updated notes",
		"items":          []map[string]interface{}{},
		"reason":         "Updating invoice details",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/api/v1/purchase-invoices/1", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PurchaseInvoiceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Updated notes", response.Notes)
}

// TestPurchaseInvoiceHandler_UpdatePurchaseInvoice_InvalidID tests invalid ID handling
// Story 10.2: Handler validates ID parameter for updates
func TestPurchaseInvoiceHandler_UpdatePurchaseInvoice_InvalidID(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{}
	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	requestBody := map[string]interface{}{
		"items":  []map[string]interface{}{},
		"reason": "Test",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/api/v1/purchase-invoices/invalid", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPurchaseInvoiceHandler_UpdatePurchaseInvoice_NotFound tests not found handling
// Story 10.2: Handler returns 404 for non-existent invoices
func TestPurchaseInvoiceHandler_UpdatePurchaseInvoice_NotFound(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{
		updateFunc: func(ctx context.Context, id uint, req *services.UpdatePurchaseInvoiceRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
			return nil, &services.InvoiceNotFoundError{ID: id}
		},
	}

	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	requestBody := map[string]interface{}{
		"items":  []map[string]interface{}{},
		"reason": "Test",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/api/v1/purchase-invoices/999", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestPurchaseInvoiceHandler_DeletePurchaseInvoice_Success tests successful deletion
// Story 10.2: Admin can soft delete purchase invoices
func TestPurchaseInvoiceHandler_DeletePurchaseInvoice_Success(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{
		deleteFunc: func(ctx context.Context, id uint, userID uint, ipAddress string) error {
			assert.Equal(t, uint(1), id)
			assert.Equal(t, uint(1), userID)
			return nil
		},
	}

	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	requestBody := map[string]interface{}{
		"reason": "Invoice cancelled - supplier error",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("DELETE", "/api/v1/purchase-invoices/1", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Purchase invoice deleted", response["message"])
	assert.Equal(t, "Invoice cancelled - supplier error", response["reason"])
}

// TestPurchaseInvoiceHandler_DeletePurchaseInvoice_InvalidID tests invalid ID handling
// Story 10.2: Handler validates ID parameter for deletion
func TestPurchaseInvoiceHandler_DeletePurchaseInvoice_InvalidID(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{}
	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	requestBody := map[string]interface{}{
		"reason": "Test",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("DELETE", "/api/v1/purchase-invoices/invalid", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPurchaseInvoiceHandler_DeletePurchaseInvoice_NotFound tests not found handling
// Story 10.2: Handler returns 404 for non-existent invoices
func TestPurchaseInvoiceHandler_DeletePurchaseInvoice_NotFound(t *testing.T) {
	mockService := &MockPurchaseInvoiceService{
		deleteFunc: func(ctx context.Context, id uint, userID uint, ipAddress string) error {
			return &services.InvoiceNotFoundError{ID: id}
		},
	}

	router := setupPurchaseInvoiceTestRouter(mockService, user.RoleAdmin, 1, 1)

	requestBody := map[string]interface{}{
		"reason": "Test",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("DELETE", "/api/v1/purchase-invoices/999", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestPurchaseInvoiceHandler_extractUserContext_TypeSafety tests user context extraction with type safety
// Story 10.2: Handler safely extracts user context with type checking
func TestPurchaseInvoiceHandler_extractUserContext_TypeSafety(t *testing.T) {
	testCases := []struct {
		name          string
		setupContext  func(*gin.Context)
		expectSuccess bool
	}{
		{
			name: "Valid uint user_id",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uint(1))
				c.Set("username", "testuser")
			},
			expectSuccess: true,
		},
		{
			name: "Valid int user_id",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("username", "testuser")
			},
			expectSuccess: true,
		},
		{
			name: "Valid int64 user_id",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", int64(1))
				c.Set("username", "testuser")
			},
			expectSuccess: true,
		},
		{
			name: "Valid float64 user_id",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", float64(1))
				c.Set("username", "testuser")
			},
			expectSuccess: true,
		},
		{
			name: "Valid string user_id",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", "1")
				c.Set("username", "testuser")
			},
			expectSuccess: true,
		},
		{
			name: "Invalid string user_id",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", "invalid")
				c.Set("username", "testuser")
			},
			expectSuccess: false,
		},
		{
			name: "Missing user_id",
			setupContext: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectSuccess: false,
		},
		{
			name: "Missing username",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uint(1))
			},
			expectSuccess: false,
		},
		{
			name: "Zero user_id",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uint(0))
				c.Set("username", "testuser")
			},
			expectSuccess: false,
		},
		{
			name: "Empty username",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uint(1))
				c.Set("username", "")
			},
			expectSuccess: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			tc.setupContext(c)

			handler := NewPurchaseInvoiceHandler(&MockPurchaseInvoiceService{})
			userCtx, ok := handler.extractUserContext(c)

			if tc.expectSuccess {
				assert.True(t, ok, "Expected successful extraction")
				assert.Equal(t, uint(1), userCtx.userID)
				assert.Equal(t, "testuser", userCtx.username)
			} else {
				assert.False(t, ok, "Expected extraction failure")
			}
		})
	}
}

// TestPurchaseInvoiceHandler_sanitizePurchaseInvoiceReason tests input sanitization
// Story 10.2: Handler sanitizes user input to prevent injection
func TestPurchaseInvoiceHandler_sanitizePurchaseInvoiceReason(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal text",
			input:    "Updating invoice details",
			expected: "Updating invoice details",
		},
		{
			name:     "Text with extra whitespace",
			input:    "   Updating invoice details   ",
			expected: "Updating invoice details",
		},
		{
			name:     "Text with null bytes",
			input:    "Updating\x00invoice\x00details",
			expected: "Updatinginvoicedetails",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Very long text",
			input:    string(make([]byte, 600)),
			expected: string(make([]byte, 500)), // Truncated to 500
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizePurchaseInvoiceReason(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestPurchaseInvoiceHandler_toPurchaseInvoiceResponse tests response conversion
// Story 10.2: Handler correctly converts model to DTO
func TestPurchaseInvoiceHandler_toPurchaseInvoiceResponse(t *testing.T) {
	invoice := &models.PurchaseInvoice{
		ID:            1,
		InvoiceNumber: "INV-2025-001",
		InvoiceDate:   time.Now(),
		SupplierID:    1,
		BranchID:      1,
		TotalAmount:   150000.00,
		PaymentStatus: "unpaid",
		Notes:         "Test notes",
		DocumentURL:   "https://example.com/invoice.pdf",
		Supplier: models.Supplier{
			ID:   1,
			Name: "PT Pharma Jaya",
		},
		Items: []models.PurchaseInvoiceItem{
			{
				ID:        1,
				ProductID: 1,
				Quantity:  10,
				UnitCost:  10000.00,
				Subtotal:  100000.00,
				Product: models.Product{
					ID:   1,
					Name: "Paracetamol 500mg",
					SKU:  "PARA-001",
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler := NewPurchaseInvoiceHandler(&MockPurchaseInvoiceService{})
	response := handler.toPurchaseInvoiceResponse(invoice)

	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, "INV-2025-001", response.InvoiceNumber)
	assert.Equal(t, "PT Pharma Jaya", response.SupplierName)
	assert.Equal(t, "unpaid", response.PaymentStatus)
	assert.Equal(t, "Test notes", response.Notes)
	assert.Equal(t, "https://example.com/invoice.pdf", response.DocumentURL)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, "Paracetamol 500mg", response.Items[0].ProductName)
	assert.Equal(t, "PARA-001", response.Items[0].ProductSKU)
}

// TestPurchaseInvoiceHandler_toPurchaseInvoiceListResponse tests list response conversion
// Story 10.2: Handler correctly converts list with pagination
func TestPurchaseInvoiceHandler_toPurchaseInvoiceListResponse(t *testing.T) {
	invoices := []*models.PurchaseInvoice{
		{
			ID:            1,
			InvoiceNumber: "INV-001",
			InvoiceDate:   time.Now(),
			TotalAmount:   100000.00,
			PaymentStatus: "unpaid",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            2,
			InvoiceNumber: "INV-002",
			InvoiceDate:   time.Now(),
			TotalAmount:   200000.00,
			PaymentStatus: "paid",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	handler := NewPurchaseInvoiceHandler(&MockPurchaseInvoiceService{})
	response := handler.toPurchaseInvoiceListResponse(invoices, 25, 1, 20)

	assert.Len(t, response.Data, 2)
	assert.Equal(t, int64(25), response.Pagination.Total)
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 20, response.Pagination.Limit)
	assert.Equal(t, 2, response.Pagination.TotalPages) // 25 / 20 = 2 pages
}
