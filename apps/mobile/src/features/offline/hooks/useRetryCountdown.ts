/**
 * useRetryCountdown Hook
 * Story 8.4: Implement Visual Sync Status Indicators
 *
 * Calculates and updates retry countdown every minute
 * Reads from retry schedule persisted by SyncQueueService (Story 8.2)
 */

import { useState, useEffect } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { formatCountdown } from '../constants/syncMessages';

/**
 * Retry countdown hook
 *
 * Calculates remaining time until next retry attempt
 * Updates countdown every 60 seconds
 * Handles countdown reaching zero
 *
 * @param nextRetryAt - ISO timestamp of next retry attempt
 * @returns countdown - Number of minutes remaining (0 if no retry scheduled)
 */
export function useRetryCountdown(nextRetryAt: string | null) {
  const [countdown, setCountdown] = useState<number>(0);

  useEffect(() => {
    // No retry scheduled - reset countdown
    if (!nextRetryAt) {
      setCountdown(0);
      return;
    }

    let mounted = true;

    /**
     * Calculate remaining minutes until retry
     */
    const calculateCountdown = () => {
      try {
        const now = new Date().getTime();
        const retryTime = new Date(nextRetryAt).getTime();
        const millisecondsRemaining = retryTime - now;
        const minutes = Math.max(0, Math.floor(millisecondsRemaining / 60000));

        if (mounted) {
          setCountdown(minutes);
        }
      } catch (error) {
        console.warn('[useRetryCountdown] Failed to calculate countdown:', error);
        if (mounted) {
          setCountdown(0);
        }
      }
    };

    // Initial calculation
    calculateCountdown();

    // Update countdown every minute
    const interval = setInterval(calculateCountdown, 60000);

    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, [nextRetryAt]);

  return countdown;
}

/**
 * Get next retry timestamp from retry schedule
 * Reads from AsyncStorage retry schedule maintained by SyncQueueService
 *
 * @returns Next retry timestamp or null if no retry scheduled
 */
export async function getNextRetryTimestamp(): Promise<string | null> {
  try {
    const retryScheduleJson = await AsyncStorage.getItem('@simpo_sync_retry_schedule');
    if (!retryScheduleJson) {
      return null;
    }

    const retrySchedule = JSON.parse(retryScheduleJson) as Record<
      string,
      { retryAt: string; retryCount: number }
    >;

    // Find the earliest pending retry
    const now = new Date().getTime();
    let earliestRetry: string | null = null;
    let earliestTime = Infinity;

    Object.entries(retrySchedule).forEach(([transactionId, retryInfo]) => {
      const retryTime = new Date(retryInfo.retryAt).getTime();
      if (retryTime > now && retryTime < earliestTime) {
        earliestTime = retryTime;
        earliestRetry = retryInfo.retryAt;
      }
    });

    return earliestRetry;
  } catch (error) {
    console.warn('[useRetryCountdown] Failed to load retry schedule:', error);
    return null;
  }
}

export default useRetryCountdown;
