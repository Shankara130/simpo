/**
 * Sync Details Modal Component
 * Story 8.4: Implement Visual Sync Status Indicators
 *
 * Modal displaying detailed sync information
 * Shows different content based on sync state (pending, failed, synced)
 * AC2: Sync Details Modal on Tap
 */

import React, { useEffect, useState } from 'react';
import {
  Modal,
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
} from 'react-native';
import { useNetworkStatus } from '../hooks/useNetworkStatus';
import { useBidirectionalSync } from '../hooks/useBidirectionalSync';
import { useSyncProgress } from '../hooks/useSyncProgress';
import { useRetryCountdown, getNextRetryTimestamp } from '../hooks/useRetryCountdown';
import {
  formatTimestamp,
  formatPendingCount,
  formatRetryCountdown,
  translateError,
  getMessage,
} from '../constants/syncMessages';
import SyncOrchestrator from '../services/SyncOrchestrator';
import * as Haptics from 'expo-haptics';

/**
 * Sync Details Modal Props
 */
interface SyncDetailsModalProps {
  visible: boolean;
  onClose: () => void;
}

/**
 * Indonesian sync phase messages
 */
const PHASE_MESSAGES: Record<string, string> = {
  uploading: 'Mengupload transaksi...',
  downloading_stock: 'Download data stok...',
  downloading_products: 'Download produk baru...',
  downloading_user: 'Download data pengguna...',
  synced: 'Semua data ter-sync',
  failed: 'Sync gagal',
};

/**
 * Get Indonesian phase message
 */
function getPhaseMessage(phase: string): string {
  return PHASE_MESSAGES[phase] || 'Sync dalam proses...';
}

/**
 * Sync Details Modal Component
 *
 * Displays sync details with different content for each state:
 * - Pending: transaction count, current phase, progress, offline message
 * - Failed: error message, retry countdown, manual retry button
 * - Synced: success message, last sync timestamp
 *
 * All variants include "Tutup" button to dismiss
 */
