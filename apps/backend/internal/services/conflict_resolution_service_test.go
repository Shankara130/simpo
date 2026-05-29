package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// MockProductRepository is a mock for the product repository
type MockProductRepositoryForConflict struct {
	mock.Mock
}

func (m *MockProductRepositoryForConflict) GetStockForProduct(ctx context.Context, productID uint) (int64, error) {
	args := m.Called(productID)
	if args.Error(1) != nil {
		return 0, args.Error(1)
	}
	return args.Get(0).(int64), nil
}

// MockConflictAuditService is a mock for the conflict audit service
type MockConflictAuditService struct {
	mock.Mock
}

func (m *MockConflictAuditService) LogConflictResolution(ctx context.Context, log ConflictResolutionLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// TestProcessBatchWithValidation_Success tests successful batch processing
func TestConflictResolutionService_ProcessBatchWithValidation_Success(t *testing.T) {
	// GREEN phase: Now service is implemented, test with proper setup
	mockRepo := new(MockProductRepositoryForConflict)
	mockAudit := new(MockConflictAuditService)
	service := NewConflictResolutionService(mockRepo, mockAudit)

	transactions := []OfflineTransaction{
		{ID: 1, TransactionNumber: "TRX-001", Timestamp: time.Now().Add(-2 * time.Hour)},
		{ID: 2, TransactionNumber: "TRX-002", Timestamp: time.Now().Add(-1 * time.Hour)},
	}

	results, errors := service.ProcessBatchWithValidation(context.Background(), transactions)

	// GREEN phase: Service processes transactions successfully
	assert.NotNil(t, results, "Results should not be nil")
	assert.NotNil(t, errors, "Errors should not be nil")
	assert.Len(t, results.SuccessfulTransactions, 2, "Both transactions should succeed (no items to validate)")
	assert.Empty(t, results.FailedTransactions, "No failed transactions")
	assert.Empty(t, results.ConflictErrors, "No conflict errors")
}

// TestValidateStockAvailability tests stock validation logic
func TestConflictResolutionService_ValidateStockAvailability(t *testing.T) {
	// GREEN phase: Test with proper mock setup
	mockRepo := new(MockProductRepositoryForConflict)
	mockAudit := new(MockConflictAuditService)
	service := NewConflictResolutionService(mockRepo, mockAudit)

	ctx := context.Background()
	transaction := OfflineTransaction{
		ID:                1,
		TransactionNumber: "TRX-001",
		Items: []TransactionItem{
			{ProductID: 1, ProductSKU: "SKU-001", Quantity: 5},
		},
	}

	batchStock := map[uint]int64{
		1: 100,
	}

	// Mock repository call - simulate stock available
	mockRepo.On("GetStockForProduct", uint(1)).Return(int64(150), nil)

	sufficient, details := service.ValidateStockAvailability(ctx, transaction, batchStock)

	// GREEN phase: Validation should succeed (150 - 100 = 50 available, need 5)
	assert.True(t, sufficient, "Should have sufficient stock")
	assert.Nil(t, details, "No conflict details when stock is sufficient")
	mockRepo.AssertExpectations(t)
}

// TestBuildConflictErrorResponse tests RFC 8807 error response building
func TestConflictResolutionService_BuildConflictErrorResponse(t *testing.T) {
	mockRepo := new(MockProductRepositoryForConflict)
	mockAudit := new(MockConflictAuditService)
	service := NewConflictResolutionService(mockRepo, mockAudit)

	details := dto.ConflictDetails{
		ProductID:         123,
		ProductSKU:        "SKU-12345",
		RequestedQuantity: 10,
		AvailableStock:    5,
		Shortfall:          5,
	}

	response := service.BuildConflictErrorResponse(details, "TRX-TEST")

	assert.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, "https://api.simpo.com/errors/conflict-insufficient-stock", response.Type)
	assert.Equal(t, "Insufficient Stock", response.Title)
	assert.Equal(t, 409, response.Status)
	assert.Contains(t, response.Detail, "SKU-12345", "Detail should contain product SKU")
	assert.Contains(t, response.Detail, "10", "Detail should contain requested quantity")
	assert.Contains(t, response.Detail, "5", "Detail should contain available stock")
}

