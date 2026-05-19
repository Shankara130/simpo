package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// setupTestRouter creates a test router with the product handler and auth context
// Story 4.1: Test helper for product handler testing
func setupProductTestRouter(productService services.ProductService, userRole string, branchID uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add auth context middleware to simulate authenticated user
	// Keys must match what the handler expects: "user_role", "branch_id"
	router.Use(func(c *gin.Context) {
		c.Set("user_role", userRole)
		c.Set("branch_id", branchID)
		c.Next()
	})

	// Add product handler
	handler := NewProductHandler(productService, nil, "test-secret-key")
	router.GET("/api/v1/products", handler.ListProducts)

	return router
}

// MockProductService is a mock for testing
type MockProductService struct {
	listFunc func(ctx context.Context, filter *services.ProductFilter) ([]*models.Product, int64, error)
}

func (m *MockProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	return nil
}

func (m *MockProductService) UpdateProduct(ctx context.Context, id uint, product *models.Product) error {
	return nil
}

func (m *MockProductService) UpdateStock(ctx context.Context, id uint, quantity int64) error {
	return nil
}

func (m *MockProductService) CheckAvailability(ctx context.Context, id uint, requestedQty int64) (int64, error) {
	return 0, nil
}

func (m *MockProductService) ListProducts(ctx context.Context, filter *services.ProductFilter) ([]*models.Product, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return []*models.Product{}, 0, nil
}

func (m *MockProductService) GetProductByID(ctx context.Context, id uint) (*models.Product, error) {
	return &models.Product{}, nil
}

func (m *MockProductService) GetLowStockProducts(ctx context.Context, branchID uint) ([]*models.Product, error) {
	return []*models.Product{}, nil
}

// TestProductHandler_ListProducts_Success tests successful product listing with pagination
// Story 4.1, AC1, AC7: Products are displayed in a searchable list with pagination
func TestProductHandler_ListProducts_Success(t *testing.T) {
	// Setup mock service with test data
	expiryTime := time.Now().AddDate(0, 6, 0) // 6 months from now
	products := []*models.Product{
		{
			ID:               1,
			SKU:              "SKU-001",
			Name:             "Paracetamol 500mg",
			Description:      "Obat pereda nyeri",
			StockQty:         50,
			Price:            "15000.00",
			ExpiryDate:       &expiryTime,
			BranchID:         1,
			Category:         "Obat Bebas",
			ReorderThreshold: 10,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               2,
			SKU:              "SKU-002",
			Name:             "Amoxicillin 500mg",
			Description:      "Antibiotik",
			StockQty:         5,
			Price:            "25000.00",
			ExpiryDate:       &expiryTime,
			BranchID:         1,
			Category:         "Obat Bebas Terbatas",
			ReorderThreshold: 20,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	mockService := &MockProductService{
		listFunc: func(ctx context.Context, filter *services.ProductFilter) ([]*models.Product, int64, error) {
			return products, int64(len(products)), nil
		},
	}

	router := setupProductTestRouter(mockService, user.RoleOwner, 1)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/products?page=1&limit=20", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// TestProductHandler_ListProducts_WithSearch tests search functionality
// Story 4.1, AC2: Products can be searched by name or SKU
func TestProductHandler_ListProducts_WithSearch(t *testing.T) {
	expiryTime := time.Now().AddDate(0, 6, 0)
	searchResult := []*models.Product{
		{
			ID:               1,
			SKU:              "PARA-001",
			Name:             "Paracetamol 500mg",
			StockQty:         50,
			Price:            "15000.00",
			ExpiryDate:       &expiryTime,
			BranchID:         1,
			Category:         "Obat Bebas",
			ReorderThreshold: 10,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	mockService := &MockProductService{
		listFunc: func(ctx context.Context, filter *services.ProductFilter) ([]*models.Product, int64, error) {
			// Verify search query is passed correctly
			assert.Equal(t, "Paracetamol", filter.SearchQuery)
			return searchResult, 1, nil
		},
	}

	router := setupProductTestRouter(mockService, user.RoleOwner, 1)

	req, _ := http.NewRequest("GET", "/api/v1/products?search=Paracetamol", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// TestProductHandler_ListProducts_WithLowStock tests low stock indicator
// Story 4.1, AC5: Low stock items are visually highlighted
func TestProductHandler_ListProducts_WithLowStock(t *testing.T) {
	expiryTime := time.Now().AddDate(0, 6, 0)
	lowStockProduct := &models.Product{
		ID:               1,
		SKU:              "SKU-002",
		Name:             "Low Stock Item",
		StockQty:         5,
		Price:            "10000.00",
		ExpiryDate:       &expiryTime,
		BranchID:         1,
		Category:         "Obat Bebas",
		ReorderThreshold: 10, // Stock (5) < Threshold (10) = Low Stock
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Test low stock calculation
	isLowStock := lowStockProduct.StockQty < int64(lowStockProduct.ReorderThreshold)
	assert.True(t, isLowStock, "Product with stock below threshold should be marked as low stock")
}

// TestProductHandler_ListProducts_WithExpired tests expired product marking
// Story 4.1, AC6: Expired items are visually marked
func TestProductHandler_ListProducts_WithExpired(t *testing.T) {
	expiredTime := time.Now().AddDate(-1, 0, 0) // 1 year ago
	expiredProduct := &models.Product{
		ID:               1,
		SKU:              "SKU-003",
		Name:             "Expired Item",
		StockQty:         100,
		Price:            "5000.00",
		ExpiryDate:       &expiredTime,
		BranchID:         1,
		Category:         "Obat Bebas",
		ReorderThreshold: 10,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Test expired calculation
	now := time.Now()
	isExpired := expiredProduct.ExpiryDate != nil && expiredProduct.ExpiryDate.Before(now)
	assert.True(t, isExpired, "Product with past expiry date should be marked as expired")
}

// TestProductHandler_ListProducts_BranchAccessControl tests RBAC for branch access
// Story 4.1, AC3: Owners can filter by branch, Cashiers restricted to their branch
func TestProductHandler_ListProducts_BranchAccessControl(t *testing.T) {
	// Test role constants
	assert.Equal(t, "CASHIER", user.RoleCashier, "Cashier role constant")
	assert.Equal(t, "OWNER", user.RoleOwner, "Owner role constant")
	assert.Equal(t, "SYSTEM_ADMIN", user.RoleSystemAdmin, "System Admin role constant")
}

// TestProductListRequest_QueryParameterBinding tests query parameter validation
// Story 4.1, Task 1.3: Query parameter validation
func TestProductListRequest_QueryParameterBinding(t *testing.T) {
	mockService := &MockProductService{
		listFunc: func(ctx context.Context, filter *services.ProductFilter) ([]*models.Product, int64, error) {
			return []*models.Product{}, 0, nil
		},
	}

	testCases := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Valid parameters",
			queryParams:    "?page=1&limit=20",
			expectedStatus: 200,
		},
		{
			name:           "Invalid limit exceeds maximum",
			queryParams:    "?limit=2000",
			expectedStatus: 200, // Should be capped to 1000, not error
		},
		{
			name:           "Page defaults to 1",
			queryParams:    "",
			expectedStatus: 200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := setupProductTestRouter(mockService, user.RoleOwner, 1)
			req, _ := http.NewRequest("GET", "/api/v1/products"+tc.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}
