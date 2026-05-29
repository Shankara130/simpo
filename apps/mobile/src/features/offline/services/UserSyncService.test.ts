/**
 * UserSyncService Tests
 * Story 8.3: Implement Bidirectional Data Synchronization
 *
 * Test coverage for user data sync operations
 */

import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { UserSyncService } from './UserSyncService';
import { UserSyncResponse, USER_PROFILE_KEY } from '../types/bidirectional-sync.types';

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
}));

// Mock fetch for API calls
global.fetch = jest.fn();

describe('UserSyncService', () => {
  let service: UserSyncService;
  const mockGetItem = AsyncStorage.getItem as jest.MockedFunction<typeof AsyncStorage.getItem>;
  const mockSetItem = AsyncStorage.setItem as jest.MockedFunction<typeof AsyncStorage.setItem>;
  const mockFetch = global.fetch as jest.MockedFunction<typeof global.fetch>;

  beforeEach(() => {
    jest.clearAllMocks();
    service = UserSyncService.getInstance();
  });

  describe('getInstance', () => {
    it('should return singleton instance', () => {
      const instance1 = UserSyncService.getInstance();
      const instance2 = UserSyncService.getInstance();

      expect(instance1).toBe(instance2);
    });
  });

  describe('syncUser()', () => {
    it('should fetch and store user data', async () => {
      // Mock API response
      const mockResponse: UserSyncResponse = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        role: 'cashier',
        status: 'active',
        profile: {
          firstName: 'Test',
          lastName: 'User',
          phone: '1234567890',
        },
        updated_at: '2026-05-29T10:00:00Z',
      };

      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      const result = await service.syncUser();

      // Verify API was called
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/users/me'),
        expect.any(Object)
      );

      // Verify result
      expect(result.role).toBe('cashier');
      expect(result.status).toBe('active');

      // Verify user data was saved
      expect(mockSetItem).toHaveBeenCalledWith(
        USER_PROFILE_KEY,
        JSON.stringify(mockResponse)
      );
    });

    it('should return user inactive status when user is deactivated', async () => {
      const mockResponse: UserSyncResponse = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        role: 'cashier',
        status: 'inactive',
        updated_at: '2026-05-29T10:00:00Z',
      };

      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      const result = await service.syncUser();

      expect(result.status).toBe('inactive');
      expect(mockSetItem).toHaveBeenCalledWith(
        USER_PROFILE_KEY,
        JSON.stringify(mockResponse)
      );
    });

    it('should return updated role when role changed', async () => {
      // Mock existing user profile with 'cashier' role
      const existingProfile = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        role: 'cashier',
        status: 'active',
        updated_at: '2026-05-29T09:00:00Z',
      };

      mockGetItem.mockResolvedValue(JSON.stringify(existingProfile));

      // Mock API response with 'owner' role
      const mockResponse: UserSyncResponse = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        role: 'owner',
        status: 'active',
        updated_at: '2026-05-29T10:00:00Z',
      };

      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      const result = await service.syncUser();

      // Verify role change is detected
      expect(service.hasRoleChanged()).toBe(true);
      expect(result.role).toBe('owner');
    });

    it('should handle network errors gracefully', async () => {
      mockFetch.mockRejectedValue(new Error('Network error'));

      await expect(service.syncUser()).rejects.toThrow('Network error');
    });

    it('should handle API errors gracefully', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 401,
      } as Response);

      await expect(service.syncUser()).rejects.toThrow();
    });

    it('should handle auth token missing gracefully', async () => {
      // Mock no auth token
      mockGetItem.mockResolvedValue(null);

      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => ({ id: 1, username: 'test', email: 'test@test.com', role: 'cashier', status: 'active', updated_at: '2026-05-29T10:00:00Z' }),
      } as Response);

      const result = await service.syncUser();

      // Should throw error about missing auth token
      await expect(service.syncUser()).rejects.toThrow('No authentication token');
    });

    it('should detect user status change from active to inactive', async () => {
      // Mock existing active user profile
      const existingProfile = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        role: 'cashier',
        status: 'active',
        updated_at: '2026-05-29T09:00:00Z',
      };

      mockGetItem.mockResolvedValue(JSON.stringify(existingProfile));

      // Mock API response with inactive status
      const mockResponse: UserSyncResponse = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        role: 'cashier',
        status: 'inactive',
        updated_at: '2026-05-29T10:00:00Z',
      };

      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      await service.syncUser();

      // Verify status change is detected
      expect(service.isUserDeactivated()).toBe(true);
    });
  });

  describe('getUserProfile()', () => {
    it('should return stored user profile', async () => {
      const mockProfile = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        role: 'cashier',
        status: 'active',
        updated_at: '2026-05-29T09:00:00Z',
      };

      mockGetItem.mockResolvedValue(JSON.stringify(mockProfile));

      const profile = await service.getUserProfile();

      expect(profile).toEqual(mockProfile);
    });

    it('should return null if no profile exists', async () => {
      mockGetItem.mockResolvedValue(null);

      const profile = await service.getUserProfile();

      expect(profile).toBeNull();
    });
  });

  describe('clearUserProfile()', () => {
    it('should remove user profile from storage', async () => {
      await service.clearUserProfile();

      expect(mockSetItem).toHaveBeenCalledWith(USER_PROFILE_KEY, null);
    });
  });
});
