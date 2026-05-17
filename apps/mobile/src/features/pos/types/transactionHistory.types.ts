/**
 * Transaction History Types for Point of Sale
 * Defines interfaces for transaction history and detail views
 * Matches backend API contract from Story 3.7
 */

/**
 * Transaction Summary - Displayed in transaction history list
 * Maps to backend GET /api/v1/transactions response
 */
export interface TransactionSummary {
  id: number;
  transactionNumber: string;
  total: string;
  status: 'COMPLETED' | 'CANCELLED' | 'PENDING';
  createdAt: string;
  paymentMethod: string;
}

/**
 * Transaction Item Detail - Line item in transaction detail
 */
export interface TransactionItemDetail {
  id: number;
  transactionId: number;
  productId: number;
  productName: string;
  productSKU: string;
  quantity: number;
  unitPrice: string;
  subtotal: string;
}

/**
 * Cashier Information
 */
export interface CashierInfo {
  id: number;
  name: string;
}

/**
 * Branch Information
 */
export interface BranchInfo {
  id: number;
  name: string;
}

/**
 * Receipt Data - ESC/POS formatted receipt for reprint
 * Reused from Story 3.5 (receipt.types.ts)
 */
export interface ReceiptData {
  transactionNumber: string;
  transactionDate: string;
  cashierName: string;
  branchName: string;
  customerName: string;
  items: ReceiptItem[];
  subtotal: string;
  taxAmount: string;
  discountAmount: string;
  total: string;
  paymentMethod: string;
}

/**
 * Receipt Item - Line item in receipt
 */
export interface ReceiptItem {
  productName: string;
  sku: string;
  quantity: number;
  unitPrice: string;
  total: string;
}

/**
 * Transaction Detail - Complete transaction data
 * Maps to backend GET /api/v1/transactions/:id response
 */
export interface TransactionDetail extends TransactionSummary {
  items: TransactionItemDetail[];
  cashier: CashierInfo;
  branch: BranchInfo;
  receiptData?: ReceiptData; // Optional, included only for COMPLETED transactions
}

/**
 * Transaction Filters - Filter options for transaction history
 */
export interface TransactionFilters {
  startDate: Date | null;
  endDate: Date | null;
  status: 'ALL' | 'COMPLETED' | 'CANCELLED' | 'PENDING';
}

/**
 * Pagination Metadata
 */
export interface PaginationMeta {
  total: number;
  totalPages: number;
  currentPage: number;
}

/**
 * Transaction List Response - API response for transaction list
 * Maps to backend GET /api/v1/transactions response structure
 */
export interface TransactionListResponse {
  data: TransactionSummary[];
  pagination: PaginationMeta;
}

/**
 * Transaction Filter Presets - Quick date range options
 */
export type DateRangePreset = 'today' | 'yesterday' | 'thisWeek' | 'thisMonth' | 'custom';

/**
 * Transaction History State - Internal state for TransactionHistoryScreen
 */
export interface TransactionHistoryState {
  transactions: TransactionSummary[];
  loading: boolean;
  error: string | null;
  filters: TransactionFilters;
  pagination: {
    currentPage: number;
    totalPages: number;
    hasMore: boolean;
  };
  refreshing: boolean; // For pull-to-refresh
}
