/**
 * StockUpdateBanner Component
 * Story 4.2, Task 13: Create Mobile Stock Update Indicator (AC: 2, 3, 5)
 *
 * This component displays an animated banner when stock levels change.
 * Features:
 * - Animated banner appearance/disappearance
 * - Displays product name and stock change
 * - Auto-dismiss after 5 seconds
 * - Swipe to dismiss support
 */

import React, { useEffect, useState, useRef } from 'react';
import {
  View,
  Text,
  StyleSheet,
  Animated,
  PanGestureHandler,
  GestureHandlerRootView,
  State,
} from 'react-native';
import { StockUpdatedEvent } from '../services/realTimeStockService';

interface StockUpdateBannerProps {
  event: StockUpdatedEvent;
  onDismiss: () => void;
  autoDismissDelay?: number;
}

/**
 * Animated banner for stock updates
 * Story 4.2, Task 13.1-13.4: Stock update banner with animation and auto-dismiss
 */
export function StockUpdateBanner({
  event,
  onDismiss,
  autoDismissDelay = 5000,
}: StockUpdateBannerProps) {
  const [isVisible, setIsVisible] = useState(true);
  const translateY = useRef(new Animated.Value(-100));
  const opacity = useRef(new Animated.Value(0));

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
   * Story 4.2, Task 13.4: Auto-dismiss after 5 seconds
   * Story 4.2, Task 13.5: Support swipe to dismiss
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
   * Handle pan gesture for swipe to dismiss
   * Story 4.2, Task 13.5: Support swipe to dismiss
   */
  const handlePanGestureEvent = Animated.event(
    [{ nativeEvent: ({ translateY }) => translateY }],
    {
      listener: ({ nativeEvent }) => {
        const { translateY } = nativeEvent;
        // Dismiss if swiped down more than 50px
        if (translateY > 50) {
          handleDismiss();
        }
      },
    }
  );

  // Determine colors based on change direction
  const isIncrease = event.change > 0;
  const isDecrease = event.change < 0;
  const backgroundColor = isIncrease
    ? '#D1FAE5' // green-50
    : isDecrease
    ? '#FEE2E2' // red-50
    : '#F3F4F6'; // gray-50

  const textColor = isIncrease
    ? '#065F46' // green-800
    : isDecrease
    ? '#991B1B' // red-800
    : '#1F2937'; // gray-800

  const icon = isIncrease ? '↑' : isDecrease ? '↓' : '→';
  const changeText = event.change > 0 ? `+${event.change}` : `${event.change}`;

  if (!isVisible) {
    return null;
  }

  return (
    <GestureHandlerRootView>
      <Animated.View
        style={[
          styles.banner,
          { backgroundColor, transform: [{ translateY: translateY.current }], opacity: opacity.current },
        ]}
      >
        <PanGestureHandler onGestureEvent={handlePanGestureEvent}>
          <View style={styles.bannerContent}>
            {/* Icon */}
            <View style={[styles.iconContainer, { backgroundColor: textColor + '20' }]}>
              <Text style={[styles.icon, { color: textColor }]}>
                {icon}
              </Text>
            </View>

            {/* Content */}
            <View style={styles.textContainer}>
              <Text style={[styles.productName, { color: textColor }]}>
                {event.name}
              </Text>
              <Text style={[styles.stockChange, { color: textColor }]}>
                <Text style={styles.oldStock}>{event.oldStock}</Text>
                {' → '}
                <Text style={styles.newStock}>{event.newStock}</Text>
                {' '}
                <Text style={styles.changeText}>({changeText})</Text>
              </Text>
              <View style={styles.metaInfo}>
                <Text style={styles.sku}>{event.sku}</Text>
                {event.branchId && (
                  <>
                    <Text style={styles.separator}>•</Text>
                    <Text style={styles.metaText}>Branch {event.branchId}</Text>
                  </>
                )}
                <Text style={styles.separator}>•</Text>
                <Text style={styles.metaText}>
                  {new Date(event.updatedAt).toLocaleTimeString()}
                </Text>
              </View>
              <Text style={styles.updatedBy}>Updated by: {event.updatedBy}</Text>
            </View>

            {/* Dismiss Button */}
            <TouchableOpacity onPress={handleDismiss} style={styles.dismissButton}>
              <Text style={styles.dismissButtonText}>✕</Text>
            </TouchableOpacity>
          </View>
        </PanGestureHandler>
      </Animated.View>
    </GestureHandlerRootView>
  );
}

/**
 * Container for managing multiple stock update banners
 * Story 4.2, Task 13.1-13.4: Stock update banner management
 */
interface BannerContainerProps {
  banners: Array<{
    id: string;
    event: StockUpdatedEvent;
  }>;
  onDismiss: (id: string) => void;
  maxVisible?: number;
}

export function StockUpdateBannerContainer({
  banners,
  onDismiss,
  maxVisible = 3,
}: BannerContainerProps) {
  if (banners.length === 0) {
    return null;
  }

  // Limit number of visible banners
  const visibleBanners = banners.slice(-maxVisible);

  return (
    <View style={styles.container} pointerEvents="box-none">
      {visibleBanners.map(({ id, event }, index) => (
        <Animated.View
          key={id}
          style={[
            styles.bannerWrapper,
            { top: 60 + index * 80 }, // Stack banners with offset
          ]}
        >
          <StockUpdateBanner
            event={event}
            onDismiss={() => onDismiss(id)}
          />
        </Animated.View>
      ))}
    </View>
  );
}

/**
 * Hook for managing stock update banners
 * Story 4.2, Task 13.4: Auto-dismiss after 5 seconds
 */
interface UseStockBannersReturn {
  banners: Array<{
    id: string;
    event: StockUpdatedEvent;
  }>;
  addBanner: (event: StockUpdatedEvent) => void;
  removeBanner: (id: string) => void;
  clearAll: () => void;
}

export function useStockBanners(maxBanners: number = 5): UseStockBannersReturn {
  const [banners, setBanners] = useState<Array<{
    id: string;
    event: StockUpdatedEvent;
  }>>([]);

  const addBanner = (event: StockUpdatedEvent) => {
    const id = `${event.productId}-${event.branchId}-${Date.now()}`;
    setBanners(prev => {
      // Remove existing banner for same product/branch if exists
      const filtered = prev.filter(b => b.id !== id);
      // Add new banner and limit total
      return [...filtered, { id, event }].slice(-maxBanners);
    });
  };

  const removeBanner = (id: string) => {
    setBanners(prev => prev.filter(b => b.id !== id));
  };

  const clearAll = () => {
    setBanners([]);
  };

  return {
    banners,
    addBanner,
    removeBanner,
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
    alignItems: 'center',
    padding: 16,
    gap: 12,
  },

  iconContainer: {
    width: 40,
    height: 40,
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
  },

  icon: {
    fontSize: 20,
    fontWeight: 'bold',
  },

  textContainer: {
    flex: 1,
    gap: 2,
  },

  productName: {
    fontSize: 15,
    fontWeight: '600',
  },

  stockChange: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    fontSize: 14,
  },

  oldStock: {
    textDecorationLine: 'underline',
    opacity: 0.6,
  },

  newStock: {
    fontWeight: '700',
  },

  changeText: {
    fontSize: 12,
    fontWeight: '600',
  },

  metaInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    fontSize: 12,
  },

  sku: {
    fontFamily: 'monospace',
    fontSize: 11,
  },

  separator: {
    color: '#757575',
  },

  metaText: {
    color: '#757575',
    fontSize: 11,
  },

  updatedBy: {
    fontSize: 10,
    color: '#757575',
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
