/**
 * ProductList Component
 * Displays a list of products with search/filter functionality
 * Uses FlatList for efficient rendering of large product catalogs
 */

import React from 'react';
import { View, Text, ScrollView, ActivityIndicator, StyleSheet } from 'react-native';
import { Product } from '../types/product.types';
import { ProductCard } from './ProductCard';

interface ProductListProps {
  products: Product[];
  onAddToCart: (product: Product) => void;
  searchQuery?: string;
  loading?: boolean;
  error?: string;
}

export const ProductList: React.FC<ProductListProps> = ({
  products,
  onAddToCart,
  searchQuery = '',
  loading = false,
  error,
}) => {
  // Filter products based on search query
  const filteredProducts = products.filter((product) => {
    const query = searchQuery.toLowerCase();

    // Safely check name and SKU (required fields)
    const nameMatch = product.name?.toLowerCase().includes(query) || false;
    const skuMatch = product.sku?.toLowerCase().includes(query) || false;

    // Safely check description (optional field)
    const descriptionMatch = product.description
      ? product.description.toLowerCase().includes(query)
      : false;

    return nameMatch || skuMatch || descriptionMatch;
  });

  // Show loading indicator
  if (loading) {
    return (
      <View style={styles.centerContainer} testID="loading-indicator">
        <ActivityIndicator size="large" color="#2196F3" />
        <Text style={styles.loadingText}>Loading products...</Text>
      </View>
    );
  }

  // Show error message
  if (error) {
    return (
      <View style={styles.centerContainer}>
        <Text style={styles.errorText}>{error}</Text>
      </View>
    );
  }

  // Show empty state when no products
  if (products.length === 0) {
    return (
      <View style={styles.centerContainer}>
        <Text style={styles.emptyText}>No products found</Text>
        <Text style={styles.emptySubtext}>Try adjusting your search or check back later</Text>
      </View>
    );
  }

  // Show empty state when no results match search
  if (filteredProducts.length === 0) {
    return (
      <View style={styles.centerContainer}>
        <Text style={styles.emptyText}>No products found</Text>
        <Text style={styles.emptySubtext}>No products match "{searchQuery}"</Text>
      </View>
    );
  }

  // Render product list
  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={styles.contentContainer}
      showsVerticalScrollIndicator={true}
      testID="product-list-scroll"
    >
      {filteredProducts.map((product) => (
        <ProductCard
          key={product.id}
          product={product}
          onAdd={onAddToCart}
        />
      ))}
    </ScrollView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },

  contentContainer: {
    paddingVertical: 8,
  },

  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
    backgroundColor: '#F5F5F5',
  },

  loadingText: {
    marginTop: 12,
    fontSize: 16,
    color: '#757575',
  },

  errorText: {
    fontSize: 16,
    color: '#F44336',
    textAlign: 'center',
  },

  emptyText: {
    fontSize: 18,
    fontWeight: '600',
    color: '#757575',
    textAlign: 'center',
    marginBottom: 8,
  },

  emptySubtext: {
    fontSize: 14,
    color: '#9E9E9E',
    textAlign: 'center',
  },
});
