/**
 * Stock Adjustment Modal Component
 * Story 4.3, Task 7: Create Stock Adjustment Modal Component (AC: 1, 2, 3)
 *
 * This modal allows administrators to manually adjust stock quantities with reason logging.
 * Only Admin and Owner roles can access this functionality.
 */

'use client';

import { useState, useEffect } from 'react';

// Product data interface (matches backend DTO)
interface Product {
  id: number;
  sku: string;
  name: string;
  stockQty: number;
  branchId: number;
  reorderThreshold: number;
}

// Stock adjustment reason enum (matches backend)
export type StockAdjustmentReason =
  | 'Damage'
  | 'Expiration'
  | 'DeliveryReceipt'
  | 'PhysicalCount'
  | 'TheftLoss'
  | 'Other';

// Stock adjustment request interface (matches backend DTO)
interface StockAdjustmentRequest {
  productId: number;
  branchId: number;
  newStockQty: number;
  reason: StockAdjustmentReason;
  reasonNotes?: string;
}

// Stock adjustment result interface (matches backend DTO)
export interface StockAdjustmentResult {
  productId: number;
  sku: string;
  name: string;
  oldStockQty: number;
  newStockQty: number;
  change: number;
  reason: string;
  adjustedBy: string;
  adjustedAt: string;
}

// Component props
interface StockAdjustmentModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (result: StockAdjustmentResult) => void;
  product?: Product;
  userRole: 'OWNER' | 'ADMIN' | 'CASHIER';
  userBranchId?: number;
  availableProducts?: Product[];
  availableBranches?: Array<{ id: number; name: string }>;
}

// Valid stock adjustment reasons
const VALID_REASONS: StockAdjustmentReason[] = [
  'Damage',
  'Expiration',
  'DeliveryReceipt',
  'PhysicalCount',
  'TheftLoss',
  'Other',
];

