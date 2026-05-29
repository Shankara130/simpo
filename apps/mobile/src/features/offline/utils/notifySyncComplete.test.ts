/**
 * Sync Completion Notification Utility Tests
 * Story 8.4: Implement Visual Sync Status Indicators
 *
 * Tests for sync completion notification utility
 */

import { notifySyncComplete, isNotificationActive, clearNotificationTimeout } from '../notifySyncComplete';
import { Alert } from 'react-native';

// Mock React Native Alert
jest.mock('react-native', () => ({
  Alert: {
    alert: jest.fn(),
  },
}));

describe('notifySyncComplete utility', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();
    // Reset notification state
    clearNotificationTimeout();
  });

  afterEach(() => {
    jest.useRealTimers();
    clearNotificationTimeout();
  });

  describe('notifySyncComplete', () => {
    it('should display sync completion message with transaction count', () => {
      notifySyncComplete(5);

      expect(Alert.alert).toHaveBeenCalledWith(
        'Sync Selesai',
        'Sync selesai - 5 transaksi berhasil',
        expect.arrayContaining([
          expect.objectContaining({
            text: 'OK',
          }),
        ]),
        { cancelable: true, onDismiss: expect.any(Function) }
      );
    });

    it('should format message correctly for single transaction', () => {
      notifySyncComplete(1);

      expect(Alert.alert).toHaveBeenCalledWith(
        'Sync Selesai',
        'Sync selesai - 1 transaksi berhasil',
        expect.any(Array),
        expect.any(Object)
      );
    });

    it('should prevent duplicate notifications', () => {
      notifySyncComplete(3);

      // Try to show another notification while first is visible
      notifySyncComplete(5);

      // Should only call Alert.alert once
      expect(Alert.alert).toHaveBeenCalledTimes(1);
    });

    it('should set notification as active when shown', () => {
      expect(isNotificationActive()).toBe(false);

      notifySyncComplete(2);

      expect(isNotificationActive()).toBe(true);
    });

    it('should auto-dismiss after 3 seconds by default', () => {
      jest.spyOn(global, 'setTimeout');

      notifySyncComplete(3);

      expect(setTimeout).toHaveBeenCalledWith(expect.any(Function), 3000);
    });

    it('should not auto-dismiss when disabled', () => {
      jest.spyOn(global, 'setTimeout');

      notifySyncComplete(4, false);

      expect(setTimeout).not.toHaveBeenCalled();
    });

    it('should clear timeout on OK button press', () => {
      const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout');

      notifySyncComplete(2);

      // Get the OK button callback from the Alert.alert call
      const alertCall = (Alert.alert as jest.Mock).mock.calls[0];
      const okButton = alertCall[2][0];

      // Simulate OK button press
      okButton.onPress();

      expect(clearTimeout).toHaveBeenCalled();
      expect(isNotificationActive()).toBe(false);
    });

    it('should clear timeout on dismiss', () => {
      const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout');

      notifySyncComplete(2);

      // Get the onDismiss callback from the Alert.alert call
      const alertCall = (Alert.alert as jest.Mock).mock.calls[0];
      const onDismiss = alertCall[3].onDismiss;

      // Simulate dismiss
      onDismiss();

      expect(clearTimeout).toHaveBeenCalled();
      expect(isNotificationActive()).toBe(false);
    });
  });

  describe('isNotificationActive', () => {
    it('should return false when no notification is visible', () => {
      expect(isNotificationActive()).toBe(false);
    });

    it('should return true when notification is visible', () => {
      notifySyncComplete(1);
      expect(isNotificationActive()).toBe(true);
    });

    it('should return false after notification is dismissed', () => {
      notifySyncComplete(1);

      // Get the OK button callback
      const alertCall = (Alert.alert as jest.Mock).mock.calls[0];
      const okButton = alertCall[2][0];

      // Simulate OK button press
      okButton.onPress();

      expect(isNotificationActive()).toBe(false);
    });
  });

  describe('clearNotificationTimeout', () => {
    it('should clear active timeout', () => {
      const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout');

      notifySyncComplete(3);

      clearNotificationTimeout();

      expect(clearTimeoutSpy).toHaveBeenCalled();
      expect(isNotificationActive()).toBe(false);
    });

    it('should be safe to call when no timeout is active', () => {
      expect(() => {
        clearNotificationTimeout();
      }).not.toThrow();
    });
  });
});
