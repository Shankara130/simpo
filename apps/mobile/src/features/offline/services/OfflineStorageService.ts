/**
 * OfflineStorageService - Local SQLite storage for offline transactions
 * Story 8.1: Implement Local SQLite Storage for Offline Transactions
 *
 * Provides offline transaction storage using expo-sqlite with:
 * - Transaction header and line items storage
 * - Pending transaction retrieval
 * - Sync status management
 * - Transaction cleanup after sync
 */

import * as SQLite from 'expo-sqlite';
import { v4 as uuidv4 } from 'uuid';
import {
  OfflineTransactionWithItems,
  OfflineTransaction,
  OfflineTransactionItem,
  OfflineTransactionStatus,
  DATABASE_NAME,
  TABLE_TRANSACTIONS,
  TABLE_TRANSACTION_ITEMS,
  CREATE_TRANSACTIONS_TABLE,
  CREATE_TRANSACTION_ITEMS_TABLE,
  CREATE_STATUS_INDEX,
  CREATE_CREATED_AT_INDEX,
  SCHEMA_VERSION,
  PRAGMA_SCHEMA_VERSION,
  OfflineStorageError,
} from '../types/offline.types';
import { SaleRequest, SaleItem, TransactionResponse } from '../../pos/types/transaction.types';

/**
 * OfflineStorageService - Singleton service for offline transaction storage
 * Follows service class pattern from PrinterManager.ts
 */
class OfflineStorageService {
  private static instance: OfflineStorageService;
  private db: SQLite.SQLiteDatabase | null = null;
  private isInitialized = false;

  private constructor() {
    // Private constructor for singleton pattern
  }

  /**
   * Get singleton instance
   * Follows pattern from PrinterManager.ts and TransactionService.ts
   */
  static getInstance(): OfflineStorageService {
    if (!OfflineStorageService.instance) {
      OfflineStorageService.instance = new OfflineStorageService();
    }
    return OfflineStorageService.instance;
  }

  /**
   * Initialize database and create schema
   * Creates offline_transactions and offline_transaction_items tables with indexes
   * Enables WAL mode for better crash recovery and concurrency
   * Handles schema migrations
   */
  async initialize(): Promise<void> {
    if (this.isInitialized && this.db) {
      return Promise.resolve(); // Already initialized
    }

    try {
      this.db = await SQLite.openDatabaseAsync(DATABASE_NAME);

      // Enable WAL mode for better crash recovery (CRITICAL-005)
      // WAL provides better atomicity and allows reads during writes
      await this.db.execAsync('PRAGMA journal_mode=WAL;');

      // Check and run migrations if needed (CRITICAL-004)
      await this.runMigrations();

      // Create tables and indexes
      await this.db.execAsync(`
        ${CREATE_TRANSACTIONS_TABLE}
        ${CREATE_TRANSACTION_ITEMS_TABLE}
        ${CREATE_STATUS_INDEX}
        ${CREATE_CREATED_AT_INDEX}
      `);

      // Set current schema version
      await this.db.execAsync(`PRAGMA ${PRAGMA_SCHEMA_VERSION}=${SCHEMA_VERSION};`);

      this.isInitialized = true;
    } catch (error) {
      throw new OfflineStorageError('Failed to initialize offline storage database', error);
    }
  }

  /**
   * Run database migrations if schema version is outdated
   * CRITICAL-004: Handle app updates with schema changes
   */
  private async runMigrations(): Promise<void> {
    try {
      const result = await this.db!.getAsync<{ user_version: number }>(
        `PRAGMA ${PRAGMA_SCHEMA_VERSION};`
      );
      const currentVersion = result?.user_version || 0;

      if (currentVersion < SCHEMA_VERSION) {
        console.info(`[OfflineStorage] Migrating database from v${currentVersion} to v${SCHEMA_VERSION}`);

        // Future migrations will be added here
        // Example:
        // if (currentVersion < 2) {
        //   await this.migrateToV2();
        // }

        console.info('[OfflineStorage] Database migration complete');
      }
    } catch (error) {
      // If version check fails, log but continue
      // This allows first-time initialization to succeed
      console.warn('[OfflineStorage] Could not check schema version:', error);
    }
  }

