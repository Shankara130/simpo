/**
 * useNetworkStatus Hook Tests
 * Story 8.1: Implement Local SQLite Storage for Offline Transactions
 *
 * Test Coverage:
 * - Network state subscription
 * - isConnected state management
 * - Network state changes
 * - Cleanup on unmount
 * - Error handling
 */

import { renderHook, waitFor } from '@testing-library/react-native';
import { useNetworkStatus } from './useNetworkStatus';
import NetInfo from '@react-native-community/netinfo';

// Mock @react-native-community/netinfo
jest.mock('@react-native-community/netinfo', () => ({
  fetch: jest.fn(),
  addEventListener: jest.fn(),
  NetInfoStateType: {
    wifi: 'wifi',
    cellular: 'cellular',
    none: 'none',
    unknown: 'unknown',
  },
}));

describe('useNetworkStatus Hook', () => {
  const mockNetInfoState = {
    type: 'wifi',
    isConnected: true,
    isInternetReachable: true,
    details: {
      isConnectionExpensive: false,
    },
  };

  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('should initialize with connected status', async () => {
    (NetInfo.fetch as jest.Mock).mockResolvedValue({
      ...mockNetInfoState,
      isConnected: true,
    });

    const { result } = renderHook(() => useNetworkStatus());

    await waitFor(() => {
      expect(result.current.isConnected).toBe(true);
    });
  });

  it('should initialize with disconnected status when offline', async () => {
    (NetInfo.fetch as jest.Mock).mockResolvedValue({
      ...mockNetInfoState,
      isConnected: false,
      isInternetReachable: false,
    });

    const { result } = renderHook(() => useNetworkStatus());

    await waitFor(() => {
      expect(result.current.isConnected).toBe(false);
    });
  });

  it('should subscribe to network state changes', async () => {
    let listenerCallback: ((state: any) => void) | null = null;

    (NetInfo.fetch as jest.Mock).mockResolvedValue(mockNetInfoState);
    (NetInfo.addEventListener as jest.Mock).mockImplementation((callback) => {
      listenerCallback = callback as any;
      return { remove: jest.fn() }; // Return subscription object with remove method
    });

    const { result } = renderHook(() => useNetworkStatus());

    // Wait for initial state
    await waitFor(() => {
      expect(result.current.isConnected).toBe(true);
    });

    // Simulate network disconnection
    if (listenerCallback) {
      listenerCallback({
        ...mockNetInfoState,
        isConnected: false,
      });
    }

    await waitFor(() => {
      expect(result.current.isConnected).toBe(false);
    });
  });

  it('should debounce network state changes (500ms)', async () => {
    let listenerCallback: ((state: any) => void) | null = null;

    (NetInfo.fetch as jest.Mock).mockResolvedValue(mockNetInfoState);
    (NetInfo.addEventListener as jest.Mock).mockImplementation((callback) => {
      listenerCallback = callback as any;
      return { remove: jest.fn() };
    });

    const { result } = renderHook(() => useNetworkStatus());

    await waitFor(() => {
      expect(result.current.isConnected).toBe(true);
    });

    // Trigger rapid state changes
    if (listenerCallback) {
      listenerCallback({ ...mockNetInfoState, isConnected: false });
      jest.advanceTimersByTime(200);
      listenerCallback({ ...mockNetInfoState, isConnected: true });
      jest.advanceTimersByTime(200);
      listenerCallback({ ...mockNetInfoState, isConnected: false });
    }

    // State should not change before debounce period
    expect(result.current.isConnected).toBe(true);

    // Advance past debounce period
    jest.advanceTimersByTime(300);
    jest.runAllTimers();

    await waitFor(() => {
      expect(result.current.isConnected).toBe(false);
    });
  });

  it('should cleanup subscription on unmount', async () => {
    const removeMock = jest.fn();

    (NetInfo.fetch as jest.Mock).mockResolvedValue(mockNetInfoState);
    (NetInfo.addEventListener as jest.Mock).mockReturnValue({
      remove: removeMock,
    });

    const { unmount } = renderHook(() => useNetworkStatus());

    await waitFor(() => {
      expect(NetInfo.addEventListener).toHaveBeenCalled();
    });

    unmount();

    expect(removeMock).toHaveBeenCalled();
  });

  it('should handle NetInfo fetch errors gracefully', async () => {
    (NetInfo.fetch as jest.Mock).mockRejectedValue(
      new Error('Network check failed')
    );

    const { result } = renderHook(() => useNetworkStatus());

    // Should default to true on error (optimistic approach)
    await waitFor(() => {
      expect(result.current.isConnected).toBe(true);
    });
  });

  it('should handle addEventListener errors gracefully', async () => {
    (NetInfo.fetch as jest.Mock).mockResolvedValue(mockNetInfoState);
    (NetInfo.addEventListener as jest.Mock).mockImplementation(() => {
      throw new Error('Subscription failed');
    });

    const { result } = renderHook(() => useNetworkStatus());

    // Should not throw, just log error
    await waitFor(() => {
      expect(result.current.isConnected).toBeDefined();
    });
  });
});
