/**
 * Notification Service
 * Story 4.4, Task 11: Implement Push Notification Service (AC: 3)
 *
 * This service handles push notifications for low stock alerts using Expo Notifications.
 * Features:
 * - Request push notification permissions on app startup
 * - Register device token with backend (future: device token management)
 * - Handle incoming push notifications for stock.low events
 * - Display local notification with product info and actionable message
 */

import * as Notifications from 'expo-notifications';
import * as Device from 'expo-device';
import { Platform } from 'react-native';

// Low stock notification data structure
// Story 4.4, AC4: Low stock notification event structure
export interface LowStockNotificationData {
	productId: number;
	sku: string;
	productName: string;
	currentStock: number;
	reorderThreshold: number;
	suggestedOrderQty: number;
	branchId: number;
	branchName: string;
}

// Push notification payload structure
interface NotificationPayload {
	type: string;
	data?: LowStockNotificationData;
}

class NotificationService {
	private isInitialized = false;

	/**
	 * Initialize notification service
	 * Story 4.4, Task 11.2: Request push notification permissions on app startup
	 */
	async initialize(): Promise<void> {
		if (this.isInitialized) {
			return;
		}

		if (Platform.OS === 'android') {
			// Android: Create notification channel
			await Notifications.setNotificationChannelAsync('stock-alerts', {
				name: 'Stock Alerts',
				importance: Notifications.AndroidImportance.HIGH,
				vibrationPattern: [0, 250, 250, 250],
				lightColor: '#FF5722',
			});
		}

		// Request permissions
		const { status: existingStatus } = await Notifications.getPermissionsAsync();
		let finalStatus = existingStatus;

		if (existingStatus !== 'granted') {
			const { status } = await Notifications.requestPermissionsAsync();
			finalStatus = status;
		}

		if (finalStatus !== 'granted') {
			console.warn('Push notification permissions not granted');
		}

		// Story 4.4, Task 11.1: Set up Expo Notifications (already configured in app.json)
		// Configure notification handlers
		this.setupNotificationHandlers();

		this.isInitialized = true;
		console.log('Notification service initialized');
	}

	/**
	 * Setup notification handlers
	 * Story 4.4, Task 11.5: Handle incoming push notifications for stock.low events
	 */
	private setupNotificationHandlers(): void {
		// Handle notification interactions (tap, dismiss, etc.)
		Notifications.addNotificationReceivedListener((notification) => {
			console.log('Notification received:', notification);
		});

		Notifications.addNotificationResponseReceivedListener((response) => {
			console.log('Notification response:', response);
			// Handle notification response (e.g., user tapped on notification)
			this.handleNotificationResponse(response);
		});

		// Handle foreground notifications
		Notifications.setNotificationHandler({
			handleNotification: async (notification) => {
				console.log('Foreground notification:', notification);
				// Show notification in foreground
				await Notifications.presentNotificationAsync({
					title: notification.request.content.title,
					body: notification.request.content.body,
					data: notification.request.content.data,
				});
			},
		});
	}

	/**
	 * Handle notification response when user taps on notification
	 */
	private handleNotificationResponse(response: Notifications.NotificationResponse): void {
		const { data } = response.notification.request.content;

		if (data && typeof data === 'object') {
			const notificationData = data as unknown;

			// Check if this is a low stock notification
			if (
				notificationData &&
				typeof notificationData === 'object' &&
				'productId' in notificationData
			) {
				const lowStockData = notificationData as LowStockNotificationData;

				// Navigate to low stock screen or product details
				// This would typically trigger navigation to the appropriate screen
				console.log('Navigate to product details:', lowStockData.productId);
			}
		}
	}

	/**
	 * Display local notification for low stock alert
	 * Story 4.4, Task 11.6: Display local notification with product info and actionable message
	 */
	async showLowStockAlert(data: LowStockNotificationData): Promise<void> {
		await this.scheduleLowStockNotification(data);
	}

	/**
	 * Schedule low stock notification
	 * Story 4.4, Task 11.6: Display local notification with product info and actionable message
	 */
	private async scheduleLowStockNotification(
		data: LowStockNotificationData
	): Promise<string> {
		const content = {
			title: 'Low Stock Alert',
			body: `${data.productName} (SKU: ${data.sku}) is running low at ${data.branchName}. Order ${data.suggestedOrderQty} units.`,
			data: data,
		};

		await Notifications.scheduleNotificationAsync({
			content,
			trigger: null, // Show immediately
			identifier: `low-stock-${data.productId}-${data.branchId}`,
		});

		return `low-stock-${data.productId}-${data.branchId}`;
	}

	/**
	 * Cancel a specific notification
	 */
	async cancelNotification(identifier: string): Promise<void> {
		await Notifications.cancelScheduledNotificationAsync(identifier);
	}

	/**
	 * Cancel all low stock notifications
	 */
	async cancelAllLowStockNotifications(): Promise<void> {
		// Note: Expo doesn't provide a way to cancel all notifications by pattern
		// You would need to track identifiers and cancel them individually
		console.log('Cancel all low stock notifications called');
	}

	/**
	 * Get notification permissions status
	 */
	async getPermissionsStatus(): Promise<Notifications.PermissionStatus> {
		const { status } = await Notifications.getPermissionsAsync();
		return status;
	}

	/**
	 * Request notification permissions
	 */
	async requestPermissions(): Promise<Notifications.PermissionStatus> {
		const { status } = await Notifications.requestPermissionsAsync();
		return status;
	}
}

// Export singleton instance
export const notificationService = new NotificationService();

// Default export
export default notificationService;
