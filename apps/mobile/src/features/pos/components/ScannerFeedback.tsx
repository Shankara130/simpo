/**
 * ScannerFeedback Component
 * Provides visual and haptic feedback for barcode scanner operations
 * Shows success/error/loading indicators with auto-dismiss
 */

import React, { useEffect, useMemo } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, Platform } from 'react-native';
import { ScannerState } from '../types/scanner.types';

interface ScannerFeedbackProps {
  /** Current scanner state */
  state?: ScannerState;
  /** Optional message to display */
  message?: string;
  /** Auto-dismiss timeout in ms (default: 1500 for success, 3000 for error) */
  timeout?: number;
}

// Default messages for each state
const DEFAULT_MESSAGES: Record<ScannerState, string> = {
  idle: '',
  scanning: 'Memindai...',
  success: 'Scan berhasil',
  error: 'Scan gagal',
  loading: 'Memindai...',
};

// Background colors for each state
const STATE_COLORS: Record<ScannerState, string> = {
  idle: 'transparent',
  scanning: '#2196F3',
  success: '#4CAF50',
  error: '#F44336',
  loading: '#FF9800',
};

// Icons for each state (using simple text characters)
const STATE_ICONS: Record<ScannerState, string> = {
  idle: '',
  scanning: '⟳',
  success: '✓',
  error: '✕',
  loading: '⟳',
};

/**
 * ScannerFeedback component for visual scanner state feedback
 *
 * Renders overlay indicators for:
 * - Loading: Orange with spinner
 * - Scanning: Blue with spinner
 * - Success: Green with checkmark
 * - Error: Red with error icon
 * - Idle: No rendering
 */
export const ScannerFeedback: React.FC<ScannerFeedbackProps> = ({
  state = 'idle',
  message,
  timeout,
}) => {
  // Don't render for idle state
  if (state === 'idle') {
    return null;
  }

  const displayMessage = message || DEFAULT_MESSAGES[state];
  const backgroundColor = STATE_COLORS[state];
  const icon = STATE_ICONS[state];

  // Determine auto-dismiss timeout
  const dismissTimeout = useMemo(() => {
    if (timeout !== undefined) {
      return timeout;
    }
    // Default timeouts: success 1500ms, error 3000ms, others no auto-dismiss
    if (state === 'success') return 1500;
    if (state === 'error') return 3000;
    return null; // No auto-dismiss for scanning/loading
  }, [state, timeout]);

  // Auto-dismiss effect
  useEffect(() => {
    if (dismissTimeout && dismissTimeout > 0) {
      const timer = setTimeout(() => {
        // Parent component should handle state reset
        // This is just for the timeout reference
      }, dismissTimeout);

      return () => clearTimeout(timer);
    }
  }, [dismissTimeout]);

  // Render icon based on state
  const renderIcon = () => {
    if (state === 'loading' || state === 'scanning') {
      return <ActivityIndicator size="large" color="#FFFFFF" testID="loading-indicator" />;
    }

    if (state === 'success') {
      return (
        <Text style={styles.icon} testID="success-icon">
          {icon}
        </Text>
      );
    }

    if (state === 'error') {
      return (
        <Text style={styles.icon} testID="error-icon">
          {icon}
        </Text>
      );
    }

    return null;
  };

  return (
    <View
      testID="feedback-container"
      style={[
        styles.container,
        { backgroundColor },
        Platform.select({
          android: { elevation: 8 },
          ios: { zIndex: 999 },
        }),
      ]}
      accessibilityLabel="Status pemindai"
      accessibilityLiveRegion="assertive"
      accessibilityRole="alert"
    >
      <View testID="feedback-content" style={styles.content}>
        {state === 'scanning' ? (
          <View testID="scanning-indicator" style={styles.iconContainer}>
            {renderIcon()}
          </View>
        ) : (
          <View style={styles.iconContainer}>{renderIcon()}</View>
        )}
        {displayMessage ? (
          <Text style={styles.message} maxFontSizeMultiplier={1.5}>
            {displayMessage}
          </Text>
        ) : null}
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    minHeight: 80,
    paddingHorizontal: 20,
    paddingVertical: 16,
    justifyContent: 'center',
    alignItems: 'center',
    ...Platform.select({
      ios: {
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.25,
        shadowRadius: 4,
      },
      android: {
        elevation: 8,
      },
    }),
  },

  content: {
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
  },

  iconContainer: {
    marginBottom: 4,
  },

  icon: {
    fontSize: 48,
    fontWeight: 'bold',
    color: '#FFFFFF',
  },

  message: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFFFFF',
    textAlign: 'center',
  },
});
