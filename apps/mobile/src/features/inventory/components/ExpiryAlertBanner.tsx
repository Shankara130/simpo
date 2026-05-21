/**
 * ExpiryAlertBanner Component
 * Story 4.5, Task 11: Create Mobile Expiry Alert Banner (AC: 5, 7)
 *
 * This component displays an animated banner when products are approaching expiry.
 * Features:
 * - Animated banner appearance/disappearance
 * - Color-coded alert levels (yellow → orange → red)
 * - Displays product info and days remaining
 * - Urgent styling for 7-day alerts (red background, bold)
 * - Swipe-to-dismiss support
 * - Auto-dismiss after 60 seconds
 */

import React, { useEffect, useState, useRef } from 'react';
import {
  View,
  Text,
  StyleSheet,
  Animated,
  TouchableOpacity,
} from 'react-native';
import { PanGestureHandler, GestureHandlerRootView } from 'react-native-gesture-handler';
import { ExpiryEvent } from '../services/realTimeStockService';

interface ExpiryAlertBannerProps {
  event: ExpiryEvent;
  onDismiss: () => void;
  onViewProduct?: (productId: number) => void;
  autoDismissDelay?: number;
}

/**
 * Get alert colors based on alert level
 * Story 4.5, Task 11.3: Display alert banner with color coding
 */
function getAlertColors(alertLevel: 'warning' | 'critical' | 'urgent') {
  switch (alertLevel) {
    case 'urgent':
      return {
        background: '#FEE2E2', // red-50
        text: '#991B1B', // red-800
        iconBackground: '#DC2626', // red-600
        border: '#EF4444', // red-500
      };
    case 'critical':
      return {
        background: '#FED7AA', // orange-200
        text: '#9A3412', // orange-800
        iconBackground: '#EA580C', // orange-600
        border: '#F97316', // orange-500
      };
    case 'warning':
    default:
      return {
        background: '#FEF3C7', // yellow-100
        text: '#92400E', // yellow-800
        iconBackground: '#D97706', // yellow-600
        border: '#F59E0B', // yellow-500
      };
  }
}

/**
 * Animated banner for expiry alerts
 * Story 4.5, Task 11.1-11.5: Expiry alert banner with animation and auto-dismiss
 */
