/**
 * ProductList component tests
 * Tests product list rendering, filtering, and product selection
 */

import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { ProductList } from './ProductList';
import { Product } from '../types/product.types';

// Mock products for testing
const mockProducts: Product[] = [
  {
    id: 1,
    sku: 'SKU12345',
    name: 'Paracetamol 500mg',
    description: 'Pain reliever',
    stockQty: 25,
    price: '15000.00',
    costPrice: '10000.00',
    branchId: 1,
    reorderThreshold: 10,
    category: 'Obat',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    sku: 'SKU67890',
    name: 'Amoxicillin 500mg',
    description: 'Antibiotic',
    stockQty: 5,
    price: '25000.00',
    costPrice: '18000.00',
    branchId: 1,
    reorderThreshold: 10,
    category: 'Obat',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 3,
    sku: 'SKU11111',
    name: 'Vitamin C 1000mg',
    stockQty: 0,
    price: '35000.00',
    branchId: 1,
    reorderThreshold: 10,
    category: 'Vitamin',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

describe('ProductList', () => {
  it('renders empty message when no products', () => {
    const { getByText } = render(
      <ProductList products={[]} onAddToCart={jest.fn()} />
    );

    expect(getByText('No products found')).toBeTruthy();
  });

  it('renders list of products', () => {
    const { getByText } = render(
      <ProductList products={mockProducts} onAddToCart={jest.fn()} />
    );

    expect(getByText('Paracetamol 500mg')).toBeTruthy();
    expect(getByText('Amoxicillin 500mg')).toBeTruthy();
    expect(getByText('Vitamin C 1000mg')).toBeTruthy();
  });

  it('calls onAddToCart when add button is pressed', () => {
    const onAddToCartMock = jest.fn();
    const { getAllByText } = render(
      <ProductList products={mockProducts} onAddToCart={onAddToCartMock} />
    );

    const addButtons = getAllByText('Add');
    fireEvent.press(addButtons[0]);

    expect(onAddToCartMock).toHaveBeenCalledWith(mockProducts[0]);
  });

  it('filters products by search query', () => {
    const { getByText, queryByText } = render(
      <ProductList products={mockProducts} searchQuery="Paracetamol" onAddToCart={jest.fn()} />
    );

    expect(getByText('Paracetamol 500mg')).toBeTruthy();
    expect(queryByText('Amoxicillin 500mg')).toBeNull();
  });

  it('shows loading indicator when loading', () => {
    const { getByTestId } = render(
      <ProductList products={[]} loading={true} onAddToCart={jest.fn()} />
    );

    expect(getByTestId('loading-indicator')).toBeTruthy();
  });

  it('shows error message when error exists', () => {
    const { getByText } = render(
      <ProductList products={[]} error="Failed to load products" onAddToCart={jest.fn()} />
    );

    expect(getByText('Failed to load products')).toBeTruthy();
  });

  it('handles empty filtered results gracefully', () => {
    const { getByText } = render(
      <ProductList products={mockProducts} searchQuery="NonExistent" onAddToCart={jest.fn()} />
    );

    expect(getByText('No products found')).toBeTruthy();
  });

  it('renders with proper scroll view', () => {
    const { getByTestId } = render(
      <ProductList products={mockProducts} onAddToCart={jest.fn()} />
    );

    expect(getByTestId('product-list-scroll')).toBeTruthy();
  });
});
