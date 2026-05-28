/**
 * Printer Manager Implementation
 * Manages thermal printer connections via USB, Bluetooth, and Network
 * Implements singleton pattern for app-wide printer access
 */

import {
  PrinterConnectionType,
  PrinterStatus,
  PrinterDevice,
  PrinterError,
  PrinterErrorType,
  IPrinterManager,
  PrinterManagerConfig,
  CashDrawerOptions,
  DrawerStatus,
} from './printer';
import { ThermalPrinterModule } from '@finan-me/react-native-thermal-printer';
import { ESCPOSGenerator } from '../services/ESCPOSGenerator';
import { getAuditLogService } from '../services/AuditLogService';

/**
 * Printer Manager Class
 * Singleton implementation for managing thermal printer connections
 */
class PrinterManagerClass implements IPrinterManager {
  private static instance: PrinterManagerClass | null = null;
  private currentPrinter: PrinterDevice | null = null;
  private currentStatus: PrinterStatus = PrinterStatus.DISCONNECTED;
  private errorHandler?: (error: PrinterError) => void;
  private statusChangeHandler?: (status: PrinterStatus) => void;
  private config: PrinterManagerConfig;
  private isReconnecting: boolean = false; // Guard against concurrent reconnect attempts

  // Cash Drawer Support (Story 7.4)
  private drawerStatus: DrawerStatus = 'disconnected';
  private drawerResultHandler?: (success: boolean, error?: string) => void;
  private drawerOperationInProgress: boolean = false; // Guard against concurrent drawer opens

  private constructor(config: PrinterManagerConfig = {}) {
    this.config = {
      autoReconnect: config.autoReconnect ?? false,
      reconnectAttempts: config.reconnectAttempts ?? 3,
      reconnectDelay: config.reconnectDelay ?? 1000,
      connectionTimeout: config.connectionTimeout ?? 5000,
    };
  }

  /**
   * Get singleton instance of PrinterManager
   */
  public static getInstance(config?: PrinterManagerConfig): PrinterManagerClass {
    if (!PrinterManagerClass.instance) {
      PrinterManagerClass.instance = new PrinterManagerClass(config);
    }
    return PrinterManagerClass.instance;
  }

