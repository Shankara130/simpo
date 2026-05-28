/**
 * useOfflineMode Hook
 * Story 8.1 AC7: Offline-Only UI Restrictions
 *
 * Provides utilities for implementing UI restrictions based on network status
 * Components can use this hook to disable features when offline
 */

import { useMemo } from 'react';
import { useNetworkStatus } from '../hooks/useNetworkStatus';

/**
 * Offline mode restrictions
 * Defines which features are available when offline
 */
interface OfflineModeRestrictions {
  isOffline: boolean;
  canProcessTransactions: boolean;
  canViewMultiBranch: boolean;
  canRegisterUsers: boolean;
  canViewReports: boolean;
  offlineMessage: string;
}

/**
 * useOfflineMode - Hook for offline mode UI restrictions
 * Story 8.1 AC7: Disable features when offline
 *
 * Provides boolean flags for feature availability and offline message
 *
 * @returns OfflineModeRestrictions with feature availability flags
 *
 * @example
 * ```tsx
 * function POSScreen() {
 *   const { isOffline, canProcessTransactions, offlineMessage } = useOfflineMode();
 *
 *   return (
 *     <View>
 *       {isOffline && <Text>{offlineMessage}</Text>}
 *       <Button disabled={!canProcessTransactions}>
 *         Process Transaction
 *       </Button>
 *     </View>
 *   );
 * }
 * ```
 */
export function useOfflineMode(): OfflineModeRestrictions {
  const { isConnected } = useNetworkStatus();
  const isOffline = !isConnected;

  return useMemo(() => ({
    isOffline,
    // Core POS features work offline (Story 8.1 AC7)
    canProcessTransactions: true,
    // Multi-branch visibility disabled when offline (Story 8.1 AC7)
    canViewMultiBranch: !isOffline,
    // User registration requires online connectivity (Story 8.1 AC7)
    canRegisterUsers: !isOffline,
    // Advanced reports disabled when offline (Story 8.1 AC7)
    canViewReports: !isOffline,
    // Indonesian offline message (Story 8.1 AC7)
    offlineMessage: 'Mode Offline - Fitur terbatas tersedia',
  }), [isOffline]);
}

export default useOfflineMode;
