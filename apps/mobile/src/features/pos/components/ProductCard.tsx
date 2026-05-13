/**
 * ProductCard Component
 * Displays individual product item with add to cart functionality
 * Shows stock status with visual indicators for low/out of stock
 */

import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { Product } from '../types/product.types';

interface ProductCardProps {
  product: Product;
  onAdd: (product: Product) => void;
}

export const ProductCard: React.FC<ProductCardProps> = ({ product, onAdd }) => {
  const isOutOfStock = product.stockQty === 0;
  const isLowStock = product.stockQty > 0 && product.stockQty < product.reorderThreshold;

  const formatPrice = (price: string): string => {
    const priceNum = parseFloat(price);
    return `Rp ${priceNum.toLocaleString('id-ID')}`;
  };

  const getStockTextStyle = () => {
    if (isOutOfStock) return styles.stockOutText;
    if (isLowStock) return styles.stockLowText;
    return styles.stockNormalText;
  };

  const getStockText = () => {
    if (isOutOfStock) return 'Out of Stock';
    return `Stock: ${product.stockQty}`;
  };

  return (
    <View style={styles.container}>
      <View style={styles.productInfo}>
        <Text style={styles.name}>{product.name}</Text>
        <Text style={styles.sku}>{product.sku}</Text>
        <Text style={styles.price}>{formatPrice(product.price)}</Text>
        <Text style={getStockTextStyle()}>{getStockText()}</Text>
      </View>

      <TouchableOpacity
        testID="add-button"
        style={[styles.addButton, isOutOfStock && styles.addButtonDisabled]}
        onPress={() => !isOutOfStock && onAdd(product)}
        disabled={isOutOfStock}
        activeOpacity={0.7}
      >
        <Text style={[styles.addButtonText, isOutOfStock && styles.addButtonTextDisabled]}>
          Add
        </Text>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    backgroundColor: '#FFFFFF',
    padding: 16,
    marginVertical: 8,
    marginHorizontal: 16,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#E0E0E0',
    alignItems: 'center',
    justifyContent: 'space-between',
    minHeight: 80,
  },

  productInfo: {
    flex: 1,
    marginRight: 12,
  },

  name: {
    fontSize: 16,
    fontWeight: '600',
    color: '#212121',
    marginBottom: 4,
  },

  sku: {
    fontSize: 14,
    color: '#757575',
    marginBottom: 4,
  },

  price: {
    fontSize: 16,
    fontWeight: '600',
    color: '#4CAF50',
    marginBottom: 4,
  },

  stockNormalText: {
    fontSize: 14,
    color: '#4CAF50',
  },

  stockLowText: {
    fontSize: 14,
    color: '#FF9800', // Orange for low stock warning
    fontWeight: '500',
  },

  stockOutText: {
    fontSize: 14,
    color: '#F44336', // Red for out of stock
    fontWeight: '600',
  },

  addButton: {
    backgroundColor: '#2196F3',
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 8,
    minWidth: 80,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 48, // Touch target minimum
  },

  addButtonDisabled: {
    backgroundColor: '#BDBDBD',
  },

  addButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },

  addButtonTextDisabled: {
    color: '#757575',
  },
});
