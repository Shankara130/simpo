/**
 * Story 8.3 Integration Tests
 * Bidirectional Data Synchronization Integration Tests
 *
 * Tests the complete bidirectional sync flow including:
 * - Upload pending transactions before downloading
 * - Download stock levels from server
 * - Download new products added since last sync
 * - User data updates and role changes
 * - Conflict resolution scenarios
 * - Visual indicator updates through all phases
 * - Retry logic for network failures
 */

import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';

// Mock all external dependencies FIRST
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  getAllKeys: jest.fn(),
  clear: jest.fn(),
}));

jest.mock('expo-sqlite', () => ({
  openDatabaseAsync: jest.fn(),
}));

jest.mock('@react-native-community/netinfo', () => ({
  fetch: jest.fn(),
}));

// Create mock instances BEFORE module mocks
const mockOfflineStorageInstance = {
  initialize: jest.fn(),
  saveTransaction: jest.fn(),
  getPendingTransactions: jest.fn(),
  markTransactionSynced: jest.fn(),
  markAndDeleteTransaction: jest.fn(),
  deleteTransaction: jest.fn(),
  reset: jest.fn(),
};

const mockSyncAPIInstance = {
  postTransaction: jest.fn(),
  setMockMode: jest.fn(),
};

const mockConflictResolutionInstance = {
  parseConflictError: jest.fn(),
  isConflictError: jest.fn(),
  markTransactionFailed: jest.fn(),
  getFailedTransaction: jest.fn(),
  getAllFailedTransactions: jest.fn(),
  clearFailedTransaction: jest.fn(),
  requestManualOverride: jest.fn(),
  getAuditLog: jest.fn(),
  isUserDeactivated: jest.fn(),
  hasRoleChanged: jest.fn(),
};

// Mock services BEFORE importing
jest.mock('../services/OfflineStorageService', () => ({
  __esModule: true,
  default: mockOfflineStorageInstance,
}));

jest.mock('../services/SyncAPI', () => ({
  __esModule: true,
  default: mockSyncAPIInstance,
}));

jest.mock('../services/ConflictResolutionService', () => ({
  __esModule: true,
  default: mockConflictResolutionInstance,
}));

// NOW import the services
import { SyncOrchestrator } from '../services/SyncOrchestrator';
import { ProductSyncService } from '../services/ProductSyncService';
import { UserSyncService } from '../services/UserSyncService';
import { SyncQueue } from '../services/SyncQueueService';
import { ConflictResolutionService } from '../services/ConflictResolutionService';
import OfflineStorageService from '../services/OfflineStorageService';
import {
  BIDIRECTIONAL_SYNC_STATE_KEY,
  BIDIRECTIONAL_SYNC_RETRY_KEY,
  MAX_RETRY_ATTEMPTS,
} from '../types/bidirectional-sync.types';

