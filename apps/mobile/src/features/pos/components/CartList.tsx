/**
 * CartList Component
 * Scrollable list of cart items with empty state and loading indicator
 */

import React from 'react';
import { View, Text, FlatList, ActivityIndicator, StyleSheet } from 'react-native';
import { CartItem } from '../types/cart.types';
import { CartItem as CartItemComponent } from './CartItem';

interface CartListProps {
  cartItems: CartItem[];
  onUpdateQuantity: (productId: number, quantity: number) => void;
  onRemoveItem: (productId: number) => void;
  loading?: boolean;
}

export const CartList: React.FC<CartListProps> = ({
  cartItems,
  onUpdateQuantity,
  onRemoveItem,
  loading = false,
}) => {
  const renderEmptyState = () => (
    <View style={styles.emptyContainer}>
      <Text style={styles.emptyIcon}>🛒</Text>
      <Text style={styles.emptyMessage}>Keranjang masih kosong</Text>
      <Text style={styles.emptySubtext}>
        Scan atau cari produk untuk menambahkan ke keranjang
      </Text>
    </View>
  );

  const renderLoadingState = () => (
    <View style={styles.loadingContainer}>
      <ActivityIndicator size="large" color="#1976D2" testID="cart-loading-indicator" />
      <Text style={styles.loadingText}>Memuat keranjang...</Text>
    </View>
  );

  const renderCartItem = ({ item }: { item: CartItem }) => (
    <CartItemComponent
      productId={item.productId}
      sku={item.sku}
      name={item.name}
      price={item.price}
      quantity={item.quantity}
      subtotal={item.subtotal}
      stockQty={item.stockQty}
      onUpdateQuantity={onUpdateQuantity}
      onRemove={onRemoveItem}
    />
  );

  const keyExtractor = (item: CartItem) => item.productId.toString();

  if (loading) {
    return renderLoadingState();
  }

  return (
    <View style={styles.container}>
      <FlatList
        data={cartItems}
        renderItem={renderCartItem}
        keyExtractor={keyExtractor}
        ListEmptyComponent={renderEmptyState}
        contentContainerStyle={[
          styles.listContent,
          cartItems.length === 0 && styles.listContentEmpty,
        ]}
        removeClippedSubviews={true}
        maxToRenderPerBatch={10}
        windowSize={5}
        testID="cart-list"
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },

  listContent: {
    paddingVertical: 8,
    paddingHorizontal: 8,
  },

  listContentEmpty: {
    flex: 1,
  },

  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 60,
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
  },

  loadingContainer: {
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
});
