/**
 * Receipt Printer Service
 * Generates ESC/POS formatted receipts for thermal printers
 * Supports 58mm and 80mm paper widths with UTF-8 encoding for Indonesian text
 */

import { PaymentMethod, EWalletType } from '../types/payment.types';
import {
  PaperWidth,
  ReceiptItem,
  ReceiptData,
  PaymentDetails,
} from '../types/receipt.types';

/**
 * ESC/POS Command Constants
 * Standard ESC/POS commands for thermal printer formatting
 */
export const ESC_POS_COMMANDS = {
  // Text Alignment
  ALIGN_LEFT: '\x1Ba\x00',
  ALIGN_CENTER: '\x1Ba\x01',
  ALIGN_RIGHT: '\x1Ba\x02',

  // Bold Mode
  BOLD_ON: '\x1BE\x01',
  BOLD_OFF: '\x1BE\x00',

  // Cut Paper
  FULL_CUT: '\x1DV\x41\x00',
  PARTIAL_CUT: '\x1DV\x42\x00',

  // Line Feed
  LINE_FEED: '\n',
  LINE_FEED_2: '\n\n',
  LINE_FEED_3: '\n\n\n',
};

/**
 * Character limits per paper width
 * 58mm printer: ~32 characters per line
 * 80mm printer: ~48 characters per line
 */
const LINE_WIDTHS: Record<PaperWidth, number> = {
  58: 32,
  80: 48,
};

/**
 * Receipt Printer Service Class
 * Handles ESC/POS receipt generation for thermal printers
 */
export class ReceiptPrinterService {
  /**
   * Generate ESC_ALIGN_LEFT command
   */
  ESC_ALIGN_LEFT(): string {
    return ESC_POS_COMMANDS.ALIGN_LEFT;
  }

  /**
   * Generate ESC_ALIGN_CENTER command
   */
  ESC_ALIGN_CENTER(): string {
    return ESC_POS_COMMANDS.ALIGN_CENTER;
  }

  /**
   * Generate ESC_ALIGN_RIGHT command
   */
  ESC_ALIGN_RIGHT(): string {
    return ESC_POS_COMMANDS.ALIGN_RIGHT;
  }

  /**
   * Generate BOLD_ON command
   */
  BOLD_ON(): string {
    return ESC_POS_COMMANDS.BOLD_ON;
  }

  /**
   * Generate BOLD_OFF command
   */
  BOLD_OFF(): string {
    return ESC_POS_COMMANDS.BOLD_OFF;
  }

  /**
   * Generate FULL_CUT command
   */
  FULL_CUT(): string {
    return ESC_POS_COMMANDS.FULL_CUT;
  }

  /**
   * Generate PARTIAL_CUT command
   */
  PARTIAL_CUT(): string {
    return ESC_POS_COMMANDS.PARTIAL_CUT;
  }

  /**
   * Format receipt items section
   * @param items Cart items to format
   * @param width Paper width (58mm or 80mm)
   * @returns Formatted items string
   */
  formatItems(items: ReceiptItem[], width: PaperWidth): string {
    const maxChars = LINE_WIDTHS[width];
    let formatted = '';

    // Header
    formatted += this.BOLD_ON();
    formatted += 'Item';
    formatted += ' '.repeat(Math.max(1, maxChars - 26)) + 'Jml';
    formatted += ' '.repeat(Math.max(1, 6)) + 'Harga';
    formatted += ' '.repeat(Math.max(1, 6)) + 'Subtotal';
    formatted += this.BOLD_OFF();
    formatted += ESC_POS_COMMANDS.LINE_FEED;

    // Separator line
    formatted += '-'.repeat(maxChars);
    formatted += ESC_POS_COMMANDS.LINE_FEED;

    // Items
    for (const item of items) {
      // Name and quantity
      const name = this.truncateText(item.name, maxChars - 16);
      formatted += name;
      formatted += ' '.repeat(Math.max(1, maxChars - 16 - name.length - 5));
      formatted += item.quantity.toString().padStart(5);
      formatted += ESC_POS_COMMANDS.LINE_FEED;

      // Price and subtotal
      formatted += item.unitPrice.padStart(10);
      formatted += ' '.repeat(Math.max(1, 5));
      formatted += item.subtotal.padStart(10);
      formatted += ESC_POS_COMMANDS.LINE_FEED;
    }

    // Separator line
    formatted += '-'.repeat(maxChars);
    formatted += ESC_POS_COMMANDS.LINE_FEED_2;

    return formatted;
  }

