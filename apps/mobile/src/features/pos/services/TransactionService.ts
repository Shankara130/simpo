/**
 * TransactionService - API integration for Transaction processing
 * Handles transaction creation with backend API
 * Story 3.6: Implement Transaction Processing <30 Seconds
 */

import axios, { AxiosError } from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { v4 as uuidv4 } from 'uuid'; // CRITICAL-003: UUID for idempotency keys
import {
  SaleRequest,
  TransactionResponse,
  TransactionServiceError,
  TransactionErrorResponse,
} from '../types/transaction.types';
import { CartItem } from '../types/cart.types';
import { PaymentData, PaymentMethod } from '../types/payment.types';

// API base URL - configure for different environments
const API_BASE_URL = __DEV__
  ? 'http://localhost:8080' // Development - local backend
  : 'https://api.simpo.id';  // Production - actual API endpoint

// API version prefix
const API_VERSION = '/api/v1';

// Full API base URL
const FULL_API_URL = `${API_BASE_URL}${API_VERSION}`;

// JWT token storage key
const JWT_TOKEN_KEY = '@simpo_jwt_token';

/**
 * Indonesian error messages for common transaction failures
 * Story 3.6 AC5: Mobile displays Indonesian error messages
 */
const INDONESIAN_ERRORS: Record<string, string> = {
  'Insufficient Stock': 'Stok tidak mencukupi',
  'Product out of stock': 'Produk stok habis',
  'insufficient stock': 'Stok tidak mencukupi',
  'stock not available': 'Produk stok habis',
  'empty-cart': 'Keranjang tidak boleh kosong',
  'unauthorized': 'Sesi tidak valid, silakan login kembali',
  'network error': 'Koneksi gagal, coba lagi',
  'timeout': 'Waktu habis, silakan coba lagi',
  'transaction-failed': 'Transaksi gagal, silakan coba lagi',
};

/**
 * Convert PaymentData discriminated union to backend payment method string
 * Story 3.6 AC2: Convert PaymentData to backend format
 */
const convertPaymentMethod = (paymentData: PaymentData): string => {
  switch (paymentData.method) {
    case PaymentMethod.CASH:
      return 'CASH';
    case PaymentMethod.TRANSFER:
      return 'TRANSFER';
    case PaymentMethod.E_WALLET:
      return 'E-WALLET';
    default:
      // HIGH FIX: Throw error for unknown payment method instead of silent fallback
      throw new TransactionServiceError('Invalid payment method', 400);
  }
};

/**
 * Convert CartItem array to SaleItem array for backend API
 * Story 3.6 AC2: Convert CartContext items to SaleRequest format
 */
const convertCartItemsToSaleItems = (cartItems: CartItem[]): SaleRequest['items'] => {
  return cartItems.map(item => ({
    product_id: item.productId,
    quantity: item.quantity,
    unit_price: item.price,
  }));
};

/**
 * Get JWT token from AsyncStorage
 * Story 3.6 AC2: Service includes JWT authentication from AsyncStorage
 */
const getAuthToken = async (): Promise<string> => {
  try {
    const token = await AsyncStorage.getItem(JWT_TOKEN_KEY);
    if (!token) {
      throw new TransactionServiceError('No authentication token found', 401);
    }
    return token;
  } catch (error) {
    throw new TransactionServiceError('Failed to retrieve authentication token', 401);
  }
};

/**
 * Map API error response to Indonesian error message
 * Story 3.6 AC5: Indonesian error message mapping
 */
const mapErrorToIndonesian = (errorDetail: string): string => {
  // Check if any known error substring exists in the error detail
  for (const [english, indonesian] of Object.entries(INDONESIAN_ERRORS)) {
    if (errorDetail.toLowerCase().includes(english.toLowerCase())) {
      return indonesian;
    }
  }

  // Default Indonesian message if no specific match found
  return errorDetail;
};

/**
 * TransactionService - Transaction API operations
 * Story 3.6 AC2: Mobile TransactionService calls backend API
 */
export const TransactionService = {
  /**
   * Create a new transaction with cart and payment data
   * Story 3.6 AC2: Service timeout set to 15 seconds
   */
  createTransaction: async (
    cartItems: CartItem[],
    paymentData: PaymentData,
    customerName: string = '',
    taxAmount: string = '0',
    discountAmount: string = '0',
    idempotencyKeyOverride?: string // CRITICAL FIX: Allow passing existing idempotency key for retry
  ): Promise<TransactionResponse> => {
    try {
      // Validate cart is not empty
      if (!cartItems || cartItems.length === 0) {
        throw new TransactionServiceError('Keranjang tidak boleh kosong', 400);
      }

      // Get JWT token for authentication
      const token = await getAuthToken();

      // CRITICAL FIX: Generate and persist idempotency key BEFORE API call
      // This prevents duplicate charges if app crashes before response
      const idempotencyKey = idempotencyKeyOverride || uuidv4(); // Use override for retry, or generate new

      // Persist idempotency key to AsyncStorage for crash recovery
      const pendingKey = `@simpo_pending_idempotency_${Date.now()}`;
      await AsyncStorage.setItem(pendingKey, idempotencyKey);

      // Convert cart items to sale items
      const items = convertCartItemsToSaleItems(cartItems);

      // Convert payment method
      const paymentMethod = convertPaymentMethod(paymentData);

      // Build sale request
      const saleRequest: SaleRequest = {
        items,
        payment_method: paymentMethod,
        customer_name: customerName,
        idempotency_key: idempotencyKey, // CRITICAL-003: Include idempotency key
        tax_amount: taxAmount,
        discount_amount: discountAmount,
      };

      // Make API call with 15 second timeout
      const response = await axios.post<TransactionResponse>(
        `${FULL_API_URL}/transactions`,
        saleRequest,
        {
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
          timeout: 15000, // 15 seconds (Story 3.6 AC2)
        }
      );

      // CRITICAL FIX: Clear persisted idempotency key after successful transaction
      await AsyncStorage.removeItem(pendingKey);

      return response.data;
    } catch (error) {
      throw TransactionService._handleError(error);
    }
  },

  /**
   * Handle API errors and convert to TransactionServiceError with Indonesian messages
   * Story 3.6 AC5: Indonesian error messages for common failures
   * @private
   */
  _handleError: (error: unknown): TransactionServiceError => {
    if (error instanceof TransactionServiceError) {
      return error;
    }

    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError<TransactionErrorResponse>;

      // Network error (no response)
      if (!axiosError.response) {
        const indonesianMessage = INDONESIAN_ERRORS['network error'];
        return new TransactionServiceError(indonesianMessage, undefined, axiosError);
      }

      // Server responded with error status
      const status = axiosError.response.status;
      const data = axiosError.response.data;

      // Extract error detail and convert to Indonesian
      const errorDetail = data?.detail || '';
      const indonesianMessage = mapErrorToIndonesian(errorDetail);

      return new TransactionServiceError(
        indonesianMessage || 'Terjadi kesalahan',
        status,
        axiosError
      );
    }

    // Unknown error - default Indonesian message
    return new TransactionServiceError('Terjadi kesalahan tidak terduga', undefined, error);
  },
};

/**
 * Transaction Hooks - React hooks for transaction operations
 * These can be used in components to process transactions
 */
export const useTransactionService = () => {
  return {
    createTransaction: TransactionService.createTransaction,
  };
};