export function ExpiryAlertBanner({
  event,
  onDismiss,
  onViewProduct,
  autoDismissDelay = 60000,
}: ExpiryAlertBannerProps) {
  const [isVisible, setIsVisible] = useState(true);
  const translateY = useRef(new Animated.Value(-100));
  const opacity = useRef(new Animated.Value(0));
  const colors = getAlertColors(event.alertLevel);

  // Auto-dismiss timer
  useEffect(() => {
    const timer = setTimeout(() => {
      handleDismiss();
    }, autoDismissDelay);

    return () => clearTimeout(timer);
  }, [autoDismissDelay]);

  // Animate in on mount
  useEffect(() => {
    Animated.parallel([
      Animated.timing(translateY.current, {
        toValue: 0,
        duration: 300,
        useNativeDriver: true,
      }),
      Animated.timing(opacity.current, {
        toValue: 1,
        duration: 300,
        useNativeDriver: true,
      }),
    ]).start();
  }, []);

  /**
   * Handle dismiss with animation
   * Story 4.5, Task 11.5: Swipe-to-dismiss functionality
   */
  const handleDismiss = () => {
    Animated.parallel([
      Animated.timing(translateY.current, {
        toValue: -100,
        duration: 300,
        useNativeDriver: true,
      }),
      Animated.timing(opacity.current, {
        toValue: 0,
        duration: 300,
        useNativeDriver: true,
      }),
    ]).start(() => {
      setIsVisible(false);
      onDismiss();
    });
  };

  /**
   * Handle view product button press
   */
  const handleViewProduct = () => {
    if (onViewProduct) {
      onViewProduct(event.productId);
    }
    handleDismiss();
  };

  /**
   * Handle pan gesture for swipe to dismiss
   * Story 4.5, Task 11.5: Swipe-to-dismiss functionality
   */
  const handlePanGestureEvent = Animated.event(
    [{ nativeEvent: ({ translateY }) => translateY }],
    {
      useNativeDriver: true,
      listener: ({ nativeEvent }) => {
        const { translateY } = nativeEvent;
        // Dismiss if swiped up more than 50px
        if (translateY < -50) {
          handleDismiss();
        }
      },
    }
  );

  // Format expiry date for display
  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  if (!isVisible) {
    return null;
  }

  const isUrgent = event.alertLevel === 'urgent';

  return (
    <GestureHandlerRootView>
      <Animated.View
        style={[
          styles.banner,
          {
            backgroundColor: colors.background,
            borderColor: colors.border,
            transform: [{ translateY: translateY.current }],
            opacity: opacity.current,
          },
        ]}
      >
        <PanGestureHandler onGestureEvent={handlePanGestureEvent}>
          <View style={styles.bannerContent}>
            {/* Alert Icon */}
            <View style={[styles.iconContainer, { backgroundColor: colors.iconBackground + '20' }]}>
              <Text style={[styles.icon, { color: colors.iconBackground }]}>
                ⚠
              </Text>
            </View>

            {/* Content */}
            <View style={styles.textContainer}>
              <Text style={[styles.alertTitle, { color: colors.text }]}>
                {isUrgent && <Text style={styles.urgentLabel}>URGENT: </Text>}
                Expiry Alert
              </Text>

              <Text style={[styles.productName, { color: colors.text }]}>
                {event.productName}
              </Text>

              <Text style={[styles.sku, { color: colors.text }]}>
                SKU: {event.sku}
              </Text>

              <View style={styles.infoRow}>
                <Text style={[styles.infoLabel, { color: colors.text }]}>
                  Expiry:
                </Text>
                <Text style={[styles.infoValue, { color: colors.text }]}>
                  {formatDate(event.expiryDate)}
                </Text>
              </View>

              <View style={styles.infoRow}>
                <Text style={[styles.infoLabel, { color: colors.text }]}>
                  Location:
                </Text>
                <Text style={[styles.infoValue, { color: colors.text }]}>
                  {event.branchName}
                </Text>
              </View>

              <View style={styles.daysContainer}>
                <Text style={[
                  styles.daysText,
                  {
                    color: colors.text,
                    fontWeight: isUrgent ? '700' : '600',
                  },
                ]}>
                  {event.daysRemaining} day{event.daysRemaining !== 1 ? 's' : ''} remaining
                </Text>
              </View>
            </View>

            {/* Action Buttons */}
            <View style={styles.buttonContainer}>
              {onViewProduct && (
                <TouchableOpacity
                  onPress={handleViewProduct}
                  style={[styles.viewButton, { backgroundColor: colors.iconBackground }]}
                >
                  <Text style={styles.viewButtonText}>View</Text>
                </TouchableOpacity>
              )}

              <TouchableOpacity onPress={handleDismiss} style={styles.dismissButton}>
                <Text style={styles.dismissButtonText}>✕</Text>
              </TouchableOpacity>
            </View>
          </View>
        </PanGestureHandler>
      </Animated.View>
    </GestureHandlerRootView>
  );
}

/**
 * Container for managing multiple expiry alert banners
 * Story 4.5, Task 11.5: Support multiple alerts
 */
interface BannerContainerProps {
  alerts: Array<{
    id: string;
    event: ExpiryEvent;
  }>;
  onDismiss: (id: string) => void;
  onViewProduct?: (productId: number) => void;
  maxVisible?: number;
}

