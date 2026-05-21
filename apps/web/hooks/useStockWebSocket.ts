/**
 * useStockWebSocket Hook
 * Story 4.2, Task 7: Create WebSocket Client Hook (AC: 1, 5)
 *
 * This hook manages WebSocket connections for real-time stock updates.
 * It provides auto-reconnect, subscription management, and connection state handling.
 */

import { useState, useEffect, useCallback, useRef } from 'react';

// Connection states for the WebSocket
export type ConnectionState =
  | 'connecting'    // Initial connection attempt
  | 'connected'     // Successfully connected and receiving events
  | 'disconnected'  // Intentionally disconnected
  | 'reconnecting'  // Attempting to reconnect after connection loss
  | 'error';        // Connection error (will try to reconnect)

// Stock update event payload from backend
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

// Low stock event payload from backend
// Story 4.4, AC2, AC4: Low stock notification event structure
export interface LowStockEvent {
  productId: number;
  sku: string;
  productName: string;
  currentStock: number;
  reorderThreshold: number;
  suggestedOrderQty: number;
  branchId: number;
  branchName: string;
}

// Expiry alert event payload from backend
// Story 4.5, AC4, AC6: Expiry date alert notification event structure
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

// Union type for all stock events
export type StockEvent = StockUpdatedEvent | LowStockEvent | ExpiryEvent;

// Configuration options for the hook
interface UseStockWebSocketOptions {
  // JWT token for authentication
  token: string | null;
  // Branch IDs to filter events (empty array = all branches)
  branches?: number[];
  // WebSocket server URL (defaults to current host)
  wsUrl?: string;
  // Auto-reconnect with exponential backoff
  autoReconnect?: boolean;
  // Maximum reconnection delay (ms)
  maxReconnectDelay?: number;
  // Event handler for stock updates
  onStockUpdate?: (event: StockUpdatedEvent) => void;
  // Story 4.4: Event handler for low stock notifications
  onLowStock?: (event: LowStockEvent) => void;
  // Story 4.5: Event handler for expiry date alerts
  onExpiry?: (event: ExpiryEvent) => void;
  // Connection state change handler
  onConnectionStateChange?: (state: ConnectionState) => void;
  // Error handler
  onError?: (error: Error) => void;
}

// Return type for the hook
interface UseStockWebSocketReturn {
  // Current connection state
  connectionState: ConnectionState;
  // Whether currently connected
  isConnected: boolean;
  // Latest stock update received (if any)
  lastStockUpdate: StockUpdatedEvent | null;
  // All received stock updates (in-memory history)
  stockUpdateHistory: StockUpdatedEvent[];
  // Manually trigger reconnection
  reconnect: () => void;
  // Manually disconnect
  disconnect: () => void;
  // Manually connect
  connect: () => void;
}

// Reconnection delays (exponential backoff): 1s, 2s, 4s, 8s, 16s, max 30s
const RECONNECT_DELAYS = [1000, 2000, 4000, 8000, 16000, 30000];

/**
 * Custom hook for managing WebSocket connections to real-time stock updates
 * Story 4.2, Task 7: Create WebSocket Client Hook
 */
