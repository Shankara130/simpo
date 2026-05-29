/**
 * useSyncProgress Hook
 * Story 8.2: Implement Transaction Sync Queue
 *
 * Hook for real-time sync progress state management
 * Subscribes to SyncQueueService events and persists state to AsyncStorage
 */

import { useState, useEffect, useCallback } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { SyncQueue } from '../services/SyncQueueService';
import { SyncState, SYNC_PROGRESS_KEY } from '../types/sync.types';

/**
 * Sync Progress Hook
 * Manages real-time sync state with persistence
 *
 * @returns SyncState object with current sync metrics
 */
export function useSyncProgress() {
  const [syncState, setSyncState] = useState<SyncState>({
    pendingCount: 0,
    processingCount: 0,
    syncedCount: 0,
    failedCount: 0,
    currentTransaction: null,
  });

  const [isInitialized, setIsInitialized] = useState(false);

  /**
   * Load persisted sync state from AsyncStorage
   */
  const loadPersistedState = useCallback(async (): Promise<SyncState | null> => {
    try {
      const savedState = await AsyncStorage.getItem(SYNC_PROGRESS_KEY);
      if (savedState) {
        return JSON.parse(savedState) as SyncState;
      }
    } catch (error) {
      console.warn('[useSyncProgress] Failed to load persisted state:', error);
    }
    return null;
  }, []);

  /**
   * Save sync state to AsyncStorage
   */
  const savePersistedState = useCallback(async (state: SyncState): Promise<void> => {
    try {
      await AsyncStorage.setItem(SYNC_PROGRESS_KEY, JSON.stringify(state));
    } catch (error) {
      console.warn('[useSyncProgress] Failed to persist state:', error);
    }
  }, []);

  /**
   * Refresh pending count from offline storage
   */
  const refreshPendingCount = useCallback(async (): Promise<void> => {
    try {
      const syncQueue = SyncQueue.getInstance();
      const pendingTransactions = await syncQueue.getPendingTransactions();

      setSyncState((prevState) => {
        const newState = {
          ...prevState,
          pendingCount: pendingTransactions.length,
        };

        // Persist updated state
        savePersistedState(newState);

        return newState;
      });
    } catch (error) {
      console.warn('[useSyncProgress] Failed to refresh pending count:', error);
    }
  }, [savePersistedState]);

  /**
   * Initialize on mount
   */
  useEffect(() => {
    let mounted = true;

    const initialize = async () => {
      // Load persisted state
      const savedState = await loadPersistedState();

      if (mounted && savedState) {
        setSyncState(savedState);
      }

      // Refresh pending count from actual storage
      await refreshPendingCount();

      // Check for orphaned processing state and reset if needed
      const syncQueue = SyncQueue.getInstance();
      const isOrphaned = await syncQueue.isProcessingStateOrphaned();

      if (isOrphaned) {
        console.info('[useSyncProgress] Detected orphaned processing state, resetting');
        setSyncState((prevState) => ({
          ...prevState,
          processingCount: 0,
        }));
      }

      setIsInitialized(true);
    };

    initialize();

    return () => {
      mounted = false;
    };
  }, [loadPersistedState, refreshPendingCount]);

  /**
   * Monitor sync queue state changes
   * Note: In a full implementation, this would subscribe to events emitted by SyncQueueService
   * For now, we poll the state periodically
   */
  useEffect(() => {
    if (!isInitialized) {
      return;
    }

    let mounted = true;

    // Poll every 2 seconds for state changes
    const interval = setInterval(() => {
      if (mounted) {
        refreshPendingCount();
      }
    }, 2000);

    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, [isInitialized, refreshPendingCount]);

  /**
   * Reset session counters (syncedCount, failedCount)
   * Useful for starting a new sync session
   */
  const resetSessionCounters = useCallback(() => {
    setSyncState((prevState) => {
      const newState = {
        ...prevState,
        syncedCount: 0,
        failedCount: 0,
      };

      savePersistedState(newState);

      return newState;
    });
  }, [savePersistedState]);

  /**
   * Update processing state
   * Called by SyncQueueService during processing
   */
  const setProcessing = useCallback(
    (isProcessing: boolean, transactionNumber?: string) => {
      setSyncState((prevState) => {
        const newState = {
          ...prevState,
          processingCount: isProcessing ? 1 : 0,
          currentTransaction: transactionNumber || null,
        };

        savePersistedState(newState);

        return newState;
      });
    },
    [savePersistedState]
  );

  /**
   * Increment synced count
   * Called by SyncQueueService after successful sync
   */
  const incrementSyncedCount = useCallback((transactionNumber: string) => {
    setSyncState((prevState) => {
      const newState = {
        ...prevState,
        syncedCount: prevState.syncedCount + 1,
        processingCount: 0,
        currentTransaction: null,
      };

      savePersistedState(newState);

      return newState;
    });
  }, [savePersistedState]);

  /**
   * Increment failed count
   * Called by SyncQueueService after failed sync
   */
  const incrementFailedCount = useCallback((transactionNumber: string) => {
    setSyncState((prevState) => {
      const newState = {
        ...prevState,
        failedCount: prevState.failedCount + 1,
        processingCount: 0,
        currentTransaction: null,
      };

      savePersistedState(newState);

      return newState;
    });
  }, [savePersistedState]);

  return {
    ...syncState,
    isInitialized,
    refreshPendingCount,
    resetSessionCounters,
    setProcessing,
    incrementSyncedCount,
    incrementFailedCount,
  };
}
