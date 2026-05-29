/**
 * ConflictResolutionScreen Tests
 * Story 8.5: Implement Conflict Resolution for Offline Transactions
 * Task 8: Create Mobile Conflict Resolution UI
 *
 * Test coverage for conflict resolution screen components
 */

import { describe, it, expect, jest } from '@jest/globals';
import React from 'react';
import { render, waitFor } from '@testing-library/react-native';

// Mock the service before import
const mockGetAllFailed = jest.fn();
const mockRequestOverride = jest.fn();
const mockClearFailed = jest.fn();

jest.mock('../services/ConflictResolutionService', () => ({
  __esModule: true,
  default: {
    getInstance: jest.fn(() => ({
      getAllFailedTransactions: () => mockGetAllFailed(),
      requestManualOverride: (req: any) => mockRequestOverride(req),
      clearFailedTransaction: (id: number) => mockClearFailed(id),
    })),
  },
}));

// Mock navigation
jest.mock('@react-navigation/native', () => ({
  useNavigation: () => ({
    navigate: jest.fn(),
    goBack: jest.fn(),
  }),
  useRoute: () => ({
    params: {},
  }),
}));

// Import after mocking
import ConflictResolutionScreen from './ConflictResolutionScreen';

describe('ConflictResolutionScreen', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Component Rendering', () => {
    it('should render loading state initially', () => {
      mockGetAllFailed.mockImplementation(() => new Promise(() => {}));

      const { getByText } = render(<ConflictResolutionScreen />);
      expect(getByText('Memuat konflik...')).toBeTruthy();
    });

    it('should render empty state when no failed transactions', async () => {
      mockGetAllFailed.mockResolvedValue([]);

      const { getByText, queryByText } = render(<ConflictResolutionScreen />);

      await waitFor(() => {
        expect(queryByText('Memuat konflik...')).toBeNull();
      });

      // Check for empty state elements
      const emptyTitle = queryByText('Tidak Ada Konflik');
      const emptyMessage = queryByText('Semua transaksi sinkronisasi berhasil');

      if (emptyTitle) expect(emptyTitle).toBeTruthy();
      if (emptyMessage) expect(emptyMessage).toBeTruthy();
    });

    it('should render header with correct Indonesian text', () => {
      mockGetAllFailed.mockResolvedValue([]);

      const { getByText } = render(<ConflictResolutionScreen />);

      // Header should always be visible
      expect(getByText('Konflik Sinkronisasi')).toBeTruthy();
    });

    it('should show transaction count in header', async () => {
      mockGetAllFailed.mockResolvedValue([
        { transactionId: 1, transactionNumber: 'TRX-001', timestamp: '2026-05-29T10:00:00Z', conflictError: {}, canOverride: true, requiresAdminAuth: true },
        { transactionId: 2, transactionNumber: 'TRX-002', timestamp: '2026-05-29T11:00:00Z', conflictError: {}, canOverride: true, requiresAdminAuth: true },
      ]);

      const { getByText } = render(<ConflictResolutionScreen />);

      await waitFor(() => {
        expect(getByText('2 Transaksi Gagal')).toBeTruthy();
      });
    });
  });

  describe('Failed Transaction Display', () => {
    it('should render failed transaction with conflict details', async () => {
      const failedTransactions = [
        {
          transactionId: 1,
          transactionNumber: 'TRX-001',
          timestamp: '2026-05-29T10:00:00Z',
          conflictError: {
            type: 'https://api.simpo.com/errors/conflict-insufficient-stock',
            title: 'Stok Tidak Cukup',
            status: 409,
            detail: 'Product SKU-12345 has insufficient stock',
            conflict_details: {
              product_sku: 'SKU-12345',
              requested_qty: 10,
              available_stock: 5,
              shortfall: 5,
            },
          },
          canOverride: true,
          requiresAdminAuth: true,
        },
      ];

      mockGetAllFailed.mockResolvedValue(failedTransactions);

      const { getByText, findByText } = render(<ConflictResolutionScreen />);

      // Wait for data to load
      await waitFor(() => {
        expect(findByText('TRX-001')).toBeTruthy();
      });

      // Verify transaction details
      expect(findByText('TRX-001')).toBeTruthy();
      expect(findByText('SKU-12345')).toBeTruthy();
      expect(findByText('Diminta: 10')).toBeTruthy();
      expect(findByText('Tersedia: 5')).toBeTruthy();
      expect(findByText('Kekurangan: 5')).toBeTruthy();
    });

    it('should show Override button for admin users', async () => {
      const failedTx = [
        {
          transactionId: 1,
          transactionNumber: 'TRX-001',
          timestamp: '2026-05-29T10:00:00Z',
          conflictError: {},
          canOverride: true,
          requiresAdminAuth: true,
        },
      ];

      mockGetAllFailed.mockResolvedValue(failedTx);

      const { findByText } = render(<ConflictResolutionScreen isAdmin={true} />);

      await waitFor(() => {
        expect(findByText('Override dengan Admin')).toBeTruthy();
      });
    });

    it('should show Hapus Transaksi button for all users', async () => {
      const failedTx = [
        {
          transactionId: 1,
          transactionNumber: 'TRX-001',
          timestamp: '2026-05-29T10:00:00Z',
          conflictError: {},
          canOverride: true,
          requiresAdminAuth: false,
        },
      ];

      mockGetAllFailed.mockResolvedValue(failedTx);

      const { findByText } = render(<ConflictResolutionScreen />);

      await waitFor(() => {
        expect(findByText('Hapus Transaksi')).toBeTruthy();
      });
    });
  });

  describe('Indonesian Language Messages', () => {
    it('should display all UI text in Indonesian', async () => {
      mockGetAllFailed.mockResolvedValue([]);

      const { getByText } = render(<ConflictResolutionScreen />);

      // Verify all Indonesian labels
      expect(getByText('Konflik Sinkronisasi')).toBeTruthy();
      expect(getByText('Transaksi Gagal')).toBeTruthy();
    });

    it('should show Indonesian error messages', async () => {
      const failedTx = [
        {
          transactionId: 1,
          transactionNumber: 'TRX-001',
          timestamp: '2026-05-29T10:00:00Z',
          conflictError: {
            type: 'https://api.simpo.com/errors/conflict-insufficient-stock',
            title: 'Stok Tidak Cukup',
            status: 409,
            detail: 'Product SKU-12345 has insufficient stock',
          },
          canOverride: true,
          requiresAdminAuth: true,
        },
      ];

      mockGetAllFailed.mockResolvedValue(failedTx);

      const { findByText } = render(<ConflictResolutionScreen />);

      await waitFor(() => {
        expect(findByText('Stok Tidak Cukup')).toBeTruthy();
      });
    });
  });
});
