/**
 * Tests for CartList component
 * Tests cart items display, empty state, and loading state
 */

import React from 'react';
import { render, waitFor } from '@testing-library/react-native';
import { CartList } from './CartList';
import { CartItem } from '../types/cart.types';

// Mock dependencies
jest.mock('../utils/formatCurrency', () => ({
  formatCurrency: (amount: number) => `Rp ${amount.toLocaleString('id-ID')}`,
}));

describe('CartList', () => {
  const mockCartItems: CartItem[] = [
    {
      productId: 1,
      sku: 'SKU001',
      name: 'Paracetamol 500mg',
      price: '15000.00',
      quantity: 2,
      subtotal: '30000.00',
      stockQty: 50,
    },
    {
      productId: 2,
      sku: 'SKU002',
      name: 'Amoxicillin 500mg',
      price: '25000.00',
      quantity: 1,
      subtotal: '25000.00',
      stockQty: 30,
    },
  ];

  const mockOnUpdateQuantity = jest.fn();
  const mockOnRemoveItem = jest.fn();

  describe('when cart has items', () => {
    it('should render list of cart items correctly', () => {
      const { getByText, getAllByTestId } = render(
        <CartList
          cartItems={mockCartItems}
          onUpdateQuantity={mockOnUpdateQuantity}
          onRemoveItem={mockOnRemoveItem}
        />
      );

      // Check if product names are displayed
      expect(getByText('Paracetamol 500mg')).toBeTruthy();
      expect(getByText('Amoxicillin 500mg')).toBeTruthy();

      // Check if SKUs are displayed (with "SKU:" prefix)
      expect(getByText('SKU: SKU001')).toBeTruthy();
      expect(getByText('SKU: SKU002')).toBeTruthy();

      // Check if quantities are displayed
      expect(getByText('2')).toBeTruthy();
      expect(getByText('1')).toBeTruthy();
    });

    it('should display item details (name, SKU, qty, price, subtotal)', () => {
      const { getByText, getByTestId } = render(
        <CartList
          cartItems={mockCartItems}
          onUpdateQuantity={mockOnUpdateQuantity}
          onRemoveItem={mockOnRemoveItem}
        />
      );

      // Product name
      expect(getByText('Paracetamol 500mg')).toBeTruthy();

      // SKU (with prefix)
      expect(getByText('SKU: SKU001')).toBeTruthy();

      // Price (formatted)
      expect(getByText('Rp 15.000')).toBeTruthy();

      // Quantity
      expect(getByText('2')).toBeTruthy();

      // Subtotal (via testID)
      expect(getByTestId('subtotal-1')).toBeTruthy();
    });

    it('should handle scroll behavior with FlatList', async () => {
      // Create more items to test scrolling
      const manyItems: CartItem[] = Array.from({ length: 50 }, (_, i) => ({
        productId: i + 1,
        sku: `SKU${i + 1}`,
        name: `Product ${i + 1}`,
        price: '10000.00',
        quantity: 1,
        subtotal: '10000.00',
        stockQty: 100,
      }));

      const { getByText } = render(
        <CartList
          cartItems={manyItems}
          onUpdateQuantity={mockOnUpdateQuantity}
          onRemoveItem={mockOnRemoveItem}
        />
      );

      // Check if first item is rendered
      expect(getByText('Product 1')).toBeTruthy();

      // Check if a later item is rendered (FlatList should handle this)
      await waitFor(() => {
        expect(getByText('Product 10')).toBeTruthy();
      });
    });

    it('should update when cart items change', () => {
      const { rerender, getByText, queryByText } = render(
        <CartList
          cartItems={mockCartItems}
          onUpdateQuantity={mockOnUpdateQuantity}
          onRemoveItem={mockOnRemoveItem}
        />
      );

      // Initial items
      expect(getByText('Paracetamol 500mg')).toBeTruthy();
      expect(getByText('Amoxicillin 500mg')).toBeTruthy();

      // Update with fewer items
      const fewerItems = [mockCartItems[0]];
      rerender(
        <CartList
          cartItems={fewerItems}
          onUpdateQuantity={mockOnUpdateQuantity}
          onRemoveItem={mockOnRemoveItem}
        />
      );

      // First item still there
      expect(getByText('Paracetamol 500mg')).toBeTruthy();

      // Second item removed
      expect(queryByText('Amoxicillin 500mg')).toBeNull();
    });
  });

  describe('when cart is empty', () => {
    it('should render empty state with Indonesian message', () => {
      const { getByText } = render(
        <CartList
          cartItems={[]}
          onUpdateQuantity={mockOnUpdateQuantity}
          onRemoveItem={mockOnRemoveItem}
        />
      );

      expect(getByText('Keranjang masih kosong')).toBeTruthy();
    });

    it('should not display cart items when empty', () => {
      const { queryByTestId } = render(
        <CartList
          cartItems={[]}
          onUpdateQuantity={mockOnUpdateQuantity}
          onRemoveItem={mockOnRemoveItem}
        />
      );

      // No cart item should be rendered
      expect(queryByTestId(/cart-item/i)).toBeNull();
    });
  });

  describe('loading state', () => {
    it('should show loading indicator when loading is true', () => {
      const { getByTestId } = render(
        <CartList
          cartItems={mockCartItems}
          onUpdateQuantity={mockOnUpdateQuantity}
          onRemoveItem={mockOnRemoveItem}
          loading={true}
        />
      );

      expect(getByTestId('cart-loading-indicator')).toBeTruthy();
    });

    it('should not show loading indicator when loading is false', () => {
      const { queryByTestId } = render(
        <CartList
          cartItems={mockCartItems}
          onUpdateQuantity={mockOnUpdateQuantity}
          onRemoveItem={mockOnRemoveItem}
          loading={false}
        />
      );

      expect(queryByTestId('cart-loading-indicator')).toBeNull();
    });

    it('should show loading state during cart initialization', () => {
      const { getByTestId, getByText } = render(
        <CartList
          cartItems={[]}
          onUpdateQuantity={mockOnUpdateQuantity}
          onRemoveItem={mockOnRemoveItem}
          loading={true}
        />
      );

      // Loading indicator should be shown instead of empty state
      expect(getByTestId('cart-loading-indicator')).toBeTruthy();
      expect(() => getByText('Keranjang masih kosong')).toThrow();
    });
  });
});
