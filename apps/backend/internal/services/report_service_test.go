package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// MockReportRepository is a mock implementation for testing
// Story 5.1, Task 3: Testing ReportService business logic
// Story 5.2, Task 3: Testing profit/loss report generation
type MockReportRepository struct {
	GetDailySalesSummaryFunc    func(ctx context.Context, date string, branchID uint) (*dto.DailySalesSummaryDTO, error)
	GetProfitLossSummaryFunc   func(ctx context.Context, startDate, endDate string, branchID uint, breakdownBy string) (*dto.ProfitLossSummaryDTO, error)
}

func (m *MockReportRepository) GetDailySalesSummary(ctx context.Context, date string, branchID uint) (*dto.DailySalesSummaryDTO, error) {
	if m.GetDailySalesSummaryFunc != nil {
		return m.GetDailySalesSummaryFunc(ctx, date, branchID)
	}
	return &dto.DailySalesSummaryDTO{}, nil
}

func (m *MockReportRepository) GetProfitLossSummary(ctx context.Context, startDate, endDate string, branchID uint, breakdownBy string) (*dto.ProfitLossSummaryDTO, error) {
	if m.GetProfitLossSummaryFunc != nil {
		return m.GetProfitLossSummaryFunc(ctx, startDate, endDate, branchID, breakdownBy)
	}
	return &dto.ProfitLossSummaryDTO{}, nil
}

// TestReportService_GenerateDailySalesSummary_Success tests successful report generation
// Story 5.1, Task 3.1-3.6: Full workflow test
func TestReportService_GenerateDailySalesSummary_Success(t *testing.T) {
	// Arrange: Create mock repository with test data
	mockRepo := &MockReportRepository{
		GetDailySalesSummaryFunc: func(ctx context.Context, date string, branchID uint) (*dto.DailySalesSummaryDTO, error) {
			return &dto.DailySalesSummaryDTO{
				Date:              date,
				BranchID:          branchID,
				BranchName:        "Test Branch",
				TotalSales:        "1000000.00",
				TotalTransactions: 10,
				PaymentBreakdown: []dto.PaymentBreakdown{
					{PaymentMethod: "CASH", Amount: "500000.00", TransactionCount: 5, Percentage: 50.0},
				},
				GeneratedAt: time.Now(),
			}, nil
		},
	}

	// Create service without Redis (optional)
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewReportService(mockTxnRepo, mockProdRepo, mockRepo, mockAudit, nil)

	// Act: Generate report
	req := &dto.DailySalesRequest{
		Date:     "2026-05-23",
		BranchID: uintPtrForTest(1),
	}

	result, err := service.GenerateDailySalesSummary(context.Background(), req)

	// Assert: Verify success
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "2026-05-23", result.Date)
	assert.Equal(t, uint(1), result.BranchID)
	assert.Equal(t, "Test Branch", result.BranchName)
	assert.Equal(t, "1000000.00", result.TotalSales)
	assert.Equal(t, 10, result.TotalTransactions)
}

// TestReportService_GenerateDailySalesSummary_AllBranches tests report for all branches
// Story 5.1, AC2: Branch filtering with nil branch_id
func TestReportService_GenerateDailySalesSummary_AllBranches(t *testing.T) {
	// Arrange
	mockRepo := &MockReportRepository{
		GetDailySalesSummaryFunc: func(ctx context.Context, date string, branchID uint) (*dto.DailySalesSummaryDTO, error) {
			return &dto.DailySalesSummaryDTO{
				Date:       date,
				BranchID:   0, // 0 means all branches
				BranchName: "All Branches",
				GeneratedAt: time.Now(),
			}, nil
		},
	}

	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewReportService(mockTxnRepo, mockProdRepo, mockRepo, mockAudit, nil)
	req := &dto.DailySalesRequest{
		Date:     "2026-05-23",
		BranchID: nil, // nil means all branches
	}

	// Act
	result, err := service.GenerateDailySalesSummary(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(0), result.BranchID)
	assert.Equal(t, "All Branches", result.BranchName)
}

// TestReportService_GenerateDailySalesSummary_InvalidDate tests date validation
// Story 5.1, Task 3.2: Date format validation
func TestReportService_GenerateDailySalesSummary_InvalidDate(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockReportRepo := new(MockReportRepository)
	mockAudit := new(MockAuditService)
	service := NewReportService(mockTxnRepo, mockProdRepo, mockReportRepo, mockAudit, nil)
	req := &dto.DailySalesRequest{
		Date:     "invalid-date",
		BranchID: nil,
	}

	// Act
	result, err := service.GenerateDailySalesSummary(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &InvalidInputError{}, err)
}

// TestReportService_GenerateDailySalesSummary_EmptyDate tests empty date validation
// Story 5.1, Task 3.2: Required field validation
func TestReportService_GenerateDailySalesSummary_EmptyDate(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockReportRepo := new(MockReportRepository)
	mockAudit := new(MockAuditService)
	service := NewReportService(mockTxnRepo, mockProdRepo, mockReportRepo, mockAudit, nil)
	req := &dto.DailySalesRequest{
		Date:     "",
		BranchID: nil,
	}

	// Act
	result, err := service.GenerateDailySalesSummary(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &InvalidInputError{}, err)
}

// uintPtr is a helper function to create a pointer to uint
// Note: This function is already defined in product_service_impl_test.go
// We're creating it here for test isolation
func uintPtrForTest(i uint) *uint {
	return &i
}
