/**
 * Low Stock Screen
 * Story 4.4, Task 12: Create Mobile Low Stock Screen (AC: 3, 4, 5)
 *
 * This screen displays products with low stock levels for mobile users.
 * Features:
 * - Fetch low stock products from API (same as web)
 * - Display list with product details, current stock, threshold
 * - Pull-to-refresh functionality
 * - Tap product to view details or create order
 */

import React, { useState, useEffect, useCallback } from 'react';
import {
	View,
	Text,
	FlatList,
	StyleSheet,
	RefreshControl,
	TouchableOpacity,
	ActivityIndicator,
	Alert,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';

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
	success: boolean;
}

interface NavigationProps {
	navigate: (screen: string, params?: object) => void;
}

const LowStockScreen: React.FC<NavigationProps> = ({ navigate }) => {
	const [products, setProducts] = useState<LowStockProduct[]>([]);
	const [loading, setLoading] = useState(true);
	const [refreshing, setRefreshing] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [userBranch, setUserBranch] = useState<number | null>(null);

	// Fetch low stock products
	const fetchLowStockProducts = useCallback(async (branchId?: number) => {
		setLoading(true);
		setError(null);

		try {
			// TODO: Replace with actual API endpoint
			// For now, using mock data
			const mockProducts: LowStockProduct[] = [
				{
					id: 123,
					sku: 'SKU-12345',
					name: 'Paracetamol 500mg',
					description: 'Pain reliever',
					stockQty: 5,
					price: '50000.00',
					branchId: 1,
					reorderThreshold: 10,
					isLowStock: true,
					isExpired: false,
					createdAt: '2026-05-20T10:00:00Z',
					updatedAt: '2026-05-20T10:00:00Z',
				},
				{
					id: 456,
					sku: 'SKU-45678',
					name: 'Amoxicillin 500mg',
					description: 'Antibiotic',
					stockQty: 0,
					price: '75000.00',
					branchId: 1,
					reorderThreshold: 10,
					isLowStock: true,
					isExpired: false,
					createdAt: '2026-05-20T10:00:00Z',
					updatedAt: '2026-05-20T10:00:00Z',
				},
			];

			// Sort by severity (most below threshold first)
			const sortedProducts = mockProducts.sort((a, b) => {
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
	}, []);

	// Initial load
	useEffect(() => {
		fetchLowStockProducts(userBranch ?? undefined);
	}, [userBranch]);

	// Pull to refresh handler
	const onRefresh = useCallback(() => {
		setRefreshing(true);
		fetchLowStockProducts(userBranch ?? undefined).finally(() => {
			setRefreshing(false);
		});
	}, [userBranch]);

	// Handle product tap
	const handleProductPress = (product: LowStockProduct) => {
		// Story 4.4, Task 12.5: Tap product to view details or create order
		Alert.alert(
			'Product Details',
			`Product: ${product.name}\nSKU: ${product.sku}\nCurrent Stock: ${product.stockQty}\nThreshold: ${product.reorderThreshold}`,
			[
				{
					text: 'View Details',
					onPress: () => {
						// TODO: Navigate to product details screen
						console.log('Navigate to product details:', product.id);
					},
				},
				{
					text: 'Order Stock',
					onPress: () => {
						// TODO: Navigate to order stock screen (future Story 10.x)
						console.log('Order stock for product:', product.id);
					},
				},
				{
					text: 'Cancel',
					style: 'cancel',
				},
			]
		);
	};

	// Get severity color based on stock percentage
	const getSeverityColor = (product: LowStockProduct) => {
		const percentage = (product.stockQty / product.reorderThreshold) * 100;
		if (percentage === 0) return '#EF4444'; // red - critical
		if (percentage < 50) return '#F97316'; // orange - high
		return '#F59E0B'; // yellow - medium
	};

	// Get severity label
	const getSeverityLabel = (product: LowStockProduct) => {
		const percentage = (product.stockQty / product.reorderThreshold) * 100;
		if (percentage === 0) return 'CRITICAL';
		if (percentage < 50) return 'HIGH';
		return 'MEDIUM';
	};

	const renderProduct = ({ item }: { item: LowStockProduct }) => {
		const severityColor = getSeverityColor(item);
		const severityLabel = getSeverityLabel(item);
		const suggestedOrderQty = item.reorderThreshold - item.stockQty;

		return (
			<TouchableOpacity
				style={styles.productCard}
				onPress={() => handleProductPress(item)}
				activeOpacity={0.7}
			>
				<View style={styles.productHeader}>
					<View style={styles.productInfo}>
						<Text style={styles.productName}>{item.name}</Text>
						<Text style={styles.productSku}>SKU: {item.sku}</Text>
						<Text style={styles.branchLabel}>Branch {item.branchId}</Text>
					</View>
					<View style={[styles.severityBadge, { backgroundColor: severityColor }]}>
						<Text style={styles.severityText}>{severityLabel}</Text>
					</View>
				</View>

				<View style={styles.stockInfo}>
					<View style={styles.stockRow}>
						<Text style={styles.stockLabel}>Current:</Text>
						<Text style={styles.stockValue}>{item.stockQty}</Text>
					</View>
					<View style={styles.stockRow}>
						<Text style={styles.stockLabel}>Threshold:</Text>
						<Text style={styles.stockValue}>{item.reorderThreshold}</Text>
					</View>
					<View style={styles.stockRow}>
						<Text style={styles.stockLabel}>Order:</Text>
						<Text style={styles.orderValue}>{suggestedOrderQty} units</Text>
					</View>
				</View>
			</TouchableOpacity>
		);
	};

	// Render list empty component
	const renderEmptyComponent = () => (
		<View style={styles.emptyContainer}>
			<Text style={styles.emptyText}>No products with low stock</Text>
			<Text style={styles.emptySubtext}>
				All products are above their reorder thresholds
			</Text>
		</View>
	);

	// Render list header component
	const renderListHeader = () => (
		<View style={styles.listHeader}>
			<Text style={styles.listHeaderTitle}>
				{products.length} {products.length === 1 ? 'product' : 'products'} with low stock
			</Text>
		</View>
	);

		return (
		<View style={styles.container}>
			{/* Header */}
			<View style={styles.header}>
				<Text style={styles.headerTitle}>Low Stock Products</Text>
				<Text style={styles.headerSubtitle}>
					Products below reorder threshold
				</Text>
			</View>

			{/* Loading State */}
			{loading ? (
				<View style={styles.loadingContainer}>
					<ActivityIndicator size="large" color="#3B82F6" />
					<Text style={styles.loadingText}>Loading low stock products...</Text>
				</View>
			) : error ? (
				<View style={styles.errorContainer}>
					<Text style={styles.errorText}>{error}</Text>
					<TouchableOpacity
						style={styles.retryButton}
						onPress={() => fetchLowStockProducts(userBranch ?? undefined)}
					>
						<Text style={styles.retryButtonText}>Retry</Text>
					</TouchableOpacity>
				</View>
			) : (
				<>
					{products.length === 0 ? (
						renderEmptyComponent()
					) : (
						<FlatList
							data={products}
							renderItem={renderProduct}
							keyExtractor={(item) => `${item.id}-${item.branchId}`}
							contentContainerStyle={styles.listContent}
							ListHeaderComponent={renderListHeader}
							refreshControl={
								<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
							}
						/>
					)}
				</>
			)}
		</View>
	);
};

const styles = StyleSheet.create({
	container: {
		flex: 1,
		backgroundColor: '#F9FAFB',
	},
	header: {
		backgroundColor: '#FFFFFF',
		padding: 20,
		borderBottomWidth: 1,
		borderBottomColor: '#E5E7EB',
	},
	headerTitle: {
		fontSize: 24,
		fontWeight: 'bold',
		color: '#111827',
	},
	headerSubtitle: {
		fontSize: 14,
		color: '#6B7280',
		marginTop: 4,
	},
	loadingContainer: {
		flex: 1,
		justifyContent: 'center',
		alignItems: 'center',
		padding: 20,
	},
	loadingText: {
		marginTop: 12,
		fontSize: 14,
		color: '#6B7280',
	},
	errorContainer: {
		flex: 1,
		justifyContent: 'center',
		alignItems: 'center',
		padding: 20,
	},
	errorText: {
		fontSize: 14,
		color: '#EF4444',
		marginBottom: 12,
		textAlign: 'center',
	},
	retryButton: {
		backgroundColor: '#3B82F6',
		paddingHorizontal: 16,
		paddingVertical: 8,
		borderRadius: 6,
	},
	retryButtonText: {
		color: '#FFFFFF',
		fontSize: 14,
		fontWeight: '600',
	},
	productCard: {
		backgroundColor: '#FFFFFF',
		marginHorizontal: 16,
		marginBottom: 12,
		borderRadius: 8,
		padding: 16,
		borderWidth: 1,
		borderColor: '#E5E7EB',
		shadowColor: '#000',
		shadowOffset: { width: 0, height: 1 },
		shadowOpacity: 0.05,
		shadowRadius: 2,
		elevation: 2,
	},
	productHeader: {
		flexDirection: 'row',
		justifyContent: 'space-between',
		alignItems: 'flex-start',
		marginBottom: 12,
	},
	productInfo: {
		flex: 1,
	},
	productName: {
		fontSize: 16,
		fontWeight: '600',
		color: '#111827',
		marginBottom: 4,
	},
	productSku: {
		fontSize: 12,
		color: '#6B7280',
	},
	branchLabel: {
		fontSize: 11,
		color: '#9CA3AF',
		marginTop: 2,
	},
	severityBadge: {
		paddingHorizontal: 8,
		paddingVertical: 4,
		borderRadius: 12,
		alignSelf: 'flex-start',
	},
	severityText: {
		fontSize: 10,
		fontWeight: '600',
		color: '#FFFFFF',
		textTransform: 'uppercase',
	},
	stockInfo: {
		gap: 8,
	},
	stockRow: {
		flexDirection: 'row',
		alignItems: 'center',
	},
	stockLabel: {
		fontSize: 13,
		color: '#6B7280',
	},
	stockValue: {
		fontSize: 13,
		fontWeight: '600',
		color: '#111827',
		marginLeft: 8,
	},
	orderValue: {
		fontSize: 13,
		fontWeight: '600',
		color: '#059669',
		marginLeft: 8,
	},
	listContent: {
		paddingBottom: 20,
	},
	listHeader: {
		paddingHorizontal: 16,
		paddingVertical: 12,
		backgroundColor: '#F3F4F6',
	},
	listHeaderTitle: {
		fontSize: 14,
		fontWeight: '600',
		color: '#374151',
	},
	emptyContainer: {
		flex: 1,
		justifyContent: 'center',
		alignItems: 'center',
		padding: 40,
	},
	emptyText: {
		fontSize: 16,
		fontWeight: '600',
		color: '#6B7280',
		marginBottom: 8,
		textAlign: 'center',
	},
	emptySubtext: {
		fontSize: 14,
		color: '#9CA3AF',
		textAlign: 'center',
	},
});

export default LowStockScreen;
