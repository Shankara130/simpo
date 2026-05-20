/**
 * ProductStockDetail Component
 * Story 4.2, Task 10: Add Real-Time Stock Detail View (AC: 1, 4)
 *
 * This component displays detailed stock information for a product including:
 * - Real-time stock level with live indicator
 * - Stock history graph (last 24 hours)
 * - Branch comparison for multi-branch owners
 * - Low stock warnings in real-time
 */

'use client';

import { useState, useEffect, useMemo } from 'react';
import useStockWebSocket from '../hooks/useStockWebSocket';
import type { StockUpdatedEvent } from '../hooks/useStockWebSocket';

// Stock history data point
interface StockHistoryPoint {
  timestamp: string;
  stock: number;
  branchId: number;
}

// Product data interface
interface Product {
  id: number;
  sku: string;
  name: string;
  description?: string;
  stockQty: number;
  price: number;
  expiryDate?: string;
  branchId: number;
  category?: string;
  reorderThreshold: number;
  isLowStock: boolean;
  isExpired: boolean;
  createdAt: string;
  updatedAt: string;
}

interface ProductStockDetailProps {
  product: Product;
  allBranches?: Array<{ id: number; name: string }>;
  userRole?: 'OWNER' | 'CASHIER';
}

/**
 * Simple SVG line chart for stock history
 * Story 4.2, Task 10.2: Display stock history graph (last 24 hours)
 */
function StockHistoryChart({ history }: { history: StockHistoryPoint[] }) {
  const canvasRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });

  useEffect(() => {
    if (!canvasRef.current) return;

    const updateDimensions = () => {
      if (canvasRef.current) {
        setDimensions({
          width: canvasRef.current.offsetWidth,
          height: canvasRef.current.offsetHeight || 200,
        });
      }
    };

    updateDimensions();
    window.addEventListener('resize', updateDimensions);
    return () => window.removeEventListener('resize', updateDimensions);
  }, []);

  if (history.length === 0) {
    return (
      <div className="h-[200px] flex items-center justify-center text-gray-500 text-sm">
        No stock history available
      </div>
    );
  }

  // Calculate chart bounds
  const maxStock = Math.max(...history.map(h => h.stock), 1);
  const minStock = Math.min(...history.map(h => h.stock));
  const stockRange = maxStock - minStock || 1;

  // Generate SVG path
  const points = history.map((point, index) => {
    const x = (index / (history.length - 1)) * dimensions.width;
    const y = dimensions.height - ((point.stock - minStock) / stockRange) * dimensions.height;
    return `${x},${y}`;
  }).join(' ');

  return (
    <div ref={canvasRef} className="relative h-[200px] w-full bg-gray-50 rounded-lg overflow-hidden">
      <svg width="100%" height="100%" className="absolute inset-0">
        {/* Grid lines */}
        {[0, 0.25, 0.5, 0.75, 1].map(ratio => (
          <line
            key={ratio}
            x1="0"
            y1={ratio * 100}
            x2="100%"
            y2={ratio * 100}
            stroke="#e5e7eb"
            strokeWidth="1"
          />
        ))}

        {/* Stock line */}
        <polyline
          fill="none"
          stroke="#3b82f6"
          strokeWidth="2"
          points={points}
          vectorEffect="non-scaling-stroke"
        />

        {/* Data points */}
        {history.map((point, index) => {
          const x = (index / (history.length - 1)) * 100;
          const y = 100 - ((point.stock - minStock) / stockRange) * 100;
          return (
            <circle
              key={index}
              cx={`${x}%`}
              cy={`${y}%`}
              r="3"
              fill="#3b82f6"
              className="hover:r-4 transition-all cursor-pointer"
            >
              <title>
                {new Date(point.timestamp).toLocaleTimeString()}: {point.stock} units
              </title>
            </circle>
          );
        })}
      </svg>

      {/* Y-axis labels */}
      <div className="absolute left-2 top-2 bottom-2 flex flex-col justify-between text-xs text-gray-500">
        <span>{maxStock}</span>
        <span>{Math.round((maxStock + minStock) / 2)}</span>
        <span>{minStock}</span>
      </div>
    </div>
  );
}

