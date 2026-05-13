/**
 * ScannerConfigService - Persistent storage for scanner settings
 * Handles loading and saving scanner configuration using AsyncStorage
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import { ScannerConfig, DEFAULT_SCANNER_CONFIG } from '../types/scanner.types';

// AsyncStorage key for scanner preferences
export const SCANNER_CONFIG_KEY = '@simpo_scanner_config';

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

/**
 * ScannerConfigService - Public API
 */
export const ScannerConfigService = {
  load: loadScannerConfig,
  save: saveScannerConfig,
  reset: resetScannerConfig,
  KEY: SCANNER_CONFIG_KEY,
};
