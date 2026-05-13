/**
 * POSScreen Component
 * Main POS screen that integrates all POS components
 * Layout: Top control bar, Product list, Cart summary, Action buttons
 */

import React, { useState } from 'react';
import { View, StyleSheet, SafeAreaView } from 'react-native';
import { Product } from '../types/product.types';
import { useCartContext } from '../context/CartContext';
import { TopControlBar } from '../components/TopControlBar';
import { ProductList } from '../components/ProductList';
import { CartSummary } from '../components/CartSummary';
import { ActionButtons } from '../components/ActionButtons';

interface POSScreenProps {
  products?: Product[];
  loading?: boolean;
  error?: string;
}

export const POSScreen: React.FC<POSScreenProps> = ({
  products = [],
  loading = false,
  error,
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const { state, actions } = useCartContext();

  const handleAddToCart = (product: Product) => {
    actions.addItem(product);
  };

  const handleRemoveFromCart = (productId: number) => {
    actions.removeItem(productId);
  };

  const handleUpdateQuantity = (productId: number, quantity: number) => {
    actions.updateQuantity(productId, quantity);
  };

  const handleClearCart = () => {
    actions.clearCart();
  };

  const handleCheckout = () => {
    // TODO: Navigate to checkout/payment screen (future story)
    console.log('Checkout clicked', state);
  };

  const handlePayment = () => {
    // TODO: Navigate to payment screen (future story)
    console.log('Payment clicked', state);
  };

  return (
    <SafeAreaView style={styles.safeArea} testID="pos-screen-container">
      <View style={styles.container}>
        {/* Top Control Bar (15%) */}
        <TopControlBar
          itemCount={state.itemCount}
          total={state.total}
          searchQuery={searchQuery}
          onSearch={setSearchQuery}
          onPayment={handlePayment}
        />

        {/* Center Product Area (55%) */}
        <View style={styles.productArea}>
          <ProductList
            products={products}
            searchQuery={searchQuery}
            loading={loading}
            error={error}
            onAddToCart={handleAddToCart}
          />
        </View>

        {/* Cart Summary Panel (15%) */}
        <CartSummary
          items={state.items}
          total={state.total}
          itemCount={state.itemCount}
          onRemove={handleRemoveFromCart}
          onUpdateQuantity={handleUpdateQuantity}
        />

        {/* Bottom Action Buttons (15%) */}
        <ActionButtons
          itemCount={state.itemCount}
          onCheckout={handleCheckout}
          onClearCart={handleClearCart}
        />
      </View>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#FFFFFF',
  },

  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },

  productArea: {
    flex: 1, // Takes remaining space (~55%)
    backgroundColor: '#F5F5F5',
  },
});
