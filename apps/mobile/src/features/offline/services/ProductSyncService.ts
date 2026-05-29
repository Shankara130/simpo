/**
 * ProductSyncService - Product data sync from server
 * Story 8.3: Implement Bidirectional Data Synchronization
 *
 * Provides product and stock level sync from backend:
 * - syncStockLevels(): Download latest stock levels
 * - syncProducts(): Download new/updated products
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import {
  ProductSyncResponse,
  ProductSyncError,
  LAST_PRODUCT_SYNC_KEY,
} from '../types/bidirectional-sync.types';

/**
 * ProductSyncService - Singleton service for product sync
 * Follows service class pattern from SyncQueue
 */
class ProductSyncServiceClass {
  private static instance: ProductSyncServiceClass;
  private baseUrl: string;
  private mockMode: boolean = false;

  private constructor() {
    // Base URL from environment or default to localhost
    this.baseUrl = __DEV__
      ? 'http://localhost:8080/api/v1'
      : 'https://api.simpo.pharmacy/api/v1';

    // Mock mode: enabled in development until backend is ready
    this.mockMode = __DEV__;
  }

  /**
   * Get singleton instance
   */
  static getInstance(): ProductSyncServiceClass {
    if (!ProductSyncServiceClass.instance) {
      ProductSyncServiceClass.instance = new ProductSyncServiceClass();
    }
    return ProductSyncServiceClass.instance;
  }

  /**
   * Set mock mode for testing
   */
  setMockMode(enabled: boolean): void {
    this.mockMode = enabled;
  }

  /**
   * Get auth token from AsyncStorage
   * Reuses JWT token from authentication
   */
  async getAuthToken(): Promise<string | null> {
    try {
      const token = await AsyncStorage.getItem('@simpo_auth_token');
      if (!token) {
        console.warn('[ProductSyncService] No auth token found');
        return null;
      }
      return token;
    } catch (error) {
      console.error('[ProductSyncService] Failed to get auth token:', error);
      return null;
    }
  }

