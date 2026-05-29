/**
 * Sync Details Modal Component Tests
 * Story 8.4: Implement Visual Sync Status Indicators
 *
 * Tests for modal rendering with different sync states
 */

import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { SyncDetailsModal } from '../SyncDetailsModal';
import * as Haptics from 'expo-haptics';

// Mock expo-haptics
jest.mock('expo-haptics');

// Mock SyncOrchestrator
jest.mock('../../services/SyncOrchestrator', () => ({
  __esModule: true,
  default: {
    sync: jest.fn(),
  },
}));

// Mock hooks with factory functions for different test scenarios
const mockUseNetworkStatus = jest.fn().mockReturnValue({ isConnected: true });
const mockUseBidirectionalSync = jest.fn().mockReturnValue({
  status: 'synced' as const,
  phase: 'synced',
  pendingCount: 0,
  processingCount: 0,
  syncedCount: 10,
  failedCount: 0,
  currentPhase: 'All data synchronized',
  error: undefined,
  lastSyncTime: '2026-05-29T10:30:00Z',
});
const mockUseSyncProgress = jest.fn().mockReturnValue({
  pendingCount: 0,
  processingCount: 0,
  syncedCount: 10,
  failedCount: 0,
  currentTransaction: null,
});
const mockUseRetryCountdown = jest.fn().mockReturnValue(0);

jest.mock('../../hooks/useNetworkStatus', () => ({
  useNetworkStatus: () => mockUseNetworkStatus(),
}));

jest.mock('../../hooks/useBidirectionalSync', () => ({
  useBidirectionalSync: () => mockUseBidirectionalSync(),
}));

jest.mock('../../hooks/useSyncProgress', () => ({
  useSyncProgress: () => mockUseSyncProgress(),
}));

jest.mock('../../hooks/useRetryCountdown', () => ({
  useRetryCountdown: () => mockUseRetryCountdown(),
  getNextRetryTimestamp: jest.fn(),
}));

// Import mocked SyncOrchestrator
import SyncOrchestrator from '../../services/SyncOrchestrator';

