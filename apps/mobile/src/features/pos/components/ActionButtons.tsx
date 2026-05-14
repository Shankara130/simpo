/**
 * ActionButtons Component
 * Bottom action buttons for checkout, payment, and clear cart
 * Prominent payment button with clear cart secondary action
 */

import React, { useState } from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Modal } from 'react-native';
import { PaymentMethod, PaymentData } from '../types/payment.types';
import { PaymentModal } from './PaymentModal';
import { formatCurrency } from '../utils/formatCurrency';

export interface ActionButtonsProps {
  itemCount: number;
  cartTotal: string;
  onCheckout: () => void;
  onClearCart: () => void;
  onPaymentMethodSelected?: (paymentData: PaymentData) => void;
}

export const ActionButtons: React.FC<ActionButtonsProps> = ({
  itemCount,
  cartTotal,
  onCheckout,
  onClearCart,
  onPaymentMethodSelected,
}) => {
  const isCartEmpty = itemCount === 0;
  const [isPaymentModalVisible, setIsPaymentModalVisible] = useState(false);

  const handleClearCart = () => {
    if (isCartEmpty) return;

    // In production: Show confirmation dialog
    onClearCart();
  };

  const handlePayment = () => {
    if (isCartEmpty) return;
    setIsPaymentModalVisible(true);
  };

  const handlePaymentMethodSelected = (paymentData: PaymentData) => {
    setIsPaymentModalVisible(false);
    if (onPaymentMethodSelected) {
      onPaymentMethodSelected(paymentData);
    }
  };

  const handleCheckout = () => {
    if (isCartEmpty) return;
    onCheckout();
  };

  const handleCancelPayment = () => {
    setIsPaymentModalVisible(false);
  };

  return (
    <View style={styles.container}>
      {/* Payment Modal */}
      <PaymentModal
        isVisible={isPaymentModalVisible}
        cartTotal={cartTotal}
        onConfirm={handlePaymentMethodSelected}
        onCancel={handleCancelPayment}
      />

      {/* Action Buttons */}
      <View style={styles.buttonContainer}>
        <TouchableOpacity
          testID="clear-button"
          style={[styles.button, styles.clearButton, isCartEmpty && styles.buttonDisabled]}
          onPress={handleClearCart}
          disabled={isCartEmpty}
          activeOpacity={0.7}
        >
          <Text style={[styles.buttonText, isCartEmpty && styles.buttonTextDisabled]}>
            Clear Cart
          </Text>
        </TouchableOpacity>

        <TouchableOpacity
          testID="payment-button"
          style={[styles.button, styles.paymentButton, isCartEmpty && styles.buttonDisabled]}
          onPress={handlePayment}
          disabled={isCartEmpty}
          activeOpacity={0.7}
        >
          <Text style={[styles.buttonText, isCartEmpty && styles.buttonTextDisabled]}>
            Payment
          </Text>
        </TouchableOpacity>

        <TouchableOpacity
          testID="checkout-button"
          style={[styles.button, styles.checkoutButton, isCartEmpty && styles.buttonDisabled]}
          onPress={handleCheckout}
          disabled={isCartEmpty}
          activeOpacity={0.7}
        >
          <Text style={[styles.buttonText, styles.checkoutButtonText, isCartEmpty && styles.buttonTextDisabled]}>
            Checkout
          </Text>
        </TouchableOpacity>
      </View>

      {itemCount > 0 && (
        <View style={styles.itemCountContainer}>
          <Text style={styles.itemCountText}>
            {itemCount} {itemCount === 1 ? 'item' : 'items'} in cart
          </Text>
        </View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#FFFFFF',
    borderTopWidth: 1,
    borderTopColor: '#E0E0E0',
    padding: 16,
    minHeight: 80,
  },

  buttonContainer: {
    flexDirection: 'row',
    gap: 8,
  },

  button: {
    flex: 1,
    paddingHorizontal: 16,
    paddingVertical: 14,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 48,
  },

  clearButton: {
    backgroundColor: '#F44336',
  },

  paymentButton: {
    backgroundColor: '#2196F3',
  },

  checkoutButton: {
    backgroundColor: '#4CAF50',
  },

  buttonDisabled: {
    backgroundColor: '#BDBDBD',
    opacity: 0.5,
  },

  buttonText: {
    color: '#FFFFFF',
    fontSize: 14,
    fontWeight: '600',
  },

  checkoutButtonText: {
    fontSize: 16,
    fontWeight: '700',
  },

  buttonTextDisabled: {
    color: '#757575',
  },

  itemCountContainer: {
    marginTop: 8,
    alignItems: 'center',
  },

  itemCountText: {
    fontSize: 14,
    color: '#757575',
  },
});
