/**
 * Payment Types for Point of Sale
 * Defines payment method enums and discriminated union types for type-safe payment data
 */

/**
 * Payment Method Enum
 * Maps to backend API payment method values
 */
export enum PaymentMethod {
  CASH = 'CASH',
  TRANSFER = 'TRANSFER',
  E_WALLET = 'E_WALLET',
}

/**
 * E-Wallet Type Enum
 * Indonesian e-wallet providers
 */
export enum EWalletType {
  GOPAY = 'GOPAY',
  OVO = 'OVO',
  DANA = 'DANA',
  SHOPEE_PAY = 'SHOPEE_PAY',
}

/**
 * Cash Payment Data
 * No additional fields required for cash payments
 */
export interface CashPaymentData {
  method: PaymentMethod.CASH;
}

/**
 * Bank Transfer Payment Data
 * Requires account name and reference number
 */
export interface BankTransferPaymentData {
  method: PaymentMethod.TRANSFER;
  accountName: string;
  referenceNumber: string;
}

/**
 * E-Wallet Payment Data
 * Requires wallet type and confirmation input
 */
export interface EWalletPaymentData {
  method: PaymentMethod.E_WALLET;
  walletType: EWalletType;
  confirmationInput: string;
}

/**
 * Payment Data Union Type
 * Discriminated union using 'method' field as discriminator
 * Enables type-safe payment data handling with TypeScript type narrowing
 */
export type PaymentData = CashPaymentData | BankTransferPaymentData | EWalletPaymentData;

/**
 * Payment Selection Interface
 * Combines payment data with validation state
 */
export interface PaymentSelection {
  paymentData: PaymentData;
  isValid: boolean;
}
