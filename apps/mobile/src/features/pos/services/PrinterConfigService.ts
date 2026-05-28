/**
 * PrinterConfigService - Persistent storage for printer and drawer settings
 * Handles loading and saving printer configuration using AsyncStorage
 * Story 7.4: Extended with cash drawer configuration support
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import {
  CashDrawerConfig,
  DEFAULT_CASH_DRAWER_CONFIG,
} from '../hardware/printer';

// Re-export DEFAULT_CASH_DRAWER_CONFIG for tests
export { DEFAULT_CASH_DRAWER_CONFIG };

// AsyncStorage keys for printer preferences
export const PRINTER_CONFIG_KEY = '@simpo_printer_config';
export const DRAWER_CONFIG_KEY = '@simpo_drawer_config';

/**
 * Printer Configuration Interface
 */
export interface PrinterConfig {
  defaultPrinterId?: string;
  paperWidth?: 58 | 80;
  autoCut?: boolean;
}

/**
 * Default printer configuration
 */
export const DEFAULT_PRINTER_CONFIG: PrinterConfig = {
  paperWidth: 58,
  autoCut: true,
};

// ============================================================================
// Printer Settings Persistence
// ============================================================================

/**
 * Load printer configuration from AsyncStorage
 * Returns default config if no saved config exists or on error
 * Tracks load success for telemetry purposes
 */
export const loadPrinterConfig = async (): Promise<PrinterConfig> => {
  try {
    const savedConfig = await AsyncStorage.getItem(PRINTER_CONFIG_KEY);

    if (savedConfig) {
      try {
        const parsed = JSON.parse(savedConfig);

        // Merge with defaults to ensure all properties exist
        const mergedConfig = {
          ...DEFAULT_PRINTER_CONFIG,
          ...parsed,
        };

        // Track successful config load for monitoring
        console.log('[PrinterConfigService] Config loaded successfully');
        return mergedConfig;
      } catch (parseError) {
        // JSON parsing failed - config might be corrupted
        console.error('[PrinterConfigService] Failed to parse saved config, using defaults:', parseError);
        console.warn('[PrinterConfigService] Consider clearing app data if this persists');
        return DEFAULT_PRINTER_CONFIG;
      }
    }

    // No saved config, return defaults
    console.log('[PrinterConfigService] No saved config found, using defaults');
    return DEFAULT_PRINTER_CONFIG;
  } catch (error) {
    // AsyncStorage access failed (permissions, storage full, etc.)
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('[PrinterConfigService] Failed to access AsyncStorage:', errorMessage);
    console.warn('[PrinterConfigService] Using default config - check app permissions and storage availability');
    return DEFAULT_PRINTER_CONFIG;
  }
};

/**
 * Save printer configuration to AsyncStorage
 */
export const savePrinterConfig = async (config: Partial<PrinterConfig>): Promise<void> => {
  try {
    // Load current config first
    const currentConfig = await loadPrinterConfig();

    // Merge with new values
    const updatedConfig = {
      ...currentConfig,
      ...config,
    };

    await AsyncStorage.setItem(PRINTER_CONFIG_KEY, JSON.stringify(updatedConfig));
  } catch (error) {
    console.error('[PrinterConfigService] Failed to save config:', error);
    throw error;
  }
};

/**
 * Reset printer configuration to defaults
 */
export const resetPrinterConfig = async (): Promise<void> => {
  try {
    await AsyncStorage.removeItem(PRINTER_CONFIG_KEY);
  } catch (error) {
    console.error('[PrinterConfigService] Failed to reset config:', error);
    throw error;
  }
};

// ============================================================================
// Cash Drawer Settings Persistence (Story 7.4)
// ============================================================================

/**
 * Load cash drawer configuration from AsyncStorage
 * Returns default config if no saved config exists
 */
export const loadDrawerConfig = async (): Promise<CashDrawerConfig> => {
  try {
    const savedConfig = await AsyncStorage.getItem(DRAWER_CONFIG_KEY);

    if (savedConfig) {
      const parsed = JSON.parse(savedConfig);

      // Merge with defaults to ensure all properties exist
      return {
        ...DEFAULT_CASH_DRAWER_CONFIG,
        ...parsed,
      };
    }

    // No saved config, return defaults
    return DEFAULT_CASH_DRAWER_CONFIG;
  } catch (error) {
    console.warn('[PrinterConfigService] Failed to load drawer config, using defaults:', error);
    return DEFAULT_CASH_DRAWER_CONFIG;
  }
};

/**
 * Save cash drawer configuration to AsyncStorage
 */
export const saveDrawerConfig = async (config: Partial<CashDrawerConfig>): Promise<void> => {
  try {
    // Load current config first
    const currentConfig = await loadDrawerConfig();

    // Merge with new values
    const updatedConfig = {
      ...currentConfig,
      ...config,
    };

    await AsyncStorage.setItem(DRAWER_CONFIG_KEY, JSON.stringify(updatedConfig));
  } catch (error) {
    console.error('[PrinterConfigService] Failed to save drawer config:', error);
    throw error;
  }
};

/**
 * Reset cash drawer configuration to defaults
 */
export const resetDrawerConfig = async (): Promise<void> => {
  try {
    await AsyncStorage.removeItem(DRAWER_CONFIG_KEY);
  } catch (error) {
    console.error('[PrinterConfigService] Failed to reset drawer config:', error);
    throw error;
  }
};

/**
 * PrinterConfigService - Public API
 */
export const PrinterConfigService = {
  // Printer settings API
  loadPrinter: loadPrinterConfig,
  savePrinter: savePrinterConfig,
  resetPrinter: resetPrinterConfig,
  PRINTER_KEY: PRINTER_CONFIG_KEY,
  // Drawer settings API
  loadDrawer: loadDrawerConfig,
  saveDrawer: saveDrawerConfig,
  resetDrawer: resetDrawerConfig,
  DRAWER_KEY: DRAWER_CONFIG_KEY,
};
