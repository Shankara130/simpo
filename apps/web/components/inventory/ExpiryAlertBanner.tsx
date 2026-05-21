/**
 * ExpiryAlertBanner Component
 * Story 4.5, Task 7: Create Expiry Alert Banner Component (AC: 5, 6, 7)
 *
 * This component displays alert banners when products are approaching their expiry dates.
 * Features:
 * - Real-time expiry notifications via WebSocket (product.expiry events)
 * - Color-coded alert levels:
 *   - 30-day: Yellow/Orange background (warning)
 *   - 14-day: Orange background (critical)
 *   - 7-day: Red background with bold text (urgent)
 * - Product info display (SKU, name, expiry date, days remaining, branch)
 * - "View Product" and "Dismiss" buttons
 * - Multiple alerts support (stacked)
 * - Auto-dismiss after 60 seconds or manual dismiss
 */

'use client';

import { useState, useEffect } from 'react';
import { ErrorBoundary } from '../ErrorBoundary';
import type { ExpiryEvent } from '../../hooks/useStockWebSocket';

interface ExpiryAlertBannerProps {
  // Expiry event to display
  event: ExpiryEvent;
  // Callback when dismissed
  onDismiss: () => void;
  // Callback when "View Product" clicked
  onViewProduct?: (productId: number) => void;
  // Auto-dismiss delay in milliseconds (default: 60 seconds)
  autoDismissDelay?: number;
}

export default function ExpiryAlertBanner({
  event,
  onDismiss,
  onViewProduct,
  autoDismissDelay = 60000,
}: ExpiryAlertBannerProps) {
  const [isVisible, setIsVisible] = useState(true);
  // PATCH: Track hover state to pause auto-dismiss timer
  const [isPaused, setIsPaused] = useState(false);

  // Auto-dismiss after specified delay
  // PATCH: Pause timer when user hovers over the alert
  useEffect(() => {
    if (isPaused) {
      return; // Don't start timer if paused
    }

    const timer = setTimeout(() => {
      handleDismiss();
    }, autoDismissDelay);

    return () => clearTimeout(timer);
  }, [autoDismissDelay, isPaused]);

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

  // Color coding based on alert level
  // Story 4.5, Task 7.3: Display alert banner with color coding
  // PATCH: AC7 requires "red background, bold text" for urgent (7-day) alerts
  const alertLevelColors = {
    warning: 'bg-yellow-50 border-yellow-500 text-yellow-900', // 30-day
    critical: 'bg-orange-50 border-orange-500 text-orange-900', // 14-day
    urgent: 'bg-red-600 border-red-700 text-white font-bold', // 7-day - PATCH: Dark red with white text per AC7
  };

  const iconColors = {
    warning: 'text-yellow-600',
    critical: 'text-orange-600',
    urgent: 'text-white', // PATCH: White icon for dark red background
  };

  const bgGradient = {
    warning: 'from-yellow-50 to-yellow-100',
    critical: 'from-orange-50 to-orange-100',
    urgent: 'from-red-600 to-red-700', // PATCH: Dark red gradient for urgent
  };

  // Format expiry date for display
  const expiryDate = new Date(event.expiryDate);
  const formattedExpiryDate = expiryDate.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });

  const colors = alertLevelColors[event.alertLevel];

  return (
    <div
      className={`transform transition-all duration-300 ${
        isVisible
          ? 'translate-y-0 opacity-100'
          : 'translate-y-[-100%] opacity-0'
      }`}
    >
      <div
        className={`${colors} border-l-4 rounded-md shadow-md p-4 mb-3 bg-gradient-to-r ${bgGradient[event.alertLevel]}`}
        // PATCH: Pause auto-dismiss timer when user interacts with alert
        onMouseEnter={() => setIsPaused(true)}
        onMouseLeave={() => setIsPaused(false)}
      >
        <div className="flex items-start justify-between">
          <div className="flex items-start gap-3 flex-1">
            {/* Alert Icon */}
            <div className={`${iconColors[event.alertLevel]} mt-0.5`}>
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
                {event.alertLevel === 'urgent' && (
                  <span className="font-bold">URGENT: </span>
                )}
                Expiry Alert
              </h3>
              <p className="text-sm mt-1">
                <strong>{event.productName}</strong> (SKU: {event.sku}) expires soon at{' '}
                <strong>{event.branchName}</strong>
              </p>
              <p className="text-xs mt-1">
                Expiry Date: {formattedExpiryDate}
              </p>
              <p className={`text-sm mt-2 font-medium ${event.alertLevel === 'urgent' ? 'font-bold' : ''}`}>
                {event.daysRemaining} day{event.daysRemaining !== 1 ? 's' : ''} remaining
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
 * ExpiryAlertBannerManager Component
 * Manages multiple expiry alerts and displays them as stacked banners
 * Story 4.5, Task 7.7: Support multiple alerts (stack or carousel)
 */
interface ExpiryAlertBannerManagerProps {
  // Array of active expiry events
  alerts: ExpiryEvent[];
  // Callback when alert is dismissed
  onDismiss: (eventId: string) => void;
  // Callback when "View Product" clicked
  onViewProduct?: (productId: number) => void;
}

export function ExpiryAlertBannerManager({
  alerts,
  onDismiss,
  onViewProduct,
}: ExpiryAlertBannerManagerProps) {
  if (alerts.length === 0) {
    return null;
  }

  // Sort alerts by urgency (urgent first, then critical, then warning)
  const sortedAlerts = [...alerts].sort((a, b) => {
    const urgencyOrder = { urgent: 0, critical: 1, warning: 2 };
    return urgencyOrder[a.alertLevel] - urgencyOrder[b.alertLevel];
  });

  // Generate unique IDs for alerts
  const getAlertId = (event: ExpiryEvent): string => {
    return `expiry-${event.productId}-${event.branchId}-${event.expiryDate}`;
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
        {sortedAlerts.map((alert) => (
          <ExpiryAlertBanner
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
