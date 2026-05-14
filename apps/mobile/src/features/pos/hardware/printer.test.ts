/**
 * Printer Interface and Manager Tests
 * Test printer hardware abstraction and connection management
 */

import {
  PrinterConnectionType,
  PrinterStatus,
  PrinterErrorType,
  PrinterDevice,
  PrinterError,
  IPrinterConnection,
  IPrinterManager,
  PrinterManagerConfig,
} from './printer';
import { PrinterManager, getPrinterManager, resetPrinterManager } from './PrinterManager';

describe('Printer Interface Types', () => {
  describe('PrinterConnectionType', () => {
    it('should have USB connection type', () => {
      expect(PrinterConnectionType.USB).toBe('usb');
    });

    it('should have BLUETOOTH connection type', () => {
      expect(PrinterConnectionType.BLUETOOTH).toBe('bluetooth');
    });

    it('should have NETWORK connection type', () => {
      expect(PrinterConnectionType.NETWORK).toBe('network');
    });
  });

  describe('PrinterStatus', () => {
    it('should have CONNECTED status', () => {
      expect(PrinterStatus.CONNECTED).toBe('connected');
    });

    it('should have DISCONNECTED status', () => {
      expect(PrinterStatus.DISCONNECTED).toBe('disconnected');
    });

    it('should have ERROR status', () => {
      expect(PrinterStatus.ERROR).toBe('error');
    });

    it('should have OUT_OF_PAPER status', () => {
      expect(PrinterStatus.OUT_OF_PAPER).toBe('out_of_paper');
    });

    it('should have CONNECTING status', () => {
      expect(PrinterStatus.CONNECTING).toBe('connecting');
    });
  });

  describe('PrinterErrorType', () => {
    it('should have CONNECTION_FAILED error type', () => {
      expect(PrinterErrorType.CONNECTION_FAILED).toBe('connection_failed');
    });

    it('should have SEND_FAILED error type', () => {
      expect(PrinterErrorType.SEND_FAILED).toBe('send_failed');
    });

    it('should have NOT_CONNECTED error type', () => {
      expect(PrinterErrorType.NOT_CONNECTED).toBe('not_connected');
    });

    it('should have OUT_OF_PAPER error type', () => {
      expect(PrinterErrorType.OUT_OF_PAPER).toBe('out_of_paper');
    });
  });

  describe('PrinterDevice', () => {
    it('should create USB printer device', () => {
      const device: PrinterDevice = {
        id: 'usb-1',
        name: 'PT-210 Thermal Printer',
        connectionType: PrinterConnectionType.USB,
        vendorId: 0x0DD4,
        productId: 0x0141,
      };

      expect(device.id).toBe('usb-1');
      expect(device.name).toBe('PT-210 Thermal Printer');
      expect(device.connectionType).toBe(PrinterConnectionType.USB);
      expect(device.vendorId).toBe(0x0DD4);
      expect(device.productId).toBe(0x0141);
    });

    it('should create Bluetooth printer device', () => {
      const device: PrinterDevice = {
        id: 'bt-1',
        name: 'POS-58BT',
        connectionType: PrinterConnectionType.BLUETOOTH,
        address: '00:11:22:33:44:55',
      };

      expect(device.id).toBe('bt-1');
      expect(device.name).toBe('POS-58BT');
      expect(device.connectionType).toBe(PrinterConnectionType.BLUETOOTH);
      expect(device.address).toBe('00:11:22:33:44:55');
    });

    it('should create Network printer device', () => {
      const device: PrinterDevice = {
        id: 'net-1',
        name: 'Network Printer',
        connectionType: PrinterConnectionType.NETWORK,
        address: '192.168.1.100',
      };

      expect(device.id).toBe('net-1');
      expect(device.name).toBe('Network Printer');
      expect(device.connectionType).toBe(PrinterConnectionType.NETWORK);
      expect(device.address).toBe('192.168.1.100');
    });
  });

  describe('PrinterError', () => {
    it('should create printer error with required fields', () => {
      const error: PrinterError = {
        type: PrinterErrorType.CONNECTION_FAILED,
        message: 'Failed to connect to printer',
      };

      expect(error.type).toBe(PrinterErrorType.CONNECTION_FAILED);
      expect(error.message).toBe('Failed to connect to printer');
    });

    it('should create printer error with original error', () => {
      const originalError = new Error('USB connection failed');
      const error: PrinterError = {
        type: PrinterErrorType.CONNECTION_FAILED,
        message: 'Failed to connect to printer',
        originalError,
      };

      expect(error.originalError).toBe(originalError);
    });
  });
});

