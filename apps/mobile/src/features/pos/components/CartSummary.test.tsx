/**
 * CartSummary component tests
 * Tests cart display, item management, and total calculations
 */

import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { CartSummary } from './CartSummary';
import { CartProvider } from '../context/CartContext';
import { CartItem } from '../types/cart.types';

// Wrapper to provide CartContext
const renderWithCartProvider = (component: React.ReactElement) => {
  return render(<CartProvider>{component}</CartProvider>);
};

// Mock cart items for testing
const mockCartItems: CartItem[] = [
  {
    productId: 1,
    sku: 'SKU12345',
    name: 'Paracetamol 500mg',
    price: '15000.00',
    quantity: 2,
    subtotal: '30000.00',
  },
  {
    productId: 2,
    sku: 'SKU67890',
    name: 'Amoxicillin 500mg',
    price: '25000.00',
    quantity: 1,
    subtotal: '25000.00',
  },
];

describe('CartSummary', () => {
  it('renders empty cart message when cart is empty', () => {
    const { getByText } = renderWithCartProvider(<CartSummary />);
    expect(getByText('Cart is empty')).toBeTruthy();
  });

  it('renders cart items when cart has items', () => {
    const { getByText } = renderWithCartProvider(
      <CartSummary items={mockCartItems} total="55000.00" itemCount={3} />
    );

    expect(getByText('Paracetamol 500mg')).toBeTruthy();
    expect(getByText('Amoxicillin 500mg')).toBeTruthy();
  });

  it('displays item quantity and subtotal correctly', () => {
    const { getByText } = renderWithCartProvider(
      <CartSummary items={mockCartItems} total="55000.00" itemCount={3} />
    );

    expect(getByText('2')).toBeTruthy(); // Quantity shown in quantity control
    expect(getByText('Rp 30.000')).toBeTruthy(); // Subtotal for 2 x Rp 15.000
    expect(getByText('Rp 25.000')).toBeTruthy(); // Subtotal for 1 x Rp 25.000
  });

  it('displays total correctly', () => {
    const { getByText } = renderWithCartProvider(
      <CartSummary items={mockCartItems} total="55000.00" itemCount={3} />
    );

    expect(getByText('Total: Rp 55.000')).toBeTruthy();
  });

  it('calls onRemove when remove button is pressed', () => {
    const onRemoveMock = jest.fn();
    const { getAllByText } = renderWithCartProvider(
      <CartSummary items={mockCartItems} total="55000.00" itemCount={3} onRemove={onRemoveMock} />
    );

    const removeButtons = getAllByText('Remove');
    fireEvent.press(removeButtons[0]);

    expect(onRemoveMock).toHaveBeenCalledWith(1);
  });

  it('calls onUpdateQuantity when + button is pressed', () => {
    const onUpdateQuantityMock = jest.fn();
    const { getAllByText } = renderWithCartProvider(
      <CartSummary items={mockCartItems} total="55000.00" itemCount={3} onUpdateQuantity={onUpdateQuantityMock} />
    );

    const increaseButtons = getAllByText('+');
    fireEvent.press(increaseButtons[0]);

    expect(onUpdateQuantityMock).toHaveBeenCalledWith(1, 3);
  });

  it('calls onUpdateQuantity when - button is pressed', () => {
    const onUpdateQuantityMock = jest.fn();
    const { getAllByText } = renderWithCartProvider(
      <CartSummary items={mockCartItems} total="55000.00" itemCount={3} onUpdateQuantity={onUpdateQuantityMock} />
    );

    const decreaseButtons = getAllByText('-');
    fireEvent.press(decreaseButtons[0]);

    expect(onUpdateQuantityMock).toHaveBeenCalledWith(1, 1);
  });

  it('renders with proper styling and layout', () => {
    const { getByTestId } = renderWithCartProvider(
      <CartSummary items={mockCartItems} total="55000.00" itemCount={3} />
    );

    const cartContainer = getByTestId('cart-summary-container');
    expect(cartContainer).toBeTruthy();
  });
});
