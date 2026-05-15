/**
 * Transaction Types for Point of Sale
 * Defines interfaces for transaction API requests and responses
 * Matches backend API contract from Story 3.6
 */

/**
 * Sale Item - Line item in a sale request
 * Maps to backend SaleItem structure
 */
export interface SaleItem {
  product_id: number;
  quantity: number;
  unit_price: string; // Decimal as string for precision
}

/**
 * Sale Request - Transaction creation request
 * Maps to backend SaleRequest structure
 */
export interface SaleRequest {
  items: SaleItem[];
  payment_method: string;
  customer_name?: string;
  idempotency_key: string; // CRITICAL-003: Idempotency key to prevent duplicate charges
  tax_amount: string;
  discount_amount: string;
}

/**
 * Transaction Response - Created transaction from backend
 * Maps to backend Transaction model
 */
export interface TransactionResponse {
  id: number;
  transactionNumber: string;
  cashierId: number;
  branchId: number;
  total: string;
  subtotal: string;
  tax: string;
  discount: string;
  paymentMethod: string;
  status: string;
  customerName?: string;
  created_at: string;
  updated_at: string;
  transactionItems?: TransactionItemResponse[];
}

/**
 * Transaction Item Response - Line item in transaction
 */
export interface TransactionItemResponse {
  id: number;
  transactionId: number;
  productId: number;
  quantity: number;
  unitPrice: string;
  subtotal: string;
  productName: string;
  productSKU: string;
}

/**
 * API Error Response (RFC 7807 format)
 * Maps to backend RFC 7807 error structure
 */
export interface TransactionErrorResponse {
  type: string;
  title: string;
  status: number;
  detail: string;
  instance: string;
}

/**
 * Transaction Service Error
 * Custom error class for transaction-specific errors
 */
export class TransactionServiceError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public originalError?: any
  ) {
    super(message);
    this.name = 'TransactionServiceError';
  }
}
