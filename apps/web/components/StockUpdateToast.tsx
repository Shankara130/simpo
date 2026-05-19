/**
 * StockUpdateToast Component
 * Story 4.2, Task 9: Create Stock Update Notification Component (AC: 2, 3, 5)
 *
 * This component displays brief notifications when stock levels change.
 * Features:
 * - Shows product name, old stock, new stock, and change delta
 * - Different colors for increases vs decreases
 * - Auto-dismiss after 3 seconds
 * - Manual dismiss support
 */

'use client';

import { useEffect, useState } from 'react';
import type { StockUpdatedEvent } from '../hooks/useStockWebSocket';

interface StockUpdateToastProps {
  event: StockUpdatedEvent;
  onDismiss: () => void;
  autoDismissDelay?: number;
}

export default function StockUpdateToast({
  event,
  onDismiss,
  autoDismissDelay = 3000,
}: StockUpdateToastProps) {
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

  // Determine color based on change direction
  const isIncrease = event.change > 0;
  const isDecrease = event.change < 0;
  const bgColor = isIncrease
    ? 'bg-green-50 border-green-200'
    : isDecrease
    ? 'bg-red-50 border-red-200'
    : 'bg-gray-50 border-gray-200';

  const textColor = isIncrease
    ? 'text-green-800'
    : isDecrease
    ? 'text-red-800'
    : 'text-gray-800';

  const icon = isIncrease ? '↑' : isDecrease ? '↓' : '→';

  return (
    <div
      className={`
        ${isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-2'}
        transition-all duration-300 ease-in-out
        ${bgColor} border ${textColor}
        px-4 py-3 rounded-lg shadow-lg
        flex items-start gap-3
        min-w-[300px] max-w-md
      `}
      role="alert"
      aria-live="polite"
    >
      {/* Icon */}
      <div className={`flex-shrink-0 text-2xl ${
        isIncrease ? 'text-green-600' : isDecrease ? 'text-red-600' : 'text-gray-600'
      }`}>
        {icon}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between mb-1">
          <p className="text-sm font-semibold">
            {event.name}
          </p>
          <span className="text-xs text-gray-500">
            {new Date(event.updatedAt).toLocaleTimeString()}
          </span>
        </div>

        <div className="space-y-1">
          {/* SKU and Branch */}
          <div className="flex items-center gap-2 text-xs opacity-75">
            <span className="font-mono">{event.sku}</span>
            {event.branchId && (
              <>
                <span>•</span>
                <span>Branch {event.branchId}</span>
              </>
            )}
          </div>

          {/* Stock Change */}
          <div className="flex items-center gap-3 text-sm">
            <div className="flex items-center gap-1">
              <span className="line-through opacity-60">
                {event.oldStock}
              </span>
              <span>→</span>
              <span className="font-semibold">
                {event.newStock}
              </span>
            </div>

            {/* Change Delta */}
            <div className={`font-semibold ${
              isIncrease ? 'text-green-700' : isDecrease ? 'text-red-700' : 'text-gray-700'
            }`}>
              {event.change > 0 ? '+' : ''}{event.change}
            </div>
          </div>

          {/* Updated By */}
          <div className="text-xs opacity-75">
            Updated by: {event.updatedBy}
          </div>
        </div>
      </div>

      {/* Dismiss Button */}
      <button
        onClick={handleDismiss}
        className="flex-shrink-0 p-1 hover:bg-black/5 rounded-full transition-colors"
        aria-label="Dismiss notification"
      >
        <svg
          className="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </button>
    </div>
  );
}

/**
 * Container for managing multiple toast notifications
 * Story 4.2, Task 9: Stock update notification component
 */
interface ToastContainerProps {
  toasts: Array<{
    id: string;
    event: StockUpdatedEvent;
  }>;
  onDismiss: (id: string) => void;
}

export function ToastContainer({ toasts, onDismiss }: ToastContainerProps) {
  if (toasts.length === 0) {
    return null;
  }

  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
      {toasts.map(({ id, event }) => (
        <StockUpdateToast
          key={id}
          event={event}
          onDismiss={() => onDismiss(id)}
        />
      ))}
    </div>
  );
}

/**
 * Hook for managing toast notifications
 * Story 4.2, Task 9.5: Auto-dismiss after 3 seconds
 */
export function useToastNotifications(maxToasts: number = 5) {
  const [toasts, setToasts] = useState<Array<{
    id: string;
    event: StockUpdatedEvent;
  }>([]);

  const addToast = (event: StockUpdatedEvent) => {
    const id = `${event.productId}-${event.branchId}-${Date.now()}`;
    setToasts(prev => {
      // Remove existing toast for same product/branch if exists
      const filtered = prev.filter(t => t.id !== id);
      // Add new toast and limit total
      return [...filtered, { id, event }].slice(-maxToasts);
    });
  };

  const removeToast = (id: string) => {
    setToasts(prev => prev.filter(t => t.id !== id));
  };

  return {
    toasts,
    addToast,
    removeToast,
  };
}
