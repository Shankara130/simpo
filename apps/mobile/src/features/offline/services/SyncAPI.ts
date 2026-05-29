/**
 * SyncAPI - Backend sync endpoint communication
 * Story 8.2: Implement Transaction Sync Queue
 *
 * Provides client for POST /api/v1/sync endpoint
 * Mock implementation for testing until backend endpoint is ready
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import {
  SyncRequest,
  SyncResponse,
  SyncErrorResponse,
  SyncQueueError,
  SyncErrorType,
} from '../types/sync.types';
import { OfflineTransactionWithItems } from '../types/offline.types';

/**
 * Convert OfflineTransactionWithItems to SyncRequest
 * Maps local storage format to backend API format
 */
function mapToSyncRequest(
  transaction: OfflineTransactionWithItems
): SyncRequest {
  return {
    transaction_number: transaction.transaction_number,
    timestamp: transaction.timestamp,
    cashier_id: transaction.cashier_id,
    payment_method: transaction.payment_method,
    total: transaction.total,
    subtotal: transaction.subtotal,
    tax: transaction.tax,
    discount: transaction.discount,
    customer_name: transaction.customer_name,
    items: transaction.items.map((item) => ({
      product_id: item.product_id,
      product_sku: item.product_sku,
      product_name: item.product_name,
      quantity: item.quantity,
      unit_price: item.unit_price,
      subtotal: item.subtotal,
    })),
  };
}

/**
 * SyncAPI - Singleton service for backend sync communication
 */
class SyncAPI {
  private static instance: SyncAPI;
  private baseUrl: string;
  private mockMode: boolean = false;
  private mockDelay: number = 500; // Simulate network latency

  private constructor() {
    // Base URL from environment or default to localhost
    // TODO: Configure from app config
    this.baseUrl = __DEV__
      ? 'http://localhost:8080/api/v1'
      : 'https://api.simpo.pharmacy/api/v1';

    // Mock mode: enabled in development (backend endpoint not ready), disabled in production
    // Use setMockMode() to override at runtime for testing
    this.mockMode = __DEV__;
  }

  /**
   * Get singleton instance
   */
  static getInstance(): SyncAPI {
    if (!SyncAPI.instance) {
      SyncAPI.instance = new SyncAPI();
    }
    return SyncAPI.instance;
  }

  /**
   * Set mock mode for testing
   */
  setMockMode(enabled: boolean, delay: number = 500): void {
    this.mockMode = enabled;
    this.mockDelay = delay;
  }

  /**
   * Get auth token from AsyncStorage
   * Reuses JWT token from authentication
   */
  private async getAuthToken(): Promise<string | null> {
    try {
      const token = await AsyncStorage.getItem('@simpo_auth_token');
      if (!token) {
        console.warn('[SyncAPI] No auth token found in AsyncStorage');
        return null;
      }
      return token;
    } catch (error) {
      console.error('[SyncAPI] Failed to get auth token:', error);
      return null;
    }
  }

  /**
   * Post transaction to sync endpoint
   * Handles RFC 7807 error responses
   */
  async postTransaction(
    transaction: OfflineTransactionWithItems
  ): Promise<SyncResponse> {
    const payload = mapToSyncRequest(transaction);

    if (this.mockMode) {
      return this.mockPostTransaction(payload);
    }

    // Real implementation when backend is ready
    return this.realPostTransaction(payload);
  }

  /**
   * Mock implementation for testing
   * Simulates backend responses with configurable delays
   */
  private async mockPostTransaction(
    payload: SyncRequest
  ): Promise<SyncResponse> {
    // Simulate network latency
    await new Promise((resolve) =>
      setTimeout(resolve, this.mockDelay + Math.random() * 200)
    );

    // Check for mock error scenarios via AsyncStorage
    const mockErrorType = await AsyncStorage.getItem(
      `@simpo_mock_sync_error_${payload.transaction_number}`
    );

    if (mockErrorType) {
      return this.throwMockError(mockErrorType, payload);
    }

    // Success response
    return {
      status: 'synced',
      transaction_id: Date.now(), // Mock server ID
      server_timestamp: new Date().toISOString(),
    };
  }

