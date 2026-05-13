package services

import (
	"context"
	"fmt"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// syncService implements SyncService interface
// AC2: Services use repository interfaces (not concrete implementations)
type syncService struct {
	transactionRepo repositories.TransactionRepository
	auditService    AuditService
}

// NewSyncService creates a new sync service with dependency injection
// AC2: Services accept repository interfaces via constructor injection
func NewSyncService(
	transactionRepo repositories.TransactionRepository,
	auditService AuditService,
) SyncService {
	// Fail fast on nil dependencies
	if transactionRepo == nil {
		panic("syncService: transactionRepo cannot be nil")
	}
	if auditService == nil {
		panic("syncService: auditService cannot be nil")
	}

	return &syncService{
		transactionRepo: transactionRepo,
		auditService:    auditService,
	}
}

// QueueTransactionSync queues a transaction for offline sync
// Stub for future offline sync story
func (s *syncService) QueueTransactionSync(ctx context.Context, transactionID uint) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate transaction ID
	if transactionID == 0 {
		return &InvalidInputError{Field: "transaction_id", Message: "transaction ID is required"}
	}

	// Stub implementation for future story
	return &InvalidInputError{
		Field:   "sync_queue",
		Message: "transaction sync queue not implemented - scheduled for future offline sync story",
	}
}

// ProcessSyncQueue processes pending sync operations
// Stub for future offline sync story
func (s *syncService) ProcessSyncQueue(ctx context.Context) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Stub implementation for future story
	return &InvalidInputError{
		Field:   "sync_process",
		Message: "sync queue processing not implemented - scheduled for future offline sync story",
	}
}

// ResolveConflict resolves conflicts in offline transactions
// Stub for future offline sync story
func (s *syncService) ResolveConflict(ctx context.Context, conflictID uint, resolution interface{}) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate conflict ID
	if conflictID == 0 {
		return &InvalidInputError{Field: "conflict_id", Message: "conflict ID is required"}
	}

	// Stub implementation for future story
	return &InvalidInputError{
		Field:   "conflict_resolution",
		Message: "conflict resolution not implemented - scheduled for future offline sync story",
	}
}
