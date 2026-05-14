/**
 * PaymentModal Component
 * Modal container for payment method selection and data collection
 * Orchestrates payment flow: method selection → data collection → confirmation
 */

import React, { useState, useEffect, useMemo } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  Modal,
  StyleSheet,
  BackHandler,
  ScrollView,
} from 'react-native';
import { PaymentMethod, EWalletType, PaymentData } from '../types/payment.types';
import { PaymentMethodSelector } from './PaymentMethodSelector';
import { BankTransferForm } from './BankTransferForm';
import { EWalletSelector } from './EWalletSelector';
import { formatCurrency } from '../utils/formatCurrency';

export interface PaymentModalProps {
  isVisible: boolean;
  cartTotal: string;
  onConfirm: (paymentData: PaymentData) => void;
  onCancel: () => void;
  selectedMethod?: PaymentMethod | null;
  accountName?: string;
  referenceNumber?: string;
  selectedWallet?: EWalletType | null;
  confirmationInput?: string;
}

export const PaymentModal: React.FC<PaymentModalProps> = ({
  isVisible,
  cartTotal,
  onConfirm,
  onCancel,
  selectedMethod: propSelectedMethod,
  accountName: propAccountName = '',
  referenceNumber: propReferenceNumber = '',
  selectedWallet: propSelectedWallet,
  confirmationInput: propConfirmationInput = '',
}) => {
  // Local state for payment method selection
  const [selectedMethod, setSelectedMethod] = useState<PaymentMethod | null>(
    propSelectedMethod || null
  );
  const [accountName, setAccountName] = useState(propAccountName);
  const [referenceNumber, setReferenceNumber] = useState(propReferenceNumber);
  const [selectedWallet, setSelectedWallet] = useState<EWalletType | null>(
    propSelectedWallet || null
  );
  const [confirmationInput, setConfirmationInput] = useState(propConfirmationInput);

  // Validation errors
  const [errors, setErrors] = useState<{
    accountName?: string;
    referenceNumber?: string;
    wallet?: string;
    confirmationInput?: string;
  }>({});

  // Reset state when modal closes
  useEffect(() => {
    if (!isVisible) {
      setSelectedMethod(null);
      setAccountName('');
      setReferenceNumber('');
      setSelectedWallet(null);
      setConfirmationInput('');
      setErrors({});
    }
  }, [isVisible]);

  // Handle Android back button
  useEffect(() => {
    const backHandler = BackHandler.addEventListener('hardwareBackPress', () => {
      if (isVisible) {
        onCancel();
        return true; // Prevent default back behavior
      }
      return false; // Allow default back behavior
    });

    return () => backHandler.remove();
  }, [isVisible, onCancel]);

  // Validate payment data
  const validatePaymentData = (): boolean => {
    const newErrors: typeof errors = {};

    if (selectedMethod === PaymentMethod.TRANSFER) {
      if (accountName.trim().length < 2) {
        newErrors.accountName = 'Nama akun minimal 2 karakter';
      }
      if (referenceNumber.trim().length < 3) {
        newErrors.referenceNumber = 'Nomor referensi minimal 3 karakter';
      }
    }

    if (selectedMethod === PaymentMethod.E_WALLET) {
      if (!selectedWallet) {
        newErrors.wallet = 'Pilih e-wallet terlebih dahulu';
      }
      if (confirmationInput.trim().length === 0) {
        newErrors.confirmationInput = 'Nomor konfirmasi harus diisi';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Check if payment data is valid
  const isPaymentValid = useMemo(() => {
    if (selectedMethod === PaymentMethod.CASH) {
      return true; // Cash requires no additional input
    }
    if (selectedMethod === PaymentMethod.TRANSFER) {
      return accountName.length >= 2 && referenceNumber.length >= 3;
    }
    if (selectedMethod === PaymentMethod.E_WALLET) {
      return selectedWallet !== null && confirmationInput.length > 0;
    }
    return false; // No method selected
  }, [selectedMethod, accountName, referenceNumber, selectedWallet, confirmationInput]);

  const handleSelectMethod = (method: PaymentMethod) => {
    setSelectedMethod(method);
    // Clear errors when method changes
    setErrors({});
  };

  const handleConfirm = () => {
    if (!validatePaymentData()) {
      return;
    }

    let paymentData: PaymentData;

    if (selectedMethod === PaymentMethod.CASH) {
      paymentData = { method: PaymentMethod.CASH };
    } else if (selectedMethod === PaymentMethod.TRANSFER) {
      paymentData = {
        method: PaymentMethod.TRANSFER,
        accountName: accountName.trim(),
        referenceNumber: referenceNumber.trim(),
      };
    } else if (selectedMethod === PaymentMethod.E_WALLET) {
      paymentData = {
        method: PaymentMethod.E_WALLET,
        walletType: selectedWallet!,
        confirmationInput: confirmationInput.trim(),
      };
    } else {
      return; // No method selected
    }

    onConfirm(paymentData);
  };

  const getPaymentMethodForm = () => {
    if (selectedMethod === PaymentMethod.TRANSFER) {
      return (
        <BankTransferForm
          accountName={accountName}
          referenceNumber={referenceNumber}
          onAccountNameChange={setAccountName}
          onReferenceNumberChange={setReferenceNumber}
          errors={{
            accountName: errors.accountName,
            referenceNumber: errors.referenceNumber,
          }}
        />
      );
    }

    if (selectedMethod === PaymentMethod.E_WALLET) {
      return (
        <EWalletSelector
          selectedWallet={selectedWallet}
          confirmationInput={confirmationInput}
          onSelectWallet={setSelectedWallet}
          onConfirmationInputChange={setConfirmationInput}
          errors={{
            wallet: errors.wallet,
            confirmationInput: errors.confirmationInput,
          }}
        />
      );
    }

    return null;
  };

  return (
    <Modal
      visible={isVisible}
      animationType="slide"
      transparent={true}
      onRequestClose={onCancel}
      testID="payment-modal"
    >
      <View style={styles.modalOverlay}>
        <View style={styles.modalContent}>
          {/* Header */}
          <View style={styles.header}>
            <Text style={styles.headerTitle}>Pilih Metode Pembayaran</Text>
            <Text style={styles.headerTotal}>
              Total: {formatCurrency(parseFloat(cartTotal))}
            </Text>
          </View>

          {/* Scrollable Content */}
          <ScrollView style={styles.scrollContent}>
            <PaymentMethodSelector
              selectedMethod={selectedMethod}
              onSelectMethod={handleSelectMethod}
            />

            {getPaymentMethodForm()}
          </ScrollView>

          {/* Footer Buttons */}
          <View style={styles.footer}>
            <TouchableOpacity
              testID="cancel-button"
              style={styles.cancelButton}
              onPress={onCancel}
              activeOpacity={0.7}
            >
              <Text style={styles.cancelButtonText}>Batal</Text>
            </TouchableOpacity>

            <TouchableOpacity
              testID="confirm-button"
              style={[
                styles.confirmButton,
                !isPaymentValid && styles.confirmButtonDisabled,
              ]}
              onPress={handleConfirm}
              disabled={!isPaymentValid}
              activeOpacity={0.7}
            >
              <Text
                style={[
                  styles.confirmButtonText,
                  !isPaymentValid && styles.confirmButtonTextDisabled,
                ]}
              >
                Konfirmasi
              </Text>
            </TouchableOpacity>
          </View>
        </View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  modalContent: {
    backgroundColor: '#FFFFFF',
    borderRadius: 16,
    width: '90%',
    maxWidth: 400,
    maxHeight: '80%',
  },
  header: {
    padding: 20,
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: '600',
    marginBottom: 8,
    color: '#212121',
  },
  headerTotal: {
    fontSize: 20,
    fontWeight: '700',
    color: '#4CAF50',
  },
  scrollContent: {
    paddingHorizontal: 20,
    paddingBottom: 20,
  },
  footer: {
    flexDirection: 'row',
    padding: 16,
    borderTopWidth: 1,
    borderTopColor: '#E0E0E0',
    gap: 12,
  },
  cancelButton: {
    flex: 1,
    backgroundColor: '#F5F5F5',
    borderRadius: 8,
    paddingVertical: 14,
    paddingHorizontal: 24,
    alignItems: 'center',
  },
  cancelButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#616161',
  },
  confirmButton: {
    flex: 1,
    backgroundColor: '#4CAF50',
    borderRadius: 8,
    paddingVertical: 14,
    paddingHorizontal: 24,
    alignItems: 'center',
  },
  confirmButtonDisabled: {
    backgroundColor: '#BDBDBD',
  },
  confirmButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFFFFF',
  },
  confirmButtonTextDisabled: {
    color: '#757575',
  },
});
