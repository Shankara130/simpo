/**
 * TopControlBar component tests
 * Tests search functionality, cart summary display, and actions
 */

import React from 'react';
import { render, fireEvent, changeTextInputValue } from '@testing-library/react-native';
import { TopControlBar } from './TopControlBar';

describe('TopControlBar', () => {
  it('renders search input with placeholder', () => {
    const { getByPlaceholderText } = render(
      <TopControlBar
        itemCount={0}
        total="0.00"
        onSearch={jest.fn()}
        onPayment={jest.fn()}
      />
    );

    expect(getByPlaceholderText('Search products or scan barcode...')).toBeTruthy();
  });

  it('displays cart item count correctly', () => {
    const { getByText } = render(
      <TopControlBar
        itemCount={3}
        total="45000.00"
        onSearch={jest.fn()}
        onPayment={jest.fn()}
      />
    );

    expect(getByText('Cart: 3 items')).toBeTruthy();
  });

  it('displays cart total correctly', () => {
    const { getByText } = render(
      <TopControlBar
        itemCount={2}
        total="35000.00"
        onSearch={jest.fn()}
        onPayment={jest.fn()}
      />
    );

    expect(getByText('Rp 35.000')).toBeTruthy();
  });

  it('calls onSearch when search text changes', () => {
    const onSearchMock = jest.fn();
    const { getByPlaceholderText } = render(
      <TopControlBar
        itemCount={0}
        total="0.00"
        onSearch={onSearchMock}
        onPayment={jest.fn()}
      />
    );

    const searchInput = getByPlaceholderText('Search products or scan barcode...');
    fireEvent.changeText(searchInput, 'Paracetamol');

    expect(onSearchMock).toHaveBeenCalledWith('Paracetamol');
  });

  it('calls onPayment when payment button is pressed', () => {
    const onPaymentMock = jest.fn();
    const { getByText } = render(
      <TopControlBar
        itemCount={1}
        total="15000.00"
        onSearch={jest.fn()}
        onPayment={onPaymentMock}
      />
    );

    fireEvent.press(getByText('Pay'));

    expect(onPaymentMock).toHaveBeenCalled();
  });

  it('disables payment button when cart is empty', () => {
    const onPaymentMock = jest.fn();
    const { getByTestId } = render(
      <TopControlBar
        itemCount={0}
        total="0.00"
        onSearch={jest.fn()}
        onPayment={onPaymentMock}
      />
    );

    const payButton = getByTestId('payment-button');
    fireEvent.press(payButton);

    // Button should not call onPayment when disabled
    expect(onPaymentMock).not.toHaveBeenCalled();
  });

  it('enables payment button when cart has items', () => {
    const onPaymentMock = jest.fn();
    const { getByTestId } = render(
      <TopControlBar
        itemCount={1}
        total="15000.00"
        onSearch={jest.fn()}
        onPayment={onPaymentMock}
      />
    );

    const payButton = getByTestId('payment-button');
    fireEvent.press(payButton);

    // Button should call onPayment when enabled
    expect(onPaymentMock).toHaveBeenCalled();
  });

  it('has minimum touch target size for payment button', () => {
    const { getByTestId } = render(
      <TopControlBar
        itemCount={1}
        total="15000.00"
        onSearch={jest.fn()}
        onPayment={jest.fn()}
      />
    );

    const payButton = getByTestId('payment-button');
    const style = payButton.props.style;

    if (Array.isArray(style)) {
      const heightStyle = style.find(s => s.minHeight);
      expect(heightStyle?.minHeight).toBeGreaterThanOrEqual(44);
    }
  });
});