// TestValidateStockAvailability_InsufficientStock tests insufficient stock scenario
func TestConflictResolutionService_ValidateStockAvailability_InsufficientStock(t *testing.T) {
	mockRepo := new(MockProductRepositoryForConflict)
	mockAudit := new(MockConflictAuditService)
	service := NewConflictResolutionService(mockRepo, mockAudit)

	ctx := context.Background()
	transaction := OfflineTransaction{
		ID:                1,
		TransactionNumber: "TRX-001",
		Items: []TransactionItem{
			{ProductID: 1, ProductSKU: "SKU-001", Quantity: 50},
		},
	}

	batchStock := map[uint]int64{
		1: 100,
	}

	// Mock repository call - simulate insufficient stock
	mockRepo.On("GetStockForProduct", uint(1)).Return(int64(140), nil)

	sufficient, details := service.ValidateStockAvailability(ctx, transaction, batchStock)

	// Should fail: 140 (DB) - 100 (batch) = 40 available, need 50
	assert.False(t, sufficient, "Should NOT have sufficient stock")
	assert.NotNil(t, details, "Conflict details should be provided")
	assert.Equal(t, uint(1), details.ProductID)
	assert.Equal(t, "SKU-001", details.ProductSKU)
	assert.Equal(t, 50, details.RequestedQuantity)
	assert.Equal(t, int64(40), details.AvailableStock)
	assert.Equal(t, 10, details.Shortfall)
	mockRepo.AssertExpectations(t)
}

// TestSortTransactionsChronologically tests chronological ordering
func TestConflictResolutionService_SortTransactionsChronologically(t *testing.T) {
	mockRepo := new(MockProductRepositoryForConflict)
	mockAudit := new(MockConflictAuditService)
	service := NewConflictResolutionService(mockRepo, mockAudit)

	now := time.Now()
	transactions := []OfflineTransaction{
		{ID: 1, TransactionNumber: "TRX-001", Timestamp: now.Add(-2 * time.Hour)},
		{ID: 3, TransactionNumber: "TRX-003", Timestamp: now},
		{ID: 2, TransactionNumber: "TRX-002", Timestamp: now.Add(-1 * time.Hour)},
	}

	// Process batch should sort chronologically (oldest first)
	results, _ := service.ProcessBatchWithValidation(context.Background(), transactions)

	// Verify chronological order
	assert.Len(t, results.SuccessfulTransactions, 3, "All transactions should succeed")
	assert.Equal(t, "TRX-001", results.SuccessfulTransactions[0].TransactionNumber, "Oldest should be first")
	assert.Equal(t, "TRX-002", results.SuccessfulTransactions[1].TransactionNumber, "Middle should be second")
	assert.Equal(t, "TRX-003", results.SuccessfulTransactions[2].TransactionNumber, "Newest should be last")
}

// TestBatchContextValidation tests running batch stock counter
func TestConflictResolutionService_BatchContextValidation(t *testing.T) {
	mockRepo := new(MockProductRepositoryForConflict)
	mockAudit := new(MockConflictAuditService)
	service := NewConflictResolutionService(mockRepo, mockAudit)

	ctx := context.Background()
	now := time.Now()

	// Product 1 has 100 units in DB
	mockRepo.On("GetStockForProduct", uint(1)).Return(int64(100), nil).Times(3)

	transactions := []OfflineTransaction{
		{
			ID: 1, TransactionNumber: "TRX-001", Timestamp: now.Add(-2 * time.Hour),
			Items: []TransactionItem{{ProductID: 1, ProductSKU: "SKU-001", Quantity: 30}},
		},
		{
			ID: 2, TransactionNumber: "TRX-002", Timestamp: now.Add(-1 * time.Hour),
			Items: []TransactionItem{{ProductID: 1, ProductSKU: "SKU-001", Quantity: 40}},
		},
		{
			ID: 3, TransactionNumber: "TRX-003", Timestamp: now,
			Items: []TransactionItem{{ProductID: 1, ProductSKU: "SKU-001", Quantity: 35}},
		},
	}

	results, _ := service.ProcessBatchWithValidation(ctx, transactions)

	// TX1: 100 - 0 = 100 available, need 30 ✓
	// TX2: 100 - 30 = 70 available, need 40 ✓
	// TX3: 100 - 70 = 30 available, need 35 ✗ FAIL
	assert.Len(t, results.SuccessfulTransactions, 2, "First two transactions should succeed")
	assert.Len(t, results.FailedTransactions, 1, "Third transaction should fail")
	assert.Len(t, results.ConflictErrors, 1, "Should have one conflict error")
	assert.Equal(t, "TRX-001", results.SuccessfulTransactions[0].TransactionNumber)
	assert.Equal(t, "TRX-002", results.SuccessfulTransactions[1].TransactionNumber)
	assert.Equal(t, "TRX-003", results.FailedTransactions[0].TransactionNumber)

	// Verify conflict details for failed transaction
	errorResp := results.ConflictErrors[0]
	assert.Equal(t, 409, errorResp.Status)
	assert.Equal(t, "SKU-001", errorResp.ConflictDetails.ProductSKU)
	assert.Equal(t, 35, errorResp.ConflictDetails.RequestedQuantity)
	assert.Equal(t, int64(30), errorResp.ConflictDetails.AvailableStock)
	assert.Equal(t, 5, errorResp.ConflictDetails.Shortfall)

	mockRepo.AssertExpectations(t)
}

