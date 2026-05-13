/**
 * TopControlBar Component
 * Top control area with product search/barcode scan input and cart summary
 * Displays running total and payment button
 */

import React from 'react';
import { View, Text, TextInput, TouchableOpacity, StyleSheet } from 'react-native';

interface TopControlBarProps {
  itemCount: number;
  total: string;
  searchQuery?: string;
  onSearch: (query: string) => void;
  onPayment: () => void;
}

export const TopControlBar: React.FC<TopControlBarProps> = ({
  itemCount,
  total,
  searchQuery = '',
  onSearch,
  onPayment,
}) => {
  const formatPrice = (price: string): string => {
    // Validate price input
    if (!price || typeof price !== 'string') {
      return 'Rp 0';
    }

    const priceNum = parseFloat(price);

    // Validate parsed price
    if (isNaN(priceNum) || priceNum < 0) {
      return 'Rp 0';
    }

    return `Rp ${priceNum.toLocaleString('id-ID')}`;
  };

  const isCartEmpty = itemCount === 0;

  return (
    <View style={styles.container}>
      <View style={styles.searchContainer}>
        <TextInput
          style={styles.searchInput}
          placeholder="Search products or scan barcode..."
          placeholderTextColor="#757575"
          value={searchQuery}
          onChangeText={onSearch}
          autoCapitalize="none"
          autoCorrect={false}
          testID="search-input"
        />
      </View>

      <View style={styles.cartSummary}>
        <View style={styles.cartInfo}>
          <Text style={styles.cartCount}>
            Cart: {itemCount} {itemCount === 1 ? 'item' : 'items'}
          </Text>
          <Text style={styles.cartTotal}>{formatPrice(total)}</Text>
        </View>

        <TouchableOpacity
          testID="payment-button"
          style={[styles.paymentButton, isCartEmpty && styles.paymentButtonDisabled]}
          onPress={onPayment}
          disabled={isCartEmpty}
          activeOpacity={0.7}
        >
          <Text style={[styles.paymentButtonText, isCartEmpty && styles.paymentButtonTextDisabled]}>
            Pay
          </Text>
        </TouchableOpacity>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#FFFFFF',
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
    padding: 16,
    minHeight: 100,
  },

  searchContainer: {
    marginBottom: 12,
  },

  searchInput: {
    backgroundColor: '#F5F5F5',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#212121',
    borderWidth: 1,
    borderColor: '#E0E0E0',
    minHeight: 48,
  },

  cartSummary: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },

  cartInfo: {
    flex: 1,
    marginRight: 12,
  },

  cartCount: {
    fontSize: 14,
    color: '#757575',
    marginBottom: 4,
  },

  cartTotal: {
    fontSize: 20,
    fontWeight: '700',
    color: '#4CAF50',
  },

  paymentButton: {
    backgroundColor: '#4CAF50',
    paddingHorizontal: 32,
    paddingVertical: 12,
    borderRadius: 8,
    minWidth: 100,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 48,
  },

  paymentButtonDisabled: {
    backgroundColor: '#BDBDBD',
    opacity: 0.5,
  },

  paymentButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },

  paymentButtonTextDisabled: {
    color: '#757575',
  },
});
