/**
 * Printer Manager Implementation
 * Manages thermal printer connections and printing operations
 * This is a mock implementation for development - production would use native modules
 */

import {
  PrinterDevice,
  PrinterStatus,
  PrinterConnectionType,
  PrinterErrorType,
  PrinterError,
  IPrinterManager,
  IPrinterConnection,
  PrinterManagerConfig,
} from './printer';

/**
 * Mock USB Printer Connection
 */
class MockUSBCPrinterConnection implements IPrinterConnection {
  private connected = false;

  constructor(private device: PrinterDevice) {}

  async connect(): Promise<boolean> {
    // Simulate connection delay
    await new Promise(resolve => setTimeout(resolve, 500));
    this.connected = true;
    return true;
  }

  async disconnect(): Promise<boolean> {
    this.connected = false;
    return true;
  }

  async send(data: Uint8Array): Promise<number> {
    if (!this.connected) {
      throw new Error('Printer not connected');
    }
    // Simulate printing delay
    await new Promise(resolve => setTimeout(resolve, 100));
    return data.length;
  }

  getStatus(): PrinterStatus {
    return this.connected ? PrinterStatus.CONNECTED : PrinterStatus.DISCONNECTED;
  }

  isReady(): boolean {
    return this.connected;
  }
}

/**
 * Mock Bluetooth Printer Connection
 */
class MockBluetoothPrinterConnection implements IPrinterConnection {
  private connected = false;

  constructor(private device: PrinterDevice) {}

  async connect(): Promise<boolean> {
    await new Promise(resolve => setTimeout(resolve, 800));
    this.connected = true;
    return true;
  }

  async disconnect(): Promise<boolean> {
    this.connected = false;
    return true;
  }

  async send(data: Uint8Array): Promise<number> {
    if (!this.connected) {
      throw new Error('Printer not connected');
    }
    await new Promise(resolve => setTimeout(resolve, 150));
    return data.length;
  }

  getStatus(): PrinterStatus {
    return this.connected ? PrinterStatus.CONNECTED : PrinterStatus.DISCONNECTED;
  }

  isReady(): boolean {
    return this.connected;
  }
}

/**
 * Mock Network Printer Connection
 */
class MockNetworkPrinterConnection implements IPrinterConnection {
  private connected = false;

  constructor(private device: PrinterDevice) {}

  async connect(): Promise<boolean> {
    await new Promise(resolve => setTimeout(resolve, 300));
    this.connected = true;
    return true;
  }

  async disconnect(): Promise<boolean> {
    this.connected = false;
    return true;
  }

  async send(data: Uint8Array): Promise<number> {
    if (!this.connected) {
      throw new Error('Printer not connected');
    }
    await new Promise(resolve => setTimeout(resolve, 50));
    return data.length;
  }

  getStatus(): PrinterStatus {
    return this.connected ? PrinterStatus.CONNECTED : PrinterStatus.DISCONNECTED;
  }

  isReady(): boolean {
    return this.connected;
  }
}

/**
 * Printer Manager Implementation
 */
export class PrinterManager implements IPrinterManager {
  private currentPrinter: PrinterDevice | null = null;
  private currentConnection: IPrinterConnection | null = null;
  private currentStatus: PrinterStatus = PrinterStatus.DISCONNECTED;
  private errorHandler?: (error: PrinterError) => void;
  private statusHandler?: (status: PrinterStatus) => void;
  private config: PrinterManagerConfig;

  constructor(config: PrinterManagerConfig = {}) {
    this.config = {
      autoReconnect: false,
      reconnectAttempts: 3,
      reconnectDelay: 1000,
      connectionTimeout: 5000,
      ...config,
    };
  }

  /**
   * Discover available printers
   * In production, this would use platform-specific APIs
   */
  async discoverPrinters(): Promise<PrinterDevice[]> {
    // Mock implementation - returns sample printers
    // In production, this would use native modules for USB/Bluetooth/Network discovery
    return [
      {
        id: 'usb-1',
        name: 'PT-210 Thermal Printer',
        connectionType: PrinterConnectionType.USB,
        vendorId: 0x0DD4,
        productId: 0x0141,
      },
      {
        id: 'bt-1',
        name: 'POS-58BT',
        connectionType: PrinterConnectionType.BLUETOOTH,
        address: '00:11:22:33:44:55',
      },
      {
        id: 'net-1',
        name: 'Network Printer',
        connectionType: PrinterConnectionType.NETWORK,
        address: '192.168.1.100',
      },
    ];
  }

