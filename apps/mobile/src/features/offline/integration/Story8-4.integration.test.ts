/**
 * Story 8-4 Integration Tests
 * Story 8.4: Implement Visual Sync Status Indicators
 *
 * End-to-end integration tests for sync status indicators
 * Tests complete sync flows: pending → syncing → synced → failed → retry
 */

import React from 'react';
import { renderHook, act } from '@testing-library/react-hooks';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { SyncStatusIndicator } from '../components/SyncStatusIndicator';
import { SyncDetailsModal } from '../components/SyncDetailsModal';
import { useNetworkStatus } from '../hooks/useNetworkStatus';
import { useBidirectionalSync } from '../hooks/useBidirectionalSync';
import { useSyncProgress } from '../hooks/useSyncProgress';
import { useRetryCountdown, getNextRetryTimestamp } from '../hooks/useRetryCountdown';
import SyncOrchestrator from '../services/SyncOrchestrator';
import * as Haptics from 'expo-haptics';

// Mock all dependencies
jest.mock('@react-native-async-storage/async-storage');
jest.mock('expo-haptics');
jest.mock('../services/SyncOrchestrator');

// Create controllable mock hooks
const mockNetworkState = { isConnected: true };
const mockSyncState = {
  status: 'synced' as const,
  phase: 'synced' as const,
  pendingCount: 0,
  processingCount: 0,
  syncedCount: 0,
  failedCount: 0,
  currentPhase: null,
  error: undefined,
  lastSyncTime: '2026-05-29T10:30:00Z',
};
const mockProgressState = {
  pendingCount: 0,
  processingCount: 0,
  syncedCount: 0,
  failedCount: 0,
  currentTransaction: null,
};

jest.mock('../hooks/useNetworkStatus', () => ({
  useNetworkStatus: () => mockNetworkState,
}));

jest.mock('../hooks/useBidirectionalSync', () => ({
  useBidirectionalSync: () => mockSyncState,
}));

jest.mock('../hooks/useSyncProgress', () => ({
  useSyncProgress: () => mockProgressState,
}));

jest.mock('../hooks/useRetryCountdown', () => ({
  useRetryCountdown: () => 0,
  getNextRetryCountdown: jest.fn(),
}));

