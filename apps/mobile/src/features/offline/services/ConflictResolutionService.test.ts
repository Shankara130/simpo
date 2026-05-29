/**
 * ConflictResolutionService Tests
 * Story 8.3: Implement Bidirectional Data Synchronization
 *
 * Test coverage for conflict resolution and manual overrides
 */

import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import AsyncStorage from '@react-native-async-storage/async-storage';

// Mock dependencies
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  getAllKeys: jest.fn(),
}));

jest.mock('./OfflineStorageService', () => {
  const mockInstance = {
    initialize: jest.fn(),
    getPendingTransactions: jest.fn(),
    markTransactionSynced: jest.fn(),
    markAndDeleteTransaction: jest.fn(),
    getInstance: jest.fn(() => mockInstance),
  };
  return {
    __esModule: true,
    default: mockInstance,
    OfflineStorageService: {
      getInstance: jest.fn(() => mockInstance),
    },
  };
});

jest.mock('./SyncAPI', () => ({
  SyncAPI: {
    getInstance: jest.fn(() => ({
      postTransaction: jest.fn(),
      setMockMode: jest.fn(),
    })),
  },
}));

// Import after mocks
import { ConflictResolutionService } from './ConflictResolutionService';
import {
  ConflictErrorResponse,
  FailedTransactionInfo,
  ManualOverrideRequest,
} from '../types/bidirectional-sync.types';

