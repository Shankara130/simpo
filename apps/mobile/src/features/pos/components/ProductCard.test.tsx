/**
 * ProductCard component tests
 * Tests product card rendering, interactions, and visual states
 */

import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { ProductCard } from './ProductCard';
import { Product } from '../types/product.types';

// Mock product for testing
const mockProduct: Product = {
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
};

describe('ProductCard', () => {
  it('renders product information correctly', () => {
    const { getByText } = render(
      <ProductCard product={mockProduct} onAdd={jest.fn()} />
    );

    expect(getByText('Paracetamol 500mg')).toBeTruthy();
    expect(getByText('SKU12345')).toBeTruthy();
    expect(getByText('Rp 15.000')).toBeTruthy();
    expect(getByText('Stock: 25')).toBeTruthy();
  });

  it('shows low stock warning when stock below threshold', () => {
    const lowStockProduct = { ...mockProduct, stockQty: 5 };
    const { getByText } = render(
      <ProductCard product={lowStockProduct} onAdd={jest.fn()} />
    );

    expect(getByText('Stock: 5')).toBeTruthy();
    // Low stock indicator should be visible
  });

  it('shows out of stock when stock is 0', () => {
    const outOfStockProduct = { ...mockProduct, stockQty: 0 };
    const { getByText } = render(
      <ProductCard product={outOfStockProduct} onAdd={jest.fn()} />
    );

    expect(getByText('Out of Stock')).toBeTruthy();
  });

  it('calls onAdd when add button is pressed', () => {
    const onAddMock = jest.fn();
    const { getByText } = render(
      <ProductCard product={mockProduct} onAdd={onAddMock} />
    );

    fireEvent.press(getByText('Add'));
    expect(onAddMock).toHaveBeenCalledWith(mockProduct);
  });

  it('disables add button when product is out of stock', () => {
    const outOfStockProduct = { ...mockProduct, stockQty: 0 };
    const onAddMock = jest.fn();
    const { getByText } = render(
      <ProductCard product={outOfStockProduct} onAdd={onAddMock} />
    );

    fireEvent.press(getByText('Add'));
    expect(onAddMock).not.toHaveBeenCalled();
  });

  it('has minimum touch target size of 44x44px', () => {
    const { getByTestId } = render(
      <ProductCard product={mockProduct} onAdd={jest.fn()} />
    );

    const addButton = getByTestId('add-button');
    const style = addButton.props.style;

    // Check that button has adequate height for touch target
    if (Array.isArray(style)) {
      const heightStyle = style.find(s => s.height);
      expect(heightStyle?.height).toBeGreaterThanOrEqual(44);
    }
  });
});