export function ExpiryAlertBannerContainer({
  alerts,
  onDismiss,
  onViewProduct,
  maxVisible = 3,
}: BannerContainerProps) {
  if (alerts.length === 0) {
    return null;
  }

  // Sort alerts by urgency (urgent first)
  const sortedAlerts = [...alerts].sort((a, b) => {
    const urgencyOrder = { urgent: 0, critical: 1, warning: 2 };
    return urgencyOrder[a.event.alertLevel] - urgencyOrder[b.event.alertLevel];
  });

  // Limit number of visible banners
  const visibleAlerts = sortedAlerts.slice(-maxVisible);

  return (
    <View style={styles.container} pointerEvents="box-none">
      {visibleAlerts.map(({ id, event }, index) => (
        <Animated.View
          key={id}
          style={[
            styles.bannerWrapper,
            { top: 60 + index * 140 }, // Stack banners with offset (expiry banners are taller)
          ]}
        >
          <ExpiryAlertBanner
            event={event}
            onDismiss={() => onDismiss(id)}
            onViewProduct={onViewProduct}
          />
        </Animated.View>
      ))}
    </View>
  );
}

/**
 * Hook for managing expiry alert banners
 * Story 4.5, Task 11.5: Support multiple alerts
 */
interface UseExpiryAlertsReturn {
  alerts: Array<{
    id: string;
    event: ExpiryEvent;
  }>;
  addAlert: (event: ExpiryEvent) => void;
  removeAlert: (id: string) => void;
  clearAll: () => void;
}

export function useExpiryAlerts(maxAlerts: number = 5): UseExpiryAlertsReturn {
  const [alerts, setAlerts] = useState<Array<{
    id: string;
    event: ExpiryEvent;
  }>>([]);

  const addAlert = (event: ExpiryEvent) => {
    const id = `expiry-${event.productId}-${event.branchId}-${event.expiryDate}`;
    setAlerts(prev => {
      // Remove existing alert for same product/branch/expiry if exists
      const filtered = prev.filter(a => a.id !== id);
      // Add new alert and limit total
      return [...filtered, { id, event }].slice(-maxAlerts);
    });
  };

  const removeAlert = (id: string) => {
    setAlerts(prev => prev.filter(a => a.id !== id));
  };

  const clearAll = () => {
    setAlerts([]);
  };

  return {
    alerts,
    addAlert,
    removeAlert,
    clearAll,
  };
}

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    zIndex: 9999,
    pointerEvents: 'none',
  },

  bannerWrapper: {
    position: 'absolute',
    left: 16,
    right: 16,
  },

  banner: {
    borderRadius: 12,
    borderWidth: 2,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.25,
    shadowRadius: 4,
    elevation: 5,
    marginHorizontal: 16,
    overflow: 'hidden',
  },

  bannerContent: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    padding: 16,
    gap: 12,
  },

  iconContainer: {
    width: 40,
    height: 40,
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
    marginTop: 4,
  },

  icon: {
    fontSize: 24,
  },

  textContainer: {
    flex: 1,
    gap: 4,
  },

  alertTitle: {
    fontSize: 14,
    fontWeight: '600',
  },

  urgentLabel: {
    fontWeight: '800',
  },

  productName: {
    fontSize: 16,
    fontWeight: '700',
  },

  sku: {
    fontSize: 12,
    fontFamily: 'monospace',
  },

  infoRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },

  infoLabel: {
    fontSize: 12,
    fontWeight: '500',
  },

  infoValue: {
    fontSize: 12,
    fontWeight: '600',
  },

  daysContainer: {
    marginTop: 4,
    paddingVertical: 6,
    paddingHorizontal: 10,
    borderRadius: 8,
    backgroundColor: 'rgba(0, 0, 0, 0.05)',
  },

  daysText: {
    fontSize: 14,
  },

  buttonContainer: {
    gap: 8,
  },

  viewButton: {
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: 8,
  },

  viewButtonText: {
    color: '#FFFFFF',
    fontSize: 12,
    fontWeight: '600',
  },

  dismissButton: {
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: 'rgba(0, 0, 0, 0.05)',
    justifyContent: 'center',
    alignItems: 'center',
  },

  dismissButtonText: {
    fontSize: 18,
    color: '#757575',
  },
});