describe('ConflictResolutionService', () => {
  let conflictResolution: ReturnType<typeof ConflictResolutionService.getInstance>;
  const mockGetItem = AsyncStorage.getItem as jest.Mock;
  const mockSetItem = AsyncStorage.setItem as jest.Mock;
  const mockRemoveItem = AsyncStorage.removeItem as jest.Mock;
  const mockGetAllKeys = AsyncStorage.getAllKeys as jest.Mock;

  beforeEach(() => {
    jest.clearAllMocks();

    // Default AsyncStorage behavior
    mockGetItem.mockResolvedValue(null);
    mockSetItem.mockResolvedValue(undefined);
    mockRemoveItem.mockResolvedValue(undefined);
    mockGetAllKeys.mockResolvedValue([]);

    conflictResolution = ConflictResolutionService.getInstance();
  });

  describe('parseConflictError()', () => {
    it('should parse RFC 7807 conflict error response', () => {
      const errorResponse = {
        type: 'https://api.simpo.com/errors/conflict-insufficient-stock',
        title: 'Insufficient Stock',
        status: 409,
        detail: 'Product SKU-12345 has insufficient stock. Requested: 10, Available: 5',
        instance: '/api/v1/sync',
        transaction_id: 'OFFLINE-1716930000000-abc123',
        available_stock: 5,
        requested_quantity: 10,
        product_id: 123,
        product_sku: 'SKU-12345',
      };

      const result = conflictResolution.parseConflictError(errorResponse);

      expect(result).toEqual(errorResponse);
    });

    it('should return null for non-conflict errors', () => {
      const errorResponse = {
        type: 'https://api.simpo.com/errors/validation-error',
        title: 'Validation Error',
        status: 400,
        detail: 'Invalid data format',
      };

      const result = conflictResolution.parseConflictError(errorResponse);

      expect(result).toBeNull();
    });

    it('should parse 409 status as conflict error', () => {
      const errorResponse = {
        status: 409,
        detail: 'Conflict occurred',
      };

      const result = conflictResolution.parseConflictError(errorResponse);

      expect(result).not.toBeNull();
      expect(result?.status).toBe(409);
      expect(result?.title).toBe('Conflict');
    });

    it('should return null for other status codes', () => {
      const errorResponse = {
        status: 400,
        detail: 'Bad request',
      };

      const result = conflictResolution.parseConflictError(errorResponse);

      expect(result).toBeNull();
    });
  });

  describe('isConflictError()', () => {
    it('should return true for conflict errors', () => {
      const errorResponse = {
        type: 'https://api.simpo.com/errors/conflict-insufficient-stock',
        title: 'Insufficient Stock',
        status: 409,
      };

      expect(conflictResolution.isConflictError(errorResponse)).toBe(true);
    });

    it('should return false for non-conflict errors', () => {
      const errorResponse = {
        type: 'https://api.simpo.com/errors/validation-error',
        title: 'Validation Error',
        status: 400,
      };

      expect(conflictResolution.isConflictError(errorResponse)).toBe(false);
    });
  });

  describe('markTransactionFailed()', () => {
    it('should mark transaction as failed with conflict error', async () => {
      const conflictError: ConflictErrorResponse = {
        type: 'https://api.simpo.com/errors/conflict-insufficient-stock',
        title: 'Insufficient Stock',
        status: 409,
        detail: 'Insufficient stock',
        instance: '/api/v1/sync',
        transaction_id: 'TX-123',
        available_stock: 5,
        requested_quantity: 10,
      };

      await conflictResolution.markTransactionFailed(1, conflictError);

      // Verify failed transaction was persisted
      expect(mockSetItem).toHaveBeenCalledWith(
        '@simpo_failed_tx_1',
        expect.stringContaining('TX-123')
      );
    });

    it('should persist failed transaction to AsyncStorage', async () => {
      const conflictError: ConflictErrorResponse = {
        type: 'https://api.simpo.com/errors/conflict-insufficient-stock',
        title: 'Insufficient Stock',
        status: 409,
        detail: 'Insufficient stock',
        instance: '/api/v1/sync',
        transaction_id: 'TX-456',
      };

      await conflictResolution.markTransactionFailed(100, conflictError);

      expect(mockSetItem).toHaveBeenCalledWith(
        '@simpo_failed_tx_100',
        expect.any(String)
      );
    });
  });

  describe('getFailedTransaction()', () => {
    it('should return null for non-existent failed transaction', async () => {
      mockGetItem.mockResolvedValue(null);

      const result = await conflictResolution.getFailedTransaction(999);

      expect(result).toBeNull();
    });

    it('should return cached failed transaction', async () => {
      const conflictError: ConflictErrorResponse = {
        type: 'https://api.simpo.com/errors/conflict-insufficient-stock',
        title: 'Insufficient Stock',
        status: 409,
        detail: 'Insufficient stock',
        instance: '/api/v1/sync',
        transaction_id: 'TX-123',
      };

      const failedInfo: FailedTransactionInfo = {
        transactionId: 1,
        transactionNumber: 'TX-123',
        conflictError,
        timestamp: '2026-05-29T10:00:00Z',
        canOverride: true,
        requiresAdminAuth: true,
      };

      mockGetItem.mockResolvedValue(JSON.stringify(failedInfo));

      const result = await conflictResolution.getFailedTransaction(1);

      expect(result).toMatchObject({
        transactionId: 1,
        transactionNumber: 'TX-123',
        canOverride: true,
        requiresAdminAuth: true,
      });
      expect(result?.conflictError.type).toBe(conflictError.type);
    });
  });

  describe('getAllFailedTransactions()', () => {
    it('should return empty array when no failed transactions', async () => {
      mockGetAllKeys.mockResolvedValue([]);

      const result = await conflictResolution.getAllFailedTransactions();

      expect(result).toEqual([]);
    });

    it('should return all failed transactions sorted by timestamp', async () => {
      const failedKeys = [
        '@simpo_failed_tx_1',
        '@simpo_failed_tx_2',
        '@simpo_other_key',
      ];
      mockGetAllKeys.mockResolvedValue(failedKeys as any);

      const tx1 = {
        transactionId: 1,
        transactionNumber: 'TX-001',
        timestamp: '2026-05-29T10:05:00Z',
        conflictError: {} as any,
        canOverride: true,
        requiresAdminAuth: true,
      };

      const tx2 = {
        transactionId: 2,
        transactionNumber: 'TX-002',
        timestamp: '2026-05-29T10:00:00Z',
        conflictError: {} as any,
        canOverride: true,
        requiresAdminAuth: true,
      };

      mockGetItem.mockImplementation((key) => {
        if (key === '@simpo_failed_tx_1') {
          return Promise.resolve(JSON.stringify(tx1));
        }
        if (key === '@simpo_failed_tx_2') {
          return Promise.resolve(JSON.stringify(tx2));
        }
        return Promise.resolve(null);
      });

      const result = await conflictResolution.getAllFailedTransactions();

      // Should be sorted by timestamp (oldest first)
      expect(result).toHaveLength(2);
      expect(result[0].transactionId).toBe(2); // Older timestamp
      expect(result[1].transactionId).toBe(1); // Newer timestamp
    });
  });

  describe('requestManualOverride()', () => {
    it('should return failure for non-existent transaction', async () => {
      mockGetItem.mockResolvedValue(null);

      const request: ManualOverrideRequest = {
        transactionId: 999,
        adminUserId: 1,
        reason: 'Test override',
        forceProcessing: true,
      };

      const result = await conflictResolution.requestManualOverride(request);

      expect(result.success).toBe(false);
      expect(result.message).toContain('not found');
    });

    it('should require admin authorization when needed', async () => {
      const failedInfo: FailedTransactionInfo = {
        transactionId: 1,
        transactionNumber: 'TX-123',
        conflictError: {
          type: 'https://api.simpo.com/errors/conflict-insufficient-stock',
          title: 'Insufficient Stock',
          status: 409,
          detail: 'Insufficient stock',
          instance: '/api/v1/sync',
        },
        timestamp: '2026-05-29T10:00:00Z',
        canOverride: true,
        requiresAdminAuth: true,
      };

      mockGetItem.mockResolvedValue(JSON.stringify(failedInfo));

      const request: ManualOverrideRequest = {
        transactionId: 1,
        adminUserId: 0, // No admin
        reason: 'Test override',
        forceProcessing: true,
      };

      const result = await conflictResolution.requestManualOverride(request);

      expect(result.success).toBe(false);
      expect(result.message).toContain('Admin authorization required');
    });

    it('should perform override with valid admin authorization', async () => {
      const failedInfo: FailedTransactionInfo = {
        transactionId: 1,
        transactionNumber: 'TX-123',
        conflictError: {
          type: 'https://api.simpo.com/errors/conflict-insufficient-stock',
          title: 'Insufficient Stock',
          status: 409,
          detail: 'Insufficient stock',
          instance: '/api/v1/sync',
        },
        timestamp: '2026-05-29T10:00:00Z',
        canOverride: true,
        requiresAdminAuth: true,
      };

      mockGetItem.mockResolvedValue(JSON.stringify(failedInfo));

      const request: ManualOverrideRequest = {
        transactionId: 1,
        adminUserId: 1, // Valid admin
        reason: 'Override for urgent customer',
        forceProcessing: true,
      };

      const result = await conflictResolution.requestManualOverride(request);

      expect(result.success).toBe(true);
      expect(result.transactionId).toBe(1);

      // Verify failed transaction was cleared
      expect(mockRemoveItem).toHaveBeenCalledWith('@simpo_failed_tx_1');
    });
  });

  describe('getAuditLog()', () => {
    it('should return empty array when no audit log exists', async () => {
      mockGetItem.mockResolvedValue(null);

      const result = await conflictResolution.getAuditLog();

      expect(result).toEqual([]);
    });

    it('should return audit log entries', async () => {
      const auditLog = [
        { action: 'test1', timestamp: '2026-05-29T10:00:00Z' },
        { action: 'test2', timestamp: '2026-05-29T11:00:00Z' },
      ];

      mockGetItem.mockResolvedValue(JSON.stringify(auditLog));

      const result = await conflictResolution.getAuditLog();

      expect(result).toHaveLength(2);
      // Should be in reverse chronological order
      expect(result[0].action).toBe('test2');
      expect(result[1].action).toBe('test1');
    });

    it('should respect limit parameter', async () => {
      const auditLog = [
        { action: 'test1', timestamp: '2026-05-29T10:00:00Z' },
        { action: 'test2', timestamp: '2026-05-29T11:00:00Z' },
        { action: 'test3', timestamp: '2026-05-29T12:00:00Z' },
      ];

      mockGetItem.mockResolvedValue(JSON.stringify(auditLog));

      const result = await conflictResolution.getAuditLog(2);

      expect(result).toHaveLength(2);
    });
  });
});
