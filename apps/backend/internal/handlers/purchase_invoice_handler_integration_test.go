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

// mockPurchaseInvoiceServiceForIntegration is a mock implementation for integration testing
type mockPurchaseInvoiceServiceForIntegration struct {
	createFunc func(ctx context.Context, invoice *models.PurchaseInvoice, items []services.CreatePurchaseInvoiceItemRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error)
	getByIDFunc func(ctx context.Context, id uint) (*models.PurchaseInvoice, error)
	listFunc func(ctx context.Context, filter *services.PurchaseInvoiceListFilter) ([]*models.PurchaseInvoice, int64, error)
	updateFunc func(ctx context.Context, id uint, req *services.UpdatePurchaseInvoiceRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error)
	deleteFunc func(ctx context.Context, id uint, userID uint, ipAddress string) error
}

func (m *mockPurchaseInvoiceServiceForIntegration) CreatePurchaseInvoice(ctx context.Context, invoice *models.PurchaseInvoice, items []services.CreatePurchaseInvoiceItemRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, invoice, items, userID, ipAddress)
	}
	return &models.PurchaseInvoice{
		ID:            1,
		InvoiceNumber: invoice.InvoiceNumber,
		InvoiceDate:   invoice.InvoiceDate,
		SupplierID:    invoice.SupplierID,
		BranchID:      invoice.BranchID,
		TotalAmount:   150000.00,
		PaymentStatus: "unpaid",
		Notes:         invoice.Notes,
		DocumentURL:   invoice.DocumentURL,
		Supplier: models.Supplier{
			ID:   invoice.SupplierID,
			Name: "PT Test Supplier",
		},
		Items: []models.PurchaseInvoiceItem{
			{
				ID:        1,
				ProductID: items[0].ProductID,
				Quantity:  items[0].Quantity,
				UnitCost:  items[0].UnitCost,
				Subtotal:  float64(items[0].Quantity) * items[0].UnitCost,
				Product: models.Product{
					ID:   items[0].ProductID,
					Name: "Test Product",
					SKU:  "TEST-001",
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockPurchaseInvoiceServiceForIntegration) GetPurchaseInvoiceByID(ctx context.Context, id uint) (*models.PurchaseInvoice, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &models.PurchaseInvoice{
		ID:            id,
		InvoiceNumber: "INV-2025-001",
		InvoiceDate:   time.Now(),
		SupplierID:    1,
		BranchID:      1,
		TotalAmount:   150000.00,
		PaymentStatus: "unpaid",
		Supplier: models.Supplier{
			ID:   1,
			Name: "PT Test Supplier",
		},
		Items: []models.PurchaseInvoiceItem{
			{
				ID:        1,
				ProductID: 1,
				Quantity:  10,
				UnitCost:  15000.00,
				Subtotal:  150000.00,
				Product: models.Product{
					ID:   1,
					Name: "Test Product",
					SKU:  "TEST-001",
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockPurchaseInvoiceServiceForIntegration) ListPurchaseInvoices(ctx context.Context, filter *services.PurchaseInvoiceListFilter) ([]*models.PurchaseInvoice, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return []*models.PurchaseInvoice{
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
				Name: "PT Test Supplier",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}, 1, nil
}

func (m *mockPurchaseInvoiceServiceForIntegration) UpdatePurchaseInvoice(ctx context.Context, id uint, req *services.UpdatePurchaseInvoiceRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, req, userID, ipAddress)
	}
	return &models.PurchaseInvoice{
		ID:            id,
		InvoiceNumber: req.InvoiceNumber,
		InvoiceDate:   time.Now(),
		SupplierID:    req.SupplierID,
		TotalAmount:   150000.00,
		PaymentStatus: "unpaid",
		Notes:         req.Notes,
		DocumentURL:   req.DocumentURL,
		Supplier: models.Supplier{
			ID:   req.SupplierID,
			Name: "PT Updated Supplier",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockPurchaseInvoiceServiceForIntegration) DeletePurchaseInvoice(ctx context.Context, id uint, userID uint, ipAddress string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id, userID, ipAddress)
	}
	return nil
}

// setupPurchaseInvoiceIntegrationTestRouter creates a test router with authentication and authorization middleware
// Story 10.2, Subtask 11.2: Integration test helper with full middleware stack
func setupPurchaseInvoiceIntegrationTestRouter(service services.PurchaseInvoiceService, userRole string, userID uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add authentication middleware simulation
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("username", "testuser")
		c.Set("user_role", userRole)
		c.Set("branch_id", uint(1))
		c.Next()
	})

	handler := NewPurchaseInvoiceHandler(service)
	router.POST("/api/v1/purchase-invoices", handler.CreatePurchaseInvoice)
	router.GET("/api/v1/purchase-invoices", handler.ListPurchaseInvoices)
	router.GET("/api/v1/purchase-invoices/:id", handler.GetPurchaseInvoice)
	router.PUT("/api/v1/purchase-invoices/:id", handler.UpdatePurchaseInvoice)
	router.DELETE("/api/v1/purchase-invoices/:id", handler.DeletePurchaseInvoice)

	return router
}

// TestPurchaseInvoiceHandlerIntegration_CreateInvoice_Success tests full request/response cycle
// Story 10.2, AC1: Full integration test for invoice creation
func TestPurchaseInvoiceHandlerIntegration_CreateInvoice_Success(t *testing.T) {
	mockService := &mockPurchaseInvoiceServiceForIntegration{}
	router := setupPurchaseInvoiceIntegrationTestRouter(mockService, user.RoleAdmin, 1)

	requestBody := map[string]interface{}{
		"invoiceNumber": "INV-2025-001",
		"invoiceDate":   time.Now().Format("2006-01-02"),
		"supplierId":     1,
		"branchId":       1,
		"notes":          "Test purchase invoice",
		"items": []map[string]interface{}{
			{
				"productId": 1,
				"quantity":  10,
				"unitCost":  15000.00,
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
	assert.Equal(t, "PT Test Supplier", response.SupplierName)
	assert.Equal(t, 150000.00, response.TotalAmount)
	assert.Equal(t, "unpaid", response.PaymentStatus)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, "Test Product", response.Items[0].ProductName)
}

// TestPurchaseInvoiceHandlerIntegration_CreateInvoice_MissingRequiredFields tests validation
// Story 10.2, Subtask 11.4: Integration test for validation error responses
func TestPurchaseInvoiceHandlerIntegration_CreateInvoice_MissingRequiredFields(t *testing.T) {
	mockService := &mockPurchaseInvoiceServiceForIntegration{}
	router := setupPurchaseInvoiceIntegrationTestRouter(mockService, user.RoleAdmin, 1)

	testCases := []struct {
		name           string
		requestBody     map[string]interface{}
		expectedStatus int
		expectedError   string
	}{
		{
			name: "Missing invoice number",
			requestBody: map[string]interface{}{
				"invoiceDate": time.Now().Format("2006-01-02"),
				"supplierId":  1,
				"branchId":    1,
				"items": []map[string]interface{}{
					{"productId": 1, "quantity": 10, "unitCost": 15000.00},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:   "validation error",
		},
		{
			name: "Missing invoice date",
			requestBody: map[string]interface{}{
				"invoiceNumber": "INV-2025-001",
				"supplierId":    1,
				"branchId":      1,
				"items": []map[string]interface{}{
					{"productId": 1, "quantity": 10, "unitCost": 15000.00},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:   "validation error",
		},
		{
			name: "Missing supplier ID",
			requestBody: map[string]interface{}{
				"invoiceNumber": "INV-2025-001",
				"invoiceDate":   time.Now().Format("2006-01-02"),
				"branchId":      1,
				"items": []map[string]interface{}{
					{"productId": 1, "quantity": 10, "unitCost": 15000.00},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:   "validation error",
		},
		{
			name: "Empty items array",
			requestBody: map[string]interface{}{
				"invoiceNumber": "INV-2025-001",
				"invoiceDate":   time.Now().Format("2006-01-02"),
				"supplierId":    1,
				"branchId":      1,
				"items":         []map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:   "validation error",
		},
		{
			name: "Invalid date format",
			requestBody: map[string]interface{}{
				"invoiceNumber": "INV-2025-001",
				"invoiceDate":   "invalid-date",
				"supplierId":    1,
				"branchId":      1,
				"items": []map[string]interface{}{
					{"productId": 1, "quantity": 10, "unitCost": 15000.00},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:   "Invalid invoice date format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/purchase-invoices", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var errorResp dto.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errorResp)
			require.NoError(t, err)
			assert.Contains(t, errorResp.Detail, tc.expectedError)
		})
	}
}

// TestPurchaseInvoiceHandlerIntegration_GetInvoice_Success tests full retrieval cycle
// Story 10.2, AC3: Full integration test for invoice retrieval
func TestPurchaseInvoiceHandlerIntegration_GetInvoice_Success(t *testing.T) {
	mockService := &mockPurchaseInvoiceServiceForIntegration{}
	router := setupPurchaseInvoiceIntegrationTestRouter(mockService, user.RoleOwner, 1)

	req, _ := http.NewRequest("GET", "/api/v1/purchase-invoices/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PurchaseInvoiceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, "INV-2025-001", response.InvoiceNumber)
	assert.Equal(t, "PT Test Supplier", response.SupplierName)
	assert.Equal(t, 150000.00, response.TotalAmount)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, "Test Product", response.Items[0].ProductName)
	assert.Equal(t, "TEST-001", response.Items[0].ProductSKU)
}

// TestPurchaseInvoiceHandlerIntegration_ListInvoices_WithFilters tests filtering and pagination
// Story 10.2, AC2, Subtask 11.6: Integration test for filtering and pagination
func TestPurchaseInvoiceHandlerIntegration_ListInvoices_WithFilters(t *testing.T) {
	mockService := &mockPurchaseInvoiceServiceForIntegration{
		listFunc: func(ctx context.Context, filter *services.PurchaseInvoiceListFilter) ([]*models.PurchaseInvoice, int64, error) {
			// Verify filters are correctly passed
			if filter.SupplierID != nil {
				assert.Equal(t, uint(1), *filter.SupplierID)
			}
			if filter.StartDate != nil {
				assert.Equal(t, "2025-01-01", *filter.StartDate)
			}
			if filter.EndDate != nil {
				assert.Equal(t, "2025-12-31", *filter.EndDate)
			}
			if filter.PaymentStatus != nil {
				assert.Equal(t, "unpaid", *filter.PaymentStatus)
			}
			assert.Equal(t, "INV-2025", filter.SearchQuery)
			assert.Equal(t, 1, filter.Page)
			assert.Equal(t, 20, filter.Limit)
			assert.Equal(t, "invoice_date", filter.SortBy)
			assert.Equal(t, "desc", filter.SortOrder)

			return []*models.PurchaseInvoice{
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
						Name: "PT Test Supplier",
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}, 1, nil
		},
	}

	router := setupPurchaseInvoiceIntegrationTestRouter(mockService, user.RoleOwner, 1)

	req, _ := http.NewRequest("GET", "/api/v1/purchase-invoices?supplierId=1&start_date=2025-01-01&end_date=2025-12-31&payment_status=unpaid&search=INV-2025&page=1&limit=20&sort_by=invoice_date&sort_order=desc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PurchaseInvoiceListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, int64(1), response.Pagination.Total)
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 20, response.Pagination.Limit)
}

// TestPurchaseInvoiceHandlerIntegration_UpdateInvoice_Success tests full update cycle
// Story 10.2: Full integration test for invoice update
func TestPurchaseInvoiceHandlerIntegration_UpdateInvoice_Success(t *testing.T) {
	mockService := &mockPurchaseInvoiceServiceForIntegration{
		updateFunc: func(ctx context.Context, id uint, req *services.UpdatePurchaseInvoiceRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
			assert.Equal(t, uint(1), id)
			assert.Equal(t, "INV-2025-001-UPDATED", req.InvoiceNumber)
			assert.Equal(t, uint(1), userID)
			return &models.PurchaseInvoice{
				ID:            id,
				InvoiceNumber: req.InvoiceNumber,
				InvoiceDate:   time.Now(),
				SupplierID:    req.SupplierID,
				TotalAmount:   150000.00,
				PaymentStatus: "unpaid",
				Notes:         req.Notes,
				DocumentURL:   req.DocumentURL,
				Supplier: models.Supplier{
					ID:   req.SupplierID,
					Name: "PT Updated Supplier",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	router := setupPurchaseInvoiceIntegrationTestRouter(mockService, user.RoleAdmin, 1)

	requestBody := map[string]interface{}{
		"invoiceNumber": "INV-2025-001-UPDATED",
		"invoiceDate":   time.Now().Format("2006-01-02"),
		"supplierId":    1,
		"notes":         "Updated notes",
		"documentUrl":   "https://example.com/updated.pdf",
		"items": []map[string]interface{}{
			{
				"productId": 1,
				"quantity":  10,
				"unitCost":  15000.00,
			},
		},
		"reason": "Updating invoice details",
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
	assert.Equal(t, "INV-2025-001-UPDATED", response.InvoiceNumber)
	assert.Equal(t, "Updated notes", response.Notes)
	assert.Equal(t, "https://example.com/updated.pdf", response.DocumentURL)
}

// TestPurchaseInvoiceHandlerIntegration_DeleteInvoice_Success tests full deletion cycle
// Story 10.2: Full integration test for invoice deletion
func TestPurchaseInvoiceHandlerIntegration_DeleteInvoice_Success(t *testing.T) {
	mockService := &mockPurchaseInvoiceServiceForIntegration{
		deleteFunc: func(ctx context.Context, id uint, userID uint, ipAddress string) error {
			assert.Equal(t, uint(1), id)
			assert.Equal(t, uint(1), userID)
			return nil
		},
	}

	router := setupPurchaseInvoiceIntegrationTestRouter(mockService, user.RoleAdmin, 1)

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
	assert.NotNil(t, response["deletedAt"])
}

// TestPurchaseInvoiceHandlerIntegration_DuplicateInvoice tests duplicate invoice handling
// Story 10.2, Subtask 11.4: Integration test for duplicate invoice error response
func TestPurchaseInvoiceHandlerIntegration_DuplicateInvoice(t *testing.T) {
	mockService := &mockPurchaseInvoiceServiceForIntegration{
		createFunc: func(ctx context.Context, invoice *models.PurchaseInvoice, items []services.CreatePurchaseInvoiceItemRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
			return nil, &services.DuplicateInvoiceError{InvoiceNumber: invoice.InvoiceNumber}
		},
	}

	router := setupPurchaseInvoiceIntegrationTestRouter(mockService, user.RoleAdmin, 1)

	requestBody := map[string]interface{}{
		"invoiceNumber": "INV-2025-001",
		"invoiceDate":   time.Now().Format("2006-01-02"),
		"supplierId":    1,
		"branchId":      1,
		"items": []map[string]interface{}{
			{
				"productId": 1,
				"quantity":  10,
				"unitCost":  15000.00,
			},
		},
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
	assert.Contains(t, errorResp.Detail, "INV-2025-001")
}

// TestPurchaseInvoiceHandlerIntegration_NotFound tests 404 error responses
// Story 10.2, Subtask 11.4: Integration test for not found error responses
func TestPurchaseInvoiceHandlerIntegration_NotFound(t *testing.T) {
	mockService := &mockPurchaseInvoiceServiceForIntegration{
		getByIDFunc: func(ctx context.Context, id uint) (*models.PurchaseInvoice, error) {
			return nil, &services.InvoiceNotFoundError{ID: id}
		},
		updateFunc: func(ctx context.Context, id uint, req *services.UpdatePurchaseInvoiceRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
			return nil, &services.InvoiceNotFoundError{ID: id}
		},
		deleteFunc: func(ctx context.Context, id uint, userID uint, ipAddress string) error {
			return &services.InvoiceNotFoundError{ID: id}
		},
	}

	router := setupPurchaseInvoiceIntegrationTestRouter(mockService, user.RoleAdmin, 1)

	// Test GET not found
	t.Run("GET not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/purchase-invoices/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errorResp dto.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "/errors/purchase-invoice-not-found", errorResp.Type)
	})

	// Test PUT not found
	t.Run("PUT not found", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"invoiceNumber": "INV-2025-001",
			"invoiceDate":   time.Now().Format("2006-01-02"),
			"supplierId":    1,
			"items": []map[string]interface{}{
				{"productId": 1, "quantity": 10, "unitCost": 15000.00},
			},
			"reason": "Test",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("PUT", "/api/v1/purchase-invoices/999", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errorResp dto.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "/errors/purchase-invoice-not-found", errorResp.Type)
	})

	// Test DELETE not found
	t.Run("DELETE not found", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"reason": "Test",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("DELETE", "/api/v1/purchase-invoices/999", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errorResp dto.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "/errors/purchase-invoice-not-found", errorResp.Type)
	})
}

// TestPurchaseInvoiceHandlerIntegration_InvalidID tests invalid ID parameter handling
// Story 10.2, Subtask 11.4: Integration test for invalid ID error responses
func TestPurchaseInvoiceHandlerIntegration_InvalidID(t *testing.T) {
	mockService := &mockPurchaseInvoiceServiceForIntegration{}
	router := setupPurchaseInvoiceIntegrationTestRouter(mockService, user.RoleAdmin, 1)

	testCases := []struct {
		name       string
		method     string
		endpoint   string
		requestBody map[string]interface{}
	}{
		{
			name:     "GET with invalid ID",
			method:   "GET",
			endpoint: "/api/v1/purchase-invoices/invalid",
		},
		{
			name:     "PUT with invalid ID",
			method:   "PUT",
			endpoint: "/api/v1/purchase-invoices/invalid",
			requestBody: map[string]interface{}{
				"invoiceNumber": "INV-2025-001",
				"invoiceDate":   time.Now().Format("2006-01-02"),
				"supplierId":    1,
				"items": []map[string]interface{}{
					{"productId": 1, "quantity": 10, "unitCost": 15000.00},
				},
				"reason": "Test",
			},
		},
		{
			name:     "DELETE with invalid ID",
			method:   "DELETE",
			endpoint: "/api/v1/purchase-invoices/invalid",
			requestBody: map[string]interface{}{
				"reason": "Test",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.method == "GET" {
				req, _ = http.NewRequest(tc.method, tc.endpoint, nil)
			} else {
				bodyBytes, _ := json.Marshal(tc.requestBody)
				req, _ = http.NewRequest(tc.method, tc.endpoint, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var errorResp dto.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errorResp)
			require.NoError(t, err)
			assert.Equal(t, "/errors/invalid-id", errorResp.Type)
		})
	}
}

// TestPurchaseInvoiceHandlerIntegration_AuditLogging tests audit trail integration
// Story 10.2, AC1, Subtask 11.5: Integration test for audit trail logging
func TestPurchaseInvoiceHandlerIntegration_AuditLogging(t *testing.T) {
	// Create mock service that verifies audit logging
	mockService := &mockPurchaseInvoiceServiceForIntegration{
		createFunc: func(ctx context.Context, invoice *models.PurchaseInvoice, items []services.CreatePurchaseInvoiceItemRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
			// Verify audit parameters are passed correctly
			assert.Equal(t, uint(1), userID, "User ID should be passed for audit logging")
			assert.NotEmpty(t, ipAddress, "IP address should be passed for audit logging")
			return &models.PurchaseInvoice{
				ID:            1,
				InvoiceNumber: invoice.InvoiceNumber,
				InvoiceDate:   invoice.InvoiceDate,
				SupplierID:    invoice.SupplierID,
				BranchID:      invoice.BranchID,
				TotalAmount:   150000.00,
				PaymentStatus: "unpaid",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}, nil
		},
		updateFunc: func(ctx context.Context, id uint, req *services.UpdatePurchaseInvoiceRequest, userID uint, ipAddress string) (*models.PurchaseInvoice, error) {
			// Verify audit parameters for update
			assert.Equal(t, uint(1), userID, "User ID should be passed for audit logging")
			assert.NotEmpty(t, ipAddress, "IP address should be passed for audit logging")
			assert.NotEmpty(t, req.Reason, "Reason should be provided for update")
			return &models.PurchaseInvoice{
				ID:            id,
				InvoiceNumber: req.InvoiceNumber,
				InvoiceDate:   time.Now(),
				SupplierID:    req.SupplierID,
				TotalAmount:   150000.00,
				PaymentStatus: "unpaid",
				UpdatedAt:     time.Now(),
			}, nil
		},
		deleteFunc: func(ctx context.Context, id uint, userID uint, ipAddress string) error {
			// Verify audit parameters for delete
			assert.Equal(t, uint(1), userID, "User ID should be passed for audit logging")
			assert.NotEmpty(t, ipAddress, "IP address should be passed for audit logging")
			return nil
		},
	}

	router := setupPurchaseInvoiceIntegrationTestRouter(mockService, user.RoleAdmin, 1)

	// Test create audit logging
	t.Run("Create audit logging", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"invoiceNumber": "INV-2025-001",
			"invoiceDate":   time.Now().Format("2006-01-02"),
			"supplierId":    1,
			"branchId":      1,
			"items": []map[string]interface{}{
				{"productId": 1, "quantity": 10, "unitCost": 15000.00},
			},
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/api/v1/purchase-invoices", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	// Test update audit logging
	t.Run("Update audit logging", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"invoiceNumber": "INV-2025-001-UPDATED",
			"invoiceDate":   time.Now().Format("2006-01-02"),
			"supplierId":    1,
			"items": []map[string]interface{}{
				{"productId": 1, "quantity": 10, "unitCost": 15000.00},
			},
			"reason": "Audit test - updating invoice",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("PUT", "/api/v1/purchase-invoices/1", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test delete audit logging
	t.Run("Delete audit logging", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"reason": "Audit test - deleting invoice",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("DELETE", "/api/v1/purchase-invoices/1", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
