/**
 * useNetworkStatus Hook
 * Story 8.1: Implement Local SQLite Storage for Offline Transactions
 *
 * Provides network connectivity status for offline mode detection
 * Returns isConnected boolean state with debounced updates
 *
 * Follows hook pattern from useReceiptPrinter.ts and useBarcodeScanner.ts
 */

import { useState, useEffect, useRef } from 'react';
import NetInfo, { NetInfoState } from '@react-native-community/netinfo';

/**
 * Debounce delay for network state changes (500ms)
 * Prevents rapid re-renders from flaky network connections
 */
const DEBOUNCE_DELAY = 500;

/**
 * Network Status Hook Result
 */
interface NetworkStatus {
  isConnected: boolean;
}

/**
 * useNetworkStatus - Hook for network connectivity monitoring
 *
 * Subscribes to network state changes and provides isConnected boolean
 * Implements debouncing to prevent rapid state changes from flaky connections
 *
 * @returns NetworkStatus with isConnected boolean state
 *
 * @example
 * ```tsx
 * function POSScreen() {
 *   const { isConnected } = useNetworkStatus();
 *
 *   return (
 *     <View>
 *       {!isConnected && <Text>Mode Offline</Text>}
 *       {/* POS UI *\/}
 *     </View>
 *   );
 * }
 * ```
 */
export function useNetworkStatus(): NetworkStatus {
  const [isConnected, setIsConnected] = useState<boolean>(true);
  const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const subscriptionRef = useRef<any>(null);

  useEffect(() => {
    let mounted = true;

    // Initial network state check
    const checkInitialNetworkState = async () => {
      try {
        const state: NetInfoState = await NetInfo.fetch();
        if (mounted) {
          setIsConnected(state.isConnected ?? true); // Default to true on null
        }
      } catch (error) {
        // Default to true (optimistic) on error
        if (mounted) {
          setIsConnected(true);
        }
      }
    };

    checkInitialNetworkState();

    // Subscribe to network state changes
    try {
      subscriptionRef.current = NetInfo.addEventListener((state: NetInfoState) => {
        // Clear existing timer
        if (debounceTimerRef.current) {
          clearTimeout(debounceTimerRef.current);
        }

        // Debounce state update
        debounceTimerRef.current = setTimeout(() => {
          if (mounted) {
            setIsConnected(state.isConnected ?? true);
          }
        }, DEBOUNCE_DELAY);
      });
    } catch (error) {
      // Gracefully handle subscription errors
      // Network state will remain at initial check value
    }

    // Cleanup function
    return () => {
      mounted = false;

      // Clear debounce timer
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
        debounceTimerRef.current = null;
      }

      // Remove network subscription
      if (subscriptionRef.current) {
        subscriptionRef.current.remove();
        subscriptionRef.current = null;
      }
    };
  }, []);

  return { isConnected };
}

export default useNetworkStatus;
