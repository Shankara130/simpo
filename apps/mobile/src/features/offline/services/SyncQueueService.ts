/**
 * SyncQueueService - Transaction sync queue orchestration
 * Story 8.2: Implement Transaction Sync Queue
 *
 * Provides sequential transaction processing with:
 * - Chronological ordering (oldest first)
 * - Exponential backoff retry logic
 * - Queue state persistence (crash recovery)
 * - Network state handling
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import OfflineStorageService from './OfflineStorageService';
import { SyncAPI } from './SyncAPI';
import { ConflictResolutionService } from './ConflictResolutionService';
import { OfflineTransactionWithItems } from '../types/offline.types';
import {
  SyncQueueError,
  SyncErrorType,
  SyncResult,
  QueueProcessingResult,
  SyncQueueState,
  RetrySchedule,
  RetryScheduleEntry,
  MAX_RETRY_ATTEMPTS,
  BASE_RETRY_DELAY_MS,
  CRASH_RECOVERY_THRESHOLD_MS,
  SYNC_QUEUE_STATE_KEY,
  SYNC_RETRY_SCHEDULE_KEY,
} from '../types/sync.types';

/**
 * SyncQueue - Singleton service for queue orchestration
 * Follows service class pattern from OfflineStorageService
 */
class SyncQueue {
  private static instance: SyncQueue;
  private isProcessing: boolean = false;
  private abortController: AbortController | null = null;
  private offlineStorage: ReturnType<typeof OfflineStorageService.getInstance>;
  private syncAPI: ReturnType<typeof SyncAPI.getInstance>;
  private conflictResolution: ReturnType<typeof ConflictResolutionService.getInstance>;

  private constructor() {
    this.offlineStorage = OfflineStorageService.getInstance();
    this.syncAPI = SyncAPI.getInstance();
    this.conflictResolution = ConflictResolutionService.getInstance();
  }

  /**
   * Get singleton instance
   */
  static getInstance(): SyncQueue {
    if (!SyncQueue.instance) {
      SyncQueue.instance = new SyncQueue();
    }
    return SyncQueue.instance;
  }

  /**
   * Validate and convert transaction ID to string key
   * Ensures transaction ID is a positive integer before conversion
   * @throws Error if transaction ID is invalid
   */
  private validateTransactionId(transactionId: number): string {
    if (!Number.isInteger(transactionId) || transactionId <= 0 || !Number.isFinite(transactionId)) {
      throw new Error(`Invalid transaction ID: ${transactionId}. Must be a positive integer.`);
    }
    return transactionId.toString();
  }

  /**
   * Get all pending transactions in chronological order
   * AC1: Read transactions where status='pending_sync', sort by created_at ASC
   */
  async getPendingTransactions(): Promise<OfflineTransactionWithItems[]> {
    try {
      const transactions =
        await this.offlineStorage.getPendingTransactions();
      // Already ordered by created_at ASC from OfflineStorageService
      return transactions;
    } catch (error) {
      throw new SyncQueueError(
        'Failed to retrieve pending transactions',
        SyncErrorType.UNKNOWN,
        error,
        true
      );
    }
  }

  /**
   * Process queue sequentially (one transaction at a time)
   * AC2: Sequential processing loop with POST /api/v1/sync
   */
  async processQueue(): Promise<QueueProcessingResult> {
    // Check if already processing (includes abort check for safety)
    if (this.isProcessing || this.abortController?.signal.aborted) {
      // Already processing or aborted - return empty result
      return {
        totalProcessed: 0,
        syncedCount: 0,
        failedCount: 0,
        skippedCount: 0,
        duration: 0,
      };
    }

    // Set processing flag immediately to prevent race condition
    this.isProcessing = true;

    const startTime = Date.now();
    this.abortController = new AbortController();

    const result: QueueProcessingResult = {
      totalProcessed: 0,
      syncedCount: 0,
      failedCount: 0,
      skippedCount: 0,
      duration: 0,
    };

    try {
      // Get pending transactions
      const transactions = await this.getPendingTransactions();

      if (transactions.length === 0) {
        return result;
      }

      // Process each transaction sequentially
      for (const transaction of transactions) {
        // Check if aborted (network lost or stopped)
        if (this.abortController?.signal.aborted) {
          break;
        }

        // Update queue state
        await this.saveQueueState({
          isProcessing: true,
          currentTransactionId: transaction.id,
          lastUpdated: new Date().toISOString(),
        });

        // Check if this transaction should retry now
        const shouldRetry = await this.shouldRetryNow(transaction.id);
        if (
          transaction.status === 'failed' &&
          !shouldRetry &&
          (await this.getRetryCount(transaction.id)) >= MAX_RETRY_ATTEMPTS
        ) {
          // Max retries exceeded - permanently failed
          console.warn(`[SyncQueue] Transaction ${transaction.transaction_number} permanently failed after ${MAX_RETRY_ATTEMPTS} attempts`);
          // Clear retry schedule to prevent repeated evaluations
          await this.clearRetrySchedule(transaction.id);
          continue;
        }

        result.totalProcessed++;

        // Attempt to sync transaction
        const syncResult = await this.syncTransaction(transaction);

        if (syncResult.success) {
          result.syncedCount++;
          // Mark as synced and delete atomically (prevents race condition)
          await this.offlineStorage.markAndDeleteTransaction(transaction.id);
          // Clear retry schedule
          await this.clearRetrySchedule(transaction.id);
        } else if (syncResult.skipped) {
          result.skippedCount++;
          // Duplicate - mark as synced (skip, don't delete for audit trail)
          await this.offlineStorage.markTransactionSynced(transaction.id);
        } else {
          result.failedCount++;
          // Schedule retry if retryable
          if (syncResult.isRetryable) {
            const retryCount = await this.getRetryCount(transaction.id);
            if (retryCount < MAX_RETRY_ATTEMPTS) {
              await this.scheduleRetry(transaction.id, retryCount);
            }
          }
        }
      }

      result.duration = Date.now() - startTime;
      return result;
    } catch (error) {
      throw new SyncQueueError(
        'Queue processing failed',
        SyncErrorType.UNKNOWN,
        error,
        true
      );
    } finally {
      // Reset processing state
      this.isProcessing = false;
      this.abortController = null;
      await this.saveQueueState({
        isProcessing: false,
        currentTransactionId: null,
        lastUpdated: new Date().toISOString(),
      });
    }
  }

