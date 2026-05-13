/**
 * CartSummary Component
 * Displays cart items with quantity controls and running total
 * Shows item details: name, SKU, quantity, unit price, subtotal
 */

import React from 'react';
import { View, Text, ScrollView, TouchableOpacity, StyleSheet } from 'react-native';
import { CartItem } from '../types/cart.types';

interface CartSummaryProps {
  items?: CartItem[];
  total?: string;
  itemCount?: number;
  onRemove?: (productId: number) => void;
  onUpdateQuantity?: (productId: number, quantity: number) => void;
}

export const CartSummary: React.FC<CartSummaryProps> = ({
  items = [],
  total = '0.00',
  itemCount = 0,
  onRemove,
  onUpdateQuantity,
}) => {
  const formatPrice = (price: string): string => {
    const priceNum = parseFloat(price);
    return `Rp ${priceNum.toLocaleString('id-ID')}`;
  };

  const formatTotal = (total: string): string => {
    const totalNum = parseFloat(total);
    return `Total: ${formatPrice(total)}`;
  };

  if (items.length === 0) {
    return (
      <View style={styles.container} testID="cart-summary-container">
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyText}>Cart is empty</Text>
        </View>
      </View>
    );
  }

  return (
    <View style={styles.container} testID="cart-summary-container">
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Cart ({itemCount} {itemCount === 1 ? 'item' : 'items'})</Text>
      </View>

      <ScrollView style={styles.itemsContainer} showsVerticalScrollIndicator={false}>
        {items.map((item) => (
          <View key={item.productId} style={styles.cartItem}>
            <View style={styles.itemDetails}>
              <Text style={styles.itemName}>{item.name}</Text>
              <Text style={styles.itemSku}>{item.sku}</Text>
              <Text style={styles.itemPrice}>{formatPrice(item.price)} each</Text>
              <Text style={styles.itemSubtotal}>{formatPrice(item.subtotal)}</Text>
            </View>

            <View style={styles.quantityControls}>
              <TouchableOpacity
                testID={`decrease-${item.productId}`}
                style={styles.quantityButton}
                onPress={() => onUpdateQuantity?.(item.productId, item.quantity - 1)}
                disabled={!onUpdateQuantity}
              >
                <Text style={styles.quantityButtonText}>-</Text>
              </TouchableOpacity>

              <Text style={styles.quantityText}>{item.quantity}</Text>

              <TouchableOpacity
                testID={`increase-${item.productId}`}
                style={styles.quantityButton}
                onPress={() => onUpdateQuantity?.(item.productId, item.quantity + 1)}
                disabled={!onUpdateQuantity}
              >
                <Text style={styles.quantityButtonText}>+</Text>
              </TouchableOpacity>
            </View>

            <TouchableOpacity
              testID={`remove-${item.productId}`}
              style={styles.removeButton}
              onPress={() => onRemove?.(item.productId)}
              disabled={!onRemove}
            >
              <Text style={styles.removeButtonText}>Remove</Text>
            </TouchableOpacity>
          </View>
        ))}
      </ScrollView>

      <View style={styles.totalContainer}>
        <Text style={styles.totalText}>{formatTotal(total)}</Text>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#FFFFFF',
    borderTopWidth: 1,
    borderTopColor: '#E0E0E0',
    minHeight: 120,
  },

  emptyContainer: {
    padding: 24,
    alignItems: 'center',
    justifyContent: 'center',
  },

  emptyText: {
    fontSize: 16,
    color: '#757575',
    fontStyle: 'italic',
  },

  header: {
    padding: 12,
    paddingHorizontal: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
    backgroundColor: '#F5F5F5',
  },

  headerTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#212121',
  },

  itemsContainer: {
    maxHeight: 200,
    padding: 12,
  },

  cartItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
    paddingHorizontal: 12,
    backgroundColor: '#FAFAFA',
    borderRadius: 8,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: '#EEEEEE',
  },

  itemDetails: {
    flex: 1,
    marginRight: 8,
  },

  itemName: {
    fontSize: 14,
    fontWeight: '600',
    color: '#212121',
    marginBottom: 2,
  },

  itemSku: {
    fontSize: 12,
    color: '#757575',
    marginBottom: 2,
  },

  itemPrice: {
    fontSize: 12,
    color: '#757575',
    marginBottom: 2,
  },

  itemSubtotal: {
    fontSize: 14,
    fontWeight: '600',
    color: '#4CAF50',
  },

  quantityControls: {
    flexDirection: 'row',
    alignItems: 'center',
    marginRight: 8,
  },

  quantityButton: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: '#2196F3',
    alignItems: 'center',
    justifyContent: 'center',
    marginHorizontal: 4,
  },

  quantityButtonText: {
    color: '#FFFFFF',
    fontSize: 18,
    fontWeight: '600',
  },

  quantityText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#212121',
    minWidth: 24,
    textAlign: 'center',
  },

  removeButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 4,
    backgroundColor: '#F44336',
    minHeight: 36,
    justifyContent: 'center',
  },

  removeButtonText: {
    color: '#FFFFFF',
    fontSize: 12,
    fontWeight: '600',
  },

  totalContainer: {
    padding: 16,
    borderTopWidth: 2,
    borderTopColor: '#2196F3',
    backgroundColor: '#F5F5F5',
  },

  totalText: {
    fontSize: 20,
    fontWeight: '700',
    color: '#212121',
    textAlign: 'right',
  },
});
