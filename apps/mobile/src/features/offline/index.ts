/**
 * Offline Feature Module
 * Story 8.1: Implement Local SQLite Storage for Offline Transactions
 * Story 8.2: Implement Transaction Sync Queue
 *
 * Exports all offline functionality for use across the mobile app
 */

// Services
export { default as OfflineStorageService } from './services/OfflineStorageService';
export { CacheService } from './services/CacheService';
export { SyncQueue } from './services/SyncQueueService';
export { SyncAPI } from './services/SyncAPI';

// Hooks
export { useNetworkStatus } from './hooks/useNetworkStatus';
export { useSyncProgress } from './hooks/useSyncProgress';
export { useSyncQueue } from './hooks/useSyncQueue';

// Types
export type {
  OfflineTransaction,
  OfflineTransactionItem,
  OfflineTransactionWithItems,
  OfflineTransactionStatus,
  CachedStockData,
} from './types/offline.types';
export type {
  SyncState,
  SyncQueueState,
  RetrySchedule,
  RetryScheduleEntry,
  SyncResult,
  QueueProcessingResult,
  SyncResponse,
  SyncErrorResponse,
  SyncRequest,
} from './types/sync.types';

export {
  OfflineStorageError,
  SyncQueueError,
  DATABASE_NAME,
  TABLE_TRANSACTIONS,
  TABLE_TRANSACTION_ITEMS,
  CACHE_LAST_STOCK_SYNC,
  CACHE_STOCK_DATA,
  SYNC_QUEUE_STATE_KEY,
  SYNC_RETRY_SCHEDULE_KEY,
  SYNC_PROGRESS_KEY,
  MAX_RETRY_ATTEMPTS,
  BASE_RETRY_DELAY_MS,
  CRASH_RECOVERY_THRESHOLD_MS,
  NETWORK_DEBOUNCE_MS,
} from './types/offline.types';
export {
  SyncErrorType,
} from './types/sync.types';
