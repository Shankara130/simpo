/**
 * ProductListScreen Component
 * Story 4.1, AC1, AC4, AC7: Product list with search, filters, and pagination
 * Story 4.2, Task 12: Integrate Real-Time Stock into Product List (AC: 1, 5)
 */

import React, { useState, useCallback, useEffect, useRef } from 'react';
import {
  View,
  Text,
  FlatList,
  TextInput,
  TouchableOpacity,
  ActivityIndicator,
  StyleSheet,
  RefreshControl,
  Alert,
  AppState,
  AppStateStatus,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { InventoryService, ProductListItem, ProductListParams } from '../services/inventoryService';
import { ProductCard } from '../components/ProductCard';
import {
  createRealTimeStockService,
  type RealTimeStockServiceConfig,
  type RealTimeStockServiceInstance,
  type StockUpdatedEvent,
  type ConnectionState,
} from '../services/realTimeStockService';

interface ProductListScreenProps {
  // Optional initial filters
  initialCategory?: string;
  initialBranchId?: number;
}

/**
 * ProductListScreen - Main screen for viewing products
 * Story 4.1, AC1: Display products in searchable list with filters
 */
export const ProductListScreen: React.FC<ProductListScreenProps> = ({
  initialCategory,
  initialBranchId,
}) => {
  const navigation = useNavigation();

  // State management
  const [products, setProducts] = useState<ProductListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  // Pagination state
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);

  // Filter state (Story 4.1, AC3: Category and branch filters)
  const [selectedCategory, setSelectedCategory] = useState<string | undefined>(initialCategory);
  const [showFilters, setShowFilters] = useState(false);

  // Story 4.1, AC2: Search functionality
  const [searchDebounce, setSearchDebounce] = useState<string>('');

  // Story 4.2, Task 12.2: Initialize real-time stock service on mount
  useEffect(() => {
    // Get JWT token from storage (in real app, this would come from auth context)
    const token = 'your-jwt-token'; // TODO: Get from auth context

    // Create real-time stock service
    const config: RealTimeStockServiceConfig = {
      wsUrl: 'ws://localhost:8080/api/v1/products/stock/subscribe', // TODO: Use from config
      token,
      branches: initialBranchId ? [initialBranchId] : undefined,
      autoReconnect: true,
    };

    realTimeStockServiceRef.current = createRealTimeStockService(config);

    // Set up event listeners
    realTimeStockServiceRef.current.on('connectionStateChange', (state) => {
      setConnectionState(state);
    });

    realTimeStockServiceRef.current.on('stockUpdate', handleStockUpdate);
    realTimeStockServiceRef.current.on('error', (error) => {
      console.error('[RealTimeStockService] Error:', error);
    });

    // Start monitoring online/offline status
    realTimeStockServiceRef.current.startOnlineMonitoring();

    // Connect to WebSocket
    realTimeStockServiceRef.current.connect();

    // Cleanup on unmount
    return () => {
      realTimeStockServiceRef.current?.destroy();
    };
  }, [initialBranchId]);

  /**
   * Handle real-time stock updates from WebSocket
   * Story 4.2, Task 12.3: Update product stock quantities from WebSocket events
   */
  const handleStockUpdate = useCallback((event: StockUpdatedEvent) => {
    setProducts(prevProducts => {
      return prevProducts.map(product => {
        // Only update if the product matches the event
        if (product.id === event.productId && product.branchId === event.branchId) {
          // Trigger flash animation
          triggerFlash(product.id);

          // Return updated product with new stock
          return {
            ...product,
            stockQty: event.newStock,
            isLowStock: event.newStock < product.reorderThreshold,
            updatedAt: event.updatedAt,
          };
        }
        return product;
      });
    });
  }, []);

  /**
   * Trigger flash animation for a product
   * Story 4.2, Task 12.4: Add visual feedback for stock updates (color flash)
   */
  const triggerFlash = useCallback((productId: number) => {
    setFlashingProducts(prev => new Set([...prev, productId]));

    // Remove flash after animation completes
    setTimeout(() => {
      setFlashingProducts(prev => {
        const newSet = new Set(prev);
        newSet.delete(productId);
        return newSet;
      });
    }, 2000); // 2 second flash animation
  }, []);

  /**
   * Handle app state changes (foreground/background)
   * Story 4.2, Task 12.5: Clean up subscriptions on unmount
   */
  useEffect(() => {
    const subscription = AppState.addEventListener('change', nextAppState => {
      if (
        appState.current.match(/inactive|background/) &&
        nextAppState === 'active'
      ) {
        // App coming from background to foreground
        console.log('[ProductListScreen] App coming to foreground, reconnecting...');
        realTimeStockServiceRef.current?.reconnect();
      } else if (
        nextAppState.match(/inactive|background/) &&
        appState.current === 'active'
      ) {
        // App going to background
        console.log('[ProductListScreen] App going to background, disconnecting...');
        realTimeStockServiceRef.current?.disconnect();
      }

      appState.current = nextAppState;
    });

    return () => subscription.remove();
  }, []);

  // Story 4.2, Task 12.2-12.5: Real-time stock service integration
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');
  const [flashingProducts, setFlashingProducts] = useState<Set<number>>(new Set());
  const realTimeStockServiceRef = useRef<RealTimeStockServiceInstance | null>(null);
  const appState = useRef(AppState.currentState);

  // Fetch products function
  const fetchProducts = useCallback(async (pageNum = 1, reset = false) => {
    setLoading(true);
    setError(null);

    try {
      const params: ProductListParams = {
        page: pageNum,
        limit: 20,
        search: searchDebounce || undefined,
        category: selectedCategory || undefined,
        branch_id: initialBranchId || undefined,
      };

      const response = await InventoryService.listProducts(params);

      if (reset || pageNum === 1) {
        setProducts(response.data);
      } else {
        setProducts(prev => [...prev, ...response.data]);
      }

      setTotal(response.pagination.total);
      setTotalPages(response.pagination.totalPages);
      setPage(pageNum);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Gagal memuat produk');
      }
    } finally {
      setLoading(false);
    }
  }, [searchDebounce, selectedCategory, initialBranchId]);

  // Initial load
  useEffect(() => {
    fetchProducts(1, true);
  }, []);

  // Story 4.1, AC2: Debounced search (300ms delay)
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      setSearchDebounce(searchQuery);
      setPage(1); // Reset to first page when searching
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [searchQuery]);

  // Refetch when search changes
  useEffect(() => {
    if (searchDebounce !== '') {
      fetchProducts(1, true);
    }
  }, [searchDebounce, fetchProducts]);

  // Story 4.1, AC7: Infinite scroll pagination
  const loadMore = useCallback(() => {
    if (!loading && page < totalPages) {
      fetchProducts(page + 1, false);
    }
  }, [loading, page, totalPages, fetchProducts]);

  // Story 4.1, AC2: Clear search
  const handleClearSearch = useCallback(() => {
    setSearchQuery('');
    setSearchDebounce('');
  }, []);

  // Story 4.1, AC2: Search input handler
  const handleSearchChange = useCallback((text: string) => {
    setSearchQuery(text);
  }, []);

  // Story 4.1, AC3: Apply category filter
  const handleCategoryFilter = useCallback((category?: string) => {
    setSelectedCategory(category);
    setPage(1);
    fetchProducts(1, true);
  }, [fetchProducts]);

  // Refresh handler
  const handleRefresh = useCallback(() => {
    fetchProducts(1, true);
  }, [fetchProducts]);

  // Story 4.1, AC4: Render product item
  const renderProduct = useCallback(({ item }: { item: ProductListItem }) => (
    <ProductCard
      id={item.id}
      sku={item.sku}
      name={item.name}
      description={item.description}
      stockQty={item.stockQty}
      price={item.price}
      expiryDate={item.expiryDate}
      category={item.category}
      reorderThreshold={item.reorderThreshold}
      isLowStock={item.isLowStock}
      isExpired={item.isExpired}
      onPress={() => {
        // TODO: Navigate to product details or add to cart
        // For now, just show the product info
        Alert.alert(
          item.name,
          `SKU: ${item.sku}\nStok: ${item.stockQty}\nHarga: ${item.price}`,
        );
      }}
    />
  ), []);

  // Story 4.1, AC7: Loading indicator at bottom when loading more
  const renderFooter = useCallback(() => {
    if (loading) {
      return (
        <View style={styles.loadingMore}>
          <ActivityIndicator size="small" color="#1976D2" />
          <Text style={styles.loadingMoreText}>Memuat produk...</Text>
        </View>
      );
    }
    return null;
  }, [loading]);

  // Story 4.1, AC7: End of list indicator
  const renderListEmptyComponent = useCallback(() => {
    if (loading) {
      return (
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color="#1976D2" testID="product-loading-indicator" />
          <Text style={styles.loadingText}>Memuat produk...</Text>
        </View>
      );
    }

    if (products.length === 0) {
      return (
        <View style={styles.centerContainer}>
          <Text style={styles.emptyIcon}>📦</Text>
          <Text style={styles.emptyMessage}>Tidak ada produk</Text>
          <Text style={styles.emptySubtext}>
            {searchDebounce
              ? `Tidak ada produk untuk "${searchDebounce}"`
              : 'Belum ada produk yang ditambahkan'}
          </Text>
        </View>
      );
    }

    return null;
  }, [loading, products.length, searchDebounce]);

  // Story 4.1, AC2: Clear search button
  const renderClearButton = useCallback(() => {
    if (searchQuery.length > 0) {
      return (
        <TouchableOpacity
          onPress={handleClearSearch}
          style={styles.clearButton}
          testID="clear-search-button"
        >
          <Text style={styles.clearButtonText}>✕</Text>
        </TouchableOpacity>
      );
    }
    return null;
  }, [searchQuery, handleClearSearch]);

  return (
    <View style={styles.container} testID="product-list-screen">
      {/* Story 4.1, AC2: Search Bar */}
      <View style={styles.searchContainer}>
        <View style={styles.searchInputWrapper}>
          <TextInput
            style={styles.searchInput}
            placeholder="Cari produk berdasarkan nama atau SKU..."
            value={searchQuery}
            onChangeText={handleSearchChange}
            autoCorrect={false}
            autoCapitalize="none"
            testID="product-search-input"
          />
          {renderClearButton()}
        </View>
      </View>

      {/* Story 4.1, AC3: Filter buttons (placeholder - would show modal in full implementation) */}
      <View style={styles.filterContainer}>
        <View style={styles.filterInfoContainer}>
          <Text style={styles.filterInfo}>
            {searchDebounce
              ? `Hasil: ${searchDebounce}`
              : `Total: ${total} produk`}
          </Text>
          {/* Story 4.2, Task 12.2: Connection status indicator */}
          <View style={[
            styles.connectionStatus,
            connectionState === 'connected' && styles.connectionStatusConnected,
            connectionState === 'connecting' && styles.connectionStatusConnecting,
            connectionState === 'reconnecting' && styles.connectionStatusReconnecting,
            connectionState === 'error' && styles.connectionStatusError,
          ]}>
            <View style={[
              styles.connectionStatusDot,
              connectionState === 'connected' && styles.connectionStatusDotConnected,
            ]} />
            <Text style={styles.connectionStatusText}>
              {connectionState === 'connected'
                ? 'Live'
                : connectionState === 'connecting'
                ? 'Menghubung...'
                : connectionState === 'reconnecting'
                ? 'Menghubung kembali...'
                : connectionState === 'error'
                ? 'Error'
                : 'Terputus'}
            </Text>
          </View>
        </View>
        <TouchableOpacity
          onPress={() => setShowFilters(!showFilters)}
          style={styles.filterButton}
          testID="filter-button"
        >
          <Text style={styles.filterButtonText}>
            {showFilters ? 'Sembunyikan Filter' : 'Tampilkan Filter'}
          </Text>
        </TouchableOpacity>
      </View>

      {/* Story 4.1, AC7: Product List with Infinite Scroll */}
      <FlatList
        data={products}
        renderItem={renderProduct}
        keyExtractor={(item) => item.id.toString()}
        ListEmptyComponent={renderListEmptyComponent}
        ListFooterComponent={renderFooter}
        onEndReached={loadMore}
        onEndReachedThreshold={0.5}
        refreshControl={
          <RefreshControl
            refreshing={loading && page === 1}
            onRefresh={handleRefresh}
            colors={['#1976D2', '#689F38']}
            tintColor="#1976D2"
          />
        }
        contentContainerStyle={[styles.listContent, products.length === 0 && styles.listContentEmpty]}
        removeClippedSubviews={true}
        maxToRenderPerBatch={10}
        windowSize={5}
        initialNumToRender={10}
        testID="product-list"
      />

      {/* Story 4.1, AC3: Simple filter display (when expanded) */}
      {showFilters && (
        <View style={styles.filterModal}>
          <View style={styles.filterHeader}>
            <Text style={styles.filterTitle}>Filter Produk</Text>
            <TouchableOpacity onPress={() => setShowFilters(false)}>
              <Text style={styles.closeFilterText}>✕</Text>
            </TouchableOpacity>
          </View>

          {/* Category filter */}
          <View style={styles.filterSection}>
            <Text style={styles.filterLabel}>Kategori:</Text>
            <View style={styles.chipContainer}>
              <TouchableOpacity
                style={[styles.chip, !selectedCategory && styles.chipSelected]}
                onPress={() => handleCategoryFilter(undefined)}
              >
                <Text style={styles.chipText}>Semua</Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[styles.chip, selectedCategory === 'Obat Bebas' && styles.chipSelected]}
                onPress={() => handleCategoryFilter('Obat Bebas')}
              >
                <Text style={styles.chipText}>Obat Bebas</Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[styles.chip, selectedCategory === 'Obat Bebas Terbatas' && styles.chipSelected]}
                onPress={() => handleCategoryFilter('Obat Bebas Terbatas')}
              >
                <Text style={styles.chipText}>Obat Bebas Terbatas</Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[styles.chip, selectedCategory === 'Obat Resep' && styles.chipSelected]}
                onPress={() => handleCategoryFilter('Obat Resep')}
              >
                <Text style={styles.chipText}>Obat Resep</Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },

  searchContainer: {
    backgroundColor: '#FFFFFF',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
  },

  searchInputWrapper: {
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#E0E0E0',
    borderRadius: 8,
    paddingHorizontal: 12,
    height: 44,
    backgroundColor: '#F5F5F5',
  },

  searchInput: {
    flex: 1,
    fontSize: 15,
    color: '#212121',
  },

  clearButton: {
    marginLeft: 8,
    width: 32,
    height: 32,
    justifyContent: 'center',
    alignItems: 'center',
  },

  clearButtonText: {
    fontSize: 18,
    color: '#757575',
  },

  filterContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 8,
    backgroundColor: '#FFFFFF',
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
  },

  filterInfoContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },

  filterInfo: {
    fontSize: 13,
    color: '#757575',
  },

  // Story 4.2, Task 12.2: Connection status indicator styles
  connectionStatus: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
    backgroundColor: '#E0E0E0',
  },

  connectionStatusConnected: {
    backgroundColor: '#D1FADF',
  },

  connectionStatusConnecting: {
    backgroundColor: '#FFF3CD',
  },

  connectionStatusReconnecting: {
    backgroundColor: '#FFF3CD',
  },

  connectionStatusError: {
    backgroundColor: '#FEE2E2',
  },

  connectionStatusDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: '#9E9E9E',
  },

  connectionStatusDotConnected: {
    backgroundColor: '#4CAF50',
  },

  connectionStatusText: {
    fontSize: 11,
    color: '#424242',
    fontWeight: '500',
  },

  filterButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    backgroundColor: '#1976D2',
    borderRadius: 4,
  },

  filterButtonText: {
    color: '#FFFFFF',
    fontSize: 13,
    fontWeight: '600',
  },

  listContent: {
    paddingVertical: 8,
    flexGrow: 1,
  },

  listContentEmpty: {
    flexGrow: 1,
  },

  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 60,
  },

  loadingText: {
    fontSize: 14,
    color: '#757575',
    marginTop: 12,
  },

  emptyIcon: {
    fontSize: 64,
    marginBottom: 16,
  },

  emptyMessage: {
    fontSize: 18,
    fontWeight: '600',
    color: '#424242',
    marginBottom: 8,
  },

  emptySubtext: {
    fontSize: 14,
    color: '#757575',
    textAlign: 'center',
    paddingHorizontal: 40,
  },

  loadingMore: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 16,
  },

  loadingMoreText: {
    marginLeft: 8,
    fontSize: 13,
    color: '#757575',
  },

  // Filter Modal Styles
  filterModal: {
    backgroundColor: '#FFFFFF',
    borderTopWidth: 1,
    borderTopColor: '#E0E0E0',
  },

  filterHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
  },

  filterTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#212121',
  },

  closeFilterText: {
    fontSize: 24,
    color: '#757575',
    paddingHorizontal: 8,
  },

  filterSection: {
    padding: 16,
  },

  filterLabel: {
    fontSize: 14,
    fontWeight: '600',
    color: '#424242',
    marginBottom: 8,
  },

  chipContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },

  chip: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
    borderWidth: 1,
    borderColor: '#E0E0E0',
    backgroundColor: '#FFFFFF',
  },

  chipSelected: {
    backgroundColor: '#1976D2',
    borderColor: '#1976D2',
  },

  chipText: {
    fontSize: 13,
    color: '#424242',
  },
});
