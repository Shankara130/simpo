/**
 * TransactionHistoryScreen Component
 * Displays transaction history with filters and pagination
 * Story 3.7: Transaction History View
 * AC3: Mobile Transaction History Screen
 */

import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  StyleSheet,
  FlatList,
  RefreshControl,
  TouchableOpacity,
  ActivityIndicator,
  Text,
  SafeAreaView,
} from 'react-native';
import { useNavigation, useFocusEffect } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import Icon from 'react-native-vector-icons/MaterialIcons';
import {
  TransactionSummary,
  TransactionFilters,
  DateRangePreset,
  TransactionHistoryState,
} from '../types/transactionHistory.types';
import { POSStackParamList } from '../types/navigation.types';
import { useTransactionHistoryService } from '../services/TransactionHistoryService';
import { TransactionFilterModal } from '../components/TransactionFilterModal';

type TransactionHistoryNavigationProp = StackNavigationProp<
  POSStackParamList,
  'TransactionHistory'
>;

/**
 * Format date to Indonesian locale
 */
const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
};

/**
 * Format time to HH:MM format
 */
const formatTime = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleTimeString('id-ID', {
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
 * Transaction List Item Component
 */
interface TransactionListItemProps {
  transaction: TransactionSummary;
  onPress: (transaction: TransactionSummary) => void;
}

const TransactionListItem: React.FC<TransactionListItemProps> = ({
  transaction,
  onPress,
}) => (
  <TouchableOpacity
    style={styles.transactionItem}
    onPress={() => onPress(transaction)}
  >
    <View style={styles.transactionHeader}>
      <Text style={styles.transactionNumber}>{transaction.transactionNumber}</Text>
      <View style={[styles.statusBadge, { backgroundColor: getStatusColor(transaction.status) }]}>
        <Text style={styles.statusText}>{getStatusLabel(transaction.status)}</Text>
      </View>
    </View>
    <View style={styles.transactionDetails}>
      <Text style={styles.transactionTotal}>{formatCurrency(transaction.total)}</Text>
      <Text style={styles.transactionTime}>{formatTime(transaction.createdAt)}</Text>
    </View>
    <Text style={styles.transactionDate}>{formatDate(transaction.createdAt)}</Text>
  </TouchableOpacity>
);

/**
 * Transaction History Screen Component
 */
export const TransactionHistoryScreen: React.FC = () => {
  const navigation = useNavigation<TransactionHistoryNavigationProp>();
  const { getTransactions, loadFilters, saveFilters } = useTransactionHistoryService();

  // State management
  const [state, setState] = useState<TransactionHistoryState>({
    transactions: [],
    loading: true,
    error: null,
    filters: {
      startDate: new Date(), // Default to today
      endDate: null,
      status: 'ALL',
    },
    pagination: {
      currentPage: 1,
      totalPages: 1,
      hasMore: false,
    },
    refreshing: false,
  });

  // Filter modal state
  const [filterModalVisible, setFilterModalVisible] = useState(false);

  /**
   * Load filters from AsyncStorage on screen focus
   */
  useFocusEffect(
    useCallback(() => {
      const loadSavedFilters = async () => {
        try {
          const savedFilters = await loadFilters();
          setState((prev) => ({
            ...prev,
            filters: savedFilters,
          }));
          // Fetch transactions with saved filters
          await fetchTransactions(1, savedFilters);
        } catch (error) {
          console.error('Failed to load filters:', error);
        }
      };
      loadSavedFilters();
    }, [])
  );

  /**
   * Fetch transactions with filters and pagination
   */
  const fetchTransactions = async (
    page: number,
    filters?: TransactionFilters,
    append = false
  ) => {
    try {
      setState((prev) => ({
        ...prev,
        loading: !append,
        error: null,
      }));

      const response = await getTransactions(
        filters || state.filters,
        page,
        20 // limit
      );

      setState((prev) => ({
        ...prev,
        transactions: append
          ? [...prev.transactions, ...response.data]
          : response.data,
        pagination: {
          currentPage: response.pagination.currentPage,
          totalPages: response.pagination.totalPages,
          hasMore: response.pagination.currentPage < response.pagination.totalPages,
        },
        loading: false,
        refreshing: false,
      }));
    } catch (error: any) {
      setState((prev) => ({
        ...prev,
        error: error.message || 'Gagal memuat transaksi',
        loading: false,
        refreshing: false,
      }));
    }
  };

  /**
   * Handle pull-to-refresh
   */
  const handleRefresh = async () => {
    setState((prev) => ({ ...prev, refreshing: true }));
    await fetchTransactions(1, state.filters);
  };

  /**
   * Handle infinite scroll (load more)
   */
  const handleLoadMore = () => {
    if (state.pagination.hasMore && !state.loading) {
      fetchTransactions(state.pagination.currentPage + 1, state.filters, true);
    }
  };

  /**
   * Handle filter button press
   */
  const handleFilterPress = () => {
    setFilterModalVisible(true);
  };

  /**
   * Handle apply filters
   */
  const handleApplyFilters = async (newFilters: TransactionFilters) => {
    setState((prev) => ({ ...prev, filters: newFilters }));
    await saveFilters(newFilters);
    await fetchTransactions(1, newFilters);
  };

  /**
   * Handle transaction item press
   */
  const handleTransactionPress = (transaction: TransactionSummary) => {
    navigation.navigate('TransactionDetail' as any, { transactionId: transaction.id });
  };

  /**
   * Render list footer (loading indicator)
   */
  const renderListFooter = () => {
    if (!state.loading) return null;
    return (
      <View style={styles.listFooter}>
        <ActivityIndicator size="small" color="#2196F3" />
      </View>
    );
  };

  /**
   * Render empty state
   */
  const renderEmptyState = () => {
    if (state.loading) return null;

    return (
      <View style={styles.emptyState}>
        <Icon name="receipt-long" size={64} color="#9E9E9E" />
        <Text style={styles.emptyStateTitle}>Tidak ada transaksi</Text>
        <Text style={styles.emptyStateMessage}>
          Belum ada transaksi untuk periode yang dipilih
        </Text>
      </View>
    );
  };

  /**
   * Render error state
   */
  const renderErrorState = () => {
    if (!state.error) return null;

    return (
      <View style={styles.errorState}>
        <Icon name="error-outline" size={48} color="#F44336" />
        <Text style={styles.errorTitle}>Gagal memuat transaksi</Text>
        <Text style={styles.errorMessage}>{state.error}</Text>
        <TouchableOpacity
          style={styles.retryButton}
          onPress={() => fetchTransactions(1, state.filters)}
        >
          <Text style={styles.retryButtonText}>Coba Lagi</Text>
        </TouchableOpacity>
      </View>
    );
  };

  /**
   * Render header
   */
  const renderHeader = () => {
    const hasActiveFilters =
      state.filters.startDate !== null ||
      state.filters.endDate !== null ||
      state.filters.status !== 'ALL';

    return (
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Riwayat Transaksi</Text>
        <TouchableOpacity
          style={[styles.filterButton, hasActiveFilters && styles.filterButtonActive]}
          onPress={handleFilterPress}
        >
          <Icon
            name="filter-list"
            size={24}
            color={hasActiveFilters ? '#2196F3' : '#666'}
          />
          {hasActiveFilters && <View style={styles.filterBadge} />}
        </TouchableOpacity>
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      {renderHeader()}
      <FlatList
        data={state.transactions}
        keyExtractor={(item) => item.id.toString()}
        renderItem={({ item }) => (
          <TransactionListItem
            transaction={item}
            onPress={handleTransactionPress}
          />
        )}
        contentContainerStyle={
          state.transactions.length === 0 ? styles.emptyListContainer : null
        }
        ListEmptyComponent={state.error ? renderErrorState() : renderEmptyState()}
        ListFooterComponent={renderListFooter}
        refreshControl={
          <RefreshControl
            refreshing={state.refreshing}
            onRefresh={handleRefresh}
            colors={['#2196F3']}
          />
        }
        onEndReached={handleLoadMore}
        onEndReachedThreshold={0.5}
      />

      {/* Filter Modal */}
      <TransactionFilterModal
        visible={filterModalVisible}
        onClose={() => setFilterModalVisible(false)}
        onApply={handleApplyFilters}
        initialFilters={state.filters}
      />
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
    backgroundColor: '#FFFFFF',
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
  },
  headerTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#212121',
  },
  filterButton: {
    padding: 8,
    borderRadius: 20,
  },
  filterButtonActive: {
    backgroundColor: '#E3F2FD',
  },
  filterBadge: {
    position: 'absolute',
    top: 8,
    right: 8,
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: '#2196F3',
  },
  transactionItem: {
    backgroundColor: '#FFFFFF',
    marginHorizontal: 16,
    marginVertical: 8,
    padding: 16,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#E0E0E0',
  },
  transactionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  transactionNumber: {
    fontSize: 14,
    fontWeight: '600',
    color: '#212121',
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
  },
  statusText: {
    fontSize: 12,
    fontWeight: '500',
    color: '#FFFFFF',
  },
  transactionDetails: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  transactionTotal: {
    fontSize: 18,
    fontWeight: '700',
    color: '#212121',
  },
  transactionTime: {
    fontSize: 14,
    color: '#757575',
  },
  transactionDate: {
    fontSize: 12,
    color: '#9E9E9E',
  },
  emptyListContainer: {
    flex: 1,
  },
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  emptyStateTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#424242',
    marginTop: 16,
  },
  emptyStateMessage: {
    fontSize: 14,
    color: '#757575',
    marginTop: 8,
    textAlign: 'center',
  },
  errorState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
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
  listFooter: {
    paddingVertical: 16,
    alignItems: 'center',
  },
});
