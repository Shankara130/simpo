/**
 * Expiring Products Page
 * Story 4.5, Task 8: Create Expiring Products Page (AC: 5, 6)
 *
 * This page displays all products approaching their expiry dates.
 * Features:
 * - Fetch expiring products from GET /api/v1/products/expiring
 * - Filter by days threshold (30, 14, 7) and branch
 * - Display table with: Product, SKU, Expiry Date, Days Remaining, Branch, Actions
 * - Sort by urgency (closest expiry first)
 * - "Create Discount" button for each product (future: discount management)
 * - Real-time updates via WebSocket subscription
 */

'use client';

import { useState, useEffect, useCallback } from 'react';
import { useStockWebSocket, type ExpiryEvent } from '../../../hooks/useStockWebSocket';

// Types for expiring products
interface ExpiringProduct {
  id: number;
  sku: string;
  name: string;
  description?: string;
  stockQty: number;
  price: string;
  expiryDate?: string;
  branchId: number;
  category?: string;
  reorderThreshold: number;
  isLowStock: boolean;
  isExpired: boolean;
  createdAt: string;
  updatedAt: string;
}

interface PaginationMetadata {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

interface ExpiringProductsResponse {
  data: ExpiringProduct[];
  pagination: PaginationMetadata;
}

export default function ExpiringPage() {
  const [products, setProducts] = useState<ExpiringProduct[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedBranch, setSelectedBranch] = useState<number | null>(null);
  const [daysThreshold, setDaysThreshold] = useState<number>(30);

  // Real-time expiry alerts from WebSocket
  const [activeAlerts, setActiveAlerts] = useState<ExpiryEvent[]>([]);

  // Fetch expiring products
  // Story 4.5, Task 8.2: Fetch expiring products from GET /api/v1/products/expiring
  const fetchExpiringProducts = async (branchId?: number, days?: number) => {
    setLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams();
      if (branchId !== undefined) {
        params.append('branch_id', branchId.toString());
      }
      if (days !== undefined) {
        params.append('days', days.toString());
      }

      const response = await fetch(
        `/api/v1/products/expiring?${params.toString()}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
        }
      );

      if (!response.ok) {
        throw new Error('Failed to fetch expiring products');
      }

      const data: ExpiringProductsResponse = await response.json();

      // Story 4.5, Task 8.5: Sort by urgency (closest expiry first)
      const sortedProducts = data.data.sort((a, b) => {
        if (!a.expiryDate) return 1;
        if (!b.expiryDate) return -1;
        return new Date(a.expiryDate).getTime() - new Date(b.expiryDate).getTime();
      });

      setProducts(sortedProducts);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  // Fetch user context and load products on mount
  useEffect(() => {
    // TODO: Get user context from JWT or API
    fetchExpiringProducts(selectedBranch ?? undefined, daysThreshold);
  }, []);

  // Handle real-time expiry alerts from WebSocket
  // Story 4.5, Task 8.7: Real-time updates via WebSocket subscription
  const handleExpiryAlert = (event: ExpiryEvent) => {
    // Add to active alerts if not already present
    setActiveAlerts((prev) => {
      const exists = prev.some(
        (a) => a.productId === event.productId && a.branchId === event.branchId
      );
      if (exists) return prev;

      // Keep only the most recent 5 alerts
      return [event, ...prev].slice(0, 5);
    });

    // Refresh the product list to show updated data
    fetchExpiringProducts(selectedBranch ?? undefined, daysThreshold);
  };

  // Setup WebSocket for real-time expiry notifications
  // PATCH: Memoize JWT token extraction to avoid parsing cookies on every render
  const getJwtToken = useCallback((): string | null => {
    if (typeof window === 'undefined') return null;
    const cookies = document.cookie.split(';');
    for (const cookie of cookies) {
      const [name, value] = cookie.trim().split('=');
      if (name === 'token' || name === 'jwt') {
        return value;
      }
    }
    return null;
  }, []);

  const token = getJwtToken();

  // Only connect if we have a token (authenticated user)
  useEffect(() => {
    if (token) {
      useStockWebSocket({
        token,
        branches: selectedBranch ? [selectedBranch] : [],
        onExpiry: handleExpiryAlert,
      });
    }
  }, [token, selectedBranch]);

  const handleBranchChange = (branchId: string) => {
    const value = branchId === 'all' ? null : parseInt(branchId, 10);
    setSelectedBranch(value);
    fetchExpiringProducts(value ?? undefined, daysThreshold);
  };

  // Story 4.5, Task 8.3: Add filter by days (30, 14, 7) and branch
  const handleDaysChange = (days: number) => {
    setDaysThreshold(days);
    fetchExpiringProducts(selectedBranch ?? undefined, days);
  };

  const handleDismissAlert = (alertId: string) => {
    // PATCH: Properly clean up dismissed alerts from state
    setActiveAlerts((prev) => prev.filter((alert) => {
      // Generate consistent ID for comparison
      const id = `expiry-${alert.productId}-${alert.branchId}-${alert.expiryDate}`;
      return id !== alertId;
    }));
  };

  // Calculate days remaining for display
  const getDaysRemaining = (product: ExpiringProduct): number => {
    if (!product.expiryDate) return 0;
    const expiryDate = new Date(product.expiryDate);
    const today = new Date();
    const diffTime = expiryDate.getTime() - today.getTime();
    return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  };

  // Get alert level and color based on days remaining
  const getAlertLevel = (daysRemaining: number): 'urgent' | 'critical' | 'warning' => {
    if (daysRemaining <= 7) return 'urgent';
    if (daysRemaining <= 14) return 'critical';
    return 'warning';
  };

  const getAlertLevelColor = (level: 'urgent' | 'critical' | 'warning'): string => {
    switch (level) {
      case 'urgent':
        return 'bg-red-100 text-red-800';
      case 'critical':
        return 'bg-orange-100 text-orange-800';
      case 'warning':
        return 'bg-yellow-100 text-yellow-800';
    }
  };

  const getAlertLevelLabel = (level: 'urgent' | 'critical' | 'warning'): string => {
    switch (level) {
      case 'urgent':
        return 'Urgent';
      case 'critical':
        return 'Critical';
      case 'warning':
        return 'Warning';
    }
  };

  // Format expiry date for display
  const formatExpiryDate = (dateString?: string): string => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">
          Expiring Products
        </h1>
        <p className="text-gray-600 mt-2">
          Products approaching their expiry dates
        </p>
      </div>

      {/* Real-time Expiry Alerts */}
      {/* Story 4.5, AC5: Notifications displayed via alert banner */}
      {activeAlerts.length > 0 && (
        <div className="mb-6">
          {activeAlerts.map((alert) => {
            const level = alert.alertLevel;
            const bgColor = level === 'urgent' ? 'bg-red-50 border-red-500' :
                            level === 'critical' ? 'bg-orange-50 border-orange-500' :
                            'bg-yellow-50 border-yellow-500';
            const textColor = level === 'urgent' ? 'text-red-900' :
                               level === 'critical' ? 'text-orange-900' :
                               'text-yellow-900';

            return (
              <div
                key={`expiry-${alert.productId}-${alert.branchId}-${alert.expiryDate}`}
                className={`${bgColor} border-l-4 rounded-md shadow-md p-4 mb-3`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className={`font-semibold text-sm ${textColor}`}>
                      {level === 'urgent' && <span className="font-bold">URGENT: </span>}
                      Expiry Alert
                    </h3>
                    <p className={`text-sm mt-1 ${textColor.replace('900', '800')}`}>
                      <strong>{alert.productName}</strong> (SKU: {alert.sku}) expires soon at{' '}
                      <strong>{alert.branchName}</strong>
                    </p>
                    <p className="text-xs mt-1 opacity-75">
                      {alert.daysRemaining} day{alert.daysRemaining !== 1 ? 's' : ''} remaining
                    </p>
                  </div>
                  <button
                    onClick={() =>
                      handleDismissAlert(
                        `expiry-${alert.productId}-${alert.branchId}-${alert.expiryDate}`
                      )
                    }
                    className="ml-4 text-gray-600 hover:text-gray-800"
                  >
                    Dismiss
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Filters */}
      {/* Story 4.5, Task 8.3: Add filter by days (30, 14, 7) and branch */}
      <div className="mb-6 flex items-center gap-4 flex-wrap">
        <div className="flex items-center gap-2">
          <label htmlFor="days-filter" className="text-sm font-medium text-gray-700">
            Days Threshold:
          </label>
          <select
            id="days-filter"
            value={daysThreshold}
            onChange={(e) => handleDaysChange(parseInt(e.target.value, 10))}
            className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value={30}>30 days</option>
            <option value={14}>14 days</option>
            <option value={7}>7 days</option>
          </select>
        </div>

        <div className="flex items-center gap-2">
          <label htmlFor="branch-filter" className="text-sm font-medium text-gray-700">
            Branch:
          </label>
          <select
            id="branch-filter"
            value={selectedBranch ?? 'all'}
            onChange={(e) => handleBranchChange(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All Branches</option>
            {/* TODO: Fetch available branches from API */}
            <option value="1">Jakarta Branch</option>
            <option value="2">Bandung Branch</option>
          </select>
        </div>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-gray-600">Loading expiring products...</p>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-6">
          <p className="text-red-800">{error}</p>
        </div>
      )}

      {/* Products Table */}
      {/* Story 4.5, Task 8.4: Display table with columns: Product, SKU, Expiry Date, Days Remaining, Branch, Actions */}
      {!loading && !error && (
        <>
          {products.length === 0 ? (
            <div className="text-center py-12 bg-white rounded-lg shadow">
              <p className="text-gray-500">No products expiring within the selected timeframe</p>
            </div>
          ) : (
            <div className="bg-white shadow-lg rounded-lg overflow-hidden">
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Product
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        SKU
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Expiry Date
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Days Remaining
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Branch
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Alert Level
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {products.map((product) => {
                      const daysRemaining = getDaysRemaining(product);
                      const alertLevel = getAlertLevel(daysRemaining);

                      return (
                        <tr key={`${product.id}-${product.branchId}`}>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="text-sm font-medium text-gray-900">
                              {product.name}
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="text-sm text-gray-500">{product.sku}</div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="text-sm text-gray-900">{formatExpiryDate(product.expiryDate)}</div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className={`text-sm font-medium ${
                              daysRemaining <= 7 ? 'text-red-600' :
                              daysRemaining <= 14 ? 'text-orange-600' :
                              'text-yellow-600'
                            }`}>
                              {daysRemaining} day{daysRemaining !== 1 ? 's' : ''}
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="text-sm text-gray-500">
                              Branch {product.branchId}
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span
                              className={`px-2 py-1 text-xs font-semibold rounded-full ${getAlertLevelColor(
                                alertLevel
                              )}`}
                            >
                              {getAlertLevelLabel(alertLevel)}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                            <button
                              className="text-blue-600 hover:text-blue-800 mr-3"
                              onClick={() => {
                                // TODO: Navigate to product details
                                console.log('View product:', product.id);
                              }}
                            >
                              View
                            </button>
                            {/* Story 4.5, Task 8.6: "Create Discount" button for each product (future: discount management) */}
                            <button
                              className="text-green-600 hover:text-green-800"
                              onClick={() => {
                                // TODO: Navigate to discount creation (future Story 10.x or separate feature)
                                console.log('Create discount for product:', product.id);
                              }}
                            >
                              Create Discount
                            </button>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
