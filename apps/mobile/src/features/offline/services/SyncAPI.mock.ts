/**
 * Mock for SyncAPI (at same directory level)
 * This allows Jest to mock it properly with relative imports
 */

import { OfflineTransactionWithItems } from '../types/offline.types';
import { SyncResponse, SyncQueueError, SyncErrorType } from '../types/sync.types';

export class SyncAPI {
  private static instance: SyncAPI;

  static getInstance(): SyncAPI {
    if (!SyncAPI.instance) {
      SyncAPI.instance = new SyncAPI();
    }
    return SyncAPI.instance;
  }

  setMockMode(enabled: boolean, delay: number = 500): void {
    // Mock implementation
  }

  async postTransaction(
    transaction: OfflineTransactionWithItems
  ): Promise<SyncResponse> {
    // Mock implementation
    return {
      status: 'synced',
      transaction_id: Date.now(),
      server_timestamp: new Date().toISOString(),
    };
  }
}

// Export singleton instance
export default new SyncAPI();
