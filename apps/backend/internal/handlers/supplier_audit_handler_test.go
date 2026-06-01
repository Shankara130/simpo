package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// MockSupplierAuditService is a mock implementation of SupplierAuditService
type MockSupplierAuditService struct {
	mock.Mock
}

func (m *MockSupplierAuditService) LogSupplierOperation(ctx context.Context, auditLog *models.SupplierAuditTrail) error {
	args := m.Called(ctx, auditLog)
	return args.Error(0)
}

func (m *MockSupplierAuditService) QueryAuditTrail(ctx context.Context, request *services.SupplierAuditQueryRequest) (*services.SupplierAuditTrailResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.SupplierAuditTrailResponse), args.Error(1)
}

func (m *MockSupplierAuditService) ExportAuditTrail(ctx context.Context, request *services.SupplierAuditExportRequest) ([]models.SupplierAuditTrail, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SupplierAuditTrail), args.Error(1)
}

func (m *MockSupplierAuditService) GetAuditByEntityID(ctx context.Context, entityType string, entityID uint) ([]models.SupplierAuditTrail, error) {
	args := m.Called(ctx, entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SupplierAuditTrail), args.Error(1)
}

func (m *MockSupplierAuditService) GetAuditByUserID(ctx context.Context, userID uint, startDate, endDate time.Time) ([]models.SupplierAuditTrail, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SupplierAuditTrail), args.Error(1)
}

// setupSupplierAuditTestRouter sets up a test router with the supplier audit handler
func setupSupplierAuditTestRouter(handler *SupplierAuditHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/audit/supplier", handler.QueryAuditTrail)
	router.GET("/audit/supplier/entity/:type/:id", handler.GetAuditByEntity)
	router.GET("/audit/supplier/user/:id", handler.GetAuditByUser)
	router.GET("/audit/supplier/export/csv", handler.ExportAuditTrailCSV)
	router.GET("/audit/supplier/export/pdf", handler.ExportAuditTrailPDF)

	return router
}

