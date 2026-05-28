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
  loadPairedBluetoothDevices,
  savePairedBluetoothDevices,
  loadLastConnectedDevice,
  saveLastConnectedDevice,
  clearLastConnectedDevice,
  loadBluetoothConfig,
  saveBluetoothConfig,
  BLUETOOTH_DEVICES_KEY,
  LAST_CONNECTED_DEVICE_KEY,
  BLUETOOTH_CONFIG_KEY,
} from './ScannerConfigService';
import { DEFAULT_SCANNER_CONFIG, DEFAULT_BLUETOOTH_CONFIG } from '../types/scanner.types';

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

  // ============================================================================
  // Bluetooth Settings Tests (Story 7.3)
  // ============================================================================

  describe('Bluetooth Paired Devices', () => {
    it('should return empty array when no paired devices exist', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(null);

      const devices = await loadPairedBluetoothDevices();

      expect(devices).toEqual([]);
      expect(AsyncStorage.getItem).toHaveBeenCalledWith(BLUETOOTH_DEVICES_KEY);
    });

    it('should return saved paired devices', async () => {
      const savedDevices = [
        { id: 'device-1', name: 'Zebra Scanner', type: 'BLE', connected: false, paired: true, rssi: -60 },
        { id: 'device-2', name: 'Honeywell Scanner', type: 'BLE', connected: true, paired: true, rssi: -50 },
      ];

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(JSON.stringify(savedDevices));

      const devices = await loadPairedBluetoothDevices();

      expect(devices).toEqual(savedDevices);
    });

    it('should return empty array on invalid data', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValue('invalid json');

      const devices = await loadPairedBluetoothDevices();

      expect(devices).toEqual([]);
    });

    it('should return empty array on error', async () => {
      (AsyncStorage.getItem as jest.Mock).mockRejectedValue(new Error('Storage error'));

      const devices = await loadPairedBluetoothDevices();

      expect(devices).toEqual([]);
    });

    it('should save paired devices', async () => {
      const devices = [
        { id: 'device-1', name: 'Zebra Scanner', type: 'BLE', connected: false, paired: true, rssi: -60 },
      ];

      (AsyncStorage.setItem as jest.Mock).mockResolvedValue(undefined);

      await savePairedBluetoothDevices(devices);

      expect(AsyncStorage.setItem).toHaveBeenCalledWith(
        BLUETOOTH_DEVICES_KEY,
        JSON.stringify(devices)
      );
    });

    it('should throw error on save failure', async () => {
      const devices = [];
      (AsyncStorage.setItem as jest.Mock).mockRejectedValue(new Error('Storage error'));

      await expect(savePairedBluetoothDevices(devices)).rejects.toThrow('Storage error');
    });
  });

  describe('Last Connected Device', () => {
    it('should return null when no last device exists', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(null);

      const deviceId = await loadLastConnectedDevice();

      expect(deviceId).toBeNull();
      expect(AsyncStorage.getItem).toHaveBeenCalledWith(LAST_CONNECTED_DEVICE_KEY);
    });

    it('should return saved last device ID', async () => {
      const savedDeviceId = 'device-123';

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(savedDeviceId);

      const deviceId = await loadLastConnectedDevice();

      expect(deviceId).toBe(savedDeviceId);
    });

    it('should return null on error', async () => {
      (AsyncStorage.getItem as jest.Mock).mockRejectedValue(new Error('Storage error'));

      const deviceId = await loadLastConnectedDevice();

      expect(deviceId).toBeNull();
    });

    it('should save last connected device ID', async () => {
      const deviceId = 'device-123';

      (AsyncStorage.setItem as jest.Mock).mockResolvedValue(undefined);

      await saveLastConnectedDevice(deviceId);

      expect(AsyncStorage.setItem).toHaveBeenCalledWith(LAST_CONNECTED_DEVICE_KEY, deviceId);
    });

    it('should throw error on save failure', async () => {
      (AsyncStorage.setItem as jest.Mock).mockRejectedValue(new Error('Storage error'));

      await expect(saveLastConnectedDevice('device-123')).rejects.toThrow('Storage error');
    });

    it('should clear last connected device', async () => {
      (AsyncStorage.removeItem as jest.Mock).mockResolvedValue(undefined);

      await clearLastConnectedDevice();

      expect(AsyncStorage.removeItem).toHaveBeenCalledWith(LAST_CONNECTED_DEVICE_KEY);
    });

    it('should throw error on clear failure', async () => {
      (AsyncStorage.removeItem as jest.Mock).mockRejectedValue(new Error('Storage error'));

      await expect(clearLastConnectedDevice()).rejects.toThrow('Storage error');
    });
  });

  describe('Bluetooth Configuration', () => {
    it('should return default config when no saved config exists', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(null);

      const config = await loadBluetoothConfig();

      expect(config).toEqual(DEFAULT_BLUETOOTH_CONFIG);
      expect(AsyncStorage.getItem).toHaveBeenCalledWith(BLUETOOTH_CONFIG_KEY);
    });

    it('should return saved config when it exists', async () => {
      const savedConfig = {
        autoReconnect: false,
        maxReconnectAttempts: 10,
        reconnectDelays: [500, 1000, 2000, 4000, 8000],
        connectionTimeout: 15000,
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(JSON.stringify(savedConfig));

      const config = await loadBluetoothConfig();

      expect(config).toEqual(savedConfig);
    });

    it('should merge saved config with defaults for missing properties', async () => {
      const partialConfig = {
        autoReconnect: false,
        // Missing other properties
      };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(JSON.stringify(partialConfig));

      const config = await loadBluetoothConfig();

      expect(config.autoReconnect).toBe(false);
      expect(config.maxReconnectAttempts).toBe(DEFAULT_BLUETOOTH_CONFIG.maxReconnectAttempts);
      expect(config.reconnectDelays).toEqual(DEFAULT_BLUETOOTH_CONFIG.reconnectDelays);
      expect(config.connectionTimeout).toBe(DEFAULT_BLUETOOTH_CONFIG.connectionTimeout);
    });

    it('should return defaults on AsyncStorage error', async () => {
      (AsyncStorage.getItem as jest.Mock).mockRejectedValue(new Error('Storage error'));

      const config = await loadBluetoothConfig();

      expect(config).toEqual(DEFAULT_BLUETOOTH_CONFIG);
    });

    it('should save partial config by merging with existing', async () => {
      const currentConfig = DEFAULT_BLUETOOTH_CONFIG;
      const partialUpdate = { autoReconnect: false };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(JSON.stringify(currentConfig));
      (AsyncStorage.setItem as jest.Mock).mockResolvedValue(undefined);

      await saveBluetoothConfig(partialUpdate);

      const expectedConfig = { ...currentConfig, ...partialUpdate };
      expect(AsyncStorage.setItem).toHaveBeenCalledWith(
        BLUETOOTH_CONFIG_KEY,
        JSON.stringify(expectedConfig)
      );
    });

    it('should save new config when no existing config', async () => {
      const newConfig = { autoReconnect: false };

      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(null);
      (AsyncStorage.setItem as jest.Mock).mockResolvedValue(undefined);

      await saveBluetoothConfig(newConfig);

      const expectedConfig = { ...DEFAULT_BLUETOOTH_CONFIG, ...newConfig };
      expect(AsyncStorage.setItem).toHaveBeenCalledWith(
        BLUETOOTH_CONFIG_KEY,
        JSON.stringify(expectedConfig)
      );
    });

    it('should throw error on save failure', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValue(JSON.stringify(DEFAULT_BLUETOOTH_CONFIG));
      (AsyncStorage.setItem as jest.Mock).mockRejectedValue(new Error('Storage error'));

      await expect(saveBluetoothConfig({ autoReconnect: false })).rejects.toThrow('Storage error');
    });
  });

  describe('Bluetooth Settings Public API', () => {
    it('should export Bluetooth settings methods', () => {
      expect(ScannerConfigService.loadPairedDevices).toBe(loadPairedBluetoothDevices);
      expect(ScannerConfigService.savePairedDevices).toBe(savePairedBluetoothDevices);
      expect(ScannerConfigService.loadLastConnectedDevice).toBe(loadLastConnectedDevice);
      expect(ScannerConfigService.saveLastConnectedDevice).toBe(saveLastConnectedDevice);
      expect(ScannerConfigService.clearLastConnectedDevice).toBe(clearLastConnectedDevice);
      expect(ScannerConfigService.loadBluetoothConfig).toBe(loadBluetoothConfig);
      expect(ScannerConfigService.saveBluetoothConfig).toBe(saveBluetoothConfig);
    });
  });
});
