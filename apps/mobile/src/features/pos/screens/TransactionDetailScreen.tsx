/**
 * TransactionDetailScreen Component
 * Displays complete transaction details with receipt reprint option
 * Story 3.7: Transaction Detail View
 * AC4: Transaction Detail View
 * AC6: Receipt Reprint from History
 */

import React, { useState, useEffect } from 'react';
import {
  View,
  StyleSheet,
  ScrollView,
  ActivityIndicator,
  Text,
  SafeAreaView,
  TouchableOpacity,
  Alert,
} from 'react-native';
import { useRoute, useNavigation, RouteProp } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import Icon from 'react-native-vector-icons/MaterialIcons';
import {
  TransactionDetail,
  TransactionItemDetail,
} from '../types/transactionHistory.types';
import { POSStackParamList } from '../types/navigation.types';
import { useTransactionHistoryService } from '../services/TransactionHistoryService';
import { useReceiptPrinter } from '../hooks/useReceiptPrinter';

type TransactionDetailRouteProp = RouteProp<
  POSStackParamList,
  'TransactionDetail'
>;

type TransactionDetailNavigationProp = StackNavigationProp<
  POSStackParamList,
  'TransactionDetail'
>;

/**
 * Format date to Indonesian locale with time
 */
const formatDateTime = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

/**
 * Format currency to Indonesian Rupiah
 */
const formatCurrency = (amount: string): string => {
  const num = parseFloat(amount);
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(num);
};

/**
 * Get status color
 */
const getStatusColor = (status: string): string => {
  switch (status) {
    case 'COMPLETED':
      return '#4CAF50';
    case 'CANCELLED':
      return '#F44336';
    case 'PENDING':
      return '#FF9800';
    default:
      return '#9E9E9E';
  }
};

/**
 * Get status label in Indonesian
 */
const getStatusLabel = (status: string): string => {
  switch (status) {
    case 'COMPLETED':
      return 'Selesai';
    case 'CANCELLED':
      return 'Batal';
    case 'PENDING':
      return 'Tertunda';
    default:
      return status;
  }
};

/**
 * Transaction Item Component
 */
interface TransactionItemRowProps {
  item: TransactionItemDetail;
}

const TransactionItemRow: React.FC<TransactionItemRowProps> = ({ item }) => (
  <View style={styles.itemRow}>
    <View style={styles.itemInfo}>
      <Text style={styles.itemName}>{item.productName}</Text>
      <Text style={styles.itemSku}>SKU: {item.productSKU}</Text>
      <Text style={styles.itemQuantity}>{item.quantity} x {formatCurrency(item.unitPrice)}</Text>
    </View>
    <Text style={styles.itemSubtotal}>{formatCurrency(item.subtotal)}</Text>
  </View>
);

/**
 * Transaction Detail Screen Component
 */
