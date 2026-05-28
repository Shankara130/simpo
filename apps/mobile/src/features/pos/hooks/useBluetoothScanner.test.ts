/**
 * useBluetoothScanner Hook Tests
 * Tests for Bluetooth scanner connection management and state
 */

import { renderHook, act, waitFor } from '@testing-library/react-native';
import { useBluetoothScanner } from './useBluetoothScanner';
import { BluetoothManager } from '../hardware/BluetoothManager';
import type { BluetoothDevice } from '../types/scanner.types';

// Mock BluetoothManager
jest.mock('../hardware/BluetoothManager');

// Mock React Native Vibration API
jest.mock('react-native', () => ({
  Vibration: {
    vibrate: jest.fn(),
  },
}));

describe('useBluetoothScanner Hook', () => {
  let mockBluetoothManager: jest.Mocked<BluetoothManager>;
  const mockOnDataReceived = jest.fn();
  const mockOnError = jest.fn();
  let mockDevice: BluetoothDevice;

  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();

    // Create mock device
    mockDevice = {
      id: 'device-123',
      name: 'Zebra Scanner',
      type: 'BLE',
      connected: false,
      paired: false,
      rssi: -60,
    };

    // Create mock instance
    mockBluetoothManager = {
      requestPermissions: jest.fn().mockResolvedValue(true),
      startDiscovery: jest.fn().mockResolvedValue(undefined),
      stopDiscovery: jest.fn().mockResolvedValue(undefined),
      connectToDevice: jest.fn().mockResolvedValue(undefined),
      cancelDeviceConnection: jest.fn().mockResolvedValue(undefined),
      disconnectDevice: jest.fn().mockResolvedValue(undefined),
      getPairedDevices: jest.fn().mockResolvedValue([]),
      getDiscoveredDevices: jest.fn().mockResolvedValue([]),
      startConnectionMonitoring: jest.fn().mockResolvedValue(undefined),
      stopConnectionMonitoring: jest.fn(),
      destroy: jest.fn(),
      enableAutoReconnect: jest.fn(),
      autoReconnect: jest.fn().mockResolvedValue(undefined),
    } as jest.Mocked<BluetoothManager>;

    // Mock BluetoothManager constructor
    (BluetoothManager as jest.MockedClass<typeof BluetoothManager>).mockImplementation(
      (callbacks) => {
        // Store callbacks so tests can invoke them
        (mockBluetoothManager as any)._callbacks = callbacks;
        return mockBluetoothManager;
      }
    );
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  const invokeCallback = (callbackName: string, ...args: any[]) => {
    const callbacks = (mockBluetoothManager as any)._callbacks;
    if (callbacks && callbacks[callbackName]) {
      callbacks[callbackName](...args);
    }
  };

  describe('Connection State Management', () => {
    it('should initialize with disconnected state', () => {
      const { result } = renderHook(() => useBluetoothScanner());

      expect(result.current.connectionState).toBe('disconnected');
    });

    it('should update connection state when Bluetooth manager calls callback', async () => {
      const { result } = renderHook(() =>
        useBluetoothScanner({
          onDataReceived: mockOnDataReceived,
        })
      );

      expect(result.current.connectionState).toBe('disconnected');

      // Simulate BluetoothManager calling state change callback
      await act(async () => {
        invokeCallback('onConnectionStateChanged', 'connected', 'device-123');
      });

      await waitFor(() => {
        expect(result.current.connectionState).toBe('connected');
      });
    });

    it('should reset state on unmount', () => {
      const { unmount, result } = renderHook(() => useBluetoothScanner());

      expect(result.current.connectionState).toBe('disconnected');

      unmount();

      // State should be cleaned up
      expect(mockBluetoothManager.destroy).toHaveBeenCalled();
    });
  });

  describe('Discovery Management', () => {
    it('should start device discovery when startDiscovery is called', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      await act(async () => {
        await result.current.startDiscovery();
      });

      expect(mockBluetoothManager.startDiscovery).toHaveBeenCalled();
    });

    it('should stop device discovery when stopDiscovery is called', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      await act(async () => {
        await result.current.stopDiscovery();
      });

      expect(mockBluetoothManager.stopDiscovery).toHaveBeenCalled();
    });

    it('should update isScanning state when starting discovery', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      await act(async () => {
        await result.current.startDiscovery();
      });

      expect(result.current.isScanning).toBe(true);
    });

    it('should update isScanning state when stopping discovery', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      await act(async () => {
        await result.current.startDiscovery();
      });

      await act(async () => {
        await result.current.stopDiscovery();
      });

      expect(result.current.isScanning).toBe(false);
    });
  });

  describe('Device Discovery Callbacks', () => {
    it('should add discovered devices to list', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      await act(async () => {
        await result.current.startDiscovery();
        invokeCallback('onDeviceDiscovered', mockDevice);
      });

      await waitFor(() => {
        expect(result.current.discoveredDevices.length).toBeGreaterThan(0);
      });
    });

    it('should avoid duplicate devices in discovered list', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      await act(async () => {
        await result.current.startDiscovery();
        invokeCallback('onDeviceDiscovered', mockDevice);
        invokeCallback('onDeviceDiscovered', mockDevice); // Same device again
      });

      await waitFor(() => {
        const count = result.current.discoveredDevices.filter(
          (d) => d.id === mockDevice.id
        ).length;
        expect(count).toBe(1);
      });
    });
  });

  describe('Connection Management', () => {
    it('should connect to device when connectToDevice is called', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      await act(async () => {
        await result.current.connectToDevice('device-123');
      });

      expect(mockBluetoothManager.connectToDevice).toHaveBeenCalledWith('device-123');
    });

    it('should disconnect from device when disconnect is called', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      await act(async () => {
        invokeCallback('onDeviceDiscovered', mockDevice);
        invokeCallback('onConnectionStateChanged', 'connected', 'device-123');
      });

      await act(async () => {
        await result.current.disconnect();
      });

      expect(mockBluetoothManager.disconnectDevice).toHaveBeenCalledWith(
        mockDevice.id
      );
    });

    it('should update connectedDevice when connection succeeds', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      await act(async () => {
        await result.current.startDiscovery();
        invokeCallback('onDeviceDiscovered', mockDevice);
        invokeCallback('onConnectionStateChanged', 'connected', mockDevice.id);
      });

      await waitFor(() => {
        expect(result.current.connectedDevice).toBeTruthy();
        expect(result.current.connectedDevice?.id).toBe(mockDevice.id);
      });
    });

    it('should clear connectedDevice on disconnection', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      await act(async () => {
        invokeCallback('onDeviceDiscovered', mockDevice);
        invokeCallback('onConnectionStateChanged', 'connected', mockDevice.id);
      });

      await waitFor(() => {
        expect(result.current.connectedDevice).toBeTruthy();
      });

      await act(async () => {
        invokeCallback('onConnectionStateChanged', 'disconnected', mockDevice.id);
      });

      await waitFor(() => {
        expect(result.current.connectedDevice).toBeNull();
      });
    });
  });

  describe('Auto-Reconnection', () => {
    it('should enable auto-reconnect when autoReconnect prop is true', () => {
      const { result } = renderHook(() =>
        useBluetoothScanner({
          autoReconnect: true,
        })
      );

      expect(mockBluetoothManager.enableAutoReconnect).toHaveBeenCalledWith(true);
    });

    it('should disable auto-reconnect when autoReconnect prop is false', () => {
      const { result } = renderHook(() =>
        useBluetoothScanner({
          autoReconnect: false,
        })
      );

      expect(mockBluetoothManager.enableAutoReconnect).toHaveBeenCalledWith(false);
    });

    it('should enable auto-reconnect by default', () => {
      const { result } = renderHook(() => useBluetoothScanner());

      // Default should be true based on implementation
      expect(mockBluetoothManager.enableAutoReconnect).toHaveBeenCalledWith(true);
    });
  });

  describe('Permission Handling', () => {
    it('should request permissions when requestPermissions is called', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      await act(async () => {
        await result.current.requestPermissions();
      });

      expect(mockBluetoothManager.requestPermissions).toHaveBeenCalled();
    });

    it('should return permission result', async () => {
      mockBluetoothManager.requestPermissions = jest.fn().mockResolvedValue(false);

      const { result } = renderHook(() => useBluetoothScanner());

      let granted;
      await act(async () => {
        granted = await result.current.requestPermissions();
      });

      expect(granted).toBe(false);
    });

    it('should return true when permissions granted', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      let granted;
      await act(async () => {
        granted = await result.current.requestPermissions();
      });

      expect(granted).toBe(true);
    });
  });

  describe('Data Reception', () => {
    it('should call onDataReceived callback when barcode data received', async () => {
      const onDataReceived = jest.fn();
      const { result } = renderHook(() =>
        useBluetoothScanner({
          onDataReceived,
        })
      );

      await act(async () => {
        invokeCallback('onDataReceived', '8991234567');
      });

      await waitFor(() => {
        expect(onDataReceived).toHaveBeenCalledWith('8991234567');
      });
    });
  });

  describe('Error Handling', () => {
    it('should call onError callback when error occurs', async () => {
      const onError = jest.fn();
      const { result } = renderHook(() =>
        useBluetoothScanner({
          onError,
        })
      );

      await act(async () => {
        invokeCallback('onError', {
          type: 'connection_failed',
          message: 'Connection failed',
        });
      });

      await waitFor(() => {
        expect(onError).toHaveBeenCalledWith({
          type: 'connection_failed',
          message: 'Connection failed',
        });
      });
    });

    it('should handle connection errors gracefully', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      // Mock connection error
      mockBluetoothManager.connectToDevice = jest.fn().mockRejectedValue(
        new Error('Connection failed')
      );

      // Should not throw
      await act(async () => {
        await expect(
          result.current.connectToDevice('device-123')
        ).rejects.toThrow();
      });
    });

    it('should update error state on connection failure', async () => {
      const onError = jest.fn();
      const { result } = renderHook(() =>
        useBluetoothScanner({
          onError,
        })
      );

      await act(async () => {
        invokeCallback('onConnectionStateChanged', 'error', 'device-123');
        invokeCallback('onError', {
          type: 'connection_failed',
          message: 'Test error',
        });
      });

      await waitFor(() => {
        expect(result.current.connectionState).toBe('error');
        expect(onError).toHaveBeenCalled();
      });
    });
  });

  describe('Connection Monitoring', () => {
    it('should start connection monitoring when hook mounts', () => {
      renderHook(() => useBluetoothScanner());

      // Should start monitoring automatically
      expect(mockBluetoothManager.startConnectionMonitoring).toHaveBeenCalled();
    });

    it('should stop monitoring when hook unmounts', () => {
      const { unmount } = renderHook(() => useBluetoothScanner());

      unmount();

      expect(mockBluetoothManager.stopConnectionMonitoring).toHaveBeenCalled();
    });
  });

  describe('Memory Management', () => {
    it('should clear discovered devices when starting new discovery', async () => {
      const { result } = renderHook(() => useBluetoothScanner());

      // Start first discovery
      await act(async () => {
        await result.current.startDiscovery();
        invokeCallback('onDeviceDiscovered', mockDevice);
      });

      await waitFor(() => {
        expect(result.current.discoveredDevices.length).toBeGreaterThan(0);
      });

      // Stop and start new discovery
      await act(async () => {
        await result.current.stopDiscovery();
        await result.current.startDiscovery();
      });

      // Should start with empty list
      expect(result.current.discoveredDevices).toEqual([]);
    });
  });
});
