/**
 * Tests for useStockWebSocket Hook
 * Story 4.2, Task 7: Create WebSocket Client Hook
 *
 * These tests verify the hook's connection management, auto-reconnect,
 * subscription management, and event handling functionality.
 */

import { renderHook, act, waitFor } from '@testing-library/react';
import { useStockWebSocket, ConnectionState, StockUpdatedEvent } from './useStockWebSocket';

// Mock WebSocket for testing
class MockWebSocket {
  static url: string = '';
  static instances: MockWebSocket[] = [];
  readyState: number = WebSocket.CONNECTING;
  onopen: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;

  constructor(url: string) {
    MockWebSocket.url = url;
    MockWebSocket.instances.push(this);
  }

  // Simulate successful connection
  open() {
    this.readyState = WebSocket.OPEN;
    if (this.onopen) {
      this.onopen(new Event('open'));
    }
  }

  // Simulate receiving a message
  message(data: any) {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { data: JSON.stringify(data) }));
    }
  }

  // Simulate connection close
  close(code?: number, reason?: string) {
    this.readyState = WebSocket.CLOSED;
    if (this.onclose) {
      this.onclose(new CloseEvent('close', { code, reason }));
    }
  }

  // Simulate error
  error() {
    if (this.onerror) {
      this.onerror(new Event('error'));
    }
  }

  static reset() {
    MockWebSocket.instances = [];
    MockWebSocket.url = '';
  }
}

// Mock global WebSocket
global.WebSocket = MockWebSocket as any;

