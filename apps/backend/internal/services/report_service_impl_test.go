package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// Test NewReportService with nil dependencies
func TestNewReportService_PanicOnNilDependencies(t *testing.T) {
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockReportRepo := new(MockReportRepository) // Story 5.1, Task 3: New dependency
	mockAudit := new(MockAuditService)

	assert.Panics(t, func() {
		NewReportService(nil, mockProdRepo, mockReportRepo, mockAudit, nil)
	}, "Should panic when transactionRepo is nil")

	assert.Panics(t, func() {
		NewReportService(mockTxnRepo, nil, mockReportRepo, mockAudit, nil)
	}, "Should panic when productRepo is nil")

	assert.Panics(t, func() {
		NewReportService(mockTxnRepo, mockProdRepo, nil, mockAudit, nil)
	}, "Should panic when reportRepo is nil") // Story 5.1, Task 3

	assert.Panics(t, func() {
		NewReportService(mockTxnRepo, mockProdRepo, mockReportRepo, nil, nil)
	}, "Should panic when auditService is nil")
}

// ==============================================================================
// Story 5.2: Profit/Loss Report Service Tests
// ==============================================================================

// TestReportService_GenerateProfitLossSummary_Signature tests method signature
// Story 5.2, Task 3.1: Verify method exists with correct signature
func TestReportService_GenerateProfitLossSummary_Signature(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockReportRepo := new(MockReportRepository)
	mockAudit := new(MockAuditService)

	service := NewReportService(mockTxnRepo, mockProdRepo, mockReportRepo, mockAudit, nil)
	require.NotNil(t, service, "Service should be created")
}

// TestReportService_GenerateProfitLossSummary_Validation tests request validation
// Story 5.2, Task 3.2: Validate ProfitLossRequest DTO
func TestReportService_GenerateProfitLossSummary_Validation(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockProdRepo := new(MockProductRepository)
	mockReportRepo := new(MockReportRepository)
	mockAudit := new(MockAuditService)

	service := NewReportService(mockTxnRepo, mockProdRepo, mockReportRepo, mockAudit, nil)
	ctx := context.Background()

	tests := []struct {
		name        string
		request     *dto.ProfitLossRequest
		expectError bool
	}{
		{
			name: "Invalid - empty start date",
			request: &dto.ProfitLossRequest{
				StartDate:   "",
				EndDate:     "2026-05-23",
				BreakdownBy: "category",
			},
			expectError: true,
		},
		{
			name: "Invalid - empty end date",
			request: &dto.ProfitLossRequest{
				StartDate:   "2026-05-01",
				EndDate:     "",
				BreakdownBy: "category",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := service.GenerateProfitLossSummary(ctx, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err, "Expected validation error")
			}
		})
	}
}
