/**
 * Tests for useBarcodeScanner hook
 * Tests barcode scanner detection, debouncing, validation, and error handling
 */

import { renderHook, act } from '@testing-library/react-native';
import { useBarcodeScanner } from './useBarcodeScanner';
import { ScannerConfig, ScannerState } from '../types/scanner.types';

// Mock React Native Vibration API
jest.mock('react-native', () => ({
  Vibration: {
    vibrate: jest.fn(),
  },
}));

const { Vibration } = require('react-native');

// Mock AsyncStorage for config persistence
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
}));

describe('useBarcodeScanner', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  const mockOnBarcodeScanned = jest.fn();
  const mockOnError = jest.fn();
  const mockOnStateChange = jest.fn();

  const defaultCallbacks = {
    onBarcodeScanned: mockOnBarcodeScanned,
    onError: mockOnError,
    onStateChange: mockOnStateChange,
  };

  describe('scanner detection', () => {
    it('should detect fast input with Enter as scanner input', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      // Simulate fast scanner input (< 100ms total)
      await act(async () => {
        // Simulate character input with small delays (scanner-like)
        result.current.handleScannerInput('8', 0);
        result.current.handleScannerInput('9', 10);
        result.current.handleScannerInput('9', 20);
        result.current.handleScannerInput('1', 30);
        result.current.handleScannerInput('2', 40);
        result.current.handleScannerInput('3', 50);
        result.current.handleScannerInput('4', 60);
        result.current.handleScannerInput('5', 70);
        result.current.handleScannerInput('6', 80);
        result.current.handleScannerInput('7', 90);
        result.current.handleScannerInput('\n', 100); // Enter key

        // Fast-forward any timers
        jest.runAllTimers();
      });

      // Should detect as scanner and call onBarcodeScanned
      expect(mockOnBarcodeScanned).toHaveBeenCalledWith('8991234567');
      expect(result.current.state).toBe('success');
    });

    it('should treat slow input as manual typing', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      // Simulate slow manual typing (> 300ms total)
      await act(async () => {
        result.current.handleScannerInput('8', 0);
        result.current.handleScannerInput('9', 50);
        result.current.handleScannerInput('9', 100);
        result.current.handleScannerInput('1', 150);
        result.current.handleScannerInput('2', 200);
        result.current.handleScannerInput('3', 250);
        result.current.handleScannerInput('4', 400); // Total time > 300ms (gap before last char)
        result.current.handleScannerInput('\n', 450);

        jest.runAllTimers();
      });

      // Should NOT detect as scanner
      expect(mockOnBarcodeScanned).not.toHaveBeenCalled();
    });

    it('should treat input without Enter as manual typing', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      await act(async () => {
        // Fast input but NO Enter key
        result.current.handleScannerInput('8', 0);
        result.current.handleScannerInput('9', 10);
        result.current.handleScannerInput('9', 20);
        result.current.handleScannerInput('1', 30);

        jest.runAllTimers();
      });

      // Should NOT detect as scanner (no Enter key)
      expect(mockOnBarcodeScanned).not.toHaveBeenCalled();
    });
  });

  describe('debouncing', () => {
    it('should ignore duplicate scans within debounce period', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      await act(async () => {
        // First scan with fast timing
        for (let i = 0; i < 12; i++) {
          result.current.handleScannerInput('8', i * 5); // 5ms intervals = 55ms total
        }
        result.current.handleScannerInput('\n', 55);
        jest.advanceTimersByTime(100);

        // Second scan immediately (same barcode)
        for (let i = 0; i < 12; i++) {
          result.current.handleScannerInput('8', i * 5);
        }
        result.current.handleScannerInput('\n', 55);
        jest.advanceTimersByTime(100);
      });

      // Should only call once (debounced)
      expect(mockOnBarcodeScanned).toHaveBeenCalledTimes(1);
    });

    it('should allow same barcode after debounce period', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      await act(async () => {
        // First scan with fast timing
        for (let i = 0; i < 12; i++) {
          result.current.handleScannerInput('8', i * 5); // 5ms intervals = 55ms total
        }
        result.current.handleScannerInput('\n', 55);
        jest.advanceTimersByTime(100);

        // Wait for debounce period + 100ms
        jest.advanceTimersByTime(600);

        // Second scan (same barcode, after debounce)
        for (let i = 0; i < 12; i++) {
          result.current.handleScannerInput('8', i * 5);
        }
        result.current.handleScannerInput('\n', 55);
        jest.advanceTimersByTime(100);
      });

      // Should call twice (debounce period passed)
      expect(mockOnBarcodeScanned).toHaveBeenCalledTimes(2);
    });

    it('should process different barcodes rapidly', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      await act(async () => {
        // First scan with fast timing
        for (let i = 0; i < 8; i++) {
          result.current.handleScannerInput('8', i * 5); // 5ms intervals = 35ms total
        }
        result.current.handleScannerInput('\n', 35);
        jest.advanceTimersByTime(100);

        // Second scan (different barcode, immediately)
        for (let i = 0; i < 8; i++) {
          result.current.handleScannerInput('9', i * 5);
        }
        result.current.handleScannerInput('\n', 35);
        jest.advanceTimersByTime(100);
      });

      // Should call twice (different barcodes)
      expect(mockOnBarcodeScanned).toHaveBeenCalledTimes(2);
    });
  });

  describe('barcode validation', () => {
    it('should reject empty barcode', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      await act(async () => {
        result.current.handleScannerInput('   ', 0); // Whitespace only
        result.current.handleScannerInput('\n', 50);
        jest.runAllTimers();
      });

      expect(mockOnBarcodeScanned).not.toHaveBeenCalled();
      expect(mockOnError).toHaveBeenCalledWith({
        type: 'invalid_barcode',
        message: 'Barcode tidak boleh kosong',
      });
    });

    it('should reject barcode that is too short', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      await act(async () => {
        // Only 5 characters (min is 8)
        for (let i = 0; i < 5; i++) {
          result.current.handleScannerInput('8', i * 10);
        }
        result.current.handleScannerInput('\n', 50);
        jest.runAllTimers();
      });

      expect(mockOnBarcodeScanned).not.toHaveBeenCalled();
      expect(mockOnError).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'invalid_barcode',
          message: expect.stringContaining('minimal'),
        })
      );
    });

    it('should reject barcode that is too long', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      await act(async () => {
        // 15 characters (max is 13)
        for (let i = 0; i < 15; i++) {
          result.current.handleScannerInput('8', i * 5); // 5ms intervals = 70ms total
        }
        result.current.handleScannerInput('\n', 70); // Under 100ms threshold
        jest.advanceTimersByTime(100);
      });

      expect(mockOnBarcodeScanned).not.toHaveBeenCalled();
      expect(mockOnError).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'invalid_barcode',
          message: expect.stringContaining('maksimal'),
        })
      );
    });

    it('should trim whitespace from barcode', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      await act(async () => {
        // Barcode with leading/trailing spaces (ONLY at ends, not internal)
        // "  8991234567  " after trim becomes "8991234567"
        result.current.handleScannerInput(' ', 0);
        result.current.handleScannerInput(' ', 4);
        result.current.handleScannerInput('8', 8);
        result.current.handleScannerInput('9', 12);
        result.current.handleScannerInput('9', 16);
        result.current.handleScannerInput('1', 20);
        result.current.handleScannerInput('2', 24);
        result.current.handleScannerInput('3', 28);
        result.current.handleScannerInput('4', 32);
        result.current.handleScannerInput('5', 36);
        result.current.handleScannerInput('6', 40);
        result.current.handleScannerInput('7', 44);
        result.current.handleScannerInput(' ', 48);
        result.current.handleScannerInput(' ', 52);
        result.current.handleScannerInput('\n', 56); // Total time: 56ms (under 100ms)

        // Advance timers
        jest.advanceTimersByTime(100);
      });

      // Should trim and process '8991234567'
      expect(mockOnBarcodeScanned).toHaveBeenCalledWith('8991234567');
    });
  });

  describe('successful scan flow', () => {
    it('should transition states correctly during successful scan', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      expect(result.current.state).toBe('idle');

      // Complete scan with fast timing (< 100ms total)
      result.current.handleScannerInput('8', 0);

      for (let i = 1; i < 12; i++) {
        result.current.handleScannerInput('9', i * 5); // 5ms intervals = 55ms total
      }
      result.current.handleScannerInput('\n', 55); // Enter at 55ms (under 100ms threshold)

      // Wait for async operations (but not the idle timeout)
      await act(async () => {
        // Advance time just enough for the scan to process, not the full 1 second
        jest.advanceTimersByTime(100);
      });

      // Should be in success state
      expect(result.current.state).toBe('success');
      expect(mockOnBarcodeScanned).toHaveBeenCalled();
    });

    it('should trigger vibration feedback on success when enabled', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      await act(async () => {
        for (let i = 0; i < 12; i++) {
          result.current.handleScannerInput('8', i * 5); // 5ms intervals = 55ms total
        }
        result.current.handleScannerInput('\n', 55); // Enter at 55ms

        // Advance time just enough for the scan to process
        jest.advanceTimersByTime(100);
      });

      expect(Vibration.vibrate).toHaveBeenCalled();
    });

    it('should not trigger vibration when feedback disabled', async () => {
      const config: Partial<ScannerConfig> = { feedbackEnabled: false };
      const { result } = renderHook(() =>
        useBarcodeScanner({ ...defaultCallbacks, config })
      );

      for (let i = 0; i < 12; i++) {
        result.current.handleScannerInput('8', i * 10);
      }
      result.current.handleScannerInput('\n', 120);
      jest.runAllTimers();

      expect(Vibration.vibrate).not.toHaveBeenCalled();
    });
  });

  describe('error handling', () => {
    it('should handle callback errors gracefully', async () => {
      // Callback that throws an error
      const errorCallback = jest.fn().mockImplementation(() => {
        throw new Error('API Error');
      });

      const { result } = renderHook(() =>
        useBarcodeScanner({
          ...defaultCallbacks,
          onBarcodeScanned: errorCallback,
        })
      );

      await act(async () => {
        // Enter 10 characters to match expected barcode
        for (let i = 0; i < 10; i++) {
          result.current.handleScannerInput('8', i * 5); // 5ms intervals = 45ms total
        }
        result.current.handleScannerInput('\n', 45); // Enter at 45ms

        // Advance time just enough for the scan to process and trigger error
        jest.advanceTimersByTime(100);
      });

      // Should report error via onError callback
      expect(mockOnError).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'api_error',
          message: 'Gagal memproses barcode: API Error',
          barcode: '8888888888',
        })
      );

      expect(result.current.state).toBe('error');
    });

    it('should provide scanner error details', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      await act(async () => {
        // Send invalid (too short) barcode
        for (let i = 0; i < 4; i++) {
          result.current.handleScannerInput('8', i * 10);
        }
        result.current.handleScannerInput('\n', 40);
        jest.runAllTimers();
      });

      expect(mockOnError).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'invalid_barcode',
          barcode: '8888',
        })
      );
    });
  });

  describe('state management', () => {
    it('should call onStateChange when state changes', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      await act(async () => {
        result.current.handleScannerInput('8', 0);
        jest.runAllTimers();
      });

      expect(mockOnStateChange).toHaveBeenCalledWith('scanning');
    });

    it('should reset to idle after success', async () => {
      const { result } = renderHook(() => useBarcodeScanner(defaultCallbacks));

      // Complete scan with fast timing
      for (let i = 0; i < 12; i++) {
        result.current.handleScannerInput('8', i * 5); // 5ms intervals = 55ms total
      }
      result.current.handleScannerInput('\n', 55); // Enter at 55ms

      // Wait for success state (but don't advance full second)
      await act(async () => {
        jest.advanceTimersByTime(100);
      });

      // Check success state is set
      expect(result.current.state).toBe('success');

      // Now reset
      act(() => {
        result.current.reset();
      });

      expect(result.current.state).toBe('idle');
    });
  });

  describe('configuration', () => {
    it('should use custom debounce time from config', async () => {
      const customConfig: Partial<ScannerConfig> = { debounceMs: 1000 };
      const { result } = renderHook(() =>
        useBarcodeScanner({ ...defaultCallbacks, config: customConfig })
      );

      // First scan
      await act(async () => {
        for (let i = 0; i < 12; i++) {
          result.current.handleScannerInput('8', i * 5); // 5ms intervals = 55ms total
        }
        result.current.handleScannerInput('\n', 55); // Enter at 55ms

        // Advance time just enough for processing
        jest.advanceTimersByTime(100);
      });

      // Second scan within custom debounce period
      await act(async () => {
        for (let i = 0; i < 12; i++) {
          result.current.handleScannerInput('8', i * 5); // 5ms intervals = 55ms total
        }
        result.current.handleScannerInput('\n', 55); // Enter at 55ms

        // Advance time just enough for processing
        jest.advanceTimersByTime(100);
      });

      // Should still debounce with custom time (first scan should be called)
      expect(mockOnBarcodeScanned).toHaveBeenCalledTimes(1);
    });

    it('should use custom barcode length limits from config', async () => {
      const customConfig: Partial<ScannerConfig> = {
        minBarcodeLength: 5,
        maxBarcodeLength: 20,
      };
      const { result } = renderHook(() =>
        useBarcodeScanner({ ...defaultCallbacks, config: customConfig })
      );

      await act(async () => {
        // 6 characters (would be rejected with default min of 8)
        for (let i = 0; i < 6; i++) {
          result.current.handleScannerInput('8', i * 5); // 5ms intervals = 25ms total
        }
        result.current.handleScannerInput('\n', 25); // Enter at 25ms

        // Advance time just enough for processing
        jest.advanceTimersByTime(100);
      });

      // Should accept with custom config
      expect(mockOnBarcodeScanned).toHaveBeenCalled();
    });
  });
});