describe('useStockWebSocket', () => {
  beforeEach(() => {
    MockWebSocket.reset();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  describe('Connection Management', () => {
    test('should not connect without token', () => {
      const { result } = renderHook(() =>
        useStockWebSocket({ token: null })
      );

      expect(result.current.connectionState).toBe('disconnected');
      expect(result.current.isConnected).toBe(false);
      expect(MockWebSocket.instances.length).toBe(0);
    });

    test('should connect when token is provided', () => {
      const { result } = renderHook(() =>
        useStockWebSocket({ token: 'valid-jwt-token' })
      );

      // Should create WebSocket instance
      expect(MockWebSocket.instances.length).toBe(1);
      expect(result.current.connectionState).toBe('connecting');

      // Simulate connection success
      act(() => {
        MockWebSocket.instances[0].open();
      });

      expect(result.current.connectionState).toBe('connected');
      expect(result.current.isConnected).toBe(true);
    });

    test('should build correct WebSocket URL with token and branches', () => {
      renderHook(() =>
        useStockWebSocket({
          token: 'test-token',
          branches: [1, 2, 3],
          wsUrl: 'localhost:8080',
        })
      );

      const url = MockWebSocket.url;
      expect(url).toContain('token=test-token');
      expect(url).toContain('branches=1%2C2%2C3'); // URL-encoded comma
    });

    test('should handle manual disconnect', () => {
      const { result } = renderHook(() =>
        useStockWebSocket({ token: 'test-token' })
      );

      // Connect first
      act(() => {
        MockWebSocket.instances[0].open();
      });

      expect(result.current.isConnected).toBe(true);

      // Disconnect
      act(() => {
        result.current.disconnect();
      });

      expect(result.current.connectionState).toBe('disconnected');
      expect(result.current.isConnected).toBe(false);
    });

    test('should support manual reconnect', () => {
      const { result } = renderHook(() =>
        useStockWebSocket({ token: 'test-token' })
      );

      // Connect first
      act(() => {
        MockWebSocket.instances[0].open();
      });

      expect(result.current.isConnected).toBe(true);

      // Manually reconnect
      act(() => {
        result.current.reconnect();
      });

      // Should create new WebSocket instance
      expect(MockWebSocket.instances.length).toBe(2);
    });
  });

  describe('Event Handling', () => {
    test('should receive and parse stock update events', () => {
      const onStockUpdate = jest.fn();
      const { result } = renderHook(() =>
        useStockWebSocket({
          token: 'test-token',
          onStockUpdate,
        })
      );

      // Connect
      act(() => {
        MockWebSocket.instances[0].open();
      });

      // Simulate receiving stock update
      const mockEvent: StockUpdatedEvent = {
        productId: 123,
        branchId: 1,
        sku: 'TEST-123',
        name: 'Test Product',
        oldStock: 100,
        newStock: 95,
        change: -5,
        updatedBy: 'John Doe',
        updatedAt: '2026-05-19T10:30:00Z',
      };

      act(() => {
        MockWebSocket.instances[0].message({
          event: 'stock.updated',
          data: mockEvent,
        });
      });

      // Verify state updated
      expect(result.current.lastStockUpdate).toEqual(mockEvent);
      expect(result.current.stockUpdateHistory).toHaveLength(1);
      expect(result.current.stockUpdateHistory[0]).toEqual(mockEvent);

      // Verify callback called
      expect(onStockUpdate).toHaveBeenCalledWith(mockEvent);
    });

    test('should maintain stock update history', () => {
      const { result } = renderHook(() =>
        useStockWebSocket({ token: 'test-token' })
      );

      // Connect
      act(() => {
        MockWebSocket.instances[0].open();
      });

      // Send multiple events
      for (let i = 0; i < 5; i++) {
        act(() => {
          MockWebSocket.instances[0].message({
            event: 'stock.updated',
            data: {
              productId: i,
              branchId: 1,
              sku: `TEST-${i}`,
              name: `Product ${i}`,
              oldStock: 100,
              newStock: 95 - i,
              change: -5,
              updatedBy: 'Test',
              updatedAt: '2026-05-19T10:30:00Z',
            },
          });
        });
      }

      expect(result.current.stockUpdateHistory).toHaveLength(5);
    });

    test('should limit stock update history to 100 events', () => {
      const { result } = renderHook(() =>
        useStockWebSocket({ token: 'test-token' })
      );

      // Connect
      act(() => {
        MockWebSocket.instances[0].open();
      });

      // Send more than 100 events
      for (let i = 0; i < 150; i++) {
        act(() => {
          MockWebSocket.instances[0].message({
            event: 'stock.updated',
            data: {
              productId: i,
              branchId: 1,
              sku: `TEST-${i}`,
              name: `Product ${i}`,
              oldStock: 100,
              newStock: 95,
              change: -5,
              updatedBy: 'Test',
              updatedAt: '2026-05-19T10:30:00Z',
            },
          });
        });
      }

      // Should keep only last 100
      expect(result.current.stockUpdateHistory).toHaveLength(100);
    });
  });

  describe('Auto-Reconnect', () => {
    test('should reconnect automatically on connection loss', () => {
      const { result } = renderHook(() =>
        useStockWebSocket({
          token: 'test-token',
          autoReconnect: true,
        })
      );

      // Connect first
      act(() => {
        MockWebSocket.instances[0].open();
      });

      expect(result.current.isConnected).toBe(true);
      expect(result.current.connectionState).toBe('connected');

      // Simulate connection loss
      act(() => {
        MockWebSocket.instances[0].close();
      });

      // Should transition to reconnecting
      expect(result.current.connectionState).toBe('reconnecting');

      // Fast-forward past reconnection delay
      act(() => {
        jest.advanceTimersByTime(1000);
      });

      // Should create new WebSocket instance
      expect(MockWebSocket.instances.length).toBe(2);
    });

    test('should not reconnect when autoReconnect is disabled', () => {
      const { result } = renderHook(() =>
        useStockWebSocket({
          token: 'test-token',
          autoReconnect: false,
        })
      );

      // Connect first
      act(() => {
        MockWebSocket.instances[0].open();
      });

      // Simulate connection loss
      act(() => {
        MockWebSocket.instances[0].close();
      });

      // Should transition to disconnected (not reconnecting)
      expect(result.current.connectionState).toBe('disconnected');

      // Should NOT create new WebSocket instance
      expect(MockWebSocket.instances.length).toBe(1);
    });

    test('should use exponential backoff for reconnection', () => {
      const onConnectionStateChange = jest.fn();
      renderHook(() =>
        useStockWebSocket({
          token: 'test-token',
          autoReconnect: true,
          onConnectionStateChange,
        })
      );

      // First connection attempt
      let ws = MockWebSocket.instances[0];
      act(() => ws.open());

      // First failure - should retry in 1s
      act(() => ws.close());

      // Advance to 1s - should reconnect
      act(() => jest.advanceTimersByTime(1000));
      expect(MockWebSocket.instances.length).toBe(2);

      // Second failure - should retry in 2s
      ws = MockWebSocket.instances[1];
      act(() => ws.open());
      act(() => ws.close());

      act(() => jest.advanceTimersByTime(2000));
      expect(MockWebSocket.instances.length).toBe(3);

      // Third failure - should retry in 4s
      ws = MockWebSocket.instances[2];
      act(() => ws.open());
      act(() => ws.close());

      act(() => jest.advanceTimersByTime(4000));
      expect(MockWebSocket.instances.length).toBe(4);
    });
  });

  describe('Connection State Changes', () => {
    test('should call onConnectionStateChange callback', () => {
      const onConnectionStateChange = jest.fn();
      const { result } = renderHook(() =>
        useStockWebSocket({
          token: 'test-token',
          onConnectionStateChange,
        })
      );

      // Initial state
      expect(onConnectionStateChange).toHaveBeenCalledWith('connecting');

      // Connected
      act(() => {
        MockWebSocket.instances[0].open();
      });

      expect(onConnectionStateChange).toHaveBeenCalledWith('connected');

      // Disconnected
      act(() => {
        result.current.disconnect();
      });

      expect(onConnectionStateChange).toHaveBeenCalledWith('disconnected');
    });
  });

  describe('Branch Filtering', () => {
    test('should include branches in WebSocket URL', () => {
      renderHook(() =>
        useStockWebSocket({
          token: 'test-token',
          branches: [1, 2, 3],
        })
      );

      const url = MockWebSocket.url;
      expect(url).toContain('branches=1%2C2%2C3');
    });

    test('should not include branches parameter when empty', () => {
      renderHook(() =>
        useStockWebSocket({
          token: 'test-token',
          branches: [],
        })
      );

      const url = MockWebSocket.url;
      expect(url).not.toContain('branches=');
    });

    test('should reconnect when branches change', () => {
      const { result, rerender } = renderHook(
        ({ branches }) =>
          useStockWebSocket({
            token: 'test-token',
            branches,
          }),
        { initialProps: { branches: [1, 2] } }
      );

      // Connect with initial branches
      act(() => {
        MockWebSocket.instances[0].open();
      });

      expect(MockWebSocket.instances.length).toBe(1);

      // Change branches
      rerender({ branches: [1, 2, 3] });

      // Should create new WebSocket with updated branches
      expect(MockWebSocket.instances.length).toBe(2);
      expect(MockWebSocket.url).toContain('branches=1%2C2%2C3');
    });
  });

  describe('Error Handling', () => {
    test('should handle WebSocket errors', () => {
      const onError = jest.fn();
      renderHook(() =>
        useStockWebSocket({
          token: 'test-token',
          onError,
        })
      );

      // Simulate error
      act(() => {
        MockWebSocket.instances[0].error();
      });

      expect(result.current.connectionState).toBe('error');
      expect(onError).toHaveBeenCalled();
    });

    test('should handle malformed messages gracefully', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

      renderHook(() =>
        useStockWebSocket({ token: 'test-token' })
      );

      // Connect
      act(() => {
        MockWebSocket.instances[0].open();
      });

      // Send malformed message
      act(() => {
        MockWebSocket.instances[0].message('invalid json{');
      });

      // Should not crash, just log error
      expect(consoleSpy).toHaveBeenCalled();
      consoleSpy.mockRestore();
    });
  });

  describe('Cleanup', () => {
    test('should disconnect on unmount', () => {
      const { result, unmount } = renderHook(() =>
        useStockWebSocket({ token: 'test-token' })
      );

      // Connect
      act(() => {
        MockWebSocket.instances[0].open();
      });

      expect(result.current.isConnected).toBe(true);

      // Unmount
      act(() => {
        unmount();
      });

      // Should close connection and not attempt reconnection
      expect(result.current.connectionState).toBe('disconnected');
    });

    test('should clear reconnection timeout on unmount', () => {
      jest.useRealTimers();

      const { unmount } = renderHook(() =>
        useStockWebSocket({ token: 'test-token' })
      );

      // Simulate connection loss to trigger reconnection
      act(() => {
        MockWebSocket.instances[0].close();
      });

      // Unmount before reconnection
      unmount();

      // Wait to ensure no reconnection attempt
      setTimeout(() => {
        expect(MockWebSocket.instances.length).toBe(1); // Only initial instance
      }, 2000);
    });
  });
});
