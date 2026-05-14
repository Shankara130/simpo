/**
 * Printer Hardware Interface Abstraction
 * Handles thermal printer connections via USB, Bluetooth, and Network
 * Provides platform-specific printer integration for React Native
 */

/**
 * Printer connection type enum
 */
export enum PrinterConnectionType {
  USB = 'usb',
  BLUETOOTH = 'bluetooth',
  NETWORK = 'network',
}

/**
 * Printer status enum
 */
export enum PrinterStatus {
  CONNECTED = 'connected',
  DISCONNECTED = 'disconnected',
  ERROR = 'error',
  OUT_OF_PAPER = 'out_of_paper',
  CONNECTING = 'connecting',
}

/**
 * Printer device information
 */
export interface PrinterDevice {
  id: string;
  name: string;
  connectionType: PrinterConnectionType;
  address?: string;  // MAC address for Bluetooth, IP for Network
  vendorId?: number;  // USB vendor ID
  productId?: number;  // USB product ID
}

/**
 * Printer connection interface
 * Abstract interface for printer connectivity
 */
export interface IPrinterConnection {
  /**
   * Connect to printer
   * @returns Promise resolving to true if connection successful
   */
  connect(): Promise<boolean>;

  /**
   * Disconnect from printer
   * @returns Promise resolving to true if disconnection successful
   */
  disconnect(): Promise<boolean>;

  /**
   * Send data to printer
   * @param data Data to send (ESC/POS commands as Uint8Array)
   * @returns Promise resolving to bytes written
   */
  send(data: Uint8Array): Promise<number>;

  /**
   * Get current printer status
   * @returns Current printer status
   */
  getStatus(): PrinterStatus;

  /**
   * Check if printer is ready
   * @returns true if printer is connected and ready
   */
  isReady(): boolean;
}

/**
 * Printer error types
 */
export enum PrinterErrorType {
  CONNECTION_FAILED = 'connection_failed',
  DISCONNECTION_FAILED = 'disconnection_failed',
  SEND_FAILED = 'send_failed',
  NOT_CONNECTED = 'not_connected',
  OUT_OF_PAPER = 'out_of_paper',
  DEVICE_NOT_FOUND = 'device_not_found',
  PERMISSION_DENIED = 'permission_denied',
  TIMEOUT = 'timeout',
}

/**
 * Printer error interface
 */
export interface PrinterError {
  type: PrinterErrorType;
  message: string;
  originalError?: unknown;
}

/**
 * Printer manager configuration
 */
export interface PrinterManagerConfig {
  autoReconnect?: boolean;
  reconnectAttempts?: number;
  reconnectDelay?: number;  // milliseconds
  connectionTimeout?: number;  // milliseconds
}

/**
 * Printer manager interface
 * Manages printer connections and provides high-level printing API
 */
export interface IPrinterManager {
  /**
   * Discover available printers
   * @returns Promise resolving to list of discovered printers
   */
  discoverPrinters(): Promise<PrinterDevice[]>;

  /**
   * Connect to a specific printer
   * @param device Printer device to connect to
   * @returns Promise resolving to true if connection successful
   */
  connect(device: PrinterDevice): Promise<boolean>;

  /**
   * Disconnect from current printer
   * @returns Promise resolving to true if disconnection successful
   */
  disconnect(): Promise<boolean>;

  /**
   * Print receipt data
   * @param data ESC/POS formatted receipt data
   * @returns Promise resolving to true if print successful
   */
  print(data: Uint8Array): Promise<boolean>;

  /**
   * Get current printer status
   * @returns Current printer status
   */
  getStatus(): PrinterStatus;

  /**
   * Get current connected printer
   * @returns Currently connected printer device or null
   */
  getCurrentPrinter(): PrinterDevice | null;

  /**
   * Set error handler callback
   * @param handler Error handler function
   */
  onError(handler: (error: PrinterError) => void): void;

  /**
   * Set status change handler callback
   * @param handler Status change handler function
   */
  onStatusChange(handler: (status: PrinterStatus) => void): void;
}

/**
 * USB Printer connection interface
 */
export interface IUSBCPrinterConnection extends IPrinterConnection {
  getVendorId(): number;
  getProductId(): number;
}

/**
 * Bluetooth Printer connection interface
 */
export interface IBluetoothPrinterConnection extends IPrinterConnection {
  getMacAddress(): string;
}

/**
 * Network Printer connection interface
 */
export interface INetworkPrinterConnection extends IPrinterConnection {
  getIpAddress(): string;
  getPort(): number;
}