export function useStockWebSocket(options: UseStockWebSocketOptions): UseStockWebSocketReturn {
  const {
    token,
    branches = [],
    wsUrl,
    autoReconnect = true,
    maxReconnectDelay = 30000,
    onStockUpdate,
    onLowStock,
    onExpiry,
    onConnectionStateChange,
    onError,
  } = options;

  // Connection state
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');
  const [lastStockUpdate, setLastStockUpdate] = useState<StockUpdatedEvent | null>(null);
  const [stockUpdateHistory, setStockUpdateHistory] = useState<StockUpdatedEvent[]>([]);

  // Refs to avoid stale closures
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptRef = useRef(0);
  const isManualDisconnectRef = useRef(false);

  /**
   * Build WebSocket URL with authentication and branch filters
   * Story 4.2, Task 7.2: WebSocket connection with JWT token
   */
  const buildWsUrl = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = wsUrl || window.location.host;
    const path = '/api/v1/products/stock/subscribe';

    // Build query parameters
    const params = new URLSearchParams();
    if (token) {
      params.append('token', token);
    }
    if (branches.length > 0) {
      params.append('branches', branches.join(','));
    }

    return `${protocol}//${host}${path}?${params.toString()}`;
  }, [token, branches, wsUrl]);

  /**
   * Handle WebSocket message
   * Story 4.2, Task 7.4: Add event handlers for stock updates
   * Story 4.4: Extended to handle stock.low events for low stock notifications
   * Story 4.5: Extended to handle product.expiry events for expiry date alerts
   */
  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const data = JSON.parse(event.data);

      // Validate event structure and handle based on event type
      if (data.event === 'stock.updated' && data.data) {
        const stockEvent: StockUpdatedEvent = data.data;

        // Update state
        setLastStockUpdate(stockEvent);
        setStockUpdateHistory(prev => [stockEvent, ...prev].slice(0, 100)); // Keep last 100

        // Call external handler if provided
        if (onStockUpdate) {
          onStockUpdate(stockEvent);
        }
      } else if (data.event === 'stock.low' && data.data) {
        // Story 4.4: Handle low stock notifications
        const lowStockEvent: LowStockEvent = data.data;

        // Call external handler if provided
        if (onLowStock) {
          onLowStock(lowStockEvent);
        }
      } else if (data.event === 'product.expiry' && data.data) {
        // Story 4.5: Handle expiry date alerts
        const expiryEvent: ExpiryEvent = data.data;

        // Call external handler if provided
        if (onExpiry) {
          onExpiry(expiryEvent);
        }
      }
    } catch (error) {
      console.error('[useStockWebSocket] Failed to parse message:', error);
      if (onError) {
        onError(error as Error);
      }
    }
  }, [onStockUpdate, onLowStock, onExpiry, onError]);

  /**
   * Handle WebSocket close
   * Story 4.2, Task 7.2: Auto-reconnect with exponential backoff
   * PATCH: Reconnection is fully implemented with exponential backoff (1s, 2s, 4s, 8s, 16s, max 30s)
   * - Clears existing reconnection timeout to prevent duplicates
   * - Respects manual disconnect flag (won't reconnect if user manually disconnected)
   * - Uses exponential backoff delays up to maxReconnectDelay
   * - Updates connection state to 'reconnecting' during backoff
   */
  const handleClose = useCallback(() => {
    // Clear existing reconnection timeout
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    // If this was a manual disconnect, don't reconnect
    if (isManualDisconnectRef.current) {
      setConnectionState('disconnected');
      return;
    }

    // Auto-reconnect if enabled
    if (autoReconnect) {
      const delay = Math.min(
        RECONNECT_DELAYS[Math.min(reconnectAttemptRef.current, RECONNECT_DELAYS.length - 1)],
        maxReconnectDelay
      );

      setConnectionState('reconnecting');
      onConnectionStateChange?.('reconnecting');

      reconnectTimeoutRef.current = setTimeout(() => {
        reconnectAttemptRef.current++;
        connect();
      }, delay);
    } else {
      setConnectionState('disconnected');
      onConnectionStateChange?.('disconnected');
    }
  }, [autoReconnect, maxReconnectDelay, onConnectionStateChange]);

  /**
   * Handle WebSocket error
   * Story 4.2, Task 7.5: Handle connection state (error)
   */
  const handleError = useCallback((error: Event) => {
    console.error('[useStockWebSocket] WebSocket error:', error);
    setConnectionState('error');
    onConnectionStateChange?.('error');
    onError?.(new Error('WebSocket connection error'));
  }, [onConnectionStateChange, onError]);

  /**
   * Establish WebSocket connection
   * Story 4.2, Task 7.2: Implement WebSocket connection
   */
  const connect = useCallback(() => {
    // Don't connect if already connected or connecting
    if (wsRef.current && (wsRef.current.readyState === WebSocket.OPEN || wsRef.current.readyState === WebSocket.CONNECTING)) {
      return;
    }

    // Reset manual disconnect flag
    isManualDisconnectRef.current = false;

    // Update state
    setConnectionState('connecting');
    onConnectionStateChange?.('connecting');

    try {
      // Create WebSocket connection
      const url = buildWsUrl();
      const ws = new WebSocket(url);

      // Store reference
      wsRef.current = ws;

      // Set up event handlers
      ws.onopen = () => {
        // Reset reconnection attempt counter on successful connection
        reconnectAttemptRef.current = 0;
        setConnectionState('connected');
        onConnectionStateChange?.('connected');
      };

      ws.onmessage = handleMessage;
      ws.onclose = handleClose;
      ws.onerror = handleError;

    } catch (error) {
      console.error('[useStockWebSocket] Failed to create WebSocket:', error);
      setConnectionState('error');
      onConnectionStateChange?.('error');
      onError?.(error as Error);
    }
  }, [buildWsUrl, handleMessage, handleClose, handleError, onConnectionStateChange, onError]);

  /**
   * Disconnect WebSocket
   * Story 4.2, Task 7.3: Implement subscription management (subscribe/unsubscribe)
   */
  const disconnect = useCallback(() => {
    // Set manual disconnect flag to prevent auto-reconnect
    isManualDisconnectRef.current = true;

    // Clear reconnection timeout
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    // Close WebSocket connection
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    // Update state
    setConnectionState('disconnected');
    onConnectionStateChange?.('disconnected');
  }, [onConnectionStateChange]);

  /**
   * Manual reconnect
   * Story 4.2, Task 7.3: Implement subscription management
   */
  const reconnect = useCallback(() => {
    // Disconnect first if connected
    disconnect();

    // Reset reconnection attempt counter
    reconnectAttemptRef.current = 0;

    // Clear manual disconnect flag to allow reconnection
    isManualDisconnectRef.current = false;

    // Reconnect after a short delay
    setTimeout(() => {
      connect();
    }, 100);
  }, [disconnect, connect]);

  /**
   * Effect: Establish connection on mount when token is available
   * Story 4.2, Task 7.2: Implement WebSocket connection
   */
  useEffect(() => {
    // Only connect if we have a token
    if (!token) {
      setConnectionState('disconnected');
      return;
    }

    connect();

    // Cleanup on unmount
    return () => {
      // Set manual disconnect to prevent auto-reconnect during cleanup
      isManualDisconnectRef.current = true;

      // Clear reconnection timeout
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }

      // Close WebSocket
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [token]); // Only reconnect when token changes

  /**
   * Effect: Handle branch filter changes
   * Story 4.2, Task 7.3: Implement subscription management
   */
  useEffect(() => {
    // If already connected and branches change, reconnect with new filters
    if (connectionState === 'connected' && wsRef.current) {
      reconnect();
    }
  }, [branches, connectionState, reconnect]);

  return {
    connectionState,
    isConnected: connectionState === 'connected',
    lastStockUpdate,
    stockUpdateHistory,
    reconnect,
    disconnect,
    connect,
  };
}

/**
 * Default export for convenience
 */
export default useStockWebSocket;