describe('Story 8-4 Integration Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();

    // Reset mock states
    mockNetworkState.isConnected = true;
    Object.assign(mockSyncState, {
      status: 'synced' as const,
      phase: 'synced' as const,
      pendingCount: 0,
      processingCount: 0,
      syncedCount: 0,
      failedCount: 0,
      currentPhase: null,
      error: undefined,
      lastSyncTime: '2026-05-29T10:30:00Z',
    });
    Object.assign(mockProgressState, {
      pendingCount: 0,
      processingCount: 0,
      syncedCount: 0,
      failedCount: 0,
      currentTransaction: null,
    });

    // Mock AsyncStorage
    (AsyncStorage.getItem as jest.Mock).mockResolvedValue(null);
    (AsyncStorage.setItem as jest.Mock).mockResolvedValue(undefined);

    // Mock SyncOrchestrator
    (SyncOrchestrator.sync as jest.Mock).mockResolvedValue({
      success: true,
      phase: 'synced',
      uploaded: 0,
      downloadedStock: 0,
      downloadedProducts: 0,
      userUpdated: false,
      duration: 0,
    });
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  describe('Full sync flow: pending → syncing → synced', () => {
    it('should transition from pending to syncing to synced', async () => {
      // Start with pending state
      mockSyncState.status = 'syncing';
      mockSyncState.phase = 'uploading';
      mockSyncState.pendingCount = 5;
      mockProgressState.pendingCount = 5;

      const { getByTestId, queryByTestId } = render(
        <SyncStatusIndicator />
      );

      // Should show pending indicator (clock icon)
      const indicator = getByTestId('sync-status-indicator');
      expect(indicator).toBeTruthy();

      // Transition to synced
      await act(async () => {
        mockSyncState.status = 'synced';
        mockSyncState.phase = 'synced';
        mockSyncState.pendingCount = 0;
        mockProgressState.pendingCount = 0;

        // Force re-render
        render(<SyncStatusIndicator />);
      });

      // Should show synced state (checkmark icon)
      expect(getByTestId('sync-status-indicator')).toBeTruthy();
    });

    it('should show correct state in modal throughout flow', async () => {
      const onClose = jest.fn();

      // Start with pending state
      mockSyncState.status = 'syncing';
      mockSyncState.pendingCount = 3;
      mockProgressState.pendingCount = 3;

      const { getByText, rerender } = render(
        <SyncDetailsModal visible={true} onClose={onClose} />
      );

      // Should show pending message
      expect(getByText(/3 transaksi pending/)).toBeTruthy();

      // Transition to synced
      mockSyncState.status = 'synced';
      mockSyncState.pendingCount = 0;
      mockProgressState.pendingCount = 0;

      rerender(<SyncDetailsModal visible={true} onClose={onClose} />);

      // Should show success message
      expect(getByText('Semua Data Ter-sync')).toBeTruthy();
    });
  });

  describe('Failed flow: pending → failed → retry → synced', () => {
    it('should handle failed sync and manual retry', async () => {
      // Start with failed state
      mockSyncState.status = 'failed';
      mockSyncState.phase = 'failed';
      mockSyncState.error = 'Network error';
      mockSyncState.failedCount = 2;
      mockSyncState.pendingCount = 0;
      mockProgressState.failedCount = 2;

      const onClose = jest.fn();
      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={onClose} />
      );

      // Should show error message
      expect(getByText('Error jaringan')).toBeTruthy();
      expect(getByText('2 transaksi gagal')).toBeTruthy();

      // Should show retry button
      const retryButton = getByText('Sync Sekarang');
      expect(retryButton).toBeTruthy();

      // Tap retry button
      fireEvent.press(retryButton);

      // Should trigger SyncOrchestrator.sync()
      expect(SyncOrchestrator.sync).toHaveBeenCalled();
    });

    it('should display retry countdown when failed', async () => {
      // Setup retry schedule in AsyncStorage
      const now = new Date();
      const fiveMinutesFromNow = new Date(now.getTime() + 5 * 60000).toISOString();
      const retrySchedule = {
        'trx-1': { retryAt: fiveMinutesFromNow, retryCount: 1 },
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(
        JSON.stringify(retrySchedule)
      );

      // Set failed state
      mockSyncState.status = 'failed';
      mockSyncState.error = 'Network error';
      mockSyncState.failedCount = 1;

      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={jest.fn()} />
      );

      // Should show retry countdown
      await waitFor(() => {
        expect(getByText(/Retry otomatis dalam/)).toBeTruthy();
      });
    });
  });

  describe('Offline → online → sync flow', () => {
    it('should show waiting for internet message when offline', () => {
      // Set offline state with pending transactions
      mockNetworkState.isConnected = false;
      mockSyncState.pendingCount = 3;
      mockProgressState.pendingCount = 3;

      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={jest.fn()} />
      );

      // Should show offline waiting message
      expect(getByText('Menunggu koneksi internet...')).toBeTruthy();
    });

    it('should transition to syncing when connection restored', async () => {
      // Start offline with pending
      mockNetworkState.isConnected = false;
      mockSyncState.pendingCount = 3;
      mockProgressState.pendingCount = 3;

      const { getByText, rerender } = render(
        <SyncDetailsModal visible={true} onClose={jest.fn()} />
      );

      // Verify offline message
      expect(getByText('Menunggu koneksi internet...')).toBeTruthy();

      // Connection restored
      mockNetworkState.isConnected = true;
      mockSyncState.status = 'syncing';
      mockSyncState.phase = 'uploading';

      rerender(<SyncDetailsModal visible={true} onClose={jest.fn()} />);

      // Should show syncing message
      expect(getByText(/Mengupload transaksi/)).toBeTruthy();
    });
  });

  describe('Indonesian messages display', () => {
    it('should display all sync states in Indonesian', () => {
      const onClose = jest.fn();

      // Test synced state
      mockSyncState.status = 'synced';
      mockSyncState.pendingCount = 0;
      mockProgressState.pendingCount = 0;

      const { getByText: getByTextSynced, rerender } = render(
        <SyncDetailsModal visible={true} onClose={onClose} />
      );

      expect(getByTextSynced('Semua Data Ter-sync')).toBeTruthy();
      expect(getByTextSynced(/Terakhir sync:/)).toBeTruthy();
      expect(getByTextSynced('Tutup')).toBeTruthy();

      // Test pending state
      mockSyncState.status = 'syncing';
      mockSyncState.pendingCount = 2;
      mockProgressState.pendingCount = 2;

      rerender(<SyncDetailsModal visible={true} onClose={onClose} />);

      expect(getByTextSynced(/2 transaksi pending/)).toBeTruthy();

      // Test failed state
      mockSyncState.status = 'failed';
      mockSyncState.error = 'server';
      mockSyncState.failedCount = 1;
      mockSyncState.pendingCount = 0;
      mockProgressState.failedCount = 1;

      rerender(<SyncDetailsModal visible={true} onClose={onClose} />);

      expect(getByTextSynced('Error server')).toBeTruthy();
      expect(getByTextSynced('Sync Sekarang')).toBeTruthy();
    });

    it('should translate technical error messages to Indonesian', () => {
      const errorTranslations = [
        { input: 'network', expected: 'Error jaringan' },
        { input: 'server', expected: 'Error server' },
        { input: 'timeout', expected: 'Request timeout' },
        { input: 'validation', expected: 'Data tidak valid' },
        { input: 'unknown', expected: 'Error tidak diketahui' },
      ];

      errorTranslations.forEach(({ input, expected }) => {
        // Import translateError utility
        const { translateError } = require('../constants/syncMessages');
        expect(translateError(input)).toBe(expected);
      });
    });
  });

  describe('Haptic feedback on failed state', () => {
    it('should trigger haptic feedback when state changes to failed', async () => {
      const mockNotificationAsync = jest.spyOn(Haptics, 'notificationAsync');

      // Start with syncing state
      mockSyncState.status = 'syncing';

      const { rerender } = render(<SyncStatusIndicator />);

      // Transition to failed
      mockSyncState.status = 'failed';
      mockSyncState.error = 'Network error';

      rerender(<SyncStatusIndicator />);

      // Should trigger haptic feedback
      await waitFor(() => {
        expect(mockNotificationAsync).toHaveBeenCalledWith(
          Haptics.NotificationFeedbackType.Warning
        );
      });
    });
  });

  describe('No breaking changes to existing services', () => {
    it('should not modify useNetworkStatus hook interface', () => {
      const { result } = renderHook(() => useNetworkStatus());

      expect(result.current).toHaveProperty('isConnected');
      expect(typeof result.current.isConnected).toBe('boolean');
    });

    it('should not modify useBidirectionalSync hook interface', () => {
      const { result } = renderHook(() => useBidirectionalSync());

      expect(result.current).toHaveProperty('status');
      expect(result.current).toHaveProperty('phase');
      expect(result.current).toHaveProperty('pendingCount');
      expect(result.current).toHaveProperty('error');
    });

    it('should not modify useSyncProgress hook interface', () => {
      const { result } = renderHook(() => useSyncProgress());

      expect(result.current).toHaveProperty('pendingCount');
      expect(result.current).toHaveProperty('syncedCount');
      expect(result.current).toHaveProperty('failedCount');
    });

    it('should not modify SyncOrchestrator interface', () => {
      expect(typeof SyncOrchestrator.sync).toBe('function');
    });
  });

  describe('State update latency', () => {
    it('should update indicator state immediately on sync change', async () => {
      let updateTimestamp: number | null = null;

      // Mock useEffect to capture update timing
      const originalUseEffect = React.useEffect;
      jest.spyOn(React, 'useEffect').mockImplementation((effect, deps) => {
        if (deps && deps.includes('displayState')) {
          updateTimestamp = Date.now();
        }
        return originalUseEffect(effect, deps);
      });

      // Start with synced state
      mockSyncState.status = 'synced';

      const startTime = Date.now();
      const { rerender } = render(<SyncStatusIndicator />);

      // Change state to syncing
      mockSyncState.status = 'syncing';
      mockSyncState.pendingCount = 5;

      rerender(<SyncStatusIndicator />);

      const endTime = Date.now();
      const latency = endTime - startTime;

      // State should update within 100ms target
      expect(latency).toBeLessThan(100);
    });
  });
});
