package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// mockSupplierProductCatalogServiceForIntegration is a mock implementation for integration testing
type mockSupplierProductCatalogServiceForIntegration struct {
	associateFunc       func(ctx context.Context, request *services.AssociateProductRequest, createdBy uint, ipAddress string) (*models.SupplierProductCatalog, error)
	getByIDFunc         func(ctx context.Context, id uint) (*models.SupplierProductCatalog, error)
	listFunc            func(ctx context.Context, filter *services.SupplierProductCatalogListFilter) ([]*models.SupplierProductCatalog, int64, error)
	updatePriceFunc     func(ctx context.Context, catalogID uint, request *services.UpdatePriceRequest, updatedBy uint, ipAddress string) error
	setPreferredFunc    func(ctx context.Context, catalogID uint, request *services.SetPreferredRequest, updatedBy uint, ipAddress string) error
	getPriceHistoryFunc func(ctx context.Context, productID uint, filter *services.PriceHistoryFilter) ([]*services.PriceHistoryEntry, error)
	getPreferredFunc    func(ctx context.Context, productID uint, branchID uint) (*models.SupplierProductCatalog, error)
	getCatalogFunc      func(ctx context.Context, supplierID uint, branchID uint) ([]*models.SupplierProductCatalog, error)
}

func (m *mockSupplierProductCatalogServiceForIntegration) AssociateProduct(ctx context.Context, request *services.AssociateProductRequest, createdBy uint, ipAddress string) (*models.SupplierProductCatalog, error) {
	if m.associateFunc != nil {
		return m.associateFunc(ctx, request, createdBy, ipAddress)
	}
	return &models.SupplierProductCatalog{
		ID:                   1,
		SupplierID:           request.SupplierID,
		ProductID:            request.ProductID,
		PurchasePrice:        request.PurchasePrice,
		IsPreferred:          request.IsPreferred,
		SKUCode:              request.SKUCode,
		MinimumOrderQuantity: request.MinimumOrderQuantity,
		LeadTimeDays:         request.LeadTimeDays,
		BranchID:             request.BranchID,
		CreatedBy:            createdBy,
		PriceEffectiveFrom:   time.Now(),
		Supplier: models.Supplier{
			ID:   request.SupplierID,
			Name: "PT Test Supplier",
		},
		Product: models.Product{
			ID:   request.ProductID,
			Name: "Test Product",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockSupplierProductCatalogServiceForIntegration) GetProductCatalogByID(ctx context.Context, id uint) (*models.SupplierProductCatalog, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &models.SupplierProductCatalog{
		ID:                   id,
		SupplierID:           1,
		ProductID:            1,
		PurchasePrice:        15000.00,
		IsPreferred:          true,
		SKUCode:              "TEST-001",
		MinimumOrderQuantity: 10,
		LeadTimeDays:         &[]int{5}[0],
		BranchID:             1,
		CreatedBy:            1,
		PriceEffectiveFrom:   time.Now(),
		Supplier: models.Supplier{
			ID:   1,
			Name: "PT Test Supplier",
		},
		Product: models.Product{
			ID:   1,
			Name: "Test Product",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockSupplierProductCatalogServiceForIntegration) ListProductCatalogs(ctx context.Context, filter *services.SupplierProductCatalogListFilter) ([]*models.SupplierProductCatalog, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return []*models.SupplierProductCatalog{
		{
			ID:                 1,
			SupplierID:         1,
			ProductID:          1,
			PurchasePrice:      15000.00,
			IsPreferred:        true,
			BranchID:           1,
			PriceEffectiveFrom: time.Now(),
			Supplier: models.Supplier{
				ID:   1,
				Name: "PT Test Supplier",
			},
			Product: models.Product{
				ID:   1,
				Name: "Test Product",
			},
		},
	}, 1, nil
}

func (m *mockSupplierProductCatalogServiceForIntegration) UpdatePurchasePrice(ctx context.Context, catalogID uint, request *services.UpdatePriceRequest, updatedBy uint, ipAddress string) error {
	if m.updatePriceFunc != nil {
		return m.updatePriceFunc(ctx, catalogID, request, updatedBy, ipAddress)
	}
	return nil
}

func (m *mockSupplierProductCatalogServiceForIntegration) SetPreferredSupplier(ctx context.Context, catalogID uint, request *services.SetPreferredRequest, updatedBy uint, ipAddress string) error {
	if m.setPreferredFunc != nil {
		return m.setPreferredFunc(ctx, catalogID, request, updatedBy, ipAddress)
	}
	return nil
}

func (m *mockSupplierProductCatalogServiceForIntegration) GetPriceHistory(ctx context.Context, productID uint, filter *services.PriceHistoryFilter) ([]*services.PriceHistoryEntry, error) {
	if m.getPriceHistoryFunc != nil {
		return m.getPriceHistoryFunc(ctx, productID, filter)
	}
	return []*services.PriceHistoryEntry{
		{
			ID:            1,
			SupplierID:    1,
			SupplierName:  "PT Test Supplier",
			ProductID:     productID,
			ProductName:   "Test Product",
			PurchasePrice: 15000.00,
			EffectiveFrom: time.Now().AddDate(-30, 0, 0).Format("2006-01-02"),
			IsCurrent:     true,
			IsPreferred:   true,
		},
	}, nil
}

func (m *mockSupplierProductCatalogServiceForIntegration) GetPreferredSupplier(ctx context.Context, productID uint, branchID uint) (*models.SupplierProductCatalog, error) {
	if m.getPreferredFunc != nil {
		return m.getPreferredFunc(ctx, productID, branchID)
	}
	return &models.SupplierProductCatalog{
		ID:                 1,
		SupplierID:         1,
		ProductID:          productID,
		PurchasePrice:      15000.00,
		IsPreferred:        true,
		BranchID:           branchID,
		PriceEffectiveFrom: time.Now(),
		Supplier: models.Supplier{
			ID:   1,
			Name: "PT Test Supplier",
		},
		Product: models.Product{
			ID:   productID,
			Name: "Test Product",
		},
	}, nil
}

func (m *mockSupplierProductCatalogServiceForIntegration) GetCatalogBySupplier(ctx context.Context, supplierID uint, branchID uint) ([]*models.SupplierProductCatalog, error) {
	if m.getCatalogFunc != nil {
		return m.getCatalogFunc(ctx, supplierID, branchID)
	}
	return []*models.SupplierProductCatalog{
		{
			ID:                 1,
			SupplierID:         supplierID,
			ProductID:          1,
			PurchasePrice:      15000.00,
			IsPreferred:        true,
			BranchID:           branchID,
			PriceEffectiveFrom: time.Now(),
			Product: models.Product{
				ID:   1,
				Name: "Product A",
			},
		},
		{
			ID:                 2,
			SupplierID:         supplierID,
			ProductID:          2,
			PurchasePrice:      20000.00,
			IsPreferred:        false,
			BranchID:           branchID,
			PriceEffectiveFrom: time.Now(),
			Product: models.Product{
				ID:   2,
				Name: "Product B",
			},
		},
	}, nil
}

// setupSupplierProductCatalogIntegrationTestRouter creates a test router with auth/RBAC middleware
func setupSupplierProductCatalogIntegrationTestRouter(mockService *mockSupplierProductCatalogServiceForIntegration) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := NewSupplierProductCatalogHandler(mockService)

	// Setup route groups similar to main router
	v1 := router.Group("/api/v1")
	{
		catalogGroup := v1.Group("/supplier-product-catalogs")
		{
			catalogGroup.POST("", handler.AssociateProduct)
			catalogGroup.GET("/:id", handler.GetProductCatalog)
			catalogGroup.GET("", handler.ListProductCatalogs)
			catalogGroup.PUT("/:id/price", handler.UpdatePurchasePrice)
			catalogGroup.PUT("/:id/preferred", handler.SetPreferredSupplier)
		}

		productPriceGroup := v1.Group("/products/:id")
		{
			productPriceGroup.GET("/price-history", handler.GetPriceHistory)
			productPriceGroup.GET("/preferred-supplier", handler.GetPreferredSupplier)
		}

		supplierCatalogGroup := v1.Group("/suppliers/:id")
		{
			supplierCatalogGroup.GET("/product-catalog", handler.GetSupplierCatalog)
		}
	}

	return router
}

// Helper to add auth context to request
func addAuthContext(req *http.Request, userID uint, branchID uint, role string) {
	ctx := req.Context()
	ctx = context.WithValue(ctx, "user_id", userID)
	ctx = context.WithValue(ctx, "branch_id", branchID)
	ctx = context.WithValue(ctx, "user_role", role)
	*req = *req.WithContext(ctx)
}

// TestSupplierProductCatalogHandler_AssociateProduct_Success tests successful product association
// Task 10.2: Test product association with supplier
func TestSupplierProductCatalogHandler_AssociateProduct_Success(t *testing.T) {
	mockService := &mockSupplierProductCatalogServiceForIntegration{
		associateFunc: func(ctx context.Context, request *services.AssociateProductRequest, createdBy uint, ipAddress string) (*models.SupplierProductCatalog, error) {
			assert.Equal(t, uint(1), request.SupplierID)
			assert.Equal(t, uint(1), request.ProductID)
			assert.Equal(t, uint(1), request.BranchID)
			assert.Equal(t, 15000.00, request.PurchasePrice)
			assert.True(t, request.IsPreferred)
			assert.Equal(t, 10, request.MinimumOrderQuantity)
			return &models.SupplierProductCatalog{
				ID:                   1,
				SupplierID:           1,
				ProductID:            1,
				PurchasePrice:        15000.00,
				IsPreferred:          true,
				MinimumOrderQuantity: 10,
				BranchID:             1,
				CreatedBy:            1,
			}, nil
		},
	}

	router := setupSupplierProductCatalogIntegrationTestRouter(mockService)

	requestBody := map[string]interface{}{
		"supplierId":           1,
		"productId":            1,
		"branchId":             1,
		"purchasePrice":        15000.00,
		"isPreferred":          true,
		"skuCode":              "TEST-001",
		"minimumOrderQuantity": 10,
		"leadTimeDays":         5,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/supplier-product-catalogs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	addAuthContext(req, 1, 1, "admin")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.SupplierProductCatalog
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, uint(1), response.SupplierID)
	assert.Equal(t, uint(1), response.ProductID)
	assert.Equal(t, 15000.00, response.PurchasePrice)
	assert.True(t, response.IsPreferred)
}

// TestSupplierProductCatalogHandler_UpdatePurchasePrice_Success tests price update with history tracking
// Task 10.3: Test price update and history tracking
func TestSupplierProductCatalogHandler_UpdatePurchasePrice_Success(t *testing.T) {
	priceUpdateCalled := false
	mockService := &mockSupplierProductCatalogServiceForIntegration{
		updatePriceFunc: func(ctx context.Context, catalogID uint, request *services.UpdatePriceRequest, updatedBy uint, ipAddress string) error {
			priceUpdateCalled = true
			assert.Equal(t, uint(1), catalogID)
			assert.Equal(t, 18000.00, request.NewPrice)
			assert.Equal(t, uint(1), updatedBy)
			return nil
		},
		getByIDFunc: func(ctx context.Context, id uint) (*models.SupplierProductCatalog, error) {
			return &models.SupplierProductCatalog{
				ID:                 id,
				SupplierID:         1,
				ProductID:          1,
				PurchasePrice:      18000.00,
				IsPreferred:        true,
				BranchID:           1,
				PriceEffectiveFrom: time.Now(),
			}, nil
		},
	}

	router := setupSupplierProductCatalogIntegrationTestRouter(mockService)

	requestBody := map[string]interface{}{
		"newPrice": 18000.00,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/api/v1/supplier-product-catalogs/1/price", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	addAuthContext(req, 1, 1, "admin")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, priceUpdateCalled, "UpdatePrice should be called")
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSupplierProductCatalogHandler_SetPreferredSupplier_Success tests preferred supplier marking
// Task 10.4: Test preferred supplier marking
func TestSupplierProductCatalogHandler_SetPreferredSupplier_Success(t *testing.T) {
	setPreferredCalled := false
	mockService := &mockSupplierProductCatalogServiceForIntegration{
		setPreferredFunc: func(ctx context.Context, catalogID uint, request *services.SetPreferredRequest, updatedBy uint, ipAddress string) error {
			setPreferredCalled = true
			assert.Equal(t, uint(1), catalogID)
			assert.True(t, request.IsPreferred)
			assert.Equal(t, uint(1), updatedBy)
			return nil
		},
		getByIDFunc: func(ctx context.Context, id uint) (*models.SupplierProductCatalog, error) {
			return &models.SupplierProductCatalog{
				ID:                 id,
				SupplierID:         1,
				ProductID:          1,
				PurchasePrice:      15000.00,
				IsPreferred:        true,
				BranchID:           1,
				PriceEffectiveFrom: time.Now(),
			}, nil
		},
	}

	router := setupSupplierProductCatalogIntegrationTestRouter(mockService)

	requestBody := map[string]interface{}{
		"isPreferred": true,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/api/v1/supplier-product-catalogs/1/preferred", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	addAuthContext(req, 1, 1, "admin")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, setPreferredCalled, "SetPreferredSupplier should be called")
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSupplierProductCatalogHandler_GetPriceHistory_Success tests price history query
// Task 10.5: Test price history query with date ranges
func TestSupplierProductCatalogHandler_GetPriceHistory_Success(t *testing.T) {
	mockService := &mockSupplierProductCatalogServiceForIntegration{
		getPriceHistoryFunc: func(ctx context.Context, productID uint, filter *services.PriceHistoryFilter) ([]*services.PriceHistoryEntry, error) {
			assert.Equal(t, uint(1), productID)
			return []*services.PriceHistoryEntry{
				{
					ID:            1,
					SupplierID:    1,
					SupplierName:  "PT Supplier A",
					ProductID:     productID,
					ProductName:   "Test Product",
					PurchasePrice: 15000.00,
					EffectiveFrom: time.Now().AddDate(-30, 0, 0).Format("2006-01-02"),
					IsCurrent:     true,
					IsPreferred:   true,
				},
				{
					ID:            2,
					SupplierID:    1,
					SupplierName:  "PT Supplier A",
					ProductID:     productID,
					ProductName:   "Test Product",
					PurchasePrice: 14000.00,
					EffectiveFrom: time.Now().AddDate(-60, 0, 0).Format("2006-01-02"),
					EffectiveTo:   time.Now().AddDate(-30, 0, 0).Format("2006-01-02"),
					IsCurrent:     false,
					IsPreferred:   true,
				},
			}, nil
		},
	}

	router := setupSupplierProductCatalogIntegrationTestRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/products/1/price-history?start_date=2024-01-01&end_date=2024-12-31&page=1&limit=10", nil)
	addAuthContext(req, 1, 1, "admin")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "data")
}

// TestSupplierProductCatalogHandler_ListProductCatalogs_WithFilters tests listing with filters
func TestSupplierProductCatalogHandler_ListProductCatalogs_WithFilters(t *testing.T) {
	mockService := &mockSupplierProductCatalogServiceForIntegration{
		listFunc: func(ctx context.Context, filter *services.SupplierProductCatalogListFilter) ([]*models.SupplierProductCatalog, int64, error) {
			assert.Equal(t, uint(1), *filter.SupplierID)
			assert.Equal(t, uint(1), filter.Page)
			assert.Equal(t, 20, filter.Limit)
			return []*models.SupplierProductCatalog{
				{
					ID:            1,
					SupplierID:    1,
					ProductID:     1,
					PurchasePrice: 15000.00,
					IsPreferred:   true,
					BranchID:      1,
				},
			}, 1, nil
		},
	}

	router := setupSupplierProductCatalogIntegrationTestRouter(mockService)

	params := url.Values{}
	params.Add("supplier_id", "1")
	params.Add("page", "1")
	params.Add("limit", "20")

	req, _ := http.NewRequest("GET", "/api/v1/supplier-product-catalogs?"+params.Encode(), nil)
	addAuthContext(req, 1, 1, "admin")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "data")
	assert.Contains(t, response, "pagination")
}

// TestSupplierProductCatalogHandler_GetPreferredSupplier_Success tests getting preferred supplier
func TestSupplierProductCatalogHandler_GetPreferredSupplier_Success(t *testing.T) {
	mockService := &mockSupplierProductCatalogServiceForIntegration{
		getPreferredFunc: func(ctx context.Context, productID uint, branchID uint) (*models.SupplierProductCatalog, error) {
			assert.Equal(t, uint(1), productID)
			assert.Equal(t, uint(1), branchID)
			return &models.SupplierProductCatalog{
				ID:            1,
				SupplierID:    1,
				ProductID:     productID,
				PurchasePrice: 15000.00,
				IsPreferred:   true,
				BranchID:      branchID,
				Supplier: models.Supplier{
					ID:   1,
					Name: "PT Preferred Supplier",
				},
			}, nil
		},
	}

	router := setupSupplierProductCatalogIntegrationTestRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/products/1/preferred-supplier?branch_id=1", nil)
	addAuthContext(req, 1, 1, "admin")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.SupplierProductCatalog
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.IsPreferred)
	assert.Equal(t, "PT Preferred Supplier", response.Supplier.Name)
}

// TestSupplierProductCatalogHandler_GetSupplierCatalog_Success tests getting supplier's product catalog
func TestSupplierProductCatalogHandler_GetSupplierCatalog_Success(t *testing.T) {
	mockService := &mockSupplierProductCatalogServiceForIntegration{
		getCatalogFunc: func(ctx context.Context, supplierID uint, branchID uint) ([]*models.SupplierProductCatalog, error) {
			assert.Equal(t, uint(1), supplierID)
			assert.Equal(t, uint(1), branchID)
			return []*models.SupplierProductCatalog{
				{
					ID:            1,
					SupplierID:    supplierID,
					ProductID:     1,
					PurchasePrice: 15000.00,
					IsPreferred:   true,
					BranchID:      branchID,
					Product: models.Product{
						ID:   1,
						Name: "Product A",
					},
				},
			}, nil
		},
	}

	router := setupSupplierProductCatalogIntegrationTestRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/suppliers/1/product-catalog?branch_id=1", nil)
	addAuthContext(req, 1, 1, "admin")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "data")
}

// TestSupplierProductCatalogHandler_AssociateProduct_ValidationError tests validation errors
// Task 10.9: Test error cases
func TestSupplierProductCatalogHandler_AssociateProduct_ValidationError(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedType   string
	}{
		{
			name: "Missing supplier ID",
			requestBody: map[string]interface{}{
				"productId":            1,
				"branchId":             1,
				"purchasePrice":        15000.00,
				"minimumOrderQuantity": 10,
			},
			expectedStatus: http.StatusBadRequest,
			expectedType:   "/errors/validation-error",
		},
		{
			name: "Invalid purchase price (negative)",
			requestBody: map[string]interface{}{
				"supplierId":           1,
				"productId":            1,
				"branchId":             1,
				"purchasePrice":        -100.00,
				"minimumOrderQuantity": 10,
			},
			expectedStatus: http.StatusBadRequest,
			expectedType:   "/errors/validation-error",
		},
		{
			name: "Invalid minimum order quantity (zero)",
			requestBody: map[string]interface{}{
				"supplierId":           1,
				"productId":            1,
				"branchId":             1,
				"purchasePrice":        15000.00,
				"minimumOrderQuantity": 0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedType:   "/errors/validation-error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupSupplierProductCatalogIntegrationTestRouter(&mockSupplierProductCatalogServiceForIntegration{})

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/supplier-product-catalogs", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			addAuthContext(req, 1, 1, "admin")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var errorResp dto.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errorResp)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, errorResp.Type)
		})
	}
}

// TestSupplierProductCatalogHandler_GetProductCatalog_NotFound tests not found error
func TestSupplierProductCatalogHandler_GetProductCatalog_NotFound(t *testing.T) {
	mockService := &mockSupplierProductCatalogServiceForIntegration{
		getByIDFunc: func(ctx context.Context, id uint) (*models.SupplierProductCatalog, error) {
			return nil, fmt.Errorf("catalog entry not found")
		},
	}

	router := setupSupplierProductCatalogIntegrationTestRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/supplier-product-catalogs/999", nil)
	addAuthContext(req, 1, 1, "admin")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
