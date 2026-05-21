/**
 * Real-Time Stock Service
 * Story 4.2, Task 11: Create Real-Time Stock Service (AC: 1, 5)
 *
 * This service manages WebSocket connections for real-time stock updates in React Native.
 * Features:
 * - WebSocket connection management
 * - Reconnection logic with exponential backoff
 * - Branch-based subscription filtering
 * - Offline detection and event queueing
 */

import { EventEmitter } from 'events';
import NetInfo from '@react-native-community/netinfo';

// Connection states
export type ConnectionState =
  | 'disconnected'
  | 'connecting'
  | 'connected'
  | 'reconnecting'
  | 'error';

// Stock update event from backend
export interface StockUpdatedEvent {
  productId: number;
  branchId: number;
  sku: string;
  name: string;
  oldStock: number;
  newStock: number;
  change: number;
  updatedBy: string;
  updatedAt: string;
}

// Expiry alert event from backend
// Story 4.5, Task 11.2: Subscribe to product.expiry events via realTimeStockService
export interface ExpiryEvent {
  productId: number;
  sku: string;
  productName: string;
  expiryDate: string;
  daysRemaining: number;
  alertLevel: 'warning' | 'critical' | 'urgent';
  branchId: number;
  branchName: string;
}

// Configuration options
interface RealTimeStockServiceConfig {
  // WebSocket server URL
  wsUrl: string;
  // JWT token for authentication
  token: string;
  // Branch IDs to filter events
  branches?: number[];
  // Reconnection settings
  autoReconnect?: boolean;
  maxReconnectDelay?: number;
  // Event queue size for offline mode
  offlineQueueSize?: number;
}

// Service events
interface ServiceEvents {
  'connectionStateChange': (state: ConnectionState) => void;
  'stockUpdate': (event: StockUpdatedEvent) => void;
  'expiry': (event: ExpiryEvent) => void;
  'error': (error: Error) => void;
}

/**
 * Real-Time Stock Service for React Native
 * Story 4.2, Task 11: WebSocket connection management with reconnection and offline detection
 */
declare interface RealTimeStockService {
  on<EventName extends keyof ServiceEvents>(
    event: EventName,
    listener: (...args: any[]) => void
  ): this;
  off<EventName extends keyof ServiceEvents>(
    event: EventName,
    listener: (...args: any[]) => void
  ): this;
  emit<EventName extends keyof ServiceEvents>(
    event: EventName,
    ...args: Parameters<ServiceEvents[EventName]>
  ): boolean;
}

class RealTimeStockServiceImpl extends EventEmitter implements RealTimeStockService {
  private ws: WebSocket | null = null;
  private config: RealTimeStockServiceConfig;
  private connectionState: ConnectionState = 'disconnected';
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private reconnectAttempts: number = 0;
  private isManualDisconnect: boolean = false;
  private isOnline: boolean = true;
  private offlineQueue: StockUpdatedEvent[] = [];

  // Reconnection delays (exponential backoff)
  private readonly RECONNECT_DELAYS = [1000, 2000, 4000, 8000, 16000, 30000];

  constructor(config: RealTimeStockServiceConfig) {
    super();
    this.config = {
      autoReconnect: true,
      maxReconnectDelay: 30000,
      offlineQueueSize: 100,
      ...config,
    };
  }

  /**
   * Get current connection state
   */
  getConnectionState(): ConnectionState {
    return this.connectionState;
  }

