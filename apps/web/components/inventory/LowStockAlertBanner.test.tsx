/**
 * LowStockAlertBanner Component Tests
 * Story 4.4, Task 10.1-10.6: Test Low Stock Alert Banner functionality
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import LowStockAlertBanner, { LowStockAlertBannerManager } from '../LowStockAlertBanner';
import type { LowStockEvent } from '../../hooks/useStockWebSocket';

// Mock LowStockEvent for testing
const mockLowStockEvent: LowStockEvent = {
	productId: 123,
	sku: 'SKU-12345',
	productName: 'Paracetamol 500mg',
	currentStock: 5,
	reorderThreshold: 10,
	suggestedOrderQty: 15,
	branchId: 1,
	branchName: 'Jakarta Branch',
};

const mockCriticalLowStockEvent: LowStockEvent = {
	productId: 456,
	sku: 'SKU-45678',
	productName: 'Amoxicillin 500mg',
	currentStock: 0,
	reorderThreshold: 10,
	suggestedOrderQty: 20,
	branchId: 1,
	branchName: 'Jakarta Branch',
};

describe('LowStockAlertBanner', () => {
	beforeEach(() => {
		jest.useFakeTimers();
	});

	afterEach(() => {
		jest.runOnlyPendingTimers();
		jest.useRealTimers();
	});

	// Task 10.2: Test WebSocket event subscription and handling
	describe('WebSocket event handling', () => {
		it('should display alert with low stock event data', () => {
			const onDismiss = jest.fn();

			render(
				<LowStockAlertBanner
					event={mockLowStockEvent}
					onDismiss={onDismiss}
				/>
			);

			expect(screen.getByText('Low Stock Alert')).toBeInTheDocument();
			expect(screen.getByText('Paracetamol 500mg')).toBeInTheDocument();
			expect(screen.getByText('SKU-12345')).toBeInTheDocument();
			expect(screen.getByText('Jakarta Branch')).toBeInTheDocument();
		});

		it('should display current stock and threshold', () => {
			const onDismiss = jest.fn();

			render(
				<LowStockAlertBanner
					event={mockLowStockEvent}
					onDismiss={onDismiss}
				/>
			);

			expect(screen.getByText('Current: 5 / Threshold: 10')).toBeInTheDocument();
		});

		it('should display actionable message with suggested order quantity', () => {
			const onDismiss = jest.fn();

			render(
				<LowStockAlertBanner
					event={mockLowStockEvent}
					onDismiss={onDismiss}
				/>
			);

			expect(screen.getByText('Order 15 units for Jakarta Branch')).toBeInTheDocument();
		});
	});

	// Task 10.4: Test dismiss functionality
	describe('Dismiss functionality', () => {
		it('should call onDismiss when dismiss button is clicked', () => {
			const onDismiss = jest.fn();

			render(
				<LowStockAlertBanner
					event={mockLowStockEvent}
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

		it('should auto-dismiss after 30 seconds', async () => {
			const onDismiss = jest.fn();

			render(
				<LowStockAlertBanner
					event={mockLowStockEvent}
					onDismiss={onDismiss}
					autoDismissDelay={30000}
				/>
			);

			// Fast-forward 30 seconds
			jest.advanceTimersByTime(30000);

			await waitFor(() => {
				expect(onDismiss).toHaveBeenCalled();
			});
		});
	});

	// Task 10.3: Test alert display with sample event data
	describe('Severity display', () => {
		it('should show critical severity for out of stock items', () => {
			const onDismiss = jest.fn();

			const { container } = render(
				<LowStockAlertBanner
					event={mockCriticalLowStockEvent}
					onDismiss={onDismiss}
				/>
			);

			expect(container.querySelector('.bg-red-50')).toBeInTheDocument();
			expect(container.querySelector('.border-red-500')).toBeInTheDocument();
		});

		it('should show high severity for items below 50% threshold', () => {
			const onDismiss = jest.fn();

			const { container } = render(
				<LowStockAlertBanner
					event={mockLowStockEvent}
					onDismiss={onDismiss}
				/>
			);

			expect(container.querySelector('.bg-orange-50')).toBeInTheDocument();
			expect(container.querySelector('.border-orange-500')).toBeInTheDocument();
		});
	});

	// Task 10.5: Test multiple alerts handling
	describe('Multiple alerts handling', () => {
		it('should display multiple alerts in manager', () => {
			const alerts = [mockLowStockEvent, mockCriticalLowStockEvent];
			const onDismiss = jest.fn();

			render(
				<LowStockAlertBannerManager
					alerts={alerts}
					onDismiss={onDismiss}
				/>
			);

			expect(screen.getAllByText('Low Stock Alert')).toHaveLength(2);
		});

		it('should handle empty alerts array', () => {
			const onDismiss = jest.fn();

			const { container } = render(
				<LowStockAlertBannerManager
					alerts={[]}
					onDismiss={onDismiss}
				/>
			);

			expect(container.firstChild).toBeNull();
		});
	});

	// Task 10.6: Test low stock page rendering and data fetching
	describe('View Product button', () => {
		it('should call onViewProduct when View Product button is clicked', () => {
			const onDismiss = jest.fn();
			const onViewProduct = jest.fn();

			render(
				<LowStockAlertBanner
					event={mockLowStockEvent}
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
				<LowStockAlertBanner
					event={mockLowStockEvent}
					onDismiss={onDismiss}
				/>
			);

			expect(screen.queryByText('View Product')).not.toBeInTheDocument();
		});
	});
});
