/**
 * useKeyboardInput Hook Tests
 * Tests USB HID keyboard input capture for barcode scanners
 */

import { renderHook, act } from '@testing-library/react-native';
import { useKeyboardInput, UseKeyboardInputProps } from './useKeyboardInput';

describe('useKeyboardInput', () => {
  let mockOnCharReceived: jest.Mock;
  let defaultProps: UseKeyboardInputProps;

  beforeEach(() => {
    mockOnCharReceived = jest.fn();
    defaultProps = {
      onCharReceived: mockOnCharReceived,
      enabled: true,
    };
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.clearAllMocks();
    jest.useRealTimers();
  });

  describe('initialization', () => {
    it('should provide textInputRef', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      expect(result.current.textInputRef).toBeDefined();
      expect(result.current.textInputRef.current).toBeNull();
    });

    it('should provide setEnabled function', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      expect(result.current.setEnabled).toBeDefined();
      expect(typeof result.current.setEnabled).toBe('function');
    });

    it('should provide handleChange function', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      expect(result.current.handleChange).toBeDefined();
      expect(typeof result.current.handleChange).toBe('function');
    });

    it('should provide handleSubmit function', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      expect(result.current.handleSubmit).toBeDefined();
      expect(typeof result.current.handleSubmit).toBe('function');
    });
  });

  describe('character capture', () => {
    it('should forward last character with timestamp on text change', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      // Create mock text input ref
      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      // Simulate text change with multiple characters
      act(() => {
        result.current.handleChange('ABC');
      });

      // Should forward only the last character 'C'
      expect(mockOnCharReceived).toHaveBeenCalledTimes(1);
      expect(mockOnCharReceived).toHaveBeenCalledWith('C', expect.any(Number));
      expect(mockSetNativeProps).toHaveBeenCalledWith({ text: '' });
    });

    it('should handle single character input', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      act(() => {
        result.current.handleChange('A');
      });

      expect(mockOnCharReceived).toHaveBeenCalledWith('A', expect.any(Number));
      expect(mockSetNativeProps).toHaveBeenCalledWith({ text: '' });
    });

    it('should extract correct last character from various inputs', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      // Test various inputs
      const testCases = [
        { input: '123', expectedChar: '3' },
        { input: 'ABC', expectedChar: 'C' },
        { input: 'XYZ', expectedChar: 'Z' },
        { input: '9', expectedChar: '9' },
      ];

      testCases.forEach(({ input, expectedChar }) => {
        mockOnCharReceived.mockClear();

        act(() => {
          result.current.handleChange(input);
        });

        expect(mockOnCharReceived).toHaveBeenCalledWith(expectedChar, expect.any(Number));
      });
    });

    it('should clear input buffer after character processing', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      act(() => {
        result.current.handleChange('A');
      });

      // Input should be cleared after each character
      expect(mockSetNativeProps).toHaveBeenCalledWith({ text: '' });
      expect(mockSetNativeProps).toHaveBeenCalledTimes(1);
    });
  });

  describe('Enter key detection', () => {
    it('should forward Enter key as \\n character with timestamp', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      act(() => {
        result.current.handleSubmit();
      });

      expect(mockOnCharReceived).toHaveBeenCalledWith('\n', expect.any(Number));
      expect(mockSetNativeProps).toHaveBeenCalledWith({ text: '' });
    });

    it('should clear input after Enter key', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      act(() => {
        result.current.handleSubmit();
      });

      expect(mockSetNativeProps).toHaveBeenCalledWith({ text: '' });
    });
  });

  describe('enabled state', () => {
    it('should forward characters when enabled', () => {
      const { result } = renderHook(() => useKeyboardInput({
        onCharReceived: mockOnCharReceived,
        enabled: true,
      }));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      act(() => {
        result.current.handleChange('A');
      });

      expect(mockOnCharReceived).toHaveBeenCalled();
    });

    it('should not forward characters when disabled initially', () => {
      const { result } = renderHook(() => useKeyboardInput({
        onCharReceived: mockOnCharReceived,
        enabled: false,
      }));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      act(() => {
        result.current.handleChange('A');
      });

      expect(mockOnCharReceived).not.toHaveBeenCalled();
      expect(mockSetNativeProps).not.toHaveBeenCalled();
    });

    it('should allow toggling enabled state', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      // Initially enabled - should forward
      act(() => {
        result.current.handleChange('A');
      });
      expect(mockOnCharReceived).toHaveBeenCalledTimes(1);

      // Disable
      act(() => {
        result.current.setEnabled(false);
      });

      // Should not forward when disabled
      mockOnCharReceived.mockClear();
      act(() => {
        result.current.handleChange('B');
      });
      expect(mockOnCharReceived).not.toHaveBeenCalled();

      // Re-enable
      act(() => {
        result.current.setEnabled(true);
      });

      // Should forward again
      act(() => {
        result.current.handleChange('C');
      });
      expect(mockOnCharReceived).toHaveBeenCalledWith('C', expect.any(Number));
    });

    it('should not forward Enter key when disabled', () => {
      const { result } = renderHook(() => useKeyboardInput({
        onCharReceived: mockOnCharReceived,
        enabled: false,
      }));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      act(() => {
        result.current.handleSubmit();
      });

      expect(mockOnCharReceived).not.toHaveBeenCalled();
      expect(mockSetNativeProps).not.toHaveBeenCalled();
    });
  });

  describe('error handling', () => {
    it('should handle empty text input gracefully', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      // Empty string should not trigger callback
      act(() => {
        result.current.handleChange('');
      });

      expect(mockOnCharReceived).not.toHaveBeenCalled();
      expect(mockSetNativeProps).toHaveBeenCalledWith({ text: '' });
    });

    it('should handle missing textInputRef gracefully', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      // Don't set the ref - should not crash
      act(() => {
        result.current.handleChange('A');
      });

      // Should complete without error (callback might still be called)
      expect(() => act(() => result.current.handleChange('A'))).not.toThrow();
    });
  });

  describe('timing', () => {
    it('should provide timestamp for each character', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      const beforeTime = Date.now();

      act(() => {
        result.current.handleChange('A');
      });

      const afterTime = Date.now();

      expect(mockOnCharReceived).toHaveBeenCalledWith(
        'A',
        expect.any(Number)
      );

      const callArgs = mockOnCharReceived.mock.calls[0];
      const timestamp = callArgs[1];

      // Timestamp should be between before and after time
      expect(timestamp).toBeGreaterThanOrEqual(beforeTime);
      expect(timestamp).toBeLessThanOrEqual(afterTime);
    });

    it('should provide unique timestamps for consecutive calls', () => {
      const { result } = renderHook(() => useKeyboardInput(defaultProps));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      act(() => {
        result.current.handleChange('A');
      });

      const firstTimestamp = mockOnCharReceived.mock.calls[0][1];

      // Advance time slightly
      jest.advanceTimersByTime(10);

      act(() => {
        result.current.handleChange('B');
      });

      const secondTimestamp = mockOnCharReceived.mock.calls[1][1];

      // Timestamps should be different
      expect(secondTimestamp).toBeGreaterThan(firstTimestamp);
    });
  });

  describe('cleanup', () => {
    it('should cleanup on unmount', () => {
      const { result, unmount } = renderHook(() => useKeyboardInput(defaultProps));

      const mockSetNativeProps = jest.fn();
      result.current.textInputRef.current = {
        setNativeProps: mockSetNativeProps,
      } as any;

      // Set some content
      act(() => {
        result.current.handleChange('A');
      });

      // Unmount should not crash
      expect(() => unmount()).not.toThrow();
    });
  });
});
