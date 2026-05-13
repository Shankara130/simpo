package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test NewSyncService with nil dependencies
func TestNewSyncService_PanicOnNilDependencies(t *testing.T) {
	mockTxnRepo := new(MockTransactionRepository)
	mockAudit := new(MockAuditService)

	assert.Panics(t, func() {
		NewSyncService(nil, mockAudit)
	}, "Should panic when transactionRepo is nil")

	assert.Panics(t, func() {
		NewSyncService(mockTxnRepo, nil)
	}, "Should panic when auditService is nil")
}

// Test QueueTransactionSync
func TestSyncService_QueueTransactionSync_NotImplemented(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockAudit := new(MockAuditService)
	service := NewSyncService(mockTxnRepo, mockAudit)

	// Act
	err := service.QueueTransactionSync(context.Background(), 1)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "sync_queue", invErr.Field)
}

func TestSyncService_QueueTransactionSync_ZeroTransactionID(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockAudit := new(MockAuditService)
	service := NewSyncService(mockTxnRepo, mockAudit)

	// Act
	err := service.QueueTransactionSync(context.Background(), 0)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "transaction_id", invErr.Field)
}

func TestSyncService_QueueTransactionSync_ContextCanceled(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockAudit := new(MockAuditService)
	service := NewSyncService(mockTxnRepo, mockAudit)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	err := service.QueueTransactionSync(ctx, 1)

	// Assert
	assert.Error(t, err)
}

// Test ProcessSyncQueue
func TestSyncService_ProcessSyncQueue_NotImplemented(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockAudit := new(MockAuditService)
	service := NewSyncService(mockTxnRepo, mockAudit)

	// Act
	err := service.ProcessSyncQueue(context.Background())

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "sync_process", invErr.Field)
}

func TestSyncService_ProcessSyncQueue_ContextCanceled(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockAudit := new(MockAuditService)
	service := NewSyncService(mockTxnRepo, mockAudit)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	err := service.ProcessSyncQueue(ctx)

	// Assert
	assert.Error(t, err)
}

// Test ResolveConflict
func TestSyncService_ResolveConflict_NotImplemented(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockAudit := new(MockAuditService)
	service := NewSyncService(mockTxnRepo, mockAudit)

	// Act
	err := service.ResolveConflict(context.Background(), 1, nil)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "conflict_resolution", invErr.Field)
}

func TestSyncService_ResolveConflict_ZeroConflictID(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockAudit := new(MockAuditService)
	service := NewSyncService(mockTxnRepo, mockAudit)

	// Act
	err := service.ResolveConflict(context.Background(), 0, nil)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "conflict_id", invErr.Field)
}

func TestSyncService_ResolveConflict_ContextCanceled(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockAudit := new(MockAuditService)
	service := NewSyncService(mockTxnRepo, mockAudit)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	err := service.ResolveConflict(ctx, 1, nil)

	// Assert
	assert.Error(t, err)
}
