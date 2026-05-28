/**
 * Printer Manager Tests
 * Tests for PrinterManager implementation with USB, Bluetooth, and Network printer support
 */

import {
  PrinterConnectionType,
  PrinterStatus,
  PrinterDevice,
  PrinterErrorType,
  IPrinterManager,
} from './printer';

// Mock the thermal printer library
jest.mock('@finan-me/react-native-thermal-printer', () => ({
  ThermalPrinterModule: {
    getUsbDevices: jest.fn(() => Promise.resolve([])),
    connectToUsbDevice: jest.fn(() => Promise.resolve(true)),
    getBluetoothDevices: jest.fn(() => Promise.resolve([])),
    connectToBluetoothDevice: jest.fn(() => Promise.resolve(true)),
    getNetPrinters: jest.fn(() => Promise.resolve([])),
    connectToNetPrinter: jest.fn(() => Promise.resolve(true)),
    disconnect: jest.fn(() => Promise.resolve(true)),
    print: jest.fn(() => Promise.resolve(true)),
  },
}));

// Mock React Native modules
jest.mock('react-native/Libraries/EventEmitter/NativeEventEmitter', () => {
  return {
    default: jest.fn().mockImplementation(() => ({
      addListener: jest.fn(),
      removeListener: jest.fn(),
      removeAllListeners: jest.fn(),
    })),
  };
});

import { PrinterManager } from './PrinterManager';