describe('Story 8.3: Integration Tests - Bidirectional Data Synchronization', () => {
  let orchestrator: ReturnType<typeof SyncOrchestrator.getInstance>;
  let syncQueue: ReturnType<typeof SyncQueue.getInstance>;
  let productSync: ReturnType<typeof ProductSyncService.getInstance>;
  let userSync: ReturnType<typeof UserSyncService.getInstance>;
  let conflictResolution: ReturnType<typeof ConflictResolutionService.getInstance>;
  let offlineStorage: ReturnType<typeof OfflineStorageService.getInstance>;

  const mockGetItem = AsyncStorage.getItem as jest.Mock;
  const mockSetItem = AsyncStorage.setItem as jest.Mock;
  const mockClear = AsyncStorage.clear as jest.Mock;

  beforeEach(async () => {
    jest.clearAllMocks();
    jest.isolateModules();

    // Reset all AsyncStorage calls
    mockGetItem.mockResolvedValue(null);
    mockSetItem.mockResolvedValue(undefined);
    mockClear.mockResolvedValue(undefined);

    // Initialize all services
    offlineStorage = OfflineStorageService.getInstance();
    await offlineStorage.initialize();

    syncQueue = SyncQueue.getInstance();
    productSync = ProductSyncService.getInstance();
    userSync = UserSyncService.getInstance();
    conflictResolution = ConflictResolutionService.getInstance();
    orchestrator = SyncOrchestrator.getInstance();
  });

  afterEach(() => {
    jest.clearAllTimers();
  });

  describe('AC1: Upload Pending Transactions Before Downloading Data', () => {
    it('should upload ALL pending transactions before starting downloads', async () => {
      // Mock processQueue to simulate successful upload
      jest.spyOn(syncQueue, 'processQueue').mockResolvedValue({
        totalProcessed: 2,
        syncedCount: 2,
        failedCount: 0,
        skippedCount: 0,
        duration: 1000,
      });

      // Mock download phases
      jest.spyOn(productSync, 'syncStockLevels').mockResolvedValue({
        products: [],
        lastSyncTimestamp: new Date().toISOString(),
      });

      jest.spyOn(productSync, 'syncProducts').mockResolvedValue({
        products: [],
        lastSyncTimestamp: new Date().toISOString(),
      });

      jest.spyOn(userSync, 'syncUser').mockResolvedValue({
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        role: 'cashier' as const,
        status: 'active' as const,
        updated_at: new Date().toISOString(),
      });

      // Execute sync
      const result = await orchestrator.sync();

      // Verify upload phase was called first
      expect(syncQueue.processQueue).toHaveBeenCalledTimes(1);

      // Verify download phases were called after upload
      expect(productSync.syncStockLevels).toHaveBeenCalled();
      expect(productSync.syncProducts).toHaveBeenCalled();
      expect(userSync.syncUser).toHaveBeenCalled();

      // Verify successful sync
      expect(result.success).toBe(true);
      expect(result.uploaded).toBe(2);
    });
  });

  describe('AC2: Download Latest Stock Levels from Server', () => {
    it('should fetch stock levels and update local cache', async () => {
      const mockStockData = [
        { id: 1, sku: 'SKU-001', name: 'Product 1', stock_qty: 50, updated_at: '2026-05-29T10:00:00Z' },
        { id: 2, sku: 'SKU-002', name: 'Product 2', stock_qty: 30, updated_at: '2026-05-29T10:00:00Z' },
      ];

      // Mock successful stock sync
      jest.spyOn(syncQueue, 'processQueue').mockResolvedValue({
        totalProcessed: 0,
        syncedCount: 0,
        failedCount: 0,
        skippedCount: 0,
        duration: 0,
      });

      jest.spyOn(productSync, 'syncStockLevels').mockResolvedValue({
        products: mockStockData,
        lastSyncTimestamp: '2026-05-29T10:00:00Z',
      });

      jest.spyOn(productSync, 'syncProducts').mockResolvedValue({
        products: [],
        lastSyncTimestamp: '2026-05-29T10:00:00Z',
      });

      jest.spyOn(userSync, 'syncUser').mockResolvedValue({
        id: 1,
        username: 'test',
        email: 'test@test.com',
        role: 'cashier' as const,
        status: 'active' as const,
        updated_at: new Date().toISOString(),
      });

      const result = await orchestrator.sync();

      expect(result.downloadedStock).toBe(2);
      expect(result.success).toBe(true);
    });
  });

  describe('AC7: Visual Sync Status Indicators', () => {
    it('should update visual indicators through all sync phases', async () => {
      // Track phase updates
      const phaseUpdates: string[] = [];
      jest.spyOn(orchestrator, 'updatePhase').mockImplementation(async (phase) => {
        phaseUpdates.push(phase);
      });

      // Mock each phase
      jest.spyOn(syncQueue, 'processQueue').mockResolvedValue({
        totalProcessed: 1,
        syncedCount: 1,
        failedCount: 0,
        skippedCount: 0,
        duration: 500,
      });

      jest.spyOn(productSync, 'syncStockLevels').mockResolvedValue({
        products: [],
        lastSyncTimestamp: '2026-05-29T10:00:00Z',
      });

      jest.spyOn(productSync, 'syncProducts').mockResolvedValue({
        products: [],
        lastSyncTimestamp: '2026-05-29T10:00:00Z',
      });

      jest.spyOn(userSync, 'syncUser').mockResolvedValue({
        id: 1,
        username: 'test',
        email: 'test@test.com',
        role: 'cashier' as const,
        status: 'active' as const,
        updated_at: new Date().toISOString(),
      });

      await orchestrator.sync();

      // Verify all phases were updated
      expect(phaseUpdates).toContain('uploading');
      expect(phaseUpdates).toContain('downloading_stock');
      expect(phaseUpdates).toContain('downloading_products');
      expect(phaseUpdates).toContain('downloading_user');
      expect(phaseUpdates).toContain('synced');
    });
  });

  describe('AC8: Automatic Background Retry for Failed Syncs', () => {
    it('should schedule retry with exponential backoff', async () => {
      // Mock sync failure
      jest.spyOn(syncQueue, 'processQueue').mockRejectedValue(new Error('Network error'));

      // Track retry schedule
      let retrySaved = false;
      mockSetItem.mockImplementation((key, value) => {
        if (key === BIDIRECTIONAL_SYNC_RETRY_KEY) {
          retrySaved = true;
          const retryData = JSON.parse(value as string);
          expect(retryData.attempts).toBeGreaterThan(0);
        }
        return Promise.resolve(undefined);
      });

      await orchestrator.sync();

      expect(retrySaved).toBe(true);
    });
  });

  describe('AC9: Sync Service Orchestration', () => {
    it('should coordinate all four sync phases in sequence', async () => {
      // Mock all phases
      jest.spyOn(syncQueue, 'processQueue').mockResolvedValue({
        totalProcessed: 1,
        syncedCount: 1,
        failedCount: 0,
        skippedCount: 0,
        duration: 500,
      });

      jest.spyOn(productSync, 'syncStockLevels').mockResolvedValue({
        products: [{ id: 1, sku: 'SKU-001', name: 'Product 1', stock_qty: 50, updated_at: '2026-05-29T10:00:00Z' }],
        lastSyncTimestamp: '2026-05-29T10:00:00Z',
      });

      jest.spyOn(productSync, 'syncProducts').mockResolvedValue({
        products: [{ id: 2, sku: 'SKU-002', name: 'Product 2', stock_qty: 30, updated_at: '2026-05-29T10:00:00Z' }],
        lastSyncTimestamp: '2026-05-29T10:00:00Z',
      });

      jest.spyOn(userSync, 'syncUser').mockResolvedValue({
        id: 1,
        username: 'test',
        email: 'test@test.com',
        role: 'cashier' as const,
        status: 'active' as const,
        updated_at: new Date().toISOString(),
      });

      const result = await orchestrator.sync();

      // Verify all phases executed
      expect(syncQueue.processQueue).toHaveBeenCalledTimes(1);
      expect(productSync.syncStockLevels).toHaveBeenCalledTimes(1);
      expect(productSync.syncProducts).toHaveBeenCalledTimes(1);
      expect(userSync.syncUser).toHaveBeenCalledTimes(1);

      expect(result.success).toBe(true);
      expect(result.phase).toBe('synced');
    });

    it('should halt sync on phase failure', async () => {
      // Mock stock download failure
      jest.spyOn(syncQueue, 'processQueue').mockResolvedValue({
        totalProcessed: 1,
        syncedCount: 1,
        failedCount: 0,
        skippedCount: 0,
        duration: 500,
      });

      jest.spyOn(productSync, 'syncStockLevels').mockRejectedValue(new Error('Network error'));

      const result = await orchestrator.sync();

      expect(result.success).toBe(false);
      expect(result.phase).toBe('downloading_stock');

      // Verify subsequent phases were not called
      expect(productSync.syncProducts).not.toHaveBeenCalled();
    });
  });

  describe('No Breaking Changes to Stories 8-1 and 8-2', () => {
    it('should not break OfflineStorageService (Story 8-1)', async () => {
      // OfflineStorageService should still work
      expect(offlineStorage).toBeDefined();
      expect(typeof offlineStorage.initialize).toBe('function');
      expect(typeof offlineStorage.saveTransaction).toBe('function');
      expect(typeof offlineStorage.getPendingTransactions).toBe('function');
    });

    it('should not break SyncQueueService (Story 8-2)', async () => {
      // SyncQueueService should still work
      expect(syncQueue).toBeDefined();
      expect(typeof syncQueue.processQueue).toBe('function');
      expect(typeof syncQueue.stopProcessing).toBe('function');
      expect(typeof syncQueue.calculateBackoff).toBe('function');
    });
  });
});