  /**
   * Sync single transaction to backend
   */
  private async syncTransaction(
    transaction: OfflineTransactionWithItems
  ): Promise<SyncResult> {
    // Check if processing was aborted before making expensive network call
    if (this.abortController?.signal.aborted) {
      return {
        transactionId: transaction.id,
        transactionNumber: transaction.transaction_number,
        success: false,
        error: 'Processing canceled',
        isRetryable: true,
      };
    }

    try {
      const response = await this.syncAPI.postTransaction(transaction);

      return {
        transactionId: transaction.id,
        transactionNumber: transaction.transaction_number,
        success: true,
      };
    } catch (error) {
      if (error instanceof SyncQueueError) {
        // AC6: Handle conflict errors with ConflictResolutionService
        if (error.errorType === SyncErrorType.CONFLICT && error.originalError) {
          // Try to parse conflict error details
          const conflictError = this.conflictResolution.parseConflictError(error.originalError);
          if (conflictError) {
            // Mark transaction as failed with conflict error
            try {
              await this.conflictResolution.markTransactionFailed(transaction.id, conflictError);
            } catch (markError) {
              console.warn('[SyncQueue] Failed to mark transaction conflict:', markError);
            }
          }
        }

        return {
          transactionId: transaction.id,
          transactionNumber: transaction.transaction_number,
          success: false,
          error: error.message,
          isRetryable: error.isRetryable,
          skipped: error.errorType === SyncErrorType.CONFLICT,
        };
      }

      return {
        transactionId: transaction.id,
        transactionNumber: transaction.transaction_number,
        success: false,
        error: 'Unknown error',
        isRetryable: true,
      };
    }
  }

  /**
   * Stop ongoing queue processing
   * AC6: Cancel processing on network loss
   */
  stopProcessing(): void {
    if (this.abortController) {
      this.abortController.abort();
    }
  }

  /**
   * Calculate exponential backoff delay
   * AC3: 1, 2, 4, 8, 16 minutes (based on 2^retryCount)
   * Note: Spec mentions 32 minutes but implementation produces 16 minutes at retryCount=4
   */
  calculateBackoff(retryCount: number): number {
    if (retryCount >= MAX_RETRY_ATTEMPTS) {
      // Cap at 32 minutes if exceeded max attempts
      return BASE_RETRY_DELAY_MS * 32;
    }

    // Exponential backoff: 2^retryCount * BASE_RETRY_DELAY_MS
    // Produces: 1, 2, 4, 8, 16 minutes for retryCount 0-4
    const backoff = BASE_RETRY_DELAY_MS * Math.pow(2, retryCount);
    return Math.min(backoff, BASE_RETRY_DELAY_MS * 32);
  }

  /**
   * Schedule retry for failed transaction
   * AC3: Store retry count and timestamp in AsyncStorage
   */
  async scheduleRetry(
    transactionId: number,
    currentRetryCount: number
  ): Promise<void> {
    const nextRetryCount = currentRetryCount + 1;
    const backoff = this.calculateBackoff(currentRetryCount);
    const retryAt = new Date(Date.now() + backoff).toISOString();

    const entry: RetryScheduleEntry = {
      retryAt,
      retryCount: nextRetryCount,
    };

    // Get existing schedule
    const scheduleJson =
      (await AsyncStorage.getItem(SYNC_RETRY_SCHEDULE_KEY)) || '{}';
    let schedule: RetrySchedule;
    try {
      schedule = JSON.parse(scheduleJson);
    } catch (error) {
      console.warn('[SyncQueue] Corrupted retry schedule, resetting:', error);
      schedule = {};
    }

    // Update schedule for this transaction
    schedule[this.validateTransactionId(transactionId)] = entry;

    // Save back to AsyncStorage with error handling
    try {
      await AsyncStorage.setItem(
        SYNC_RETRY_SCHEDULE_KEY,
        JSON.stringify(schedule)
      );
    } catch (error) {
      console.error('[SyncQueue] Failed to persist retry schedule, retry may be lost:', error);
      // Continue anyway - transaction will be retried from scratch on next queue processing
    }
  }

