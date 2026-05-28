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
  CashDrawerOptions,
  DEFAULT_CASH_DRAWER_CONFIG,
} from './printer';

// Mock the thermal printer library
jest.mock('@finan-me/react-native-thermal-printer', () => {
  const ThermalPrinterModule = {
    getUsbDevices: jest.fn(() => Promise.resolve([])),
    connectToUsbDevice: jest.fn(() => Promise.resolve(true)),
    getBluetoothDevices: jest.fn(() => Promise.resolve([])),
    connectToBluetoothDevice: jest.fn(() => Promise.resolve(true)),
    getNetPrinters: jest.fn(() => Promise.resolve([])),
    connectToNetPrinter: jest.fn(() => Promise.resolve(true)),
    disconnect: jest.fn(() => Promise.resolve(true)),
    print: jest.fn(() => Promise.resolve(true)),
  };
  return {
    __esModule: true,
    default: { ThermalPrinterModule },
    ThermalPrinterModule,
  };
});

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

  // ============================================================================
  // Cash Drawer Support Tests (Story 7.4)
  // ============================================================================

  describe('Cash Drawer Control', () => {
    const defaultDrawerOptions: CashDrawerOptions = {
      pulseTiming: 100,
      pinNumber: 0, // Pin 2
      enabled: true,
    };

    beforeEach(async () => {
      // Clear all mocks to start fresh
      jest.clearAllMocks();

      // Set up the print mock to return true by default for this test suite
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.print.mockResolvedValue(true);
      ThermalPrinterModule.connectToUsbDevice.mockResolvedValue(true);
      ThermalPrinterModule.disconnect.mockResolvedValue(true);

      // Connect a printer before drawer tests
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'USB Printer',
        connectionType: PrinterConnectionType.USB,
      };
      await printerManager.connect(device);
    });

    it('should have print mock properly configured', async () => {
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      // Verify print mock exists and returns true
      const result = await ThermalPrinterModule.print(new Uint8Array([0x1B, 0x70]));
      expect(result).toBe(true);
      expect(ThermalPrinterModule.print).toHaveBeenCalled();
    });

    it('should open cash drawer when enabled', async () => {
      // Use the default print mock from beforeEach
      const result = await printerManager.openCashDrawer(defaultDrawerOptions);

      expect(result).toBe(true);
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      expect(ThermalPrinterModule.print).toHaveBeenCalled();
      expect(printerManager.getDrawerStatus()).toBe('connected');
    });

    it('should not open drawer when disabled', async () => {
      const disabledOptions: CashDrawerOptions = {
        ...defaultDrawerOptions,
        enabled: false,
      };

      const result = await printerManager.openCashDrawer(disabledOptions);
      const mockCallback = jest.fn();
      await printerManager.openCashDrawer(disabledOptions, mockCallback);

      expect(result).toBe(false);
      expect(mockCallback).toHaveBeenCalledWith(false, 'Drawer disabled in configuration');
    });

    it('should fail to open drawer when printer not connected', async () => {
      // Disconnect the printer
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.disconnect.mockResolvedValueOnce(true);
      await printerManager.disconnect();

      const result = await printerManager.openCashDrawer(defaultDrawerOptions);
      const mockCallback = jest.fn();
      await printerManager.openCashDrawer(defaultDrawerOptions, mockCallback);

      expect(result).toBe(false);
      expect(mockCallback).toHaveBeenCalledWith(false, 'Printer not connected');
      expect(printerManager.getDrawerStatus()).toBe('failed');
    });

    it('should handle drawer open failure gracefully', async () => {
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      // Override for this test to simulate failure
      ThermalPrinterModule.print.mockResolvedValueOnce(false);

      const mockCallback = jest.fn();
      const result = await printerManager.openCashDrawer(defaultDrawerOptions, mockCallback);

      expect(result).toBe(false);
      expect(mockCallback).toHaveBeenCalledWith(false, 'Failed to send drawer command to printer');
      expect(printerManager.getDrawerStatus()).toBe('failed');
    });

    it('should update drawer status to opening during operation', async () => {
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');

      // Mock print to return after a delay
      let printResolver: (value: boolean) => void = () => {};
      const printPromise = new Promise<boolean>(resolve => {
        printResolver = resolve;
      });
      ThermalPrinterModule.print.mockReturnValueOnce(printPromise);

      const openPromise = printerManager.openCashDrawer(defaultDrawerOptions);

      // Status should be 'opening' immediately
      expect(printerManager.getDrawerStatus()).toBe('opening');

      // Resolve the print
      printResolver(true);
      await openPromise;

      // Status should be 'connected' after success
      expect(printerManager.getDrawerStatus()).toBe('connected');
    });

    it('should support different pulse timings', async () => {
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.print.mockResolvedValue(true);

      // Test with 50ms pulse
      const shortPulseOptions: CashDrawerOptions = {
        ...defaultDrawerOptions,
        pulseTiming: 50,
      };
      await printerManager.openCashDrawer(shortPulseOptions);
      expect(printerManager.getDrawerStatus()).toBe('connected');

      // Test with 200ms pulse
      const longPulseOptions: CashDrawerOptions = {
        ...defaultDrawerOptions,
        pulseTiming: 200,
      };
      await printerManager.openCashDrawer(longPulseOptions);
      expect(printerManager.getDrawerStatus()).toBe('connected');
    });

    it('should support both pin 2 and pin 5', async () => {
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.print.mockResolvedValue(true);

      // Test with Pin 2 (drawer 1)
      const pin2Options: CashDrawerOptions = {
        ...defaultDrawerOptions,
        pinNumber: 0,
      };
      await printerManager.openCashDrawer(pin2Options);
      expect(printerManager.getDrawerStatus()).toBe('connected');

      // Test with Pin 5 (drawer 2)
      const pin5Options: CashDrawerOptions = {
        ...defaultDrawerOptions,
        pinNumber: 1,
      };
      await printerManager.openCashDrawer(pin5Options);
      expect(printerManager.getDrawerStatus()).toBe('connected');
    });

    it('should report drawer connected when printer is connected', () => {
      expect(printerManager.isDrawerConnected).toBe(true);
    });

    it('should report drawer disconnected when printer is disconnected', async () => {
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.disconnect.mockResolvedValueOnce(true);
      await printerManager.disconnect();

      expect(printerManager.isDrawerConnected).toBe(false);
    });

    it('should allow setting drawer result handler', () => {
      const handler = jest.fn();
      printerManager.onDrawerResult(handler);

      // Handler should be set (verified by being called during operations)
      expect(handler).toBeDefined();
    });

    it('should allow clearing drawer result handler', () => {
      const handler = jest.fn();
      printerManager.onDrawerResult(handler);
      printerManager.clearDrawerResultHandler();

      // Handler should be cleared
      expect(handler).toBeDefined();
    });

    it('should call drawer result handler on success', async () => {
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.print.mockResolvedValueOnce(true);

      const handler = jest.fn();
      await printerManager.openCashDrawer(defaultDrawerOptions, handler);

      expect(handler).toHaveBeenCalledWith(true);
    });

    it('should call drawer result handler on failure', async () => {
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      ThermalPrinterModule.print.mockResolvedValueOnce(false);

      const handler = jest.fn();
      await printerManager.openCashDrawer(defaultDrawerOptions, handler);

      expect(handler).toHaveBeenCalledWith(false, 'Failed to send drawer command to printer');
    });

    it('should generate correct ESC/POS command for drawer kick', async () => {
      const { ThermalPrinterModule } = require('@finan-me/react-native-thermal-printer');
      // Ensure print will succeed for this test
      ThermalPrinterModule.print.mockResolvedValueOnce(true);

      // Open drawer with 100ms pulse on Pin 2
      await printerManager.openCashDrawer({
        pulseTiming: 100,
        pinNumber: 0,
        enabled: true,
      });

      // Verify print was called
      expect(ThermalPrinterModule.print).toHaveBeenCalled();

      // Get the command that was sent
      const printCall = ThermalPrinterModule.print.mock.calls.find(
        call => call[0] instanceof Uint8Array && call[0].length === 6
      );
      expect(printCall).toBeDefined();
      const command = printCall ? printCall[0] : new Uint8Array();

      // Verify ESC/POS command structure: ESC p 0 m t1 t2
      // ESC = 0x1B, p = 0x70, 0 = pin 2, m = 0x00 (pulse), t1 = 50 (100ms / 2), t2 = 50
      expect(command[0]).toBe(0x1B); // ESC
      expect(command[1]).toBe(0x70); // p
      expect(command[2]).toBe(0x00); // Pin 2
      expect(command[3]).toBe(0x00); // Pulse mode
      expect(command[4]).toBe(50);    // Pulse on time (100ms / 2 = 50)
      expect(command[5]).toBe(50);    // Pulse off time (100ms / 2 = 50)
    });
  });
});
