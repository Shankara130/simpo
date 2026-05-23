package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
