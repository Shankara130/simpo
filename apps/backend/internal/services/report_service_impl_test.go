package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// Test NewReportService with nil dependencies
func TestNewReportService_PanicOnNilDependencies(t *testing.T) {
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)

	assert.Panics(t, func() {
		NewReportService(nil, mockProdRepo, mockAudit)
	}, "Should panic when transactionRepo is nil")

	assert.Panics(t, func() {
		NewReportService(mockTxnRepo, nil, mockAudit)
	}, "Should panic when productRepo is nil")

	assert.Panics(t, func() {
		NewReportService(mockTxnRepo, mockProdRepo, nil)
	}, "Should panic when auditService is nil")
}

// Test GenerateDailySales
func TestReportService_GenerateDailySales_Success(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewReportService(mockTxnRepo, mockProdRepo, mockAudit)

	startDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)

	summary := &repositories.TransactionSummary{
		TotalTransactions: 10,
		TotalAmount:       "150000.00",
		SubtotalAmount:    "150000.00",
		TaxAmount:         "0.00",
		DiscountAmount:    "0.00",
		PaymentMethods: []repositories.PaymentMethodSummary{
			{PaymentMethod: "CASH", Count: 5, TotalAmount: "75000.00"},
			{PaymentMethod: "E-WALLET", Count: 5, TotalAmount: "75000.00"},
		},
	}

	// Mock expectations
	// PATCH: Updated to expect startDate (date range fix)
	mockTxnRepo.On("GetDailySummary", mock.Anything, uint(1), startDate).Return(summary, nil)

	// Act
	report, err := service.GenerateDailySales(context.Background(), 1, startDate, endDate)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, uint(1), report.BranchID)
	assert.Equal(t, "150000.00", report.TotalSales)
	assert.Equal(t, int64(10), report.TotalTransactions)
	assert.Len(t, report.PaymentMethods, 2)
	mockTxnRepo.AssertExpectations(t)
}

func TestReportService_GenerateDailySales_InvalidDateRange(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewReportService(mockTxnRepo, mockProdRepo, mockAudit)

	startDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC) // End before start

	// Act
	_, err := service.GenerateDailySales(context.Background(), 1, startDate, endDate)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "date_range", invErr.Field)
}

func TestReportService_GenerateDailySales_ZeroBranchID(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewReportService(mockTxnRepo, mockProdRepo, mockAudit)

	startDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)

	// Act
	_, err := service.GenerateDailySales(context.Background(), 0, startDate, endDate)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "branch_id", invErr.Field)
}

func TestReportService_GenerateDailySales_ContextCanceled(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewReportService(mockTxnRepo, mockProdRepo, mockAudit)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	startDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)

	// Act
	_, err := service.GenerateDailySales(ctx, 1, startDate, endDate)

	// Assert
	assert.Error(t, err)
}

// Test GenerateProfitLoss
func TestReportService_GenerateProfitLoss_Success(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewReportService(mockTxnRepo, mockProdRepo, mockAudit)

	startDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)

	summary := &repositories.TransactionSummary{
		TotalTransactions: 10,
		TotalAmount:       "150000.00",
		SubtotalAmount:    "150000.00",
	}

	// Mock expectations
	// PATCH: Updated to expect startDate (date range fix)
	mockTxnRepo.On("GetDailySummary", mock.Anything, uint(1), startDate).Return(summary, nil)

	// Act
	report, err := service.GenerateProfitLoss(context.Background(), 1, startDate, endDate)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, uint(1), report.BranchID)
	assert.Equal(t, "150000.00", report.Revenue)
	assert.Equal(t, "150000.00", report.NetProfit)
	mockTxnRepo.AssertExpectations(t)
}

func TestReportService_GenerateProfitLoss_InvalidDateRange(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewReportService(mockTxnRepo, mockProdRepo, mockAudit)

	startDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC) // End before start

	// Act
	_, err := service.GenerateProfitLoss(context.Background(), 1, startDate, endDate)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "date_range", invErr.Field)
}

// Test ExportReport
func TestReportService_ExportReport_NotImplemented(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewReportService(mockTxnRepo, mockProdRepo, mockAudit)

	// Act
	_, err := service.ExportReport(context.Background(), "daily_sales", "pdf")

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "export", invErr.Field)
}
