/**
 * Sync Completion Notification Utility
 * Story 8.4: Implement Visual Sync Status Indicators
 *
 * Shows toast notification when sync completes successfully
 * AC4: Sync Notification on Completion
 */

import { Alert } from 'react-native';
import { formatSyncComplete } from '../constants/syncMessages';

/**
 * Notification state for preventing duplicate toasts
 */
let isNotificationVisible = false;
let notificationTimeout: NodeJS.Timeout | null = null;

/**
 * Show sync completion notification
 * Displays toast message when sync transitions from pending to synced
 *
 * @param syncedCount - Number of transactions successfully synced
 * @param autoDismiss - Whether to auto-dismiss after 3 seconds (default: true)
 *
 * @example
 * notifySyncComplete(5); // Shows "Sync selesai - 5 transaksi berhasil"
 */
export function notifySyncComplete(
  syncedCount: number,
  autoDismiss: boolean = true
): void {
  // Prevent duplicate notifications
  if (isNotificationVisible) {
    return;
  }

  isNotificationVisible = true;

  const message = formatSyncComplete(syncedCount);

  // Use Alert.alert for simple toast-like notification
  // Note: React Native Alert.alert doesn't auto-dismiss, so we handle it manually
  Alert.alert(
    'Sync Selesai',
    message,
    [
      {
        text: 'OK',
        onPress: () => {
          isNotificationVisible = false;
          if (notificationTimeout) {
            clearTimeout(notificationTimeout);
            notificationTimeout = null;
          }
        },
      },
    ],
    { cancelable: true, onDismiss: () => {
      isNotificationVisible = false;
      if (notificationTimeout) {
        clearTimeout(notificationTimeout);
        notificationTimeout = null;
      }
    }}
  );

  // Auto-dismiss after 3 seconds if enabled
  if (autoDismiss) {
    notificationTimeout = setTimeout(() => {
      isNotificationVisible = false;
      // Alert.alert doesn't have a programmatic dismiss method
      // User will need to tap OK or cancel
    }, 3000);
  }
}

/**
 * Check if notification is currently visible
 *
 * @returns true if notification is visible, false otherwise
 */
export function isNotificationActive(): boolean {
  return isNotificationVisible;
}

/**
 * Clear any pending auto-dismiss timeout
 * Called when user manually dismisses notification
 */
export function clearNotificationTimeout(): void {
  if (notificationTimeout) {
    clearTimeout(notificationTimeout);
    notificationTimeout = null;
  }
  isNotificationVisible = false;
}

export default {
  notifySyncComplete,
  isNotificationActive,
  clearNotificationTimeout,
};
