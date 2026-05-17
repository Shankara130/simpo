/**
 * TransactionHistoryService - API integration for transaction history
 * Handles transaction listing, detail retrieval, and API communication
 * Story 3.7: Transaction History View
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import axios, { AxiosError } from 'axios';
import {
  TransactionSummary,
  TransactionDetail,
  TransactionFilters,
  TransactionListResponse,
} from '../types/transactionHistory.types';

// API base URL - configure for different environments
const API_BASE_URL = __DEV__
  ? 'http://localhost:8080' // Development - local backend
  : 'https://api.simpo.id';  // Production - actual API endpoint

// API version prefix
const API_VERSION = '/api/v1';

// Full API base URL
const FULL_API_URL = `${API_BASE_URL}${API_VERSION}`;

// Storage key for filter persistence
const FILTERS_STORAGE_KEY = 'transaction_history_filters';

/**
 * API Error class for handling service-specific errors
 */
export class TransactionHistoryServiceError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public originalError?: any
  ) {
    super(message);
    this.name = 'TransactionHistoryServiceError';
  }
}

/**
 * TransactionHistoryService - Transaction history API operations
 */
export const TransactionHistoryService = {
  /**
   * Fetch list of transactions with pagination and filters
   * @param filters - Filter options (startDate, endDate, status)
   * @param page - Page number (default: 1)
   * @param limit - Items per page (default: 20)
   * @returns TransactionListResponse with data and pagination metadata
   */
  getTransactions: async (
    filters?: TransactionFilters,
    page = 1,
    limit = 20
  ): Promise<TransactionListResponse> => {
    try {
      // Validate and sanitize pagination parameters
      if (page < 1 || !Number.isFinite(page)) {
        page = 1;
      }

      if (limit < 1 || !Number.isFinite(limit)) {
        limit = 20;
      }

      // Enforce reasonable maximum limit
      if (limit > 100) {
        limit = 100;
      }

      // Get JWT token from storage
      const token = await AsyncStorage.getItem('jwt_token');
      if (!token) {
        throw new TransactionHistoryServiceError(
          'Authentication required. Please log in.',
          401
        );
      }

      // Build query parameters
      const params: Record<string, any> = {
        page,
        limit,
      };

      // Add date range filters if provided
      if (filters?.startDate) {
        params.startDate = filters.startDate.toISOString().split('T')[0]; // YYYY-MM-DD format
      }

      if (filters?.endDate) {
        params.endDate = filters.endDate.toISOString().split('T')[0];
      }

      // Add status filter if not 'ALL'
      if (filters?.status && filters.status !== 'ALL') {
        params.status = filters.status;
      }

      const response = await axios.get<TransactionListResponse>(
        `${FULL_API_URL}/transactions`,
        {
          params,
          headers: {
            Authorization: `Bearer ${token}`,
          },
          timeout: 10000, // 10 second timeout
        }
      );

      return response.data;
    } catch (error) {
      throw TransactionHistoryService._handleError(
        error,
        'Gagal memuat riwayat transaksi'
      );
    }
  },

  /**
   * Get transaction details by ID
   * @param transactionId - Transaction ID
   * @returns TransactionDetail with items, cashier, branch, and receipt data
   */
  getTransactionById: async (transactionId: number): Promise<TransactionDetail> => {
    try {
      if (!transactionId || transactionId <= 0) {
        throw new TransactionHistoryServiceError(
          'ID transaksi tidak valid',
          400
        );
      }

      // Get JWT token from storage
      const token = await AsyncStorage.getItem('jwt_token');
      if (!token) {
        throw new TransactionHistoryServiceError(
          'Authentication required. Please log in.',
          401
        );
      }

      const response = await axios.get<TransactionDetail>(
        `${FULL_API_URL}/transactions/${transactionId}`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
          timeout: 10000,
        }
      );

      return response.data;
    } catch (error) {
      throw TransactionHistoryService._handleError(
        error,
        'Gagal memuat detail transaksi'
      );
    }
  },

  /**
   * Save filters to AsyncStorage for persistence
   * @param filters - Filter options to save
   */
  saveFilters: async (filters: TransactionFilters): Promise<void> => {
    try {
      const filtersToSave = {
        startDate: filters.startDate?.toISOString() || null,
        endDate: filters.endDate?.toISOString() || null,
        status: filters.status,
      };
      await AsyncStorage.setItem(
        FILTERS_STORAGE_KEY,
        JSON.stringify(filtersToSave)
      );
    } catch (error) {
      console.error('Failed to save filters:', error);
      // Non-critical error, don't throw
    }
  },

  /**
   * Load filters from AsyncStorage
   * @returns Saved filters or default filters
   */
  loadFilters: async (): Promise<TransactionFilters> => {
    try {
      const savedFilters = await AsyncStorage.getItem(FILTERS_STORAGE_KEY);
      if (!savedFilters) {
        return TransactionHistoryService._getDefaultFilters();
      }

      const parsed = JSON.parse(savedFilters);
      return {
        startDate: parsed.startDate ? new Date(parsed.startDate) : null,
        endDate: parsed.endDate ? new Date(parsed.endDate) : null,
        status: parsed.status || 'ALL',
      };
    } catch (error) {
      console.error('Failed to load filters:', error);
      return TransactionHistoryService._getDefaultFilters();
    }
  },

  /**
   * Clear saved filters
   */
  clearFilters: async (): Promise<void> => {
    try {
      await AsyncStorage.removeItem(FILTERS_STORAGE_KEY);
    } catch (error) {
      console.error('Failed to clear filters:', error);
      // Non-critical error, don't throw
    }
  },

  /**
   * Get default filters (today, all statuses)
   * @private
   */
  _getDefaultFilters: (): TransactionFilters => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    return {
      startDate: today,
      endDate: null,
      status: 'ALL',
    };
  },

  /**
   * Handle API errors and convert to TransactionHistoryServiceError
   * Maps RFC 7807 error responses to user-friendly Indonesian messages
   * @private
   */
  _handleError: (
    error: unknown,
    defaultMessage: string
  ): TransactionHistoryServiceError => {
    if (error instanceof TransactionHistoryServiceError) {
      return error;
    }

    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError<any>;

      // Network error (no response)
      if (!axiosError.response) {
        return new TransactionHistoryServiceError(
          'Koneksi internet terputus. Silakan periksa koneksi Anda.',
          undefined,
          axiosError
        );
      }

      // Server responded with error status
      const status = axiosError.response.status;
      const data = axiosError.response.data;

      // Handle RFC 7807 error responses
      if (data?.type && data?.detail) {
        switch (status) {
          case 400:
            return new TransactionHistoryServiceError(
              data.detail || 'Permintaan tidak valid',
              status,
              axiosError
            );
          case 401:
            return new TransactionHistoryServiceError(
              'Sesi Anda telah berakhir. Silakan login kembali.',
              status,
              axiosError
            );
          case 403:
            return new TransactionHistoryServiceError(
              'Anda tidak memiliki akses ke transaksi ini',
              status,
              axiosError
            );
          case 404:
            return new TransactionHistoryServiceError(
              'Transaksi tidak ditemukan',
              status,
              axiosError
            );
          case 500:
            return new TransactionHistoryServiceError(
              'Terjadi kesalahan server. Silakan coba lagi nanti.',
              status,
              axiosError
            );
          default:
            return new TransactionHistoryServiceError(
              data.detail || defaultMessage,
              status,
              axiosError
            );
        }
      }

      // Fallback for non-RFC 7807 responses
      switch (status) {
        case 401:
          return new TransactionHistoryServiceError(
            'Sesi Anda telah berakhir. Silakan login kembali.',
            status,
            axiosError
          );
        case 403:
          return new TransactionHistoryServiceError(
            'Akses ditolak',
            status,
            axiosError
          );
        case 404:
          return new TransactionHistoryServiceError(
            'Transaksi tidak ditemukan',
            status,
            axiosError
          );
        case 500:
          return new TransactionHistoryServiceError(
            'Terjadi kesalahan server. Silakan coba lagi nanti.',
            status,
            axiosError
          );
        default:
          return new TransactionHistoryServiceError(
            defaultMessage,
            status,
            axiosError
          );
      }
    }

    // Unknown error
    return new TransactionHistoryServiceError(defaultMessage, undefined, error);
  },
};

/**
 * Transaction History Hooks - React hooks for transaction history operations
 */
export const useTransactionHistoryService = () => {
  return {
    getTransactions: TransactionHistoryService.getTransactions,
    getTransactionById: TransactionHistoryService.getTransactionById,
    saveFilters: TransactionHistoryService.saveFilters,
    loadFilters: TransactionHistoryService.loadFilters,
    clearFilters: TransactionHistoryService.clearFilters,
  };
};
