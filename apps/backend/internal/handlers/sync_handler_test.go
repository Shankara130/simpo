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
	"github.com/stretchr/testify/mock"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// MockProductRepoForConflict is a mock for product repository
type MockProductRepoForConflict struct {
	mock.Mock
}

func (m *MockProductRepoForConflict) GetStockForProduct(ctx context.Context, productID uint) (int64, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).(int64), args.Error(1)
}

// MockConflictAudit is a mock for conflict audit service
type MockConflictAudit struct {
	mock.Mock
}

func (m *MockConflictAudit) LogConflictResolution(ctx context.Context, log services.ConflictResolutionLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// MockSyncService is a mock for SyncService
type MockSyncService struct {
	mock.Mock
}

func (m *MockSyncService) ProcessSyncQueue(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSyncService) QueueTransactionSync(ctx context.Context, transactionID uint) error {
	args := m.Called(ctx, transactionID)
	return args.Error(1)
}

func (m *MockSyncService) ResolveConflict(ctx context.Context, conflictID uint, resolution interface{}) error {
	args := m.Called(ctx, conflictID, resolution)
	return args.Error(1)
}

// TestSyncHandler_SuccessfulSync tests successful transaction sync
func TestSyncHandler_SuccessfulSync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockProductRepoForConflict)
	mockAudit := new(MockConflictAudit)
	mockConflictSvc := services.NewConflictResolutionService(mockRepo, mockAudit)
	mockSyncSvc := new(MockSyncService)
	handler := NewSyncHandler(mockConflictSvc, mockSyncSvc)

	router := gin.New()
	router.POST("/api/v1/sync", handler.SyncTransactions)

	// Prepare test request - transactions without items will pass validation
	transactions := []services.OfflineTransaction{
		{ID: 1, TransactionNumber: "TRX-001", Timestamp: time.Now(), Items: []services.TransactionItem{}},
		{ID: 2, TransactionNumber: "TRX-002", Timestamp: time.Now(), Items: []services.TransactionItem{}},
	}

	body, _ := json.Marshal(map[string][]services.OfflineTransaction{"transactions": transactions})
	req, _ := http.NewRequest("POST", "/api/v1/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SyncResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Transactions synchronized successfully", response.Message)
	assert.Len(t, response.SuccessfulTransactions, 2)
	assert.Len(t, response.FailedTransactions, 0)
	assert.Empty(t, response.ConflictErrors)
}

// TestSyncHandler_ConflictError tests insufficient stock conflict response
func TestSyncHandler_ConflictError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock repository to return insufficient stock
	mockRepo := new(MockProductRepoForConflict)
	mockAudit := new(MockConflictAudit)
	mockConflictSvc := services.NewConflictResolutionService(mockRepo, mockAudit)

	// Mock: Product 1 has only 50 units in stock
	mockRepo.On("GetStockForProduct", mock.Anything, uint(1)).Return(int64(50), nil)

	mockSyncSvc := new(MockSyncService)
	handler := NewSyncHandler(mockConflictSvc, mockSyncSvc)

	router := gin.New()
	router.POST("/api/v1/sync", handler.SyncTransactions)

	// Prepare test request with items that will conflict
	transactions := []services.OfflineTransaction{
		{
			ID: 1, TransactionNumber: "TRX-001",
			Items: []services.TransactionItem{
				{ProductID: 1, ProductSKU: "SKU-001", Quantity: 100},
			},
		},
	}

	body, _ := json.Marshal(map[string][]services.OfflineTransaction{"transactions": transactions})
	req, _ := http.NewRequest("POST", "/api/v1/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response SyncResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Conflict detected during transaction sync", response.Message)
	assert.Len(t, response.FailedTransactions, 1)
	assert.Len(t, response.ConflictErrors, 1)
	assert.Equal(t, "TRX-001", response.ConflictErrors[0].TransactionID)
	assert.Equal(t, "SKU-001", response.ConflictErrors[0].ConflictDetails.ProductSKU)
	assert.Equal(t, 409, response.ConflictErrors[0].Status)

	mockRepo.AssertExpectations(t)
}
