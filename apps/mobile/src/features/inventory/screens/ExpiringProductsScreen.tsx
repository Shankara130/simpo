/**
 * Expiring Products Screen
 * Story 4.5, Task 12: Create Mobile Expiring Products Screen (AC: 5, 6)
 *
 * This screen displays products approaching their expiry dates for mobile users.
 * Features:
 * - Fetch expiring products from API (same as web)
 * - Display list with product details, expiry date, days remaining
 * - Pull-to-refresh functionality
 * - Tap product to view details or create discount
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
	ScrollView,
} from 'react-native';

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
	success: boolean;
}

// Alert level type
type AlertLevel = 'warning' | 'critical' | 'urgent';

interface NavigationProps {
	navigate: (screen: string, params?: object) => void;
}

const ExpiringProductsScreen: React.FC<NavigationProps> = ({ navigate }) => {
	const [products, setProducts] = useState<ExpiringProduct[]>([]);
	const [loading, setLoading] = useState(true);
	const [refreshing, setRefreshing] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [userBranch, setUserBranch] = useState<number | null>(null);
	const [daysThreshold, setDaysThreshold] = useState<number>(30);

	// Fetch expiring products
	// Story 4.5, Task 12.2: Fetch expiring products from API (same as web)
	const fetchExpiringProducts = useCallback(async (branchId?: number, days?: number) => {
		setLoading(true);
		setError(null);

		try {
			// Build query parameters
			const params = new URLSearchParams();
			if (branchId !== undefined) {
				params.append('branch_id', branchId.toString());
			}
			if (days !== undefined) {
				params.append('days', days.toString());
			}

			// TODO: Replace with actual API endpoint
			// const response = await fetch(`/api/v1/products/expiring?${params.toString()}`, {
			//   credentials: 'include',
			// });
			// For now, using mock data
			const mockProducts: ExpiringProduct[] = [
				{
					id: 123,
					sku: 'SKU-12345',
					name: 'Paracetamol 500mg',
					description: 'Pain reliever',
					stockQty: 50,
					price: '50000.00',
					expiryDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(), // 7 days from now
					branchId: 1,
					reorderThreshold: 10,
					isLowStock: false,
					isExpired: false,
					createdAt: '2026-05-20T10:00:00Z',
					updatedAt: '2026-05-20T10:00:00Z',
				},
				{
					id: 456,
					sku: 'SKU-45678',
					name: 'Amoxicillin 500mg',
					description: 'Antibiotic',
					stockQty: 30,
					price: '75000.00',
					expiryDate: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString(), // 14 days from now
					branchId: 1,
					reorderThreshold: 10,
					isLowStock: false,
					isExpired: false,
					createdAt: '2026-05-20T10:00:00Z',
					updatedAt: '2026-05-20T10:00:00Z',
				},
				{
					id: 789,
					sku: 'SKU-78901',
					name: 'Ibuprofen 400mg',
					description: 'Anti-inflammatory',
					stockQty: 25,
					price: '60000.00',
					expiryDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(), // 30 days from now
					branchId: 2,
					reorderThreshold: 10,
					isLowStock: false,
					isExpired: false,
					createdAt: '2026-05-20T10:00:00Z',
					updatedAt: '2026-05-20T10:00:00Z',
				},
			];

			// Story 4.5, Task 12.3: Sort by urgency (closest expiry first)
			const sortedProducts = mockProducts
				.filter(p => p.expiryDate)
				.sort((a, b) => {
					return new Date(a.expiryDate!).getTime() - new Date(b.expiryDate!).getTime();
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
		fetchExpiringProducts(userBranch ?? undefined, daysThreshold);
	}, [userBranch, daysThreshold]);

	// Pull to refresh handler
	// Story 4.5, Task 12.4: Pull-to-refresh functionality
	const onRefresh = useCallback(() => {
		setRefreshing(true);
		fetchExpiringProducts(userBranch ?? undefined, daysThreshold).finally(() => {
			setRefreshing(false);
		});
	}, [userBranch, daysThreshold]);

	// Handle product tap
	// Story 4.5, Task 12.5: Tap product to view details or create discount
	const handleProductPress = (product: ExpiringProduct) => {
		Alert.alert(
			'Expiring Product',
			`Product: ${product.name}\nSKU: ${product.sku}\nExpiry: ${formatExpiryDate(product.expiryDate)}\nDays Remaining: ${getDaysRemaining(product)}`,
			[
				{
					text: 'View Details',
					onPress: () => {
						// TODO: Navigate to product details screen
						console.log('Navigate to product details:', product.id);
					},
				},
				{
					text: 'Create Discount',
					onPress: () => {
						// TODO: Navigate to discount creation (future Story 10.x or separate feature)
						console.log('Create discount for product:', product.id);
					},
				},
				{
					text: 'Cancel',
					style: 'cancel',
				},
			]
		);
	};

	// Calculate days remaining for display
	const getDaysRemaining = (product: ExpiringProduct): number => {
		if (!product.expiryDate) return 0;
		const expiryDate = new Date(product.expiryDate);
		const today = new Date();
		const diffTime = expiryDate.getTime() - today.getTime();
		return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
	};

	// Get alert level based on days remaining
	const getAlertLevel = (daysRemaining: number): AlertLevel => {
		if (daysRemaining <= 7) return 'urgent';
		if (daysRemaining <= 14) return 'critical';
		return 'warning';
	};

	// Get alert colors based on alert level
	const getAlertColors = (alertLevel: AlertLevel) => {
		switch (alertLevel) {
			case 'urgent':
				return {
					background: '#FEF2F2', // red-50
					border: '#EF4444', // red-500
					badge: '#DC2626', // red-600
					text: '#991B1B', // red-800
				};
			case 'critical':
				return {
					background: '#FFF7ED', // orange-50
					border: '#F97316', // orange-500
					badge: '#EA580C', // orange-600
					text: '#9A3412', // orange-800
				};
			case 'warning':
			default:
				return {
					background: '#FEFCE8', // yellow-50
					border: '#F59E0B', // yellow-500
					badge: '#D97706', // yellow-600
					text: '#92400E', // yellow-800
				};
		}
	};

	// Get alert level label
	const getAlertLevelLabel = (alertLevel: AlertLevel): string => {
		switch (alertLevel) {
			case 'urgent':
				return 'URGENT';
			case 'critical':
				return 'CRITICAL';
			case 'warning':
				return 'WARNING';
		}
	};

	// Format expiry date for display
	const formatExpiryDate = (dateString?: string): string => {
		if (!dateString) return 'N/A';
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
		});
	};

	const renderProduct = ({ item }: { item: ExpiringProduct }) => {
		const daysRemaining = getDaysRemaining(item);
		const alertLevel = getAlertLevel(daysRemaining);
		const colors = getAlertColors(alertLevel);
		const alertLabel = getAlertLevelLabel(alertLevel);

		return (
			<TouchableOpacity
				style={[styles.productCard, { borderColor: colors.border, backgroundColor: colors.background }]}
				onPress={() => handleProductPress(item)}
				activeOpacity={0.7}
			>
				<View style={styles.productHeader}>
					<View style={styles.productInfo}>
						<Text style={[styles.productName, { color: colors.text }]}>
							{item.name}
						</Text>
						<Text style={styles.productSku}>SKU: {item.sku}</Text>
						<Text style={styles.branchLabel}>Branch {item.branchId}</Text>
					</View>
					<View style={[styles.alertBadge, { backgroundColor: colors.badge }]}>
						<Text style={styles.alertText}>{alertLabel}</Text>
					</View>
				</View>

				<View style={styles.expiryInfo}>
					<View style={styles.infoRow}>
						<Text style={[styles.infoLabel, { color: colors.text }]}>Expiry Date:</Text>
						<Text style={[styles.infoValue, { color: colors.text }]}>
							{formatExpiryDate(item.expiryDate)}
						</Text>
					</View>
					<View style={styles.infoRow}>
						<Text style={[styles.infoLabel, { color: colors.text }]}>Days Remaining:</Text>
						<Text style={[styles.daysValue, { color: colors.badge, fontWeight: '700' }]}>
							{daysRemaining} day{daysRemaining !== 1 ? 's' : ''}
						</Text>
					</View>
					<View style={styles.infoRow}>
						<Text style={[styles.infoLabel, { color: colors.text }]}>Stock:</Text>
						<Text style={[styles.infoValue, { color: colors.text }]}>
							{item.stockQty} units
						</Text>
					</View>
				</View>
			</TouchableOpacity>
		);
	};

	// Render list empty component
	const renderEmptyComponent = () => (
		<View style={styles.emptyContainer}>
			<Text style={styles.emptyText}>No expiring products</Text>
			<Text style={styles.emptySubtext}>
				All products are within safe expiry limits
			</Text>
		</View>
	);

	// Render list header component
	const renderListHeader = () => (
		<View style={styles.listHeader}>
			<Text style={styles.listHeaderTitle}>
				{products.length} {products.length === 1 ? 'product' : 'products'} expiring
			</Text>
		</View>
	);

	// Render filter buttons
	const renderFilterButtons = () => (
		<ScrollView
			horizontal
			showsHorizontalScrollIndicator={false}
			style={styles.filterContainer}
			contentContainerStyle={styles.filterContent}
		>
			<TouchableOpacity
				style={[styles.filterButton, daysThreshold === 30 && styles.filterButtonActive]}
				onPress={() => setDaysThreshold(30)}
			>
				<Text style={[styles.filterButtonText, daysThreshold === 30 && styles.filterButtonTextActive]}>
					30 Days
				</Text>
			</TouchableOpacity>
			<TouchableOpacity
				style={[styles.filterButton, daysThreshold === 14 && styles.filterButtonActive]}
				onPress={() => setDaysThreshold(14)}
			>
				<Text style={[styles.filterButtonText, daysThreshold === 14 && styles.filterButtonTextActive]}>
					14 Days
				</Text>
			</TouchableOpacity>
			<TouchableOpacity
				style={[styles.filterButton, daysThreshold === 7 && styles.filterButtonActive]}
				onPress={() => setDaysThreshold(7)}
			>
				<Text style={[styles.filterButtonText, daysThreshold === 7 && styles.filterButtonTextActive]}>
					7 Days
				</Text>
			</TouchableOpacity>
		</ScrollView>
	);

	return (
		<View style={styles.container}>
			{/* Header */}
			<View style={styles.header}>
				<Text style={styles.headerTitle}>Expiring Products</Text>
				<Text style={styles.headerSubtitle}>
					Products approaching expiry dates
				</Text>
			</View>

			{/* Filter Buttons */}
			{renderFilterButtons()}

			{/* Loading State */}
			{loading ? (
				<View style={styles.loadingContainer}>
					<ActivityIndicator size="large" color="#3B82F6" />
					<Text style={styles.loadingText}>Loading expiring products...</Text>
				</View>
			) : error ? (
				<View style={styles.errorContainer}>
					<Text style={styles.errorText}>{error}</Text>
					<TouchableOpacity
						style={styles.retryButton}
						onPress={() => fetchExpiringProducts(userBranch ?? undefined, daysThreshold)}
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
	filterContainer: {
		backgroundColor: '#FFFFFF',
		borderBottomWidth: 1,
		borderBottomColor: '#E5E7EB',
	},
	filterContent: {
		paddingHorizontal: 16,
		paddingVertical: 12,
		gap: 8,
	},
	filterButton: {
		paddingHorizontal: 16,
		paddingVertical: 8,
		borderRadius: 20,
		backgroundColor: '#F3F4F6',
		borderWidth: 1,
		borderColor: '#E5E7EB',
	},
	filterButtonActive: {
		backgroundColor: '#3B82F6',
		borderColor: '#3B82F6',
	},
	filterButtonText: {
		fontSize: 14,
		fontWeight: '600',
		color: '#6B7280',
	},
	filterButtonTextActive: {
		color: '#FFFFFF',
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
		marginHorizontal: 16,
		marginBottom: 12,
		borderRadius: 8,
		padding: 16,
		borderWidth: 2,
		shadowColor: '#000',
		shadowOffset: { width: 0, height: 1 },
		shadowOpacity: 0.1,
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
		fontWeight: '700',
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
	alertBadge: {
		paddingHorizontal: 10,
		paddingVertical: 4,
		borderRadius: 12,
		alignSelf: 'flex-start',
	},
	alertText: {
		fontSize: 10,
		fontWeight: '700',
		color: '#FFFFFF',
		textTransform: 'uppercase',
	},
	expiryInfo: {
		gap: 6,
	},
	infoRow: {
		flexDirection: 'row',
		alignItems: 'center',
	},
	infoLabel: {
		fontSize: 13,
		fontWeight: '500',
	},
	infoValue: {
		fontSize: 13,
		fontWeight: '600',
		marginLeft: 8,
	},
	daysValue: {
		fontSize: 14,
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

export default ExpiringProductsScreen;
