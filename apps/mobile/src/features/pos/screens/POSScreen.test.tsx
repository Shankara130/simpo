/**
 * POSScreen component tests
 * Tests main POS screen integration with all sub-components
 */

import React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
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
  return render(<CartProvider>{component}</CartProvider>);
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
});
