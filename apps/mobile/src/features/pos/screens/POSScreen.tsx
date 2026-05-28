/**
 * POSScreen Component
 * Main POS screen that integrates all POS components
 * Layout: Top control bar, Product list, Cart summary (CartList + CartTotal), Action buttons
 * Story 3.6: Transaction Processing Integration
 * Story 7.2: USB Barcode Scanner Integration
 */

import React, { useState, useRef, useEffect } from 'react';
import { View, StyleSheet, SafeAreaView, ScrollView, Alert, ActivityIndicator, Text, TextInput } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { Product } from '../types/product.types';
import { PaymentData } from '../types/payment.types';
import { useCartContext } from '../context/CartContext';
import { TopControlBar } from '../components/TopControlBar';
import { ProductList } from '../components/ProductList';
import { CartList } from '../components/CartList';
import { CartTotal } from '../components/CartTotal';
import { ActionButtons } from '../components/ActionButtons';
import { PrinterStatusComponent } from '../components/PrinterStatus';
import { ScannerFeedback } from '../components/ScannerFeedback';
import { useReceiptPrinter } from '../hooks/useReceiptPrinter';
import { useBarcodeScanner } from '../hooks/useBarcodeScanner';
import { useKeyboardInput } from '../hooks/useKeyboardInput';
import { ReceiptData } from '../types/receipt.types';
import { PrinterStatus } from '../hardware/printer';
import { ScannerState } from '../types/scanner.types';
import { TransactionService } from '../services/TransactionService';
import { ProductService } from '../services/ProductService';
import { TransactionResponse } from '../types/transaction.types';

interface POSScreenProps {
  products?: Product[];
  loading?: boolean;
  error?: string;
  pharmacyName?: string;
  pharmacyAddress?: string;
  pharmacyPhone?: string;
}