  /**
   * Ensure database is initialized before operations
   */
  private async ensureInitialized(): Promise<void> {
    if (!this.isInitialized || !this.db) {
      await this.initialize();
    }
  }

  /**
   * Generate offline transaction number
   * Format: OFFLINE-{timestamp}-{random}
   */
  private generateTransactionNumber(): string {
    const timestamp = Date.now();
    const random = Math.floor(Math.random() * 10000);
    return `OFFLINE-${timestamp}-${random}`;
  }

  /**
   * Get current timestamp in ISO 8601 format
   */
  private getCurrentTimestamp(): string {
    return new Date().toISOString();
  }

  /**
   * Convert SaleItem to OfflineTransactionItem
   * Maps API sale item to database storage format
   * Note: SKU and name are temporary placeholders enriched during sync (Story 8-2)
   */
  private mapSaleItemToOfflineItem(
    saleItem: SaleItem,
    transactionId: number
  ): Omit<OfflineTransactionItem, 'id'> {
    return {
      transaction_id: transactionId,
      product_id: saleItem.product_id,
      product_sku: `SKU-${saleItem.product_id}`, // Temporary - enriched during sync
      product_name: `Product-${saleItem.product_id}`, // Temporary - enriched during sync
      quantity: saleItem.quantity,
      unit_price: saleItem.unit_price,
      subtotal: this.calculateItemSubtotal(saleItem),
    };
  }

  /**
   * Save transaction to offline storage
   * Stores transaction header and items in SQLite database
   * Returns TransactionResponse in consistent format with online API
   */
  async saveTransaction(
    saleRequest: SaleRequest,
    cashierId: number
  ): Promise<TransactionResponse> {
    await this.ensureInitialized();

    if (!this.db) {
      throw new OfflineStorageError('Database not initialized');
    }

    try {
      // Begin transaction for atomic insert
      await this.db!.execAsync('BEGIN;');

      const transactionNumber = this.generateTransactionNumber();
      const timestamp = this.getCurrentTimestamp();
      const total = this.calculateTotal(saleRequest);

      // Insert transaction header
      const transactionResult = await this.db!.runAsync(
        `INSERT INTO ${TABLE_TRANSACTIONS} (
          transaction_number, timestamp, cashier_id, payment_method,
          total, subtotal, tax, discount, customer_name, status, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`,
        [
          transactionNumber,
          timestamp,
          cashierId,
          saleRequest.payment_method,
          total.total,
          total.subtotal,
          saleRequest.tax_amount,
          saleRequest.discount_amount,
          saleRequest.customer_name || null,
          'pending_sync',
          timestamp,
          timestamp,
        ]
      );

      const transactionId = transactionResult.lastInsertRowId;

      // Insert transaction items
      for (const item of saleRequest.items) {
        const itemSubtotal = this.calculateItemSubtotal(item);
        await this.db!.runAsync(
          `INSERT INTO ${TABLE_TRANSACTION_ITEMS} (
            transaction_id, product_id, product_sku, product_name,
            quantity, unit_price, subtotal
          ) VALUES (?, ?, ?, ?, ?, ?, ?);`,
          [
            transactionId,
            item.product_id,
            `SKU-${item.product_id}`, // Temporary SKU - enriched during sync (Story 8-2)
            `Product-${item.product_id}`, // Temporary name - enriched during sync (Story 8-2)
            item.quantity,
            item.unit_price,
            itemSubtotal,
          ]
        );
      }

      // Commit transaction
      await this.db!.execAsync('COMMIT;');

      // Return in consistent format with online API
      return {
        id: transactionId as number,
        transactionNumber,
        cashierId,
        branchId: 0, // Offline - branch ID not applicable
        total: total.total,
        subtotal: total.subtotal,
        tax: saleRequest.tax_amount,
        discount: saleRequest.discount_amount,
        paymentMethod: saleRequest.payment_method,
        customerName: saleRequest.customer_name,
        status: 'pending_sync',
        created_at: timestamp,
        updated_at: timestamp,
      };
    } catch (error) {
      // Rollback on error
      if (this.db) {
        try {
          await this.db.execAsync('ROLLBACK;');
        } catch (rollbackError) {
          // Log rollback error but still throw original error
          // Rollback failure is non-critical since transaction will be rejected
          console.warn('[OfflineStorage] Failed to rollback transaction:', rollbackError);
        }
      }
      throw new OfflineStorageError('Failed to save offline transaction', error);
    }
  }

