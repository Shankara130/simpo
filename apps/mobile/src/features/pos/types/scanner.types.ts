/**
 * Scanner type definitions for POS feature
 * Defines barcode scanner state, configuration, and events
 */

/**
 * Scanner state represents current status of barcode scanner
 */
export type ScannerState = 'idle' | 'scanning' | 'success' | 'error' | 'loading';

/**
 * Scanner configuration options
 */
export interface ScannerConfig {
  /** Debounce time in milliseconds to prevent duplicate scans (default: 500ms) */
  debounceMs: number;
  /** Maximum time for complete scan in milliseconds (default: 100ms) */
  maxScanTimeMs: number;
  /** Minimum barcode length (default: 8 characters) */
  minBarcodeLength: number;
  /** Maximum barcode length (default: 13 characters for EAN-13) */
  maxBarcodeLength: number;
  /** Whether to enable sound/vibration feedback (default: true) */
  feedbackEnabled: boolean;
}

/**
 * Default scanner configuration
 */
export const DEFAULT_SCANNER_CONFIG: ScannerConfig = {
  debounceMs: 500,
  maxScanTimeMs: 100,
  minBarcodeLength: 8,
  maxBarcodeLength: 13,
  feedbackEnabled: true,
};

/**
 * Scanner input tracking state for detecting scanner vs keyboard input
 */
export interface ScannerInputState {
  /** Buffer of characters received */
  inputBuffer: string;
  /** Timestamp of first character in current input */
  firstCharTime: number | null;
  /** Timestamp of last character received */
  lastCharTime: number | null;
  /** Whether currently processing a scan */
  isScanning: boolean;
}

/**
 * Scanner result from successful scan
 */
export interface ScannerResult {
  /** The scanned barcode value */
  barcode: string;
  /** Timestamp when scan was completed */
  timestamp: number;
}

/**
 * Scanner error types
 */
export type ScannerErrorType = 'invalid_barcode' | 'product_not_found' | 'api_error' | 'unknown';

/**
 * Scanner error details
 */
export interface ScannerError {
  type: ScannerErrorType;
  message: string;
  barcode?: string;
  originalError?: unknown;
}

/**
 * Scanner callbacks for hook consumers
 */
export interface ScannerCallbacks {
  /** Called when a valid barcode is scanned successfully */
  onBarcodeScanned: (barcode: string) => void | Promise<void>;
  /** Called when an error occurs during scanning */
  onError?: (error: ScannerError) => void;
  /** Called when scanner state changes */
  onStateChange?: (state: ScannerState) => void;
}
