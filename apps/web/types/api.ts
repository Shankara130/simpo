/**
 * TypeScript type definitions for simpo API responses
 */

/**
 * User role enumeration
 */
export type UserRole = 'Admin' | 'Owner' | 'Cashier';

/**
 * User entity
 */
export interface User {
  id: number;
  email: string;
  username: string;
  role: UserRole;
  branchId?: number;
  status: 'active' | 'inactive';
  createdAt: string;
  updatedAt: string;
}

/**
 * Login request
 */
export interface LoginRequest {
  email: string;
  password: string;
}

/**
 * Login response
 */
export interface LoginResponse {
  token: string;
  user: User;
}

/**
 * Product entity
 */
export interface Product {
  id: number;
  sku: string;
  name: string;
  description?: string;
  stockQty: number;
  price: number;
  expiryDate?: string;
  branchId: number;
  createdAt: string;
  updatedAt: string;
}

/**
 * Transaction entity
 */
export interface Transaction {
  id: number;
  transactionNumber: string;
  cashierId: number;
  total: number;
  paymentMethod: 'CASH' | 'TRANSFER' | 'E_WALLET';
  status: 'completed' | 'pending' | 'cancelled';
  createdAt: string;
  updatedAt: string;
}

/**
 * Transaction item
 */
export interface TransactionItem {
  id: number;
  transactionId: number;
  productId: number;
  quantity: number;
  unitPrice: number;
  subtotal: number;
}

/**
 * Branch entity
 */
export interface Branch {
  id: number;
  name: string;
  address: string;
  phone: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Daily sales report
 */
export interface DailySalesReport {
  date: string;
  totalSales: number;
  transactionCount: number;
  paymentMethodBreakdown: {
    cash: number;
    transfer: number;
    eWallet: number;
  };
  topProducts: Array<{
    productId: number;
    productName: string;
    quantity: number;
    revenue: number;
  }>;
}

/**
 * Profit/loss report
 */
export interface ProfitLossReport {
  period: {
    startDate: string;
    endDate: string;
  };
  totalRevenue: number;
  costOfGoodsSold: number;
  grossProfit: number;
  grossProfitMargin: number;
  breakdown?: {
    byProduct?: Array<{
      categoryId: number;
      categoryName: string;
      revenue: number;
      cogs: number;
      profit: number;
    }>;
    byBranch?: Array<{
      branchId: number;
      branchName: string;
      revenue: number;
      cogs: number;
      profit: number;
    }>;
    byPaymentMethod?: Array<{
      method: string;
      revenue: number;
    }>;
  };
}

/**
 * Paginated list response
 */
export interface PaginatedList<T> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

/**
 * API error response (RFC 7807)
 */
export interface ApiErrorResponse {
  type: string;
  title: string;
  status: number;
  detail: string;
  instance?: string;
}

/**
 * Health check response
 */
export interface HealthCheckResponse {
  status: 'healthy' | 'degraded';
  database: 'connected' | 'disconnected';
  redis: 'connected' | 'disconnected';
  uptime: number;
  version: string;
}
