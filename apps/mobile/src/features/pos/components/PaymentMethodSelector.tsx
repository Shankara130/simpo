/**
 * PaymentMethodSelector Component
 * Allows cashiers to select payment method (Cash, Bank Transfer, E-Wallet)
 * Displays three payment options with icons and highlights selected option
 */

import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { PaymentMethod } from '../types/payment.types';

export interface PaymentMethodSelectorProps {
  selectedMethod: PaymentMethod | null;
  onSelectMethod: (method: PaymentMethod) => void;
}

interface PaymentMethodOption {
  method: PaymentMethod;
  label: string;
  icon: string;
  accessibilityLabel: string;
}

const paymentMethods: PaymentMethodOption[] = [
  {
    method: PaymentMethod.CASH,
    label: 'Tunai',
    icon: '💵',
    accessibilityLabel: 'Pilih metode pembayaran Tunai',
  },
  {
    method: PaymentMethod.TRANSFER,
    label: 'Transfer Bank',
    icon: '🏦',
    accessibilityLabel: 'Pilih metode pembayaran Transfer Bank',
  },
  {
    method: PaymentMethod.E_WALLET,
    label: 'E-Wallet',
    icon: '📱',
    accessibilityLabel: 'Pilih metode pembayaran E-Wallet',
  },
];

export const PaymentMethodSelector: React.FC<PaymentMethodSelectorProps> = ({
  selectedMethod,
  onSelectMethod,
}) => {
  return (
    <View style={styles.container} testID="payment-method-selector">
      <Text style={styles.title}>Pilih Metode Pembayaran</Text>
      <View style={styles.optionsContainer}>
        {paymentMethods.map((option) => {
          const isSelected = selectedMethod === option.method;
          return (
            <TouchableOpacity
              key={option.method}
              testID={`payment-method-option-${option.method}`}
              style={[
                styles.optionButton,
                isSelected && styles.selectedOptionButton,
              ]}
              onPress={() => onSelectMethod(option.method)}
              accessible={true}
              accessibilityLabel={option.accessibilityLabel}
              accessibilityRole="button"
              accessibilityState={{ selected: isSelected }}
            >
              <View style={styles.optionContent}>
                <Text style={styles.icon}>{option.icon}</Text>
                <Text
                  style={[
                    styles.optionLabel,
                    isSelected && styles.selectedOptionLabel,
                  ]}
                >
                  {option.label}
                </Text>
              </View>
            </TouchableOpacity>
          );
        })}
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    paddingVertical: 16,
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 12,
    color: '#212121',
  },
  optionsContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 12,
  },
  optionButton: {
    flex: 1,
    backgroundColor: '#F5F5F5',
    borderRadius: 8,
    padding: 16,
    minHeight: 88, // Meets 44×44px touch target requirement
    borderWidth: 2,
    borderColor: 'transparent',
  },
  selectedOptionButton: {
    borderColor: '#4CAF50',
    backgroundColor: '#E8F5E9',
  },
  optionContent: {
    alignItems: 'center',
    gap: 8,
  },
  icon: {
    fontSize: 32,
  },
  optionLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#616161',
    textAlign: 'center',
  },
  selectedOptionLabel: {
    color: '#2E7D32',
    fontWeight: '600',
  },
});
