/**
 * PaymentMethodSelector Component Tests
 * Test payment method selection UI and interactions
 */

import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { PaymentMethodSelector } from './PaymentMethodSelector';
import { PaymentMethod } from '../types/payment.types';

describe('PaymentMethodSelector Component', () => {
  const mockOnSelectMethod = jest.fn();

  beforeEach(() => {
    mockOnSelectMethod.mockClear();
  });

  it('should render three payment method options', () => {
    const { getAllByTestId } = render(
      <PaymentMethodSelector
        selectedMethod={null}
        onSelectMethod={mockOnSelectMethod}
      />
    );

    const options = getAllByTestId(/payment-method-option/);
    expect(options).toHaveLength(3);
  });

  it('should display correct labels for each method (Tunai, Transfer Bank, E-Wallet)', () => {
    const { getByText } = render(
      <PaymentMethodSelector
        selectedMethod={null}
        onSelectMethod={mockOnSelectMethod}
      />
    );

    expect(getByText('Tunai')).toBeTruthy();
    expect(getByText('Transfer Bank')).toBeTruthy();
    expect(getByText('E-Wallet')).toBeTruthy();
  });

  it('should call onSelectMethod with CASH when Tunai option is tapped', () => {
    const { getByTestId } = render(
      <PaymentMethodSelector
        selectedMethod={null}
        onSelectMethod={mockOnSelectMethod}
      />
    );

    const cashOption = getByTestId('payment-method-option-CASH');
    fireEvent.press(cashOption);

    expect(mockOnSelectMethod).toHaveBeenCalledTimes(1);
    expect(mockOnSelectMethod).toHaveBeenCalledWith(PaymentMethod.CASH);
  });

  it('should call onSelectMethod with TRANSFER when Transfer Bank option is tapped', () => {
    const { getByTestId } = render(
      <PaymentMethodSelector
        selectedMethod={null}
        onSelectMethod={mockOnSelectMethod}
      />
    );

    const transferOption = getByTestId('payment-method-option-TRANSFER');
    fireEvent.press(transferOption);

    expect(mockOnSelectMethod).toHaveBeenCalledTimes(1);
    expect(mockOnSelectMethod).toHaveBeenCalledWith(PaymentMethod.TRANSFER);
  });

  it('should call onSelectMethod with E_WALLET when E-Wallet option is tapped', () => {
    const { getByTestId } = render(
      <PaymentMethodSelector
        selectedMethod={null}
        onSelectMethod={mockOnSelectMethod}
      />
    );

    const ewalletOption = getByTestId('payment-method-option-E_WALLET');
    fireEvent.press(ewalletOption);

    expect(mockOnSelectMethod).toHaveBeenCalledTimes(1);
    expect(mockOnSelectMethod).toHaveBeenCalledWith(PaymentMethod.E_WALLET);
  });

  it('should highlight selected payment method', () => {
    const { getByTestId } = render(
      <PaymentMethodSelector
        selectedMethod={PaymentMethod.CASH}
        onSelectMethod={mockOnSelectMethod}
      />
    );

    const cashOption = getByTestId('payment-method-option-CASH');
    // Check that selected option has green border color
    expect(cashOption.props.style.borderColor).toBe('#4CAF50');
    expect(cashOption.props.style.borderWidth).toBe(2);
    expect(cashOption.props.style.backgroundColor).toBe('#E8F5E9');
  });

  it('should not highlight unselected payment methods', () => {
    const { getByTestId } = render(
      <PaymentMethodSelector
        selectedMethod={PaymentMethod.CASH}
        onSelectMethod={mockOnSelectMethod}
      />
    );

    const transferOption = getByTestId('payment-method-option-TRANSFER');
    expect(transferOption.props.style).not.toContainEqual({
      borderColor: '#4CAF50',
      borderWidth: 2,
    });
  });

  it('should have accessibility labels for screen readers', () => {
    const { getByLabelText } = render(
      <PaymentMethodSelector
        selectedMethod={null}
        onSelectMethod={mockOnSelectMethod}
      />
    );

    expect(getByLabelText('Pilih metode pembayaran Tunai')).toBeTruthy();
    expect(getByLabelText('Pilih metode pembayaran Transfer Bank')).toBeTruthy();
    expect(getByLabelText('Pilih metode pembayaran E-Wallet')).toBeTruthy();
  });

  it('should have large touch targets (minimum 44×44px)', () => {
    const { getByTestId } = render(
      <PaymentMethodSelector
        selectedMethod={null}
        onSelectMethod={mockOnSelectMethod}
      />
    );

    const cashOption = getByTestId('payment-method-option-CASH');
    const touchTargetStyle = cashOption.props.style;

    // Check if height or minHeight meets 44px minimum
    const hasValidTouchTarget =
      touchTargetStyle?.height >= 44 || touchTargetStyle?.minHeight >= 44;

    expect(hasValidTouchTarget).toBe(true);
  });
});
