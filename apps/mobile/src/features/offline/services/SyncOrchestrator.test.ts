/**
 * SyncOrchestrator Tests
 * Story 8.3: Implement Bidirectional Data Synchronization
 *
 * Test coverage for bidirectional sync orchestration
 */

import { describe, it, expect, beforeEach, jest, afterEach } from '@jest/globals';
import AsyncStorage from '@react-native-async-storage/async-storage';

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
}));

// Mock OfflineStorageService
// The actual export is: export default OfflineStorageService.getInstance();
// So the default is already an instance, not the class
jest.mock('./OfflineStorageService', () => {
  const mockInstance = {
    initialize: jest.fn(),
    getPendingTransactions: jest.fn(),
    markTransactionSynced: jest.fn(),
    markAndDeleteTransaction: jest.fn(),
    getInstance: jest.fn(() => mockInstance), // Add getInstance to the instance
  };
  return {
    __esModule: true,
    default: mockInstance,
  };
});

// Mock SyncAPI
jest.mock('./SyncAPI', () => ({
  SyncAPI: {
    getInstance: jest.fn(() => ({
      postTransaction: jest.fn(),
      setMockMode: jest.fn(),
    })),
  },
}));

// Import after mock is set up
import { SyncOrchestrator } from './SyncOrchestrator';
import { SyncQueue } from './SyncQueueService';
import { ProductSyncService } from './ProductSyncService';
import { UserSyncService } from './UserSyncService';
import {
  SyncPhase,
  ProductSyncResponse,
  UserSyncResponse,
  BIDIRECTIONAL_SYNC_STATE_KEY,
  BIDIRECTIONAL_SYNC_RETRY_KEY,
} from '../types/bidirectional-sync.types';

describe('SyncOrchestrator', () => {
  let orchestrator: ReturnType<typeof SyncOrchestrator.getInstance>;
  let processQueueSpy: jest.SpyInstance;
  let syncStockLevelsSpy: jest.SpyInstance;
  let syncProductsSpy: jest.SpyInstance;
  let syncUserSpy: jest.SpyInstance;

  const mockGetItem = AsyncStorage.getItem as jest.Mock;
  const mockSetItem = AsyncStorage.setItem as jest.Mock;

  beforeEach(() => {
    jest.clearAllMocks();

    // Default AsyncStorage behavior
    mockGetItem.mockResolvedValue(null);
    mockSetItem.mockResolvedValue(undefined);

    // Get service instances
    const syncQueue = SyncQueue.getInstance();
    const productSync = ProductSyncService.getInstance();
    const userSync = UserSyncService.getInstance();

    // Spy on methods
    processQueueSpy = jest.spyOn(syncQueue, 'processQueue');
    syncStockLevelsSpy = jest.spyOn(productSync, 'syncStockLevels');
    syncProductsSpy = jest.spyOn(productSync, 'syncProducts');
    syncUserSpy = jest.spyOn(userSync, 'syncUser');

    // Set default return values
    processQueueSpy.mockResolvedValue({
      totalProcessed: 0,
      syncedCount: 0,
      failedCount: 0,
      skippedCount: 0,
      duration: 0,
    });

    syncStockLevelsSpy.mockResolvedValue({
      products: [],
      lastSyncTimestamp: new Date().toISOString(),
    });

    syncProductsSpy.mockResolvedValue({
      products: [],
      lastSyncTimestamp: new Date().toISOString(),
    });

    syncUserSpy.mockResolvedValue({
      id: 1,
      username: 'test',
      email: 'test@test.com',
      role: 'cashier',
      status: 'active',
      updated_at: new Date().toISOString(),
    });

    orchestrator = SyncOrchestrator.getInstance();
  });

  afterEach(() => {
    processQueueSpy.mockRestore();
    syncStockLevelsSpy.mockRestore();
    syncProductsSpy.mockRestore();
    syncUserSpy.mockRestore();
  });

  describe('sync()', () => {
    it('should execute all four phases', async () => {
      processQueueSpy.mockResolvedValue({
        totalProcessed: 1,
        syncedCount: 1,
        failedCount: 0,
        skippedCount: 0,
        duration: 100,
      });

      syncStockLevelsSpy.mockResolvedValue({
        products: [{ id: 1, sku: 'SKU001', name: 'Test', stock_qty: 10, updated_at: '2026-05-29T10:00:00Z' }],
        lastSyncTimestamp: '2026-05-29T10:00:00Z',
      });

      syncProductsSpy.mockResolvedValue({
        products: [{ id: 2, sku: 'SKU002', name: 'Test2', stock_qty: 20, updated_at: '2026-05-29T10:00:00Z' }],
        lastSyncTimestamp: '2026-05-29T10:00:00Z',
      });

      const result = await orchestrator.sync();

      expect(result.success).toBe(true);
      expect(result.uploaded).toBe(1);
      expect(result.downloadedStock).toBe(1);
      expect(result.downloadedProducts).toBe(1);
      expect(result.phase).toBe('synced');
    });

    it('should halt on upload failure', async () => {
      processQueueSpy.mockResolvedValue({
        totalProcessed: 2,
        syncedCount: 0,
        failedCount: 2,
        skippedCount: 0,
        duration: 100,
      });

      const result = await orchestrator.sync();

      expect(result.success).toBe(false);
      expect(result.phase).toBe('uploading');
      expect(syncStockLevelsSpy).not.toHaveBeenCalled();
    });

    it('should handle zero transactions', async () => {
      const result = await orchestrator.sync();

      expect(result.success).toBe(true);
      expect(result.uploaded).toBe(0);
    });
  });

  describe('updatePhase()', () => {
    it('should persist phase state', async () => {
      await orchestrator.updatePhase('uploading');

      expect(mockSetItem).toHaveBeenCalledWith(
        BIDIRECTIONAL_SYNC_STATE_KEY,
        expect.stringContaining('"phase":"uploading"')
      );
    });
  });

  describe('calculateBackoff()', () => {
    it('should calculate exponential backoff', () => {
      expect(orchestrator.calculateBackoff(0)).toBe(60000);
      expect(orchestrator.calculateBackoff(1)).toBe(120000);
      expect(orchestrator.calculateBackoff(2)).toBe(240000);
    });
  });
});
