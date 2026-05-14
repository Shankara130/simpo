/**
 * CartItem Component
 * Individual cart item with quantity controls and remove button
 * Displays product details and handles user interactions
 */

import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet, AccessibilityInfo } from 'react-native';
import { CartItem as CartItemType } from '../types/cart.types';
import { formatCurrency } from '../utils/formatCurrency';

interface CartItemProps {
  productId: number;
  sku: string;
  name: string;
  price: string;
  quantity: number;
  subtotal: string;
  stockQty: number;
  onUpdateQuantity: (productId: number, quantity: number) => void;
  onRemove: (productId: number) => void;
}

export const CartItem: React.FC<CartItemProps> = ({
  productId,
  sku,
  name,
  price,
  quantity,
  subtotal,
  stockQty,
  onUpdateQuantity,
  onRemove,
}) => {
  const isAtStockLimit = quantity >= stockQty;

  const handleIncrease = () => {
    if (quantity < stockQty) {
      onUpdateQuantity(productId, quantity + 1);
    } else {
      // Show feedback that stock is limited
      AccessibilityInfo.announceForAccessibility('Stok terbatas');
    }
  };

  const handleDecrease = () => {
    if (quantity > 1) {
      onUpdateQuantity(productId, quantity - 1);
    }
    // If quantity === 1, decrease button does nothing
    // User must click remove button to delete
  };

  const handleRemove = () => {
    onRemove(productId);
  };

  const priceNum = parseFloat(price);
  const subtotalNum = parseFloat(subtotal);

  // Validate parsed values
  if (isNaN(priceNum) || !Number.isFinite(priceNum)) {
    console.warn('CartItem: Invalid price value', price);
  }
  if (isNaN(subtotalNum) || !Number.isFinite(subtotalNum)) {
    console.warn('CartItem: Invalid subtotal value', subtotal);
  }

  return (
    <View style={styles.container} testID={`cart-item-${productId}`}>
      {/* Product Info */}
      <View style={styles.productInfo}>
        <Text style={styles.name} numberOfLines={2}>
          {name}
        </Text>
        <Text style={styles.sku}>SKU: {sku}</Text>
        <Text style={styles.price}>{formatCurrency(priceNum)}</Text>
      </View>

      {/* Quantity Controls */}
      <View style={styles.quantityControls}>
        <TouchableOpacity
          testID={`decrease-qty-${productId}`}
          style={styles.qtyButton}
          onPress={handleDecrease}
          accessibilityLabel="Kurangi jumlah"
          accessibilityRole="button"
        >
          <Text style={styles.qtyButtonText}>-</Text>
        </TouchableOpacity>

        <Text style={styles.quantity}>{quantity}</Text>

        <TouchableOpacity
          testID={`increase-qty-${productId}`}
          style={[
            styles.qtyButton,
            isAtStockLimit && styles.qtyButtonDisabled,
          ]}
          onPress={handleIncrease}
          disabled={isAtStockLimit}
          accessibilityLabel="Tambah jumlah"
          accessibilityRole="button"
          accessibilityState={{ disabled: isAtStockLimit }}
        >
          <Text
            style={[
              styles.qtyButtonText,
              isAtStockLimit && styles.qtyButtonTextDisabled,
            ]}
          >
            +
          </Text>
        </TouchableOpacity>
      </View>

      {/* Subtotal and Actions */}
      <View style={styles.actions}>
        <Text style={styles.subtotal} testID={`subtotal-${productId}`}>
          {formatCurrency(subtotalNum)}
        </Text>

        {isAtStockLimit && (
          <Text style={styles.stockWarning}>Stok terbatas</Text>
        )}

        <TouchableOpacity
          testID={`remove-item-${productId}`}
          style={styles.removeButton}
          onPress={handleRemove}
          accessibilityLabel="Hapus item"
          accessibilityRole="button"
        >
          <Text style={styles.removeButtonText}>🗑️</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    padding: 12,
    marginVertical: 4,
    marginHorizontal: 8,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },

  productInfo: {
    flex: 1,
    marginRight: 12,
  },

  name: {
    fontSize: 14,
    fontWeight: '600',
    color: '#1A1A1A',
    marginBottom: 4,
  },

  sku: {
    fontSize: 12,
    color: '#757575',
    marginBottom: 4,
  },

  price: {
    fontSize: 13,
    color: '#424242',
    fontWeight: '500',
  },

  quantityControls: {
    flexDirection: 'row',
    alignItems: 'center',
    marginRight: 12,
  },

  qtyButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#1976D2',
    justifyContent: 'center',
    alignItems: 'center',
    marginHorizontal: 4,
  },

  qtyButtonDisabled: {
    backgroundColor: '#BDBDBD',
  },

  qtyButtonText: {
    fontSize: 18,
    fontWeight: '600',
    color: '#FFFFFF',
  },

  qtyButtonTextDisabled: {
    color: '#757575',
  },

  quantity: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1A1A1A',
    minWidth: 24,
    textAlign: 'center',
  },

  actions: {
    alignItems: 'flex-end',
    minWidth: 80,
  },

  subtotal: {
    fontSize: 14,
    fontWeight: '600',
    color: '#1976D2',
    marginBottom: 4,
  },

  stockWarning: {
    fontSize: 11,
    color: '#D32F2F',
    marginBottom: 4,
  },

  removeButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#FF5252',
    justifyContent: 'center',
    alignItems: 'center',
  },

  removeButtonText: {
    fontSize: 16,
  },
});