export function SyncDetailsModal({ visible, onClose }: SyncDetailsModalProps) {
  const { isConnected } = useNetworkStatus();
  const { status, phase, pendingCount, failedCount, error, lastSyncTime } =
    useBidirectionalSync();
  const { syncedCount } = useSyncProgress();
  const [nextRetryTimestamp, setNextRetryTimestamp] = useState<string | null>(null);

  // Load retry timestamp when modal opens
  useEffect(() => {
    if (visible) {
      getNextRetryTimestamp().then(setNextRetryTimestamp);
    }
  }, [visible]);

  const retryCountdown = useRetryCountdown(nextRetryTimestamp);

  // Haptic feedback on modal open
  useEffect(() => {
    if (visible) {
      Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light).catch(() => {
        // Ignore haptic errors
      });
    }
  }, [visible]);

  const handleClose = () => {
    // Provide haptic feedback on close
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light).catch(() => {
      // Ignore haptic errors
    });
    onClose();
  };

  const handleManualRetry = async () => {
    // Provide haptic feedback on retry
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium).catch(() => {
      // Ignore haptic errors
    });

    // Trigger manual retry via SyncOrchestrator
    await SyncOrchestrator.sync();
  };

  /**
   * Render content based on sync state
   */
  const renderContent = () => {
    // Failed state
    if (status === 'failed' || error) {
      return (
        <>
          <View style={styles.iconContainer}>
            <Text style={styles.errorIcon}>⚠️</Text>
          </View>

          <Text style={styles.title}>Sync Gagal</Text>

          {error && (
            <Text style={styles.errorMessage}>{translateError(error)}</Text>
          )}

          {failedCount > 0 && (
            <Text style={styles.detailText}>
              {failedCount} transaksi gagal
            </Text>
          )}

          {retryCountdown > 0 && (
            <Text style={styles.countdownText}>
              {formatRetryCountdown(retryCountdown)}
            </Text>
          )}

          <TouchableOpacity
            style={styles.retryButton}
            onPress={handleManualRetry}
            activeOpacity={0.7}
          >
            <Text style={styles.retryButtonText}>Sync Sekarang</Text>
          </TouchableOpacity>
        </>
      );
    }

    // Pending/syncing state
    if (status === 'syncing' || pendingCount > 0) {
      return (
        <>
          <View style={styles.iconContainer}>
            <ActivityIndicator size="large" color="#F59E0B" />
          </View>

          <Text style={styles.title}>Sync Sedang Berjalan</Text>

          {!isConnected && (
            <Text style={styles.offlineMessage}>
              Menunggu koneksi internet...
            </Text>
          )}

          {pendingCount > 0 && (
            <Text style={styles.detailText}>
              {formatPendingCount(pendingCount)}
            </Text>
          )}

          {phase && phase !== 'idle' && (
            <Text style={styles.phaseText}>
              {getPhaseMessage(phase)}
            </Text>
          )}

          {syncedCount > 0 && (
            <Text style={styles.progressText}>
              {syncedCount} transaksi berhasil
            </Text>
          )}
        </>
      );
    }

    // Synced state (default)
    return (
      <>
        <View style={styles.iconContainer}>
          <Text style={styles.successIcon}>✓</Text>
        </View>

        <Text style={styles.title}>Semua Data Ter-sync</Text>

        {lastSyncTime && (
          <Text style={styles.timestampText}>
            Terakhir sync: {formatTimestamp(lastSyncTime)}
          </Text>
        )}

        {pendingCount === 0 && (
          <Text style={styles.detailText}>Tidak ada transaksi pending</Text>
        )}
      </>
    );
  };

  return (
    <Modal
      visible={visible}
      transparent={true}
      animationType="fade"
      onRequestClose={handleClose}
    >
      <View style={styles.modalOverlay}>
        <View style={styles.modalContent}>
          {renderContent()}

          <TouchableOpacity
            style={styles.closeButton}
            onPress={handleClose}
            activeOpacity={0.7}
          >
            <Text style={styles.closeButtonText}>Tutup</Text>
          </TouchableOpacity>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  modalContent: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    padding: 24,
    width: '100%',
    maxWidth: 400,
    alignItems: 'center',
  },
  iconContainer: {
    marginBottom: 16,
  },
  errorIcon: {
    fontSize: 48,
  },
  successIcon: {
    fontSize: 48,
    color: '#10B981',
  },
  title: {
    fontSize: 20,
    fontWeight: '600',
    marginBottom: 12,
    textAlign: 'center',
    color: '#1F2937',
  },
  errorMessage: {
    fontSize: 16,
    color: '#EF4444',
    marginBottom: 8,
    textAlign: 'center',
  },
  detailText: {
    fontSize: 14,
    color: '#6B7280',
    marginBottom: 8,
    textAlign: 'center',
  },
  countdownText: {
    fontSize: 14,
    color: '#6B7280',
    marginBottom: 16,
    textAlign: 'center',
  },
  phaseText: {
    fontSize: 14,
    color: '#6B7280',
    marginBottom: 8,
    textAlign: 'center',
  },
  progressText: {
    fontSize: 14,
    color: '#10B981',
    marginBottom: 8,
    textAlign: 'center',
  },
  offlineMessage: {
    fontSize: 14,
    color: '#F59E0B',
    marginBottom: 12,
    textAlign: 'center',
  },
  timestampText: {
    fontSize: 14,
    color: '#6B7280',
    marginBottom: 8,
    textAlign: 'center',
  },
  retryButton: {
    backgroundColor: '#3B82F6',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
    marginBottom: 12,
    minWidth: 150,
    alignItems: 'center',
  },
  retryButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
  closeButton: {
    backgroundColor: '#E5E7EB',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
    minWidth: 150,
    alignItems: 'center',
  },
  closeButtonText: {
    color: '#374151',
    fontSize: 16,
    fontWeight: '500',
  },
});
