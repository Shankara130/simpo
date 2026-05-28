/**
 * Mock for react-native-ble-plx
 * Used for testing BluetoothManager without actual hardware
 */

// Mock device list for testing
let mockConnectedDevices: any[] = [];
let mockIsScanning: boolean = false;
let mockDeviceScanCallback: ((error: any, device: any) => void) | null = null;

export class BleClient {
  static async create() {
    return new BleClient();
  }

  async startDeviceScan(
    serviceUUIDs: string[],
    options: any,
    callback: (error: any, device: any) => void
  ): Promise<void> {
    mockIsScanning = true;
    mockDeviceScanCallback = callback;

    // Emit mock device after delay for testing
    setTimeout(() => {
      if (mockDeviceScanCallback) {
        mockDeviceScanCallback(null, {
          id: '00:11:22:33:44:55',
          name: 'Zebra Scanner',
          localName: 'Zebra Scanner',
          rssi: -60,
        });
      }
    }, 100);

    return Promise.resolve();
  }

  async stopDeviceScan(): Promise<void> {
    mockIsScanning = false;
    mockDeviceScanCallback = null;
    return Promise.resolve();
  }

  async connectToDevice(
    deviceId: string,
    connectionOptions?: any
  ): Promise<void> {
    // Simulate connection success/failure based on deviceId
    if (deviceId === 'unknown-device') {
      throw new Error('Device not found');
    }

    // Simulate connection process
    return new Promise((resolve) => {
      setTimeout(() => {
        // Add to connected devices
        mockConnectedDevices.push({
          id: deviceId,
          name: 'Test Device',
          rssi: -60,
        });
        resolve();
      }, 50);
    });
  }

  async disconnectFromDevice(deviceId: string): Promise<void> {
    mockConnectedDevices = mockConnectedDevices.filter(d => d.id !== deviceId);
    return Promise.resolve();
  }

  async cancelDeviceConnection(deviceId: string): Promise<void> {
    mockConnectedDevices = mockConnectedDevices.filter(d => d.id !== deviceId);
    return Promise.resolve();
  }

  async connectedDevices(serviceUUIDs: string[]): Promise<any[]> {
    return Promise.resolve(mockConnectedDevices);
  }

  async requestConnectionPriority(
    deviceId: string,
    priority: any
  ): Promise<void> {
    return Promise.resolve();
  }

  async isConnected(deviceId: string): Promise<boolean> {
    return Promise.resolve(mockConnectedDevices.some(d => d.id === deviceId));
  }

  async read(
    deviceId: string,
    serviceUUID: string,
    characteristicUUID: string
  ): Promise<void> {
    return Promise.resolve();
  }

  async write(
    deviceId: string,
    serviceUUID: string,
    characteristicUUID: string,
    data: any,
    writeType?: any
  ): Promise<void> {
    return Promise.resolve();
  }

  async startNotifications(
    deviceId: string,
    serviceUUID: string,
    characteristicUUID: string,
    listener: (error: any, notification: any) => void
  ): Promise<void> {
    // Simulate data reception after delay
    setTimeout(() => {
      if (listener) {
        listener(null, {
          value: { 0: 0x38, 1: 0x39, 2: 0x39 }, // Test data: "899"
        });
      }
    }, 50);
    return Promise.resolve();
  }

  async stopNotifications(
    deviceId: string,
    serviceUUID: string,
    characteristicUUID: string
  ): Promise<void> {
    return Promise.resolve();
  }
}

// Helper to reset mock state between tests
export function resetMock() {
  mockConnectedDevices = [];
  mockIsScanning = false;
  mockDeviceScanCallback = null;
}

export interface Device {
  id: string;
  name?: string;
  rssi?: number;
  localName?: string;
  serviceUUIDs?: string[];
}

export type ConnectionState = 'connecting' | 'connected' | 'disconnecting' | 'disconnected';
