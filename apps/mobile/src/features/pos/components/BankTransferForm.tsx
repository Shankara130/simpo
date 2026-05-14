/**
 * BankTransferForm Component
 * Input form for bank transfer payment method
 * Collects account name and reference number with validation
 */

import React from 'react';
import { View, Text, TextInput, StyleSheet } from 'react-native';

export interface BankTransferFormProps {
  accountName: string;
  referenceNumber: string;
  onAccountNameChange: (value: string) => void;
  onReferenceNumberChange: (value: string) => void;
  errors?: {
    accountName?: string;
    referenceNumber?: string;
  };
}

export const BankTransferForm: React.FC<BankTransferFormProps> = ({
  accountName,
  referenceNumber,
  onAccountNameChange,
  onReferenceNumberChange,
  errors = {},
}) => {
  return (
    <View style={styles.container} testID="bank-transfer-form">
      <View style={styles.fieldContainer}>
        <Text style={styles.label}>Nama Akun Bank</Text>
        <TextInput
          testID="account-name-input"
          style={[
            styles.input,
            errors.accountName && styles.inputError,
          ]}
          value={accountName}
          onChangeText={onAccountNameChange}
          placeholder="Nama Akun"
          placeholderTextColor="#9E9E9E"
          autoCapitalize="words"
          autoComplete="off"
          accessible={true}
          accessibilityLabel="Nama Akun Bank"
          accessibilityHint="Masukkan nama akun bank pengirim"
        />
        {errors.accountName && (
          <Text style={styles.errorText}>{errors.accountName}</Text>
        )}
      </View>

      <View style={styles.fieldContainer}>
        <Text style={styles.label}>Nomor Referensi</Text>
        <TextInput
          testID="reference-number-input"
          style={[
            styles.input,
            errors.referenceNumber && styles.inputError,
          ]}
          value={referenceNumber}
          onChangeText={onReferenceNumberChange}
          placeholder="Nomor Referensi"
          placeholderTextColor="#9E9E9E"
          autoCapitalize="characters"
          autoComplete="off"
          accessible={true}
          accessibilityLabel="Nomor Referensi"
          accessibilityHint="Masukkan nomor referensi transfer"
        />
        {errors.referenceNumber && (
          <Text style={styles.errorText}>{errors.referenceNumber}</Text>
        )}
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    paddingVertical: 16,
  },
  fieldContainer: {
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
