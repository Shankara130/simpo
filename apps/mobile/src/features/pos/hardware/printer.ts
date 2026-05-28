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

  // Cash Drawer Methods (Story 7.4)

  /**
   * Open cash drawer via ESC/POS kick command
   * @param options - Cash drawer options (pulse timing, pin number, enabled)
   * @param onResult - Callback for success/failure result
   * @returns Promise resolving to true if drawer opened successfully
   */
  openCashDrawer(
    options: CashDrawerOptions,
    onResult?: (success: boolean, error?: string) => void
  ): Promise<boolean>;

  /**
   * Check if cash drawer is connected (inferred from printer connection)
   * @returns true if drawer is connected via printer
   */
  readonly isDrawerConnected: boolean;

  /**
   * Get current drawer status
   * @returns Current drawer status
   */
  getDrawerStatus(): DrawerStatus;

  /**
   * Set drawer result handler callback
   * @param handler - Callback function for drawer operation results
   */
  onDrawerResult(handler: (success: boolean, error?: string) => void): void;

  /**
   * Clear drawer result handler callback
   */
  clearDrawerResultHandler(): void;
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

// ============================================================================
// Cash Drawer Support (Story 7.4)
// ============================================================================

/**
 * Cash Drawer Options Interface
 * Configuration options for cash drawer control
 */
export interface CashDrawerOptions {
  /** Pulse duration in milliseconds (typically 50-200ms) */
  pulseTiming: number;
  /** Drawer pin number (0 = pin 2, 1 = pin 5) */
  pinNumber: 0 | 1;
  /** Enable/disable automatic drawer opening */
  enabled: boolean;
  /** Timeout for drawer open operation in milliseconds (default: 10000) */
  drawerOpenTimeoutMs?: number;
  /** Mechanical delay for drawer opening in milliseconds (default: 200) */
  mechanicalDelayMs?: number;
}

/**
 * Cash Drawer Status Type
 * Represents the current state of the cash drawer connection
 */
export type DrawerStatus = 'disconnected' | 'connected' | 'opening' | 'failed';

/**
 * Cash Drawer Pin Number Enum
 * RJ-12 connector has two drawer pins: Pin 2 (drawer 1) and Pin 5 (drawer 2)
 */
export enum DrawerPin {
  PIN_2 = 0, // Drawer 1 trigger (most commonly used)
  PIN_5 = 1, // Drawer 2 trigger (for dual drawer systems)
}

/**
 * Cash Drawer Configuration Interface
 * Persisted configuration for cash drawer behavior
 */
export interface CashDrawerConfig {
  /** Enable automatic drawer opening for cash payments */
  autoOpen: boolean;
  /** Pulse duration in milliseconds (50-500ms) */
  pulseMs: number;
  /** Drawer pin selection (Pin 2 or Pin 5) */
  pinNumber: 0 | 1;
}

/**
 * Default cash drawer configuration
 */
export const DEFAULT_CASH_DRAWER_CONFIG: CashDrawerConfig = {
  autoOpen: true,
  pulseMs: 100, // 100ms default pulse
  pinNumber: DrawerPin.PIN_2, // Pin 2 is most common
};
