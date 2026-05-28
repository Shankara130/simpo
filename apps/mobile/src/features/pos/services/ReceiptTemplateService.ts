/**
 * Receipt Template Service
 * Generates formatted receipt templates using ESC/POS commands
 * Supports 58mm and 80mm paper widths with Indonesian language
 */

import { ReceiptData, ReceiptItem, PaymentDetails } from '../types/receipt.types';
import { ESCPOSGenerator } from './ESCPOSGenerator';

/**
 * Receipt Template Service Interface
 */
export interface IReceiptTemplateService {
  generateReceipt(data: ReceiptData): Uint8Array;
  generateHeader(data: any): Uint8Array;
  generateTransactionInfo(data: any): Uint8Array;
  generateItemsTable58mm(items: ReceiptItem[]): Uint8Array;
  generateItemsTable80mm(items: ReceiptItem[]): Uint8Array;
  generateTotals(data: any): Uint8Array;
  generatePaymentDetails(payment: PaymentDetails): Uint8Array;
  generateFooter(data: any): Uint8Array;
  generateConfigurableReceipt(data: ReceiptData, config: any): Uint8Array;
}

/**
 * Receipt Configuration Interface
 */
export interface ReceiptConfig {
  showHeader?: boolean;
  showItems?: boolean;
  showTotals?: boolean;
  showPayment?: boolean;
  showFooter?: boolean;
}

/**
 * Receipt Template Service Implementation
 */
export class ReceiptTemplateService implements IReceiptTemplateService {
  private generator: ESCPOSGenerator;

  constructor() {
    this.generator = new ESCPOSGenerator();
  }

  /**
   * Generate complete receipt
   */
  public generateReceipt(data: ReceiptData): Uint8Array {
    const commands: Uint8Array[] = [];

    // Initialize printer
    commands.push(this.generator.initialize());

    // Generate based on paper width
    if (data.paperWidth === 58) {
      return this.generateReceipt58mm(data);
    } else {
      return this.generateReceipt80mm(data);
    }
  }

  /**
   * Generate 58mm receipt
   */
  private generateReceipt58mm(data: ReceiptData): Uint8Array {
    const commands: Uint8Array[] = [];

    // Initialize
    commands.push(this.generator.initialize());

    // Header
    commands.push(this.generator.bold(data.pharmacyName));
    commands.push(this.generator.alignCenter(data.pharmacyName));
    commands.push(this.generator.alignCenter(data.pharmacyAddress));
    commands.push(this.generator.alignCenter(`Telp: ${data.pharmacyPhone}`));
    commands.push(this.generator.lineFeeds(2));

    // Transaction info
    commands.push(this.generator.alignLeft(`No: ${data.transactionNumber}`));
    commands.push(this.generator.alignLeft(`Tgl: ${this.formatDate(data.transactionDate)}`));
    commands.push(this.generator.lineFeeds(2));

    // Items table header
    commands.push(this.generator.lineFeed());
    commands.push(this.generator.bold('Item        Qty    Total'));
    commands.push(this.generator.lineFeed());

    // Items
    data.items.forEach(item => {
      const itemLine = this.formatItem58mm(item);
      commands.push(this.generator.alignLeft(itemLine));
      commands.push(this.generator.lineFeed());
    });

    // Separator
    commands.push(this.generator.lineFeed());

    // Totals
    commands.push(this.generator.alignLeft(`Subtotal:        ${this.formatCurrency(data.subtotal)}`));
    if (data.tax) {
      commands.push(this.generator.alignLeft(`Pajak:           ${this.formatCurrency(data.tax)}`));
    }
    commands.push(this.generator.bold(`TOTAL:           ${this.formatCurrency(data.total)}`));
    commands.push(this.generator.lineFeeds(2));

    // Payment details
    commands.push(this.generator.alignLeft(this.formatPaymentText(data.payment)));
    if (data.payment.method === 'cash' && data.payment.cashDetails) {
      commands.push(this.generator.alignLeft(`Kembalian:       ${this.formatCurrency(data.payment.cashDetails.change)}`));
    }
    commands.push(this.generator.lineFeeds(2));

    // Footer
    commands.push(this.generator.alignCenter('Terima Kasih'));
    commands.push(this.generator.alignCenter(`Untuk informasi: ${data.pharmacyPhone}`));
    commands.push(this.generator.lineFeeds(3));

    // Cut
    commands.push(this.generator.partialCut());

    return this.generator.appendCommands(...commands);
  }

