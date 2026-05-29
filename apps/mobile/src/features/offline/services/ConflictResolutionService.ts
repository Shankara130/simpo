/**
 * ConflictResolutionService - Handle sync conflicts and manual overrides
 * Story 8.3: Implement Bidirectional Data Synchronization
 *
 * Provides conflict resolution for insufficient stock scenarios:
 * - Parse backend conflict error responses (RFC 7807)
 * - Mark failed transactions with error messages
 * - Handle manual override requests (with admin authorization)
 * - Log all resolution attempts to audit trail
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import OfflineStorageService from './OfflineStorageService';
import { SyncAPI } from './SyncAPI';
import {
  ConflictErrorResponse,
  FailedTransactionInfo,
  ManualOverrideRequest,
  ManualOverrideResult,
  TransactionStatusWithError,
} from '../types/bidirectional-sync.types';

/**
 * ConflictResolutionService - Singleton service for conflict resolution
 * Follows service class pattern from other services
 */
class ConflictResolutionServiceClass {
  private static instance: ConflictResolutionServiceClass;
  private failedTransactionsCache: Map<number, FailedTransactionInfo> = new Map();
  private offlineStorage: ReturnType<typeof OfflineStorageService.getInstance>;
  private syncAPI: ReturnType<typeof SyncAPI.getInstance>;

  private constructor() {
    this.offlineStorage = OfflineStorageService.getInstance();
    this.syncAPI = SyncAPI.getInstance();
  }

  /**
   * Get singleton instance
   */
  static getInstance(): ConflictResolutionServiceClass {
    if (!ConflictResolutionServiceClass.instance) {
      ConflictResolutionServiceClass.instance = new ConflictResolutionServiceClass();
    }
    return ConflictResolutionServiceClass.instance;
  }

  /**
   * Parse backend error response for conflict information
   * AC6: Parse backend error responses for "insufficient stock" errors
   */
  parseConflictError(errorResponse: any): ConflictErrorResponse | null {
    // Check if error response follows RFC 7807 format
    if (errorResponse && typeof errorResponse === 'object') {
      // Check for conflict type
      const type = errorResponse.type;
      if (type && typeof type === 'string' && type.includes('conflict')) {
        return {
          type: errorResponse.type || '',
          title: errorResponse.title || 'Conflict Error',
          status: errorResponse.status || 409,
          detail: errorResponse.detail || 'Conflict occurred during sync',
          instance: errorResponse.instance || '',
          transaction_id: errorResponse.transaction_id,
          available_stock: errorResponse.available_stock,
          requested_quantity: errorResponse.requested_quantity,
          product_id: errorResponse.product_id,
          product_sku: errorResponse.product_sku,
        };
      }
    }

    // Try to parse from standard HTTP error
    if (errorResponse && errorResponse.status === 409) {
      return {
        type: 'https://api.simpo.com/errors/conflict',
        title: 'Conflict',
        status: 409,
        detail: errorResponse.detail || 'Conflict occurred',
        instance: errorResponse.instance || '/api/v1/sync',
      };
    }

    return null;
  }

  /**
   * Check if error is a conflict error
   */
  isConflictError(error: any): boolean {
    const conflictError = this.parseConflictError(error);
    return conflictError !== null;
  }

  /**
   * Mark transaction as failed with conflict error
   * AC6: Mark failed transactions in SQLite with error messages
   */
  async markTransactionFailed(
    transactionId: number,
    conflictError: ConflictErrorResponse
  ): Promise<void> {
    try {
      // Store failed transaction info in cache
      const failedInfo: FailedTransactionInfo = {
        transactionId,
        transactionNumber: conflictError.transaction_id || `TX-${transactionId}`,
        conflictError,
        timestamp: new Date().toISOString(),
        canOverride: true,
        requiresAdminAuth: true,
      };

      this.failedTransactionsCache.set(transactionId, failedInfo);

      // Persist to AsyncStorage for manual review
      await this.persistFailedTransaction(failedInfo);

      // Log to audit trail
      await this.logToAuditTrail({
        action: 'transaction_failed_conflict',
        transactionId,
        transactionNumber: failedInfo.transactionNumber,
        conflictType: conflictError.type,
        errorMessage: conflictError.detail,
        timestamp: failedInfo.timestamp,
      });
    } catch (error) {
      console.error('[ConflictResolution] Failed to mark transaction:', error);
      throw error;
    }
  }

  /**
   * Persist failed transaction to AsyncStorage
   */
  private async persistFailedTransaction(info: FailedTransactionInfo): Promise<void> {
    try {
      const key = `@simpo_failed_tx_${info.transactionId}`;
      await AsyncStorage.setItem(key, JSON.stringify(info));
    } catch (error) {
      console.error('[ConflictResolution] Failed to persist failed transaction:', error);
    }
  }

  /**
   * Get failed transaction info by ID
   */
  async getFailedTransaction(transactionId: number): Promise<FailedTransactionInfo | null> {
    // Check cache first
    if (this.failedTransactionsCache.has(transactionId)) {
      return this.failedTransactionsCache.get(transactionId)!;
    }

    // Load from AsyncStorage
    try {
      const key = `@simpo_failed_tx_${transactionId}`;
      const data = await AsyncStorage.getItem(key);
      if (data) {
        const info = JSON.parse(data) as FailedTransactionInfo;
        this.failedTransactionsCache.set(transactionId, info);
        return info;
      }
    } catch (error) {
      console.error('[ConflictResolution] Failed to load failed transaction:', error);
    }

    return null;
  }

