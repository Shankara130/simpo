/**
 * ProductSyncService Tests
 * Story 8.3: Implement Bidirectional Data Synchronization
 *
 * Test coverage for product sync operations
 */

import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { ProductSyncService } from './ProductSyncService';
import { ProductSyncResponse, LAST_PRODUCT_SYNC_KEY } from '../types/bidirectional-sync.types';

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
}));

// Mock fetch for API calls
global.fetch = jest.fn();

describe('ProductSyncService', () => {
  let service: ProductSyncService;
  const mockGetItem = AsyncStorage.getItem as jest.MockedFunction<typeof AsyncStorage.getItem>;
  const mockSetItem = AsyncStorage.setItem as jest.MockedFunction<typeof AsyncStorage.setItem>;
  const mockFetch = global.fetch as jest.MockedFunction<typeof global.fetch>;

  beforeEach(() => {
    jest.clearAllMocks();
    service = ProductSyncService.getInstance();
  });

  describe('getInstance', () => {
    it('should return singleton instance', () => {
      const instance1 = ProductSyncService.getInstance();
      const instance2 = ProductSyncService.getInstance();

      expect(instance1).toBe(instance2);
    });
  });

  describe('syncStockLevels()', () => {
    it('should fetch stock levels updated since last sync', async () => {
      // Mock last sync timestamp
      mockGetItem.mockResolvedValue('2026-05-29T09:00:00Z');

      // Mock API response
      const mockResponse: ProductSyncResponse = {
        products: [
          { id: 1, sku: 'SKU-001', name: 'Product 1', stock_qty: 50, updated_at: '2026-05-29T10:00:00Z' },
          { id: 2, sku: 'SKU-002', name: 'Product 2', stock_qty: 30, updated_at: '2026-05-29T10:05:00Z' },
        ],
        lastSyncTimestamp: '2026-05-29T10:05:00Z',
      };

      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      const result = await service.syncStockLevels();

      // Verify API was called with since parameter
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('since=2026-05-29T09%3A00%3A00Z'),
        expect.any(Object)
      );

      // Verify result
      expect(result.products).toHaveLength(2);
      expect(result.lastSyncTimestamp).toBe('2026-05-29T10:05:00Z');

      // Verify timestamp was saved
      expect(mockSetItem).toHaveBeenCalledWith(
        '@simpo_last_stock_sync',
        '2026-05-29T10:05:00Z'
      );
    });

    it('should handle first sync (no timestamp)', async () => {
      // Mock no last sync timestamp
      mockGetItem.mockResolvedValue(null);

      const mockResponse: ProductSyncResponse = {
        products: [
          { id: 1, sku: 'SKU-001', name: 'Product 1', stock_qty: 50, updated_at: '2026-05-29T10:00:00Z' },
        ],
        lastSyncTimestamp: '2026-05-29T10:00:00Z',
      };

      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      const result = await service.syncStockLevels();

      // Verify API was called without since parameter (or with since=0)
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('sync'),
        expect.any(Object)
      );

      // Verify timestamp was saved
      expect(mockSetItem).toHaveBeenCalledWith(
        '@simpo_last_stock_sync',
        '2026-05-29T10:00:00Z'
      );
    });

    it('should throw error on network failure', async () => {
      mockGetItem.mockResolvedValue('2026-05-29T09:00:00Z');

      mockFetch.mockRejectedValue(new Error('Network error'));

      await expect(service.syncStockLevels()).rejects.toThrow('Network error');
    });

    it('should throw error on API error', async () => {
      mockGetItem.mockResolvedValue('2026-05-29T09:00:00Z');

      mockFetch.mockResolvedValue({
        ok: false,
        status: 500,
      } as Response);

      await expect(service.syncStockLevels()).rejects.toThrow();
    });
  });

  describe('syncProducts()', () => {
    it('should fetch products updated since last sync', async () => {
      // Mock last product sync timestamp
      mockGetItem.mockImplementation((key) => {
        if (key === LAST_PRODUCT_SYNC_KEY) {
          return Promise.resolve('2026-05-29T09:00:00Z');
        }
        return Promise.resolve(null);
      });

      // Mock API response
      const mockResponse: ProductSyncResponse = {
        products: [
          { id: 3, sku: 'SKU-003', name: 'Product 3', stock_qty: 20, updated_at: '2026-05-29T10:10:00Z' },
        ],
        lastSyncTimestamp: '2026-05-29T10:10:00Z',
      };

      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      const result = await service.syncProducts();

      // Verify API was called
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('since=2026-05-29T09%3A00%3A00Z'),
        expect.any(Object)
      );

      // Verify result
      expect(result.products).toHaveLength(1);

      // Verify timestamp was saved
      expect(mockSetItem).toHaveBeenCalledWith(
        LAST_PRODUCT_SYNC_KEY,
        '2026-05-29T10:10:00Z'
      );
    });

    it('should handle first product sync (no timestamp)', async () => {
      mockGetItem.mockResolvedValue(null);

      const mockResponse: ProductSyncResponse = {
        products: [],
        lastSyncTimestamp: '2026-05-29T10:00:00Z',
      };

      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      const result = await service.syncProducts();

      expect(result.products).toHaveLength(0);
      expect(mockSetItem).toHaveBeenCalledWith(
        LAST_PRODUCT_SYNC_KEY,
        '2026-05-29T10:00:00Z'
      );
    });
  });

  describe('getAuthToken()', () => {
    it('should retrieve auth token from AsyncStorage', async () => {
      mockGetItem.mockResolvedValue('mock-jwt-token');

      const token = await service.getAuthToken();

      expect(token).toBe('mock-jwt-token');
      expect(mockGetItem).toHaveBeenCalledWith('@simpo_auth_token');
    });

    it('should return null if no token found', async () => {
      mockGetItem.mockResolvedValue(null);

      const token = await service.getAuthToken();

      expect(token).toBeNull();
    });

    it('should handle AsyncStorage errors gracefully', async () => {
      mockGetItem.mockRejectedValue(new Error('Storage error'));

      const token = await service.getAuthToken();

      expect(token).toBeNull();
    });
  });
});
