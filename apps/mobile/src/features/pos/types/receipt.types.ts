/**
 * Receipt Types for Point of Sale
 * Defines receipt data structures for thermal printer integration
 */

import { PaymentMethod, EWalletType } from './payment.types';

/**
 * Paper Width Type
 * Thermal printer paper width in millimeters
 */
export type PaperWidth = 58 | 80;

/**
 * Receipt Item Interface
 * Represents a single line item on the receipt
 */
export interface ReceiptItem {
  name: string;
  quantity: number;
  unitPrice: string;  // Currency as string for precision
  subtotal: string;   // Currency as string for precision
}

/**
 * Payment Details Interface
 * Discriminated union using 'method' field as discriminator
 * Enables type-safe payment data handling with TypeScript type narrowing
 */
export interface PaymentDetails {
  method: PaymentMethod;
  cashDetails?: {
    change: string;  // Currency as string for precision
  };
  transferDetails?: {
    accountName: string;
    referenceNumber: string;
  };
  ewalletDetails?: {
    walletType: EWalletType;
    confirmationInput: string;
  };
}

/**
 * Receipt Data Interface
 * Complete receipt data structure for thermal printer
 */
export interface ReceiptData {
  transactionNumber: string;
  transactionDate: string;  // ISO 8601 format
  pharmacyName: string;
  pharmacyAddress: string;
  pharmacyPhone: string;
  items: ReceiptItem[];
  subtotal: string;
  tax?: string;  // Optional tax field
  total: string;
  payment: PaymentDetails;
  paperWidth: PaperWidth;
}

/**
 * Receipt Configuration Interface
 * Pharmacy configuration for receipt printing
 */
export interface ReceiptConfig {
  pharmacyName: string;
  pharmacyAddress: string;
  pharmacyPhone: string;
  defaultPaperWidth: PaperWidth;
}
