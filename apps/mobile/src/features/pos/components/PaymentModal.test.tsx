/**
 * PaymentModal Component Tests
 * Test payment modal orchestration and conditional rendering
 */

import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { PaymentModal } from './PaymentModal';
import { PaymentMethod, EWalletType } from '../types/payment.types';

describe('PaymentModal Component', () => {
  const defaultProps = {
    isVisible: true,
    cartTotal: '150000.00',
    onConfirm: jest.fn(),
    onCancel: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render modal when isVisible is true', () => {
    const { getByTestId } = render(
      <PaymentModal {...defaultProps} />
    );

    expect(getByTestId('payment-modal')).toBeTruthy();
  });

  it('should hide modal when isVisible is false', () => {
    const { queryByTestId } = render(
      <PaymentModal {...defaultProps} isVisible={false} />
    );

    expect(queryByTestId('payment-modal')).toBeNull();
  });

  it('should render PaymentMethodSelector component', () => {
    const { getByTestId } = render(
      <PaymentModal {...defaultProps} />
    );

    expect(getByTestId('payment-method-selector')).toBeTruthy();
  });

  it('should not render BankTransferForm when no payment method is selected', () => {
    const { queryByTestId } = render(
      <PaymentModal {...defaultProps} />
    );

    expect(queryByTestId('bank-transfer-form')).toBeNull();
  });

  it('should render BankTransferForm when TRANSFER is selected', () => {
    const { getByTestId, getByText } = render(
      <PaymentModal {...defaultProps} selectedMethod={PaymentMethod.TRANSFER} />
    );

    // Press transfer option to select it
    fireEvent.press(getByText('Transfer Bank'));

    expect(getByTestId('bank-transfer-form')).toBeTruthy();
  });

  it('should not render EWalletSelector when no payment method is selected', () => {
    const { queryByTestId } = render(
      <PaymentModal {...defaultProps} />
    );

    expect(queryByTestId('ewallet-selector')).toBeNull();
  });

  it('should render EWalletSelector when E_WALLET is selected', () => {
    const { getByTestId, getByText } = render(
      <PaymentModal {...defaultProps} selectedMethod={PaymentMethod.E_WALLET} />
    );

    // Press e-wallet option to select it
    fireEvent.press(getByText('E-Wallet'));

    expect(getByTestId('ewallet-selector')).toBeTruthy();
  });

  it('should render no additional form when CASH is selected', () => {
    const { queryByTestId, getByText } = render(
      <PaymentModal {...defaultProps} selectedMethod={PaymentMethod.CASH} />
    );

    // Press cash option to select it
    fireEvent.press(getByText('Tunai'));

    expect(queryByTestId('bank-transfer-form')).toBeNull();
    expect(queryByTestId('ewallet-selector')).toBeNull();
  });

  it('should call onCancel when Cancel button pressed', () => {
    const mockOnCancel = jest.fn();
    const { getByText } = render(
      <PaymentModal {...defaultProps} onCancel={mockOnCancel} />
    );

    fireEvent.press(getByText('Batal'));

    expect(mockOnCancel).toHaveBeenCalledTimes(1);
  });

  it('should disable Confirm button when no payment method is selected', () => {
    const { getByTestId } = render(
      <PaymentModal {...defaultProps} />
    );

    const confirmButton = getByTestId('confirm-button');
    // Check that button has disabled styling
    expect(confirmButton.props.style.backgroundColor).toBe('#BDBDBD');
  });

  it('should enable Confirm button when CASH payment method is selected', () => {
    const { getByTestId, getByText } = render(
      <PaymentModal {...defaultProps} selectedMethod={PaymentMethod.CASH} />
    );

    fireEvent.press(getByText('Tunai'));

    const confirmButton = getByTestId('confirm-button');
    // Check that button has enabled styling (green)
    expect(confirmButton.props.style.backgroundColor).toBe('#4CAF50');
  });

  it('should call onConfirm with valid cash payment data when Confirm pressed', () => {
    const mockOnConfirm = jest.fn();
    const { getByText } = render(
      <PaymentModal {...defaultProps} onConfirm={mockOnConfirm} selectedMethod={PaymentMethod.CASH} />
    );

    fireEvent.press(getByText('Tunai'));
    fireEvent.press(getByText('Konfirmasi'));

    expect(mockOnConfirm).toHaveBeenCalledWith({
      method: PaymentMethod.CASH,
    });
  });

  it('should disable Confirm button when TRANSFER payment method is selected but fields are empty', () => {
    const { getByTestId, getByText } = render(
      <PaymentModal {...defaultProps} selectedMethod={PaymentMethod.TRANSFER} />
    );

    fireEvent.press(getByText('Transfer Bank'));

    const confirmButton = getByTestId('confirm-button');
    expect(confirmButton.props.style.backgroundColor).toBe('#BDBDBD');
  });

  it('should enable Confirm button when TRANSFER payment method with valid data', () => {
    const { getByTestId, getByText } = render(
      <PaymentModal {...defaultProps} selectedMethod={PaymentMethod.TRANSFER} accountName="Test Account" referenceNumber="REF123" />
    );

    fireEvent.press(getByText('Transfer Bank'));

    const confirmButton = getByTestId('confirm-button');
    expect(confirmButton.props.style.backgroundColor).toBe('#4CAF50');
  });

  it('should call onConfirm with valid transfer payment data when Confirm pressed', () => {
    const mockOnConfirm = jest.fn();
    const { getByText, getByPlaceholderText } = render(
      <PaymentModal
        {...defaultProps}
        onConfirm={mockOnConfirm}
        selectedMethod={PaymentMethod.TRANSFER}
      />
    );

    fireEvent.press(getByText('Transfer Bank'));
    fireEvent.changeText(getByPlaceholderText('Nama Akun'), 'John Doe');
    fireEvent.changeText(getByPlaceholderText('Nomor Referensi'), 'REF123');
    fireEvent.press(getByText('Konfirmasi'));

    expect(mockOnConfirm).toHaveBeenCalledWith({
      method: PaymentMethod.TRANSFER,
      accountName: 'John Doe',
      referenceNumber: 'REF123',
    });
  });

  it('should disable Confirm button when E_WALLET payment method is selected but no wallet or confirmation', () => {
    const { getByTestId, getByText } = render(
      <PaymentModal {...defaultProps} selectedMethod={PaymentMethod.E_WALLET} />
    );

    fireEvent.press(getByText('E-Wallet'));

    const confirmButton = getByTestId('confirm-button');
    expect(confirmButton.props.style.backgroundColor).toBe('#BDBDBD');
  });

  it('should enable Confirm button when E_WALLET payment method with valid data', () => {
    const { getByTestId, getByText } = render(
      <PaymentModal
        {...defaultProps}
        selectedMethod={PaymentMethod.E_WALLET}
        selectedWallet={EWalletType.GOPAY}
        confirmationInput="081234567890"
      />
    );

    fireEvent.press(getByText('E-Wallet'));

    const confirmButton = getByTestId('confirm-button');
    expect(confirmButton.props.style.backgroundColor).toBe('#4CAF50');
  });

  it('should call onConfirm with valid e-wallet payment data when Confirm pressed', () => {
    const mockOnConfirm = jest.fn();
    const { getByText, getByPlaceholderText } = render(
      <PaymentModal
        {...defaultProps}
        onConfirm={mockOnConfirm}
        selectedMethod={PaymentMethod.E_WALLET}
        selectedWallet={EWalletType.OVO}
        confirmationInput="123456"
      />
    );

    fireEvent.press(getByText('E-Wallet'));
    fireEvent.press(getByText('Konfirmasi'));

    expect(mockOnConfirm).toHaveBeenCalledWith({
      method: PaymentMethod.E_WALLET,
      walletType: EWalletType.OVO,
      confirmationInput: '123456',
    });
  });

  it('should display cart total in modal header', () => {
    const { getByText } = render(
      <PaymentModal {...defaultProps} cartTotal="250000.00" />
    );

    // formatCurrency adds non-breaking space, need to handle that
    expect(getByText(/250\.000/)).toBeTruthy();
  });
});
