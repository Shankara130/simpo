/**
 * ScannerConfigService - Persistent storage for scanner settings
 * Handles loading and saving scanner configuration using AsyncStorage
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import { ScannerConfig, DEFAULT_SCANNER_CONFIG, BluetoothDevice, BluetoothConfig, DEFAULT_BLUETOOTH_CONFIG } from '../types/scanner.types';

// AsyncStorage keys for scanner preferences
export const SCANNER_CONFIG_KEY = '@simpo_scanner_config';
export const BLUETOOTH_DEVICES_KEY = '@simpo_bluetooth_devices';
export const LAST_CONNECTED_DEVICE_KEY = '@simpo_last_connected_device';
export const BLUETOOTH_CONFIG_KEY = '@simpo_bluetooth_config';

/**
 * Load scanner configuration from AsyncStorage
 * Returns default config if no saved config exists
 */
export const loadScannerConfig = async (): Promise<ScannerConfig> => {
  try {
    const savedConfig = await AsyncStorage.getItem(SCANNER_CONFIG_KEY);

    if (savedConfig) {
      const parsed = JSON.parse(savedConfig);

      // Merge with defaults to ensure all properties exist
      return {
        ...DEFAULT_SCANNER_CONFIG,
        ...parsed,
      };
    }

    // No saved config, return defaults
    return DEFAULT_SCANNER_CONFIG;
  } catch (error) {
    console.warn('[ScannerConfigService] Failed to load config, using defaults:', error);
    return DEFAULT_SCANNER_CONFIG;
  }
};

/**
 * Save scanner configuration to AsyncStorage
 */
export const saveScannerConfig = async (config: Partial<ScannerConfig>): Promise<void> => {
  try {
    // Load current config first
    const currentConfig = await loadScannerConfig();

    // Merge with new values
    const updatedConfig = {
      ...currentConfig,
      ...config,
    };

    await AsyncStorage.setItem(SCANNER_CONFIG_KEY, JSON.stringify(updatedConfig));
  } catch (error) {
    console.error('[ScannerConfigService] Failed to save config:', error);
    throw error;
  }
};

/**
 * Reset scanner configuration to defaults
 */
export const resetScannerConfig = async (): Promise<void> => {
  try {
    await AsyncStorage.removeItem(SCANNER_CONFIG_KEY);
  } catch (error) {
    console.error('[ScannerConfigService] Failed to reset config:', error);
    throw error;
  }
};

// ============================================================================
// Bluetooth Settings Persistence (Story 7.3)
// ============================================================================

/**
 * Load list of paired Bluetooth devices
 */
export const loadPairedBluetoothDevices = async (): Promise<BluetoothDevice[]> => {
  try {
    const savedDevices = await AsyncStorage.getItem(BLUETOOTH_DEVICES_KEY);

    if (savedDevices) {
      const parsed = JSON.parse(savedDevices);
      return Array.isArray(parsed) ? parsed : [];
    }

    return [];
  } catch (error) {
    console.warn('[ScannerConfigService] Failed to load paired devices:', error);
    return [];
  }
};

/**
 * Save list of paired Bluetooth devices
 */
export const savePairedBluetoothDevices = async (devices: BluetoothDevice[]): Promise<void> => {
  try {
    await AsyncStorage.setItem(BLUETOOTH_DEVICES_KEY, JSON.stringify(devices));
  } catch (error) {
    console.error('[ScannerConfigService] Failed to save paired devices:', error);
    throw error;
  }
};

/**
 * Load last-connected Bluetooth device
 */
export const loadLastConnectedDevice = async (): Promise<string | null> => {
  try {
    const savedDeviceId = await AsyncStorage.getItem(LAST_CONNECTED_DEVICE_KEY);
    return savedDeviceId;
  } catch (error) {
    console.warn('[ScannerConfigService] Failed to load last connected device:', error);
    return null;
  }
};

/**
 * Save last-connected Bluetooth device ID
 */
export const saveLastConnectedDevice = async (deviceId: string): Promise<void> => {
  try {
    await AsyncStorage.setItem(LAST_CONNECTED_DEVICE_KEY, deviceId);
  } catch (error) {
    console.error('[ScannerConfigService] Failed to save last connected device:', error);
    throw error;
  }
};

/**
 * Clear last-connected device (e.g., after unpairing)
 */
export const clearLastConnectedDevice = async (): Promise<void> => {
  try {
    await AsyncStorage.removeItem(LAST_CONNECTED_DEVICE_KEY);
  } catch (error) {
    console.error('[ScannerConfigService] Failed to clear last connected device:', error);
    throw error;
  }
};

/**
 * Load Bluetooth configuration
 */
export const loadBluetoothConfig = async (): Promise<BluetoothConfig> => {
  try {
    const savedConfig = await AsyncStorage.getItem(BLUETOOTH_CONFIG_KEY);

    if (savedConfig) {
      const parsed = JSON.parse(savedConfig);

      // Merge with defaults to ensure all properties exist
      return {
        ...DEFAULT_BLUETOOTH_CONFIG,
        ...parsed,
      };
    }

    return DEFAULT_BLUETOOTH_CONFIG;
  } catch (error) {
    console.warn('[ScannerConfigService] Failed to load Bluetooth config, using defaults:', error);
    return DEFAULT_BLUETOOTH_CONFIG;
  }
};

/**
 * Save Bluetooth configuration
 */
export const saveBluetoothConfig = async (config: Partial<BluetoothConfig>): Promise<void> => {
  try {
    // Load current config first
    const currentConfig = await loadBluetoothConfig();

    // Merge with new values
    const updatedConfig = {
      ...currentConfig,
      ...config,
    };

    await AsyncStorage.setItem(BLUETOOTH_CONFIG_KEY, JSON.stringify(updatedConfig));
  } catch (error) {
    console.error('[ScannerConfigService] Failed to save Bluetooth config:', error);
    throw error;
  }
};

/**
 * ScannerConfigService - Public API
 */
export const ScannerConfigService = {
  load: loadScannerConfig,
  save: saveScannerConfig,
  reset: resetScannerConfig,
  KEY: SCANNER_CONFIG_KEY,
  // Bluetooth settings API
  loadPairedDevices: loadPairedBluetoothDevices,
  savePairedDevices: savePairedBluetoothDevices,
  loadLastConnectedDevice: loadLastConnectedDevice,
  saveLastConnectedDevice: saveLastConnectedDevice,
  clearLastConnectedDevice: clearLastConnectedDevice,
  loadBluetoothConfig: loadBluetoothConfig,
  saveBluetoothConfig: saveBluetoothConfig,
};
