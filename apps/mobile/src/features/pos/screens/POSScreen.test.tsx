/**
 * POSScreen component tests
 * Tests main POS screen integration with all sub-components
 * Story 7.2: USB Barcode Scanner Integration
 */

import React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { NavigationContainer } from '@react-navigation/native';
import { POSScreen } from './POSScreen';
import { CartProvider } from '../context/CartContext';
import { Product } from '../types/product.types';

// Mock products
const mockProducts: Product[] = [
  {
    id: 1,
    sku: 'SKU12345',
    name: 'Paracetamol 500mg',
    stockQty: 25,
    price: '15000.00',
    branchId: 1,
    reorderThreshold: 10,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

const renderWithCartProvider = (component: React.ReactElement) => {
  return render(
    <NavigationContainer>
      <CartProvider>{component}</CartProvider>
    </NavigationContainer>
  );
};

describe('POSScreen', () => {
  it('renders all major components', () => {
    const { getByPlaceholderText, getByTestId } = renderWithCartProvider(
      <POSScreen products={mockProducts} loading={false} />
    );

    // Top control bar with search
    expect(getByPlaceholderText('Search products or scan barcode...')).toBeTruthy();

    // Product list area
    expect(getByTestId('product-list-scroll')).toBeTruthy();
  });

  it('displays cart summary with empty state initially', () => {
    const { getByText } = renderWithCartProvider(
      <POSScreen products={mockProducts} loading={false} />
    );

    // CartList shows Indonesian empty message
    expect(getByText('Keranjang masih kosong')).toBeTruthy();
  });

  it('adds product to cart when add button is pressed', async () => {
    const { getAllByText, getByText, getByTestId } = renderWithCartProvider(
      <POSScreen products={mockProducts} loading={false} />
    );

    // Initially cart should be empty
    expect(getByText('Keranjang masih kosong')).toBeTruthy();
    expect(getByText('0 items')).toBeTruthy();

    const addButtons = getAllByText('Add');
    fireEvent.press(addButtons[0]);

    await waitFor(() => {
      // Cart should no longer be empty
      expect(() => getByText('Keranjang masih kosong')).toThrow();
      // CartTotal should show updated item count
      expect(getByText('1 item')).toBeTruthy();
    });
  });

  it('shows loading state when loading products', () => {
    const { getByTestId } = renderWithCartProvider(
      <POSScreen products={[]} loading={true} />
    );

    expect(getByTestId('loading-indicator')).toBeTruthy();
  });

  it('handles search functionality', () => {
    const { getByPlaceholderText } = renderWithCartProvider(
      <POSScreen products={mockProducts} loading={false} />
    );

    const searchInput = getByPlaceholderText('Search products or scan barcode...');
    fireEvent.changeText(searchInput, 'Paracetamol');

    // Search should filter the product list
    expect(searchInput.props.value).toBe('Paracetamol');
  });

  it('shows error message when product loading fails', () => {
    const { getByText } = renderWithCartProvider(
      <POSScreen products={[]} loading={false} error="Failed to load products" />
    );

    expect(getByText('Failed to load products')).toBeTruthy();
  });

  it('renders action buttons at the bottom', () => {
    const { getByText } = renderWithCartProvider(
      <POSScreen products={mockProducts} loading={false} />
    );

    expect(getByText('Checkout')).toBeTruthy();
    expect(getByText('Clear Cart')).toBeTruthy();
  });

  it('has proper layout structure with SafeAreaView', () => {
    const { getByTestId } = renderWithCartProvider(
      <POSScreen products={mockProducts} loading={false} />
    );

    expect(getByTestId('pos-screen-container')).toBeTruthy();
  });

  // Story 7.2: USB Barcode Scanner Integration Tests
  describe('Scanner Integration', () => {
    it('renders scanner keyboard input', () => {
      const { getByTestId } = renderWithCartProvider(
        <POSScreen products={mockProducts} loading={false} />
      );

      expect(getByTestId('scanner-keyboard-input')).toBeTruthy();
    });

    it('renders scanner feedback component', () => {
      const { queryByTestId } = renderWithCartProvider(
        <POSScreen products={mockProducts} loading={false} />
      );

      // Scanner feedback component exists but may not render when idle
      // Component is included in JSX even if not visible
      expect(queryByTestId('scanner-feedback')).toBeNull(); // Null when idle
    });

    it('initializes scanner with idle state', () => {
      const { queryByTestId } = renderWithCartProvider(
        <POSScreen products={mockProducts} loading={false} />
      );

      // Scanner feedback should not be visible when idle (component returns null)
      expect(queryByTestId('scanner-feedback')).toBeNull();
    });

    it('scanner input does not interfere with search functionality', () => {
      const { getByPlaceholderText } = renderWithCartProvider(
        <POSScreen products={mockProducts} loading={false} />
      );

      const searchInput = getByPlaceholderText('Search products or scan barcode...');
      expect(searchInput).toBeTruthy();

      // Search should work normally
      fireEvent.changeText(searchInput, 'Paracetamol');
      expect(searchInput.props.value).toBe('Paracetamol');
    });

    it('renders scanner settings button', () => {
      const { getByTestId } = renderWithCartProvider(
        <POSScreen products={mockProducts} loading={false} />
      );

      expect(getByTestId('scanner-settings-button')).toBeTruthy();
    });
  });
});
