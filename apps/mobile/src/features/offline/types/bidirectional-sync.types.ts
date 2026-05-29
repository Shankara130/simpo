/**
 * Bidirectional Sync Types
 * Defines interfaces for bidirectional synchronization orchestration
 * Story 8.3: Implement Bidirectional Data Synchronization
 */

/**
 * Sync Phase - Represents current phase of bidirectional sync
 */
export type SyncPhase =
  | 'idle'
  | 'uploading'
  | 'downloading_stock'
  | 'downloading_products'
  | 'downloading_user'
  | 'synced'
  | 'failed';

/**
 * Bidirectional Sync State
 * Extended sync state with phase information for UI indicators
 */
export interface BidirectionalSyncState {
  status: 'idle' | 'syncing' | 'synced' | 'failed';
  phase: SyncPhase;
  pendingCount: number;          // Number of transactions waiting to sync
  processingCount: number;       // Number of transactions currently processing (0 or 1)
  syncedCount: number;           // Number successfully synced in current session
  failedCount: number;           // Number failed in current session
  currentPhase: string | null;   // Human-readable current phase description
  error?: string;                // Error message if sync failed
  lastSyncTime: string | null;   // ISO 8601 timestamp of last successful sync
}

/**
 * Product Sync Request
 * Query parameters for product sync endpoint
 */
export interface ProductSyncRequest {
  since?: string;                // ISO 8601 timestamp - get changes since this time
}

/**
 * Product Sync Response
 * Response from GET /api/v1/products/sync endpoint
 */
export interface ProductSyncResponse {
  products: Array<{
    id: number;
    sku: string;
    name: string;
    stock_qty: number;
    updated_at: string;           // ISO 8601 timestamp
  }>;
  lastSyncTimestamp: string;     // ISO 8601 timestamp
}

/**
 * User Sync Response
 * Response from GET /api/v1/users/me endpoint
 */
export interface UserSyncResponse {
  id: number;
  username: string;
  email: string;
  role: 'admin' | 'owner' | 'cashier';
  status: 'active' | 'inactive';
  profile?: {
    firstName?: string;
    lastName?: string;
    phone?: string;
  };
  updated_at: string;            // ISO 8601 timestamp
}

/**
 * Bidirectional Sync Result
 * Result of full bidirectional sync operation
 */
export interface BidirectionalSyncResult {
  success: boolean;
  phase: SyncPhase;
  uploaded: number;              // Number of transactions uploaded
  failedCount?: number;          // Number of transactions that failed to upload
  downloadedStock: number;        // Number of stock records updated
  downloadedProducts: number;     // Number of new/updated products
  userUpdated: boolean;          // Whether user data was updated
  duration: number;               // Total sync duration in milliseconds
  error?: string;                // Error message if failed
}

/**
 * Sync Orchestrator Error
 * Custom error class for sync orchestrator operations
 */
export class SyncOrchestratorError extends Error {
  constructor(
    message: string,
    public readonly phase?: SyncPhase,
    public originalError?: any
  ) {
    super(message);
    this.name = 'SyncOrchestratorError';
  }
}

/**
 * Product Sync Error
 * Custom error class for product sync operations
 */
export class ProductSyncError extends Error {
  constructor(
    message: string,
    public originalError?: any,
    public readonly isRetryable: boolean = true
  ) {
    super(message);
    this.name = 'ProductSyncError';
  }
}

/**
 * User Sync Error
 * Custom error class for user sync operations
 */
export class UserSyncError extends Error {
  constructor(
    message: string,
    public originalError?: any,
    public readonly isRetryable: boolean = true
  ) {
    super(message);
    this.name = 'UserSyncError';
  }
}

/**
 * AsyncStorage Keys for Bidirectional Sync
 */
export const LAST_PRODUCT_SYNC_KEY = '@simpo_last_product_sync';
export const USER_PROFILE_KEY = '@simpo_user_profile';
export const BIDIRECTIONAL_SYNC_STATE_KEY = '@simpo_bidirectional_sync_state';
export const BIDIRECTIONAL_SYNC_RETRY_KEY = '@simpo_bidirectional_sync_retry';

/**
 * Retry Configuration
 * AC8: Retry intervals follow Story 8.2 pattern: 1min, 2min, 4min, 8min, 32min
 */
export const MAX_RETRY_ATTEMPTS = 5;

export const BASE_RETRY_DELAY_MS = 60 * 1000; // 1 minute in milliseconds

/**
 * Conflict Error Response (RFC 7807 Problem Details)
 * AC6: Backend response format for conflict resolution errors
 */
export interface ConflictErrorResponse {
  type: string;                   // URI reference to error type
  title: string;                  // Short title
  status: number;                 // HTTP status code
  detail: string;                 // Detailed error message
  instance: string;               // URI reference to specific occurrence
  transaction_id?: string;        // Transaction number
  available_stock?: number;       // Available stock quantity
  requested_quantity?: number;    // Requested quantity
  product_id?: number;            // Product ID
  product_sku?: string;           // Product SKU
}

/**
 * Failed Transaction Info
 * AC6: Information about failed transaction due to conflict
 */
export interface FailedTransactionInfo {
  transactionId: number;         // Local SQLite transaction ID
  transactionNumber: string;      // Transaction number
  conflictError: ConflictErrorResponse; // Conflict error details
  timestamp: string;              // ISO 8601 timestamp of failure
  canOverride: boolean;           // Whether manual override is allowed
  requiresAdminAuth: boolean;     // Whether admin authorization is required
}

/**
 * Manual Override Request
 * AC6: Request for manual override of failed transaction
 */
export interface ManualOverrideRequest {
  transactionId: number;          // Local SQLite transaction ID
  adminUserId: number;            // Admin user ID authorizing override
  reason: string;                 // Reason for override
  forceProcessing: boolean;       // Force processing despite insufficient stock
}

/**
 * Manual Override Result
 * AC6: Result of manual override operation
 */
export interface ManualOverrideResult {
  success: boolean;
  transactionId: number;
  transactionNumber?: string;
  message: string;
  serverTransactionId?: number;   // Server transaction ID if successful
}

/**
 * Transaction Status with Error Message
 * Extended status for transactions with conflict errors
 */
export type TransactionStatusWithError = 'pending_sync' | 'synced' | 'failed' | 'failed_conflict';