  /**
   * Discover available printers across all connection types
   */
  public async discoverPrinters(): Promise<PrinterDevice[]> {
    const discoveredPrinters: PrinterDevice[] = [];

    try {
      // Discover USB printers
      const usbDevices = await ThermalPrinterModule.getUsbDevices();
      const usbPrinters = usbDevices.map((device: unknown) => ({
        id: (device as { id?: string }).id || `usb-${Date.now()}`,
        name: (device as { name?: string }).name || 'Unknown USB Printer',
        address: (device as { address?: string }).address,
        connectionType: PrinterConnectionType.USB,
      }));
      discoveredPrinters.push(...usbPrinters);

      // Discover Bluetooth printers
      const bluetoothDevices = await ThermalPrinterModule.getBluetoothDevices();
      const bluetoothPrinters = bluetoothDevices.map((device: unknown) => ({
        id: (device as { id?: string }).id || `bt-${Date.now()}`,
        name: (device as { name?: string }).name || 'Unknown Bluetooth Printer',
        address: (device as { address?: string }).address,
        connectionType: PrinterConnectionType.BLUETOOTH,
      }));
      discoveredPrinters.push(...bluetoothPrinters);

      // Discover Network printers
      const networkPrinters = await ThermalPrinterModule.getNetPrinters();
      const networkPrintersMapped = networkPrinters.map((device: unknown) => ({
        id: (device as { id?: string }).id || `net-${Date.now()}`,
        name: (device as { name?: string }).name || 'Unknown Network Printer',
        address: (device as { address?: string }).address,
        connectionType: PrinterConnectionType.NETWORK,
      }));
      discoveredPrinters.push(...networkPrintersMapped);
    } catch (error) {
      this.handleError({
        type: PrinterErrorType.DEVICE_NOT_FOUND,
        message: 'Failed to discover printers',
        originalError: error,
      });
      // Re-throw to allow caller to handle discovery failures
      throw new Error(`Printer discovery failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }

    return discoveredPrinters;
  }

  /**
   * Connect to a specific printer
   */
  public async connect(device: PrinterDevice): Promise<boolean> {
    try {
      this.updateStatus(PrinterStatus.CONNECTING);

      let connected = false;

      switch (device.connectionType) {
        case PrinterConnectionType.USB:
          connected = await ThermalPrinterModule.connectToUsbDevice(device);
          break;
        case PrinterConnectionType.BLUETOOTH:
          connected = await ThermalPrinterModule.connectToBluetoothDevice(device);
          break;
        case PrinterConnectionType.NETWORK:
          connected = await ThermalPrinterModule.connectToNetPrinter(device);
          break;
        default:
          throw new Error(`Unsupported connection type: ${device.connectionType}`);
      }

      if (connected) {
        this.currentPrinter = device;
        this.updateStatus(PrinterStatus.CONNECTED);
        return true;
      } else {
        this.updateStatus(PrinterStatus.ERROR);
        this.handleError({
          type: PrinterErrorType.CONNECTION_FAILED,
          message: `Failed to connect to ${device.name}`,
        });
        return false;
      }
    } catch (error) {
      this.updateStatus(PrinterStatus.ERROR);
      this.handleError({
        type: PrinterErrorType.CONNECTION_FAILED,
        message: `Error connecting to ${device.name}`,
        originalError: error,
      });
      return false;
    }
  }

  /**
   * Disconnect from current printer
   */
  public async disconnect(): Promise<boolean> {
    try {
      if (!this.currentPrinter) {
        this.updateStatus(PrinterStatus.DISCONNECTED);
        return true;
      }

      // Track which printer we're disconnecting for validation
      const disconnectingPrinterId = this.currentPrinter.id;
      const disconnectingPrinterName = this.currentPrinter.name;

      const disconnected = await ThermalPrinterModule.disconnect();

      if (disconnected) {
        // Validate we disconnected the expected printer
        if (this.currentPrinter?.id === disconnectingPrinterId) {
          this.currentPrinter = null;
          this.updateStatus(PrinterStatus.DISCONNECTED);

          // Clean up drawer state on disconnect
          this.drawerStatus = 'disconnected';
          this.drawerResultHandler = undefined;
          this.drawerOperationInProgress = false;

          return true;
        } else {
          // Printer changed during disconnect - handle unexpected state
          this.handleError({
            type: PrinterErrorType.DISCONNECTION_FAILED,
            message: 'Printer state changed during disconnect',
          });
          return false;
        }
      } else {
        this.handleError({
          type: PrinterErrorType.DISCONNECTION_FAILED,
          message: `Failed to disconnect from ${disconnectingPrinterName}`,
        });
        return false;
      }
    } catch (error) {
      this.handleError({
        type: PrinterErrorType.DISCONNECTION_FAILED,
        message: 'Error disconnecting from printer',
        originalError: error,
      });
      return false;
    }
  }

  /**
   * Print receipt data
   */
  public async print(data: Uint8Array): Promise<boolean> {
    try {
      // Double-check status before and during print operation
      if (!this.currentPrinter || this.currentStatus !== PrinterStatus.CONNECTED) {
        this.handleError({
          type: PrinterErrorType.NOT_CONNECTED,
          message: 'No printer connected',
        });
        return false;
      }

      // Store current printer ID for validation after print attempt
      const printerId = this.currentPrinter.id;

      const result = await ThermalPrinterModule.print(data);

      // Validate printer still connected after print attempt
      if (this.currentPrinter?.id !== printerId || this.currentStatus !== PrinterStatus.CONNECTED) {
        this.handleError({
          type: PrinterErrorType.SEND_FAILED,
          message: 'Printer disconnected during print operation',
        });
        return false;
      }

      if (!result) {
        this.handleError({
          type: PrinterErrorType.SEND_FAILED,
          message: 'Failed to send data to printer',
        });
      }

      return result;
    } catch (error) {
      this.handleError({
        type: PrinterErrorType.SEND_FAILED,
        message: 'Error printing data',
        originalError: error,
      });
      return false;
    }
  }

  // ============================================================================
  // Cash Drawer Support (Story 7.4)
  // ============================================================================

  /**
   * Open cash drawer via ESC/POS kick command
   * @param options - Drawer configuration options
   * @param onResult - Callback for success/failure (for audit logging)
   */
  public async openCashDrawer(
    options: CashDrawerOptions,
    onResult?: (success: boolean, error?: string) => void
  ): Promise<boolean> {
    // Guard against concurrent drawer opens
    if (this.drawerOperationInProgress) {
      const errorMsg = 'Drawer operation already in progress';
      onResult?.(false, errorMsg);
      this.drawerResultHandler?.(false, errorMsg);
      return false;
    }

    this.drawerOperationInProgress = true;

    try {
      // Check if drawer is enabled
      if (!options.enabled) {
        this.drawerStatus = 'disconnected';
        const errorMsg = 'Drawer disabled in configuration';
        onResult?.(false, errorMsg);
        this.drawerResultHandler?.(false, errorMsg);
        return false;
      }

      // Verify printer is connected (explicit state validation)
      if (!this.currentPrinter || this.currentStatus !== PrinterStatus.CONNECTED) {
        this.drawerStatus = 'disconnected'; // Use 'disconnected' instead of 'failed' for clarity
        const errorMsg = 'Printer not connected';
        onResult?.(false, errorMsg);
        this.drawerResultHandler?.(false, errorMsg);
        this.handleError({
          type: PrinterErrorType.NOT_CONNECTED,
          message: 'Cannot open drawer: printer not connected',
        });
        return false;
      }

      this.drawerStatus = 'opening';

      // Generate ESC/POS cash drawer kick command
      const escposGenerator = new ESCPOSGenerator();
      const drawerCommand = escposGenerator.generateCashDrawerKick({
        pulseTiming: options.pulseTiming,
        pinNumber: options.pinNumber,
      });

      // Send command to printer via USB with timeout (default 10 seconds)
      const printTimeout = options.drawerOpenTimeoutMs || 10000;
      const result = await Promise.race([
        ThermalPrinterModule.print(drawerCommand),
        new Promise<boolean>((_, reject) =>
          setTimeout(() => reject(new Error('Printer operation timed out')), printTimeout)
        ),
      ]) as boolean;

      if (!result) {
        // Check if printer disconnected during operation
        if (this.currentStatus !== PrinterStatus.CONNECTED) {
          this.drawerStatus = 'disconnected';
          const errorMsg = 'Printer disconnected during operation';
          onResult?.(false, errorMsg);
          this.drawerResultHandler?.(false, errorMsg);
          this.handleError({
            type: PrinterErrorType.NOT_CONNECTED,
            message: errorMsg,
          });
          return false;
        }

        this.drawerStatus = 'failed';
        const errorMsg = 'Failed to send drawer command to printer';
        this.handleDrawerError(false, errorMsg, onResult);
        this.handleError({
          type: PrinterErrorType.SEND_FAILED,
          message: errorMsg,
        });
        return false;
      }

      // Wait for drawer to open (mechanical delay - configurable via options)
      const mechanicalDelay = options.mechanicalDelayMs || 200;
      await this.delay(mechanicalDelay);

      // Reset drawer status to connected after success
      this.drawerStatus = 'connected';
      this.handleDrawerError(true, undefined, onResult);
      return true;
    } catch (error) {
      // Check if error is timeout or printer disconnected
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';

      if (errorMessage.includes('timed out') || this.currentStatus !== PrinterStatus.CONNECTED) {
        this.drawerStatus = 'disconnected'; // Use 'disconnected' for timeout/disconnect
      } else {
        this.drawerStatus = 'failed';
      }

      this.handleDrawerError(false, errorMessage, onResult);
      this.handleError({
        type: PrinterErrorType.SEND_FAILED,
        message: `Cash drawer open failed: ${errorMessage}`,
        originalError: error,
      });
      return false;
    } finally {
      this.drawerOperationInProgress = false;
    }
  }

  /**
   * Handle drawer operation errors with centralized callback logic
   */
  private handleDrawerError(
    success: boolean,
    error?: string,
    additionalCallback?: (success: boolean, error?: string) => void
  ): void {
    // Use single callback mechanism - prefer drawerResultHandler if set
    if (this.drawerResultHandler) {
      this.drawerResultHandler(success, error);
    }
    // Fall back to additional callback if drawerResultHandler not set
    else if (additionalCallback) {
      additionalCallback(success, error);
    }
  }

  /**
   * Check if cash drawer is connected
   * (Inferred from printer connection since drawer connects via printer)
   */
  public get isDrawerConnected(): boolean {
    return this.currentStatus === PrinterStatus.CONNECTED;
  }

  /**
   * Get current drawer status
   */
  public getDrawerStatus(): DrawerStatus {
    return this.drawerStatus;
  }

  /**
   * Set drawer result handler callback
   * @param handler - Callback function for drawer operation results
   */
  public onDrawerResult(handler: (success: boolean, error?: string) => void): void {
    this.drawerResultHandler = handler;
  }

  /**
   * Clear drawer result handler callback
   */
  public clearDrawerResultHandler(): void {
    this.drawerResultHandler = undefined;
  }

  /**
   * Delay helper function
   * @param ms - Milliseconds to delay
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Get current printer status
   */
  public getStatus(): PrinterStatus {
    return this.currentStatus;
  }

  /**
   * Get currently connected printer
   */
  public getCurrentPrinter(): PrinterDevice | null {
    return this.currentPrinter;
  }

  /**
   * Set error handler callback
   */
  public onError(handler: (error: PrinterError) => void): void {
    this.errorHandler = handler;
  }

  /**
   * Set status change handler callback
   */
  public onStatusChange(handler: (status: PrinterStatus) => void): void {
    this.statusChangeHandler = handler;
  }

  /**
   * Clear status change handler callback
   */
  public clearStatusChangeHandler(): void {
    this.statusChangeHandler = undefined;
  }

  /**
   * Clear error handler callback
   */
  public clearErrorHandler(): void {
    this.errorHandler = undefined;
  }

  /**
   * Update printer status and notify listeners
   */
  private updateStatus(status: PrinterStatus): void {
    if (this.currentStatus !== status) {
      this.currentStatus = status;
      if (this.statusChangeHandler) {
        this.statusChangeHandler(status);
      }
    }
  }

  /**
   * Handle printer errors
   */
  private handleError(error: PrinterError): void {
    if (this.errorHandler) {
      this.errorHandler(error);
    }

    // Attempt auto-reconnect if enabled
    if (this.config.autoReconnect && this.currentPrinter) {
      this.attemptReconnect();
    }
  }

  /**
   * Attempt to reconnect to printer with retries
   */
  private async attemptReconnect(): Promise<void> {
    // Guard: Prevent concurrent reconnect attempts
    if (this.isReconnecting) {
      return;
    }

    if (!this.currentPrinter || !this.config.autoReconnect) {
      return;
    }

    this.isReconnecting = true;

    try {
      const attempts = this.config.reconnectAttempts || 3;
      const delay = this.config.reconnectDelay || 1000;

      for (let i = 0; i < attempts; i++) {
        await new Promise(resolve => setTimeout(resolve, delay));

        const connected = await this.connect(this.currentPrinter);
        if (connected) {
          return; // Successfully reconnected
        }
      }

      // All attempts failed
      this.handleError({
        type: PrinterErrorType.CONNECTION_FAILED,
        message: `Failed to reconnect after ${attempts} attempts`,
      });
    } finally {
      this.isReconnecting = false;
    }
  }

  /**
   * Reset singleton instance (for testing purposes)
   */
  public static resetInstance(): void {
    if (PrinterManagerClass.instance) {
      PrinterManagerClass.instance.disconnect();
    }
    PrinterManagerClass.instance = null;
  }
}

// Export the singleton class with proper naming
export const PrinterManager = PrinterManagerClass;

// Legacy exports for backward compatibility
export function getPrinterManager(config?: PrinterManagerConfig): PrinterManagerClass {
  return PrinterManagerClass.getInstance(config);
}

export function resetPrinterManager(): void {
  PrinterManagerClass.resetInstance();
}
