/**
 * UserSyncService - User data sync from server
 * Story 8.3: Implement Bidirectional Data Synchronization
 *
 * Provides user profile sync from backend:
 * - syncUser(): Download and store current user data
 * - Detects role changes and deactivation
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import {
  UserSyncResponse,
  UserSyncError,
  USER_PROFILE_KEY,
} from '../types/bidirectional-sync.types';

/**
 * UserSyncService - Singleton service for user sync
 * Follows service class pattern from SyncQueue
 */
class UserSyncServiceClass {
  private static instance: UserSyncServiceClass;
  private baseUrl: string;
  private mockMode: boolean = false;
  private currentUserProfile: UserSyncResponse | null = null;
  private isUserDeactivatedFlag: boolean = false;
  private hasRoleChangedFlag: boolean = false;

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
  static getInstance(): UserSyncServiceClass {
    if (!UserSyncServiceClass.instance) {
      UserSyncServiceClass.instance = new UserSyncServiceClass();
    }
    return UserSyncServiceClass.instance;
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
        console.warn('[UserSyncService] No auth token found');
        return null;
      }
      return token;
    } catch (error) {
      console.error('[UserSyncService] Failed to get auth token:', error);
      return null;
    }
  }

  /**
   * Sync user data from server
   * AC4: Download and update user profile, detect role changes and deactivation
   */
  async syncUser(): Promise<UserSyncResponse> {
    if (this.mockMode) {
      return this.mockSyncUser();
    }

    try {
      // Get existing user profile to compare changes
      const existingProfileJson = await AsyncStorage.getItem(USER_PROFILE_KEY);
      const existingProfile = existingProfileJson
        ? (JSON.parse(existingProfileJson) as UserSyncResponse)
        : null;

      // Get auth token
      const token = await this.getAuthToken();
      if (!token) {
        throw new UserSyncError('No authentication token available', null, false);
      }

      // Create abort controller for timeout
      const abortController = new AbortController();
      const timeoutId = setTimeout(() => abortController.abort(), 30000); // 30 second timeout

      try {
        const response = await fetch(`${this.baseUrl}/users/me`, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          signal: abortController.signal,
        });

        if (!response.ok) {
          if (response.status === 401) {
            // Unauthorized - token may be expired
            throw new UserSyncError(
              'Authentication failed - token may be expired',
              null,
              false // Not retryable - requires re-login
            );
          }

          throw new UserSyncError(
            `User sync failed with status ${response.status}`,
            null,
            response.status === 503 // Retryable if 503
          );
        }

        const data: UserSyncResponse = await response.json().catch(() => ({
          id: 0,
          username: 'unknown',
          email: '',
          role: 'cashier',
          status: 'inactive',
          updated_at: new Date().toISOString(),
        }));

        // Check for role change
        if (existingProfile && existingProfile.role !== data.role) {
          this.hasRoleChangedFlag = true;
          console.log(
            `[UserSyncService] Role changed: ${existingProfile.role} → ${data.role}`
          );
        } else {
          this.hasRoleChangedFlag = false;
        }

        // Check for deactivation
        if (existingProfile && existingProfile.status === 'active' && data.status === 'inactive') {
          this.isUserDeactivatedFlag = true;
          console.warn('[UserSyncService] User has been deactivated');
        } else {
          this.isUserDeactivatedFlag = false;
        }

        // Store updated user profile
        await AsyncStorage.setItem(USER_PROFILE_KEY, JSON.stringify(data));
        this.currentUserProfile = data;

        return data;
      } finally {
        clearTimeout(timeoutId);
      }
    } catch (error) {
      if (error instanceof UserSyncError) {
        throw error;
      }

      throw new UserSyncError(
        'Failed to sync user data',
        error
      );
    }
  }

  /**
   * Get stored user profile
   */
  async getUserProfile(): Promise<UserSyncResponse | null> {
    try {
      const profileJson = await AsyncStorage.getItem(USER_PROFILE_KEY);

      if (!profileJson) {
        return null;
      }

      return JSON.parse(profileJson) as UserSyncResponse;
    } catch (error) {
      console.error('[UserSyncService] Failed to load user profile:', error);
      return null;
    }
  }

  /**
   * Clear user profile from storage
   */
  async clearUserProfile(): Promise<void> {
    try {
      await AsyncStorage.removeItem(USER_PROFILE_KEY);
      this.currentUserProfile = null;
      this.isUserDeactivatedFlag = false;
      this.hasRoleChangedFlag = false;
    } catch (error) {
      console.error('[UserSyncService] Failed to clear user profile:', error);
    }
  }

  /**
   * Check if user has been deactivated
   * AC4: Detect inactive status
   */
  isUserDeactivated(): boolean {
    return this.isUserDeactivatedFlag;
  }

  /**
   * Check if user role has changed
   * AC4: Detect role changes
   */
  hasRoleChanged(): boolean {
    return this.hasRoleChangedFlag;
  }

  /**
   * Mock syncUser for testing
   */
  private async mockSyncUser(): Promise<UserSyncResponse> {
    // Simulate network latency
    await new Promise((resolve) => setTimeout(resolve, 100));

    return {
      id: 1,
      username: 'testuser',
      email: 'test@example.com',
      role: 'cashier',
      status: 'active',
      updated_at: new Date().toISOString(),
    };
  }
}

// Export as UserSyncService for clarity
export { UserSyncServiceClass as UserSyncService };
export default UserSyncServiceClass.getInstance();
