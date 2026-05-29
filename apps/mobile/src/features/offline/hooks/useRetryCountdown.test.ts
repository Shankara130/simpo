/**
 * useRetryCountdown Hook Tests
 * Story 8.4: Implement Visual Sync Status Indicators
 */

import { renderHook, act } from '@testing-library/react-hooks';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { useRetryCountdown, getNextRetryTimestamp } from '../useRetryCountdown';

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage');

describe('useRetryCountdown', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  describe('useRetryCountdown hook', () => {
    it('should return 0 when no retry scheduled', () => {
      const { result } = renderHook(() => useRetryCountdown(null));

      expect(result.current).toBe(0);
    });

    it('should calculate initial countdown when retry scheduled', () => {
      // Create a timestamp 5 minutes in the future
      const now = new Date();
      const fiveMinutesFromNow = new Date(now.getTime() + 5 * 60000).toISOString();

      const { result } = renderHook(() => useRetryCountdown(fiveMinutesFromNow));

      expect(result.current).toBe(5);
    });

    it('should return 0 when retry time has passed', () => {
      // Create a timestamp 5 minutes in the past
      const now = new Date();
      const fiveMinutesAgo = new Date(now.getTime() - 5 * 60000).toISOString();

      const { result } = renderHook(() => useRetryCountdown(fiveMinutesAgo));

      expect(result.current).toBe(0);
    });

    it('should update countdown every minute', () => {
      // Create a timestamp 2 minutes in the future
      const now = new Date();
      const twoMinutesFromNow = new Date(now.getTime() + 2 * 60000).toISOString();

      const { result } = renderHook(() => useRetryCountdown(twoMinutesFromNow));

      expect(result.current).toBe(2);

      // Fast-forward time by 1 minute
      act(() => {
        jest.advanceTimersByTime(60000);
      });

      expect(result.current).toBe(1);

      // Fast-forward another minute
      act(() => {
        jest.advanceTimersByTime(60000);
      });

      expect(result.current).toBe(0);
    });

    it('should clear interval on unmount', () => {
      const clearIntervalSpy = jest.spyOn(global, 'clearInterval');

      const { unmount } = renderHook(() => useRetryCountdown(null));

      unmount();

      // No intervals should be active after unmount
      expect(clearIntervalSpy).not.toHaveBeenCalled();
    });
  });

  describe('getNextRetryTimestamp function', () => {
    beforeEach(() => {
      jest.clearAllMocks();
    });

    it('should return null when no retry schedule exists', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);

      const result = await getNextRetryTimestamp();

      expect(result).toBeNull();
      expect(AsyncStorage.getItem).toHaveBeenCalledWith('@simpo_sync_retry_schedule');
    });

    it('should return null when retry schedule is empty', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce('{}');

      const result = await getNextRetryTimestamp();

      expect(result).toBeNull();
    });

    it('should return next retry timestamp from schedule', async () => {
      const now = new Date();
      const tenMinutesFromNow = new Date(now.getTime() + 10 * 60000).toISOString();
      const thirtyMinutesFromNow = new Date(now.getTime() + 30 * 60000).toISOString();

      const retrySchedule = {
        'trx-1': { retryAt: thirtyMinutesFromNow, retryCount: 2 },
        'trx-2': { retryAt: tenMinutesFromNow, retryCount: 1 },
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(
        JSON.stringify(retrySchedule)
      );

      const result = await getNextRetryTimestamp();

      // Should return the earliest retry time (10 minutes)
      expect(result).toBe(tenMinutesFromNow);
    });

    it('should ignore past retry timestamps', async () => {
      const now = new Date();
      const tenMinutesAgo = new Date(now.getTime() - 10 * 60000).toISOString();
      const fiveMinutesFromNow = new Date(now.getTime() + 5 * 60000).toISOString();

      const retrySchedule = {
        'trx-1': { retryAt: tenMinutesAgo, retryCount: 2 },
        'trx-2': { retryAt: fiveMinutesFromNow, retryCount: 1 },
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(
        JSON.stringify(retrySchedule)
      );

      const result = await getNextRetryTimestamp();

      // Should only return future retry times (5 minutes)
      expect(result).toBe(fiveMinutesFromNow);
    });

    it('should return null when all retries are in the past', async () => {
      const now = new Date();
      const tenMinutesAgo = new Date(now.getTime() - 10 * 60000).toISOString();
      const fiveMinutesAgo = new Date(now.getTime() - 5 * 60000).toISOString();

      const retrySchedule = {
        'trx-1': { retryAt: tenMinutesAgo, retryCount: 2 },
        'trx-2': { retryAt: fiveMinutesAgo, retryCount: 1 },
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(
        JSON.stringify(retrySchedule)
      );

      const result = await getNextRetryTimestamp();

      expect(result).toBeNull();
    });

    it('should handle JSON parse errors gracefully', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce('invalid json');

      const result = await getNextRetryTimestamp();

      expect(result).toBeNull();
    });
  });
});
