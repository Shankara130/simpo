/**
 * useKeyboardInput Hook
 * Captures keyboard input from USB HID barcode scanners
 * USB scanners appear as keyboard devices sending keystrokes + Enter
 */

import { useRef, useCallback, useEffect } from 'react';
import { TextInput } from 'react-native';

export interface UseKeyboardInputProps {
  /** Callback when character is received */
  onCharReceived: (char: string, timestamp: number) => void;
  /** Whether input capture is active */
  enabled?: boolean;
}

export interface UseKeyboardInputReturn {
  /** Ref to attach to TextInput for keyboard capture */
  textInputRef: React.RefObject<TextInput>;
  /** Enable/disable input capture */
  setEnabled: (enabled: boolean) => void;
  /**
   * Handler for text changes (character by character)
   * This should be attached to onChangeText prop of TextInput
   */
  handleChange: (text: string) => void;
  /**
   * Handler for submit editing (Enter key)
   * This should be attached to onSubmitEditing prop of TextInput
   */
  handleSubmit: () => void;
}

/**
 * Keyboard input capture for USB HID barcode scanners
 *
 * How USB scanners work:
 * 1. Scanner appears as keyboard device to OS
 * 2. Each barcode character sent as keystroke
 * 3. Enter key (carriage return) sent after last character
 * 4. Typical scan time: <100ms total
 *
 * Implementation:
 * - Use invisible TextInput to capture keyboard events
 * - Track timing of each character
 * - Forward to useBarcodeScanner for processing
 */
export const useKeyboardInput = (props: UseKeyboardInputProps): UseKeyboardInputReturn => {
  const { onCharReceived, enabled = true } = props;

  const textInputRef = useRef<TextInput>(null);
  const enabledRef = useRef(enabled);

  // Update enabled ref without triggering re-render
  const setEnabled = useCallback((value: boolean) => {
    enabledRef.current = value;
  }, []);

  // Handle text change (character by character)
  const handleChange = useCallback((text: string) => {
    if (!enabledRef.current) return;

    const timestamp = Date.now();

    // Get the last character (new character added)
    const char = text.slice(-1);

    // Forward to scanner processing
    if (char) {
      onCharReceived(char, timestamp);
    }

    // Clear input to keep buffer empty
    // This prevents accumulation of characters
    if (textInputRef.current) {
      textInputRef.current.setNativeProps({ text: '' });
    }
  }, [onCharReceived]);

  // Handle submit (Enter key)
  const handleSubmit = useCallback(() => {
    if (!enabledRef.current) return;

    const timestamp = Date.now();

    // Send Enter key to scanner processing
    onCharReceived('\n', timestamp);

    // Clear input
    if (textInputRef.current) {
      textInputRef.current.setNativeProps({ text: '' });
    }
  }, [onCharReceived]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      // Clear any pending input
      if (textInputRef.current) {
        textInputRef.current.setNativeProps({ text: '' });
      }
    };
  }, []);

  return {
    textInputRef,
    setEnabled,
    handleChange,
    handleSubmit,
  };
};
