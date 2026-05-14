/**
 * Receipt Printer Hook
 * Custom hook for managing receipt printing operations
 * Integrates ReceiptPrinterService with PrinterManager for complete printing workflow
 */

import { useState, useCallback, useRef, useEffect } from 'react';
import { PrinterManager, getPrinterManager } from '../hardware/PrinterManager';
import { PrinterStatus, PrinterError, PrinterErrorType } from '../hardware/printer';
import { ReceiptPrinterService } from '../services/ReceiptPrinterService';
import { ReceiptData } from '../types/receipt.types';

/**
 * Receipt printing state
 */
export interface ReceiptPrintingState {
  isLoading: boolean;
  isSuccess: boolean;
  error: string | null;
  printerStatus: PrinterStatus;
  printerName: string | null;
}

/**
 * Receipt printer hook result
 */
export interface UseReceiptPrinterResult extends ReceiptPrintingState {
  printReceipt: (receiptData: ReceiptData) => Promise<boolean>;
  connectPrinter: (printerId: string) => Promise<boolean>;
  disconnectPrinter: () => Promise<boolean>;
  discoverPrinters: () => Promise<PrinterDevice[]>;
  retryPrint: () => Promise<boolean>;
  clearError: () => void;
}

/**
 * Printer device info for discovery
 */
export interface PrinterDevice {
  id: string;
  name: string;
  connectionType: string;
  address?: string;
  vendorId?: number;
  productId?: number;
}

/**
 * Receipt printer hook configuration
 */
export interface UseReceiptPrinterConfig {
  /** Enable auto-retry on print failure */
  autoRetry?: boolean;
  /** Maximum retry attempts */
  maxRetries?: number;
  /** Delay between retries in milliseconds */
  retryDelay?: number;
  /** Auto-connect to last used printer */
  autoConnect?: boolean;
}

const DEFAULT_CONFIG: UseReceiptPrinterConfig = {
  autoRetry: false,
  maxRetries: 3,
  retryDelay: 1000,
  autoConnect: false,
};

/**
 * Receipt Printer Hook
 * Manages receipt printing with automatic error handling and retry logic
 */
