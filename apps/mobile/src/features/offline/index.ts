/**
 * Offline Feature Module
 * Story 8.1: Implement Local SQLite Storage for Offline Transactions
 *
 * Exports all offline functionality for use across the mobile app
 */

// Services
export { default as OfflineStorageService } from './services/OfflineStorageService';
export { CacheService } from './services/CacheService';

// Hooks
export { useNetworkStatus } from './hooks/useNetworkStatus';

// Types
export type {
  OfflineTransaction,
  OfflineTransactionItem,
  OfflineTransactionWithItems,
  OfflineTransactionStatus,
  CachedStockData,
} from './types/offline.types';
export {
  OfflineStorageError,
  DATABASE_NAME,
  TABLE_TRANSACTIONS,
  TABLE_TRANSACTION_ITEMS,
  CACHE_LAST_STOCK_SYNC,
  CACHE_STOCK_DATA,
} from './types/offline.types';
