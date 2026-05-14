/**
 * EWalletSelector Component
 * E-wallet selection UI with Indonesian e-wallet providers
 * Displays e-wallet options and confirmation input field
 */

import React from 'react';
import { View, Text, TouchableOpacity, TextInput, StyleSheet } from 'react-native';
import { EWalletType } from '../types/payment.types';

export interface EWalletSelectorProps {
  selectedWallet: EWalletType | null;
  confirmationInput: string;
  onSelectWallet: (wallet: EWalletType) => void;
  onConfirmationInputChange: (value: string) => void;
  errors?: {
    wallet?: string;
    confirmationInput?: string;
  };
}

interface EWalletOption {
  type: EWalletType;
  label: string;
  color: string;
}

const eWalletOptions: EWalletOption[] = [
  { type: EWalletType.GOPAY, label: 'GoPay', color: '#00AED6' },
  { type: EWalletType.OVO, label: 'OVO', color: '#4C3494' },
  { type: EWalletType.DANA, label: 'Dana', color: '#118EEA' },
  { type: EWalletType.SHOPEE_PAY, label: 'ShopeePay', color: '#EE4D2D' },
];

export const EWalletSelector: React.FC<EWalletSelectorProps> = ({
  selectedWallet,
  confirmationInput,
  onSelectWallet,
  onConfirmationInputChange,
  errors = {},
}) => {
  return (
    <View style={styles.container} testID="ewallet-selector">
      <Text style={styles.title}>Pilih E-Wallet</Text>

      {/* E-Wallet Options */}
      <View style={styles.walletOptionsContainer}>
        {eWalletOptions.map((option) => {
          const isSelected = selectedWallet === option.type;
          return (
            <TouchableOpacity
              key={option.type}
              testID={`ewallet-option-${option.type}`}
              style={[
                styles.walletOption,
                isSelected && { borderColor: option.color, backgroundColor: `${option.color}15` },
              ]}
              onPress={() => onSelectWallet(option.type)}
              accessible={true}
              accessibilityLabel={`Pilih ${option.label}`}
              accessibilityRole="button"
              accessibilityState={{ selected: isSelected }}
            >
              <Text
                style={[
                  styles.walletLabel,
                  isSelected && { color: option.color },
                ]}
              >
                {option.label}
              </Text>
            </TouchableOpacity>
          );
        })}
      </View>

      {/* Wallet Selection Error */}
      {errors.wallet && (
        <Text style={styles.errorText}>{errors.wallet}</Text>
      )}

      {/* Confirmation Input */}
      <View style={styles.confirmationContainer}>
        <Text style={styles.label}>Nomor Konfirmasi</Text>
        <TextInput
          testID="confirmation-input"
          style={[
            styles.input,
            errors.confirmationInput && styles.inputError,
          ]}
          value={confirmationInput}
          onChangeText={onConfirmationInputChange}
          placeholder="Nomor Konfirmasi"
          placeholderTextColor="#9E9E9E"
          keyboardType="phone-pad"
          autoComplete="tel"
          accessible={true}
          accessibilityLabel="Nomor Konfirmasi"
          accessibilityHint="Masukkan nomor telepon atau kode konfirmasi"
        />
        {errors.confirmationInput && (
          <Text style={styles.errorText}>{errors.confirmationInput}</Text>
        )}
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
  walletOptionsContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
    marginBottom: 16,
  },
  walletOption: {
    flex: 1,
    minWidth: 80,
    backgroundColor: '#F5F5F5',
    borderRadius: 8,
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderWidth: 2,
    borderColor: 'transparent',
    alignItems: 'center',
    minHeight: 48,
  },
  walletLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#616161',
  },
  confirmationContainer: {
    marginBottom: 16,
  },
  label: {
    fontSize: 14,
    fontWeight: '500',
    marginBottom: 8,
    color: '#424242',
  },
  input: {
    backgroundColor: '#FFFFFF',
    borderWidth: 1,
    borderColor: '#E0E0E0',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#212121',
  },
  inputError: {
    borderColor: '#F44336',
  },
  errorText: {
    fontSize: 12,
    color: '#F44336',
    marginTop: 4,
  },
});