describe('PrinterManager', () => {
  let manager: PrinterManager;

  beforeEach(() => {
    resetPrinterManager();
    manager = new PrinterManager();
  });

  afterEach(() => {
    resetPrinterManager();
  });

  describe('Initialization', () => {
    it('should initialize with default configuration', () => {
      expect(manager.getStatus()).toBe(PrinterStatus.DISCONNECTED);
      expect(manager.getCurrentPrinter()).toBeNull();
    });

    it('should initialize with custom configuration', () => {
      const config: PrinterManagerConfig = {
        autoReconnect: true,
        reconnectAttempts: 5,
        reconnectDelay: 2000,
        connectionTimeout: 10000,
      };
      const customManager = new PrinterManager(config);

      expect(customManager).toBeDefined();
    });

    it('should return DISCONNECTED status when not connected', () => {
      expect(manager.getStatus()).toBe(PrinterStatus.DISCONNECTED);
    });

    it('should return null for current printer when not connected', () => {
      expect(manager.getCurrentPrinter()).toBeNull();
    });
  });

  describe('Printer Discovery', () => {
    it('should discover available printers', async () => {
      const printers = await manager.discoverPrinters();

      expect(printers).toBeDefined();
      expect(printers.length).toBeGreaterThan(0);
      expect(printers).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            id: expect.any(String),
            name: expect.any(String),
            connectionType: expect.any(String),
          }),
        ]),
      );
    });

    it('should include USB printer in discovery results', async () => {
      const printers = await manager.discoverPrinters();
      const usbPrinter = printers.find(p => p.connectionType === PrinterConnectionType.USB);

      expect(usbPrinter).toBeDefined();
      expect(usbPrinter?.vendorId).toBeDefined();
      expect(usbPrinter?.productId).toBeDefined();
    });

    it('should include Bluetooth printer in discovery results', async () => {
      const printers = await manager.discoverPrinters();
      const btPrinter = printers.find(p => p.connectionType === PrinterConnectionType.BLUETOOTH);

      expect(btPrinter).toBeDefined();
      expect(btPrinter?.address).toBeDefined();
    });

    it('should include Network printer in discovery results', async () => {
      const printers = await manager.discoverPrinters();
      const netPrinter = printers.find(p => p.connectionType === PrinterConnectionType.NETWORK);

      expect(netPrinter).toBeDefined();
      expect(netPrinter?.address).toBeDefined();
    });
  });

  describe('Printer Connection', () => {
    it('should connect to USB printer successfully', async () => {
      const printers = await manager.discoverPrinters();
      const usbPrinter = printers.find(p => p.connectionType === PrinterConnectionType.USB);

      if (!usbPrinter) {
        throw new Error('No USB printer found');
      }

      const connected = await manager.connect(usbPrinter);

      expect(connected).toBe(true);
      expect(manager.getStatus()).toBe(PrinterStatus.CONNECTED);
      expect(manager.getCurrentPrinter()).toEqual(usbPrinter);
    });

    it('should connect to Bluetooth printer successfully', async () => {
      const printers = await manager.discoverPrinters();
      const btPrinter = printers.find(p => p.connectionType === PrinterConnectionType.BLUETOOTH);

      if (!btPrinter) {
        throw new Error('No Bluetooth printer found');
      }

      const connected = await manager.connect(btPrinter);

      expect(connected).toBe(true);
      expect(manager.getStatus()).toBe(PrinterStatus.CONNECTED);
    });

    it('should connect to Network printer successfully', async () => {
      const printers = await manager.discoverPrinters();
      const netPrinter = printers.find(p => p.connectionType === PrinterConnectionType.NETWORK);

      if (!netPrinter) {
        throw new Error('No Network printer found');
      }

      const connected = await manager.connect(netPrinter);

      expect(connected).toBe(true);
      expect(manager.getStatus()).toBe(PrinterStatus.CONNECTED);
    });

    it('should set status to CONNECTING during connection', async () => {
      const printers = await manager.discoverPrinters();
      const usbPrinter = printers.find(p => p.connectionType === PrinterConnectionType.USB);

      if (!usbPrinter) {
        throw new Error('No USB printer found');
      }

      let statusDuringConnection: PrinterStatus | null = null;
      manager.onStatusChange((status) => {
        if (status === PrinterStatus.CONNECTING) {
          statusDuringConnection = status;
        }
      });

      await manager.connect(usbPrinter);

      expect(statusDuringConnection).toBe(PrinterStatus.CONNECTING);
    });
  });

  describe('Printing', () => {
    it('should print data successfully when connected', async () => {
      const printers = await manager.discoverPrinters();
      const usbPrinter = printers.find(p => p.connectionType === PrinterConnectionType.USB);

      if (!usbPrinter) {
        throw new Error('No USB printer found');
      }

      await manager.connect(usbPrinter);

      const testData = new Uint8Array([0x1B, 0x40, 0x1B, 0x61, 0x01]); // ESC/POS commands
      const printed = await manager.print(testData);

      expect(printed).toBe(true);
    });

    it('should fail to print when not connected', async () => {
      const testData = new Uint8Array([0x1B, 0x40]);
      const printed = await manager.print(testData);

      expect(printed).toBe(false);
    });

    it('should handle error callback when print fails', async () => {
      let errorReceived: PrinterError | null = null;
      manager.onError((error) => {
        errorReceived = error;
      });

      const testData = new Uint8Array([0x1B, 0x40]);
      await manager.print(testData);

      expect(errorReceived).toBeDefined();
      expect(errorReceived?.type).toBe(PrinterErrorType.NOT_CONNECTED);
    });
  });

  describe('Printer Disconnection', () => {
    it('should disconnect from printer successfully', async () => {
      const printers = await manager.discoverPrinters();
      const usbPrinter = printers.find(p => p.connectionType === PrinterConnectionType.USB);

      if (!usbPrinter) {
        throw new Error('No USB printer found');
      }

      await manager.connect(usbPrinter);
      expect(manager.getStatus()).toBe(PrinterStatus.CONNECTED);

      const disconnected = await manager.disconnect();

      expect(disconnected).toBe(true);
      expect(manager.getStatus()).toBe(PrinterStatus.DISCONNECTED);
      expect(manager.getCurrentPrinter()).toBeNull();
    });

    it('should handle disconnect when not connected', async () => {
      const disconnected = await manager.disconnect();

      expect(disconnected).toBe(true);
      expect(manager.getStatus()).toBe(PrinterStatus.DISCONNECTED);
    });
  });

  describe('Status Change Handling', () => {
    it('should call status change handler on connection', async () => {
      const statuses: PrinterStatus[] = [];
      manager.onStatusChange((status) => {
        statuses.push(status);
      });

      const printers = await manager.discoverPrinters();
      const usbPrinter = printers.find(p => p.connectionType === PrinterConnectionType.USB);

      if (!usbPrinter) {
        throw new Error('No USB printer found');
      }

      await manager.connect(usbPrinter);

      expect(statuses).toContain(PrinterStatus.CONNECTING);
      expect(statuses).toContain(PrinterStatus.CONNECTED);
    });

    it('should call status change handler on disconnection', async () => {
      const statuses: PrinterStatus[] = [];
      manager.onStatusChange((status) => {
        statuses.push(status);
      });

      const printers = await manager.discoverPrinters();
      const usbPrinter = printers.find(p => p.connectionType === PrinterConnectionType.USB);

      if (!usbPrinter) {
        throw new Error('No USB printer found');
      }

      await manager.connect(usbPrinter);
      await manager.disconnect();

      expect(statuses).toContain(PrinterStatus.DISCONNECTED);
    });
  });

  describe('Error Handling', () => {
    it('should call error handler on connection failure', async () => {
      let errorReceived: PrinterError | null = null;
      manager.onError((error) => {
        errorReceived = error;
      });

      // Try to connect with invalid device (will fail in mock)
      const invalidDevice: PrinterDevice = {
        id: 'invalid',
        name: 'Invalid Printer',
        connectionType: 'invalid' as PrinterConnectionType,
      };

      await manager.connect(invalidDevice);

      expect(errorReceived).toBeDefined();
      expect(errorReceived?.type).toBe(PrinterErrorType.CONNECTION_FAILED);
    });

    it('should set status to ERROR on connection failure', async () => {
      const invalidDevice: PrinterDevice = {
        id: 'invalid',
        name: 'Invalid Printer',
        connectionType: 'invalid' as PrinterConnectionType,
      };

      await manager.connect(invalidDevice);

      expect(manager.getStatus()).toBe(PrinterStatus.ERROR);
    });
  });
});

describe('Printer Manager Singleton', () => {
  afterEach(() => {
    resetPrinterManager();
  });

  it('should return singleton instance', () => {
    const manager1 = getPrinterManager();
    const manager2 = getPrinterManager();

    expect(manager1).toBe(manager2);
  });

  it('should reset singleton', () => {
    const manager1 = getPrinterManager();
    resetPrinterManager();
    const manager2 = getPrinterManager();

    expect(manager1).not.toBe(manager2);
  });

  it('should create singleton with custom config', () => {
    const config: PrinterManagerConfig = {
      autoReconnect: true,
      reconnectAttempts: 5,
    };
    const manager = getPrinterManager(config);

    expect(manager).toBeDefined();
  });
});
