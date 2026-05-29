/**
 * ConflictResolutionScreen - UI for managing failed transactions with conflicts
 * Story 8.5: Implement Conflict Resolution for Offline Transactions
 * Task 8: Create Mobile Conflict Resolution UI
 * AC8: Display failed transactions, show conflict details, provide override/delete actions
 *
 * Indonesian language for all user-facing messages per Story 8.5 requirements
 */

import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
  Modal,
  TextInput,
  ScrollView,
} from 'react-native';
import { useNavigation, useRoute } from '@react-navigation/native';
import ConflictResolutionService from '../services/ConflictResolutionService';
import {
  ConflictErrorResponse,
  FailedTransactionInfo,
} from '../types/bidirectional-sync.types';

/**
 * ConflictResolutionScreen Props
 */
interface ConflictResolutionScreenProps {
  isAdmin?: boolean; // Whether current user is admin (for override button)
  currentUserId?: number; // Current user ID for override authorization
}

/**
 * Main Conflict Resolution Screen Component
 */
const ConflictResolutionScreen: React.FC<ConflictResolutionScreenProps> = ({
  isAdmin = false,
  currentUserId,
}) => {
  const navigation = useNavigation();
  const route = useRoute();

  const [failedTransactions, setFailedTransactions] = useState<FailedTransactionInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  // Admin authorization modal state
  const [adminModalVisible, setAdminModalVisible] = useState(false);
  const [selectedTransactionId, setSelectedTransactionId] = useState<number | null>(null);
  const [adminUsername, setAdminUsername] = useState('');
  const [adminPassword, setAdminPassword] = useState('');
  const [overrideReason, setOverrideReason] = useState('');
  const [processingOverride, setProcessingOverride] = useState(false);

  // Load failed transactions on mount
  useEffect(() => {
    loadFailedTransactions();
  }, []);

  /**
   * Load all failed transactions from ConflictResolutionService
   */
  const loadFailedTransactions = async () => {
    try {
      setLoading(true);
      const conflicts = await ConflictResolutionService.getAllFailedTransactions();
      setFailedTransactions(conflicts);
    } catch (error) {
      console.error('[ConflictResolutionScreen] Failed to load conflicts:', error);
      Alert.alert('Error', 'Gagal memuat data konflik');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Refresh failed transactions list
   */
  const handleRefresh = async () => {
    setRefreshing(true);
    await loadFailedTransactions();
    setRefreshing(false);
  };

  /**
   * Show admin authorization modal for override
   */
  const showOverrideModal = (transactionId: number) => {
    setSelectedTransactionId(transactionId);
    setAdminModalVisible(true);
  };

  /**
   * Submit manual override request
   */
  const handleOverrideSubmit = async () => {
    if (selectedTransactionId === null) return;

    // Validate inputs
    if (!adminUsername.trim()) {
      Alert.alert('Error', 'Username admin diperlukan');
      return;
    }

    if (!overrideReason.trim()) {
      Alert.alert('Error', 'Alasan override diperlukan');
      return;
    }

    try {
      setProcessingOverride(true);

      const result = await ConflictResolutionService.requestManualOverride({
        transactionId: selectedTransactionId,
        adminUserId: currentUserId || 0,
        reason: overrideReason,
        forceProcessing: true,
      });

      if (result.success) {
        Alert.alert(
          'Berhasil',
          `Transaksi ${result.transactionNumber} berhasil di-override`,
          [
            {
              text: 'OK',
              onPress: () => {
                setAdminModalVisible(false);
                loadFailedTransactions(); // Refresh list
              },
            },
          ]
        );
      } else {
        Alert.alert('Gagal', result.message || 'Override gagal');
      }
    } catch (error) {
      console.error('[ConflictResolutionScreen] Override failed:', error);
      Alert.alert('Error', 'Terjadi kesalahan saat override');
    } finally {
      setProcessingOverride(false);
    }
  };

  /**
   * Delete failed transaction
   */
  const handleDeleteTransaction = async (transactionId: number, transactionNumber: string) => {
    Alert.alert(
      'Hapus Transaksi',
      `Apakah Anda yakin ingin menghapus transaksi ${transactionNumber}? Tindakan ini tidak dapat dibatalkan.`,
      [
        {
          text: 'Batal',
          style: 'cancel',
        },
        {
          text: 'Hapus',
          style: 'destructive',
          onPress: async () => {
            try {
              await ConflictResolutionService.clearFailedTransaction(transactionId);
              Alert.alert('Berhasil', 'Transaksi berhasil dihapus');
              loadFailedTransactions(); // Refresh list
            } catch (error) {
              console.error('[ConflictResolutionScreen] Delete failed:', error);
              Alert.alert('Error', 'Gagal menghapus transaksi');
            }
          },
        },
      ]
    );
  };

  /**
   * Render individual failed transaction item
   */
  const renderTransactionItem = ({ item }: { item: FailedTransactionInfo }) => (
    <View style={styles.transactionCard}>
      <View style={styles.transactionHeader}>
        <Text style={styles.transactionNumber}>{item.transactionNumber}</Text>
        <Text style={styles.transactionTimestamp}>
          {new Date(item.timestamp).toLocaleString('id-ID')}
        </Text>
      </View>

      {/* Conflict Details */}
      <View style={styles.conflictDetails}>
        <Text style={styles.errorTitle}>{item.conflictError.title || 'Konflik Sinkronisasi'}</Text>
        <Text style={styles.errorMessage}>{item.conflictError.detail || 'Terjadi konflik saat sinkronisasi'}</Text>

        {item.conflictError.conflict_details && (
          <View style={styles.stockInfo}>
            <Text style={styles.stockLabel}>Produk: {item.conflictError.conflict_details.product_sku || 'N/A'}</Text>
            <Text style={styles.stockLabel}>
              Diminta: {item.conflictError.conflict_details.requested_qty || 0}
            </Text>
            <Text style={styles.stockLabel}>
              Tersedia: {item.conflictError.conflict_details.available_stock || 0}
            </Text>
            {item.conflictError.conflict_details.shortfall > 0 && (
              <Text style={styles.shortfallLabel}>
                Kekurangan: {item.conflictError.conflict_details.shortfall}
              </Text>
            )}
          </View>
        )}
      </View>

      {/* Action Buttons */}
      <View style={styles.actionButtons}>
        {item.canOverride && isAdmin && (
          <TouchableOpacity
            style={[styles.button, styles.overrideButton]}
            onPress={() => showOverrideModal(item.transactionId)}
          >
            <Text style={styles.buttonText}>Override dengan Admin</Text>
          </TouchableOpacity>
        )}

        <TouchableOpacity
          style={[styles.button, styles.deleteButton]}
          onPress={() => handleDeleteTransaction(item.transactionId, item.transactionNumber)}
        >
          <Text style={styles.buttonText}>Hapus Transaksi</Text>
        </TouchableOpacity>
      </View>
    </View>
  );

  /**
   * Render empty state
   */
  const renderEmptyState = () => (
    <View style={styles.emptyState}>
      <Text style={styles.emptyStateIcon}>✓</Text>
      <Text style={styles.emptyStateTitle}>Tidak Ada Konflik</Text>
      <Text style={styles.emptyStateMessage}>Semua transaksi sinkronisasi berhasil</Text>
    </View>
  );

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#2196F3" />
        <Text style={styles.loadingText}>Memuat konflik...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Konflik Sinkronisasi</Text>
        <Text style={styles.headerSubtitle}>
          {failedTransactions.length} Transaksi Gagal
        </Text>
      </View>

      {/* Failed Transactions List */}
      {failedTransactions.length === 0 ? (
        renderEmptyState()
      ) : (
        <FlatList
          data={failedTransactions}
          keyExtractor={(item) => item.transactionId.toString()}
          renderItem={renderTransactionItem}
          contentContainerStyle={styles.listContent}
          onRefresh={handleRefresh}
          refreshing={refreshing}
        />
      )}

      {/* Admin Authorization Modal */}
      <Modal
        visible={adminModalVisible}
        animationType="slide"
        transparent={true}
        onRequestClose={() => setAdminModalVisible(false)}
      >
        <View style={styles.modalContainer}>
          <View style={styles.modalContent}>
            <Text style={styles.modalTitle}>Autorisasi Admin Diperlukan</Text>

            <View style={styles.modalForm}>
              <TextInput
                style={styles.input}
                placeholder="Username Admin"
                value={adminUsername}
                onChangeText={setAdminUsername}
                autoCapitalize="none"
              />

              <TextInput
                style={styles.input}
                placeholder="Password"
                value={adminPassword}
                onChangeText={setAdminPassword}
                secureTextEntry
              />

              <TextInput
                style={[styles.input, styles.textArea]}
                placeholder="Alasan Override"
                value={overrideReason}
                onChangeText={setOverrideReason}
                multiline
                numberOfLines={3}
              />
            </View>

            <View style={styles.modalButtons}>
              <TouchableOpacity
                style={[styles.modalButton, styles.cancelButton]}
                onPress={() => setAdminModalVisible(false)}
                disabled={processingOverride}
              >
                <Text style={styles.modalButtonText}>Batal</Text>
              </TouchableOpacity>

              <TouchableOpacity
                style={[styles.modalButton, styles.confirmButton]}
                onPress={handleOverrideSubmit}
                disabled={processingOverride}
              >
                {processingOverride ? (
                  <ActivityIndicator color="#FFF" />
                ) : (
                  <Text style={styles.modalButtonText}>Override</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  header: {
    backgroundColor: '#2196F3',
    padding: 16,
    paddingTop: 60,
  },
  headerTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#FFFFFF',
  },
  headerSubtitle: {
    fontSize: 14,
    color: '#FFFFFF',
    marginTop: 4,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 12,
    fontSize: 16,
    color: '#666',
  },
  listContent: {
    padding: 16,
  },
  transactionCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    padding: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  transactionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 12,
  },
  transactionNumber: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#333',
  },
  transactionTimestamp: {
    fontSize: 12,
    color: '#999',
  },
  conflictDetails: {
    marginBottom: 12,
  },
  errorTitle: {
    fontSize: 14,
    fontWeight: 'bold',
    color: '#D32F2F',
    marginBottom: 4,
  },
  errorMessage: {
    fontSize: 14,
    color: '#666',
    marginBottom: 8,
  },
  stockInfo: {
    backgroundColor: '#FFF3E0',
    padding: 12,
    borderRadius: 4,
  },
  stockLabel: {
    fontSize: 13,
    color: '#333',
    marginBottom: 2,
  },
  shortfallLabel: {
    fontSize: 13,
    fontWeight: 'bold',
    color: '#D32F2F',
    marginTop: 4,
  },
  actionButtons: {
    flexDirection: 'row',
    gap: 8,
  },
  button: {
    flex: 1,
    paddingVertical: 10,
    paddingHorizontal: 16,
    borderRadius: 6,
    alignItems: 'center',
  },
  buttonText: {
    fontSize: 13,
    fontWeight: '600',
    color: '#FFFFFF',
  },
  overrideButton: {
    backgroundColor: '#FF9800',
  },
  deleteButton: {
    backgroundColor: '#F44336',
  },
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  emptyStateIcon: {
    fontSize: 64,
    marginBottom: 16,
  },
  emptyStateTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#4CAF50',
    marginBottom: 8,
  },
  emptyStateMessage: {
    fontSize: 14,
    color: '#999',
    textAlign: 'center',
  },
  modalContainer: {
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
  },
  modalTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 20,
  },
  modalForm: {
    marginBottom: 20,
  },
  input: {
    backgroundColor: '#F5F5F5',
    borderRadius: 8,
    padding: 12,
    fontSize: 14,
    marginBottom: 12,
  },
  textArea: {
    height: 80,
    textAlignVertical: 'top',
  },
  modalButtons: {
    flexDirection: 'row',
    gap: 12,
  },
  modalButton: {
    flex: 1,
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
    alignItems: 'center',
  },
  cancelButton: {
    backgroundColor: '#999',
  },
  confirmButton: {
    backgroundColor: '#2196F3',
  },
  modalButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFFFFF',
  },
});

export default ConflictResolutionScreen;
