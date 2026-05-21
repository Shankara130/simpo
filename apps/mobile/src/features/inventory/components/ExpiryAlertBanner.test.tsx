/**
 * ExpiryAlertBanner Component Tests
 * Story 4.5, Task 13.1-13.3: Test banner rendering with all alert levels
 */

import React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { ExpiryAlertBanner, ExpiryAlertBannerContainer, useExpiryAlerts } from './ExpiryAlertBanner';
import { ExpiryEvent } from '../services/realTimeStockService';

// Mock ExpiryEvent for testing
const mockWarningExpiryEvent: ExpiryEvent = {
  productId: 123,
  sku: 'SKU-12345',
  productName: 'Paracetamol 500mg',
  expiryDate: '2026-06-20T00:00:00Z',
  daysRemaining: 30,
  alertLevel: 'warning',
  branchId: 1,
  branchName: 'Jakarta Branch',
};

const mockCriticalExpiryEvent: ExpiryEvent = {
  productId: 456,
  sku: 'SKU-45678',
  productName: 'Amoxicillin 500mg',
  expiryDate: '2026-06-04T00:00:00Z',
  daysRemaining: 14,
  alertLevel: 'critical',
  branchId: 1,
  branchName: 'Jakarta Branch',
};

const mockUrgentExpiryEvent: ExpiryEvent = {
  productId: 789,
  sku: 'SKU-78901',
  productName: 'Ibuprofen 400mg',
  expiryDate: '2026-05-28T00:00:00Z',
  daysRemaining: 7,
  alertLevel: 'urgent',
  branchId: 2,
  branchName: 'Bandung Branch',
};

describe('ExpiryAlertBanner', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  // Task 13.1: Test banner rendering with color coding
  describe('Banner rendering with all alert levels', () => {
    it('should display warning level (30-day) with product info', () => {
      const onDismiss = jest.fn();

      const { getByText } = render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(getByText('Expiry Alert')).toBeTruthy();
      expect(getByText('Paracetamol 500mg')).toBeTruthy();
    });

    it('should display critical level (14-day)', () => {
      const onDismiss = jest.fn();

      const { getByText } = render(
        <ExpiryAlertBanner
          event={mockCriticalExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(getByText('Expiry Alert')).toBeTruthy();
      expect(getByText('Amoxicillin 500mg')).toBeTruthy();
    });

    it('should display urgent level (7-day) with urgent label', () => {
      const onDismiss = jest.fn();

      const { getByText } = render(
        <ExpiryAlertBanner
          event={mockUrgentExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(getByText('URGENT:')).toBeTruthy();
      expect(getByText('Expiry Alert')).toBeTruthy();
    });

    it('should display SKU correctly', () => {
      const onDismiss = jest.fn();

      const { getByText } = render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(getByText(/SKU-12345/)).toBeTruthy();
    });
  });

  describe('Dismiss functionality', () => {
    it('should call onDismiss when dismiss button is pressed', async () => {
      const onDismiss = jest.fn();

      const { getByText } = render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      const dismissButton = getByText('✕');
      fireEvent.press(dismissButton);

      await waitFor(() => {
        expect(onDismiss).toHaveBeenCalled();
      });
    });

    it('should auto-dismiss after 60 seconds', async () => {
      const onDismiss = jest.fn();

      render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
          autoDismissDelay={60000}
        />
      );

      // Fast-forward 60 seconds
      jest.advanceTimersByTime(60000);

      await waitFor(() => {
        expect(onDismiss).toHaveBeenCalled();
      });
    });
  });

  describe('View Product button', () => {
    it('should call onViewProduct when View button is pressed', () => {
      const onDismiss = jest.fn();
      const onViewProduct = jest.fn();

      const { getByText } = render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
          onViewProduct={onViewProduct}
        />
      );

      const viewButton = getByText('View');
      fireEvent.press(viewButton);

      expect(onViewProduct).toHaveBeenCalledWith(123);
      expect(onDismiss).toHaveBeenCalled();
    });
  });
});

describe('ExpiryAlertBannerContainer', () => {
  it('should display multiple alerts', () => {
    const alerts = [
      { id: '1', event: mockWarningExpiryEvent },
      { id: '2', event: mockCriticalExpiryEvent },
    ];
    const onDismiss = jest.fn();

    const { getByText } = render(
      <ExpiryAlertBannerContainer
        alerts={alerts}
        onDismiss={onDismiss}
      />
    );

    expect(getByText('Paracetamol 500mg')).toBeTruthy();
  });

  it('should handle empty alerts array', () => {
    const onDismiss = jest.fn();

    const { UNSAFE_getByType } = render(
      <ExpiryAlertBannerContainer
        alerts={[]}
        onDismiss={onDismiss}
      />
    );

    // Container should not render any banners
    expect(() => UNSAFE_getByType('View')).toThrow();
  });
});

describe('useExpiryAlerts hook', () => {
  it('should add alert when addAlert is called', () => {
    const { result } = renderHook(() => useExpiryAlerts());

    act(() => {
      result.current.addAlert(mockWarningExpiryEvent);
    });

    expect(result.current.alerts.length).toBe(1);
  });

  it('should remove alert when removeAlert is called', () => {
    const { result } = renderHook(() => useExpiryAlerts());

    act(() => {
      result.current.addAlert(mockWarningExpiryEvent);
      const alerts = result.current.alerts;
      if (alerts.length > 0) {
        result.current.removeAlert(alerts[0].id);
      }
    });

    expect(result.current.alerts.length).toBe(0);
  });

  it('should clear all alerts when clearAll is called', () => {
    const { result } = renderHook(() => useExpiryAlerts());

    act(() => {
      result.current.addAlert(mockWarningExpiryEvent);
      result.current.addAlert(mockCriticalExpiryEvent);
      result.current.clearAll();
    });

    expect(result.current.alerts.length).toBe(0);
  });
});
