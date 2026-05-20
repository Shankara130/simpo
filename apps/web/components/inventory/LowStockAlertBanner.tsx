/**
 * LowStockAlertBanner Component
 * Story 4.4, Task 7: Create Low Stock Alert Banner Component (AC: 3, 4, 5)
 *
 * This component displays alert banners when products fall below their reorder thresholds.
 * Features:
 * - Real-time low stock notifications via WebSocket
 * - Product info display (SKU, name, current stock vs threshold)
 * - Actionable message: "Order {qty} units of {product} for {branch}"
 * - "View Product" and "Dismiss" buttons
 * - Multiple alerts support (stacked)
 * - Auto-dismiss after 30 seconds or manual dismiss
 */

'use client';

import { useState, useEffect } from 'react';
import { ErrorBoundary } from '../ErrorBoundary';
import type { LowStockEvent } from '../../hooks/useStockWebSocket';

interface LowStockAlertBannerProps {
	// Low stock event to display
	event: LowStockEvent;
	// Callback when dismissed
	onDismiss: () => void;
	// Callback when "View Product" clicked
	onViewProduct?: (productId: number) => void;
	// Auto-dismiss delay in milliseconds (default: 30 seconds)
	autoDismissDelay?: number;
}

export default function LowStockAlertBanner({
	event,
	onDismiss,
	onViewProduct,
	autoDismissDelay = 30000,
}: LowStockAlertBannerProps) {
	const [isVisible, setIsVisible] = useState(true);

	// Auto-dismiss after specified delay
	useEffect(() => {
		const timer = setTimeout(() => {
			handleDismiss();
		}, autoDismissDelay);

		return () => clearTimeout(timer);
	}, [autoDismissDelay]);

	const handleDismiss = () => {
		setIsVisible(false);
		// Wait for fade-out animation to complete
		setTimeout(() => {
			onDismiss();
		}, 300);
	};

	const handleViewProduct = () => {
		if (onViewProduct) {
			onViewProduct(event.productId);
		}
		handleDismiss();
	};

	// Calculate severity based on how far below threshold
	// Guard against division by zero if reorderThreshold is 0
	const stockPercentage = event.reorderThreshold > 0
		? (event.currentStock / event.reorderThreshold) * 100
		: 0;
	const severity =
		stockPercentage === 0 ? 'critical' : stockPercentage < 50 ? 'high' : 'medium';

	const severityColors = {
		critical: 'bg-red-50 border-red-500 text-red-900',
		high: 'bg-orange-50 border-orange-500 text-orange-900',
		medium: 'bg-yellow-50 border-yellow-500 text-yellow-900',
	};

	const iconColors = {
		critical: 'text-red-600',
		high: 'text-orange-600',
		medium: 'text-yellow-600',
	};

	return (
		<div
			className={`transform transition-all duration-300 ${
				isVisible
					? 'translate-y-0 opacity-100'
					: 'translate-y-[-100%] opacity-0'
			}`}
		>
			<div
				className={`${severityColors[severity]} border-l-4 rounded-md shadow-md p-4 mb-3`}
			>
				<div className="flex items-start justify-between">
					<div className="flex items-start gap-3 flex-1">
						{/* Alert Icon */}
						<div className={`${iconColors[severity]} mt-0.5`}>
							<svg
								className="w-5 h-5"
								fill="currentColor"
								viewBox="0 0 20 20"
							>
								<path
									fillRule="evenodd"
									d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
									clipRule="evenodd"
								/>
							</svg>
						</div>

						{/* Alert Content */}
						<div className="flex-1">
							<h3 className="font-semibold text-sm">
								Low Stock Alert
							</h3>
							<p className="text-sm mt-1">
								<strong>{event.productName}</strong> (SKU: {event.sku}) is running low at{' '}
								<strong>{event.branchName}</strong>
							</p>
							<p className="text-xs mt-1">
								Current: {event.currentStock} / Threshold: {event.reorderThreshold}
							</p>
							<p className="text-sm mt-2 font-medium">
								Order {event.suggestedOrderQty} units for {event.branchName}
							</p>
						</div>
					</div>

					{/* Action Buttons */}
					<div className="flex items-center gap-2 ml-4">
						{onViewProduct && (
							<button
								onClick={handleViewProduct}
								className="px-3 py-1.5 text-xs font-medium rounded bg-white bg-opacity-50 hover:bg-opacity-70 transition-colors"
							>
								View Product
							</button>
						)}
						<button
							onClick={handleDismiss}
							className="p-1 rounded-full hover:bg-black hover:bg-opacity-10 transition-colors"
							aria-label="Dismiss"
						>
							<svg
								className="w-4 h-4"
								fill="currentColor"
								viewBox="0 0 20 20"
							>
								<path
									fillRule="evenodd"
									d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
									clipRule="evenodd"
								/>
							</svg>
						</button>
					</div>
				</div>
			</div>
		</div>
	);
}

/**
 * LowStockAlertBannerManager Component
 * Manages multiple low stock alerts and displays them as stacked banners
 */
interface LowStockAlertBannerManagerProps {
	// Array of active low stock events
	alerts: LowStockEvent[];
	// Callback when alert is dismissed
	onDismiss: (eventId: string) => void;
	// Callback when "View Product" clicked
	onViewProduct?: (productId: number) => void;
}

export function LowStockAlertBannerManager({
	alerts,
	onDismiss,
	onViewProduct,
}: LowStockAlertBannerManagerProps) {
	if (alerts.length === 0) {
		return null;
	}

	// Generate unique IDs for alerts based on their content
	const getAlertId = (event: LowStockEvent): string => {
		return `${event.productId}-${event.branchId}-${event.currentStock}`;
	};

	return (
		<ErrorBoundary
			fallback={
				<div className="bg-yellow-50 border border-yellow-500 rounded-md shadow-md p-4 m-4">
					<p className="text-yellow-900 text-sm">
						Alert notifications temporarily unavailable
					</p>
				</div>
			}
		>
			<div className="fixed top-4 right-4 left-4 z-50 max-w-md mx-auto">
				{alerts.map((alert) => (
					<LowStockAlertBanner
						key={getAlertId(alert)}
						event={alert}
						onDismiss={() => onDismiss(getAlertId(alert))}
						onViewProduct={onViewProduct}
					/>
				))}
			</div>
		</ErrorBoundary>
	);
}
