/**
 * CartTotal Component
 * Displays running total, item count, and clear cart button
 * Subscribes to CartContext for real-time updates
 */

import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { useCartContext } from '../context/CartContext';
import { formatCurrency } from '../utils/formatCurrency';

interface CartTotalProps {
  onClearCart: () => void;
}

export const CartTotal: React.FC<CartTotalProps> = ({ onClearCart }) => {
  const { state } = useCartContext();
  const totalNum = parseFloat(state.total);
  const hasItems = state.itemCount > 0;

  // Validate parsed total
  if (isNaN(totalNum) || !Number.isFinite(totalNum)) {
    console.warn('CartTotal: Invalid total value', state.total);
  }

  const handleClearCart = () => {
    onClearCart();
  };

  return (
    <View style={styles.container}>
      {/* Item count */}
      <Text style={styles.itemCount}>
        {state.itemCount} {state.itemCount === 1 ? 'item' : 'items'}
      </Text>

      {/* Total amount */}
      <Text style={styles.total} testID="cart-total-amount">
        {formatCurrency(totalNum)}
      </Text>

      {/* Clear cart button - only show when cart has items */}
      {hasItems && (
        <TouchableOpacity
          testID="clear-cart-button"
          style={styles.clearButton}
          onPress={handleClearCart}
          accessibilityLabel="Kosongkan keranjang"
          accessibilityRole="button"
        >
          <Text style={styles.clearButtonText}>Kosongkan</Text>
        </TouchableOpacity>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    padding: 12,
    marginHorizontal: 8,
    marginVertical: 4,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },

  itemCount: {
    fontSize: 14,
    color: '#757575',
    fontWeight: '500',
  },

  total: {
    fontSize: 20,
    fontWeight: '700',
    color: '#1976D2',
  },

  clearButton: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 6,
    backgroundColor: '#FF5252',
  },

  clearButtonText: {
    fontSize: 12,
    fontWeight: '600',
    color: '#FFFFFF',
  },
});
