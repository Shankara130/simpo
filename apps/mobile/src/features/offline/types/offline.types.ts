/**
 * Offline Transaction Types
 * Defines interfaces for offline transaction storage using SQLite
 * Story 8.1: Implement Local SQLite Storage for Offline Transactions
 */

import { SaleRequest, SaleItem } from '../../pos/types/transaction.types';

/**
 * Offline Transaction Status
 * Tracks synchronization state of offline transactions
 */
export type OfflineTransactionStatus = 'pending_sync' | 'synced' | 'failed';

/**
 * Offline Transaction - Database model for local storage
 * Maps to offline_transactions table in SQLite
 */
export interface OfflineTransaction {
  id: number;                      // Local SQLite ID (auto-increment)
  transaction_number: string;       // Format: OFFLINE-{timestamp}-{random}
  timestamp: string;                // ISO 8601 format
  cashier_id: number;               // Cashier user ID
  payment_method: string;           // Payment method (CASH, TRANSFER, E_WALLET)
  total: string;                    // Decimal as string for precision
  subtotal: string;                 // Decimal as string
  tax: string;                      // Decimal as string
  discount: string;                 // Decimal as string
  customer_name?: string;           // Optional customer name
  status: OfflineTransactionStatus; // Sync status
  created_at: string;               // ISO 8601 creation timestamp
  updated_at: string;               // ISO 8601 update timestamp
}

/**
 * Offline Transaction Item - Line item for offline transactions
 * Maps to offline_transaction_items table in SQLite
 */
export interface OfflineTransactionItem {
  id: number;                       // Local SQLite ID (auto-increment)
  transaction_id: number;           // Foreign key to offline_transactions.id
  product_id: number;               // Product ID
  product_sku: string;              // Product SKU/barcode
  product_name: string;             // Product display name
  quantity: number;                 // Quantity sold
  unit_price: string;               // Decimal as string for precision
  subtotal: string;                 // Decimal as string (quantity * unit_price)
}

/**
 * Complete Offline Transaction with Items
 * Combined data for offline transaction with line items
 */
export interface OfflineTransactionWithItems extends OfflineTransaction {
  items: OfflineTransactionItem[];
}

/**
 * Database Constants
 * SQLite database and table names for offline storage
 */
export const DATABASE_NAME = 'simpo_offline.db';

export const TABLE_TRANSACTIONS = 'offline_transactions';

export const TABLE_TRANSACTION_ITEMS = 'offline_transaction_items';

/**
 * Database Schema Version
 * Increment when schema changes require migration
 */
export const SCHEMA_VERSION = 1;

export const PRAGMA_SCHEMA_VERSION = 'user_version';

/**
 * SQL Schema Creation Statements
 */
export const CREATE_TRANSACTIONS_TABLE = `
  CREATE TABLE IF NOT EXISTS ${TABLE_TRANSACTIONS} (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    transaction_number TEXT UNIQUE NOT NULL,
    timestamp TEXT NOT NULL,
    cashier_id INTEGER NOT NULL,
    payment_method TEXT NOT NULL,
    total TEXT NOT NULL,
    subtotal TEXT NOT NULL,
    tax TEXT NOT NULL,
    discount TEXT NOT NULL,
    customer_name TEXT,
    status TEXT NOT NULL DEFAULT 'pending_sync',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
  );
`;

export const CREATE_TRANSACTION_ITEMS_TABLE = `
  CREATE TABLE IF NOT EXISTS ${TABLE_TRANSACTION_ITEMS} (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    transaction_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    product_sku TEXT NOT NULL,
    product_name TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price TEXT NOT NULL,
    subtotal TEXT NOT NULL,
    FOREIGN KEY (transaction_id) REFERENCES ${TABLE_TRANSACTIONS}(id) ON DELETE CASCADE
  );
`;

export const CREATE_STATUS_INDEX = `
  CREATE INDEX IF NOT EXISTS idx_status ON ${TABLE_TRANSACTIONS}(status);
`;

export const CREATE_CREATED_AT_INDEX = `
  CREATE INDEX IF NOT EXISTS idx_created_at ON ${TABLE_TRANSACTIONS}(created_at);
`;

/**
 * Offline Storage Error
 * Custom error class for offline storage operations
 */
export class OfflineStorageError extends Error {
  constructor(
    message: string,
    public originalError?: any
  ) {
    super(message);
    this.name = 'OfflineStorageError';
  }
}

/**
 * Cache Keys for AsyncStorage
 */
export const CACHE_LAST_STOCK_SYNC = '@simpo_last_stock_sync';

export const CACHE_STOCK_DATA = '@simpo_stock_cache';

/**
 * Cached Stock Data Interface
 */
export interface CachedStockData {
  products: Array<{
    id: number;
    sku: string;
    name: string;
    stock_qty: number;
  }>;
  timestamp: string;  // ISO 8601 timestamp
}
