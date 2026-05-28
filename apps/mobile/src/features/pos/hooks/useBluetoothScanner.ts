/**
 * useBluetoothScanner Hook
 *
 * React hook for managing Bluetooth barcode scanner connections.
 * Provides state and methods for device discovery, connection management,
 * and data reception from Bluetooth scanners.
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import { BluetoothManager, BluetoothManagerCallbacks } from '../hardware/BluetoothManager';
import type {
  BluetoothDevice,
  BluetoothConnectionState,
  BluetoothError
} from '../types/scanner.types';

export interface UseBluetoothScannerOptions {
  /** Enable auto-reconnection when connection is lost (default: true) */
  autoReconnect?: boolean;
  /** Callback when barcode data is received */
  onDataReceived?: (data: string) => void;
  /** Callback when error occurs */
  onError?: (error: BluetoothError) => void;
}

export interface UseBluetoothScannerReturn {
  /** Current connection state */
  connectionState: BluetoothConnectionState;
  /** List of discovered devices from current scan */
  discoveredDevices: BluetoothDevice[];
  /** Currently connected device (if any) */
  connectedDevice: BluetoothDevice | null;
  /** Start scanning for Bluetooth devices */
  startDiscovery: () => Promise<void>;
  /** Stop scanning for devices */
  stopDiscovery: () => Promise<void>;
  /** Connect to a specific device by ID */
  connectToDevice: (deviceId: string) => Promise<void>;
  /** Disconnect from current device */
  disconnect: () => Promise<void>;
  /** Request Bluetooth permissions */
  requestPermissions: () => Promise<boolean>;
  /** Whether Bluetooth is currently scanning */
  isScanning: boolean;
}

export function useBluetoothScanner(options: UseBluetoothScannerOptions = {}): UseBluetoothScannerReturn {
  const {
    autoReconnect = true,
    onDataReceived,
    onError,
  } = options;

  // State management
  const [connectionState, setConnectionState] = useState<BluetoothConnectionState>('disconnected');
  const [discoveredDevices, setDiscoveredDevices] = useState<BluetoothDevice[]>([]);
  const [connectedDevice, setConnectedDevice] = useState<BluetoothDevice | null>(null);
  const [isScanning, setIsScanning] = useState(false);

  // Ref to store BluetoothManager instance
  const managerRef = useRef<BluetoothManager | null>(null);

  // Ref to track discovered devices synchronously (for callback handling)
  const devicesRef = useRef<Map<string, BluetoothDevice>>(new Map());

  // Initialize BluetoothManager
  useEffect(() => {
    const callbacks: BluetoothManagerCallbacks = {
      onDeviceDiscovered: (device) => {
        // Update ref synchronously
        devicesRef.current.set(device.id, device);

        // Update state
        setDiscoveredDevices((prev) => {
          // Avoid duplicates
          if (prev.some((d) => d.id === device.id)) {
            return prev;
          }
          return [...prev, device];
        });
      },
      onConnectionStateChanged: (state, deviceId) => {
        setConnectionState(state);

        // Update connected device based on state
        if (state === 'connected') {
          // Find device from ref (synchronous)
          const device = devicesRef.current.get(deviceId);
          if (device) {
            const connectedDevice = { ...device, connected: true, paired: true };
            setConnectedDevice(connectedDevice);

            // Also update in discovered devices
            setDiscoveredDevices((prev) =>
              prev.map((d) => (d.id === deviceId ? connectedDevice : d))
            );

            // Update ref
            devicesRef.current.set(deviceId, connectedDevice);
          }
        } else if (state === 'disconnected' || state === 'error') {
          setConnectedDevice(null);
        }
      },
      onDataReceived: (data) => {
        onDataReceived?.(data);
      },
      onError: (error) => {
        onError?.(error);
      },
    };

    const manager = new BluetoothManager(callbacks);
    managerRef.current = manager;

    // Enable auto-reconnect if specified (only if true, otherwise disable)
    manager.enableAutoReconnect(autoReconnect);

    // Start connection monitoring
    manager.startConnectionMonitoring();

    // Cleanup on unmount
    return () => {
      manager.stopConnectionMonitoring();
      manager.destroy();
    };
  }, [autoReconnect]);

  /**
   * Start device discovery
   */
  const startDiscovery = useCallback(async () => {
    if (!managerRef.current) return;

    setIsScanning(true);
    setDiscoveredDevices([]);
    devicesRef.current.clear(); // Clear ref as well

    await managerRef.current.startDiscovery();
  }, []);

  /**
   * Stop device discovery
   */
  const stopDiscovery = useCallback(async () => {
    if (!managerRef.current) return;

    await managerRef.current.stopDiscovery();
    setIsScanning(false);
  }, []);

  /**
   * Connect to a device
   */
  const connectToDevice = useCallback(async (deviceId: string) => {
    if (!managerRef.current) return;

    await managerRef.current.connectToDevice(deviceId);
  }, []);

  /**
   * Disconnect from current device
   */
  const disconnect = useCallback(async () => {
    if (!managerRef.current || !connectedDevice) return;

    await managerRef.current.disconnectDevice(connectedDevice.id);
  }, [connectedDevice]);

  /**
   * Request Bluetooth permissions
   */
  const requestPermissions = useCallback(async () => {
    if (!managerRef.current) return false;

    return await managerRef.current.requestPermissions();
  }, []);

  return {
    connectionState,
    discoveredDevices,
    connectedDevice,
    startDiscovery,
    stopDiscovery,
    connectToDevice,
    disconnect,
    requestPermissions,
    isScanning,
  };
}
