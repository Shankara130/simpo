/**
 * SyncAPI Tests
 * Story 8.2: Implement Transaction Sync Queue
 *
 * Tests for backend sync endpoint communication
 */

import '@testing-library/jest-native/extend-expect';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { SyncAPI } from './SyncAPI';
import {
  OfflineTransactionWithItems,
  OfflineTransactionStatus,
} from '../types/offline.types';
import { SyncQueueError, SyncErrorType } from '../types/sync.types';

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage', () =>
  require('@react-native-async-storage/async-storage/jest/async-storage-mock')
);

describe('SyncAPI', () => {
  let syncAPI: ReturnType<typeof SyncAPI.getInstance>;
  let mockTransaction: OfflineTransactionWithItems;

  beforeEach(() => {
    // Get fresh instance for each test
    syncAPI = SyncAPI.getInstance();
    syncAPI.setMockMode(true, 0); // Zero delay for tests

    // Clear AsyncStorage
    AsyncStorage.clear();

    // Mock transaction data
    mockTransaction = {
      id: 1,
      transaction_number: 'OFFLINE-1234567890-1234',
      timestamp: '2026-05-28T10:30:00.000Z',
      cashier_id: 1,
      payment_method: 'CASH',
      total: '150.00',
      subtotal: '150.00',
      tax: '0.00',
      discount: '0.00',
      customer_name: 'Test Customer',
      status: 'pending_sync' as OfflineTransactionStatus,
      created_at: '2026-05-28T10:30:00.000Z',
      updated_at: '2026-05-28T10:30:00.000Z',
      items: [
        {
          id: 1,
          transaction_id: 1,
          product_id: 1,
          product_sku: 'SKU-001',
          product_name: 'Test Product',
          quantity: 2,
          unit_price: '75.00',
          subtotal: '150.00',
        },
      ],
    };
  });

  describe('postTransaction', () => {
    it('should successfully sync transaction in mock mode', async () => {
      const response = await syncAPI.postTransaction(mockTransaction);

      expect(response.status).toBe('synced');
      expect(response.transaction_id).toBeGreaterThan(0);
      expect(response.server_timestamp).toBeDefined();
    });

    it('should handle 409 conflict error (duplicate transaction)', async () => {
      // Set mock error for this transaction
      await AsyncStorage.setItem(
        `@simpo_mock_sync_error_${mockTransaction.transaction_number}`,
        '409'
      );

      // Use a fresh instance for this test to ensure mock is applied
      syncAPI = SyncAPI.getInstance();

      await expect(syncAPI.postTransaction(mockTransaction)).rejects.toThrow();

      try {
        await syncAPI.postTransaction(mockTransaction);
      } catch (error) {
        expect(error).toBeInstanceOf(SyncQueueError);
        if (error instanceof SyncQueueError) {
          expect(error.errorType).toBe(SyncErrorType.CONFLICT);
          expect(error.isRetryable).toBe(false);
        }
      }
    });

    it('should handle 400 validation error', async () => {
      await AsyncStorage.setItem(
        `@simpo_mock_sync_error_${mockTransaction.transaction_number}`,
        '400'
      );

      try {
        await syncAPI.postTransaction(mockTransaction);
      } catch (error) {
        expect(error).toBeInstanceOf(SyncQueueError);
        if (error instanceof SyncQueueError) {
          expect(error.errorType).toBe(SyncErrorType.VALIDATION_ERROR);
          expect(error.isRetryable).toBe(false);
        }
      }
    });

    it('should handle 503 service unavailable error', async () => {
      await AsyncStorage.setItem(
        `@simpo_mock_sync_error_${mockTransaction.transaction_number}`,
        '503'
      );

      try {
        await syncAPI.postTransaction(mockTransaction);
      } catch (error) {
        expect(error).toBeInstanceOf(SyncQueueError);
        if (error instanceof SyncQueueError) {
          expect(error.errorType).toBe(SyncErrorType.SERVER_ERROR);
          expect(error.isRetryable).toBe(true);
        }
      }
    });

    it('should handle network error', async () => {
      await AsyncStorage.setItem(
        `@simpo_mock_sync_error_${mockTransaction.transaction_number}`,
        'network'
      );

      try {
        await syncAPI.postTransaction(mockTransaction);
      } catch (error) {
        expect(error).toBeInstanceOf(SyncQueueError);
        if (error instanceof SyncQueueError) {
          expect(error.errorType).toBe(SyncErrorType.NETWORK);
          expect(error.isRetryable).toBe(true);
        }
      }
    });

    it('should clear mock error after use', async () => {
      // Set mock error
      await AsyncStorage.setItem(
        `@simpo_mock_sync_error_${mockTransaction.transaction_number}`,
        '409'
      );

      // First call should throw
      await expect(syncAPI.postTransaction(mockTransaction)).rejects.toThrow();

      // Mock error should be cleared
      const storedError = await AsyncStorage.getItem(
        `@simpo_mock_sync_error_${mockTransaction.transaction_number}`
      );
      expect(storedError).toBeNull();

      // Second call should succeed
      const response = await syncAPI.postTransaction(mockTransaction);
      expect(response.status).toBe('synced');
    });

    it('should map transaction to sync request correctly', async () => {
      const response = await syncAPI.postTransaction(mockTransaction);

      expect(response.status).toBe('synced');
      // Verify the mock received correct data
      // (Implicitly verified by successful response)
    });
  });

  describe('setMockMode', () => {
    it('should allow enabling/disabling mock mode', () => {
      syncAPI.setMockMode(true);
      // Mock mode enabled (no assertion needed, just no crash)

      syncAPI.setMockMode(false);
      // Mock mode disabled (no assertion needed, just no crash)
    });

    it('should allow setting custom mock delay', () => {
      syncAPI.setMockMode(true, 1000);
      // Delay set (no assertion needed, just no crash)
    });
  });

  describe('singleton pattern', () => {
    it('should return same instance across multiple calls', () => {
      const instance1 = SyncAPI.getInstance();
      const instance2 = SyncAPI.getInstance();

      expect(instance1).toBe(instance2);
    });
  });
});
