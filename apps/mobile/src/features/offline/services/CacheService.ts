/**
 * Cache Service for Stock Data
 * Story 8.1 AC6: Stock Data Caching for Offline Mode
 *
 * Provides caching functionality for product stock data
 * Enables offline stock display using cached data
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import {
  CACHE_LAST_STOCK_SYNC,
  CACHE_STOCK_DATA,
  CachedStockData,
} from '../types/offline.types';

/**
 * CacheService - Stock data caching service
 */
export const CacheService = {
  /**
   * Write stock data to cache
   * Story 8.1 AC6: Cache product stock data in AsyncStorage
   */
  writeStockCache: async (products: CachedStockData['products']): Promise<void> => {
    try {
      const cacheData: CachedStockData = {
        products,
        timestamp: new Date().toISOString(),
      };

      await AsyncStorage.setItem(CACHE_STOCK_DATA, JSON.stringify(cacheData));

      // Update last sync timestamp
      await AsyncStorage.setItem(CACHE_LAST_STOCK_SYNC, cacheData.timestamp);
    } catch (error) {
      // Silently fail - cache is optional functionality
      console.warn('Failed to write stock cache:', error);
    }
  },

  /**
   * Read stock data from cache
   * Story 8.1 AC6: Display cached stock levels when offline
   */
  readStockCache: async (): Promise<CachedStockData | null> => {
    try {
      const cacheStr = await AsyncStorage.getItem(CACHE_STOCK_DATA);
      if (!cacheStr) {
        return null;
      }

      const cacheData: CachedStockData = JSON.parse(cacheStr);

      // Validate cache structure
      if (!cacheData.products || !Array.isArray(cacheData.products)) {
        return null;
      }

      return cacheData;
    } catch (error) {
      // Return null on error - treat as cache miss
      console.warn('Failed to read stock cache:', error);
      return null;
    }
  },

  /**
   * Get last stock sync timestamp
   * Story 8.1 AC6: Show "Stok terakhir sync: {timestamp}" indicator
   */
  getLastSyncTimestamp: async (): Promise<string | null> => {
    try {
      const timestamp = await AsyncStorage.getItem(CACHE_LAST_STOCK_SYNC);
      return timestamp;
    } catch (error) {
      return null;
    }
  },

  /**
   * Clear stock cache
   * Useful for testing or manual cache invalidation
   */
  clearStockCache: async (): Promise<void> => {
    try {
      await AsyncStorage.removeItem(CACHE_STOCK_DATA);
      await AsyncStorage.removeItem(CACHE_LAST_STOCK_SYNC);
    } catch (error) {
      // Silently fail
      console.warn('Failed to clear stock cache:', error);
    }
  },

  /**
   * Check if cache is stale (older than specified minutes)
   * @param maxAgeMinutes Maximum age in minutes (default: 60 minutes)
   */
  isCacheStale: async (maxAgeMinutes: number = 60): Promise<boolean> => {
    try {
      const timestamp = await CacheService.getLastSyncTimestamp();
      if (!timestamp) {
        return true; // No cache = stale
      }

      const cacheTime = new Date(timestamp).getTime();
      const now = Date.now();
      const maxAgeMs = maxAgeMinutes * 60 * 1000;

      return (now - cacheTime) > maxAgeMs;
    } catch (error) {
      return true; // Treat errors as stale
    }
  },
};

export default CacheService;
