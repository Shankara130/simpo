/**
 * SyncOrchestrator - Bidirectional sync coordination
 * Story 8.3: Implement Bidirectional Data Synchronization
 *
 * Orchestrates the complete bidirectional sync flow:
 * Phase 1: Upload pending transactions (via SyncQueueService)
 * Phase 2: Download stock levels (via ProductSyncService)
 * Phase 3: Download new products (via ProductSyncService)
 * Phase 4: Download user data (via UserSyncService)
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import { SyncQueue } from './SyncQueueService';
import { ProductSyncService } from './ProductSyncService';
import { UserSyncService } from './UserSyncService';
import { ConflictResolutionService } from './ConflictResolutionService';
import {
  BidirectionalSyncState,
  BidirectionalSyncResult,
  SyncPhase,
  ProductSyncResponse,
  UserSyncResponse,
  SyncOrchestratorError,
  LAST_PRODUCT_SYNC_KEY,
  USER_PROFILE_KEY,
  BIDIRECTIONAL_SYNC_STATE_KEY,
  BIDIRECTIONAL_SYNC_RETRY_KEY,
  MAX_RETRY_ATTEMPTS,
  BASE_RETRY_DELAY_MS,
} from '../types/bidirectional-sync.types';

/**
 * SyncOrchestrator - Singleton service for bidirectional sync coordination
 * Follows service class pattern from SyncQueue and OfflineStorageService
 */
class SyncOrchestratorClass {
  private static instance: SyncOrchestratorClass;
  private isSyncing: boolean = false;
  private abortController: AbortController | null = null;
  private syncQueue: ReturnType<typeof SyncQueue.getInstance>;
  private conflictResolution: ReturnType<typeof ConflictResolutionService.getInstance>;

  private constructor() {
    this.syncQueue = SyncQueue.getInstance();
    this.conflictResolution = ConflictResolutionService.getInstance();
  }

  /**
   * Get singleton instance
   */
  static getInstance(): SyncOrchestratorClass {
    if (!SyncOrchestratorClass.instance) {
      SyncOrchestratorClass.instance = new SyncOrchestratorClass();
    }
    return SyncOrchestratorClass.instance;
  }


  /**
   * Execute full bidirectional sync
   * AC9: Coordinate entire sync flow with phase-based orchestration
   */
  async sync(): Promise<BidirectionalSyncResult> {
    // Prevent concurrent sync operations
    if (this.isSyncing) {
      return {
        success: false,
        phase: 'idle',
        uploaded: 0,
        downloadedStock: 0,
        downloadedProducts: 0,
        userUpdated: false,
        duration: 0,
        error: 'Sync already in progress',
      };
    }

    this.isSyncing = true;
    this.abortController = new AbortController();

    const startTime = Date.now();
    const result: BidirectionalSyncResult = {
      success: false,
      phase: 'uploading',
      uploaded: 0,
      downloadedStock: 0,
      downloadedProducts: 0,
      userUpdated: false,
      duration: 0,
    };

    try {
      // Phase 1: Upload pending transactions
      await this.updatePhase('uploading');
      const queueResult = await this.syncQueue.processQueue();

      // AC6: Handle conflict errors - don't halt sync for individual conflicts
      // Only halt if ALL uploads failed (not due to conflicts)
      const hasNonConflictFailures = queueResult.failedCount > 0;
      const hasSuccessfulUploads = queueResult.syncedCount > 0;

      if (hasNonConflictFailures && !hasSuccessfulUploads) {
        // All uploads failed with no conflicts - halt sync
        throw new SyncOrchestratorError(
          'All transaction uploads failed',
          'uploading'
        );
      }

      // AC6: Continue with download phases even if some transactions failed
      // Failed transactions are marked by ConflictResolutionService
      result.uploaded = queueResult.syncedCount;
      result.failedCount = queueResult.failedCount;

      // Check if aborted (network lost during upload)
      if (this.abortController.signal.aborted) {
        throw new SyncOrchestratorError('Sync canceled during upload phase', 'uploading');
      }

      // Phase 2: Download stock levels
      await this.updatePhase('downloading_stock');

      const productSync = ProductSyncService.getInstance();
      const stockResult = await productSync.syncStockLevels();
      result.downloadedStock = stockResult.products.length;

      // Check if aborted
      if (this.abortController.signal.aborted) {
        throw new SyncOrchestratorError('Sync canceled during stock download', 'downloading_stock');
      }

      // Phase 3: Download new products
      await this.updatePhase('downloading_products');

      const productsResult = await productSync.syncProducts();
      result.downloadedProducts = productsResult.products.length;

      // Check if aborted
      if (this.abortController.signal.aborted) {
        throw new SyncOrchestratorError('Sync canceled during product download', 'downloading_products');
      }

      // Phase 4: Download user data
      await this.updatePhase('downloading_user');

      const userSync = UserSyncService.getInstance();
      const userResult = await userSync.syncUser();
      // Only mark as updated if role changed or user was deactivated
      result.userUpdated = userSync.hasRoleChanged() || userSync.isUserDeactivated();

      // Check if aborted
      if (this.abortController.signal.aborted) {
        throw new SyncOrchestratorError('Sync canceled during user download', 'downloading_user');
      }

      // All phases complete
      result.success = true;
      result.phase = 'synced';
      result.duration = Date.now() - startTime;

      await this.updatePhase('synced');
      await this.saveLastSyncTime();

      return result;
    } catch (error) {
      // Handle error
      const phase = this.abortController.signal.aborted ? 'idle' : result.phase;
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';

      result.success = false;
      result.phase = phase;
      result.error = errorMessage;
      result.duration = Date.now() - startTime;

      await this.updatePhase('failed', errorMessage);

      // Schedule retry if error is retryable (but not for aborted syncs)
      if (!this.abortController?.signal.aborted) {
        await this.scheduleRetry();
      }

      return result;
    } finally {
      // Reset syncing state
      this.isSyncing = false;
      this.abortController = null;
    }
  }