export default function StockAdjustmentModal({
  isOpen,
  onClose,
  onSuccess,
  product,
  userRole,
  userBranchId,
  availableProducts = [],
  availableBranches = [],
}: StockAdjustmentModalProps) {
  // Form state
  const [selectedProductId, setSelectedProductId] = useState<number>(product?.id || 0);
  const [selectedBranchId, setSelectedBranchId] = useState<number>(
    product?.branchId || userBranchId || 0
  );
  const [newStockQty, setNewStockQty] = useState<string>(product?.stockQty.toString() || '');
  const [reason, setReason] = useState<StockAdjustmentReason>('DeliveryReceipt');
  const [reasonNotes, setReasonNotes] = useState('');

  // Loading and error state
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Get current product for display
  const currentProduct = availableProducts.find(p => p.id === selectedProductId) || product;

  // Reset form when modal opens/closes or product changes
  useEffect(() => {
    if (isOpen) {
      if (product) {
        setSelectedProductId(product.id);
        setSelectedBranchId(product.branchId);
        setNewStockQty(product.stockQty.toString());
      } else {
        setSelectedProductId(0);
        setSelectedBranchId(userBranchId || 0);
        setNewStockQty('');
      }
      setReason('DeliveryReceipt');
      setReasonNotes('');
      setError(null);
    }
  }, [isOpen, product, userBranchId]);

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validation
    if (!selectedProductId) {
      setError('Please select a product');
      return;
    }
    if (!selectedBranchId) {
      setError('Please select a branch');
      return;
    }
    if (newStockQty === '' || parseInt(newStockQty) < 0) {
      setError('Stock quantity must be a non-negative number');
      return;
    }
    if (!reason) {
      setError('Please select a reason for the adjustment');
      return;
    }
    if (reason === 'Other' && !reasonNotes.trim()) {
      setError('Please provide additional details when selecting "Other" reason');
      return;
    }

    setIsSubmitting(true);

    try {
      const request: StockAdjustmentRequest = {
        productId: selectedProductId,
        branchId: selectedBranchId,
        newStockQty: parseInt(newStockQty),
        reason,
        reasonNotes: reason === 'Other' ? reasonNotes : undefined,
      };

      const response = await fetch('/api/v1/products/stock/adjust', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.detail || errorData.title || 'Failed to adjust stock');
      }

      const result: StockAdjustmentResult = await response.json();
      onSuccess(result);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to adjust stock');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Calculate stock change
  const stockChange = currentProduct && newStockQty !== ''
    ? parseInt(newStockQty) - currentProduct.stockQty
    : null;

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex justify-between items-center">
            <h2 className="text-xl font-semibold text-gray-900">
              Adjust Stock Quantity
            </h2>
            <button
              onClick={onClose}
              disabled={isSubmitting}
              className="text-gray-400 hover:text-gray-500 disabled:opacity-50"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <p className="text-sm text-gray-600 mt-1">
            Manually adjust stock quantity with reason logging for audit trail
          </p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="px-6 py-4">
          {/* Error message */}
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {/* Product Selector */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Product <span className="text-red-500">*</span>
            </label>
            {product ? (
              <div className="p-3 bg-gray-50 border border-gray-200 rounded-md">
                <div className="font-medium text-gray-900">{product.name}</div>
                <div className="text-sm text-gray-500">SKU: {product.sku}</div>
              </div>
            ) : (
              <select
                value={selectedProductId}
                onChange={(e) => setSelectedProductId(parseInt(e.target.value))}
                disabled={isSubmitting}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                required
              >
                <option value={0}>Select a product</option>
                {availableProducts.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.sku} - {p.name}
                  </option>
                ))}
              </select>
            )}
          </div>

          {/* Branch Selector */}
          {(userRole === 'OWNER' || userRole === 'ADMIN') && availableBranches.length > 0 && (
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Branch <span className="text-red-500">*</span>
              </label>
              <select
                value={selectedBranchId}
                onChange={(e) => setSelectedBranchId(parseInt(e.target.value))}
                disabled={isSubmitting || !!product}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                required
              >
                <option value={0}>Select a branch</option>
                {availableBranches.map((branch) => (
                  <option key={branch.id} value={branch.id}>
                    {branch.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Current Stock Display */}
          {currentProduct && (
            <div className="mb-4 p-4 bg-blue-50 border border-blue-200 rounded-md">
              <div className="flex justify-between items-center">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Current Stock
                  </label>
                  <div className="text-2xl font-bold text-gray-900">
                    {currentProduct.stockQty}
                  </div>
                  {currentProduct.stockQty < currentProduct.reorderThreshold && (
                    <div className="text-xs text-red-600 mt-1">
                      ⚠️ Low stock (min: {currentProduct.reorderThreshold})
                    </div>
                  )}
                </div>
                {stockChange !== null && newStockQty !== '' && (
                  <div className="text-right">
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Change
                    </label>
                    <div className={`text-2xl font-bold ${
                      stockChange > 0 ? 'text-green-600' : stockChange < 0 ? 'text-red-600' : 'text-gray-600'
                    }`}>
                      {stockChange > 0 ? '+' : ''}{stockChange}
                    </div>
                    <div className="text-xs text-gray-500 mt-1">
                      New stock: {parseInt(newStockQty)}
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* New Stock Quantity Input */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              New Stock Quantity <span className="text-red-500">*</span>
            </label>
            <input
              type="number"
              min="0"
              value={newStockQty}
              onChange={(e) => setNewStockQty(e.target.value)}
              disabled={isSubmitting}
              placeholder="Enter new stock quantity"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              required
            />
            <p className="text-xs text-gray-500 mt-1">
              Enter the new total stock quantity (not the increment/decrement)
            </p>
          </div>

          {/* Reason Selector */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Reason <span className="text-red-500">*</span>
            </label>
            <select
              value={reason}
              onChange={(e) => setReason(e.target.value as StockAdjustmentReason)}
              disabled={isSubmitting}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              required
            >
              {VALID_REASONS.map((r) => (
                <option key={r} value={r}>
                  {r === 'Damage' && '💥 Damage - Products damaged and removed from stock'}
                  {r === 'Expiration' && '📅 Expiration - Products expired and disposed'}
                  {r === 'DeliveryReceipt' && '📦 Delivery Receipt - New stock received from supplier'}
                  {r === 'PhysicalCount' && '🔢 Physical Count - Adjustment based on physical inventory count'}
                  {r === 'TheftLoss' && '🚨 Theft/Loss - Stock lost due to theft or loss'}
                  {r === 'Other' && '📝 Other - Other reasons with additional details'}
                </option>
              ))}
            </select>
          </div>

          {/* Reason Notes (for "Other" reason) */}
          {reason === 'Other' && (
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Additional Details <span className="text-red-500">*</span>
              </label>
              <textarea
                value={reasonNotes}
                onChange={(e) => setReasonNotes(e.target.value)}
                disabled={isSubmitting}
                placeholder="Please provide additional details about the adjustment..."
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                required={reason === 'Other'}
              />
            </div>
          )}

          {/* Warning for large reductions */}
          {stockChange !== null && stockChange < -10 && (
            <div className="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
              <p className="text-sm text-yellow-800">
                ⚠️ <strong>Warning:</strong> You are reducing stock by {Math.abs(stockChange)} units.
                This will be logged in the audit trail.
              </p>
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex justify-end gap-3 mt-6">
            <button
              type="button"
              onClick={onClose}
              disabled={isSubmitting}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
            >
              {isSubmitting ? (
                <>
                  <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Adjusting...
                </>
              ) : (
                'Adjust Stock'
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