  /**
   * Calculate total from sale request items
   * Returns total and subtotal as decimal strings
   */
  private calculateTotal(saleRequest: SaleRequest): {
    total: string;
    subtotal: string;
  } {
    let subtotal = '0.00';

    for (const item of saleRequest.items) {
      const itemSubtotal = this.calculateItemSubtotal(item);
      subtotal = this.addDecimalStrings(subtotal, itemSubtotal);
    }

    // Add tax and subtract discount
    const total = this.addDecimalStrings(
      this.addDecimalStrings(subtotal, saleRequest.tax_amount),
      this.subtractDecimalStrings('0.00', saleRequest.discount_amount)
    );

    return { total, subtotal };
  }

  /**
   * Calculate item subtotal (quantity * unit_price)
   */
  private calculateItemSubtotal(item: SaleItem): string {
    const quantity = BigInt(item.quantity);
    const price = this.decimalStringToBigInt(item.unit_price);
    const subtotal = quantity * price;
    return this.bigIntToDecimalString(subtotal);
  }

  /**
   * Add two decimal strings
   * Simple implementation for MVP - handles 2 decimal places
   */
  private addDecimalStrings(a: string, b: string): string {
    const aValue = Math.round(parseFloat(a || '0') * 100);
    const bValue = Math.round(parseFloat(b || '0') * 100);
    const result = aValue + bValue;
    return (result / 100).toFixed(2);
  }

  /**
   * Subtract b from a (both decimal strings)
   */
  private subtractDecimalStrings(a: string, b: string): string {
    const aValue = Math.round(parseFloat(a || '0') * 100);
    const bValue = Math.round(parseFloat(b || '0') * 100);
    const result = aValue - bValue;
    return (result / 100).toFixed(2);
  }

  /**
   * Convert decimal string to BigInt (for precision math)
   * Assumes 2 decimal places
   */
  private decimalStringToBigInt(value: string): bigint {
    const rounded = Math.round(parseFloat(value || '0') * 100);
    return BigInt(rounded);
  }

  /**
   * Convert BigInt to decimal string (2 decimal places)
   */
  private bigIntToDecimalString(value: bigint): string {
    const num = Number(value);
    return (num / 100).toFixed(2);
  }

  /**
   * Get all pending transactions
   * Returns transactions with status 'pending_sync' including their items
   */
  async getPendingTransactions(): Promise<OfflineTransactionWithItems[]> {
    await this.ensureInitialized();

    if (!this.db) {
      throw new OfflineStorageError('Database not initialized');
    }

    try {
      // Get pending transactions
      const transactions = await this.db.getAllAsync<OfflineTransaction>(
        `SELECT * FROM ${TABLE_TRANSACTIONS} WHERE status = 'pending_sync' ORDER BY created_at ASC;`
      );

      // Get items for each transaction
      const result: OfflineTransactionWithItems[] = [];
      for (const transaction of transactions) {
        const items = await this.db.getAllAsync<OfflineTransactionItem>(
          `SELECT * FROM ${TABLE_TRANSACTION_ITEMS} WHERE transaction_id = ?;`,
          [transaction.id]
        );

        result.push({
          ...transaction,
          items,
        });
      }

      return result;
    } catch (error) {
      throw new OfflineStorageError('Failed to retrieve pending transactions', error);
    }
  }

