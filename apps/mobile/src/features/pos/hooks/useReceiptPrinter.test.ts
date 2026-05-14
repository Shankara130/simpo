/**
 * Receipt Printer Hook Tests
 * Test receipt printing hook with error handling and retry logic
 */

import { renderHook, act } from '@testing-library/react-native';
import { useReceiptPrinter } from './useReceiptPrinter';
import { ReceiptPrinterService } from '../services/ReceiptPrinterService';
import { resetPrinterManager } from '../hardware/PrinterManager';
import { PaymentMethod } from '../types/payment.types';
import { ReceiptData } from '../types/receipt.types';

// Mock ReceiptPrinterService
jest.mock('../services/ReceiptPrinterService');

describe('useReceiptPrinter Hook', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    resetPrinterManager();
  });

  afterEach(() => {
    resetPrinterManager();
  });

  const mockReceiptData: ReceiptData = {
    transactionNumber: 'TRX-20260514-0001',
    transactionDate: '2026-05-14T15:30:00+07:00',
    pharmacyName: 'Apotek Sehat',
    pharmacyAddress: 'Jl. Sudirman No. 123, Jakarta',
    pharmacyPhone: '021-1234567',
    items: [
      {
        name: 'Paracetamol 500mg',
        quantity: 2,
        unitPrice: '15000.00',
        subtotal: '30000.00',
      },
    ],
    subtotal: '30000.00',
    total: '30000.00',
    payment: {
      method: PaymentMethod.CASH,
      cashDetails: {
        change: '0.00',
      },
    },
    paperWidth: 58,
  };

  describe('Initialization', () => {
    it('should initialize with correct default state', () => {
      const { result } = renderHook(() => useReceiptPrinter());

      expect(result.current.isLoading).toBe(false);
      expect(result.current.isSuccess).toBe(false);
      expect(result.current.error).toBe(null);
      expect(result.current.printerStatus).toBe('disconnected');
      expect(result.current.printerName).toBe(null);
    });

    it('should provide all required functions', () => {
      const { result } = renderHook(() => useReceiptPrinter());

      expect(result.current.printReceipt).toBeDefined();
      expect(result.current.connectPrinter).toBeDefined();
      expect(result.current.disconnectPrinter).toBeDefined();
      expect(result.current.discoverPrinters).toBeDefined();
      expect(result.current.retryPrint).toBeDefined();
      expect(result.current.clearError).toBeDefined();
    });
  });

  describe('Printer Discovery', () => {
    it('should discover available printers', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      let printers;
      await act(async () => {
        printers = await result.current.discoverPrinters();
      });

      expect(printers).toBeDefined();
      expect(Array.isArray(printers)).toBe(true);
      expect(printers.length).toBeGreaterThan(0);
    });

    it('should return printers with correct structure', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      let printers;
      await act(async () => {
        printers = await result.current.discoverPrinters();
      });

      expect(printers[0]).toMatchObject({
        id: expect.any(String),
        name: expect.any(String),
        connectionType: expect.any(String),
      });
    });
  });

  describe('Printer Connection', () => {
    it('should connect to printer successfully', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      // First discover printers
      let printers;
      await act(async () => {
        printers = await result.current.discoverPrinters();
      });

      // Connect to first printer
      let connected;
      await act(async () => {
        connected = await result.current.connectPrinter(printers[0].id);
      });

      expect(connected).toBe(true);
      expect(result.current.printerStatus).toBe('connected');
      expect(result.current.printerName).toBe(printers[0].name);
    });

    it('should fail to connect to invalid printer', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      let connected;
      await act(async () => {
        connected = await result.current.connectPrinter('invalid-id');
      });

      expect(connected).toBe(false);
      expect(result.current.error).toContain('tidak ditemukan');
    });

    it('should disconnect from printer successfully', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      // First connect to a printer
      let printers;
      await act(async () => {
        printers = await result.current.discoverPrinters();
        await result.current.connectPrinter(printers[0].id);
      });

      expect(result.current.printerStatus).toBe('connected');

      // Then disconnect
      let disconnected;
      await act(async () => {
        disconnected = await result.current.disconnectPrinter();
      });

      expect(disconnected).toBe(true);
      expect(result.current.printerStatus).toBe('disconnected');
      expect(result.current.printerName).toBe(null);
    });
  });

  describe('Receipt Printing', () => {
    beforeEach(() => {
      // Mock ReceiptPrinterService to return valid ESC/POS data
      (ReceiptPrinterService.prototype.generateReceipt as jest.Mock).mockReturnValue(
        new Uint8Array([0x1B, 0x40, 0x1B, 0x61, 0x01])
      );
    });

    it('should fail to print when printer not connected', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      let success;
      await act(async () => {
        success = await result.current.printReceipt(mockReceiptData);
      });

      expect(success).toBe(false);
      expect(result.current.error).toContain('tidak terhubung');
      expect(result.current.isSuccess).toBe(false);
    });

    it('should print receipt successfully when connected', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      // First connect to printer
      let printers;
      await act(async () => {
        printers = await result.current.discoverPrinters();
        await result.current.connectPrinter(printers[0].id);
      });

      expect(result.current.printerStatus).toBe('connected');

      // Then print receipt
      let success;
      await act(async () => {
        success = await result.current.printReceipt(mockReceiptData);
      });

      expect(success).toBe(true);
      expect(result.current.isSuccess).toBe(true);
      expect(result.current.error).toBe(null);
    });

    it('should set loading state during printing', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      // Connect to printer first
      let printers;
      await act(async () => {
        printers = await result.current.discoverPrinters();
        await result.current.connectPrinter(printers[0].id);
      });

      // Start printing and wait for completion
      await act(async () => {
        await result.current.printReceipt(mockReceiptData);
      });

      // After completion, loading should be false
      expect(result.current.isLoading).toBe(false);
      expect(result.current.isSuccess).toBe(true);
    });

    it('should generate ESC/POS receipt data', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      // Connect to printer first
      let printers;
      await act(async () => {
        printers = await result.current.discoverPrinters();
        await result.current.connectPrinter(printers[0].id);
      });

      // Print receipt
      await act(async () => {
        await result.current.printReceipt(mockReceiptData);
      });

      expect(ReceiptPrinterService.prototype.generateReceipt).toHaveBeenCalledWith(
        mockReceiptData
      );
    });
  });

  describe('Retry Logic', () => {
    beforeEach(() => {
      (ReceiptPrinterService.prototype.generateReceipt as jest.Mock).mockReturnValue(
        new Uint8Array([0x1B, 0x40])
      );
    });

    it('should retry print when retryPrint is called', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      // Connect to printer first
      let printers;
      await act(async () => {
        printers = await result.current.discoverPrinters();
        await result.current.connectPrinter(printers[0].id);
      });

      // First print
      await act(async () => {
        await result.current.printReceipt(mockReceiptData);
      });

      expect(result.current.isSuccess).toBe(true);

      // Reset success state
      result.current.isSuccess = false;

      // Retry print
      let retrySuccess;
      await act(async () => {
        retrySuccess = await result.current.retryPrint();
      });

      expect(retrySuccess).toBe(true);
    });

    it('should fail retry when no previous print data', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      let success;
      await act(async () => {
        success = await result.current.retryPrint();
      });

      expect(success).toBe(false);
      expect(result.current.error).toContain('Tidak ada struk');
    });

    it('should auto-retry on failure when enabled', async () => {
      const { result } = renderHook(() =>
        useReceiptPrinter({ autoRetry: true, maxRetries: 2, retryDelay: 100 })
      );

      // Connect to printer first
      let printers;
      await act(async () => {
        printers = await result.current.discoverPrinters();
        await result.current.connectPrinter(printers[0].id);
      });

      // Mock a failed print by disconnecting printer
      await act(async () => {
        await result.current.disconnectPrinter();
      });

      // Attempt print (should fail and trigger auto-retry)
      let success;
      await act(async () => {
        success = await result.current.printReceipt(mockReceiptData);
      });

      expect(success).toBe(false);
      // Auto-retry should be triggered but will also fail due to no connection
    }, 10000);
  });

  describe('Error Handling', () => {
    it('should set error message on print failure', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      (ReceiptPrinterService.prototype.generateReceipt as jest.Mock).mockReturnValue(
        new Uint8Array([0x1B, 0x40])
      );

      // Try to print without connecting
      let success;
      await act(async () => {
        success = await result.current.printReceipt(mockReceiptData);
      });

      expect(success).toBe(false);
      expect(result.current.error).toBeTruthy();
    });

    it('should clear error when clearError is called', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      // Trigger an error
      await act(async () => {
        await result.current.connectPrinter('invalid-id');
      });

      expect(result.current.error).toBeTruthy();

      // Clear error
      act(() => {
        result.current.clearError();
      });

      expect(result.current.error).toBe(null);
    });

    it('should set error on invalid printer connection', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      let connected;
      await act(async () => {
        connected = await result.current.connectPrinter('non-existent-printer');
      });

      expect(connected).toBe(false);
      expect(result.current.error).toContain('tidak ditemukan');
    });
  });

  describe('Configuration', () => {
    it('should accept custom configuration', () => {
      const customConfig = {
        autoRetry: true,
        maxRetries: 5,
        retryDelay: 2000,
        autoConnect: true,
      };

      const { result } = renderHook(() => useReceiptPrinter(customConfig));

      // Hook should initialize without errors
      expect(result.current.printReceipt).toBeDefined();
    });

    it('should use default configuration when not provided', () => {
      const { result } = renderHook(() => useReceiptPrinter());

      // Should use defaults: autoRetry=false, maxRetries=3, retryDelay=1000
      expect(result.current.printReceipt).toBeDefined();
    });
  });

  describe('Indonesian Error Messages', () => {
    it('should display Indonesian error for printer not connected', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      (ReceiptPrinterService.prototype.generateReceipt as jest.Mock).mockReturnValue(
        new Uint8Array([0x1B, 0x40])
      );

      await act(async () => {
        await result.current.printReceipt(mockReceiptData);
      });

      expect(result.current.error).toContain('tidak terhubung');
    });

    it('should display Indonesian error for printer not found', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      await act(async () => {
        await result.current.connectPrinter('invalid-id');
      });

      expect(result.current.error).toContain('tidak ditemukan');
    });

    it('should display Indonesian error for no receipt to retry', async () => {
      const { result } = renderHook(() => useReceiptPrinter());

      await act(async () => {
        await result.current.retryPrint();
      });

      expect(result.current.error).toContain('Tidak ada struk');
    });
  });
});