  /**
   * Generate 80mm receipt
   */
  private generateReceipt80mm(data: ReceiptData): Uint8Array {
    const commands: Uint8Array[] = [];

    // Initialize
    commands.push(this.generator.initialize());

    // Header
    commands.push(this.generator.bold(data.pharmacyName));
    commands.push(this.generator.alignCenter(data.pharmacyName));
    commands.push(this.generator.alignCenter(data.pharmacyAddress));
    commands.push(this.generator.alignCenter(`Telp: ${data.pharmacyPhone}`));
    commands.push(this.generator.lineFeeds(2));

    // Transaction info
    commands.push(this.generator.alignLeft(`No Transaksi: ${data.transactionNumber}`));
    commands.push(this.generator.alignLeft(`Tanggal:     ${this.formatDate(data.transactionDate)}`));
    commands.push(this.generator.lineFeeds(2));

    // Items table header
    commands.push(this.generator.bold('Item                  Qty    Price     Total'));
    commands.push(this.generator.lineFeed());

    // Items
    data.items.forEach(item => {
      const itemLine = this.formatItem80mm(item);
      commands.push(this.generator.alignLeft(itemLine));
      commands.push(this.generator.lineFeed());
    });

    // Separator
    commands.push(this.generator.lineFeed());

    // Totals
    commands.push(this.generator.alignLeft(`${' '.repeat(20)}Subtotal: ${this.formatCurrency(data.subtotal)}`));
    if (data.tax) {
      commands.push(this.generator.alignLeft(`${' '.repeat(20)}Pajak:    ${this.formatCurrency(data.tax)}`));
    }
    commands.push(this.generator.bold(`${' '.repeat(20)}TOTAL:    ${this.formatCurrency(data.total)}`));
    commands.push(this.generator.lineFeeds(2));

    // Payment details
    commands.push(this.generator.alignLeft(this.formatPaymentText(data.payment)));
    if (data.payment.method === 'cash' && data.payment.cashDetails) {
      commands.push(this.generator.alignLeft(`${' '.repeat(20)}Kembalian: ${this.formatCurrency(data.payment.cashDetails.change)}`));
    }
    commands.push(this.generator.lineFeeds(2));

    // Footer
    commands.push(this.generator.alignCenter('Terima Kasih'));
    commands.push(this.generator.alignCenter(`Untuk informasi: ${data.pharmacyPhone}`));
    commands.push(this.generator.lineFeeds(3));

    // Cut
    commands.push(this.generator.partialCut());

    return this.generator.appendCommands(...commands);
  }

  /**
   * Generate header section
   */
  public generateHeader(data: any): Uint8Array {
    const commands: Uint8Array[] = [];
    commands.push(this.generator.initialize());
    commands.push(this.generator.bold(data.pharmacyName));
    commands.push(this.generator.alignCenter(data.pharmacyName));
    commands.push(this.generator.alignCenter(data.pharmacyAddress));
    commands.push(this.generator.alignCenter(`Telp: ${data.pharmacyPhone}`));
    commands.push(this.generator.lineFeeds(2));

    return this.generator.appendCommands(...commands);
  }

  /**
   * Generate transaction info section
   */
  public generateTransactionInfo(data: any): Uint8Array {
    const commands: Uint8Array[] = [];
    commands.push(this.generator.alignLeft(`No: ${data.transactionNumber}`));
    commands.push(this.generator.alignLeft(`Tgl: ${this.formatDate(data.transactionDate)}`));
    commands.push(this.generator.lineFeeds(2));

    return this.generator.appendCommands(...commands);
  }

  /**
   * Generate items table for 58mm
   */
  public generateItemsTable58mm(items: ReceiptItem[]): Uint8Array {
    const commands: Uint8Array[] = [];

    if (items.length === 0) {
      return this.generator.appendCommands(...commands);
    }

    commands.push(this.generator.bold('Item        Qty    Total'));
    commands.push(this.generator.lineFeed());

    items.forEach(item => {
      const itemLine = this.formatItem58mm(item);
      commands.push(this.generator.alignLeft(itemLine));
      commands.push(this.generator.lineFeed());
    });

    return this.generator.appendCommands(...commands);
  }

  /**
   * Generate items table for 80mm
   */
  public generateItemsTable80mm(items: ReceiptItem[]): Uint8Array {
    const commands: Uint8Array[] = [];

    if (items.length === 0) {
      return this.generator.appendCommands(...commands);
    }

    commands.push(this.generator.bold('Item                  Qty    Price     Total'));
    commands.push(this.generator.lineFeed());

    items.forEach(item => {
      const itemLine = this.formatItem80mm(item);
      commands.push(this.generator.alignLeft(itemLine));
      commands.push(this.generator.lineFeed());
    });

    return this.generator.appendCommands(...commands);
  }

  /**
   * Generate totals section
   */
  public generateTotals(data: any): Uint8Array {
    const commands: Uint8Array[] = [];

    commands.push(this.generator.alignLeft(`Subtotal:        ${this.formatCurrency(data.subtotal)}`));
    if (data.tax) {
      commands.push(this.generator.alignLeft(`Pajak:           ${this.formatCurrency(data.tax)}`));
    }
    commands.push(this.generator.bold(`TOTAL:           ${this.formatCurrency(data.total)}`));
    commands.push(this.generator.lineFeeds(2));

    return this.generator.appendCommands(...commands);
  }