  /**
   * Mark transaction as synced
   * Updates status to 'synced' and updates timestamp
   */
  async markTransactionSynced(transactionId: number): Promise<void> {
    await this.ensureInitialized();

    if (!this.db) {
      throw new OfflineStorageError('Database not initialized');
    }

    try {
      const timestamp = this.getCurrentTimestamp();

      await this.db.runAsync(
        `UPDATE ${TABLE_TRANSACTIONS} SET status = ?, updated_at = ? WHERE id = ?;`,
        ['synced', timestamp, transactionId]
      );
    } catch (error) {
      throw new OfflineStorageError('Failed to mark transaction as synced', error);
    }
  }

  /**
   * Delete transaction by ID
   * Cascades to transaction items via foreign key constraint
   */
  async deleteTransaction(transactionId: number): Promise<void> {
    await this.ensureInitialized();

    if (!this.db) {
      throw new OfflineStorageError('Database not initialized');
    }

    try {
      await this.db.runAsync(
        `DELETE FROM ${TABLE_TRANSACTIONS} WHERE id = ?;`,
        [transactionId]
      );
    } catch (error) {
      throw new OfflineStorageError('Failed to delete transaction', error);
    }
  }

  /**
   * Mark transaction as synced and delete atomically
   * Performs both operations in a single SQLite transaction to prevent inconsistent state
   * CRITICAL-007 fix: Prevents data loss if app crashes between mark and delete
   */
  async markAndDeleteTransaction(transactionId: number): Promise<void> {
    await this.ensureInitialized();

    if (!this.db) {
      throw new OfflineStorageError('Database not initialized');
    }

    try {
      const timestamp = this.getCurrentTimestamp();

      // Begin transaction for atomic operation
      await this.db.execAsync('BEGIN;');

      // Mark as synced
      await this.db.runAsync(
        `UPDATE ${TABLE_TRANSACTIONS} SET status = ?, updated_at = ? WHERE id = ?;`,
        ['synced', timestamp, transactionId]
      );

      // Delete transaction
      await this.db.runAsync(
        `DELETE FROM ${TABLE_TRANSACTIONS} WHERE id = ?;`,
        [transactionId]
      );

      // Commit transaction
      await this.db.execAsync('COMMIT;');
    } catch (error) {
      // Rollback on error
      try {
        await this.db.execAsync('ROLLBACK;');
      } catch (rollbackError) {
        // Rollback failure is non-critical
        console.warn('[OfflineStorage] Failed to rollback transaction:', rollbackError);
      }
      throw new OfflineStorageError('Failed to mark and delete transaction', error);
    }
  }

  /**
   * Close database connection
   */
  async close(): Promise<void> {
    if (this.db) {
      try {
        await this.db.closeAsync();
      } catch (error) {
        // Ignore close errors
      } finally {
        this.db = null;
        this.isInitialized = false;
      }
    }
  }

  /**
   * Reset database (for testing purposes)
   * Drops all tables and recreates schema
   */
  async reset(): Promise<void> {
    await this.ensureInitialized();

    if (!this.db) {
      throw new OfflineStorageError('Database not initialized');
    }

    try {
      // Drop tables
      await this.db.execAsync(`
        DROP TABLE IF EXISTS ${TABLE_TRANSACTION_ITEMS};
        DROP TABLE IF EXISTS ${TABLE_TRANSACTIONS};
      `);

      // Reset initialization flag
      this.isInitialized = false;

      // Reinitialize
      await this.initialize();
    } catch (error) {
      throw new OfflineStorageError('Failed to reset database', error);
    }
  }
}

// Export singleton instance
export default OfflineStorageService.getInstance();
