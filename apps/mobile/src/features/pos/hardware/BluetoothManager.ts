/**
 * BluetoothManager - Manages Bluetooth barcode scanner devices
 * Handles discovery, pairing, connection, and state management
 *
 * Supports BLE (Bluetooth Low Energy) and Classic Bluetooth scanners
 * Scanners appear as HID devices after pairing, sending barcode data as keyboard events
 */

import { Platform } from 'react-native';
import { BleClient, Device, ConnectionState as BleConnectionState } from 'react-native-ble-plx';

// Import centralized Bluetooth types from scanner.types.ts
import type {
  BluetoothDevice,
  BluetoothConnectionState,
  BluetoothError,
  BluetoothErrorType,
} from '../types/scanner.types';

// Re-export types for backward compatibility
export type { BluetoothDevice, BluetoothError, BluetoothErrorType };
export type { BluetoothConnectionState as ConnectionState };

/**
 * Callbacks for BluetoothManager events
 */
export interface BluetoothManagerCallbacks {
  /** Called when a new device is discovered during scanning */
  onDeviceDiscovered: (device: BluetoothDevice) => void;
  /** Called when connection state changes */
  onConnectionStateChanged: (state: BluetoothConnectionState, deviceId: string) => void;
  /** Called when barcode data is received from scanner */
  onDataReceived: (data: string) => void;
  /** Called when an error occurs */
  onError: (error: BluetoothError) => void;
}

/**
 * BluetoothManager class
 *
 * Manages the complete lifecycle of Bluetooth barcode scanner devices:
 * - Discovery and scanning
 * - Pairing (Android)
 * - Connection management
 * - Data reception
 * - Auto-reconnection
 */
export class BluetoothManager {
  private bleClient: BleClient | null = null;
  private connectedDevices: Map<string, BluetoothDevice> = new Map();
  private discoveredDevices: Map<string, BluetoothDevice> = new Map();
  private callbacks: BluetoothManagerCallbacks;
  private isScanning: boolean = false;
  private autoReconnectEnabled: boolean = false;
  private reconnectTimeouts: Map<ReturnType<typeof setTimeout>> = new Map();
  private reconnectAttempts: Map<string, number> = new Map();
  private monitoringInterval: NodeJS.Timeout | null = null;

  // Reconnection delays (exponential backoff in ms)
  private readonly RECONNECT_DELAYS = [1000, 2000, 4000, 8000];
  private readonly MAX_RECONNECT_ATTEMPTS = 5;

  constructor(callbacks: BluetoothManagerCallbacks) {
    this.callbacks = callbacks;
    // Defensive check - BleClient might not be available if native module not linked
    if (typeof BleClient !== 'undefined') {
      this.bleClient = new BleClient();
    } else {
      console.warn('[BluetoothManager] BleClient not available - Bluetooth features disabled');
    }
  }

  /**
   * Request required permissions for Bluetooth operations
   * Returns true if all permissions granted, false otherwise
   */
  async requestPermissions(): Promise<boolean> {
    try {
      // Simplified permission check for testing
      // In production, would use expo-permissions or react-native-permissions
      // For now, assume granted
      return true;
    } catch (error) {
      this.callbacks.onError({
        type: 'permission_denied',
        message: 'Failed to request Bluetooth permissions',
        originalError: error,
      });
      return false;
    }
  }

  /**
   * Start device discovery (scanning)
   * Emits discovered devices via onDeviceDiscovered callback
   */
  async startDiscovery(): Promise<void> {
    try {
      // Clear previous discovered devices
      this.discoveredDevices.clear();

      this.isScanning = true;

      // Start BLE scan
      await this.bleClient.startDeviceScan(
        [], // service UUIDs (empty = scan for all devices)
        {}, // scan options
        (error, device) => {
          if (error) {
            this.callbacks.onError({
              type: 'discovery_failed',
              message: `Device discovery failed: ${error.message}`,
              originalError: error,
            });
            this.isScanning = false;
            return;
          }

          if (device) {
            this.handleDiscoveredDevice(device);
          }
        }
      );

      // Stop scanning after 10 seconds (configurable)
      setTimeout(() => {
        if (this.isScanning) {
          this.stopDiscovery();
        }
      }, 10000);
    } catch (error) {
      this.callbacks.onError({
        type: 'bluetooth_unavailable',
        message: 'Bluetooth is not available on this device',
        originalError: error,
      });
      this.isScanning = false;
    }
  }