  /**
   * Get all failed transactions pending review
   */
  async getAllFailedTransactions(): Promise<FailedTransactionInfo[]> {
    const keys = await AsyncStorage.getAllKeys();
    const failedKeys = keys.filter(key => key.startsWith('@simpo_failed_tx_'));

    const failedTransactions: FailedTransactionInfo[] = [];

    for (const key of failedKeys) {
      try {
        const data = await AsyncStorage.getItem(key);
        if (data) {
          const info = JSON.parse(data) as FailedTransactionInfo;
          failedTransactions.push(info);
        }
      } catch (error) {
        console.error('[ConflictResolution] Failed to parse failed transaction:', error);
      }
    }

    // Sort by timestamp (oldest first)
    return failedTransactions.sort((a, b) =>
      new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    );
  }

  /**
   * Clear failed transaction record
   */
  async clearFailedTransaction(transactionId: number): Promise<void> {
    try {
      const key = `@simpo_failed_tx_${transactionId}`;
      await AsyncStorage.removeItem(key);
      this.failedTransactionsCache.delete(transactionId);
    } catch (error) {
      console.error('[ConflictResolution] Failed to clear failed transaction:', error);
    }
  }

  /**
   * Request manual override for failed transaction
   * AC6: Implement manual override capability (with admin authorization flag)
   */
  async requestManualOverride(request: ManualOverrideRequest): Promise<ManualOverrideResult> {
    try {
      // Get failed transaction info
      const failedInfo = await this.getFailedTransaction(request.transactionId);
      if (!failedInfo) {
        return {
          success: false,
          transactionId: request.transactionId,
          message: 'Failed transaction not found',
        };
      }

      // Verify admin authorization (adminUserId=0 is valid)
      if (failedInfo.requiresAdminAuth && (request.adminUserId === undefined || request.adminUserId === null)) {
        return {
          success: false,
          transactionId: request.transactionId,
          message: 'Admin authorization required for this override',
        };
      }

      // Log override attempt to audit trail
      await this.logToAuditTrail({
        action: 'manual_override_requested',
        transactionId: request.transactionId,
        transactionNumber: failedInfo.transactionNumber,
        adminUserId: request.adminUserId,
        reason: request.reason,
        forceProcessing: request.forceProcessing,
        timestamp: new Date().toISOString(),
      });

      // Call backend with override request
      // Note: This would require a new backend endpoint or parameter
      // For now, we'll simulate the call
      const result = await this.performOverrideRequest(request, failedInfo);

      if (result.success) {
        // Clear failed transaction record on success
        await this.clearFailedTransaction(request.transactionId);

        // Log successful override
        await this.logToAuditTrail({
          action: 'manual_override_success',
          transactionId: request.transactionId,
          transactionNumber: result.transactionNumber,
          adminUserId: request.adminUserId,
          serverTransactionId: result.serverTransactionId,
          timestamp: new Date().toISOString(),
        });
      }

      return result;
    } catch (error) {
      console.error('[ConflictResolution] Manual override failed:', error);

      // Log failure
      await this.logToAuditTrail({
        action: 'manual_override_failed',
        transactionId: request.transactionId,
        adminUserId: request.adminUserId,
        error: error instanceof Error ? error.message : 'Unknown error',
        timestamp: new Date().toISOString(),
      });

      return {
        success: false,
        transactionId: request.transactionId,
        message: error instanceof Error ? error.message : 'Override failed',
      };
    }
  }

  /**
   * Perform override request to backend
   * Note: This would need to be implemented with actual backend endpoint
   */
  private async performOverrideRequest(
    request: ManualOverrideRequest,
    failedInfo: FailedTransactionInfo
  ): Promise<ManualOverrideResult> {
    // TODO: Implement actual backend call
    // For now, simulate success
    return {
      success: true,
      transactionId: request.transactionId,
      transactionNumber: failedInfo.transactionNumber,
      message: 'Override successful (simulated)',
      serverTransactionId: Date.now(), // Mock server ID
    };
  }

  /**
   * Log to audit trail
   * AC6: Log all conflict resolution attempts to audit trail
   */
  private async logToAuditTrail(event: Record<string, any>): Promise<void> {
    try {
      const auditKey = '@simpo_sync_audit_log';
      const existingLog = await AsyncStorage.getItem(auditKey);
      const auditLog = existingLog ? JSON.parse(existingLog) : [];

      auditLog.push({
        ...event,
        loggedAt: new Date().toISOString(),
      });

      // Keep only last 1000 entries to prevent storage bloat
      if (auditLog.length > 1000) {
        auditLog.splice(0, Math.max(0, auditLog.length - 1000));
      }

      await AsyncStorage.setItem(auditKey, JSON.stringify(auditLog));
    } catch (error) {
      console.error('[ConflictResolution] Failed to log to audit trail:', error);
    }
  }

  /**
   * Get audit log entries
   */
  async getAuditLog(limit?: number): Promise<Record<string, any>[]> {
    try {
      const auditKey = '@simpo_sync_audit_log';
      const existingLog = await AsyncStorage.getItem(auditKey);
      if (!existingLog) {
        return [];
      }

      const auditLog = JSON.parse(existingLog) as Record<string, any>[];

      // Return in reverse chronological order
      const reversedLog = [...auditLog].reverse();

      if (limit) {
        return reversedLog.slice(0, limit);
      }

      return reversedLog;
    } catch (error) {
      console.error('[ConflictResolution] Failed to get audit log:', error);
      return [];
    }
  }
}

// Export as ConflictResolutionService for clarity
export { ConflictResolutionServiceClass as ConflictResolutionService };
export default ConflictResolutionServiceClass.getInstance();
