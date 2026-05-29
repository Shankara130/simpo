/**
 * Sync Queue Types
 * Defines interfaces for transaction synchronization queue
 * Story 8.2: Implement Transaction Sync Queue
 */

import { OfflineTransactionWithItems } from './offline.types';

/**
 * Sync Progress State
 * Real-time sync metrics for UI indicators
 */
export interface SyncState {
  pendingCount: number;          // Number of transactions waiting to sync
  processingCount: number;       // Number of transactions currently processing (0 or 1)
  syncedCount: number;           // Number successfully synced in current session
  failedCount: number;           // Number failed in current session
  currentTransaction: string | null;  // Transaction number currently syncing
}

/**
 * Sync Queue State (Persisted in AsyncStorage)
 * Queue processing metadata for crash recovery
 */
export interface SyncQueueState {
  isProcessing: boolean;
  currentTransactionId: number | null;
  lastUpdated: string;           // ISO 8601 timestamp
}

/**
 * Retry Schedule Entry (Persisted in AsyncStorage)
 * Individual transaction retry information
 */
export interface RetryScheduleEntry {
  retryAt: string;               // ISO 8601 timestamp when retry should occur
  retryCount: number;            // Current retry attempt count (0-4)
}

/**
 * Retry Schedule (Persisted in AsyncStorage)
 * Maps transaction ID to retry schedule entries
 */
export type RetrySchedule = Record<string, RetryScheduleEntry>;

/**
 * Sync Response from Backend
 * Success response format for POST /api/v1/sync
 */
export interface SyncResponse {
  status: 'synced';
  transaction_id: number;       // Server-side transaction ID
  server_timestamp: string;      // ISO 8601 timestamp from server
}

/**
 * Sync Error Response from Backend
 * RFC 7807 Problem Details format
 */
export interface SyncErrorResponse {
  type: string;                  // Error type URI
  title: string;                 // Error title
  detail: string;                // Detailed error message
}

/**
 * Sync Request Payload for Backend
 * Transaction data sent to sync endpoint
 */
export interface SyncRequest {
  transaction_number: string;
  timestamp: string;             // ISO 8601
  cashier_id: number;
  payment_method: string;
  total: string;                 // Decimal as string
  subtotal: string;
  tax: string;
  discount: string;
  customer_name?: string;
  items: Array<{
    product_id: number;
    product_sku: string;
    product_name: string;
    quantity: number;
    unit_price: string;
    subtotal: string;
  }>;
}

/**
 * Sync Result
 * Result of individual transaction sync attempt
 */
export interface SyncResult {
  transactionId: number;
  transactionNumber: string;
  success: boolean;
  error?: string;
  isRetryable: boolean;
  skipped?: boolean;             // True if duplicate (409)
}

/**
 * Queue Processing Result
 * Summary of queue processing session
 */
export interface QueueProcessingResult {
  totalProcessed: number;
  syncedCount: number;
  failedCount: number;
  skippedCount: number;
  duration: number;              // Processing duration in milliseconds
}

/**
 * Sync Queue Error Types
 * Categorizes sync errors for retry logic
 */
export enum SyncErrorType {
  NETWORK = 'network',           // Network error - retryable
  SERVER_ERROR = 'server_error', // 5xx error - retryable
  VALIDATION_ERROR = 'validation_error', // 400 error - not retryable
  CONFLICT = 'conflict',         // 409 error - skip (already synced)
  UNKNOWN = 'unknown',           // Unknown error - retryable
}

/**
 * Sync Queue Error
 * Custom error class for sync queue operations
 */
export class SyncQueueError extends Error {
  constructor(
    message: string,
    public readonly errorType: SyncErrorType,
    public originalError?: any,
    public readonly isRetryable: boolean = true
  ) {
    super(message);
    this.name = 'SyncQueueError';
  }
}

/**
 * AsyncStorage Keys for Sync Queue
 */
export const SYNC_QUEUE_STATE_KEY = '@simpo_sync_queue_state';
export const SYNC_RETRY_SCHEDULE_KEY = '@simpo_sync_retry_schedule';
export const SYNC_PROGRESS_KEY = '@simpo_sync_progress';

/**
 * Retry Configuration
 */
export const MAX_RETRY_ATTEMPTS = 5;
export const BASE_RETRY_DELAY_MS = 60 * 1000; // 1 minute
export const CRASH_RECOVERY_THRESHOLD_MS = 5 * 60 * 1000; // 5 minutes
export const NETWORK_DEBOUNCE_MS = 500; // 500ms