  /**
   * Generate payment details section
   */
  public generatePaymentDetails(payment: PaymentDetails): Uint8Array {
    const commands: Uint8Array[] = [];

    commands.push(this.generator.alignLeft(this.formatPaymentText(payment)));

    if (payment.method === 'cash' && payment.cashDetails) {
      commands.push(this.generator.alignLeft(`Kembalian:       ${this.formatCurrency(payment.cashDetails.change)}`));
    }

    if (payment.method === 'transfer' && payment.transferDetails) {
      commands.push(this.generator.alignLeft(`Ke: ${payment.transferDetails.accountName}`));
      commands.push(this.generator.alignLeft(`Ref: ${payment.transferDetails.referenceNumber}`));
    }

    if (payment.method === 'ewallet' && payment.ewalletDetails) {
      commands.push(this.generator.alignLeft(`Via: ${payment.ewalletDetails.walletType}`));
      commands.push(this.generator.alignLeft(`Konfirmasi: ${payment.ewalletDetails.confirmationInput}`));
    }

    commands.push(this.generator.lineFeeds(2));

    return this.generator.appendCommands(...commands);
  }

  /**
   * Generate footer section
   */
  public generateFooter(data: any): Uint8Array {
    const commands: Uint8Array[] = [];

    commands.push(this.generator.alignCenter('Terima Kasih'));
    commands.push(this.generator.alignCenter(`Untuk informasi: ${data.pharmacyPhone}`));
    commands.push(this.generator.lineFeeds(3));
    commands.push(this.generator.partialCut());

    return this.generator.appendCommands(...commands);
  }

  /**
   * Generate configurable receipt
   */
  public generateConfigurableReceipt(data: ReceiptData, config: ReceiptConfig): Uint8Array {
    const commands: Uint8Array[] = [];

    // Initialize
    commands.push(this.generator.initialize());

    // Add sections based on configuration
    if (config.showHeader !== false) {
      commands.push(this.generateHeader(data));
    }

    if (config.showItems !== false) {
      if (data.paperWidth === 58) {
        commands.push(this.generateItemsTable58mm(data.items));
      } else {
        commands.push(this.generateItemsTable80mm(data.items));
      }
    }

    if (config.showTotals !== false) {
      const totalsData = {
        subtotal: data.subtotal,
        tax: data.tax,
        total: data.total,
      };
      commands.push(this.generateTotals(totalsData));
    }

    if (config.showPayment !== false) {
      commands.push(this.generatePaymentDetails(data.payment));
    }

    if (config.showFooter !== false) {
      commands.push(this.generateFooter(data));
    }

    return this.generator.appendCommands(...commands);
  }

  /**
   * Format item for 58mm receipt
   */
  private formatItem58mm(item: ReceiptItem): string {
    const name = item.name.substring(0, 12); // Limit to 12 chars
    const paddedName = name.padEnd(12, ' ');
    const qty = item.quantity.toString().padStart(4, ' ');
    const subtotal = this.formatCurrency(item.subtotal).padStart(8, ' ');
    return `${paddedName}${qty}${subtotal}`;
  }

  /**
   * Format item for 80mm receipt
   */
  private formatItem80mm(item: ReceiptItem): string {
    const name = item.name.substring(0, 20); // Limit to 20 chars
    const paddedName = name.padEnd(20, ' ');
    const qty = item.quantity.toString().padStart(4, ' ');
    const price = this.formatCurrency(item.unitPrice).padStart(8, ' ');
    const subtotal = this.formatCurrency(item.subtotal).padStart(8, ' ');
    return `${paddedName}${qty}${price}${subtotal}`;
  }

  /**
   * Format payment text
   */
  private formatPaymentText(payment: PaymentDetails): string {
    const methodMap: Record<string, string> = {
      cash: 'Tunai',
      transfer: 'Transfer',
      ewallet: 'E-Wallet',
    };

    const methodName = methodMap[payment.method] || payment.method;
    return `Pembayaran: ${methodName}`;
  }

  /**
   * Format currency in Indonesian format
   */
  private formatCurrency(amount: string): string {
    // Remove any existing formatting and add Indonesian format
    const numericAmount = parseInt(amount.replace(/[^\d]/g, ''), 10);
    return `Rp ${numericAmount.toLocaleString('id-ID')}`;
  }

  /**
   * Format date in Indonesian timezone
   */
  private formatDate(dateString: string): string {
    try {
      const date = new Date(dateString);
      const options: Intl.DateTimeFormatOptions = {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        timeZone: 'Asia/Jakarta',
      };

      return date.toLocaleString('id-ID', options);
    } catch {
      return dateString;
    }
  }

  /**
   * Set paper width
   */
  public setPaperWidth(width: 58 | 80): void {
    this.generator.setPaperWidth(width);
  }

  /**
   * Get current paper width
   */
  public getPaperWidth(): 58 | 80 {
    return this.generator.getPaperWidth() as 58 | 80;
  }
}

// Export for convenience
export default ReceiptTemplateService;
