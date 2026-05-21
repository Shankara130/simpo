/**
 * Low Stock Products Page
 * Story 4.4, Task 8: Create Low Stock Products Page (AC: 3, 4)
 *
 * This page displays all products with current stock below their reorder thresholds.
 * Features:
 * - Fetch low stock products from GET /api/v1/products/low-stock
 * - Display table with: Product, SKU, Current Stock, Threshold, Suggested Order, Branch, Actions
 * - Filter by branch (for multi-branch Owners)
 * - Sort by severity (most below threshold first)
 * - "Order Stock" button for each product (links to supplier management - future Story 10.x)
 * - Real-time updates via WebSocket subscription
 */

'use client';

import { useState, useEffect } from 'react';
import { useStockWebSocket, type LowStockEvent } from '../../../hooks/useStockWebSocket';

// Types for low stock products
interface LowStockProduct {
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

interface LowStockProductsResponse {
	data: LowStockProduct[];
	pagination: PaginationMetadata;
}

export default function LowStockPage() {
	const [products, setProducts] = useState<LowStockProduct[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [selectedBranch, setSelectedBranch] = useState<number | null>(null);
	const [userBranch, setUserBranch] = useState<number | null>(null);
	const [userRole, setUserRole] = useState<string | null>(null);

	// Real-time low stock alerts from WebSocket
	const [activeAlerts, setActiveAlerts] = useState<LowStockEvent[]>([]);

	// Fetch low stock products
	const fetchLowStockProducts = async (branchId?: number) => {
		setLoading(true);
		setError(null);

		try {
			const params = new URLSearchParams();
			if (branchId !== undefined) {
				params.append('branch_id', branchId.toString());
			}

			const response = await fetch(
				`/api/v1/products/low-stock?${params.toString()}`,
				{
					method: 'GET',
					headers: {
						'Content-Type': 'application/json',
					},
					credentials: 'include',
				}
			);

			if (!response.ok) {
				throw new Error('Failed to fetch low stock products');
			}

			const data: LowStockProductsResponse = await response.json();

			// Sort by severity (most below threshold first)
			const sortedProducts = data.data.sort((a, b) => {
				const aSeverity = a.stockQty / a.reorderThreshold;
				const bSeverity = b.stockQty / b.reorderThreshold;
				return aSeverity - bSeverity;
			});

			setProducts(sortedProducts);
		} catch (err) {
			setError(err instanceof Error ? err.message : 'An error occurred');
		} finally {
			setLoading(false);
		}
	};

	// Fetch user context (branch, role) - for now, default to no branch filter
	useEffect(() => {
		// TODO: Get user context from JWT or API
		// For now, default to showing all branches
		fetchLowStockProducts();
	}, []);

	// Handle real-time low stock alerts from WebSocket
	const handleLowStockAlert = (event: LowStockEvent) => {
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
		fetchLowStockProducts(selectedBranch ?? undefined);
	};

	// Setup WebSocket for real-time low stock notifications
	// Helper to extract JWT token from cookies
	const getJwtToken = (): string | null => {
		if (typeof window === 'undefined') return null;
		const cookies = document.cookie.split(';');
		for (const cookie of cookies) {
			const [name, value] = cookie.trim().split('=');
			if (name === 'token' || name === 'jwt') {
				return value;
			}
		}
		return null;
	};

	const token = getJwtToken();

	// Only connect if we have a token (authenticated user)
	useEffect(() => {
		if (token) {
			useStockWebSocket({
				token,
				branches: selectedBranch ? [selectedBranch] : [],
				onLowStock: handleLowStockAlert,
			});
		}
	}, [token, selectedBranch]);

	const handleBranchChange = (branchId: string) => {
		const value = branchId === 'all' ? null : parseInt(branchId, 10);
		setSelectedBranch(value);
		fetchLowStockProducts(value ?? undefined);
	};

	const handleDismissAlert = (alertId: string) => {
		setActiveAlerts((prev) => prev.filter((_, index) => {
			// Generate ID from alert properties
			const id = `${prev[index].productId}-${prev[index].branchId}-${prev[index].currentStock}`;
			return id !== alertId;
		}));
	};

	const getSeverityColor = (product: LowStockProduct) => {
		const percentage = (product.stockQty / product.reorderThreshold) * 100;
		if (percentage === 0) return 'bg-red-100 text-red-800';
		if (percentage < 50) return 'bg-orange-100 text-orange-800';
		return 'bg-yellow-100 text-yellow-800';
	};

	const getSeverityLabel = (product: LowStockProduct) => {
		const percentage = (product.stockQty / product.reorderThreshold) * 100;
		if (percentage === 0) return 'Critical';
		if (percentage < 50) return 'High';
		return 'Medium';
	};

	return (
		<div className="container mx-auto px-4 py-8">
			<div className="mb-6">
				<h1 className="text-3xl font-bold text-gray-900">
					Low Stock Products
				</h1>
				<p className="text-gray-600 mt-2">
					Products that have fallen below their reorder thresholds
				</p>
			</div>

			{/* Real-time Low Stock Alerts */}
			{activeAlerts.length > 0 && (
				<div className="mb-6">
					{activeAlerts.map((alert) => (
						<div
							key={`${alert.productId}-${alert.branchId}-${alert.currentStock}`}
							className="bg-red-50 border-l-4 border-red-500 rounded-md shadow-md p-4 mb-3"
						>
							<div className="flex items-start justify-between">
								<div className="flex-1">
									<h3 className="font-semibold text-sm text-red-900">
										Low Stock Alert
									</h3>
									<p className="text-sm text-red-800 mt-1">
										<strong>{alert.productName}</strong> (SKU: {alert.sku}) is running low
									</p>
									<p className="text-xs text-red-700 mt-1">
										Current: {alert.currentStock} / Threshold: {alert.reorderThreshold}
									</p>
									<p className="text-sm font-medium text-red-900 mt-2">
										Order {alert.suggestedOrderQty} units for {alert.branchName}
									</p>
								</div>
								<button
									onClick={() =>
										handleDismissAlert(
											`${alert.productId}-${alert.branchId}-${alert.currentStock}`
										)
									}
									className="ml-4 text-red-600 hover:text-red-800"
								>
									Dismiss
								</button>
							</div>
						</div>
					))}
				</div>
			)}

			{/* Branch Filter */}
			<div className="mb-6 flex items-center gap-4">
				<label htmlFor="branch-filter" className="text-sm font-medium text-gray-700">
					Filter by Branch:
				</label>
				<select
					id="branch-filter"
					value={selectedBranch ?? 'all'}
					onChange={(e) => handleBranchChange(e.target.value)}
					className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value="all">All Branches</option>
					{/* TODO: Fetch available branches from API */}
					<option value="1">Branch 1</option>
					<option value="2">Branch 2</option>
				</select>
			</div>

			{/* Loading State */}
			{loading && (
				<div className="text-center py-12">
					<div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
					<p className="mt-4 text-gray-600">Loading low stock products...</p>
				</div>
			)}

			{/* Error State */}
			{error && (
				<div className="bg-red-50 border border-red-200 rounded-md p-4 mb-6">
					<p className="text-red-800">{error}</p>
				</div>
			)}

			{/* Products Table */}
			{!loading && !error && (
				<>
					{products.length === 0 ? (
						<div className="text-center py-12 bg-white rounded-lg shadow">
							<p className="text-gray-500">No products with low stock found</p>
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
												Current Stock
											</th>
											<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
												Threshold
											</th>
											<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
												Suggested Order
											</th>
											<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
												Branch
											</th>
											<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
												Severity
											</th>
											<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
												Actions
											</th>
										</tr>
									</thead>
									<tbody className="bg-white divide-y divide-gray-200">
										{products.map((product) => (
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
													<div className="text-sm text-gray-900">{product.stockQty}</div>
												</td>
												<td className="px-6 py-4 whitespace-nowrap">
													<div className="text-sm text-gray-500">{product.reorderThreshold}</div>
												</td>
												<td className="px-6 py-4 whitespace-nowrap">
													<div className="text-sm font-medium text-blue-600">
														{product.reorderThreshold - product.stockQty}
													</div>
												</td>
												<td className="px-6 py-4 whitespace-nowrap">
													<div className="text-sm text-gray-500">
														Branch {product.branchId}
													</div>
												</td>
												<td className="px-6 py-4 whitespace-nowrap">
													<span
														className={`px-2 py-1 text-xs font-semibold rounded-full ${getSeverityColor(
															product
														)}`}
													>
														{getSeverityLabel(product)}
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
													<button
														className="text-green-600 hover:text-green-800"
														onClick={() => {
															// TODO: Navigate to order stock (future Story 10.x)
															console.log('Order stock for product:', product.id);
														}}
													>
														Order Stock
													</button>
												</td>
											</tr>
										))}
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
