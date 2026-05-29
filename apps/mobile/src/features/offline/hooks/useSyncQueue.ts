/**
 * useSyncQueue Hook
 * Story 8.2: Implement Transaction Sync Queue
 * AC6: Network State Transition Triggers
 *
 * Automatically triggers sync queue processing when network is restored
 * Handles network state transitions with logging and background prevention
 */

import { useEffect, useRef, useState } from 'react';
import { AppState, AppStateStatus } from 'react-native';
import { SyncQueue } from '../services/SyncQueueService';
import { useNetworkStatus } from './useNetworkStatus';

/**
 * useSyncQueue - Hook for automatic sync queue management
 *
 * Integrates network status monitoring with sync queue processing:
 * - Triggers processQueue() when network transitions from offline to online
 * - Cancels processing when network goes offline mid-process
 * - Prevents auto-sync when app is in background
 * - Logs all network state transitions with timestamps
 *
 * @returns Object with sync queue state and control methods
 *
 * @example
 * ```tsx
 * function POSScreen() {
 *   const { isProcessing, lastSyncTime } = useSyncQueue();
 *
 *   return (
 *     <View>
 *       <Text>Sync: {isProcessing ? 'Processing...' : 'Idle'}</Text>
 *       {/* POS UI *\/}
 *     </View>
 *   );
 * }
 * ```
 */
export function useSyncQueue() {
  const { isConnected } = useNetworkStatus();
  const [isProcessing, setIsProcessing] = useState<boolean>(false);
  const [lastSyncTime, setLastSyncTime] = useState<Date | null>(null);
  const [lastNetworkTransition, setLastNetworkTransition] = useState<{
    from: boolean;
    to: boolean;
    timestamp: string;
  } | null>(null);

  const previousConnectionRef = useRef<boolean>(true);
  const appStateRef = useRef<AppStateStatus>(AppState.currentState);
  const syncQueueRef = useRef<ReturnType<typeof SyncQueue.getInstance>>(SyncQueue.getInstance());

  /**
   * Effect: Monitor network state transitions
   */
  useEffect(() => {
    const syncQueue = syncQueueRef.current;

    // Check for network state transition
    if (previousConnectionRef.current !== isConnected) {
      const timestamp = new Date().toISOString();
      const from = previousConnectionRef.current;
      const to = isConnected;

      // Log network state transition
      console.info(`[SyncQueue] Network state transition: ${from ? 'online' : 'offline'} → ${to ? 'online' : 'offline'} (${timestamp})`);

      // Store transition
      setLastNetworkTransition({ from, to, timestamp });

      // Handle offline → online transition (network restoration)
      if (from === false && to === true) {
        // Only trigger if app is in foreground (not background)
        if (appStateRef.current === 'active') {
          console.info('[SyncQueue] Network restored, triggering sync queue processing...');
          triggerSyncProcessing();
        } else {
          console.info('[SyncQueue] Network restored but app is in background, will sync on foreground');
        }
      }

      // Handle online → offline transition (network loss)
      if (from === true && to === false) {
        console.info('[SyncQueue] Network lost, canceling ongoing sync processing...');
        syncQueue.stopProcessing();
        setIsProcessing(false);
      }

      // Update ref for next comparison
      previousConnectionRef.current = isConnected;
    }
  }, [isConnected]);

  /**
   * Effect: Monitor app state changes (foreground/background)
   */
  useEffect(() => {
    const subscription = AppState.addEventListener('change', (nextAppState: AppStateStatus) => {
      const previousAppState = appStateRef.current;
      appStateRef.current = nextAppState;

      // Log app state transition
      console.info(`[SyncQueue] App state transition: ${previousAppState} → ${nextAppState}`);

      // When app comes to foreground and network is online, trigger sync
      if (nextAppState === 'active' && previousAppState.match(/inactive|background/) && isConnected) {
        console.info('[SyncQueue] App moved to foreground with network, triggering sync...');
        // Use setTimeout to avoid state update during render
        setTimeout(() => triggerSyncProcessing(), 100);
      }

      // When app goes to background, stop ongoing sync processing
      if (nextAppState.match(/inactive|background/)) {
        console.info('[SyncQueue] App moved to background, ensuring no sync processing...');
        const syncQueue = syncQueueRef.current;
        syncQueue.stopProcessing();
        setIsProcessing(false);
      }
    });

    return () => {
      subscription.remove();
    };
  }, [isConnected]); // Include isConnected as dependency

  /**
   * Trigger sync queue processing
   */
  const triggerSyncProcessing = async () => {
    const syncQueue = syncQueueRef.current;

    try {
      setIsProcessing(true);

      // Process the queue
      const result = await syncQueue.processQueue();

      // Update last sync time
      setLastSyncTime(new Date());

      // Log results
      console.info('[SyncQueue] Sync processing complete:', {
        totalProcessed: result.totalProcessed,
        syncedCount: result.syncedCount,
        failedCount: result.failedCount,
        skippedCount: result.skippedCount,
        duration: result.duration,
      });

      setIsProcessing(false);
    } catch (error) {
      console.error('[SyncQueue] Sync processing failed:', error);
      setIsProcessing(false);
    }
  };

  /**
   * Manual trigger for sync queue processing
   * Can be called by user action (e.g., "Sync Now" button)
   */
  const manualSync = async () => {
    console.info('[SyncQueue] Manual sync triggered');
    await triggerSyncProcessing();
  };

  return {
    isProcessing,
    lastSyncTime,
    lastNetworkTransition,
    manualSync,
  };
}

export default useSyncQueue;
