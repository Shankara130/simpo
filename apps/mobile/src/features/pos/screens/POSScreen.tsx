/**
 * POSScreen Component
 * Main POS screen that integrates all POS components
 * Layout: Top control bar, Product list, Cart summary (CartList + CartTotal), Action buttons
 */

import React, { useState } from 'react';
import { View, StyleSheet, SafeAreaView, ScrollView } from 'react-native';
import { Product } from '../types/product.types';
import { PaymentData } from '../types/payment.types';
import { useCartContext } from '../context/CartContext';
import { TopControlBar } from '../components/TopControlBar';
import { ProductList } from '../components/ProductList';
import { CartList } from '../components/CartList';
import { CartTotal } from '../components/CartTotal';
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
  const [paymentData, setPaymentData] = useState<PaymentData | null>(null);
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
    // Payment is now handled via ActionButtons → PaymentModal
    console.log('Payment flow initiated');
  };

  const handlePaymentMethodSelected = (data: PaymentData) => {
    // Store payment data for transaction processing
    setPaymentData(data);

    // Log payment method selection for audit trail
    console.log('Payment method selected:', {
      method: data.method,
      timestamp: new Date().toISOString(),
      cartItemCount: state.itemCount,
      cartTotal: state.total,
    });

    // TODO: Future story (3.6) - Pass payment data to transaction creation endpoint
    // For now, just store it locally and log the selection
    console.log('Payment data stored for transaction:', data);
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

        {/* Cart Summary Panel (15%) - CartList + CartTotal */}
        <View style={styles.cartSummaryPanel}>
          <ScrollView style={styles.cartListContainer} nestedScrollEnabled={false}>
            <CartList
              cartItems={state.items}
              onUpdateQuantity={handleUpdateQuantity}
              onRemoveItem={handleRemoveFromCart}
            />
          </ScrollView>
          <CartTotal onClearCart={handleClearCart} />
        </View>

        {/* Bottom Action Buttons (15%) */}
        <ActionButtons
          itemCount={state.itemCount}
          cartTotal={state.total}
          onCheckout={handleCheckout}
          onClearCart={handleClearCart}
          onPaymentMethodSelected={handlePaymentMethodSelected}
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

  cartSummaryPanel: {
    height: 150, // ~15% of screen height (approx)
    backgroundColor: '#FFFFFF',
    borderTopWidth: 1,
    borderTopColor: '#E0E0E0',
  },

  cartListContainer: {
    flex: 1,
  },
});
