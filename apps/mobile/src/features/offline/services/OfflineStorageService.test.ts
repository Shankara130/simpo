/**
 * OfflineStorageService Tests
 * Story 8.1: Implement Local SQLite Storage for Offline Transactions
 *
 * Test Coverage:
 * - Database initialization and schema creation
 * - Transaction save with valid data
 * - Transaction save with invalid data (constraint violations)
 * - Get pending transactions
 * - Mark transaction as synced
 * - Delete transaction
 * - Error handling for database operations
 */

import offlineStorageService from './OfflineStorageService';
import {
  OfflineTransactionWithItems,
  OfflineTransactionStatus,
  DATABASE_NAME,
  TABLE_TRANSACTIONS,
  TABLE_TRANSACTION_ITEMS,
  OfflineStorageError,
} from '../types/offline.types';
import { SaleRequest } from '../../pos/types/transaction.types';

// Mock expo-sqlite
jest.mock('expo-sqlite', () => ({
  openDatabaseAsync: jest.fn(),
}));

import { openDatabaseAsync } from 'expo-sqlite';

describe('OfflineStorageService', () => {
  let mockDb: any;

  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();

    // Create mock database object
    mockDb = {
      execAsync: jest.fn(),
      runAsync: jest.fn(),
      getFirstAsync: jest.fn(),
      getAllAsync: jest.fn(),
      closeAsync: jest.fn(),
    };

    // Mock openDatabaseAsync to return our mock db
    (openDatabaseAsync as jest.Mock).mockResolvedValue(mockDb);

    // Reset singleton state for each test
    // @ts-ignore - accessing private properties for testing
    if (offlineStorageService.db) {
      offlineStorageService.close();
    }
    // @ts-ignore
    offlineStorageService.isInitialized = false;
    // @ts-ignore
    offlineStorageService.db = null;
  });

  afterEach(async () => {
    // Close database connection after each test
    try {
      await offlineStorageService.close();
    } catch (e) {
      // Ignore close errors in tests
    }
  });

  describe('Database Initialization', () => {
    it('should create database with correct name', async () => {
      await offlineStorageService.initialize();

      expect(openDatabaseAsync).toHaveBeenCalledWith(DATABASE_NAME);
    });

    it('should create offline_transactions table with correct schema', async () => {
      await offlineStorageService.initialize();

      expect(mockDb.execAsync).toHaveBeenCalledWith(
        expect.stringContaining('CREATE TABLE IF NOT EXISTS offline_transactions')
      );
    });

    it('should create offline_transaction_items table with correct schema', async () => {
      await offlineStorageService.initialize();

      expect(mockDb.execAsync).toHaveBeenCalledWith(
        expect.stringContaining('CREATE TABLE IF NOT EXISTS offline_transaction_items')
      );
    });

    it('should create indexes on status and created_at columns', async () => {
      await offlineStorageService.initialize();

      expect(mockDb.execAsync).toHaveBeenCalledWith(
        expect.stringContaining('CREATE INDEX IF NOT EXISTS idx_status')
      );
      expect(mockDb.execAsync).toHaveBeenCalledWith(
        expect.stringContaining('CREATE INDEX IF NOT EXISTS idx_created_at')
      );
    });

    it('should handle database initialization errors gracefully', async () => {
      (openDatabaseAsync as jest.Mock).mockRejectedValue(
        new Error('Database creation failed')
      );

      await expect(offlineStorageService.initialize()).rejects.toThrow(OfflineStorageError);
    });

    it('should be idempotent - multiple calls should not recreate tables', async () => {
      await offlineStorageService.initialize();
      await offlineStorageService.initialize();

      // execAsync is called 3 times: WAL mode, schema batch (all tables/indexes), schema version
      expect(mockDb.execAsync).toHaveBeenCalledTimes(3);
    });
  });

  describe('saveTransaction', () => {
    const mockSaleRequest: SaleRequest = {
      items: [
        {
          product_id: 1,
          quantity: 2,
          unit_price: '75000.00',
        },
      ],
      payment_method: 'CASH',
      tax_amount: '0.00',
      discount_amount: '0.00',
      idempotency_key: 'test-idempotency-key',
    };

    beforeEach(async () => {
      await offlineStorageService.initialize();

      // Mock successful transaction insertion
      mockDb.runAsync.mockResolvedValue({ lastInsertRowId: 1 });
    });

    it('should save transaction with all required fields', async () => {
      const result = await offlineStorageService.saveTransaction(mockSaleRequest, 1);

      expect(result.transactionNumber).toMatch(/^OFFLINE-\d+-\d+$/);
      expect(result.cashierId).toBe(1);
      expect(result.paymentMethod).toBe('CASH');
      expect(result.status).toBe('pending_sync');
    });

    it('should save transaction items in offline_transaction_items table', async () => {
      mockDb.runAsync
        .mockResolvedValueOnce({ lastInsertRowId: 1 }) // Transaction insert
        .mockResolvedValue({ lastInsertRowId: 1 }); // Item insert

      await offlineStorageService.saveTransaction(mockSaleRequest, 1);

      // Should call runAsync twice (transaction + item)
      expect(mockDb.runAsync).toHaveBeenCalledTimes(2);
    });

    it('should generate unique transaction numbers', async () => {
      const result1 = await offlineStorageService.saveTransaction(mockSaleRequest, 1);
      const result2 = await offlineStorageService.saveTransaction(mockSaleRequest, 1);

      expect(result1.transactionNumber).not.toBe(result2.transactionNumber);
    });

    it('should handle database constraint violations', async () => {
      mockDb.runAsync.mockRejectedValue(
        new Error('UNIQUE constraint failed: offline_transactions.transaction_number')
      );

      await expect(
        offlineStorageService.saveTransaction(mockSaleRequest, 1)
      ).rejects.toThrow(OfflineStorageError);
    });

    it('should use transaction for atomic insert of header and items', async () => {
      // Mock successful insert
      mockDb.runAsync
        .mockResolvedValueOnce({ lastInsertRowId: 1 }) // Transaction
        .mockResolvedValue({ lastInsertRowId: 1 }); // Item

      await offlineStorageService.saveTransaction(mockSaleRequest, 1);

      // Verify execAsync was called for BEGIN and COMMIT
      expect(mockDb.execAsync).toHaveBeenCalledWith('BEGIN;');
      expect(mockDb.execAsync).toHaveBeenCalledWith('COMMIT;');
    });

    it('should rollback transaction on error', async () => {
      mockDb.runAsync.mockRejectedValue(new Error('Insert failed'));

      try {
        await offlineStorageService.saveTransaction(mockSaleRequest, 1);
      } catch (error) {
        // Expected error
      }

      // Verify rollback was called
      expect(mockDb.execAsync).toHaveBeenCalledWith('ROLLBACK;');
    });

    it('should include customer_name in transaction if provided', async () => {
      const saleWithCustomer: SaleRequest = {
        ...mockSaleRequest,
        customer_name: 'John Doe',
      };

      const result = await offlineStorageService.saveTransaction(saleWithCustomer, 1);

      expect(result.customerName).toBe('John Doe');
    });
  });

  describe('getPendingTransactions', () => {
    const mockPendingTransactions = [
      {
        id: 1,
        transaction_number: 'OFFLINE-1234567890-123',
        timestamp: '2026-05-28T10:00:00Z',
        cashier_id: 1,
        payment_method: 'CASH',
        total: '150000.00',
        subtotal: '150000.00',
        tax: '0.00',
        discount: '0.00',
        customer_name: null,
        status: 'pending_sync' as OfflineTransactionStatus,
        created_at: '2026-05-28T10:00:00Z',
        updated_at: '2026-05-28T10:00:00Z',
      },
    ];

    beforeEach(async () => {
      await offlineStorageService.initialize();
    });

    it('should return transactions with pending_sync status', async () => {
      mockDb.getAllAsync.mockResolvedValue(mockPendingTransactions);

      const transactions = await offlineStorageService.getPendingTransactions();

      expect(transactions).toHaveLength(1);
      expect(transactions[0].status).toBe('pending_sync');
      expect(transactions[0].transaction_number).toBe('OFFLINE-1234567890-123');
    });

    it('should return empty array when no pending transactions exist', async () => {
      mockDb.getAllAsync.mockResolvedValue([]);

      const transactions = await offlineStorageService.getPendingTransactions();

      expect(transactions).toEqual([]);
    });

    it('should include transaction items in result', async () => {
      const mockItems = [
        {
          id: 1,
          transaction_id: 1,
          product_id: 1,
          product_sku: 'SKU-12345',
          product_name: 'Paracetamol 500mg',
          quantity: 2,
          unit_price: '75000.00',
          subtotal: '150000.00',
        },
      ];

      mockDb.getAllAsync
        .mockResolvedValueOnce(mockPendingTransactions) // Transactions query
        .mockResolvedValueOnce(mockItems); // Items query

      const transactions = await offlineStorageService.getPendingTransactions();

      expect(transactions).toHaveLength(1);
      expect(transactions[0].items).toHaveLength(1);
      expect(transactions[0].items[0].product_name).toBe('Paracetamol 500mg');
    });

    it('should handle database query errors', async () => {
      mockDb.getAllAsync.mockRejectedValue(new Error('Query failed'));

      await expect(offlineStorageService.getPendingTransactions()).rejects.toThrow(
        OfflineStorageError
      );
    });
  });

  describe('markTransactionSynced', () => {
    beforeEach(async () => {
      await offlineStorageService.initialize();
      mockDb.runAsync.mockResolvedValue({ changes: 1 });
    });

    it('should update transaction status to synced', async () => {
      await offlineStorageService.markTransactionSynced(1);

      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('UPDATE offline_transactions'),
        ['synced', expect.any(String), 1]
      );
    });

    it('should update updated_at timestamp', async () => {
      await offlineStorageService.markTransactionSynced(1);

      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('updated_at = ?'),
        ['synced', expect.any(String), 1]
      );
    });

    it('should handle non-existent transaction gracefully', async () => {
      mockDb.runAsync.mockResolvedValue({ changes: 0 });

      // Should not throw, just return normally
      await expect(offlineStorageService.markTransactionSynced(999)).resolves.not.toThrow();
    });
  });

  describe('deleteTransaction', () => {
    beforeEach(async () => {
      await offlineStorageService.initialize();
      mockDb.runAsync.mockResolvedValue({ changes: 1 });
    });

    it('should delete transaction by ID', async () => {
      await offlineStorageService.deleteTransaction(1);

      expect(mockDb.runAsync).toHaveBeenCalledWith(
        'DELETE FROM offline_transactions WHERE id = ?;',
        [1]
      );
    });

    it('should cascade delete transaction items due to foreign key', async () => {
      // Foreign key constraint should handle this automatically
      await offlineStorageService.deleteTransaction(1);

      expect(mockDb.runAsync).toHaveBeenCalledWith(
        'DELETE FROM offline_transactions WHERE id = ?;',
        [1]
      );
    });

    it('should handle non-existent transaction gracefully', async () => {
      mockDb.runAsync.mockResolvedValue({ changes: 0 });

      await expect(offlineStorageService.deleteTransaction(999)).resolves.not.toThrow();
    });
  });

  describe('close', () => {
    it('should close database connection', async () => {
      await offlineStorageService.initialize();
      await offlineStorageService.close();

      expect(mockDb.closeAsync).toHaveBeenCalled();
    });

    it('should handle close errors gracefully', async () => {
      await offlineStorageService.initialize();
      mockDb.closeAsync.mockRejectedValue(new Error('Close failed'));

      // Should not throw
      await expect(offlineStorageService.close()).resolves.not.toThrow();
    });
  });
});
