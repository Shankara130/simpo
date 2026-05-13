package services

import (
	"context"
)

// SyncService defines the interface for synchronization business operations
// AC1: Service interface for sync domain with clear business method signatures
type SyncService interface {
	// QueueTransactionSync queues a transaction for offline sync
	// Stub for future offline sync story
	QueueTransactionSync(ctx context.Context, transactionID uint) error

	// ProcessSyncQueue processes pending sync operations
	// Stub for future offline sync story
	ProcessSyncQueue(ctx context.Context) error

	// ResolveConflict resolves conflicts in offline transactions
	// Stub for future offline sync story
	ResolveConflict(ctx context.Context, conflictID uint, resolution interface{}) error
}