  /**
   * Throw mock error based on type
   */
  private async throwMockError(
    errorType: string,
    payload: SyncRequest
  ): Promise<SyncResponse> {
    // Clear mock error after use
    await AsyncStorage.removeItem(
      `@simpo_mock_sync_error_${payload.transaction_number}`
    );

    switch (errorType) {
      case '409':
        throw new SyncQueueError(
          'Transaction already exists',
          SyncErrorType.CONFLICT,
          { status: 409, title: 'Conflict', detail: 'Duplicate transaction number' },
          false // Not retryable
        );

      case '400':
        throw new SyncQueueError(
          'Invalid transaction data',
          SyncErrorType.VALIDATION_ERROR,
          { status: 400, title: 'Bad Request', detail: 'Validation failed' },
          false // Not retryable
        );

      case '503':
        throw new SyncQueueError(
          'Service unavailable',
          SyncErrorType.SERVER_ERROR,
          { status: 503, title: 'Service Unavailable', detail: 'Server overloaded' },
          true // Retryable
        );

      case 'network':
        throw new SyncQueueError(
          'Network error',
          SyncErrorType.NETWORK,
          new Error('Network request failed'),
          true // Retryable
        );

      default:
        throw new SyncQueueError(
          'Unknown sync error',
          SyncErrorType.UNKNOWN,
          null,
          true // Retryable
        );
    }
  }

  /**
   * Real implementation when backend endpoint is ready
   * TODO: Implement when POST /api/v1/sync is available
   */
  private async realPostTransaction(
    payload: SyncRequest
  ): Promise<SyncResponse> {
    const token = await this.getAuthToken();

    if (!token) {
      throw new SyncQueueError(
        'No authentication token available',
        SyncErrorType.VALIDATION_ERROR,
        null,
        false
      );
    }

    // Create abort controller for timeout
    const abortController = new AbortController();
    const timeoutId = setTimeout(() => abortController.abort(), 30000); // 30 second timeout

    try {
      const response = await fetch(`${this.baseUrl}/sync`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
        signal: abortController.signal,
      });

      if (response.status === 409) {
        // Conflict - transaction already exists
        const error: SyncErrorResponse = await response.json();
        throw new SyncQueueError(
          error.detail || 'Transaction already exists',
          SyncErrorType.CONFLICT,
          error,
          false // Not retryable - skip
        );
      }

      if (response.status === 400) {
        // Bad request - validation error
        const error: SyncErrorResponse = await response.json();
        throw new SyncQueueError(
          error.detail || 'Invalid transaction data',
          SyncErrorType.VALIDATION_ERROR,
          error,
          false // Not retryable
        );
      }

      if (response.status === 503) {
        // Service unavailable - retryable
        const error: SyncErrorResponse = await response.json();
        throw new SyncQueueError(
          error.detail || 'Service unavailable',
          SyncErrorType.SERVER_ERROR,
          error,
          true // Retryable
        );
      }

      if (!response.ok) {
        // Other errors
        throw new SyncQueueError(
          `Sync failed with status ${response.status}`,
          SyncErrorType.UNKNOWN,
          null,
          true // Retryable
        );
      }

      // Success
      return await response.json();
    } catch (error) {
      // Clear timeout
      clearTimeout(timeoutId);

      if (error instanceof SyncQueueError) {
        throw error;
      }

      // Check if error is due to timeout
      if (error instanceof Error && error.name === 'AbortError') {
        throw new SyncQueueError(
          'Request timeout after 30 seconds',
          SyncErrorType.NETWORK,
          error,
          true // Retryable
        );
      }

      // Network error
      throw new SyncQueueError(
        'Network request failed',
        SyncErrorType.NETWORK,
        error,
        true // Retryable
      );
    } finally {
      // Always clear timeout
      clearTimeout(timeoutId);
    }
  }
}

// Export class and singleton instance
export { SyncAPI };
export default SyncAPI.getInstance();