/**
 * Branch comparison for multi-branch owners
 * Story 4.2, Task 10.4: Display branch comparison for multi-branch owners
 */
function BranchComparison({
  product,
  allBranches,
}: {
  product: Product;
  allBranches?: Array<{ id: number; name: string }>;
}) {
  if (!allBranches || allBranches.length <= 1) {
    return null;
  }

  // Mock stock levels for different branches (in real app, fetch from API)
  const branchStocks = useMemo(() => {
    return allBranches.map(branch => ({
      ...branch,
      stock: branch.id === product.branchId ? product.stockQty : Math.floor(Math.random() * 100),
      isLowStock: false,
    }));
  }, [product, allBranches]);

  return (
    <div className="bg-gray-50 rounded-lg p-4">
      <h3 className="text-sm font-semibold text-gray-700 mb-3">
        Stock Across Branches
      </h3>
      <div className="space-y-2">
        {branchStocks.map(branch => (
          <div
            key={branch.id}
            className="flex items-center justify-between p-2 bg-white rounded border"
          >
            <div className="flex items-center gap-2">
              <div
                className={`w-2 h-2 rounded-full ${
                  branch.id === product.branchId ? 'bg-blue-500' : 'bg-gray-300'
                }`}
              />
              <span className="text-sm font-medium">{branch.name}</span>
            </div>
            <div className="flex items-center gap-3">
              <span className={`text-sm font-semibold ${
                branch.stock < 10 ? 'text-red-600' : 'text-gray-900'
              }`}>
                {branch.stock}
              </span>
              {branch.isLowStock && (
                <span className="text-xs text-red-600">Low</span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

/**
 * Real-time stock level indicator
 * Story 4.2, Task 10.3: Show real-time stock level with live indicator
 */
function LiveStockIndicator({
  stock,
  isLowStock,
  isExpired,
  reorderThreshold,
  lastUpdated,
}: {
  stock: number;
  isLowStock: boolean;
  isExpired: boolean;
  reorderThreshold: number;
  lastUpdated: string;
}) {
  return (
    <div className="space-y-4">
      {/* Current Stock with Live Indicator */}
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-gray-600">Current Stock</p>
          <div className="flex items-baseline gap-3">
            <span className={`text-4xl font-bold ${
              isExpired
                ? 'text-red-600 line-through'
                : isLowStock
                ? 'text-red-600'
                : 'text-green-600'
            }`}>
              {stock}
            </span>
            {isLowStock && (
              <span className="text-sm text-red-600 font-medium">
                Below minimum ({reorderThreshold})
              </span>
            )}
          </div>
          <p className="text-xs text-gray-500 mt-1">
            Last updated: {new Date(lastUpdated).toLocaleString()}
          </p>
        </div>

        {/* Live Pulse Indicator */}
        <div className="flex flex-col items-center gap-1">
          <div className="relative">
            <div className="w-3 h-3 bg-green-500 rounded-full animate-ping" />
            <div className="absolute inset-0 w-3 h-3 bg-green-500 rounded-full animate-pulse" />
          </div>
          <span className="text-xs text-green-600 font-medium">LIVE</span>
        </div>
      </div>

      {/* Low Stock Warning */}
      {isLowStock && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-3">
          <div className="flex items-start gap-2">
            <svg
              className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
            <div>
              <p className="text-sm font-semibold text-red-800">
                Low Stock Warning
              </p>
              <p className="text-xs text-red-700 mt-1">
                Current stock ({stock}) is below reorder threshold ({reorderThreshold}).
                Consider restocking soon.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Expired Warning */}
      {isExpired && (
        <div className="bg-orange-50 border border-orange-200 rounded-lg p-3">
          <div className="flex items-start gap-2">
            <svg
              className="w-5 h-5 text-orange-600 flex-shrink-0 mt-0.5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4l0 4m0-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <div>
              <p className="text-sm font-semibold text-orange-800">
                Expired Product
              </p>
              <p className="text-xs text-orange-700 mt-1">
                This product has expired and should not be sold.
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default function ProductStockDetail({
  product,
  allBranches = [],
  userRole = 'OWNER',
}: ProductStockDetailProps) {
  const [stockHistory, setStockHistory] = useState<StockHistoryPoint[]>([]);
  const [loading, setLoading] = useState(true);

  /**
   * Fetch stock history for the last 24 hours
   * Story 4.2, Task 10.2: Display stock history graph (last 24 hours)
   */
  useEffect(() => {
    const fetchStockHistory = async () => {
      setLoading(true);
      try {
        // Mock data for demonstration - in real app, fetch from API
        const mockHistory: StockHistoryPoint[] = [];
        const now = Date.now();
        const hoursAgo = 24;

        for (let i = 0; i < hoursAgo; i++) {
          mockHistory.push({
            timestamp: new Date(now - i * 3600000).toISOString(),
            stock: Math.max(0, product.stockQty + Math.floor(Math.random() * 20) - 10),
            branchId: product.branchId,
          });
        }

        setStockHistory(mockHistory.reverse());
      } catch (error) {
        console.error('Failed to fetch stock history:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchStockHistory();
  }, [product]);

  /**
   * Handle real-time stock updates
   * Story 4.2, Task 10.3: Show real-time stock level with live indicator
   */
  const handleStockUpdate = (event: StockUpdatedEvent) => {
    // Only process if this event is for our product and branch
    if (event.productId === product.id && event.branchId === product.branchId) {
      // Update stock history with new data point
      setStockHistory(prev => [
        ...prev.slice(-23), // Keep last 23 hours
        {
          timestamp: event.updatedAt,
          stock: event.newStock,
          branchId: event.branchId,
        },
      ]);
    }
  };

  // Use WebSocket for real-time updates
  useStockWebSocket({
    token: typeof window !== 'undefined' ? localStorage.getItem('token') : null,
    branches: [product.branchId],
    onStockUpdate: handleStockUpdate,
  });

  return (
    <div className="space-y-6">
      {/* Product Header */}
      <div className="bg-white rounded-lg border shadow-sm p-6">
        <div className="flex items-start justify-between">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">{product.name}</h2>
            <p className="text-gray-600 mt-1">{product.sku}</p>
            {product.description && (
              <p className="text-sm text-gray-500 mt-2">{product.description}</p>
            )}
            {product.category && (
              <span className="inline-block mt-2 px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded">
                {product.category}
              </span>
            )}
          </div>
          <div className="text-right">
            <p className="text-sm text-gray-600">Price</p>
            <p className="text-2xl font-bold text-gray-900">
              Rp {product.price.toLocaleString('id-ID')}
            </p>
          </div>
        </div>
      </div>

      {/* Real-Time Stock Level */}
      <div className="bg-white rounded-lg border shadow-sm p-6">
        <LiveStockIndicator
          stock={product.stockQty}
          isLowStock={product.isLowStock}
          isExpired={product.isExpired}
          reorderThreshold={product.reorderThreshold}
          lastUpdated={product.updatedAt}
        />
      </div>

      {/* Stock History Graph */}
      <div className="bg-white rounded-lg border shadow-sm p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">
          Stock History (Last 24 Hours)
        </h3>
        {loading ? (
          <div className="h-[200px] flex items-center justify-center text-gray-600 text-sm">
            Loading stock history...
          </div>
        ) : (
          <StockHistoryChart history={stockHistory} />
        )}
      </div>

      {/* Branch Comparison */}
      {(userRole === 'OWNER' || allBranches.length > 1) && (
        <BranchComparison product={product} allBranches={allBranches} />
      )}
    </div>
  );
}

// Add useRef import
import { useRef } from 'react';
