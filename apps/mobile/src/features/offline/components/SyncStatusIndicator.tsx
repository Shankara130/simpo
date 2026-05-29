/**
 * Sync Status Indicator Component
 * Story 8.4: Implement Visual Sync Status Indicators
 *
 * Visual indicator showing sync status in app header
 * Displays icon based on sync state: synced (green checkmark), pending (yellow clock), failed (red exclamation)
 * Tappable to show sync details modal
 * AC1: Visual Status Indicators in App Header
 */

import React, { useState, useEffect } from 'react';
import { TouchableOpacity, Animated, View, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { useNetworkStatus } from '../hooks/useNetworkStatus';
import { useBidirectionalSync } from '../hooks/useBidirectionalSync';
import { useSyncProgress } from '../hooks/useSyncProgress';
import { formatSyncProgress } from '../constants/syncMessages';

/**
 * Sync status indicator types
 */
type SyncIndicatorState = 'synced' | 'pending' | 'failed';

/**
 * Display state properties
 */
interface DisplayState {
  state: SyncIndicatorState;
  icon: keyof typeof Ionicons.glyphMap;
  color: string;
  message: string;
}

/**
 * Sync Status Indicator Props
 */
interface SyncStatusIndicatorProps {
  position?: 'header-left' | 'header-right';
  onShowDetails?: () => void;
  size?: number;
}

/**
 * Get display state from combined sync and network status
 * Maps sync state + network status to visual indicator
 */
function getDisplayState(
  syncStatus: 'idle' | 'syncing' | 'synced' | 'failed',
  isConnected: boolean,
  pendingCount: number,
  hasError?: string
): DisplayState {
  // Offline with pending transactions
  if (!isConnected && pendingCount > 0) {
    return {
      state: 'pending',
      icon: 'time-outline',
      color: '#F59E0B', // yellow/amber
      message: 'Menunggu koneksi internet...',
    };
  }

  // Online with pending transactions
  if (isConnected && pendingCount > 0 && syncStatus === 'syncing') {
    return {
      state: 'pending',
      icon: 'time-outline',
      color: '#F59E0B', // yellow/amber
      message: 'Sync dalam proses...',
    };
  }

  // Sync complete
  if (pendingCount === 0 && syncStatus === 'synced') {
    return {
      state: 'synced',
      icon: 'checkmark-circle-outline',
      color: '#10B981', // green
      message: 'Semua data ter-sync',
    };
  }

  // Sync failed
  if (syncStatus === 'failed' || hasError) {
    return {
      state: 'failed',
      icon: 'alert-circle-outline',
      color: '#EF4444', // red
      message: 'Gagal sync',
    };
  }

  // Default to pending/idle
  return {
    state: 'pending',
    icon: 'time-outline',
    color: '#9CA3AF', // gray
    message: 'Menunggu sync...',
  };
}

/**
 * Sync Status Indicator Component
 *
 * Visual icon indicator showing sync status in header
 * Supports tappable gesture to show details modal
 * Animated state transitions with fade and scale
 * Haptic feedback on failed state
 */
export function SyncStatusIndicator({
  position = 'header-right',
  onShowDetails,
  size = 24,
}: SyncStatusIndicatorProps) {
  const { isConnected } = useNetworkStatus();
  const { status: syncStatus, pendingCount, error } = useBidirectionalSync();
  const { pendingCount: uploadPendingCount } = useSyncProgress();

  const combinedPendingCount = Math.max(pendingCount, uploadPendingCount);

  const displayState = getDisplayState(
    syncStatus,
    isConnected,
    combinedPendingCount,
    error
  );

  const [previousState, setPreviousState] = useState<SyncIndicatorState>(displayState.state);
  const [fadeAnim] = useState(new Animated.Value(1));
  const [scaleAnim] = useState(new Animated.Value(1));

  // Animate state transitions
  useEffect(() => {
    if (previousState !== displayState.state) {
      // Fade transition
      fadeAnim.setValue(0);
      Animated.timing(fadeAnim, {
        toValue: 1,
        duration: 200,
        useNativeDriver: true,
      }).start();

      // Scale transition
      scaleAnim.setValue(0.8);
      Animated.spring(scaleAnim, {
        toValue: 1,
        tension: 50,
        friction: 7,
        useNativeDriver: true,
      }).start();

      // Haptic feedback for failed state
      if (displayState.state === 'failed') {
        Haptics.notificationAsync(Haptics.NotificationFeedbackType.Warning)
          .catch(() => {
            // Ignore haptic errors (not all devices support haptics)
          });
      }

      setPreviousState(displayState.state);
    }
  }, [displayState.state, previousState]);

  const handlePress = () => {
    // Provide haptic feedback on tap
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light).catch(() => {
      // Ignore haptic errors
    });

    // Call parent's onShowDetails callback if provided
    if (onShowDetails) {
      onShowDetails();
    }
  };

  const animatedStyle = {
    opacity: fadeAnim,
    transform: [
      {
        scale: scaleAnim,
      },
    ],
  };

  const containerStyle = position === 'header-right'
    ? styles.headerRight
    : styles.headerLeft;

  return (
    <Animated.View style={[containerStyle, animatedStyle]}>
      <TouchableOpacity
        onPress={handlePress}
        testID="sync-status-indicator"
        accessibilityLabel={`Sync status: ${displayState.message}`}
        accessibilityRole="button"
      >
        <Ionicons
          name={displayState.icon}
          size={size}
          color={displayState.color}
        />
      </TouchableOpacity>
    </Animated.View>
  );
}

const styles = StyleSheet.create({
  headerRight: {
    position: 'absolute',
    right: 16,
    top: 16,
    zIndex: 10,
  },
  headerLeft: {
    position: 'absolute',
    left: 16,
    top: 16,
    zIndex: 10,
  },
});
