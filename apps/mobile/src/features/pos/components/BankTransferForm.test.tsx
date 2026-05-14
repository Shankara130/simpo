/**
 * BankTransferForm Component Tests
 * Test bank transfer input form with validation
 */

import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { BankTransferForm } from './BankTransferForm';

describe('BankTransferForm Component', () => {
  const defaultProps = {
    accountName: '',
    referenceNumber: '',
    onAccountNameChange: jest.fn(),
    onReferenceNumberChange: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render account name input field', () => {
    const { getByPlaceholderText } = render(
      <BankTransferForm {...defaultProps} />
    );

    expect(getByPlaceholderText('Nama Akun')).toBeTruthy();
  });

  it('should render reference number input field', () => {
    const { getByPlaceholderText } = render(
      <BankTransferForm {...defaultProps} />
    );

    expect(getByPlaceholderText('Nomor Referensi')).toBeTruthy();
  });

  it('should show error message when account name is empty (< 2 chars)', () => {
    const { getByText, getByPlaceholderText } = render(
      <BankTransferForm
        {...defaultProps}
        accountName="A"
        errors={{ accountName: 'Nama akun minimal 2 karakter' }}
      />
    );

    expect(getByText('Nama akun minimal 2 karakter')).toBeTruthy();
  });

  it('should show error message when reference number is empty (< 3 chars)', () => {
    const { getByText } = render(
      <BankTransferForm
        {...defaultProps}
        referenceNumber="AB"
        errors={{ referenceNumber: 'Nomor referensi minimal 3 karakter' }}
      />
    );

    expect(getByText('Nomor referensi minimal 3 karakter')).toBeTruthy();
  });

  it('should call onAccountNameChange when account name changes', () => {
    const mockOnAccountNameChange = jest.fn();
    const { getByPlaceholderText } = render(
      <BankTransferForm
        {...defaultProps}
        onAccountNameChange={mockOnAccountNameChange}
      />
    );

    const input = getByPlaceholderText('Nama Akun');
    fireEvent.changeText(input, 'John Doe');

    expect(mockOnAccountNameChange).toHaveBeenCalledWith('John Doe');
  });

  it('should call onReferenceNumberChange when reference number changes', () => {
    const mockOnReferenceNumberChange = jest.fn();
    const { getByPlaceholderText } = render(
      <BankTransferForm
        {...defaultProps}
        onReferenceNumberChange={mockOnReferenceNumberChange}
      />
    );

    const input = getByPlaceholderText('Nomor Referensi');
    fireEvent.changeText(input, 'REF123');

    expect(mockOnReferenceNumberChange).toHaveBeenCalledWith('REF123');
  });

  it('should display current account name value', () => {
    const { getByDisplayValue } = render(
      <BankTransferForm {...defaultProps} accountName="Test Account" />
    );

    expect(getByDisplayValue('Test Account')).toBeTruthy();
  });

  it('should display current reference number value', () => {
    const { getByDisplayValue } = render(
      <BankTransferForm {...defaultProps} referenceNumber="REF001" />
    );

    expect(getByDisplayValue('REF001')).toBeTruthy();
  });

  it('should not show errors when fields are valid', () => {
    const { queryByText } = render(
      <BankTransferForm
        {...defaultProps}
        accountName="Valid Name"
        referenceNumber="REF123"
      />
    );

    expect(queryByText('Nama akun minimal 2 karakter')).toBeNull();
    expect(queryByText('Nomor referensi minimal 3 karakter')).toBeNull();
  });
});