// TestSupplierAuditHandler_QueryAuditTrail_Success tests successful audit trail query
func TestSupplierAuditHandler_QueryAuditTrail_Success(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	now := time.Now().UTC()

	audits := []models.SupplierAuditTrail{
		{
			ID:                1,
			TransactionType:   "supplier_operation",
			EntityType:       "supplier",
			EntityID:         1,
			UserID:           10,
			UserRole:         "PharmacyManager",
			ActionType:       "CREATE",
			ActionDescription: "Created supplier PT Pharma",
			BranchID:         1,
			CreatedAt:        now,
		},
	}

	response := &services.SupplierAuditTrailResponse{
		Data: audits,
		Pagination: dto.PaginationMeta{
			Page:       1,
			Limit:      10,
			Total:      1,
			TotalPages: 1,
		},
	}

	mockService.On("QueryAuditTrail", mock.Anything, mock.AnythingOfType("*services.SupplierAuditQueryRequest")).Return(response, nil)

	// Act
	req, _ := http.NewRequest("GET", "/audit/supplier?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.NotNil(t, result["data"])

	mockService.AssertExpectations(t)
}

// TestSupplierAuditHandler_QueryAuditTrail_WithFilters tests query with filters
func TestSupplierAuditHandler_QueryAuditTrail_WithFilters(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	response := &services.SupplierAuditTrailResponse{
		Data: []models.SupplierAuditTrail{},
		Pagination: dto.PaginationMeta{
			Page:       1,
			Limit:      20,
			Total:      0,
			TotalPages: 0,
		},
	}

	mockService.On("QueryAuditTrail", mock.Anything, mock.MatchedBy(func(r *services.SupplierAuditQueryRequest) bool {
		return r.TransactionType != nil && *r.TransactionType == "supplier_operation" &&
			r.EntityType != nil && *r.EntityType == "supplier" &&
			r.Page != nil && *r.Page == 1 &&
			r.Limit != nil && *r.Limit == 20
	})).Return(response, nil)

	// Act
	req, _ := http.NewRequest("GET", "/audit/supplier?transaction_type=supplier_operation&entity_type=supplier&page=1&limit=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// TestSupplierAuditHandler_QueryAuditTrail_InvalidDate tests invalid date format
func TestSupplierAuditHandler_QueryAuditTrail_InvalidDate(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	// Act
	req, _ := http.NewRequest("GET", "/audit/supplier?start_date=invalid-date&page=1&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid Request", result["title"])
}

// TestSupplierAuditHandler_QueryAuditTrail_InvalidPagination tests invalid pagination parameters
func TestSupplierAuditHandler_QueryAuditTrail_InvalidPagination(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	// Act - Test invalid page
	req, _ := http.NewRequest("GET", "/audit/supplier?page=invalid&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, "Invalid Request", result["title"])

	// Act - Test limit > 100
	req2, _ := http.NewRequest("GET", "/audit/supplier?page=1&limit=150", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

// TestSupplierAuditHandler_GetAuditByEntity_Success tests successful entity audit retrieval
func TestSupplierAuditHandler_GetAuditByEntity_Success(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	now := time.Now().UTC()
	audits := []models.SupplierAuditTrail{
		{
			ID:                1,
			TransactionType:   "supplier_operation",
			EntityType:       "supplier",
			EntityID:         1,
			UserID:           10,
			UserRole:         "PharmacyManager",
			ActionType:       "CREATE",
			ActionDescription: "Created supplier",
			BranchID:         1,
			CreatedAt:        now,
		},
	}

	mockService.On("GetAuditByEntityID", mock.Anything, "supplier", uint(1)).Return(audits, nil)

	// Act
	req, _ := http.NewRequest("GET", "/audit/supplier/entity/supplier/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.NotNil(t, result["data"])
	assert.Equal(t, float64(1), result["count"])

	mockService.AssertExpectations(t)
}

// TestSupplierAuditHandler_GetAuditByEntity_InvalidEntityType tests invalid entity type
func TestSupplierAuditHandler_GetAuditByEntity_InvalidEntityType(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	// Act
	req, _ := http.NewRequest("GET", "/audit/supplier/entity/invalid_type/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result["detail"], "Invalid entity type")
}

// TestSupplierAuditHandler_GetAuditByUser_Success tests successful user audit retrieval
func TestSupplierAuditHandler_GetAuditByUser_Success(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	now := time.Now().UTC()
	audits := []models.SupplierAuditTrail{
		{
			ID:                1,
			TransactionType:   "supplier_operation",
			EntityType:       "supplier",
			EntityID:         1,
			UserID:           10,
			UserRole:         "PharmacyManager",
			ActionType:       "CREATE",
			ActionDescription: "Created supplier",
			BranchID:         1,
			CreatedAt:        now,
		},
	}

	mockService.On("GetAuditByUserID", mock.Anything, uint(10), mock.Anything, mock.Anything).Return(audits, nil)

	// Act
	req, _ := http.NewRequest("GET", "/audit/supplier/user/10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.NotNil(t, result["data"])
	assert.Equal(t, float64(1), result["count"])

	mockService.AssertExpectations(t)
}

// TestSupplierAuditHandler_GetAuditByUser_WithDateRange tests user audit with date range
func TestSupplierAuditHandler_GetAuditByUser_WithDateRange(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	audits := []models.SupplierAuditTrail{}
	mockService.On("GetAuditByUserID", mock.Anything, uint(10), mock.MatchedBy(func(t time.Time) bool {
		return !t.IsZero()
	}), mock.MatchedBy(func(t time.Time) bool {
		return !t.IsZero()
	})).Return(audits, nil)

	// Act
	req, _ := http.NewRequest("GET", "/audit/supplier/user/10?start_date=2026-01-01&end_date=2026-12-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// TestSupplierAuditHandler_ExportAuditTrailCSV_Success tests successful CSV export
func TestSupplierAuditHandler_ExportAuditTrailCSV_Success(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	audits := []models.SupplierAuditTrail{
		{
			ID:                1,
			TransactionType:   "supplier_operation",
			EntityType:       "supplier",
			EntityID:         1,
			UserID:           10,
			UserRole:         "PharmacyManager",
			ActionType:       "CREATE",
			ActionDescription: "Created supplier PT Pharma",
			BranchID:         1,
			CreatedAt:        time.Now().UTC(),
		},
	}

	mockService.On("ExportAuditTrail", mock.Anything, mock.MatchedBy(func(r *services.SupplierAuditExportRequest) bool {
		return r.Format == "csv"
	})).Return(audits, nil)

	// Act
	req, _ := http.NewRequest("GET", "/audit/supplier/export/csv?start_date=2026-01-01&end_date=2026-12-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=")

	mockService.AssertExpectations(t)
}

// TestSupplierAuditHandler_ExportAuditTrailCSV_MissingDates tests CSV export with missing dates
func TestSupplierAuditHandler_ExportAuditTrailCSV_MissingDates(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	// Act - Missing start date
	req, _ := http.NewRequest("GET", "/audit/supplier/export/csv?end_date=2026-12-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, "Invalid Request", result["title"])

	// Act - Missing end date
	req2, _ := http.NewRequest("GET", "/audit/supplier/export/csv?start_date=2026-01-01", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

// TestSupplierAuditHandler_ExportAuditTrailPDF_Success tests successful PDF export
func TestSupplierAuditHandler_ExportAuditTrailPDF_Success(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	audits := []models.SupplierAuditTrail{
		{
			ID:                1,
			TransactionType:   "supplier_operation",
			EntityType:       "supplier",
			EntityID:         1,
			UserID:           10,
			UserRole:         "PharmacyManager",
			ActionType:       "CREATE",
			ActionDescription: "Created supplier PT Pharma",
			BranchID:         1,
			CreatedAt:        time.Now().UTC(),
		},
	}

	mockService.On("ExportAuditTrail", mock.Anything, mock.MatchedBy(func(r *services.SupplierAuditExportRequest) bool {
		return r.Format == "pdf"
	})).Return(audits, nil)

	// Act
	req, _ := http.NewRequest("GET", "/audit/supplier/export/pdf?start_date=2026-01-01&end_date=2026-12-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - PDF export returns OK with PDF content type
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=")

	mockService.AssertExpectations(t)
}

// TestSupplierAuditHandler_QueryAuditTrail_InvalidEntityID tests invalid entity ID format
func TestSupplierAuditHandler_QueryAuditTrail_InvalidEntityID(t *testing.T) {
	// Arrange
	mockService := new(MockSupplierAuditService)
	handler := NewSupplierAuditHandler(mockService)
	router := setupSupplierAuditTestRouter(handler)

	// Act
	req, _ := http.NewRequest("GET", "/audit/supplier?entity_id=invalid&page=1&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid Request", result["title"])
}
