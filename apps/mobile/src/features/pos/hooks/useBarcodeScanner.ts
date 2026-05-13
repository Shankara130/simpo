/**
 * useBarcodeScanner Hook
 * Handles barcode scanner input detection, debouncing, validation, and feedback
 * Distinguishes between scanner input (fast, consistent timing) and manual typing
 */

import { useState, useCallback, useRef, useEffect } from 'react';
import { Vibration } from 'react-native';
import {
  ScannerState,
  ScannerConfig,
  ScannerInputState,
  ScannerCallbacks,
  DEFAULT_SCANNER_CONFIG,
  ScannerError,
  ScannerErrorType,
} from '../types/scanner.types';

export interface UseBarcodeScannerProps extends Partial<ScannerCallbacks> {
  /** Custom scanner configuration (uses defaults if not provided) */
  config?: Partial<ScannerConfig>;
}

export interface UseBarcodeScannerReturn {
  /** Current scanner state */
  state: ScannerState;
  /** Handle character input from scanner/keyboard */
  handleScannerInput: (char: string, timestamp: number) => void;
  /** Reset scanner to idle state */
  reset: () => void;
}

/**
 * Custom hook for barcode scanner functionality
 *
 * Detects scanner input by analyzing timing characteristics:
 * - Scanner: Fast input (< 100ms total) ending with Enter key
 * - Keyboard: Slower input or no Enter key
 *
 * Features:
 * - Debouncing to prevent duplicate scans (default: 500ms)
 * - Barcode validation (length, characters)
 * - Haptic feedback on success/error
 * - State management with callbacks
 */
export const useBarcodeScanner = (props: UseBarcodeScannerProps = {}): UseBarcodeScannerReturn => {
  const {
    onBarcodeScanned,
    onError,
    onStateChange,
    config: userConfig,
  } = props;

  // Merge user config with defaults
  const config: ScannerConfig = {
    ...DEFAULT_SCANNER_CONFIG,
    ...userConfig,
  };

  // Scanner state
  const [state, setState] = useState<ScannerState>('idle');

  // Input tracking for scanner detection
  const inputState = useRef<ScannerInputState>({
    inputBuffer: '',
    firstCharTime: null,
    lastCharTime: null,
    isScanning: false,
  });

  // Debounce tracking
  const lastScanRef = useRef<{ barcode: string; time: number }>({
    barcode: '',
    time: 0,
  });

  // Update state with callback notification
  const updateState = useCallback((newState: ScannerState) => {
    setState(newState);
    onStateChange?.(newState);
  }, [onStateChange]);

  // Validate barcode format
  const validateBarcode = useCallback((barcode: string): { valid: boolean; error?: ScannerError } => {
    const trimmed = barcode.trim();

    // Check empty
    if (!trimmed) {
      return {
        valid: false,
        error: {
          type: 'invalid_barcode',
          message: 'Barcode tidak boleh kosong',
        },
      };
    }

    // Check length
    if (trimmed.length < config.minBarcodeLength) {
      return {
        valid: false,
        error: {
          type: 'invalid_barcode',
          message: `Barcode terlalu pendek (minimal ${config.minBarcodeLength} karakter)`,
          barcode: trimmed,
        },
      };
    }

    if (trimmed.length > config.maxBarcodeLength) {
      return {
        valid: false,
        error: {
          type: 'invalid_barcode',
          message: `Barcode terlalu panjang (maksimal ${config.maxBarcodeLength} karakter)`,
          barcode: trimmed,
        },
      };
    }

    return { valid: true };
  }, [config.minBarcodeLength, config.maxBarcodeLength]);

  // Check if barcode is debounced (duplicate within cooldown period)
  const canProcessBarcode = useCallback((barcode: string): boolean => {
    const now = Date.now();
    const { barcode: lastBarcode, time: lastTime } = lastScanRef.current;

    if (barcode === lastBarcode && (now - lastTime) < config.debounceMs) {
      return false; // Skip duplicate scan
    }

    lastScanRef.current = { barcode, time: now };
    return true;
  }, [config.debounceMs]);

  // Trigger haptic feedback
  const triggerFeedback = useCallback((success: boolean) => {
    if (!config.feedbackEnabled) return;

    // Vibrate once for success, twice for error
    if (success) {
      Vibration.vibrate(50); // Short vibration for success
    } else {
      Vibration.vibrate([50, 100, 50]); // Double vibration for error
    }
  }, [config.feedbackEnabled]);

  // Report error
  const reportError = useCallback((error: ScannerError) => {
    updateState('error');
    triggerFeedback(false);
    onError?.(error);
  }, [onError, updateState, triggerFeedback]);

  // Process completed barcode scan
  const processBarcode = useCallback(async (barcode: string) => {
    // Trim barcode first
    const trimmedBarcode = barcode.trim();

    // Validate
    const validation = validateBarcode(trimmedBarcode);
    if (!validation.valid) {
      reportError(validation.error!);
      return;
    }

    // Check debounce
    if (!canProcessBarcode(trimmedBarcode)) {
      // Silently skip duplicate scan
      return;
    }

    // Process the scan
    updateState('loading');

    try {
      await onBarcodeScanned?.(trimmedBarcode);
      // Set success state first
      setState('success'); // Direct setState to avoid onStateChange callback
      // Then trigger feedback
      triggerFeedback(true);
    } catch (error) {
      reportError({
        type: 'api_error',
        message: `Gagal memproses barcode: ${error instanceof Error ? error.message : 'Unknown error'}`,
        barcode: trimmedBarcode,
        originalError: error,
      });
      return; // Don't reset to idle if error occurred
    }

    // Reset to idle after short delay
    setTimeout(() => {
      setState('idle'); // Direct setState to avoid triggering onStateChange
    }, 1000);
  }, [validateBarcode, canProcessBarcode, onBarcodeScanned, updateState, triggerFeedback, reportError]);

  // Handle character input
  const handleScannerInput = useCallback((char: string, timestamp: number) => {
    const current = inputState.current;

    // Initialize on first character
    if (!current.isScanning && char !== '\n') {
      current.inputBuffer = '';
      current.firstCharTime = timestamp;
      current.isScanning = true;
      updateState('scanning');
    }

    // Append character (except Enter key)
    if (char !== '\n') {
      current.inputBuffer += char;
      current.lastCharTime = timestamp;
      return;
    }

    // Enter key received - check if this is scanner input
    if (char === '\n' && current.isScanning) {
      const totalTime = (current.lastCharTime || 0) - (current.firstCharTime || 0);

      // Check timing: scanner input should be fast (< maxScanTimeMs)
      if (totalTime <= config.maxScanTimeMs && current.inputBuffer.length > 0) {
        // Scanner detected - process barcode
        const barcode = current.inputBuffer;
        processBarcode(barcode);
      } else {
        // Manual typing detected - ignore
        updateState('idle');
      }

      // Reset input state
      current.inputBuffer = '';
      current.firstCharTime = null;
      current.lastCharTime = null;
      current.isScanning = false;
    }
  }, [config.maxScanTimeMs, processBarcode, updateState]);

  // Reset scanner state
  const reset = useCallback(() => {
    inputState.current = {
      inputBuffer: '',
      firstCharTime: null,
      lastCharTime: null,
      isScanning: false,
    };
    updateState('idle');
  }, [updateState]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      reset();
    };
  }, [reset]);

  return {
    state,
    handleScannerInput,
    reset,
  };
};
