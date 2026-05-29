/**
 * Story 8.2 Integration Tests
 * Story 8.2: Implement Transaction Sync Queue
 *
 * End-to-end tests for transaction synchronization flow
 */

import '@testing-library/jest-native/extend-expect';
import AsyncStorage from '@react-native-async-storage/async-storage';
import OfflineStorageService from '../../services/OfflineStorageService';
import { SyncQueue } from '../../services/SyncQueueService';
import { useNetworkStatus } from '../../hooks/useNetworkStatus';
import { useSyncProgress } from '../../hooks/useSyncProgress';
import {
  OfflineTransactionWithItems,
  OfflineTransactionStatus,
} from '../../types/offline.types';
import {
  SyncQueueError,
  SyncErrorType,
  BASE_RETRY_DELAY_MS,
  MAX_RETRY_ATTEMPTS,
} from '../../types/sync.types';
import { SaleRequest } from '../../../pos/types/transaction.types';

// Mock dependencies
jest.mock('@react-native-async-storage/async-storage', () =>
  require('@react-native-async-storage/async-storage/jest/async-storage-mock')
);
jest.mock('expo-sqlite');
jest.mock('../../services/SyncAPI');

describe('Story 8.2 Integration Tests', () => {
  let syncQueue: SyncQueue;
  let mockSyncAPI: any;

  beforeEach(async () => {
    jest.clearAllMocks();
    await AsyncStorage.clear();

    // Initialize offline storage
    await OfflineStorageService.initialize();

    // Setup SyncAPI mock
    const { SyncAPI: SyncAPIModule } = require('../../services/SyncAPI');
    SyncAPIModule.setMockMode(true, 0);
    mockSyncAPI = SyncAPIModule.getInstance();

    mockSyncAPI.postTransaction.mockResolvedValue({
      status: 'synced',
      transaction_id: Date.now(),
      server_timestamp: new Date().toISOString(),
    });

    // Get sync queue instance
    syncQueue = SyncQueue.getInstance();
  });

  afterEach(async () => {
    await OfflineStorageService.close();
  });

  describe('AC1: Pending Transaction Identification and Ordering', () => {
    it('should retrieve pending transactions in chronological order', async () => {
      // Create test transactions with different timestamps
      const saleRequest: SaleRequest = {
        items: [{ product_id: 1, quantity: 1, unit_price: '100.00' }],
        payment_method: 'CASH',
        tax_amount: '0.00',
        discount_amount: '0.00',
      };

      const cashierId = 1;

      // Save first transaction
      await new Promise((resolve) => setTimeout(resolve, 10)); // Small delay for different timestamps
      const tx1 = await OfflineStorageService.saveTransaction(
        saleRequest,
        cashierId
      );

      // Save second transaction
      await new Promise((resolve) => setTimeout(resolve, 10)); // Small delay for different timestamps
      const tx2 = await OfflineStorageService.saveTransaction(
        saleRequest,
        cashierId
      );

      // Get pending transactions
      const pending = await syncQueue.getPendingTransactions();

      // Should have 2 transactions
      expect(pending.length).toBe(2);

      // Should be ordered by created_at (oldest first)
      expect(pending[0].id).toBeLessThan(pending[1].id);
    });

    it('should return empty array when no pending transactions', async () => {
      const pending = await syncQueue.getPendingTransactions();
      expect(pending).toEqual([]);
    });
  });

  describe('AC3: Exponential Backoff Retry', () => {
    it('should calculate exponential backoff correctly', () => {
      expect(syncQueue.calculateBackoff(0)).toBe(BASE_RETRY_DELAY_MS); // 1 min
      expect(syncQueue.calculateBackoff(1)).toBe(BASE_RETRY_DELAY_MS * 2); // 2 min
      expect(syncQueue.calculateBackoff(2)).toBe(BASE_RETRY_DELAY_MS * 4); // 4 min
      expect(syncQueue.calculateBackoff(3)).toBe(BASE_RETRY_DELAY_MS * 8); // 8 min
      expect(syncQueue.calculateBackoff(4)).toBe(BASE_RETRY_DELAY_MS * 32); // 32 min (capped)
    });

    it('should enforce max retry limit', async () => {
      const transactionId = 1;

      // Schedule max + 1 retries
      await syncQueue.scheduleRetry(transactionId, MAX_RETRY_ATTEMPTS);

      const shouldRetry = await syncQueue.shouldRetryNow(transactionId);
      expect(shouldRetry).toBe(false); // Max exceeded
    });
  });

  describe('AC5: Queue Persistence and Crash Recovery', () => {
    it('should save and load queue state', async () => {
      const state = {
        isProcessing: true,
        currentTransactionId: 1,
        lastUpdated: new Date().toISOString(),
      };

      await syncQueue.saveQueueState(state);

      const loaded = await syncQueue.loadQueueState();

      expect(loaded.isProcessing).toBe(true);
      expect(loaded.currentTransactionId).toBe(1);
      expect(loaded.lastUpdated).toBe(state.lastUpdated);
    });

    it('should detect orphaned processing state after crash', async () => {
      // Set processing state 10 minutes ago
      const oldTimestamp = new Date(
        Date.now() - 10 * 60 * 1000
      ).toISOString();

      const state = {
        isProcessing: true,
        currentTransactionId: 1,
        lastUpdated: oldTimestamp,
      };

      await syncQueue.saveQueueState(state);

      const isOrphaned = await syncQueue.isProcessingStateOrphaned();

      expect(isOrphaned).toBe(true);
    });
  });

  describe('AC2 & AC7: Sequential Processing with Error Handling', () => {
    it('should handle 409 conflict (duplicate transaction)', async () => {
      // Create test transaction
      const saleRequest: SaleRequest = {
        items: [{ product_id: 1, quantity: 1, unit_price: '100.00' }],
        payment_method: 'CASH',
        tax_amount: '0.00',
        discount_amount: '0.00',
      };

      const tx = await OfflineStorageService.saveTransaction(saleRequest, 1);

      // Mock 409 response
      mockSyncAPI.postTransaction.mockImplementationOnce(() => {
        throw new SyncQueueError(
          'Duplicate',
          SyncErrorType.CONFLICT,
          { status: 409 },
          false
        );
      });

      // Process queue
      const result = await syncQueue.processQueue();

      expect(result.totalProcessed).toBe(1);
      expect(result.skippedCount).toBe(1); // Duplicates are skipped
    });

    it('should handle 400 validation error (not retryable)', async () => {
      const saleRequest: SaleRequest = {
        items: [{ product_id: 1, quantity: 1, unit_price: '100.00' }],
        payment_method: 'CASH',
        tax_amount: '0.00',
        discount_amount: '0.00',
      };

      await OfflineStorageService.saveTransaction(saleRequest, 1);

      // Mock 400 response
      mockSyncAPI.postTransaction.mockImplementationOnce(() => {
        throw new SyncQueueError(
          'Invalid data',
          SyncErrorType.VALIDATION_ERROR,
          { status: 400 },
          false
        );
      });

      const result = await syncQueue.processQueue();

      expect(result.failedCount).toBe(1);
      // Validation errors should not be retried
    });
  });

  describe('useSyncProgress Hook', () => {
    it('should initialize with default state', async () => {
      const { result } = require('@testing-library/react-hooks').renderHook(
        () => useSyncProgress()
      );

      // Wait for initialization
      await new Promise((resolve) => setTimeout(resolve, 100));

      expect(result.current.pendingCount).toBe(0);
      expect(result.current.syncedCount).toBe(0);
      expect(result.current.failedCount).toBe(0);
    });
  });

  describe('No Breaking Changes to Story 8-1', () => {
    it('should maintain OfflineStorageService functionality', async () => {
      const saleRequest: SaleRequest = {
        items: [{ product_id: 1, quantity: 2, unit_price: '50.00' }],
        payment_method: 'CASH',
        tax_amount: '0.00',
        discount_amount: '0.00',
      };

      // Should still work
      const tx = await OfflineStorageService.saveTransaction(saleRequest, 1);

      expect(tx).toBeDefined();
      expect(tx.status).toBe('pending_sync');

      // Get pending should still work
      const pending = await OfflineStorageService.getPendingTransactions();
      expect(pending.length).toBe(1);
    });

    it('should maintain useNetworkStatus functionality', async () => {
      // useNetworkStatus should still be importable and functional
      const { useNetworkStatus } = require('../../hooks/useNetworkStatus');

      expect(typeof useNetworkStatus).toBe('function');
    });
  });
});
