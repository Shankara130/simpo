/**
 * Audit Log Service
 * Handles audit trail logging for compliance and accountability
 * Story 7.4: Cash drawer audit logging for Badan POM compliance
 */

import AsyncStorage from '@react-native-async-storage/async-storage';

/**
 * Audit Event Types
 */
export enum AuditEventType {
  DRAWER_OPEN = 'cash_drawer_open', // Aligned with spec terminology
  DRAWER_FAILED = 'drawer_failed',
  PRINT = 'print',
  TRANSACTION = 'transaction',
}

/**
 * Audit Event Interface
 */
export interface AuditEvent {
  id: string;
  eventType: AuditEventType;
  transactionId?: string;
  userId?: string;
  status: 'success' | 'failed';
  timestamp: string;  // ISO 8601 format
  metadata?: Record<string, unknown>;
}

/**
 * Audit Log Storage Key
 */
const AUDIT_LOG_KEY = '@simpo_audit_logs';
const AUDIT_LOG_QUEUE_KEY = '@simpo_audit_log_queue';

/**
 * Maximum number of audit logs to keep locally
 */
const MAX_LOCAL_LOGS = 1000;

/**
 * Audit Log Service
 * Manages audit trail logging with offline queue support
 */
export class AuditLogService {
  private static instance: AuditLogService | null = null;
  private apiBaseUrl: string;

  private constructor(apiBaseUrl: string = '') {
    this.apiBaseUrl = apiBaseUrl;
  }

  /**
   * Get singleton instance
   */
  public static getInstance(apiBaseUrl?: string): AuditLogService {
    if (!AuditLogService.instance) {
      AuditLogService.instance = new AuditLogService(apiBaseUrl);
    }
    return AuditLogService.instance;
  }

  /**
   * Log cash drawer open event
   * @param transactionId - Transaction ID
   * @param userId - User ID who triggered the event
   * @param status - Success or failed status
   * @param metadata - Additional event metadata
   */
  public async logCashDrawerOpen(
    transactionId: string,
    userId?: string,
    status: 'success' | 'failed' = 'success',
    metadata?: Record<string, unknown>
  ): Promise<void> {
    const event: AuditEvent = {
      id: this.generateEventId(),
      eventType: status === 'success' ? AuditEventType.DRAWER_OPEN : AuditEventType.DRAWER_FAILED,
      transactionId,
      userId,
      status,
      timestamp: new Date().toISOString(),
      metadata,
    };

    await this.saveEventLocally(event);

    // Try to sync with backend
    await this.syncEvent(event);
  }

  /**
   * Save event to local storage
   */
  private async saveEventLocally(event: AuditEvent): Promise<void> {
    try {
      const existingLogs = await this.getLocalLogs();
      const updatedLogs = [event, ...existingLogs].slice(0, MAX_LOCAL_LOGS);

      await AsyncStorage.setItem(AUDIT_LOG_KEY, JSON.stringify(updatedLogs));
    } catch (error) {
      console.error('[AuditLogService] Failed to save event locally:', error);
    }
  }

  /**
   * Get local audit logs
   */
  public async getLocalLogs(): Promise<AuditEvent[]> {
    try {
      const logsJson = await AsyncStorage.getItem(AUDIT_LOG_KEY);
      if (logsJson) {
        return JSON.parse(logsJson) as AuditEvent[];
      }
      return [];
    } catch (error) {
      console.error('[AuditLogService] Failed to load local logs:', error);
      return [];
    }
  }

  /**
   * Sync event to backend API with retry logic
   */
  private async syncEvent(event: AuditEvent): Promise<void> {
    if (!this.apiBaseUrl) {
      // No API URL configured, queue for later
      await this.queueEvent(event);
      return;
    }

    const maxRetries = 3; // Maximum retry attempts
    const baseDelay = 1000; // Base delay in milliseconds (1 second)
    let lastError: Error | null = null;

    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        const response = await fetch(`${this.apiBaseUrl}/api/v1/audit`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(event),
        });

