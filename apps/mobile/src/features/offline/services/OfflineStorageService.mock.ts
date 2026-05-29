/**
 * Mock for OfflineStorageService (at same directory level)
 * This allows Jest to mock it properly with relative imports
 */

import { OfflineTransactionWithItems } from '../types/offline.types';

export class OfflineStorageService {
  private static instance: OfflineStorageService;

  static getInstance(): OfflineStorageService {
    if (!OfflineStorageService.instance) {
      OfflineStorageService.instance = new OfflineStorageService();
    }
    return OfflineStorageService.instance;
  }

  async getPendingTransactions(): Promise<OfflineTransactionWithItems[]> {
    return [];
  }

  async markTransactionSynced(transactionId: number): Promise<void> {
    // Mock implementation
  }

  async deleteTransaction(transactionId: number): Promise<void> {
    // Mock implementation
  }

  async initialize(): Promise<void> {
    // Mock implementation
  }

  async close(): Promise<void> {
    // Mock implementation
  }

  async saveTransaction(
    saleRequest: any,
    cashierId: number
  ): Promise<any> {
    // Mock implementation
    return {} as any;
  }

  async reset(): Promise<void> {
    // Mock implementation
  }
}

// Export singleton instance
export default new OfflineStorageService();
