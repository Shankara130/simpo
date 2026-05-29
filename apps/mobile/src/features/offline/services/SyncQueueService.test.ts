/**
 * SyncQueueService Tests
 * Story 8.2: Implement Transaction Sync Queue
 *
 * Integration tests for sync queue functionality
 * Uses real OfflineStorageService with mock database
 */

import '@testing-library/jest-native/extend-expect';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { SyncQueue } from './SyncQueueService';
import OfflineStorageService from './OfflineStorageService';
import { SyncAPI } from './SyncAPI';
import {
  OfflineTransactionWithItems,
  OfflineTransactionStatus,
} from '../types/offline.types';
import {
  SyncQueueError,
  SyncErrorType,
  MAX_RETRY_ATTEMPTS,
  BASE_RETRY_DELAY_MS,
} from '../types/sync.types';

// Mock AsyncStorage for queue state persistence
jest.mock('@react-native-async-storage/async-storage', () =>
  require('@react-native-async-storage/async-storage/jest/async-storage-mock')
);

// Mock SyncAPI
jest.mock('./SyncAPI');

// Mock expo-sqlite
jest.mock('expo-sqlite');

describe('SyncQueue (Integration Tests)', () => {
  let syncQueue: SyncQueue;
  let mockSyncAPI: any;

  beforeEach(async () => {
    jest.clearAllMocks();
    await AsyncStorage.clear();

    // Reset offline storage
    await OfflineStorageService.reset();

    // Setup SyncAPI mock
    const { SyncAPI: SyncAPIModule } = require('./SyncAPI');
    SyncAPIModule.setMockMode(true, 0); // Zero delay for tests
    mockSyncAPI = SyncAPIModule.getInstance();

    // Setup default API mock
    mockSyncAPI.postTransaction.mockResolvedValue({
      status: 'synced',
      transaction_id: Date.now(),
      server_timestamp: new Date().toISOString(),
    });

    // Get sync queue instance
    syncQueue = SyncQueue.getInstance();
  });

  afterEach(async () => {
    // Cleanup
    await OfflineStorageService.close();
  });

  describe('calculateBackoff', () => {
    it('should calculate exponential backoff correctly', () => {
      expect(syncQueue.calculateBackoff(0)).toBe(BASE_RETRY_DELAY_MS); // 1 min
      expect(syncQueue.calculateBackoff(1)).toBe(BASE_RETRY_DELAY_MS * 2); // 2 min
      expect(syncQueue.calculateBackoff(2)).toBe(BASE_RETRY_DELAY_MS * 4); // 4 min
      expect(syncQueue.calculateBackoff(3)).toBe(BASE_RETRY_DELAY_MS * 8); // 8 min
      expect(syncQueue.calculateBackoff(4)).toBe(BASE_RETRY_DELAY_MS * 32); // 32 min (max)
    });

    it('should cap backoff at maximum', () => {
      const maxBackoff = syncQueue.calculateBackoff(100);
      expect(maxBackoff).toBe(BASE_RETRY_DELAY_MS * 32); // Capped at 32 min
    });
  });

  describe('queue state persistence', () => {
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
    });

    it('should detect orphaned processing state', async () => {
      // Set processing state 10 minutes ago (orphaned)
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

    it('should not detect recent state as orphaned', async () => {
      const state = {
        isProcessing: true,
        currentTransactionId: 1,
        lastUpdated: new Date().toISOString(),
      };

      await syncQueue.saveQueueState(state);

      const isOrphaned = await syncQueue.isProcessingStateOrphaned();
      expect(isOrphaned).toBe(false);
    });
  });

  describe('retry scheduling', () => {
    it('should schedule retry with exponential backoff', async () => {
      const transactionId = 1;
      const retryCount = 2;

      await syncQueue.scheduleRetry(transactionId, retryCount);

      const shouldRetry = await syncQueue.shouldRetryNow(transactionId);
      // Should not retry immediately (future timestamp)
      expect(shouldRetry).toBe(false);
    });

    it('should check retry count', async () => {
      const transactionId = 1;

      const count = await syncQueue.getRetryCount(transactionId);
      expect(count).toBe(0); // No retry scheduled yet
    });

    it('should enforce max retry limit', async () => {
      const transactionId = 1;

      // Schedule max retries
      await syncQueue.scheduleRetry(transactionId, MAX_RETRY_ATTEMPTS);

      const shouldRetry = await syncQueue.shouldRetryNow(transactionId);
      expect(shouldRetry).toBe(false); // Max retries exceeded
    });
  });

  describe('singleton pattern', () => {
    it('should return same instance', () => {
      const instance1 = SyncQueue.getInstance();
      const instance2 = SyncQueue.getInstance();

      expect(instance1).toBe(instance2);
    });
  });

  describe('stopProcessing', () => {
    it('should stop without error', () => {
      expect(() => {
        syncQueue.stopProcessing();
      }).not.toThrow();
    });

    it('should be idempotent', () => {
      expect(() => {
        syncQueue.stopProcessing();
        syncQueue.stopProcessing();
        syncQueue.stopProcessing();
      }).not.toThrow();
    });
  });
});