export const TransactionDetailScreen: React.FC = () => {
  const route = useRoute<TransactionDetailRouteProp>();
  const navigation = useNavigation<TransactionDetailNavigationProp>();
  const { getTransactionById } = useTransactionHistoryService();

  // Get transaction ID from route params
  const { transactionId } = route.params as { transactionId: number };

  // State
  const [transaction, setTransaction] = useState<TransactionDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Receipt printer hook
  const {
    isLoading: isPrinting,
    isSuccess: printSuccess,
    error: printError,
    printReceipt,
    clearError,
  } = useReceiptPrinter({
    autoRetry: true,
  });

  /**
   * Fetch transaction detail on mount
   */
  useEffect(() => {
    const fetchTransactionDetail = async () => {
      try {
        setLoading(true);
        setError(null);

        const detail = await getTransactionById(transactionId);
        setTransaction(detail);
      } catch (err: any) {
        setError(err.message || 'Gagal memuat detail transaksi');
      } finally {
        setLoading(false);
      }
    };

    fetchTransactionDetail();
  }, [transactionId]);

  /**
   * Handle receipt reprint
   * AC6: Receipt Reprint from History
   */
  const handleReprintReceipt = async () => {
    if (!transaction?.receiptData) {
      Alert.alert('Error', 'Data struk tidak tersedia');
      return;
    }

    try {
      // Convert receipt data to ESC/POS format
      // TODO: Use the receipt formatting from Story 3.5
      await printReceipt(transaction.receiptData);

      // Log audit trail for reprint action
      // TODO: Implement audit trail logging
      console.log('Receipt reprinted:', transaction.transactionNumber);
    } catch (err: any) {
      Alert.alert('Gagal Mencetak', err.message || 'Terjadi kesalahan saat mencetak struk');
    }
  };

  /**
   * Render loading state
   */
  const renderLoading = () => (
    <View style={styles.centerContainer}>
      <ActivityIndicator size="large" color="#2196F3" />
      <Text style={styles.loadingText}>Memuat detail transaksi...</Text>
    </View>
  );

  /**
   * Render error state
   */
  const renderError = () => (
    <View style={styles.centerContainer}>
      <Icon name="error-outline" size={64} color="#F44336" />
      <Text style={styles.errorTitle}>Gagal memuat transaksi</Text>
      <Text style={styles.errorMessage}>{error}</Text>
      <TouchableOpacity
        style={styles.retryButton}
        onPress={() => navigation.goBack()}
      >
        <Text style={styles.retryButtonText}>Kembali</Text>
      </TouchableOpacity>
    </View>
  );

  /**
   * Render transaction detail
   */
  const renderDetail = () => {
    if (!transaction) return null;

    return (
      <ScrollView style={styles.content}>
        {/* Transaction Header */}
        <View style={styles.section}>
          <View style={styles.transactionNumberContainer}>
            <Text style={styles.transactionNumber}>{transaction.transactionNumber}</Text>
            <View
              style={[
                styles.statusBadge,
                { backgroundColor: getStatusColor(transaction.status) },
              ]}
            >
              <Text style={styles.statusText}>{getStatusLabel(transaction.status)}</Text>
            </View>
          </View>
          <Text style={styles.transactionDate}>{formatDateTime(transaction.createdAt)}</Text>
        </View>

        {/* Items List */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Item</Text>
          {transaction.items.map((item) => (
            <TransactionItemRow key={item.id} item={item} />
          ))}
        </View>

        {/* Payment Information */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Pembayaran</Text>
          <View style={styles.row}>
            <Text style={styles.label}>Metode</Text>
            <Text style={styles.value}>{transaction.paymentMethod}</Text>
          </View>
        </View>

        {/* Total Section */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Total</Text>
          <View style={styles.row}>
            <Text style={styles.label}>Subtotal</Text>
            <Text style={styles.value}>{formatCurrency(transaction.subtotal)}</Text>
          </View>
          {parseFloat(transaction.tax) > 0 && (
            <View style={styles.row}>
              <Text style={styles.label}>Pajak</Text>
              <Text style={styles.value}>{formatCurrency(transaction.tax)}</Text>
            </View>
          )}
          {parseFloat(transaction.discount) > 0 && (
            <View style={styles.row}>
              <Text style={styles.label}>Diskon</Text>
              <Text style={styles.value}>{formatCurrency(transaction.discount)}</Text>
            </View>
          )}
          <View style={[styles.row, styles.totalRow]}>
            <Text style={styles.totalLabel}>Total</Text>
            <Text style={styles.totalValue}>{formatCurrency(transaction.total)}</Text>
          </View>
        </View>

        {/* Cashier and Branch Info */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Informasi</Text>
          <View style={styles.row}>
            <Text style={styles.label}>Kasir</Text>
            <Text style={styles.value}>{transaction.cashier.name || '-'}</Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.label}>Cabang</Text>
            <Text style={styles.value}>{transaction.branch.name || '-'}</Text>
          </View>
        </View>

        {/* Reprint Receipt Button - AC6 */}
        {transaction.status === 'COMPLETED' && (
          <TouchableOpacity
            style={[styles.actionButton, styles.printButton]}
            onPress={handleReprintReceipt}
            disabled={isPrinting}
          >
            <Icon name="print" size={24} color="#FFFFFF" />
            <Text style={styles.actionButtonText}>
              {isPrinting ? 'Mencetak...' : 'Cetak Ulang Struk'}
            </Text>
          </TouchableOpacity>
        )}
      </ScrollView>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Icon name="arrow-back" size={24} color="#FFFFFF" />
        </TouchableOpacity>
        <Text style={styles.headerTitle}>Detail Transaksi</Text>
        <View style={{ width: 24 }} />
      </View>

      {/* Content */}
      {loading ? renderLoading() : error ? renderError() : renderDetail()}
    </SafeAreaView>
  );
};

/**
 * Styles
 */
const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    backgroundColor: '#2196F3',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#FFFFFF',
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
    color: '#757575',
  },
  errorTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#424242',
    marginTop: 16,
  },
  errorMessage: {
    fontSize: 14,
    color: '#757575',
    marginTop: 8,
    textAlign: 'center',
  },
  retryButton: {
    marginTop: 16,
    paddingHorizontal: 24,
    paddingVertical: 12,
    backgroundColor: '#2196F3',
    borderRadius: 8,
  },
  retryButtonText: {
    color: '#FFFFFF',
    fontSize: 14,
    fontWeight: '600',
  },
  content: {
    flex: 1,
  },
  section: {
    backgroundColor: '#FFFFFF',
    marginHorizontal: 16,
    marginVertical: 8,
    padding: 16,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#E0E0E0',
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#212121',
    marginBottom: 12,
  },
  transactionNumberContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  transactionNumber: {
    fontSize: 18,
    fontWeight: '700',
    color: '#212121',
  },
  statusBadge: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
  },
  statusText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFFFFF',
  },
  transactionDate: {
    fontSize: 14,
    color: '#757575',
  },
  itemRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#F0F0F0',
  },
  itemInfo: {
    flex: 1,
  },
  itemName: {
    fontSize: 14,
    fontWeight: '500',
    color: '#212121',
    marginBottom: 4,
  },
  itemSku: {
    fontSize: 12,
    color: '#9E9E9E',
    marginBottom: 2,
  },
  itemQuantity: {
    fontSize: 12,
    color: '#757575',
  },
  itemSubtotal: {
    fontSize: 14,
    fontWeight: '600',
    color: '#212121',
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 4,
  },
  label: {
    fontSize: 14,
    color: '#757575',
  },
  value: {
    fontSize: 14,
    fontWeight: '500',
    color: '#212121',
  },
  totalRow: {
    borderTopWidth: 1,
    borderTopColor: '#E0E0E0',
    marginTop: 8,
    paddingTop: 12,
  },
  totalLabel: {
    fontSize: 16,
    fontWeight: '700',
    color: '#212121',
  },
  totalValue: {
    fontSize: 18,
    fontWeight: '700',
    color: '#2196F3',
  },
  actionButton: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    marginHorizontal: 16,
    marginVertical: 16,
    padding: 16,
    borderRadius: 8,
  },
  printButton: {
    backgroundColor: '#4CAF50',
  },
  actionButtonText: {
    marginLeft: 8,
    fontSize: 16,
    fontWeight: '600',
    color: '#FFFFFF',
  },
});