describe('SyncDetailsModal', () => {
  const mockOnClose = jest.fn();
  const mockSync = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mock implementations to default (synced state)
    mockUseNetworkStatus.mockReturnValue({ isConnected: true });
    mockUseBidirectionalSync.mockReturnValue({
      status: 'synced' as const,
      phase: 'synced',
      pendingCount: 0,
      processingCount: 0,
      syncedCount: 10,
      failedCount: 0,
      currentPhase: 'All data synchronized',
      error: undefined,
      lastSyncTime: '2026-05-29T10:30:00Z',
    });
    mockUseSyncProgress.mockReturnValue({
      pendingCount: 0,
      processingCount: 0,
      syncedCount: 10,
      failedCount: 0,
      currentTransaction: null,
    });
    mockUseRetryCountdown.mockReturnValue(0);
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

  describe('modal rendering', () => {
    it('should render modal when visible', () => {
      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      // Should show close button
      const closeButton = getByText('Tutup');
      expect(closeButton).toBeTruthy();
    });

    it('should not render modal when not visible', () => {
      const { queryByText } = render(
        <SyncDetailsModal visible={false} onClose={mockOnClose} />
      );

      const closeButton = queryByText('Tutup');
      expect(closeButton).toBeNull();
    });
  });

  describe('synced state content', () => {
    it('should display success message when synced', () => {
      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const successMessage = getByText('Semua data ter-sync');
      expect(successMessage).toBeTruthy();
    });

    it('should display last sync timestamp when synced', () => {
      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const timestamp = getByText(/Terakhir sync:/);
      expect(timestamp).toBeTruthy();
    });

    it('should show no pending transactions when synced', () => {
      const { queryByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const pendingText = queryByText(/transaksi pending/);
      expect(pendingText).toBeNull();
    });
  });

  describe('pending state content', () => {
    it('should display transaction count when pending', () => {
      // Mock pending state
      mockUseBidirectionalSync.mockReturnValue({
        status: 'syncing' as const,
        phase: 'uploading',
        pendingCount: 5,
        processingCount: 1,
        syncedCount: 0,
        failedCount: 0,
        currentPhase: 'Uploading pending transactions',
        error: undefined,
        lastSyncTime: '2026-05-29T10:30:00Z',
      });
      mockUseSyncProgress.mockReturnValue({
        pendingCount: 5,
        processingCount: 1,
        syncedCount: 0,
        failedCount: 0,
        currentTransaction: null,
      });

      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const pendingText = getByText(/5 transaksi pending/);
      expect(pendingText).toBeTruthy();
    });

    it('should display current phase when syncing', () => {
      mockUseBidirectionalSync.mockReturnValue({
        status: 'syncing' as const,
        phase: 'uploading',
        pendingCount: 5,
        processingCount: 1,
        syncedCount: 0,
        failedCount: 0,
        currentPhase: 'Uploading pending transactions',
        error: undefined,
        lastSyncTime: '2026-05-29T10:30:00Z',
      });

      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const phaseText = getByText(/Mengupload transaksi/);
      expect(phaseText).toBeTruthy();
    });

    it('should display offline message when offline', () => {
      mockUseNetworkStatus.mockReturnValue({ isConnected: false });

      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const offlineMessage = getByText('Menunggu koneksi internet...');
      expect(offlineMessage).toBeTruthy();
    });
  });

  describe('failed state content', () => {
    it('should display error message when failed', () => {
      mockUseBidirectionalSync.mockReturnValue({
        status: 'failed' as const,
        phase: 'failed',
        pendingCount: 0,
        processingCount: 0,
        syncedCount: 0,
        failedCount: 3,
        currentPhase: 'Sync failed - will retry',
        error: 'Network error',
        lastSyncTime: '2026-05-29T10:30:00Z',
      });

      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const errorMessage = getByText('Error jaringan');
      expect(errorMessage).toBeTruthy();
    });

    it('should display retry countdown when failed', () => {
      mockUseBidirectionalSync.mockReturnValue({
        status: 'failed' as const,
        phase: 'failed',
        pendingCount: 0,
        processingCount: 0,
        syncedCount: 0,
        failedCount: 3,
        currentPhase: 'Sync failed - will retry',
        error: 'Network error',
        lastSyncTime: '2026-05-29T10:30:00Z',
      });
      mockUseRetryCountdown.mockReturnValue(5);

      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const countdownText = getByText(/Retry otomatis dalam 5 menit/);
      expect(countdownText).toBeTruthy();
    });

    it('should display manual retry button when failed', () => {
      mockUseBidirectionalSync.mockReturnValue({
        status: 'failed' as const,
        phase: 'failed',
        pendingCount: 0,
        processingCount: 0,
        syncedCount: 0,
        failedCount: 3,
        currentPhase: 'Sync failed - will retry',
        error: 'Network error',
        lastSyncTime: '2026-05-29T10:30:00Z',
      });

      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const retryButton = getByText('Sync Sekarang');
      expect(retryButton).toBeTruthy();
    });
  });

  describe('user interactions', () => {
    it('should call onClose when close button pressed', () => {
      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const closeButton = getByText('Tutup');
      fireEvent.press(closeButton);

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should trigger manual retry when retry button pressed', () => {
      mockUseBidirectionalSync.mockReturnValue({
        status: 'failed' as const,
        phase: 'failed',
        pendingCount: 0,
        processingCount: 0,
        syncedCount: 0,
        failedCount: 3,
        currentPhase: 'Sync failed - will retry',
        error: 'Network error',
        lastSyncTime: '2026-05-29T10:30:00Z',
      });

      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const retryButton = getByText('Sync Sekarang');
      fireEvent.press(retryButton);

      expect(SyncOrchestrator.sync).toHaveBeenCalled();
    });

    it('should provide haptic feedback on button press', () => {
      const mockImpactAsync = jest.spyOn(Haptics, 'impactAsync');

      const { getByText } = render(
        <SyncDetailsModal visible={true} onClose={mockOnClose} />
      );

      const closeButton = getByText('Tutup');
      fireEvent.press(closeButton);

      expect(mockImpactAsync).toHaveBeenCalled();
    });
  });
});
