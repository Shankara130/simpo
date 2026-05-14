/**
 * EWalletSelector Component Tests
 * Test e-wallet selection UI with validation
 */

import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { EWalletSelector } from './EWalletSelector';
import { EWalletType } from '../types/payment.types';

describe('EWalletSelector Component', () => {
  const defaultProps = {
    selectedWallet: null,
    confirmationInput: '',
    onSelectWallet: jest.fn(),
    onConfirmationInputChange: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render four e-wallet options (GoPay, OVO, Dana, ShopeePay)', () => {
    const { getAllByTestId } = render(
      <EWalletSelector {...defaultProps} />
    );

    const options = getAllByTestId(/ewallet-option-/);
    expect(options).toHaveLength(4);
  });

  it('should display correct labels for each e-wallet', () => {
    const { getByText } = render(
      <EWalletSelector {...defaultProps} />
    );

    expect(getByText('GoPay')).toBeTruthy();
    expect(getByText('OVO')).toBeTruthy();
    expect(getByText('Dana')).toBeTruthy();
    expect(getByText('ShopeePay')).toBeTruthy();
  });

  it('should call onSelectWallet with GOPAY when GoPay option is tapped', () => {
    const mockOnSelectWallet = jest.fn();
    const { getByTestId } = render(
      <EWalletSelector
        {...defaultProps}
        onSelectWallet={mockOnSelectWallet}
      />
    );

    const gopayOption = getByTestId('ewallet-option-GOPAY');
    fireEvent.press(gopayOption);

    expect(mockOnSelectWallet).toHaveBeenCalledTimes(1);
    expect(mockOnSelectWallet).toHaveBeenCalledWith(EWalletType.GOPAY);
  });

  it('should call onSelectWallet with OVO when OVO option is tapped', () => {
    const mockOnSelectWallet = jest.fn();
    const { getByTestId } = render(
      <EWalletSelector
        {...defaultProps}
        onSelectWallet={mockOnSelectWallet}
      />
    );

    const ovoOption = getByTestId('ewallet-option-OVO');
    fireEvent.press(ovoOption);

    expect(mockOnSelectWallet).toHaveBeenCalledWith(EWalletType.OVO);
  });

  it('should call onSelectWallet with DANA when Dana option is tapped', () => {
    const mockOnSelectWallet = jest.fn();
    const { getByTestId } = render(
      <EWalletSelector
        {...defaultProps}
        onSelectWallet={mockOnSelectWallet}
      />
    );

    const danaOption = getByTestId('ewallet-option-DANA');
    fireEvent.press(danaOption);

    expect(mockOnSelectWallet).toHaveBeenCalledWith(EWalletType.DANA);
  });

  it('should call onSelectWallet with SHOPEE_PAY when ShopeePay option is tapped', () => {
    const mockOnSelectWallet = jest.fn();
    const { getByTestId } = render(
      <EWalletSelector
        {...defaultProps}
        onSelectWallet={mockOnSelectWallet}
      />
    );

    const shopeeOption = getByTestId('ewallet-option-SHOPEE_PAY');
    fireEvent.press(shopeeOption);

    expect(mockOnSelectWallet).toHaveBeenCalledWith(EWalletType.SHOPEE_PAY);
  });

  it('should highlight selected e-wallet', () => {
    const { getByTestId } = render(
      <EWalletSelector
        {...defaultProps}
        selectedWallet={EWalletType.GOPAY}
      />
    );

    const gopayOption = getByTestId('ewallet-option-GOPAY');
    expect(gopayOption.props.style.borderColor).toBe('#00AED6');
  });

  it('should render confirmation input field', () => {
    const { getByPlaceholderText } = render(
      <EWalletSelector {...defaultProps} />
    );

    expect(getByPlaceholderText('Nomor Konfirmasi')).toBeTruthy();
  });

  it('should show error message when confirmation input is empty', () => {
    const { getByText } = render(
      <EWalletSelector
        {...defaultProps}
        errors={{ confirmationInput: 'Nomor konfirmasi harus diisi' }}
      />
    );

    expect(getByText('Nomor konfirmasi harus diisi')).toBeTruthy();
  });

  it('should call onConfirmationInputChange when input changes', () => {
    const mockOnConfirmationInputChange = jest.fn();
    const { getByPlaceholderText } = render(
      <EWalletSelector
        {...defaultProps}
        onConfirmationInputChange={mockOnConfirmationInputChange}
      />
    );

    const input = getByPlaceholderText('Nomor Konfirmasi');
    fireEvent.changeText(input, '081234567890');

    expect(mockOnConfirmationInputChange).toHaveBeenCalledWith('081234567890');
  });

  it('should display current confirmation input value', () => {
    const { getByDisplayValue } = render(
      <EWalletSelector
        {...defaultProps}
        confirmationInput="123456"
      />
    );

    expect(getByDisplayValue('123456')).toBeTruthy();
  });

  it('should not show error when confirmation input is valid', () => {
    const { queryByText } = render(
      <EWalletSelector
        {...defaultProps}
        selectedWallet={EWalletType.GOPAY}
        confirmationInput="081234567890"
      />
    );

    expect(queryByText('Nomor konfirmasi harus diisi')).toBeNull();
  });

  it('should show error when no e-wallet is selected', () => {
    const { getByText } = render(
      <EWalletSelector
        {...defaultProps}
        errors={{ wallet: 'Pilih e-wallet terlebih dahulu' }}
      />
    );

    expect(getByText('Pilih e-wallet terlebih dahulu')).toBeTruthy();
  });
});