export function useReceiptPrinter(config: UseReceiptPrinterConfig = {}): UseReceiptPrinterResult {
  const finalConfig = { ...DEFAULT_CONFIG, ...config };

  const [state, setState] = useState<ReceiptPrintingState>({
    isLoading: false,
    isSuccess: false,
    error: null,
    printerStatus: PrinterStatus.DISCONNECTED,
    printerName: null,
  });

  const [lastReceiptData, setLastReceiptData] = useState<ReceiptData | null>(null);
  const printerManagerRef = useRef<PrinterManager | null>(null);
  const retryCountRef = useRef(0);

  // Initialize printer manager on mount
  useEffect(() => {
    if (!printerManagerRef.current) {
      printerManagerRef.current = getPrinterManager();

      // Set up status change handler
      printerManagerRef.current.onStatusChange((status) => {
        setState((prev) => ({
          ...prev,
          printerStatus: status,
          printerName: status === PrinterStatus.CONNECTED
            ? printerManagerRef.current?.getCurrentPrinter()?.name || null
            : null,
        }));
      });

      // Set up error handler
      printerManagerRef.current.onError((error: PrinterError) => {
        setState((prev) => ({
          ...prev,
          error: error.message,
        }));
      });
    }

    return () => {
      // Cleanup on unmount
      if (printerManagerRef.current && finalConfig.autoConnect) {
        printerManagerRef.current.disconnect();
      }
    };
  }, []);

  const receiptPrinterServiceRef = useRef<ReceiptPrinterService>(
    new ReceiptPrinterService()
  );

  /**
   * Print receipt with error handling
   */
  const printReceipt = useCallback(async (receiptData: ReceiptData): Promise<boolean> => {
    setState((prev) => ({
      ...prev,
      isLoading: true,
      isSuccess: false,
      error: null,
    }));

    setLastReceiptData(receiptData);
    retryCountRef.current = 0;

    try {
      // Generate ESC/POS receipt data
      const receiptBuffer = receiptPrinterServiceRef.current.generateReceipt(receiptData);

      // Check printer connection
      const printerManager = printerManagerRef.current;
      if (!printerManager) {
        throw new Error('Printer manager not initialized');
      }

      if (printerManager.getStatus() !== PrinterStatus.CONNECTED) {
        setState((prev) => ({
          ...prev,
          isLoading: false,
          error: 'Printer tidak terhubung. Silakan hubungkan printer terlebih dahulu.',
        }));
        return false;
      }

      // Send to printer
      const printSuccess = await printerManager.print(receiptBuffer);

      if (printSuccess) {
        setState((prev) => ({
          ...prev,
          isLoading: false,
          isSuccess: true,
          error: null,
        }));
        return true;
      } else {
        throw new Error('Gagal mencetak struk');
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Gagal mencetak struk';

      setState((prev) => ({
        ...prev,
        isLoading: false,
        isSuccess: false,
        error: errorMessage,
      }));

      // Auto-retry if enabled
      if (finalConfig.autoRetry && retryCountRef.current < finalConfig.maxRetries!) {
        retryCountRef.current++;
        setTimeout(() => {
          retryPrint();
        }, finalConfig.retryDelay);
      }

      return false;
    }
  }, [finalConfig.autoRetry, finalConfig.maxRetries, finalConfig.retryDelay]);

  /**
   * Retry last print operation
   */
  const retryPrint = useCallback(async (): Promise<boolean> => {
    if (!lastReceiptData) {
      setState((prev) => ({
        ...prev,
        error: 'Tidak ada struk yang dapat dicetak ulang',
      }));
      return false;
    }

    return printReceipt(lastReceiptData);
  }, [lastReceiptData, printReceipt]);

  /**
   * Connect to printer
   */
  const connectPrinter = useCallback(async (printerId: string): Promise<boolean> => {
    setState((prev) => ({
      ...prev,
      isLoading: true,
      error: null,
    }));

    try {
      const printerManager = printerManagerRef.current;
      if (!printerManager) {
        throw new Error('Printer manager not initialized');
      }

      // Discover printers to find the one with matching ID
      const printers = await printerManager.discoverPrinters();
      const targetPrinter = printers.find((p) => p.id === printerId);

      if (!targetPrinter) {
        setState((prev) => ({
          ...prev,
          isLoading: false,
          error: `Printer dengan ID ${printerId} tidak ditemukan`,
        }));
        return false;
      }

      // Connect to printer
      const connected = await printerManager.connect(targetPrinter);

      setState((prev) => ({
        ...prev,
        isLoading: false,
      }));

      return connected;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Gagal menghubungkan printer';

      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: errorMessage,
      }));

      return false;
    }
  }, []);

  /**
   * Disconnect from printer
   */
  const disconnectPrinter = useCallback(async (): Promise<boolean> => {
    try {
      const printerManager = printerManagerRef.current;
      if (!printerManager) {
        return false;
      }

      const disconnected = await printerManager.disconnect();

      setState((prev) => ({
        ...prev,
        printerName: null,
      }));

      return disconnected;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Gagal memutuskan printer';

      setState((prev) => ({
        ...prev,
        error: errorMessage,
      }));

      return false;
    }
  }, []);

  /**
   * Discover available printers
   */
  const discoverPrinters = useCallback(async (): Promise<PrinterDevice[]> => {
    try {
      const printerManager = printerManagerRef.current;
      if (!printerManager) {
        throw new Error('Printer manager not initialized');
      }

      const printers = await printerManager.discoverPrinters();

      return printers.map((p) => ({
        id: p.id,
        name: p.name,
        connectionType: p.connectionType,
        address: p.address,
        vendorId: p.vendorId,
        productId: p.productId,
      }));
    } catch (error) {
      setState((prev) => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Gagal menemukan printer',
      }));

      return [];
    }
  }, []);

  /**
   * Clear error state
   */
  const clearError = useCallback(() => {
    setState((prev) => ({
      ...prev,
      error: null,
    }));
  }, []);

  return {
    isLoading: state.isLoading,
    isSuccess: state.isSuccess,
    error: state.error,
    printerStatus: state.printerStatus,
    printerName: state.printerName,
    printReceipt,
    connectPrinter,
    disconnectPrinter,
    discoverPrinters,
    retryPrint,
    clearError,
  };
}