  /**
   * Sync stock levels from server
   * AC2: Download latest stock levels
   */
  async syncStockLevels(): Promise<ProductSyncResponse> {
    if (this.mockMode) {
      return this.mockSyncStockLevels();
    }

    try {
      // Get last sync timestamp
      const lastSync = (await AsyncStorage.getItem('@simpo_last_stock_sync')) || '';

      // Build URL with since parameter
      const url = lastSync
        ? `${this.baseUrl}/products/sync?since=${encodeURIComponent(lastSync)}`
        : `${this.baseUrl}/products/sync`;

      // Get auth token
      const token = await this.getAuthToken();
      if (!token) {
        throw new ProductSyncError('No authentication token available');
      }

      // Create abort controller for timeout
      const abortController = new AbortController();
      const timeoutId = setTimeout(() => abortController.abort(), 30000); // 30 second timeout

      try {
        const response = await fetch(url, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          signal: abortController.signal,
        });

        if (!response.ok) {
          throw new ProductSyncError(
            `Stock sync failed with status ${response.status}`,
            null,
            response.status === 503 // Retryable if 503
          );
        }

        const data: ProductSyncResponse = await response.json().catch(() => ({
          products: [],
          lastSyncTimestamp: new Date().toISOString(),
        }));

        // Update last sync timestamp
        await AsyncStorage.setItem(
          '@simpo_last_stock_sync',
          data.lastSyncTimestamp
        );

        // Update local cache with stock data
        await this.updateStockCache(data.products);

        return data;
      } finally {
        clearTimeout(timeoutId);
      }
    } catch (error) {
      if (error instanceof ProductSyncError) {
        throw error;
      }

      throw new ProductSyncError(
        'Failed to sync stock levels',
        error
      );
    }
  }

  /**
   * Sync products from server
   * AC3: Download new products added since last sync
   */
  async syncProducts(): Promise<ProductSyncResponse> {
    if (this.mockMode) {
      return this.mockSyncProducts();
    }

    try {
      // Get last product sync timestamp
      const lastSync = (await AsyncStorage.getItem(LAST_PRODUCT_SYNC_KEY)) || '';

      // Build URL with since parameter
      const url = lastSync
        ? `${this.baseUrl}/products?since=${encodeURIComponent(lastSync)}`
        : `${this.baseUrl}/products`;

      // Get auth token
      const token = await this.getAuthToken();
      if (!token) {
        throw new ProductSyncError('No authentication token available');
      }

      // Create abort controller for timeout
      const abortController = new AbortController();
      const timeoutId = setTimeout(() => abortController.abort(), 30000); // 30 second timeout

      try {
        const response = await fetch(url, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          signal: abortController.signal,
        });

        if (!response.ok) {
          throw new ProductSyncError(
            `Product sync failed with status ${response.status}`,
            null,
            response.status === 503 // Retryable if 503
          );
        }

        const data: ProductSyncResponse = await response.json().catch(() => ({
          products: [],
          lastSyncTimestamp: new Date().toISOString(),
        }));

        // Update last product sync timestamp
        await AsyncStorage.setItem(
          LAST_PRODUCT_SYNC_KEY,
          data.lastSyncTimestamp
        );

        // Update local cache with product data
        await this.updateProductCache(data.products);

        return data;
      } finally {
        clearTimeout(timeoutId);
      }
    } catch (error) {
      if (error instanceof ProductSyncError) {
        throw error;
      }

      throw new ProductSyncError(
        'Failed to sync products',
        error
      );
    }
  }

  /**
   * Update stock cache in AsyncStorage
   * AC2: Update local product cache with current stock quantities
   */
  private async updateStockCache(products: Array<{
    id: number;
    sku: string;
    name: string;
    stock_qty: number;
    updated_at: string;
  }>): Promise<void> {
    try {
      // Get existing cache
      const cacheJson = await AsyncStorage.getItem('@simpo_stock_cache');
      let cache: Record<string, { stock_qty: number; updated_at: string }> = {};

      if (cacheJson) {
        try {
          cache = JSON.parse(cacheJson);
        } catch (error) {
          console.warn('[ProductSyncService] Corrupted stock cache, resetting:', error);
          cache = {}; // Fallback to empty cache
        }
      }

      // Update cache with new stock levels
      for (const product of products) {
        // Validate stock quantity
        const stockQty = typeof product.stock_qty === 'number' && !isNaN(product.stock_qty)
          ? product.stock_qty
          : 0;

        cache[product.sku] = {
          stock_qty: stockQty,
          updated_at: product.updated_at,
        };
      }

      // Save updated cache
      await AsyncStorage.setItem('@simpo_stock_cache', JSON.stringify(cache));
    } catch (error) {
      console.error('[ProductSyncService] Failed to update stock cache:', error);
    }
  }

  /**
   * Update product cache in AsyncStorage
   * AC3: Merge new products into local cache
   */
  private async updateProductCache(products: Array<{
    id: number;
    sku: string;
    name: string;
    stock_qty: number;
    updated_at: string;
  }>): Promise<void> {
    try {
      // Get existing cache
      const cacheJson = await AsyncStorage.getItem('@simpo_product_cache');
      let cache: Record<string, any> = {};

      if (cacheJson) {
        try {
          cache = JSON.parse(cacheJson);
        } catch (error) {
          console.warn('[ProductSyncService] Corrupted product cache, resetting:', error);
          cache = {}; // Fallback to empty cache
        }
      }

      // Merge new products into cache
      for (const product of products) {
        cache[product.sku] = product;
      }

      // Save updated cache
      await AsyncStorage.setItem('@simpo_product_cache', JSON.stringify(cache));
    } catch (error) {
      console.error('[ProductSyncService] Failed to update product cache:', error);
    }
  }

  /**
   * Mock syncStockLevels for testing
   */
  private async mockSyncStockLevels(): Promise<ProductSyncResponse> {
    // Simulate network latency
    await new Promise((resolve) => setTimeout(resolve, 200));

    return {
      products: [
        {
          id: 1,
          sku: 'SKU-001',
          name: 'Product 1',
          stock_qty: 50,
          updated_at: new Date().toISOString(),
        },
      ],
      lastSyncTimestamp: new Date().toISOString(),
    };
  }

  /**
   * Mock syncProducts for testing
   */
  private async mockSyncProducts(): Promise<ProductSyncResponse> {
    // Simulate network latency
    await new Promise((resolve) => setTimeout(resolve, 200));

    return {
      products: [],
      lastSyncTimestamp: new Date().toISOString(),
    };
  }
}

// Export as ProductSyncService for clarity
export { ProductSyncServiceClass as ProductSyncService };
export default ProductSyncServiceClass.getInstance();