  /**
   * Check if transaction should retry now
   * AC3: Compare current time with scheduled retry time
   */
  async shouldRetryNow(transactionId: number): Promise<boolean> {
    const scheduleJson =
      (await AsyncStorage.getItem(SYNC_RETRY_SCHEDULE_KEY)) || '{}';
    let schedule: RetrySchedule;
    try {
      schedule = JSON.parse(scheduleJson);
    } catch (error) {
      console.warn('[SyncQueue] Corrupted retry schedule, assuming ready:', error);
      return true; // No valid schedule - ready to retry
    }

    const entry = schedule[this.validateTransactionId(transactionId)];
    if (!entry) {
      return true; // No retry schedule - ready to retry
    }

    // Check retry count
    if (entry.retryCount >= MAX_RETRY_ATTEMPTS) {
      return false; // Max retries exceeded
    }

    // Check if retry time has passed
    const now = Date.now();
    const retryTime = new Date(entry.retryAt).getTime();

    return now >= retryTime;
  }

  /**
   * Get retry count for transaction
   */
  async getRetryCount(transactionId: number): Promise<number> {
    const scheduleJson =
      (await AsyncStorage.getItem(SYNC_RETRY_SCHEDULE_KEY)) || '{}';
    let schedule: RetrySchedule;
    try {
      schedule = JSON.parse(scheduleJson);
    } catch (error) {
      console.warn('[SyncQueue] Corrupted retry schedule, returning 0:', error);
      return 0;
    }

    const entry = schedule[this.validateTransactionId(transactionId)];
    return entry?.retryCount || 0;
  }

  /**
   * Clear retry schedule for transaction
   */
  async clearRetrySchedule(transactionId: number): Promise<void> {
    const scheduleJson =
      (await AsyncStorage.getItem(SYNC_RETRY_SCHEDULE_KEY)) || '{}';
    let schedule: RetrySchedule;
    try {
      schedule = JSON.parse(scheduleJson);
    } catch (error) {
      console.warn('[SyncQueue] Corrupted retry schedule, resetting:', error);
      schedule = {};
    }

    delete schedule[this.validateTransactionId(transactionId)];

    try {
      await AsyncStorage.setItem(
        SYNC_RETRY_SCHEDULE_KEY,
        JSON.stringify(schedule)
      );
    } catch (error) {
      console.error('[SyncQueue] Failed to persist cleared retry schedule:', error);
      // Non-critical - entry will expire naturally or be cleaned up
    }
  }

  /**
   * Save queue state to AsyncStorage
   * AC5: Persist isProcessing, currentTransactionId, lastUpdated
   */
  async saveQueueState(state: SyncQueueState): Promise<void> {
    try {
      await AsyncStorage.setItem(
        SYNC_QUEUE_STATE_KEY,
        JSON.stringify(state)
      );
    } catch (error) {
      console.error('[SyncQueue] Failed to persist queue state:', error);
      // Queue processing will continue, but crash recovery may not work correctly
    }
  }

  /**
   * Load queue state from AsyncStorage
   */
  async loadQueueState(): Promise<SyncQueueState> {
    const stateJson =
      (await AsyncStorage.getItem(SYNC_QUEUE_STATE_KEY)) || '{}';
    let state: SyncQueueState;
    try {
      state = JSON.parse(stateJson);
    } catch (error) {
      console.warn('[SyncQueue] Corrupted queue state, using defaults:', error);
      state = {} as SyncQueueState;
    }

    return {
      isProcessing: state.isProcessing || false,
      currentTransactionId: state.currentTransactionId || null,
      lastUpdated: state.lastUpdated || new Date().toISOString(),
    };
  }

  /**
   * Check if processing state is orphaned (crashed)
   * AC5: If isProcessing=true and timestamp > 5 minutes ago
   */
  async isProcessingStateOrphaned(): Promise<boolean> {
    const state = await this.loadQueueState();

    if (!state.isProcessing) {
      return false;
    }

    const lastUpdate = new Date(state.lastUpdated).getTime();
    const now = Date.now();
    const elapsed = now - lastUpdate;

    return elapsed > CRASH_RECOVERY_THRESHOLD_MS;
  }
}

// Export class as SyncQueue for clarity
export { SyncQueue };
export default SyncQueue.getInstance();