  /**
   * Format payment details section
   * @param payment Payment data to format
   * @returns Formatted payment string
   */
  formatPayment(payment: PaymentDetails): string {
    let formatted = '';

    switch (payment.method) {
      case PaymentMethod.CASH:
        formatted += 'Metode: Tunai';
        if (payment.cashDetails) {
          formatted += ESC_POS_COMMANDS.LINE_FEED;
          formatted += `Kembalian: Rp ${payment.cashDetails.change}`;
        }
        break;

      case PaymentMethod.TRANSFER:
        formatted += 'Metode: Transfer Bank';
        if (payment.transferDetails) {
          formatted += ESC_POS_COMMANDS.LINE_FEED;
          formatted += `Akun: ${payment.transferDetails.accountName}`;
          formatted += ESC_POS_COMMANDS.LINE_FEED;
          formatted += `Ref: ${payment.transferDetails.referenceNumber}`;
        }
        break;

      case PaymentMethod.E_WALLET:
        if (payment.ewalletDetails) {
          const walletName = this.getEWalletName(payment.ewalletDetails.walletType);
          formatted += `Metode: ${walletName}`;
          formatted += ESC_POS_COMMANDS.LINE_FEED;
          formatted += `Konfirmasi: ${payment.ewalletDetails.confirmationInput}`;
        }
        break;
    }

    formatted += ESC_POS_COMMANDS.LINE_FEED_2;

    return formatted;
  }

  /**
   * Generate complete ESC/POS receipt
   * @param receiptData Receipt data to format
   * @returns ESC/POS formatted receipt buffer
   */
  generateReceipt(receiptData: ReceiptData): Uint8Array {
    let receipt = '';

    // ===== HEADER =====
    receipt += this.BOLD_ON();
    receipt += ESC_POS_COMMANDS.ALIGN_CENTER;
    receipt += receiptData.pharmacyName;
    receipt += this.BOLD_OFF();
    receipt += ESC_POS_COMMANDS.LINE_FEED;
    receipt += ESC_POS_COMMANDS.ALIGN_CENTER;
    receipt += receiptData.pharmacyAddress;
    receipt += ESC_POS_COMMANDS.LINE_FEED;
    receipt += ESC_POS_COMMANDS.ALIGN_CENTER;
    receipt += `Telp: ${receiptData.pharmacyPhone}`;
    receipt += ESC_POS_COMMANDS.LINE_FEED_2;

    // ===== TRANSACTION INFO =====
    receipt += ESC_POS_COMMANDS.ALIGN_LEFT;
    receipt += `No. Transaksi: ${receiptData.transactionNumber}`;
    receipt += ESC_POS_COMMANDS.LINE_FEED;

    // Format date for Indonesian locale
    const date = new Date(receiptData.transactionDate);
    const formattedDate = this.formatIndonesianDate(date);
    receipt += `Tanggal/Waktu: ${formattedDate}`;
    receipt += ESC_POS_COMMANDS.LINE_FEED_2;

    // ===== ITEMS =====
    receipt += this.formatItems(receiptData.items, receiptData.paperWidth);

    // ===== TOTALS =====
    receipt += this.BOLD_ON();
    receipt += 'Subtotal:'.padEnd(20) + receiptData.subtotal;
    receipt += this.BOLD_OFF();
    receipt += ESC_POS_COMMANDS.LINE_FEED;

    if (receiptData.tax) {
      receipt += 'Pajak:'.padEnd(20) + receiptData.tax;
      receipt += ESC_POS_COMMANDS.LINE_FEED;
    }

    receipt += this.BOLD_ON();
    receipt += 'TOTAL:'.padEnd(20) + receiptData.total;
    receipt += this.BOLD_OFF();
    receipt += ESC_POS_COMMANDS.LINE_FEED_2;

    // ===== PAYMENT =====
    receipt += this.formatPayment(receiptData.payment);

    // ===== FOOTER =====
    receipt += ESC_POS_COMMANDS.ALIGN_CENTER;
    receipt += '-'.repeat(LINE_WIDTHS[receiptData.paperWidth]);
    receipt += ESC_POS_COMMANDS.LINE_FEED;
    receipt += 'Terima kasih atas kunjungan Anda';
    receipt += ESC_POS_COMMANDS.LINE_FEED;
    receipt += ESC_POS_COMMANDS.LINE_FEED;

    // ===== CUT COMMAND (MUST BE LAST) =====
    receipt += this.FULL_CUT();

    // Encode to UTF-8 buffer
    const encoder = new TextEncoder();
    return encoder.encode(receipt);
  }

  /**
   * Truncate text to fit within line width
   * @param text Text to truncate
   * @param maxLength Maximum characters
   * @returns Truncated text
   */
  private truncateText(text: string, maxLength: number): string {
    if (text.length <= maxLength) {
      return text;
    }
    return text.substring(0, maxLength - 3) + '...';
  }

  /**
   * Format date for Indonesian locale
   * @param date Date to format
   * @returns Formatted date string
   */
  private formatIndonesianDate(date: Date): string {
    const options: Intl.DateTimeFormatOptions = {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
      timeZone: 'Asia/Jakarta',
    };

    return date.toLocaleString('id-ID', options);
  }

  /**
   * Get e-wallet display name
   * @param walletType E-wallet type enum
   * @returns Display name for e-wallet
   */
  private getEWalletName(walletType: EWalletType): string {
    const names: Record<EWalletType, string> = {
      [EWalletType.GOPAY]: 'GoPay',
      [EWalletType.OVO]: 'OVO',
      [EWalletType.DANA]: 'Dana',
      [EWalletType.SHOPEE_PAY]: 'ShopeePay',
    };

    return names[walletType] || walletType;
  }

}
