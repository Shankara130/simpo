/**
 * TopControlBar Component
 * Top control area with product search/barcode scan input and cart summary
 * Displays running total and payment button
 * Enhanced with barcode scanner integration
 * Story 3.7: Added Transaction History button
 */

import React, { useRef, useEffect } from 'react';
import { View, Text, TextInput, TouchableOpacity, StyleSheet } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import Icon from 'react-native-vector-icons/MaterialIcons';
import { useBarcodeScanner, ScannerState } from '../hooks/useBarcodeScanner';
import { POSStackParamList } from '../types/navigation.types';

type TopControlBarNavigationProp = StackNavigationProp<
  POSStackParamList,
  'POS'
>;

interface TopControlBarProps {
  itemCount: number;
  total: string;
  searchQuery?: string;
  onSearch: (query: string) => void;
  onPayment: () => void;
  onBarcodeScanned?: (barcode: string) => void | Promise<void>;
}

export const TopControlBar: React.FC<TopControlBarProps> = ({
  itemCount,
  total,
  searchQuery = '',
  onSearch,
  onPayment,
  onBarcodeScanned,
}) => {
  const navigation = useNavigation<TopControlBarNavigationProp>();

  const formatPrice = (price: string): string => {
    // Validate price input
    if (!price || typeof price !== 'string') {
      return 'Rp 0';
    }

    const priceNum = parseFloat(price);

    // Validate parsed price
    if (isNaN(priceNum) || priceNum < 0) {
      return 'Rp 0';
    }

    return `Rp ${priceNum.toLocaleString('id-ID')}`;
  };

  const inputRef = useRef<TextInput>(null);
  const lastScanTimeRef = useRef<number>(0);

  const isCartEmpty = itemCount === 0;

  // Barcode scanner integration
  const { state: scannerState, handleScannerInput, reset: resetScanner } = useBarcodeScanner({
    onBarcodeScanned: async (barcode) => {
      // Forward to parent callback if provided
      if (onBarcodeScanned) {
        await onBarcodeScanned(barcode);
      }
      // Update search query to show scanned barcode
      onSearch(barcode);
    },
    onStateChange: (newState) => {
      // Could be used to update UI based on scanner state
      console.log('[TopControlBar] Scanner state:', newState);
    },
    onError: (error) => {
      console.error('[TopControlBar] Scanner error:', error);
      // Could show error banner here
    },
  });

  // Auto-focus input when component mounts
  useEffect(() => {
    const timer = setTimeout(() => {
      if (inputRef.current && !inputRef.current.isFocused()) {
        inputRef.current?.focus();
      }
    }, 500);

    return () => clearTimeout(timer);
  }, []);

  // Handle text input changes (manual search)
  const handleTextChange = (text: string) => {
    const now = Date.now();

    // Reset scanner if user is manually typing
    if (now - lastScanTimeRef.current > 100) {
      // This is manual input, clear any scanner state
      if (scannerState === 'scanning') {
        resetScanner();
      }
    }

    onSearch(text);
  };

  // Handle keyboard input with timing for scanner detection
  const handleKeyPress = (nativeEvent: any) => {
    const char = nativeEvent.key;

    // Track timing for scanner detection
    const now = Date.now();
    if (lastScanTimeRef.current === 0) {
      lastScanTimeRef.current = now;
    }

    // Pass to scanner hook for processing
    if (char) {
      handleScannerInput(char, now);
    }
  };

  // Handle input submission (Enter key)
  const handleSubmitEditing = () => {
    // Scanner input is already handled by handleScannerInput
    // This is for manual search submission
    if (searchQuery && searchQuery.trim().length > 0) {
      onSearch(searchQuery);
    }
  };

  // Get scanner state color for visual indicator
  const getScannerStateColor = (): string => {
    switch (scannerState) {
      case 'scanning':
        return '#2196F3'; // Blue - actively scanning
      case 'success':
        return '#4CAF50'; // Green - scan successful
      case 'error':
        return '#F44336'; // Red - scan error
      case 'loading':
        return '#FF9800'; // Orange - processing
      default:
        return '#BDBDBD'; // Gray - idle
    }
  };

  return (
    <View style={styles.container}>
      <View style={styles.searchContainer}>
        <View style={styles.scannerIndicator}>
          <View
            style={[
              styles.scannerDot,
              { backgroundColor: getScannerStateColor() },
            ]}
          />
        </View>
        <TextInput
          ref={inputRef}
          style={styles.searchInput}
          placeholder="Search products or scan barcode..."
          placeholderTextColor="#757575"
          value={searchQuery}
          onChangeText={handleTextChange}
          onSubmitEditing={handleSubmitEditing}
          onKeyPress={handleKeyPress}
          autoCapitalize="none"
          autoCorrect={false}
          blurOnSubmit={false}
          selectTextOnFocus
          testID="search-input"
        />
      </View>

      <View style={styles.cartSummary}>
        <View style={styles.cartInfo}>
          <Text style={styles.cartCount}>
            Cart: {itemCount} {itemCount === 1 ? 'item' : 'items'}
          </Text>
          <Text style={styles.cartTotal}>{formatPrice(total)}</Text>
        </View>

        <TouchableOpacity
          testID="payment-button"
          style={[styles.paymentButton, isCartEmpty && styles.paymentButtonDisabled]}
          onPress={onPayment}
          disabled={isCartEmpty}
          activeOpacity={0.7}
        >
          <Text style={[styles.paymentButtonText, isCartEmpty && styles.paymentButtonTextDisabled]}>
            Pay
          </Text>
        </TouchableOpacity>

        {/* Story 3.7: Transaction History button */}
        <TouchableOpacity
          testID="history-button"
          style={styles.historyButton}
          onPress={() => navigation.navigate('TransactionHistory')}
          activeOpacity={0.7}
        >
          <Icon name="history" size={24} color="#2196F3" />
        </TouchableOpacity>

        {/* Story 7.2: Scanner Settings button */}
        <TouchableOpacity
          style={styles.settingsButton}
          onPress={() => navigation.navigate('ScannerSettings')}
          activeOpacity={0.7}
          testID="scanner-settings-button"
        >
          <Icon name="settings" size={24} color="#2196F3" />
        </TouchableOpacity>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#FFFFFF',
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
    padding: 16,
    minHeight: 100,
  },

  searchContainer: {
    marginBottom: 12,
  },

  searchInput: {
    backgroundColor: '#F5F5F5',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#212121',
    borderWidth: 1,
    borderColor: '#E0E0E0',
    minHeight: 48,
  },

  cartSummary: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },

  cartInfo: {
    flex: 1,
    marginRight: 12,
  },

  cartCount: {
    fontSize: 14,
    color: '#757575',
    marginBottom: 4,
  },

  cartTotal: {
    fontSize: 20,
    fontWeight: '700',
    color: '#4CAF50',
  },

  paymentButton: {
    backgroundColor: '#4CAF50',
    paddingHorizontal: 32,
    paddingVertical: 12,
    borderRadius: 8,
    minWidth: 100,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 48,
  },

  paymentButtonDisabled: {
    backgroundColor: '#BDBDBD',
    opacity: 0.5,
  },

  paymentButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },

  paymentButtonTextDisabled: {
    color: '#757575',
  },

  scannerIndicator: {
    position: 'absolute',
    left: 12,
    top: 12,
    zIndex: 1,
  },

  scannerDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: '#BDBDBD',
  },

  // Story 3.7: Transaction History button styles
  historyButton: {
    backgroundColor: '#E3F2FD',
    paddingHorizontal: 12,
    paddingVertical: 12,
    borderRadius: 8,
    marginLeft: 8,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 48,
  },

  // Story 7.2: Scanner Settings button styles
  settingsButton: {
    backgroundColor: '#E8F5E9',
    paddingHorizontal: 12,
    paddingVertical: 12,
    borderRadius: 8,
    marginLeft: 8,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 48,
  },
});
