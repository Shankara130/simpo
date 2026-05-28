/**
 * PrinterConfigService Tests
 * Tests for printer and drawer configuration persistence
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import {
  PrinterConfigService,
  loadPrinterConfig,
  savePrinterConfig,
  resetPrinterConfig,
  loadDrawerConfig,
  saveDrawerConfig,
  resetDrawerConfig,
  DEFAULT_PRINTER_CONFIG,
  DEFAULT_CASH_DRAWER_CONFIG,
  PRINTER_CONFIG_KEY,
  DRAWER_CONFIG_KEY,
} from '../services/PrinterConfigService';

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
}));

describe('PrinterConfigService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Printer Configuration', () => {
    it('should load default config when no saved config exists', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);

      const config = await loadPrinterConfig();

      expect(config).toEqual(DEFAULT_PRINTER_CONFIG);
      expect(AsyncStorage.getItem).toHaveBeenCalledWith(PRINTER_CONFIG_KEY);
    });

    it('should load saved config when available', async () => {
      const savedConfig = {
        paperWidth: 80,
        autoCut: false,
        defaultPrinterId: 'printer-123',
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(
        JSON.stringify(savedConfig)
      );

      const config = await loadPrinterConfig();

      expect(config).toEqual({
        ...DEFAULT_PRINTER_CONFIG,
        ...savedConfig,
      });
    });

    it('should merge saved config with defaults', async () => {
      const partialConfig = {
        paperWidth: 80,
        // autoCut is missing, should use default
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(
        JSON.stringify(partialConfig)
      );

      const config = await loadPrinterConfig();

      expect(config).toEqual({
        ...DEFAULT_PRINTER_CONFIG,
        ...partialConfig,
      });
      expect(config.autoCut).toBe(DEFAULT_PRINTER_CONFIG.autoCut);
    });

    it('should save printer config', async () => {
      const newConfig = {
        paperWidth: 80,
        autoCut: false,
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(
        JSON.stringify(DEFAULT_PRINTER_CONFIG)
      );
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);

      await savePrinterConfig(newConfig);

      expect(AsyncStorage.setItem).toHaveBeenCalledWith(
        PRINTER_CONFIG_KEY,
        JSON.stringify({
          ...DEFAULT_PRINTER_CONFIG,
          ...newConfig,
        })
      );
    });

    it('should reset printer config to defaults', async () => {
      (AsyncStorage.removeItem as jest.Mock).mockResolvedValueOnce(undefined);

      await resetPrinterConfig();

      expect(AsyncStorage.removeItem).toHaveBeenCalledWith(PRINTER_CONFIG_KEY);
    });

    it('should handle load errors gracefully', async () => {
      (AsyncStorage.getItem as jest.Mock).mockRejectedValueOnce(
        new Error('Storage error')
      );

      const config = await loadPrinterConfig();

      expect(config).toEqual(DEFAULT_PRINTER_CONFIG);
    });

    it('should handle save errors', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(
        JSON.stringify(DEFAULT_PRINTER_CONFIG)
      );
      (AsyncStorage.setItem as jest.Mock).mockRejectedValueOnce(
        new Error('Storage error')
      );

      await expect(savePrinterConfig({ paperWidth: 80 })).rejects.toThrow(
        'Storage error'
      );
    });
  });

  describe('Cash Drawer Configuration', () => {
    it('should load default drawer config when no saved config exists', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);

      const config = await loadDrawerConfig();

      expect(config).toEqual(DEFAULT_CASH_DRAWER_CONFIG);
      expect(AsyncStorage.getItem).toHaveBeenCalledWith(DRAWER_CONFIG_KEY);
    });

    it('should load saved drawer config when available', async () => {
      const savedConfig = {
        autoOpen: false,
        pulseMs: 150,
        pinNumber: 1, // Pin 5
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(
        JSON.stringify(savedConfig)
      );

      const config = await loadDrawerConfig();

      expect(config).toEqual({
        ...DEFAULT_CASH_DRAWER_CONFIG,
        ...savedConfig,
      });
    });

    it('should merge saved drawer config with defaults', async () => {
      const partialConfig = {
        pulseMs: 200,
        // autoOpen and pinNumber are missing, should use defaults
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(
        JSON.stringify(partialConfig)
      );

      const config = await loadDrawerConfig();

      expect(config).toEqual({
        ...DEFAULT_CASH_DRAWER_CONFIG,
        ...partialConfig,
      });
      expect(config.autoOpen).toBe(DEFAULT_CASH_DRAWER_CONFIG.autoOpen);
      expect(config.pinNumber).toBe(DEFAULT_CASH_DRAWER_CONFIG.pinNumber);
    });

    it('should save drawer config', async () => {
      const newConfig = {
        autoOpen: false,
        pulseMs: 150,
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(
        JSON.stringify(DEFAULT_CASH_DRAWER_CONFIG)
      );
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);

      await saveDrawerConfig(newConfig);

      expect(AsyncStorage.setItem).toHaveBeenCalledWith(
        DRAWER_CONFIG_KEY,
        JSON.stringify({
          ...DEFAULT_CASH_DRAWER_CONFIG,
          ...newConfig,
        })
      );
    });

    it('should reset drawer config to defaults', async () => {
      (AsyncStorage.removeItem as jest.Mock).mockResolvedValueOnce(undefined);

      await resetDrawerConfig();

      expect(AsyncStorage.removeItem).toHaveBeenCalledWith(DRAWER_CONFIG_KEY);
    });

    it('should handle drawer load errors gracefully', async () => {
      (AsyncStorage.getItem as jest.Mock).mockRejectedValueOnce(
        new Error('Storage error')
      );

      const config = await loadDrawerConfig();

      expect(config).toEqual(DEFAULT_CASH_DRAWER_CONFIG);
    });

    it('should handle drawer save errors', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(
        JSON.stringify(DEFAULT_CASH_DRAWER_CONFIG)
      );
      (AsyncStorage.setItem as jest.Mock).mockRejectedValueOnce(
        new Error('Storage error')
      );

      await expect(saveDrawerConfig({ autoOpen: false })).rejects.toThrow(
        'Storage error'
      );
    });

    it('should validate pulse timing range (50-500ms)', async () => {
      const config = await loadDrawerConfig();

      // Default should be in valid range
      expect(config.pulseMs).toBeGreaterThanOrEqual(50);
      expect(config.pulseMs).toBeLessThanOrEqual(500);

      // Test boundary values
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);
      const defaultConfig = await loadDrawerConfig();
      expect(defaultConfig.pulseMs).toBe(100); // Default is 100ms
    });

    it('should validate pin number values (0 or 1)', async () => {
      const config = await loadDrawerConfig();

      // Pin number should be 0 or 1
      expect([0, 1]).toContain(config.pinNumber);
    });
  });

  describe('PrinterConfigService Public API', () => {
    it('should provide all printer methods', () => {
      expect(PrinterConfigService.loadPrinter).toBeDefined();
      expect(PrinterConfigService.savePrinter).toBeDefined();
      expect(PrinterConfigService.resetPrinter).toBeDefined();
      expect(PrinterConfigService.PRINTER_KEY).toBe(PRINTER_CONFIG_KEY);
    });

    it('should provide all drawer methods', () => {
      expect(PrinterConfigService.loadDrawer).toBeDefined();
      expect(PrinterConfigService.saveDrawer).toBeDefined();
      expect(PrinterConfigService.resetDrawer).toBeDefined();
      expect(PrinterConfigService.DRAWER_KEY).toBe(DRAWER_CONFIG_KEY);
    });
  });
});
