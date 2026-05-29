/**
 * useBidirectionalSync Hook
 * Story 8.3: Implement Bidirectional Data Synchronization
 *
 * Hook for bidirectional sync state management and UI indicators
 * Subscribes to SyncOrchestrator events and persists state to AsyncStorage
 * AC7: Visual sync status indicators with phase information
 */

import { useState, useEffect, useCallback } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { SyncOrchestrator } from '../services/SyncOrchestrator';
import {
  BidirectionalSyncState,
  SyncPhase,
  BIDIRECTIONAL_SYNC_STATE_KEY,
  BIDIRECTIONAL_SYNC_RETRY_KEY,
  MAX_RETRY_ATTEMPTS,
  BASE_RETRY_DELAY_MS,
} from '../types/bidirectional-sync.types';

/**
 * Bidirectional Sync Hook
 * Manages real-time bidirectional sync state with phase information
 *
 * @returns BidirectionalSyncState object with current sync metrics and actions
 */
export function useBidirectionalSync() {
  const [syncState, setSyncState] = useState<BidirectionalSyncState>({
    status: 'idle',
    phase: 'idle',
    pendingCount: 0,
    processingCount: 0,
    syncedCount: 0,
    failedCount: 0,
    currentPhase: null,
    lastSyncTime: null,
  });

  const [isInitialized, setIsInitialized] = useState(false);
  const [orchestrator] = useState(() => SyncOrchestrator.getInstance());

  /**
   * Load persisted sync state from AsyncStorage
   */
  const loadPersistedState = useCallback(async (): Promise<BidirectionalSyncState | null> => {
    try {
      const savedState = await AsyncStorage.getItem(BIDIRECTIONAL_SYNC_STATE_KEY);
      if (savedState) {
        return JSON.parse(savedState) as BidirectionalSyncState;
      }
    } catch (error) {
      console.warn('[useBidirectionalSync] Failed to load persisted state:', error);
    }
    return null;
  }, []);

  /**
   * Save sync state to AsyncStorage
   */
  const savePersistedState = useCallback(async (state: BidirectionalSyncState): Promise<void> => {
    try {
      await AsyncStorage.setItem(BIDIRECTIONAL_SYNC_STATE_KEY, JSON.stringify(state));
    } catch (error) {
      console.warn('[useBidirectionalSync] Failed to persist state:', error);
    }
  }, []);

  /**
   * Start bidirectional sync
   * Calls orchestrator.sync() and updates state based on result
   */
  const startSync = useCallback(async () => {
    try {
      const result = await orchestrator.sync();

      // Update state based on sync result
      setSyncState((prevState) => {
        const newState: BidirectionalSyncState = {
          ...prevState,
          status: result.success ? 'synced' : 'failed',
          phase: result.phase,
          syncedCount: result.success ? prevState.syncedCount + (result.uploaded || 0) : prevState.syncedCount,
          failedCount: result.success ? 0 : prevState.failedCount + 1,
          currentPhase: getPhaseDescription(result.phase),
          error: result.error,
          lastSyncTime: result.success ? new Date().toISOString() : prevState.lastSyncTime,
        };

        // Persist updated state
        savePersistedState(newState);

        return newState;
      });

      return result;
    } catch (error) {
      // Handle unexpected errors
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      setSyncState((prevState) => {
        const newState: BidirectionalSyncState = {
          ...prevState,
          status: 'failed',
          phase: 'failed',
          currentPhase: 'Sync failed',
          error: errorMessage,
        };

        savePersistedState(newState);
        return newState;
      });

      throw error;
    }
  }, [orchestrator, savePersistedState]);

  /**
   * Stop ongoing sync process
   * Calls orchestrator.stopSync()
   */
  const stopSync = useCallback(() => {
    orchestrator.stopSync();

    setSyncState((prevState) => {
      const newState: BidirectionalSyncState = {
        ...prevState,
        status: 'idle',
        phase: 'idle',
        currentPhase: null,
        processingCount: 0,
      };

      savePersistedState(newState);
      return newState;
    });
  }, [orchestrator, savePersistedState]);

  /**
   * Get retry information
   * Returns retry count and next retry time if applicable
   */
  const getRetryInfo = useCallback(async () => {
    try {
      const retryDataJson = await AsyncStorage.getItem(BIDIRECTIONAL_SYNC_RETRY_KEY);
      if (retryDataJson) {
        const retryData = JSON.parse(retryDataJson) as { attempts: number; lastAttempt: string };
        return {
          attempts: retryData.attempts,
          lastAttempt: retryData.lastAttempt,
          canRetry: retryData.attempts < MAX_RETRY_ATTEMPTS,
        };
      }
    } catch (error) {
      console.warn('[useBidirectionalSync] Failed to load retry info:', error);
    }
    return null;
  }, []);

  /**
   * Get human-readable phase description
   */
  const getPhaseDescription = useCallback((phase: SyncPhase): string | null => {
    const descriptions: Record<SyncPhase, string | null> = {
      idle: null,
      uploading: 'Uploading pending transactions',
      downloading_stock: 'Downloading stock levels',
      downloading_products: 'Downloading new products',
      downloading_user: 'Downloading user data',
      synced: 'All data synchronized',
      failed: 'Sync failed - will retry',
    };

    return descriptions[phase];
  }, []);

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

      setIsInitialized(true);
    };

    initialize();

    return () => {
      mounted = false;
    };
  }, [loadPersistedState]);

  /**
   * Monitor sync state changes
   * Poll orchestrator state periodically for updates
   */
  useEffect(() => {
    if (!isInitialized) {
      return;
    }

    let mounted = true;

    // Poll every 2 seconds for state changes
    const interval = setInterval(async () => {
      if (mounted) {
        const orchestratorState = await orchestrator.getSyncState();
        setSyncState((prevState) => {
          // Only update if phase or status changed
          if (
            orchestratorState.phase !== prevState.phase ||
            orchestratorState.status !== prevState.status
          ) {
            return {
              ...prevState,
              ...orchestratorState,
            };
          }
          return prevState;
        });
      }
    }, 2000);

    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, [isInitialized, orchestrator]);

  return {
    ...syncState,
    isInitialized,
    startSync,
    stopSync,
    getRetryInfo,
  };
}
