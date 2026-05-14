/**
 * POSScreen Component
 * Main POS screen that integrates all POS components
 * Layout: Top control bar, Product list, Cart summary (CartList + CartTotal), Action buttons
 */

import React, { useState } from 'react';
import { View, StyleSheet, SafeAreaView, ScrollView, Alert } from 'react-native';
import { Product } from '../types/product.types';
import { PaymentData } from '../types/payment.types';
import { useCartContext } from '../context/CartContext';
import { TopControlBar } from '../components/TopControlBar';
import { ProductList } from '../components/ProductList';
import { CartList } from '../components/CartList';
import { CartTotal } from '../components/CartTotal';
import { ActionButtons } from '../components/ActionButtons';
import { PrinterStatusComponent } from '../components/PrinterStatus';
import { useReceiptPrinter } from '../hooks/useReceiptPrinter';
import { ReceiptData } from '../types/receipt.types';
import { PrinterStatus } from '../hardware/printer';

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

  const handleAddToCart = (product: Product) => {
    actions.addItem(product);
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
   * Generate transaction number
   * Format: TRX-YYYYMMDD-XXXX
   */
  const generateTransactionNumber = (): string => {
    const now = new Date();
    const date = now.toISOString().split('T')[0].replace(/-/g, '');
    const random = Math.floor(Math.random() * 10000)
      .toString()
      .padStart(4, '0');
    return `TRX-${date}-${random}`;
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
   * Print receipt after payment confirmation
   */
  const handlePrintReceipt = async (payment: PaymentData) => {
    try {
      // Generate receipt data
      const receiptData: ReceiptData = {
        transactionNumber: generateTransactionNumber(),
        transactionDate: new Date().toISOString(),
        pharmacyName,
        pharmacyAddress,
        pharmacyPhone,
        items: convertCartToReceiptItems(),
        subtotal: state.subtotal.toFixed(2),
        tax: state.tax > 0 ? state.tax.toFixed(2) : undefined,
        total: state.total.toFixed(2),
        payment: convertPaymentToReceiptPayment(payment),
        paperWidth: 58, // Default to 58mm paper width
      };

      // Audit trail log
      console.log('Printing receipt:', {
        transactionNumber: receiptData.transactionNumber,
        timestamp: receiptData.transactionDate,
        itemCount: receiptData.items.length,
        total: receiptData.total,
        paymentMethod: payment.method,
      });

      // Print receipt
      const success = await printReceipt(receiptData);

      if (success) {
        // Receipt printed successfully - clear cart and show success message
        Alert.alert(
          'Struk Berhasil Dicetak',
          `Transaksi ${receiptData.transactionNumber} selesai.`,
          [
            {
              text: 'OK',
              onPress: () => {
                handleClearCart();
              },
            },
          ]
        );

        // Audit trail log
        console.log('Receipt printed successfully:', {
          transactionNumber: receiptData.transactionNumber,
          timestamp: new Date().toISOString(),
          status: 'success',
        });
      } else {
        // Receipt printing failed - show error with retry option
        Alert.alert(
          'Gagal Mencetak Struk',
          printError || 'Terjadi kesalahan saat mencetak struk.',
          [
            {
              text: 'Batal',
              onPress: () => {
                // User cancels - keep cart for manual retry
                clearError();
              },
            },
            {
              text: 'Coba Lagi',
              onPress: () => {
                // Retry printing
                handlePrintReceipt(payment);
              },
            },
          ]
        );

        // Audit trail log
        console.error('Receipt printing failed:', {
          transactionNumber: receiptData.transactionNumber,
          timestamp: new Date().toISOString(),
          error: printError,
          status: 'failed',
        });
      }
    } catch (error) {
      console.error('Error printing receipt:', error);
      Alert.alert(
        'Error',
        'Terjadi kesalahan saat mencetak struk. Silakan coba lagi.',
        [
          {
            text: 'OK',
            onPress: () => {
              clearError();
            },
          },
        ]
      );
    }
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

    // Trigger receipt printing after payment confirmation
    handlePrintReceipt(data);

    // TODO: Future story (3.6) - Pass payment data to transaction creation endpoint
    // For now, just store it locally and log the selection
    console.log('Payment data stored for transaction:', data);
  };

  return (
    <SafeAreaView style={styles.safeArea} testID="pos-screen-container">
      <View style={styles.container}>
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
});
