/**
 * ProductCard Component
 * Story 4.1, AC4, AC5, AC6: Display product with low stock and expired indicators
 */

import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity } from 'react-native';

interface ProductCardProps {
  id: number;
  sku: string;
  name: string;
  description?: string;
  stockQty: number;
  price: string;
  expiryDate?: string;
  category?: string;
  reorderThreshold: number;
  isLowStock: boolean;    // Story 4.1, AC5: Low stock indicator
  isExpired: boolean;     // Story 4.1, AC6: Expired indicator
  onPress?: () => void;   // Navigate to product details or add to cart
  disabled?: boolean;     // Disable interaction for expired items (AC6)
}

export const ProductCard: React.FC<ProductCardProps> = ({
  id,
  sku,
  name,
  description,
  stockQty,
  price,
  expiryDate,
  category,
  reorderThreshold,
  isLowStock,
  isExpired,
  onPress,
  disabled = false,
}) => {
  // Story 4.1, AC6: Disable expired products
  const isDisabled = disabled || isExpired;

  // Format price (e.g., "15000.00" -> "Rp 15.000")
  const formattedPrice = React.useMemo(() => {
    try {
      const numPrice = parseFloat(price);
      return `Rp ${numPrice.toLocaleString('id-ID')}`;
    } catch {
      return 'Rp -';
    }
  }, [price]);

  // Format expiry date
  const formattedExpiry = React.useMemo(() => {
    if (!expiryDate) return '-';
    try {
      const date = new Date(expiryDate);
      return date.toLocaleDateString('id-ID');
    } catch {
      return '-';
    }
  }, [expiryDate]);

  // Story 4.1, AC5: Low stock indicator color
  const stockColor = isLowStock ? '#FF5722' : '#4CAF50'; // Red for low, green for OK

  // Story 4.1, AC6: Expired indicator color
  const expiredColor = '#9E9E9E'; // Gray for expired

  return (
    <TouchableOpacity
      testID={`product-card-${id}`}
      style={[styles.container, isDisabled && styles.disabled]}
      onPress={isDisabled ? undefined : onPress}
      disabled={isDisabled}
      accessible={true}
      accessibilityLabel={`Product ${name}, SKU ${sku}, ${formattedPrice}, Stock: ${stockQty}`}
    >
      {/* Story 4.1, AC5: Low stock indicator */}
      {isLowStock && (
        <View style={styles.badge}>
          <Text style={styles.badgeText}>!</Text>
        </View>
      )}

      {/* Story 4.1, AC6: Expired badge */}
      {isExpired && (
        <View style={[styles.expiredBadge, { backgroundColor: expiredColor }]}>
          <Text style={styles.expiredBadgeText}>EXPIRED</Text>
        </View>
      )}

      {/* Category */}
      {category && (
        <Text style={styles.category} numberOfLines={1}>
          {category}
        </Text>
      )}

      {/* Product Name */}
      <Text style={styles.name} numberOfLines={2}>
        {name}
      </Text>

      {/* SKU */}
      <Text style={styles.sku}>SKU: {sku}</Text>

      {/* Description */}
      {description && (
        <Text style={styles.description} numberOfLines={1}>
          {description}
        </Text>
      )}

      {/* Stock and Price Row */}
      <View style={styles.row}>
        <View style={styles.stockContainer}>
          <Text style={[styles.stockLabel, { color: stockColor }]}>Stok:</Text>
          <Text style={styles.stockValue}>{stockQty}</Text>
        </View>

        <View style={styles.priceContainer}>
          <Text style={styles.price}>{formattedPrice}</Text>
        </View>
      </View>

      {/* Expiry Date */}
      {expiryDate && (
        <View style={styles.expiryContainer}>
          <Text style={styles.expiryLabel}>Kadaluarsa:</Text>
          <Text style={[styles.expiryValue, isExpired && { color: expiredColor }]}>
            {formattedExpiry}
          </Text>
        </View>
      )}

      {/* Story 4.1, AC6: Disabled message for expired products */}
      {isExpired && (
        <Text style={styles.disabledMessage}>
          Produk kedaluwarsa tidak dapat dijual
        </Text>
      )}
    </TouchableOpacity>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    padding: 12,
    marginVertical: 6,
    marginHorizontal: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 2,
    borderWidth: 1,
    borderColor: '#E0E0E0',
  },

  disabled: {
    opacity: 0.6,
    backgroundColor: '#F5F5F5',
  },

  badge: {
    position: 'absolute',
    top: 8,
    right: 8,
    backgroundColor: '#FF5722',
    borderRadius: 12,
    width: 24,
    height: 24,
    justifyContent: 'center',
    alignItems: 'center',
  },

  badgeText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: 'bold',
  },

  expiredBadge: {
    position: 'absolute',
    top: 8,
    left: 8,
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 4,
  },

  expiredBadgeText: {
    color: '#FFFFFF',
    fontSize: 10,
    fontWeight: '600',
    letterSpacing: 0.5,
  },

  category: {
    fontSize: 11,
    color: '#757575',
    marginBottom: 4,
    textTransform: 'uppercase',
  },

  name: {
    fontSize: 15,
    fontWeight: '600',
    color: '#212121',
    marginBottom: 2,
  },

  sku: {
    fontSize: 12,
    color: '#9E9E9E',
    marginBottom: 8,
  },

  description: {
    fontSize: 12,
    color: '#757575',
    marginBottom: 8,
  },

  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginVertical: 4,
  },

  stockContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },

  stockLabel: {
    fontSize: 12,
    color: '#757575',
    marginRight: 4,
  },

  stockValue: {
    fontSize: 14,
    fontWeight: '600',
    color: '#212121',
  },

  priceContainer: {
    alignItems: 'flex-end',
  },

  price: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#1976D2',
  },

  expiryContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },

  expiryLabel: {
    fontSize: 11,
    color: '#757575',
    marginRight: 4,
  },

  expiryValue: {
    fontSize: 11,
    color: '#212121',
  },

  disabledMessage: {
    fontSize: 10,
    color: '#D32F2F',
    fontStyle: 'italic',
    marginTop: 4,
    textAlign: 'center',
  },
});