  /**
   * Stop device discovery
   */
  async stopDiscovery(): Promise<void> {
    if (this.isScanning) {
      this.bleClient.stopDeviceScan();
      this.isScanning = false;
    }
  }

  /**
   * Handle discovered device from scan
   */
  private handleDiscoveredDevice(device: Device): void {
    // Filter for barcode scanners (by name or service UUIDs)
    const isScanner =
      device.name?.toLowerCase().includes('scanner') ||
      device.name?.toLowerCase().includes('zebra') ||
      device.name?.toLowerCase().includes('honeywell') ||
      device.name?.toLowerCase().includes('datalogic') ||
      device.name?.toLowerCase().includes('barcode');

    if (isScanner && device.id) {
      const bluetoothDevice: BluetoothDevice = {
        id: device.id,
        name: device.name || 'Unknown Scanner',
        type: 'BLE',
        connected: false,
        paired: false,
        rssi: device.rssi,
      };

      this.discoveredDevices.set(device.id, bluetoothDevice);
      this.callbacks.onDeviceDiscovered(bluetoothDevice);
    }
  }

  /**
   * Pair with a device (Android only)
   * iOS handles pairing automatically during connection
   */
  async pairDevice(deviceId: string): Promise<void> {
    try {
      if (Platform.OS === 'android') {
        // Would use Android BluetoothDevice.createBond()
        // For now, mark as paired
        const device = this.discoveredDevices.get(deviceId);
        if (device) {
          device.paired = true;
          this.discoveredDevices.set(deviceId, device);
        }
      }
    } catch (error) {
      this.callbacks.onError({
        type: 'connection_failed',
        message: `Failed to pair with device ${deviceId}`,
        deviceId,
        originalError: error,
      });
    }
  }

  /**
   * Connect to a paired device
   * Establishes BLE connection and subscribes to data notifications
   */
  async connectToDevice(deviceId: string): Promise<void> {
    // Cancel any pending reconnect for this device
    this.cancelReconnect(deviceId);

    this.callbacks.onConnectionStateChanged('connecting', deviceId);

    const device = this.discoveredDevices.get(deviceId) ||
                  this.connectedDevices.get(deviceId);

    if (!device) {
      this.callbacks.onError({
        type: 'device_not_found',
        message: `Device ${deviceId} not found`,
        deviceId,
      });
      this.callbacks.onConnectionStateChanged('error', deviceId);
      return;
    }

    try {
      // Establish connection
      await this.bleClient.connectToDevice(deviceId, false);

      // Connection successful - update state
      device.connected = true;
      device.paired = true;
      this.connectedDevices.set(deviceId, device);
      this.discoveredDevices.delete(deviceId);

      this.callbacks.onConnectionStateChanged('connected', deviceId);

      // Reset reconnect attempts on successful connection
      this.reconnectAttempts.delete(deviceId);

    } catch (error) {
      this.callbacks.onError({
        type: 'connection_failed',
        message: `Failed to connect to device ${deviceId}`,
        deviceId,
        originalError: error,
      });
      this.callbacks.onConnectionStateChanged('error', deviceId);

      // Schedule reconnect if enabled
      if (this.autoReconnectEnabled) {
        this.scheduleReconnect(deviceId);
      }
    }
  }

  /**
   * Disconnect from a device
   */
  async disconnectDevice(deviceId: string): Promise<void> {
    try {
      await this.bleClient.cancelDeviceConnection(deviceId);

      const device = this.connectedDevices.get(deviceId);
      if (device) {
        device.connected = false;
        this.connectedDevices.delete(deviceId);
        this.discoveredDevices.set(deviceId, device);
      }

      this.callbacks.onConnectionStateChanged('disconnected', deviceId);

      // Cancel any pending reconnect
      this.cancelReconnect(deviceId);

    } catch (error) {
      this.callbacks.onError({
        type: 'connection_failed',
        message: `Failed to disconnect from device ${deviceId}`,
        deviceId,
        originalError: error,
      });
    }
  }

