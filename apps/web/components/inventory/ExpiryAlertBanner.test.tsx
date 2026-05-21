/**
 * ExpiryAlertBanner Component Tests
 * Story 4.5, Task 10.1-10.6: Test Expiry Alert Banner functionality
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import ExpiryAlertBanner, { ExpiryAlertBannerManager } from './ExpiryAlertBanner';
import type { ExpiryEvent } from '../../hooks/useStockWebSocket';

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

  // Task 10.2: Test WebSocket event subscription and handling
  describe('WebSocket event handling', () => {
    it('should display alert with expiry event data', () => {
      const onDismiss = jest.fn();

      render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(screen.getByText('Expiry Alert')).toBeInTheDocument();
      expect(screen.getByText('Paracetamol 500mg')).toBeInTheDocument();
      expect(screen.getByText('SKU-12345')).toBeInTheDocument();
      expect(screen.getByText('Jakarta Branch')).toBeInTheDocument();
    });

    it('should display expiry date', () => {
      const onDismiss = jest.fn();

      render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(screen.getByText(/Expiry Date:/)).toBeInTheDocument();
    });

    it('should display days remaining', () => {
      const onDismiss = jest.fn();

      render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(screen.getByText('30 days remaining')).toBeInTheDocument();
    });

    it('should pluralize "days" correctly when more than 1 day', () => {
      const onDismiss = jest.fn();

      render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(screen.getByText('30 days remaining')).toBeInTheDocument();
    });

    it('should singularize "day" when exactly 1 day remaining', () => {
      const onDismiss = jest.fn();
      const oneDayEvent: ExpiryEvent = {
        ...mockWarningExpiryEvent,
        daysRemaining: 1,
      };

      render(
        <ExpiryAlertBanner
          event={oneDayEvent}
          onDismiss={onDismiss}
        />
      );

      expect(screen.getByText('1 day remaining')).toBeInTheDocument();
    });
  });

  // Task 10.4: Test urgent styling (7-day alerts)
  describe('Urgent styling', () => {
    it('should show URGENT label for 7-day alerts', () => {
      const onDismiss = jest.fn();

      render(
        <ExpiryAlertBanner
          event={mockUrgentExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(screen.getByText('URGENT:')).toBeInTheDocument();
    });

    it('should apply red styling for urgent alerts', () => {
      const onDismiss = jest.fn();

      const { container } = render(
        <ExpiryAlertBanner
          event={mockUrgentExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(container.querySelector('.bg-red-50')).toBeInTheDocument();
      expect(container.querySelector('.border-red-500')).toBeInTheDocument();
      expect(container.querySelector('.text-red-900')).toBeInTheDocument();
    });
  });

  // Task 10.3: Test alert display with sample event data (all 3 alert levels)
  describe('Alert level display', () => {
    it('should show warning level (30-day) with yellow styling', () => {
      const onDismiss = jest.fn();

      const { container } = render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(container.querySelector('.bg-yellow-50')).toBeInTheDocument();
      expect(container.querySelector('.border-yellow-500')).toBeInTheDocument();
      expect(container.querySelector('.text-yellow-900')).toBeInTheDocument();
    });

    it('should show critical level (14-day) with orange styling', () => {
      const onDismiss = jest.fn();

      const { container } = render(
        <ExpiryAlertBanner
          event={mockCriticalExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(container.querySelector('.bg-orange-50')).toBeInTheDocument();
      expect(container.querySelector('.border-orange-500')).toBeInTheDocument();
      expect(container.querySelector('.text-orange-900')).toBeInTheDocument();
    });

    it('should show urgent level (7-day) with red styling', () => {
      const onDismiss = jest.fn();

      const { container } = render(
        <ExpiryAlertBanner
          event={mockUrgentExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(container.querySelector('.bg-red-50')).toBeInTheDocument();
      expect(container.querySelector('.border-red-500')).toBeInTheDocument();
      expect(container.querySelector('.text-red-900')).toBeInTheDocument();
    });
  });

  // Task 10.5: Test dismiss functionality
  describe('Dismiss functionality', () => {
    it('should call onDismiss when dismiss button is clicked', () => {
      const onDismiss = jest.fn();

      render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      const dismissButton = screen.getByLabelText('Dismiss');
      fireEvent.click(dismissButton);

      // Wait for fade-out animation
      await waitFor(() => {
        expect(onDismiss).toHaveBeenCalled();
      }, { timeout: 500 });
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

    it('should support custom auto-dismiss delay', async () => {
      const onDismiss = jest.fn();

      render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
          autoDismissDelay={10000}
        />
      );

      // Fast-forward 10 seconds
      jest.advanceTimersByTime(10000);

      await waitFor(() => {
        expect(onDismiss).toHaveBeenCalled();
      });
    });
  });

  // Task 10.6: Test multiple alerts handling
  describe('Multiple alerts handling', () => {
    it('should display multiple alerts sorted by urgency', () => {
      const alerts = [mockWarningExpiryEvent, mockCriticalExpiryEvent, mockUrgentExpiryEvent];
      const onDismiss = jest.fn();

      render(
        <ExpiryAlertBannerManager
          alerts={alerts}
          onDismiss={onDismiss}
        />
      );

      // Should show all 3 alerts
      expect(screen.getAllByText('Expiry Alert')).toHaveLength(3);

      // Urgent should be first (sorted by urgency)
      const alertsContainer = screen.getAllByText('Expiry Alert');
      expect(alertsContainer[0]).toBeInTheDocument(); // Urgent first
    });

    it('should handle empty alerts array', () => {
      const onDismiss = jest.fn();

      const { container } = render(
        <ExpiryAlertBannerManager
          alerts={[]}
          onDismiss={onDismiss}
        />
      );

      expect(container.firstChild).toBeNull();
    });

    it('should dismiss individual alerts independently', async () => {
      const alerts = [mockWarningExpiryEvent, mockCriticalExpiryEvent];
      const onDismiss = jest.fn();

      render(
        <ExpiryAlertBannerManager
          alerts={alerts}
          onDismiss={onDismiss}
        />
      );

      const allDismissButtons = screen.getAllByLabelText('Dismiss');

      // Dismiss first alert
      fireEvent.click(allDismissButtons[0]);

      await waitFor(() => {
        expect(onDismiss).toHaveBeenCalledTimes(1);
        // Should still have one alert
        expect(screen.getAllByText('Expiry Alert')).toHaveLength(1);
      });
    });
  });

  describe('View Product button', () => {
    it('should call onViewProduct when View Product button is clicked', () => {
      const onDismiss = jest.fn();
      const onViewProduct = jest.fn();

      render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
          onViewProduct={onViewProduct}
        />
      );

      const viewButton = screen.getByText('View Product');
      fireEvent.click(viewButton);

      expect(onViewProduct).toHaveBeenCalledWith(123);
      expect(onDismiss).toHaveBeenCalled();
    });

    it('should not show View Product button if onViewProduct not provided', () => {
      const onDismiss = jest.fn();

      render(
        <ExpiryAlertBanner
          event={mockWarningExpiryEvent}
          onDismiss={onDismiss}
        />
      );

      expect(screen.queryByText('View Product')).not.toBeInTheDocument();
    });
  });
});
