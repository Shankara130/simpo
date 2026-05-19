/**
 * Product List Page with Real-Time Stock Updates
 * Story 4.2, Task 8: Integrate Real-Time Stock into Product List (AC: 1, 4, 5)
 *
 * This page displays products with real-time stock updates via WebSocket.
 * Owners can view stock across all branches, cashiers see only their branch.
 */

'use client';

import { useState, useEffect, useMemo } from 'react';
import useStockWebSocket from '../../../hooks/useStockWebSocket';
import { ToastContainer, useToastNotifications } from '../../../components/StockUpdateToast';
import type { StockUpdatedEvent } from '../../../hooks/useStockWebSocket';

// Product data interface (matches backend DTO)
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

// Product list response interface
interface ProductListResponse {
  data: Product[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

// Filter options interface
interface FilterOptions {
  search: string;
  category: string;
  branchId?: number;
  lowStock: boolean;
  expired: boolean;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export default function ProductsPage() {
  // Product list state
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Pagination state
  const [page, setPage] = useState(1);
  const [limit] = useState(20);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);

  // Filter state
  const [filters, setFilters] = useState<FilterOptions>({
    search: '',
    category: '',
    lowStock: false,
    expired: false,
    sortBy: 'createdAt',
    sortOrder: 'desc',
  });

  // Real-time stock update indicators
  const [flashingProducts, setFlashingProducts] = useState<Set<number>>(new Set());

  // Toast notifications for stock updates
  const { toasts, addToast, removeToast } = useToastNotifications(5);

  // User info (in real app, this would come from auth context)
  const [userRole, setUserRole] = useState<'OWNER' | 'CASHIER'>('OWNER');
  const [userBranchId, setUserBranchId] = useState<number | undefined>(undefined);

  /**
   * Fetch products from API
   * Story 4.2, Task 8.1: Modify ProductListScreen (web version)
   */
  const fetchProducts = async () => {
    setLoading(true);
    setError(null);

    try {
      // Build query parameters
      const params = new URLSearchParams({
        page: page.toString(),
        limit: limit.toString(),
        ...(filters.search && { search: filters.search }),
        ...(filters.category && { category: filters.category }),
        ...(filters.branchId && { branch_id: filters.branchId.toString() }),
        ...(filters.lowStock && { low_stock: 'true' }),
        ...(filters.expired && { expired: 'true' }),
        ...(filters.sortBy && { sort_by: filters.sortBy }),
        ...(filters.sortOrder && { sort_order: filters.sortOrder }),
      });

      const response = await fetch(`/api/v1/products?${params}`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result: ProductListResponse = await response.json();
      setProducts(result.data);
      setTotal(result.pagination.total);
      setTotalPages(result.pagination.totalPages);

    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch products');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Handle real-time stock updates from WebSocket
   * Story 4.2, Task 8.2-8.4: Update stock quantities in real-time with visual feedback
   * Story 4.2, Task 9: Show toast notification for stock changes
   */
  const handleStockUpdate = (event: StockUpdatedEvent) => {
    // Add toast notification
    addToast(event);

    setProducts(prevProducts => {
      return prevProducts.map(product => {
        // Only update if the product matches the event
        if (product.id === event.productId && product.branchId === event.branchId) {
          // Trigger flash animation
          triggerFlash(product.id);

          // Return updated product with new stock
          return {
            ...product,
            stockQty: event.newStock,
            isLowStock: event.newStock < product.reorderThreshold,
            updatedAt: event.updatedAt,
          };
        }
        return product;
      });
    });
  };

  /**
   * Trigger flash animation for a product
   * Story 4.2, Task 8.4: Show visual indicator when stock updates (flash animation)
   */
  const triggerFlash = (productId: number) => {
    setFlashingProducts(prev => new Set([...prev, productId]));

    // Remove flash after animation completes
    setTimeout(() => {
      setFlashingProducts(prev => {
        const newSet = new Set(prev);
        newSet.delete(productId);
        return newSet;
      });
    }, 2000); // 2 second flash animation
  };

  /**
   * Initialize WebSocket connection on component mount
   * Story 4.2, Task 8.2: Initialize WebSocket connection on component mount
   */
  const branchesForWebSocket = useMemo(() => {
    // Owners can see all branches (empty array = all branches)
    // Cashiers can only see their assigned branch
    if (userRole === 'CASHIER' && userBranchId) {
      return [userBranchId];
    }
    return []; // Empty = all branches
  }, [userRole, userBranchId]);

  const { connectionState, isConnected } = useStockWebSocket({
    token: typeof window !== 'undefined' ? localStorage.getItem('token') : null,
    branches: branchesForWebSocket,
    onStockUpdate: handleStockUpdate,
  });

  /**
   * Fetch products on mount and when filters/pagination change
   */
  useEffect(() => {
    fetchProducts();
  }, [page, limit, filters]);

  /**
   * Clean up WebSocket connection on unmount
   * Story 4.2, Task 8.6: Clean up WebSocket connection on unmount
   */
  useEffect(() => {
    return () => {
      // WebSocket cleanup is handled by the hook
    };
  }, []);

  /**
   * Handle filter changes
   */
  const handleFilterChange = (key: keyof FilterOptions, value: any) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setPage(1); // Reset to first page when filters change
  };

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Page Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Products</h1>
          <p className="text-gray-600 mt-1">
            Manage your inventory with real-time stock updates
          </p>
        </div>

        {/* Connection Status Indicator */}
        <div className="flex items-center gap-2">
          <div
            className={`px-3 py-1 rounded-full text-sm font-medium ${
              isConnected
                ? 'bg-green-100 text-green-800'
                : connectionState === 'connecting'
                ? 'bg-yellow-100 text-yellow-800'
                : connectionState === 'error'
                ? 'bg-red-100 text-red-800'
                : 'bg-gray-100 text-gray-800'
            }`}
          >
            <span className={`inline-block w-2 h-2 rounded-full mr-2 ${
              isConnected
                ? 'bg-green-500 animate-pulse'
                : 'bg-gray-400'
            }`} />
            {connectionState === 'connected'
              ? 'Live'
              : connectionState === 'connecting'
              ? 'Connecting...'
              : connectionState === 'error'
              ? 'Connection Error'
              : 'Disconnected'}
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white p-6 rounded-lg border shadow-sm mb-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {/* Search */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Search
            </label>
            <input
              type="text"
              placeholder="Search by name or SKU..."
              value={filters.search}
              onChange={(e) => handleFilterChange('search', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {/* Category */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Category
            </label>
            <select
              value={filters.category}
              onChange={(e) => handleFilterChange('category', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">All Categories</option>
              <option value="Medicine">Medicine</option>
              <option value="Supplements">Supplements</option>
              <option value="Medical Devices">Medical Devices</option>
              <option value="Personal Care">Personal Care</option>
            </select>
          </div>

          {/* Low Stock Filter */}
          <div className="flex items-end">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={filters.lowStock}
                onChange={(e) => handleFilterChange('lowStock', e.target.checked)}
                className="mr-2 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <span className="text-sm font-medium text-gray-700">Low Stock Only</span>
            </label>
          </div>

          {/* Expired Filter */}
          <div className="flex items-end">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={filters.expired}
                onChange={(e) => handleFilterChange('expired', e.target.checked)}
                className="mr-2 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <span className="text-sm font-medium text-gray-700">Expired Only</span>
            </label>
          </div>
        </div>
      </div>

      {/* Products Table */}
      <div className="bg-white rounded-lg border shadow-sm overflow-hidden">
        {loading ? (
          <div className="p-8 text-center text-gray-600">Loading products...</div>
        ) : error ? (
          <div className="p-8 text-center text-red-600">Error: {error}</div>
        ) : products.length === 0 ? (
          <div className="p-8 text-center text-gray-600">No products found</div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      SKU
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Name
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Category
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Stock
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Price
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {products.map((product) => (
                    <tr
                      key={`${product.id}-${product.branchId}`}
                      className={`transition-colors duration-500 ${
                        flashingProducts.has(product.id)
                          ? 'bg-yellow-50'
                          : 'hover:bg-gray-50'
                      }`}
                    >
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {product.sku}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">{product.name}</div>
                        {product.description && (
                          <div className="text-sm text-gray-500 truncate max-w-xs">
                            {product.description}
                          </div>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {product.category || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <span className={`text-sm font-medium ${
                            product.isLowStock
                              ? 'text-red-600'
                              : product.stockQty === 0
                              ? 'text-red-800'
                              : 'text-green-600'
                          }`}>
                            {product.stockQty}
                          </span>
                          {flashingProducts.has(product.id) && (
                            <span className="ml-2 text-xs text-yellow-600 animate-pulse">
                              ✓ Updated just now
                            </span>
                          )}
                          {product.branchId && (
                            <span className="ml-2 text-xs text-gray-400">
                              Branch {product.branchId}
                            </span>
                          )}
                        </div>
                        {product.isLowStock && (
                          <div className="text-xs text-red-600">
                            Low stock (min: {product.reorderThreshold})
                          </div>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        Rp {product.price.toLocaleString('id-ID')}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {product.isExpired ? (
                          <span className="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-red-100 text-red-800">
                            Expired
                          </span>
                        ) : product.isLowStock ? (
                          <span className="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-yellow-100 text-yellow-800">
                            Low Stock
                          </span>
                        ) : (
                          <span className="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                            In Stock
                          </span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            <div className="bg-white px-4 py-3 border-t border-gray-200 sm:px-6">
              <div className="flex items-center justify-between">
                <div className="flex-1 flex justify-between sm:hidden">
                  <button
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={page === 1}
                    className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Previous
                  </button>
                  <button
                    onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                    disabled={page >= totalPages}
                    className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Next
                  </button>
                </div>
                <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
                  <div>
                    <p className="text-sm text-gray-700">
                      Showing{' '}
                      <span className="font-medium">{(page - 1) * limit + 1}</span> to{' '}
                      <span className="font-medium">
                        {Math.min(page * limit, total)}
                      </span>{' '}
                      of <span className="font-medium">{total}</span> results
                    </p>
                  </div>
                  <div>
                    <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                      <button
                        onClick={() => setPage(p => Math.max(1, p - 1))}
                        disabled={page === 1}
                        className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        Previous
                      </button>
                      {[...Array(totalPages)].map((_, i) => (
                        <button
                          key={i + 1}
                          onClick={() => setPage(i + 1)}
                          className={`relative inline-flex items-center px-4 py-2 border text-sm font-medium ${
                            page === i + 1
                              ? 'z-10 bg-blue-50 border-blue-500 text-blue-600'
                              : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50'
                          }`}
                        >
                          {i + 1}
                        </button>
                      ))}
                      <button
                        onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                        disabled={page >= totalPages}
                        className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        Next
                      </button>
                    </nav>
                  </div>
                </div>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Toast Notifications */}
      <ToastContainer toasts={toasts} onDismiss={removeToast} />
    </div>
  );
}