  /**
   * Update sync phase and persist state
   * AC7: Visual sync status indicators
   */
  async updatePhase(
    phase: SyncPhase,
    errorMessage?: string
  ): Promise<void> {
    const currentState = await this.getSyncState();

    const newState: BidirectionalSyncState = {
      ...currentState,
      status: phase === 'synced' ? 'synced' : phase === 'failed' ? 'failed' : 'syncing',
      phase,
      currentPhase: this.getPhaseDescription(phase),
      error: errorMessage,
    };

    try {
      await AsyncStorage.setItem(
        BIDIRECTIONAL_SYNC_STATE_KEY,
        JSON.stringify(newState)
      );
    } catch (error) {
      console.error('[SyncOrchestrator] Failed to persist sync state:', error);
    }
  }

  /**
   * Get current sync state from AsyncStorage
   */
  async getSyncState(): Promise<BidirectionalSyncState> {
    try {
      const stateJson = await AsyncStorage.getItem(BIDIRECTIONAL_SYNC_STATE_KEY);

      if (!stateJson) {
        // Return default state
        return {
          status: 'idle',
          phase: 'idle',
          pendingCount: 0,
          processingCount: 0,
          syncedCount: 0,
          failedCount: 0,
          currentPhase: null,
          lastSyncTime: null,
        };
      }

      return (JSON.parse(stateJson) as BidirectionalSyncState) || {
        status: 'idle',
        phase: 'idle',
        pendingCount: 0,
        processingCount: 0,
        syncedCount: 0,
        failedCount: 0,
        currentPhase: null,
        lastSyncTime: null,
      };
    } catch (error) {
      console.error('[SyncOrchestrator] Failed to load sync state:', error);
      return {
        status: 'idle',
        phase: 'idle',
        pendingCount: 0,
        processingCount: 0,
        syncedCount: 0,
        failedCount: 0,
        currentPhase: null,
        lastSyncTime: null,
      };
    }
  }

  /**
   * Stop ongoing sync process
   * AC8: Automatic background retry for failed syncs (cancel on user stop)
   */
  stopSync(): void {
    if (this.abortController) {
      this.abortController.abort();
      this.syncQueue.stopProcessing();
    }
  }

  /**
   * Calculate exponential backoff delay
   * AC8: Retry intervals follow Story 8.2 pattern
   */
  calculateBackoff(retryCount: number): number {
    if (retryCount >= MAX_RETRY_ATTEMPTS) {
      // Cap at 32 minutes if exceeded max attempts
      return BASE_RETRY_DELAY_MS * 32;
    }

    // Exponential backoff: 2^retryCount * BASE_RETRY_DELAY_MS
    const backoff = BASE_RETRY_DELAY_MS * Math.pow(2, retryCount);
    return Math.min(backoff, BASE_RETRY_DELAY_MS * 32);
  }

  /**
   * Schedule retry for failed sync
   * AC8: Store failed sync attempts in AsyncStorage
   */
  async scheduleRetry(): Promise<void> {
    try {
      const retryDataJson = await AsyncStorage.getItem(BIDIRECTIONAL_SYNC_RETRY_KEY);
      let retryData: { attempts: number; lastAttempt: string } | null;

      if (retryDataJson) {
        try {
          retryData = JSON.parse(retryDataJson);
        } catch (error) {
          console.warn('[SyncOrchestrator] Corrupted retry data, resetting:', error);
          retryData = { attempts: 0, lastAttempt: new Date().toISOString() };
        }
      } else {
        retryData = { attempts: 0, lastAttempt: new Date().toISOString() };
      }

      if (retryData && retryData.attempts < MAX_RETRY_ATTEMPTS) {
        const nextAttemptCount = retryData.attempts + 1;
        const backoff = this.calculateBackoff(retryData.attempts);
        const nextRetryAt = new Date(Date.now() + backoff).toISOString();

        await AsyncStorage.setItem(
          BIDIRECTIONAL_SYNC_RETRY_KEY,
          JSON.stringify({
            attempts: nextAttemptCount,
            lastAttempt: nextRetryAt,
          })
        );

        console.log(
          `[SyncOrchestrator] Retry ${nextAttemptCount} scheduled at ${nextRetryAt}`
        );
      } else {
        // Max retries exceeded - mark as permanently failed
        console.warn(
          '[SyncOrchestrator] Max retry attempts exceeded - manual intervention required'
        );
      }
    } catch (error) {
      console.error('[SyncOrchestrator] Failed to schedule retry:', error);
    }
  }

  /**
   * Save last sync timestamp
   */
  private async saveLastSyncTime(): Promise<void> {
    try {
      await AsyncStorage.setItem(
        '@simpo_last_full_sync',
        new Date().toISOString()
      );
    } catch (error) {
      console.error('[SyncOrchestrator] Failed to save last sync time:', error);
    }
  }

  /**
   * Get human-readable phase description
   */
  private getPhaseDescription(phase: SyncPhase): string | null {
    const descriptions: Record<SyncPhase, string> = {
      idle: null,
      uploading: 'Uploading pending transactions',
      downloading_stock: 'Downloading stock levels',
      downloading_products: 'Downloading new products',
      downloading_user: 'Downloading user data',
      synced: 'All data synchronized',
      failed: 'Sync failed - will retry',
    };

    return descriptions[phase];
  }
}

// Export as SyncOrchestrator for clarity
export { SyncOrchestratorClass as SyncOrchestrator };
export default SyncOrchestratorClass.getInstance();