  /**
   * Check if currently connected
   */
  isConnected(): boolean {
    return this.connectionState === 'connected' && this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Build WebSocket URL with authentication
   * Story 4.2, Task 11.2: WebSocket connection management
   */
  private buildWsUrl(): string {
    const params = new URLSearchParams();
    params.append('token', this.config.token);

    if (this.config.branches && this.config.branches.length > 0) {
      params.append('branches', this.config.branches.join(','));
    }

    return `${this.config.wsUrl}?${params.toString()}`;
  }

  /**
   * Connect to WebSocket server
   * Story 4.2, Task 11.2: WebSocket connection management
   */
  connect(): void {
    // Don't connect if already connected or connecting
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    // Reset manual disconnect flag
    this.isManualDisconnect = false;

    // Update state
    this.setConnectionState('connecting');

    try {
      const url = this.buildWsUrl();
      this.ws = new WebSocket(url);

      // Set up event handlers
      this.ws.onopen = this.handleOpen.bind(this);
      this.ws.onmessage = this.handleMessage.bind(this);
      this.ws.onclose = this.handleClose.bind(this);
      this.ws.onerror = this.handleError.bind(this);

    } catch (error) {
      console.error('[RealTimeStockService] Failed to create WebSocket:', error);
      this.setConnectionState('error');
      this.emit('error', error as Error);
    }
  }

  /**
   * Disconnect from WebSocket server
   * Story 4.2, Task 11.3: Reconnection logic with exponential backoff
   */
  disconnect(): void {
    // Set manual disconnect flag to prevent auto-reconnect
    this.isManualDisconnect = true;

    // Clear reconnection timeout
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    // Close WebSocket
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.setConnectionState('disconnected');
  }

  /**
   * Manual reconnect
   */
  reconnect(): void {
    // Disconnect first if connected
    this.disconnect();

    // Reset reconnection attempt counter
    this.reconnectAttempts = 0;

    // Clear manual disconnect flag to allow reconnection
    this.isManualDisconnect = false;

    // Reconnect after a short delay
    setTimeout(() => {
      this.connect();
    }, 100);
  }

  /**
   * Handle WebSocket open event
   */
  private handleOpen(): void {
    // Reset reconnection attempt counter on successful connection
    this.reconnectAttempts = 0;
    this.setConnectionState('connected');

    // Process queued events from offline mode
    this.processOfflineQueue();
  }

  /**
   * Handle WebSocket message event
   * Story 4.2, Task 11.2: WebSocket connection management
   * Story 4.5, Task 11.2: Subscribe to product.expiry events via realTimeStockService
   */
  private handleMessage(event: MessageEvent): void {
    try {
      const data = JSON.parse(event.data);

      // Validate event structure
      if (data.event === 'stock.updated' && data.data) {
        const stockEvent: StockUpdatedEvent = data.data;

        // Emit stock update event
        this.emit('stockUpdate', stockEvent);

        // If offline, queue the event
        if (!this.isOnline) {
          this.queueOfflineEvent(stockEvent);
        }
      } else if (data.event === 'product.expiry' && data.data) {
        // Story 4.5, Task 11.2: Handle product.expiry events
        const expiryEvent: ExpiryEvent = data.data;
        this.emit('expiry', expiryEvent);
      }
    } catch (error) {
      console.error('[RealTimeStockService] Failed to parse message:', error);
      this.emit('error', error as Error);
    }
  }

  /**
   * Handle WebSocket close event
   * Story 4.2, Task 11.3: Reconnection logic with exponential backoff
   */
  private handleClose(): void {
    // Clear existing reconnection timeout
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    // If this was a manual disconnect, don't reconnect
    if (this.isManualDisconnect) {
      this.setConnectionState('disconnected');
      return;
    }

    // Auto-reconnect if enabled and online
    if (this.config.autoReconnect && this.isOnline) {
      const delay = Math.min(
        this.RECONNECT_DELAYS[Math.min(this.reconnectAttempts, this.RECONNECT_DELAYS.length - 1)],
        this.config.maxReconnectDelay
      );

      this.setConnectionState('reconnecting');

      this.reconnectTimeout = setTimeout(() => {
        this.reconnectAttempts++;
        this.connect();
      }, delay);
    } else {
      this.setConnectionState('disconnected');
    }
  }

  /**
   * Handle WebSocket error event
   * Story 4.2, Task 11.5: Offline detection and queueing
   */
  private handleError(error: Event): void {
    console.error('[RealTimeStockService] WebSocket error:', error);
    this.setConnectionState('error');
    this.emit('error', new Error('WebSocket connection error'));
  }

  /**
   * Update connection state and notify listeners
   */
  private setConnectionState(state: ConnectionState): void {
    this.connectionState = state;
    this.emit('connectionStateChange', state);
  }

  /**
   * Queue event when offline
   * Story 4.2, Task 11.5: Offline detection and queueing
   */
  private queueOfflineEvent(event: StockUpdatedEvent): void {
    const maxSize = this.config.offlineQueueSize || 100;

    this.offlineQueue.push(event);

    // Keep queue size under limit
    if (this.offlineQueue.length > maxSize) {
      this.offlineQueue = this.offlineQueue.slice(-maxSize);
    }

    console.log('[RealTimeStockService] Queued offline event. Queue size:', this.offlineQueue.length);
  }

  /**
   * Process queued events when coming back online
   * Story 4.2, Task 11.5: Offline detection and queueing
   */
  private processOfflineQueue(): void {
    if (this.offlineQueue.length === 0) {
      return;
    }

    console.log(`[RealTimeStockService] Processing ${this.offlineQueue.length} queued events`);

    // Emit all queued events
    const events = [...this.offlineQueue];
    this.offlineQueue = [];

    events.forEach(event => {
      this.emit('stockUpdate', event);
    });
  }

  /**
   * Start online/offline monitoring
   * Story 4.2, Task 11.5: Offline detection and queueing
   */
  async startOnlineMonitoring(): Promise<void> {
    // Check initial state
    const state = await NetInfo.fetch();
    this.isOnline = state.isConnected ?? false;
    console.log('[RealTimeStockService] Initial online status:', this.isOnline);

    // Subscribe to network status changes
    const unsubscribe = NetInfo.addEventListener(state => {
      const wasOnline = this.isOnline;
      this.isOnline = state.isConnected ?? false;

      console.log('[RealTimeStockService] Network status changed:', this.isOnline);

      // If we just came back online, process queued events
      if (!wasOnline && this.isOnline && this.isConnected()) {
        this.processOfflineQueue();
      }

      // If we just went offline and were trying to reconnect, pause reconnection
      if (wasOnline && !this.isOnline && this.connectionState === 'reconnecting') {
        if (this.reconnectTimeout) {
          clearTimeout(this.reconnectTimeout);
          this.reconnectTimeout = null;
        }
        this.setConnectionState('disconnected');
      }

      // If we just came back online and were disconnected, reconnect
      if (!wasOnline && this.isOnline && this.connectionState === 'disconnected') {
        this.connect();
      }
    });

    // Return unsubscribe function for cleanup
    return () => unsubscribe();
  }

  /**
   * Clean up resources
   */
  destroy(): void {
    this.disconnect();
    this.removeAllListeners();
  }
}

/**
 * Factory function to create a real-time stock service instance
 * Story 4.2, Task 11: Create realTimeStockService.ts
 */
export function createRealTimeStockService(
  config: RealTimeStockServiceConfig
): RealTimeStockServiceImpl {
  return new RealTimeStockServiceImpl(config);
}

/**
 * Export types
 */
export type { RealTimeStockServiceConfig, RealTimeStockServiceImpl as RealTimeStockServiceInstance };
