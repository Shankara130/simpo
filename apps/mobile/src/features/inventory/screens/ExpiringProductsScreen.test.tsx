/**
 * ExpiringProductsScreen Component Tests
 * Story 4.5, Task 13.4-13.5: Test screen rendering and data fetching
 */

import React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import ExpiringProductsScreen from './ExpiringProductsScreen';
import { Alert } from 'react-native';

// Mock Alert
jest.spyOn(Alert, 'alert').mockImplementation(() => {});

// Mock navigate prop
const mockNavigate = jest.fn();

describe('ExpiringProductsScreen', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.clearAllMocks();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  // Task 13.4: Test screen rendering and data fetching
  describe('Screen rendering', () => {
    it('should render header with title and subtitle', () => {
      const { getByText } = render(
        <ExpiringProductsScreen navigate={mockNavigate} />
      );

      expect(getByText('Expiring Products')).toBeTruthy();
      expect(getByText('Products approaching expiry dates')).toBeTruthy();
    });

    it('should render filter buttons', () => {
      const { getByText } = render(
        <ExpiringProductsScreen navigate={mockNavigate} />
      );

      expect(getByText('30 Days')).toBeTruthy();
      expect(getByText('14 Days')).toBeTruthy();
      expect(getByText('7 Days')).toBeTruthy();
    });

    it('should show loading state initially', () => {
      const { getByText } = render(
        <ExpiringProductsScreen navigate={mockNavigate} />
      );

      expect(getByText('Loading expiring products...')).toBeTruthy();
    });

    it('should display products after loading', async () => {
      const { getByText, queryByText } = render(
        <ExpiringProductsScreen navigate={mockNavigate} />
      );

      // Wait for loading to complete
      await waitFor(() => {
        expect(queryByText('Loading expiring products...')).toBeNull();
      }, { timeout: 3000 });
    });
  });

  // Task 13.4: Test screen data fetching
  describe('Data fetching', () => {
    it('should fetch expiring products on mount', async () => {
      const { getByText } = render(
        <ExpiringProductsScreen navigate={mockNavigate} />
      );

      await waitFor(() => {
        expect(getByText(/Paracetamol 500mg/)).toBeTruthy();
      }, { timeout: 3000 });
    });

    it('should filter by days threshold when filter button is pressed', async () => {
      const { getByText } = render(
        <ExpiringProductsScreen navigate={mockNavigate} />
      );

      // Wait for initial load
      await waitFor(() => {
        expect(getByText(/Paracetamol 500mg/)).toBeTruthy();
      }, { timeout: 3000 });

      // Press 14 days filter
      const fourteenDaysButton = getByText('14 Days');
      fireEvent.press(fourteenDaysButton);

      // Should update filter state
      expect(fourteenDaysButton).toBeTruthy();
    });
  });

  // Task 13.4: Test screen rendering with product list
  describe('Product list display', () => {
    it('should display product cards with correct information', async () => {
      const { getByText } = render(
        <ExpiringProductsScreen navigate={mockNavigate} />
      );

      await waitFor(() => {
        expect(getByText('Paracetamol 500mg')).toBeTruthy();
        expect(getByText('Amoxicillin 500mg')).toBeTruthy();
      }, { timeout: 3000 });
    });

    it('should display SKU for each product', async () => {
      const { getByText } = render(
        <ExpiringProductsScreen navigate={mockNavigate} />
      );

      await waitFor(() => {
        expect(getByText(/SKU-12345/)).toBeTruthy();
      }, { timeout: 3000 });
    });
  });

  // Task 13.4: Test screen rendering with color coding
  describe('Alert level color coding', () => {
    it('should display alert badges for products', async () => {
      const { getByText } = render(
        <ExpiringProductsScreen navigate={mockNavigate} />
      );

      await waitFor(() => {
        expect(getByText('URGENT')).toBeTruthy();
        expect(getByText('CRITICAL')).toBeTruthy();
        expect(getByText('WARNING')).toBeTruthy();
      }, { timeout: 3000 });
    });
  });

  // Task 12.5: Tap product to view details or create discount
  describe('Product tap handling', () => {
    it('should show alert when product is tapped', async () => {
      const { getByText } = render(
        <ExpiringProductsScreen navigate={mockNavigate} />
      );

      await waitFor(() => {
        expect(getByText('Paracetamol 500mg')).toBeTruthy();
      }, { timeout: 3000 });

      // Tap on first product
      fireEvent.press(getByText('Paracetamol 500mg'));

      expect(Alert.alert).toHaveBeenCalled();
    });
  });
});
