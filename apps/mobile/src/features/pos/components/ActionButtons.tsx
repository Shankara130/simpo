/**
 * ActionButtons Component
 * Bottom action buttons for checkout and clear cart
 * Prominent checkout button with clear cart secondary action
 */

import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Alert } from 'react-native';

interface ActionButtonsProps {
  itemCount: number;
  onCheckout: () => void;
  onClearCart: () => void;
}

export const ActionButtons: React.FC<ActionButtonsProps> = ({
  itemCount,
  onCheckout,
  onClearCart,
}) => {
  const isCartEmpty = itemCount === 0;

  const handleClearCart = () => {
    if (isCartEmpty) return;

    Alert.alert(
      'Clear Cart',
      'Are you sure you want to clear all items from the cart?',
      [
        {
          text: 'Cancel',
          style: 'cancel',
        },
        {
          text: 'Clear',
          style: 'destructive',
          onPress: onClearCart,
        },
      ]
    );
  };

  const handleCheckout = () => {
    if (isCartEmpty) return;
    onCheckout();
  };

  return (
    <View style={styles.container}>
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
    gap: 12,
  },

  button: {
    flex: 1,
    paddingHorizontal: 24,
    paddingVertical: 14,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 48,
  },

  clearButton: {
    backgroundColor: '#F44336',
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
    fontSize: 16,
    fontWeight: '600',
  },

  checkoutButtonText: {
    fontSize: 18,
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
