/**
 * TransactionService Tests
 * Story 3.6: Implement Transaction Processing <30 Seconds
 */

import { TransactionService, TransactionServiceError } from './TransactionService';
import { PaymentMethod } from '../types/payment.types';
import AsyncStorage from '@react-native-async-storage/async-storage';
import axios from 'axios';

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
}));

// Mock axios
jest.mock('axios');

// Mock OfflineStorageService for Story 8.1 tests
jest.mock('../../offline/services/OfflineStorageService', () => ({
  default: {
    saveTransaction: jest.fn(),
  },
}));

const mockedAxios = axios as jest.Mocked<typeof axios>;
const mockedAsyncStorage = AsyncStorage as jest.Mocked<typeof AsyncStorage>;

describe('TransactionService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('createTransaction', () => {
    const mockCartItems = [
      {
        productId: 1,
        sku: 'SKU001',
        name: 'Test Product 1',
        price: '10000.00',
        quantity: 2,
        subtotal: '20000.00',
        stockQty: 100,
      },
      {
        productId: 2,
        sku: 'SKU002',
        name: 'Test Product 2',
        price: '15000.00',
        quantity: 1,
        subtotal: '15000.00',
        stockQty: 50,
      },
    ];

    const mockPaymentData = {
      method: PaymentMethod.CASH,
    };

    const mockTransactionResponse = {
      id: 1,
      transactionNumber: 'TRX-20260515-0001',
      cashierId: 100,
      branchId: 1,
      total: '35000.00',
      subtotal: '35000.00',
      tax: '0',
      discount: '0',
      paymentMethod: 'CASH',
      status: 'COMPLETED',
      created_at: '2026-05-15T10:00:00Z',
      updated_at: '2026-05-15T10:00:00Z',
    };

    it('should create transaction successfully', async () => {
      // Arrange
      mockedAsyncStorage.getItem.mockResolvedValueOnce('mock-jwt-token');
      mockedAxios.post.mockResolvedValueOnce({
        data: mockTransactionResponse,
        status: 201,
        statusText: 'Created',
        headers: {},
        config: {} as any,
      });

      // Act
      const result = await TransactionService.createTransaction(
        mockCartItems,
        mockPaymentData
      );

      // Assert
      expect(result).toEqual(mockTransactionResponse);
      expect(mockedAsyncStorage.getItem).toHaveBeenCalledWith('@simpo_jwt_token');
      expect(mockedAxios.post).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/transactions'),
        expect.objectContaining({
          items: expect.arrayContaining([
            expect.objectContaining({
              product_id: 1,
              quantity: 2,
              unit_price: '10000.00',
            }),
          ]),
          payment_method: 'CASH',
        }),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer mock-jwt-token',
          }),
          timeout: 15000,
        })
      );
    });

    it('should throw error for empty cart', async () => {
      // Act & Assert
      await expect(
        TransactionService.createTransaction([], mockPaymentData)
      ).rejects.toThrow(TransactionServiceError);
      await expect(
        TransactionService.createTransaction([], mockPaymentData)
      ).rejects.toThrow('Keranjang tidak boleh kosong');
    });

    it('should throw error when no JWT token found', async () => {
      // Arrange
      mockedAsyncStorage.getItem.mockResolvedValueOnce(null);

      // Act & Assert
      await expect(
        TransactionService.createTransaction(mockCartItems, mockPaymentData)
      ).rejects.toThrow(TransactionServiceError);
    });

    it('should convert bank transfer payment data correctly', async () => {
      // Arrange
      const transferPaymentData = {
        method: PaymentMethod.TRANSFER,
        accountName: 'John Doe',
        referenceNumber: 'REF123',
      };

      mockedAsyncStorage.getItem.mockResolvedValueOnce('mock-jwt-token');
      mockedAxios.post.mockResolvedValueOnce({
        data: mockTransactionResponse,
        status: 201,
        statusText: 'Created',
        headers: {},
        config: {} as any,
      });

      // Act
      await TransactionService.createTransaction(mockCartItems, transferPaymentData);

      // Assert
      expect(mockedAxios.post).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({
          payment_method: 'TRANSFER',
        }),
        expect.anything()
      );
    });

    it('should convert e-wallet payment data correctly', async () => {
      // Arrange
      const ewalletPaymentData = {
        method: PaymentMethod.E_WALLET,
        walletType: 'GOPAY',
        confirmationInput: '123456',
      };

      mockedAsyncStorage.getItem.mockResolvedValueOnce('mock-jwt-token');
      mockedAxios.post.mockResolvedValueOnce({
        data: mockTransactionResponse,
        status: 201,
        statusText: 'Created',
        headers: {},
        config: {} as any,
      });

      // Act
      await TransactionService.createTransaction(mockCartItems, ewalletPaymentData);

      // Assert
      expect(mockedAxios.post).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({
          payment_method: 'E-WALLET',
        }),
        expect.anything()
      );
    });

    it('should map insufficient stock error to Indonesian', async () => {
      // Arrange
      mockedAsyncStorage.getItem.mockResolvedValueOnce('mock-jwt-token');
      mockedAxios.post.mockRejectedValueOnce({
        response: {
          status: 400,
          data: {
            type: 'https://api.simpo.com/errors/transaction-failed',
            title: 'Transaction Failed',
            status: 400,
            detail: 'Insufficient Stock: Test Product (tersedia: 5, diminta: 10)',
            instance: '/api/v1/transactions',
          },
        },
        isAxiosError: true,
        toJSON: () => ({}),
      } as any);

      // Act & Assert
      await expect(
        TransactionService.createTransaction(mockCartItems, mockPaymentData)
      ).rejects.toThrow('Stok tidak mencukupi');
    });

    it('should map network error to Indonesian', async () => {
      // Arrange
      mockedAsyncStorage.getItem.mockResolvedValueOnce('mock-jwt-token');
      mockedAxios.post.mockRejectedValueOnce({
        response: undefined,
        isAxiosError: true,
        toJSON: () => ({}),
      } as any);

      // Act & Assert
      await expect(
        TransactionService.createTransaction(mockCartItems, mockPaymentData)
      ).rejects.toThrow('Koneksi gagal, coba lagi');
    });

    it('should map empty cart error to Indonesian', async () => {
      // Arrange
      mockedAsyncStorage.getItem.mockResolvedValueOnce('mock-jwt-token');
      mockedAxios.post.mockRejectedValueOnce({
        response: {
          status: 400,
          data: {
            type: 'https://api.simpo.com/errors/empty-cart',
            title: 'Cart cannot be empty',
            status: 400,
            detail: 'Cart cannot be empty',
            instance: '/api/v1/transactions',
          },
        },
        isAxiosError: true,
        toJSON: () => ({}),
      } as any);

      // Act & Assert
      await expect(
        TransactionService.createTransaction(mockCartItems, mockPaymentData)
      ).rejects.toThrow('Keranjang tidak boleh kosong');
    });

    // Story 8.1 AC4: Offline transaction tests
    describe('Offline Mode (Story 8.1)', () => {
      const mockOfflineStorageService = require('../../offline/services/OfflineStorageService').default;

      beforeEach(() => {
        jest.clearAllMocks();
        // Mock user data for cashier ID
        mockedAsyncStorage.getItem.mockImplementation((key) => {
          if (key === '@simpo_user') {
            return Promise.resolve(JSON.stringify({ id: 100 }));
          }
          return Promise.resolve(null);
        });
      });

      it('should create offline transaction when isOffline is true', async () => {
        // Arrange
        const mockOfflineResponse = {
          id: 999,
          transactionNumber: 'OFFLINE-1234567890-123',
          cashierId: 100,
          branchId: 0,
          total: '35000.00',
          subtotal: '35000.00',
          tax: '0.00',
          discount: '0.00',
          paymentMethod: 'CASH',
          customerName: undefined,
          status: 'pending_sync',
          created_at: expect.any(String),
          updated_at: expect.any(String),
        };

        mockOfflineStorageService.saveTransaction.mockResolvedValue(mockOfflineResponse);

        // Act
        const result = await TransactionService.createTransaction(
          mockCartItems,
          mockPaymentData,
          '',
          '0',
          '0',
          undefined,
          true // isOffline flag
        );

        // Assert
        expect(result).toEqual(mockOfflineResponse);
        expect(mockOfflineStorageService.saveTransaction).toHaveBeenCalledWith(
          expect.objectContaining({
            payment_method: 'CASH',
            items: expect.any(Array),
          }),
          100 // cashier ID
        );
        // Should NOT call axios
        expect(mockedAxios.post).not.toHaveBeenCalled();
      });

      it('should handle offline storage errors gracefully', async () => {
        // Arrange
        mockOfflineStorageService.saveTransaction.mockRejectedValue(
          new Error('SQLite storage failed')
        );

        // Act & Assert
        await expect(
          TransactionService.createTransaction(
            mockCartItems,
            mockPaymentData,
            '',
            '0',
            '0',
            undefined,
            true // isOffline flag
          )
        ).rejects.toThrow('Gagal menyimpan transaksi offline');
      });

      it('should use default cashier ID (1) when user data not available', async () => {
        // Arrange
        mockedAsyncStorage.getItem.mockImplementation((key) => {
          if (key === '@simpo_user') {
            return Promise.resolve(null); // No user data
          }
          return Promise.resolve(null);
        });

        const mockOfflineResponse = {
          id: 999,
          transactionNumber: 'OFFLINE-1234567890-123',
          cashierId: 1, // Default ID
          branchId: 0,
          total: '35000.00',
          subtotal: '35000.00',
          tax: '0.00',
          discount: '0.00',
          paymentMethod: 'CASH',
          customerName: undefined,
          status: 'pending_sync',
          created_at: expect.any(String),
          updated_at: expect.any(String),
        };

        mockOfflineStorageService.saveTransaction.mockResolvedValue(mockOfflineResponse);

        // Act
        const result = await TransactionService.createTransaction(
          mockCartItems,
          mockPaymentData,
          '',
          '0',
          '0',
          undefined,
          true // isOffline flag
        );

        // Assert
        expect(result.cashierId).toBe(1);
        expect(mockOfflineStorageService.saveTransaction).toHaveBeenCalledWith(
          expect.anything(),
          1 // Default cashier ID
        );
      });
    });
  });
});
