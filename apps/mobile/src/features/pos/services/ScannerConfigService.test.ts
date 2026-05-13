/**
 * Tests for ScannerConfigService
 * Tests persistent storage of scanner configuration using AsyncStorage
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import {
  ScannerConfigService,
  loadScannerConfig,
  saveScannerConfig,
  resetScannerConfig,
  SCANNER_CONFIG_KEY,
} from './ScannerConfigService';
import { DEFAULT_SCANNER_CONFIG } from '../types/scanner.types';

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
}));

describe('ScannerConfigService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('loadScannerConfig', () => {
    it('should return default config when no saved config exists', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(null);

      const config = await loadScannerConfig();

      expect(config).toEqual(DEFAULT_SCANNER_CONFIG);
      expect(AsyncStorage.getItem).toHaveBeenCalledWith(SCANNER_CONFIG_KEY);
    });

    it('should return saved config when it exists', async () => {
      const savedConfig = {
        debounceMs: 1000,
        maxScanTimeMs: 200,
        minBarcodeLength: 10,
        maxBarcodeLength: 15,
        feedbackEnabled: false,
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(JSON.stringify(savedConfig));

      const config = await loadScannerConfig();

      expect(config).toEqual(savedConfig);
    });

    it('should merge saved config with defaults for missing properties', async () => {
      const partialConfig = {
        debounceMs: 1000,
        // Missing other properties
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(JSON.stringify(partialConfig));

      const config = await loadScannerConfig();

      expect(config.debounceMs).toBe(1000);
      expect(config.maxScanTimeMs).toBe(DEFAULT_SCANNER_CONFIG.maxScanTimeMs);
      expect(config.minBarcodeLength).toBe(DEFAULT_SCANNER_CONFIG.minBarcodeLength);
      expect(config.maxBarcodeLength).toBe(DEFAULT_SCANNER_CONFIG.maxBarcodeLength);
      expect(config.feedbackEnabled).toBe(DEFAULT_SCANNER_CONFIG.feedbackEnabled);
    });

    it('should return defaults on AsyncStorage error', async () => {
      (AsyncStorage.getItem as jest.Mock).mockRejectedValue(new Error('Storage error'));

      const config = await loadScannerConfig();

      expect(config).toEqual(DEFAULT_SCANNER_CONFIG);
    });

    it('should return defaults on invalid JSON', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValue('invalid json');

      const config = await loadScannerConfig();

      expect(config).toEqual(DEFAULT_SCANNER_CONFIG);
    });
  });

  describe('saveScannerConfig', () => {
    it('should save partial config by merging with existing', async () => {
      const currentConfig = DEFAULT_SCANNER_CONFIG;
      const partialUpdate = { debounceMs: 1000 };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(JSON.stringify(currentConfig));
      (AsyncStorage.setItem as jest.Mock).mockResolvedValue(undefined);

      await saveScannerConfig(partialUpdate);

      const expectedConfig = { ...currentConfig, ...partialUpdate };
      expect(AsyncStorage.setItem).toHaveBeenCalledWith(
        SCANNER_CONFIG_KEY,
        JSON.stringify(expectedConfig)
      );
    });

    it('should save new config when no existing config', async () => {
      const newConfig = { debounceMs: 1000 };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(null);
      (AsyncStorage.setItem as jest.Mock).mockResolvedValue(undefined);

      await saveScannerConfig(newConfig);

      const expectedConfig = { ...DEFAULT_SCANNER_CONFIG, ...newConfig };
      expect(AsyncStorage.setItem).toHaveBeenCalledWith(
        SCANNER_CONFIG_KEY,
        JSON.stringify(expectedConfig)
      );
    });

    it('should save all properties when provided', async () => {
      const fullConfig = {
        debounceMs: 1000,
        maxScanTimeMs: 200,
        minBarcodeLength: 10,
        maxBarcodeLength: 15,
        feedbackEnabled: false,
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(JSON.stringify(DEFAULT_SCANNER_CONFIG));
      (AsyncStorage.setItem as jest.Mock).mockResolvedValue(undefined);

      await saveScannerConfig(fullConfig);

      expect(AsyncStorage.setItem).toHaveBeenCalledWith(
        SCANNER_CONFIG_KEY,
        JSON.stringify(fullConfig)
      );
    });

    it('should throw error on AsyncStorage failure', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(JSON.stringify(DEFAULT_SCANNER_CONFIG));
      (AsyncStorage.setItem as jest.Mock).mockRejectedValue(new Error('Storage error'));

      await expect(saveScannerConfig({ debounceMs: 1000 })).rejects.toThrow('Storage error');
    });
  });

  describe('resetScannerConfig', () => {
    it('should remove config from AsyncStorage', async () => {
      (AsyncStorage.removeItem as jest.Mock).mockResolvedValue(undefined);

      await resetScannerConfig();

      expect(AsyncStorage.removeItem).toHaveBeenCalledWith(SCANNER_CONFIG_KEY);
    });

    it('should throw error on AsyncStorage failure', async () => {
      (AsyncStorage.removeItem as jest.Mock).mockRejectedValue(new Error('Storage error'));

      await expect(resetScannerConfig()).rejects.toThrow('Storage error');
    });
  });

  describe('ScannerConfigService public API', () => {
    it('should export all methods and constants', () => {
      expect(ScannerConfigService.load).toBe(loadScannerConfig);
      expect(ScannerConfigService.save).toBe(saveScannerConfig);
      expect(ScannerConfigService.reset).toBe(resetScannerConfig);
      expect(ScannerConfigService.KEY).toBe(SCANNER_CONFIG_KEY);
    });
  });
});