        if (!response.ok) {
          throw new Error(`API request failed: ${response.status}`);
        }

        // Success - return without logging to avoid test timing issues
        return;
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));

        // If this is not the last attempt, wait with exponential backoff before retrying
        if (attempt < maxRetries) {
          const delay = baseDelay * Math.pow(2, attempt); // Exponential backoff: 1s, 2s, 4s
          console.warn(
            `[AuditLogService] Sync attempt ${attempt + 1}/${maxRetries + 1} failed, ` +
            `retrying in ${delay}ms:`,
            lastError.message
          );
          await this.delay(delay);
        }
      }
    }

    // All retries exhausted - queue for later sync
    console.error('[AuditLogService] All retry attempts failed, queuing for later sync:', lastError?.message);
    await this.queueEvent(event);
  }

  /**
   * Delay helper for retry logic
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Queue event for later sync
   */
  private async queueEvent(event: AuditEvent): Promise<void> {
    try {
      const queue = await this.getQueue();
      const updatedQueue = [...queue, event];

      await AsyncStorage.setItem(AUDIT_LOG_QUEUE_KEY, JSON.stringify(updatedQueue));
    } catch (error) {
      console.error('[AuditLogService] Failed to queue event:', error);
    }
  }

  /**
   * Get queued events
   */
  private async getQueue(): Promise<AuditEvent[]> {
    try {
      const queueJson = await AsyncStorage.getItem(AUDIT_LOG_QUEUE_KEY);
      if (queueJson) {
        return JSON.parse(queueJson) as AuditEvent[];
      }
      return [];
    } catch (error) {
      console.error('[AuditLogService] Failed to load queue:', error);
      return [];
    }
  }

  /**
   * Sync queued events to backend
   */
  public async syncQueuedEvents(): Promise<number> {
    const queue = await this.getQueue();

    if (queue.length === 0) {
      return 0;
    }

    let syncedCount = 0;

    for (const event of queue) {
      try {
        await this.syncEvent(event);
        syncedCount++;

        // Remove from queue after successful sync
        await this.removeFromQueue(event.id);
      } catch (error) {
        console.error(`[AuditLogService] Failed to sync event ${event.id}:`, error);
        // Stop syncing on error to avoid infinite loops
        break;
      }
    }

    return syncedCount;
  }

  /**
   * Remove event from queue
   */
  private async removeFromQueue(eventId: string): Promise<void> {
    try {
      const queue = await this.getQueue();
      const updatedQueue = queue.filter((event) => event.id !== eventId);

      await AsyncStorage.setItem(AUDIT_LOG_QUEUE_KEY, JSON.stringify(updatedQueue));
    } catch (error) {
      console.error('[AuditLogService] Failed to remove from queue:', error);
    }
  }

  /**
   * Clear local audit logs
   */
  public async clearLocalLogs(): Promise<void> {
    try {
      await AsyncStorage.removeItem(AUDIT_LOG_KEY);
    } catch (error) {
      console.error('[AuditLogService] Failed to clear local logs:', error);
    }
  }

  /**
   * Clear queued events
   */
  public async clearQueue(): Promise<void> {
    try {
      await AsyncStorage.removeItem(AUDIT_LOG_QUEUE_KEY);
    } catch (error) {
      console.error('[AuditLogService] Failed to clear queue:', error);
    }
  }

  /**
   * Generate unique event ID
   */
  private generateEventId(): string {
    return `audit_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Reset singleton instance (for testing)
   */
  public static resetInstance(): void {
    AuditLogService.instance = null;
  }
}

// Export singleton getter for convenience
export function getAuditLogService(apiBaseUrl?: string): AuditLogService {
  return AuditLogService.getInstance(apiBaseUrl);
}

// Export service class
export default AuditLogService;