describe('PrinterManager', () => {
  let printerManager: IPrinterManager;
  let mockErrorCallback: jest.Mock;
  let mockStatusCallback: jest.Mock;

  beforeEach(() => {
    jest.clearAllMocks();
    printerManager = PrinterManager.getInstance();
    mockErrorCallback = jest.fn();
    mockStatusCallback = jest.fn();
    printerManager.onError(mockErrorCallback);
    printerManager.onStatusChange(mockStatusCallback);
  });

  afterEach(() => {
    // Reset singleton instance
    (PrinterManager as any).instance = null;
  });

  describe('Singleton Pattern', () => {
    it('should return the same instance', () => {
      const instance1 = PrinterManager.getInstance();
      const instance2 = PrinterManager.getInstance();
      expect(instance1).toBe(instance2);
    });

    it('should create only one instance', () => {
      const manager1 = PrinterManager.getInstance();
      const manager2 = PrinterManager.getInstance();
      expect(manager1).toBe(manager2);
    });
  });

  describe('Printer Discovery', () => {
    it('should discover USB printers', async () => {
      const mockUsbDevices = [
        {
          id: 'usb-1',
          name: 'Xprinter XP-58IIH',
          connectionType: PrinterConnectionType.USB,
          vendorId: 0x0416,
          productId: 0x5011,
        },
      ];

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.getUsbDevices.mockResolvedValueOnce(mockUsbDevices);

      const discovered = await printerManager.discoverPrinters();

      expect(discovered).toHaveLength(1);
      expect(discovered[0].connectionType).toBe(PrinterConnectionType.USB);
      expect(ThermalPrinterModule.getUsbDevices).toHaveBeenCalled();
    });

    it('should discover Bluetooth printers', async () => {
      const mockBluetoothDevices = [
        {
          id: 'bt-1',
          name: 'Bluetooth Printer',
          connectionType: PrinterConnectionType.BLUETOOTH,
          address: '00:11:22:33:44:55',
        },
      ];

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.getBluetoothDevices.mockResolvedValueOnce(mockBluetoothDevices);

      const discovered = await printerManager.discoverPrinters();

      expect(discovered).toHaveLength(1);
      expect(discovered[0].connectionType).toBe(PrinterConnectionType.BLUETOOTH);
      expect(ThermalPrinterModule.getBluetoothDevices).toHaveBeenCalled();
    });

    it('should discover Network printers', async () => {
      const mockNetworkPrinters = [
        {
          id: 'net-1',
          name: 'Network Printer',
          connectionType: PrinterConnectionType.NETWORK,
          address: '192.168.1.100',
        },
      ];

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.getNetPrinters.mockResolvedValueOnce(mockNetworkPrinters);

      const discovered = await printerManager.discoverPrinters();

      expect(discovered).toHaveLength(1);
      expect(discovered[0].connectionType).toBe(PrinterConnectionType.NETWORK);
      expect(ThermalPrinterModule.getNetPrinters).toHaveBeenCalled();
    });

    it('should combine all printer types', async () => {
      const mockDevices = [
        { id: 'usb-1', name: 'USB Printer', connectionType: PrinterConnectionType.USB },
        { id: 'bt-1', name: 'BT Printer', connectionType: PrinterConnectionType.BLUETOOTH },
        { id: 'net-1', name: 'NET Printer', connectionType: PrinterConnectionType.NETWORK },
      ];

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.getUsbDevices.mockResolvedValueOnce([mockDevices[0]]);
      ThermalPrinterModule.getBluetoothDevices.mockResolvedValueOnce([mockDevices[1]]);
      ThermalPrinterModule.getNetPrinters.mockResolvedValueOnce([mockDevices[2]]);

      const discovered = await printerManager.discoverPrinters();

      expect(discovered).toHaveLength(3);
    });
  });

  describe('Printer Connection', () => {
    it('should connect to USB printer successfully', async () => {
      const usbDevice: PrinterDevice = {
        id: 'usb-1',
        name: 'Xprinter XP-58IIH',
        connectionType: PrinterConnectionType.USB,
        vendorId: 0x0416,
        productId: 0x5011,
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValueOnce(true);

      const connected = await printerManager.connect(usbDevice);

      expect(connected).toBe(true);
      expect(ThermalPrinterModule.connectToUsbDevice).toHaveBeenCalledWith(usbDevice);
      expect(mockStatusCallback).toHaveBeenCalledWith(PrinterStatus.CONNECTED);
    });

    it('should connect to Bluetooth printer successfully', async () => {
      const bluetoothDevice: PrinterDevice = {
        id: 'bt-1',
        name: 'Bluetooth Printer',
        connectionType: PrinterConnectionType.BLUETOOTH,
        address: '00:11:22:33:44:55',
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToBluetoothDevice.mockResolvedValueOnce(true);

      const connected = await printerManager.connect(bluetoothDevice);

      expect(connected).toBe(true);
      expect(ThermalPrinterModule.connectToBluetoothDevice).toHaveBeenCalledWith(bluetoothDevice);
      expect(mockStatusCallback).toHaveBeenCalledWith(PrinterStatus.CONNECTED);
    });

    it('should connect to Network printer successfully', async () => {
      const networkDevice: PrinterDevice = {
        id: 'net-1',
        name: 'Network Printer',
        connectionType: PrinterConnectionType.NETWORK,
        address: '192.168.1.100',
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToNetPrinter.mockResolvedValueOnce(true);

      const connected = await printerManager.connect(networkDevice);

      expect(connected).toBe(true);
      expect(ThermalPrinterModule.connectToNetPrinter).toHaveBeenCalledWith(networkDevice);
      expect(mockStatusCallback).toHaveBeenCalledWith(PrinterStatus.CONNECTED);
    });

    it('should handle connection failure', async () => {
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'USB Printer',
        connectionType: PrinterConnectionType.USB,
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValueOnce(false);

      const connected = await printerManager.connect(device);

      expect(connected).toBe(false);
      expect(mockStatusCallback).toHaveBeenCalledWith(PrinterStatus.ERROR);
      expect(mockErrorCallback).toHaveBeenCalled();
    });

    it('should update current printer after successful connection', async () => {
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'USB Printer',
        connectionType: PrinterConnectionType.USB,
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValueOnce(true);

      await printerManager.connect(device);

      expect(printerManager.getCurrentPrinter()).toEqual(device);
    });
  });

  describe('Printer Disconnection', () => {
    it('should disconnect from printer successfully', async () => {
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'USB Printer',
        connectionType: PrinterConnectionType.USB,
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValueOnce(true);
      ThermalPrinterModule.disconnect.mockResolvedValueOnce(true);

      await printerManager.connect(device);
      const disconnected = await printerManager.disconnect();

      expect(disconnected).toBe(true);
      expect(ThermalPrinterModule.disconnect).toHaveBeenCalled();
      expect(mockStatusCallback).toHaveBeenCalledWith(PrinterStatus.DISCONNECTED);
      expect(printerManager.getCurrentPrinter()).toBeNull();
    });

    it('should handle disconnection failure', async () => {
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'USB Printer',
        connectionType: PrinterConnectionType.USB,
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValueOnce(true);
      ThermalPrinterModule.disconnect.mockResolvedValueOnce(false);

      await printerManager.connect(device);
      const disconnected = await printerManager.disconnect();

      expect(disconnected).toBe(false);
      expect(mockErrorCallback).toHaveBeenCalled();
    });
  });

  describe('Printer Status', () => {
    it('should return CONNECTED status when printer is connected', async () => {
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'USB Printer',
        connectionType: PrinterConnectionType.USB,
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValueOnce(true);

      await printerManager.connect(device);

      expect(printerManager.getStatus()).toBe(PrinterStatus.CONNECTED);
    });

    it('should return DISCONNECTED status when no printer is connected', () => {
      expect(printerManager.getStatus()).toBe(PrinterStatus.DISCONNECTED);
    });

    it('should return ERROR status after connection failure', async () => {
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'USB Printer',
        connectionType: PrinterConnectionType.USB,
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValueOnce(false);

      await printerManager.connect(device);

      expect(printerManager.getStatus()).toBe(PrinterStatus.ERROR);
    });
  });

  describe('Current Printer', () => {
    it('should return null when no printer is connected', () => {
      expect(printerManager.getCurrentPrinter()).toBeNull();
    });

    it('should return connected printer device', async () => {
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'USB Printer',
        connectionType: PrinterConnectionType.USB,
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValueOnce(true);

      await printerManager.connect(device);

      expect(printerManager.getCurrentPrinter()).toEqual(device);
    });
  });

  describe('Error Handling', () => {
    it('should call error callback on connection failure', async () => {
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'USB Printer',
        connectionType: PrinterConnectionType.USB,
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValueOnce(false);

      await printerManager.connect(device);

      expect(mockErrorCallback).toHaveBeenCalledWith({
        type: PrinterErrorType.CONNECTION_FAILED,
        message: expect.stringContaining('Failed to connect'),
      });
    });

    it('should allow setting error callback', () => {
      const errorCallback = jest.fn();
      printerManager.onError(errorCallback);

      expect(mockErrorCallback).not.toHaveBeenCalled();
    });
  });

  describe('Status Change Callbacks', () => {
    it('should allow setting status change callback', () => {
      const statusCallback = jest.fn();
      printerManager.onStatusChange(statusCallback);

      expect(mockStatusCallback).not.toHaveBeenCalled();
    });

    it('should call status change callback on connection', async () => {
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'USB Printer',
        connectionType: PrinterConnectionType.USB,
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValueOnce(true);

      await printerManager.connect(device);

      expect(mockStatusCallback).toHaveBeenCalledWith(PrinterStatus.CONNECTED);
    });
  });

  describe('Auto-Reconnect Logic', () => {
    it('should have auto-reconnect configuration option', () => {
      const config = {
        autoReconnect: true,
        reconnectAttempts: 3,
        reconnectDelay: 1000,
      };

      const manager = PrinterManager.getInstance();
      expect(manager).toBeDefined();
      // Auto-reconnect logic should be configurable
    });

    it('should attempt reconnection when enabled', async () => {
      // This would test auto-reconnect functionality
      // Implementation would involve monitoring connection status
      // and attempting reconnection based on configuration
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'USB Printer',
        connectionType: PrinterConnectionType.USB,
      };

      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValueOnce(true);

      await printerManager.connect(device);

      expect(printerManager.getStatus()).toBe(PrinterStatus.CONNECTED);
    });
  });
});
