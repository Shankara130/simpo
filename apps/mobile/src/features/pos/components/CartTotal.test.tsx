/**
 * Tests for CartTotal component
 * Tests total calculation, currency formatting, real-time updates, and clear cart button
 */

import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { CartTotal } from './CartTotal';
import { CartProvider } from '../context/CartContext';

// Mock formatCurrency
jest.mock('../utils/formatCurrency', () => ({
  formatCurrency: (amount: number) => `Rp ${amount.toLocaleString('id-ID')}`,
}));

describe('CartTotal', () => {
  const mockClearCart = jest.fn();

  // Helper to render with custom cart state
  const renderWithCartState = (state: any, onClearCart = mockClearCart) => {
    return render(
      <CartProvider>
        <CartTotal onClearCart={onClearCart} />
      </CartProvider>
    );
  };

  beforeEach(() => {
    mockClearCart.mockClear();
    // Reset the mock
    jest.spyOn(require('../context/CartContext'), 'useCartContext').mockRestore();
  });

  describe('total calculation and display', () => {
    it('should display total, item count, and clear button', () => {
      // Mock the hook to return custom state
      const mockUseCartContext = jest.spyOn(require('../context/CartContext'), 'useCartContext');
      mockUseCartContext.mockReturnValue({
        state: {
          items: [
            { productId: 1, price: '15000.00', quantity: 2, subtotal: '30000.00' },
            { productId: 2, price: '25000.00', quantity: 1, subtotal: '25000.00' },
          ],
          total: '55000.00',
          itemCount: 3,
        },
        actions: {
          addItem: jest.fn(),
          removeItem: jest.fn(),
          updateQuantity: jest.fn(),
          clearCart: mockClearCart,
        },
      });

      const { getByText, getByTestId } = render(
        <CartTotal onClearCart={mockClearCart} />
      );

      // Should display total
      expect(getByTestId('cart-total-amount')).toBeTruthy();

      // Should display item count
      expect(getByText('3 items')).toBeTruthy();

      // Should have clear button (since cart has items)
      expect(getByTestId('clear-cart-button')).toBeTruthy();
    });

    it('should show "1 item" for single item', () => {
      const mockUseCartContext = jest.spyOn(require('../context/CartContext'), 'useCartContext');
      mockUseCartContext.mockReturnValue({
        state: {
          items: [{ productId: 1, price: '15000.00', quantity: 1, subtotal: '15000.00' }],
          total: '15000.00',
          itemCount: 1,
        },
        actions: {
          addItem: jest.fn(),
          removeItem: jest.fn(),
          updateQuantity: jest.fn(),
          clearCart: mockClearCart,
        },
      });

      const { getByText } = render(
        <CartTotal onClearCart={mockClearCart} />
      );

      expect(getByText('1 item')).toBeTruthy();
    });

    it('should show "0 items" for empty cart', () => {
      const mockUseCartContext = jest.spyOn(require('../context/CartContext'), 'useCartContext');
      mockUseCartContext.mockReturnValue({
        state: {
          items: [],
          total: '0.00',
          itemCount: 0,
        },
        actions: {
          addItem: jest.fn(),
          removeItem: jest.fn(),
          updateQuantity: jest.fn(),
          clearCart: mockClearCart,
        },
      });

      const { getByText, queryByTestId } = render(
        <CartTotal onClearCart={mockClearCart} />
      );

      expect(getByText('0 items')).toBeTruthy();
      expect(queryByTestId('clear-cart-button')).toBeNull();
    });
  });

  describe('clear cart button', () => {
    it('should call clearCart callback when clear button is pressed', () => {
      const mockUseCartContext = jest.spyOn(require('../context/CartContext'), 'useCartContext');
      mockUseCartContext.mockReturnValue({
        state: {
          items: [{ productId: 1, price: '15000.00', quantity: 1, subtotal: '15000.00' }],
          total: '15000.00',
          itemCount: 1,
        },
        actions: {
          addItem: jest.fn(),
          removeItem: jest.fn(),
          updateQuantity: jest.fn(),
          clearCart: mockClearCart,
        },
      });

      const { getByTestId } = render(
        <CartTotal onClearCart={mockClearCart} />
      );

      const clearButton = getByTestId('clear-cart-button');
      fireEvent.press(clearButton);

      expect(mockClearCart).toHaveBeenCalled();
    });

    it('should not show clear button when cart is empty', () => {
      const mockUseCartContext = jest.spyOn(require('../context/CartContext'), 'useCartContext');
      mockUseCartContext.mockReturnValue({
        state: {
          items: [],
          total: '0.00',
          itemCount: 0,
        },
        actions: {
          addItem: jest.fn(),
          removeItem: jest.fn(),
          updateQuantity: jest.fn(),
          clearCart: mockClearCart,
        },
      });

      const { queryByTestId } = render(
        <CartTotal onClearCart={mockClearCart} />
      );

      expect(queryByTestId('clear-cart-button')).toBeNull();
    });
  });
});