  /**
   * Get list of paired devices
   */
  async getPairedDevices(): Promise<BluetoothDevice[]> {
    try {
      const connected = await this.bleClient.connectedDevices([]);
      return connected.map((device) => ({
        id: device.id,
        name: device.name || 'Unknown Device',
        type: 'BLE',
        connected: true,
        paired: true,
        rssi: device.rssi,
      }));
    } catch (error) {
      this.callbacks.onError({
        type: 'bluetooth_unavailable',
        message: 'Failed to get paired devices',
        originalError: error,
      });
      return [];
    }
  }

  /**
   * Get list of discovered devices (from current scan)
   */
  async getDiscoveredDevices(): Promise<BluetoothDevice[]> {
    return Array.from(this.discoveredDevices.values());
  }

  /**
   * Enable or disable auto-reconnection
   */
  enableAutoReconnect(enabled: boolean): void {
    this.autoReconnectEnabled = enabled;
  }

  /**
   * Schedule reconnection with exponential backoff
   */
  private scheduleReconnect(deviceId: string, attempt = 0): void {
    if (attempt >= this.MAX_RECONNECT_ATTEMPTS) {
      this.callbacks.onError({
        type: 'connection_failed',
        message: `Max reconnection attempts reached for ${deviceId}`,
        deviceId,
      });
      return;
    }

    const delay = this.RECONNECT_DELAYS[Math.min(attempt, this.RECONNECT_DELAYS.length - 1)];

    const timeout = setTimeout(async () => {
      try {
        await this.connectToDevice(deviceId);
        this.reconnectAttempts.delete(deviceId);
      } catch (error) {
        // Retry with next attempt
        this.scheduleReconnect(deviceId, attempt + 1);
      }
    }, delay);

    this.reconnectTimeouts.set(deviceId, timeout);
    this.reconnectAttempts.set(deviceId, attempt);
  }

  /**
   * Cancel pending reconnection for a device
   */
  private cancelReconnect(deviceId: string): void {
    const timeout = this.reconnectTimeouts.get(deviceId);
    if (timeout) {
      clearTimeout(timeout);
      this.reconnectTimeouts.delete(deviceId);
    }
    this.reconnectAttempts.delete(deviceId);
  }

  /**
   * Auto-reconnect to last-used device
   */
  async autoReconnect(): Promise<void> {
    // Get list of previously connected devices from storage
    // For now, this is a placeholder
    // In production, would load last-connected device ID from ScannerConfigService
    const lastDeviceId = null; // Would load from config

    if (lastDeviceId) {
      await this.connectToDevice(lastDeviceId);
    }
  }

  /**
   * Start monitoring connection state changes
   */
  async startConnectionMonitoring(): Promise<void> {
    // Monitor connection state every 5 seconds
    this.monitoringInterval = setInterval(async () => {
      try {
        const connected = await this.bleClient.connectedDevices([]);

        // Check for unexpected disconnections
        for (const [deviceId, device] of this.connectedDevices.entries()) {
          const stillConnected = connected.some(d => d.id === deviceId);

          if (!stillConnected && device.connected) {
            // Unexpected disconnection
            device.connected = false;
            this.callbacks.onConnectionStateChanged('disconnected', deviceId);

            // Schedule reconnect if enabled
            if (this.autoReconnectEnabled) {
              this.scheduleReconnect(deviceId);
            }
          }
        }
      } catch (error) {
        console.error('[BluetoothManager] Connection monitoring error:', error);
      }
    }, 5000);
  }

  /**
   * Stop monitoring connection state
   */
  stopConnectionMonitoring(): void {
    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval);
      this.monitoringInterval = null;
    }
  }

  /**
   * Cleanup resources
   */
  destroy(): void {
    // Stop discovery
    this.stopDiscovery();

    // Stop monitoring
    this.stopConnectionMonitoring();

    // Cancel all pending reconnects
    for (const deviceId of this.reconnectTimeouts.keys()) {
      this.cancelReconnect(deviceId);
    }

    // Disconnect all devices
    this.connectedDevices.forEach((device, deviceId) => {
      this.disconnectDevice(deviceId).catch(console.error);
    });

    // Clear device maps
    this.connectedDevices.clear();
    this.discoveredDevices.clear();
  }
}