export const POSScreen: React.FC<POSScreenProps> = ({
  products = [],
  loading = false,
  error,
  pharmacyName = 'Apotek Sehat',
  pharmacyAddress = 'Jl. Sudirman No. 123, Jakarta',
  pharmacyPhone = '021-1234567',
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [paymentData, setPaymentData] = useState<PaymentData | null>(null);
  const { state, actions } = useCartContext();

  // Story 7.2: USB Barcode Scanner Integration
  const [scannerState, setScannerState] = useState<ScannerState>('idle');

  // Story 3.6 AC4: Transaction duration tracking
  // Track when first item is added to cart (transaction start time)
  const transactionStartTimeRef = useRef<Date | null>(null);
  const hasStartedTransactionRef = useRef<boolean>(false); // HIGH FIX: Use useRef to prevent stale closure race condition
  const [isProcessingTransaction, setIsProcessingTransaction] = useState(false);
  const [transactionError, setTransactionError] = useState<string | null>(null);
  const [currentIdempotencyKey, setCurrentIdempotencyKey] = useState<string | null>(null); // CRITICAL FIX: Store idempotency key for retry

  // Update transaction start time when first item is added
  useEffect(() => {
    // HIGH FIX: Use ref to check transaction started state (prevents race condition with stale closure)
    if (state.itemCount > 0 && !hasStartedTransactionRef.current && transactionStartTimeRef.current === null) {
      transactionStartTimeRef.current = new Date();
      hasStartedTransactionRef.current = true;
      console.log('Transaction start time:', transactionStartTimeRef.current.toISOString());
    }
    // Reset start time only when explicitly cleared (not when cart temporarily empty during transaction)
    // This prevents race condition when cart goes 1→0→1 quickly
  }, [state.itemCount]);

  // Receipt printer hook
  const {
    isLoading: isPrinting,
    isSuccess: printSuccess,
    error: printError,
    printerStatus,
    printerName,
    printReceipt,
    clearError,
  } = useReceiptPrinter({
    autoRetry: false,
  });

  // Story 7.2: Barcode scanner hook - handles barcode validation, debouncing, feedback
  const scanner = useBarcodeScanner({
    onBarcodeScanned: async (barcode: string) => {
      try {
        // Fetch product by barcode
        const product = await ProductService.getProductByBarcode(barcode);

        // Add to cart
        handleAddToCart(product);

        // Success feedback is handled by ScannerFeedback component
      } catch (error) {
        // Error feedback is handled by ScannerFeedback component
        // Let the error propagate for useBarcodeScanner to handle
        throw error;
      }
    },
    onError: (error) => {
      // Display error message to user
      const errorMessage = error.message || 'Gagal memindai barcode';
      Alert.alert('Scan Gagal', errorMessage);
    },
    onStateChange: setScannerState,
  });

  // Story 7.2: Keyboard input hook - captures USB HID scanner input
  const keyboardInput = useKeyboardInput({
    onCharReceived: scanner.handleScannerInput,
    enabled: true, // Scanner input always enabled when on POSScreen
  });

  const handleAddToCart = (product: Product) => {
    actions.addItem(product);
    // Transaction start time is now handled by useEffect with hasStartedTransaction flag
  };

  const handleRemoveFromCart = (productId: number) => {
    actions.removeItem(productId);
  };

  const handleUpdateQuantity = (productId: number, quantity: number) => {
    actions.updateQuantity(productId, quantity);
  };

  const handleClearCart = () => {
    actions.clearCart();
    setPaymentData(null);
    transactionStartTimeRef.current = null; // Reset transaction timer
    hasStartedTransactionRef.current = false; // HIGH FIX: Reset ref instead of state
    setTransactionError(null);
    setCurrentIdempotencyKey(null); // CRITICAL FIX: Clear idempotency key
  };

  const handleCheckout = () => {
    // TODO: Navigate to checkout/payment screen (future story)
    console.log('Checkout clicked', state);
  };

  const handlePayment = () => {
    // Payment is now handled via ActionButtons → PaymentModal
    console.log('Payment flow initiated');
  };

  /**
   * Convert cart items to receipt items
   */
  const convertCartToReceiptItems = (): ReceiptData['items'] => {
    return state.items.map((item) => ({
      name: item.product.name,
      quantity: item.quantity,
      unitPrice: item.product.price.toFixed(2),
      subtotal: (item.product.price * item.quantity).toFixed(2),
    }));
  };

  /**
   * Convert payment data to receipt payment details
   */
  const convertPaymentToReceiptPayment = (payment: PaymentData): ReceiptData['payment'] => {
    return {
      method: payment.method,
      cashDetails: payment.cashDetails,
      transferDetails: payment.transferDetails,
      ewalletDetails: payment.ewalletDetails,
    };
  };

  /**
   * Process transaction with backend API
   * Story 3.6 Task 5: Transaction creation flow with backend integration
   * AC3: Transaction number generated by backend
   * AC4: Transaction duration tracking from first scan to completion
   * AC5: Cart preserved on error with retry option
   */
  const handleProcessTransaction = async (payment: PaymentData, isRetry: boolean = false) => {
    setIsProcessingTransaction(true);
    setTransactionError(null);

    try {
      // Story 3.6 AC4: Calculate transaction duration
      const transactionEndTime = new Date();
      let transactionDuration = transactionStartTimeRef.current
        ? transactionEndTime.getTime() - transactionStartTimeRef.current.getTime()
        : 0;

      // MEDIUM FIX: Validate duration is non-negative (handles system clock adjustments)
      if (transactionDuration < 0) {
        console.warn('Negative transaction duration detected, system clock may have been adjusted');
        transactionDuration = 0;
      }

      console.log('Processing transaction:', {
        itemCount: state.itemCount,
        total: state.total,
        paymentMethod: payment.method,
        duration: transactionDuration,
        isRetry,
      });

      // CRITICAL FIX: Use existing idempotency key for retry, or null for new transaction
      const idempotencyKeyToUse = isRetry ? currentIdempotencyKey : undefined;

      // Story 3.6 AC2: Call TransactionService with cart and payment data
      const transactionResponse: TransactionResponse = await TransactionService.createTransaction(
        state.items,
        payment,
        '', // customerName
        '0', // taxAmount
        '0', // discountAmount
        idempotencyKeyToUse || undefined // Pass stored key for retry
      );

      // Story 3.6 AC4: Log transaction duration with transaction record
      console.log('Transaction completed:', {
        transactionNumber: transactionResponse.transactionNumber,
        transactionId: transactionResponse.id,
        duration: transactionDuration,
        durationSeconds: (transactionDuration / 1000).toFixed(2),
        timestamp: transactionEndTime.toISOString(),
      });

      // Story 3.6 AC3: Transaction number generated by backend
      // Generate receipt data with backend-generated transaction number
      const receiptData: ReceiptData = {
        transactionNumber: transactionResponse.transactionNumber, // Use backend-generated number
        transactionDate: transactionResponse.created_at,
        pharmacyName,
        pharmacyAddress,
        pharmacyPhone,
        items: convertCartToReceiptItems(),
        subtotal: state.subtotal.toFixed(2),
        tax: state.tax > 0 ? state.tax.toFixed(2) : undefined,
        total: state.total.toFixed(2),
        payment: convertPaymentToReceiptPayment(payment),
        paperWidth: 58,
      };

      // Print receipt with backend transaction data
      const printSuccess = await printReceipt(receiptData);

      if (printSuccess) {
        // Story 3.6 AC3: Cart cleared ONLY after successful transaction creation
        // CRITICAL FIX: Clear idempotency key on success
        setCurrentIdempotencyKey(null);

        Alert.alert(
          'Transaksi Berhasil',
          `Transaksi ${transactionResponse.transactionNumber} selesai.\nDurasi: ${(transactionDuration / 1000).toFixed(1)} detik`,
          [
            {
              text: 'OK',
              onPress: () => {
                handleClearCart();
              },
            },
          ]
        );
      } else {
        // Receipt printing failed - transaction was created but printing failed
        // CRITICAL FIX: Clear idempotency key since transaction succeeded
        setCurrentIdempotencyKey(null);

        // Keep cart cleared for data integrity, but show error
        Alert.alert(
          'Transaksi Berhasil',
          `Transaksi ${transactionResponse.transactionNumber} telah dicatat.\nGagal mencetak struk.`,
          [
            {
              text: 'OK',
              onPress: () => {
                handleClearCart();
              },
            },
          ]
        );
      }
    } catch (error) {
      // Story 3.6 AC5: Cart preserved on error, retry option available
      const errorMessage = error instanceof Error ? error.message : 'Terjadi kesalahan tidak terduga';
      setTransactionError(errorMessage);

      // CRITICAL FIX: Store idempotency key for retry (it was already persisted by TransactionService)
      if (!currentIdempotencyKey) {
        // This was the first attempt, the TransactionService already persisted the key
        // We'll retrieve it for retry
        const pendingKeys = await AsyncStorage.getAllKeys();
        const idempotencyKeyEntry = pendingKeys.find(k => k.startsWith('@simpo_pending_idempotency_'));
        if (idempotencyKeyEntry) {
          const key = await AsyncStorage.getItem(idempotencyKeyEntry);
          if (key) setCurrentIdempotencyKey(key);
        }
      }

      console.error('Transaction failed:', {
        error: errorMessage,
        cartPreserved: true,
        idempotencyKeyStored: !!currentIdempotencyKey,
      });

      Alert.alert(
        'Transaksi Gagal',
        errorMessage,
        [
          {
            text: 'Batal',
            onPress: () => {
              setTransactionError(null);
              setCurrentIdempotencyKey(null); // Clear on cancel
            },
          },
          {
            text: 'Coba Lagi',
            onPress: () => {
              setTransactionError(null);
              // CRITICAL FIX: Retry transaction with same payment data and idempotency key
              handleProcessTransaction(payment, true);
            },
          },
        ]
      );
    } finally {
      setIsProcessingTransaction(false);
    }
  };

  /**
   * Print receipt after payment confirmation
   * DEPRECATED: Use handleProcessTransaction instead (Story 3.6)
   */
  const handlePrintReceipt = async (payment: PaymentData) => {
    // Story 3.6 Task 5: Replace with actual backend transaction processing
    await handleProcessTransaction(payment);
  };

  const handlePaymentMethodSelected = (data: PaymentData) => {
    // Store payment data for transaction processing
    setPaymentData(data);

    // Log payment method selection for audit trail
    console.log('Payment method selected:', {
      method: data.method,
      timestamp: new Date().toISOString(),
      cartItemCount: state.itemCount,
      cartTotal: state.total,
    });

    // Story 3.6 Task 5: Process transaction with backend API
    // AC3: Transaction creation triggered after payment confirmation
    handleProcessTransaction(data);
  };

  return (
    <SafeAreaView style={styles.safeArea} testID="pos-screen-container">
      <View style={styles.container}>
        {/* Story 7.2: Invisible keyboard input for USB HID barcode scanner */}
        <TextInput
          ref={keyboardInput.textInputRef}
          onChangeText={keyboardInput.handleChange}
          onSubmitEditing={keyboardInput.handleSubmit}
          style={styles.invisibleTextInput}
          autoFocus={false}
          selectTextOnFocus={false}
          testID="scanner-keyboard-input"
        />

        {/* Story 7.2: Scanner feedback overlay */}
        <ScannerFeedback state={scannerState} testID="scanner-feedback" />

        {/* Transaction Processing Indicator - Story 3.6 */}
        {isProcessingTransaction && (
          <View style={styles.processingOverlay}>
            <ActivityIndicator size="large" color="#4CAF50" />
            <View style={styles.processingTextContainer}>
              <Text style={styles.processingText}>Memproses Transaksi...</Text>
            </View>
          </View>
        )}

        {/* Top Control Bar (15%) */}
        <TopControlBar
          itemCount={state.itemCount}
          total={state.total}
          searchQuery={searchQuery}
          onSearch={setSearchQuery}
          onPayment={handlePayment}
        />

        {/* Printer Status Indicator */}
        <View style={styles.printerStatusContainer}>
          <PrinterStatusComponent
            status={printerStatus}
            printerName={printerName}
            compact={true}
            testID="pos-printer-status"
          />
        </View>

        {/* Center Product Area (55%) */}
        <View style={styles.productArea}>
          <ProductList
            products={products}
            searchQuery={searchQuery}
            loading={loading}
            error={error}
            onAddToCart={handleAddToCart}
          />
        </View>

        {/* Cart Summary Panel (15%) - CartList + CartTotal */}
        <View style={styles.cartSummaryPanel}>
          <ScrollView style={styles.cartListContainer} nestedScrollEnabled={false}>
            <CartList
              cartItems={state.items}
              onUpdateQuantity={handleUpdateQuantity}
              onRemoveItem={handleRemoveFromCart}
            />
          </ScrollView>
          <CartTotal onClearCart={handleClearCart} />
        </View>

        {/* Bottom Action Buttons (15%) */}
        <ActionButtons
          itemCount={state.itemCount}
          cartTotal={state.total}
          onCheckout={handleCheckout}
          onClearCart={handleClearCart}
          onPaymentMethodSelected={handlePaymentMethodSelected}
        />
      </View>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#FFFFFF',
  },

  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },

  printerStatusContainer: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    backgroundColor: '#FFFFFF',
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
  },

  productArea: {
    flex: 1, // Takes remaining space (~55%)
    backgroundColor: '#F5F5F5',
  },

  cartSummaryPanel: {
    height: 150, // ~15% of screen height (approx)
    backgroundColor: '#FFFFFF',
    borderTopWidth: 1,
    borderTopColor: '#E0E0E0',
  },

  cartListContainer: {
    flex: 1,
  },

  // Story 3.6: Transaction processing overlay styles
  processingOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 1000,
  },

  processingTextContainer: {
    marginTop: 16,
    backgroundColor: 'rgba(255, 255, 255, 0.9)',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },

  processingText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333333',
  },

  // Story 7.2: Invisible input for USB HID barcode scanner
  invisibleTextInput: {
    height: 0,
    width: 0,
    opacity: 0,
    position: 'absolute',
    top: -9999, // Position off-screen
  },
});