  /**
   * Connect to a specific printer
   */
  async connect(device: PrinterDevice): Promise<boolean> {
    try {
      this.updateStatus(PrinterStatus.CONNECTING);

      // Create connection based on printer type
      let connection: IPrinterConnection;
      switch (device.connectionType) {
        case PrinterConnectionType.USB:
          connection = new MockUSBCPrinterConnection(device);
          break;
        case PrinterConnectionType.BLUETOOTH:
          connection = new MockBluetoothPrinterConnection(device);
          break;
        case PrinterConnectionType.NETWORK:
          connection = new MockNetworkPrinterConnection(device);
          break;
        default:
          throw new Error(`Unsupported connection type: ${device.connectionType}`);
      }

      // Attempt connection
      const connected = await connection.connect();
      if (!connected) {
        this.handleError({
          type: PrinterErrorType.CONNECTION_FAILED,
          message: `Failed to connect to printer: ${device.name}`,
        });
        return false;
      }

      // Store connection and update status
      this.currentPrinter = device;
      this.currentConnection = connection;
      this.updateStatus(PrinterStatus.CONNECTED);
      return true;
    } catch (error) {
      this.handleError({
        type: PrinterErrorType.CONNECTION_FAILED,
        message: `Connection error: ${error instanceof Error ? error.message : 'Unknown error'}`,
        originalError: error,
      });
      this.updateStatus(PrinterStatus.ERROR);
      return false;
    }
  }

  /**
   * Disconnect from current printer
   */
  async disconnect(): Promise<boolean> {
    try {
      if (this.currentConnection) {
        await this.currentConnection.disconnect();
      }
      this.currentPrinter = null;
      this.currentConnection = null;
      this.updateStatus(PrinterStatus.DISCONNECTED);
      return true;
    } catch (error) {
      this.handleError({
        type: PrinterErrorType.DISCONNECTION_FAILED,
        message: `Disconnection error: ${error instanceof Error ? error.message : 'Unknown error'}`,
        originalError: error,
      });
      return false;
    }
  }

  /**
   * Print receipt data
   */
  async print(data: Uint8Array): Promise<boolean> {
    try {
      if (!this.currentConnection || !this.currentConnection.isReady()) {
        this.handleError({
          type: PrinterErrorType.NOT_CONNECTED,
          message: 'No printer connected',
        });
        return false;
      }

      // Send data to printer
      const bytesWritten = await this.currentConnection.send(data);
      if (bytesWritten !== data.length) {
        this.handleError({
          type: PrinterErrorType.SEND_FAILED,
          message: `Incomplete data transfer: ${bytesWritten}/${data.length} bytes`,
        });
        return false;
      }

      return true;
    } catch (error) {
      this.handleError({
        type: PrinterErrorType.SEND_FAILED,
        message: `Print error: ${error instanceof Error ? error.message : 'Unknown error'}`,
        originalError: error,
      });
      return false;
    }
  }

  /**
   * Get current printer status
   */
  getStatus(): PrinterStatus {
    return this.currentStatus;
  }

  /**
   * Get current connected printer
   */
  getCurrentPrinter(): PrinterDevice | null {
    return this.currentPrinter;
  }

  /**
   * Set error handler callback
   */
  onError(handler: (error: PrinterError) => void): void {
    this.errorHandler = handler;
  }

  /**
   * Set status change handler callback
   */
  onStatusChange(handler: (status: PrinterStatus) => void): void {
    this.statusHandler = handler;
  }

  /**
   * Update printer status and notify handlers
   */
  private updateStatus(status: PrinterStatus): void {
    this.currentStatus = status;
    if (this.statusHandler) {
      this.statusHandler(status);
    }
  }

  /**
   * Handle printer errors
   */
  private handleError(error: PrinterError): void {
    if (this.errorHandler) {
      this.errorHandler(error);
    }
  }
}

/**
 * Singleton instance for app-wide printer management
 */
let printerManagerInstance: PrinterManager | null = null;

/**
 * Get or create printer manager singleton
 */
export function getPrinterManager(config?: PrinterManagerConfig): PrinterManager {
  if (!printerManagerInstance) {
    printerManagerInstance = new PrinterManager(config);
  }
  return printerManagerInstance;
}

/**
 * Reset printer manager singleton (useful for testing)
 */
export function resetPrinterManager(): void {
  if (printerManagerInstance) {
    printerManagerInstance.disconnect();
  }
  printerManagerInstance = null;
}
