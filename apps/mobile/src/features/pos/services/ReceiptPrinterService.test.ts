/**
 * ReceiptPrinterService Tests
 * Test ESC/POS receipt generation service
 */

import { PaymentMethod, EWalletType } from '../types/payment.types';
import {
  PaperWidth,
  ReceiptItem,
  ReceiptData,
  PaymentDetails,
} from '../types/receipt.types';
import { ReceiptPrinterService } from './ReceiptPrinterService';

describe('ReceiptPrinterService', () => {
  let service: ReceiptPrinterService;

  // Mock receipt data for reuse across tests
  const mockReceiptData: ReceiptData = {
    transactionNumber: 'TRX-20260514-0001',
    transactionDate: '2026-05-14T15:30:00+07:00',
    pharmacyName: 'Apotek Sehat',
    pharmacyAddress: 'Jl. Sudirman No. 123, Jakarta',
    pharmacyPhone: '021-1234567',
    items: [
      {
        name: 'Paracetamol 500mg',
        quantity: 2,
        unitPrice: '15000.00',
        subtotal: '30000.00',
      },
    ],
    subtotal: '30000.00',
    total: '30000.00',
    payment: {
      method: PaymentMethod.CASH,
      cashDetails: {
        change: '0.00',
      },
    },
    paperWidth: 58,
  };

  beforeEach(() => {
    service = new ReceiptPrinterService();
  });

  describe('ESC/POS Commands', () => {
    it('should generate ESC_ALIGN_LEFT command', () => {
      const command = service.ESC_ALIGN_LEFT();
      expect(command).toBe('\x1Ba\x00');
    });

    it('should generate ESC_ALIGN_CENTER command', () => {
      const command = service.ESC_ALIGN_CENTER();
      expect(command).toBe('\x1Ba\x01');
    });

    it('should generate ESC_ALIGN_RIGHT command', () => {
      const command = service.ESC_ALIGN_RIGHT();
      expect(command).toBe('\x1Ba\x02');
    });

    it('should generate BOLD_ON command', () => {
      const command = service.BOLD_ON();
      expect(command).toBe('\x1BE\x01');
    });

    it('should generate BOLD_OFF command', () => {
      const command = service.BOLD_OFF();
      expect(command).toBe('\x1BE\x00');
    });

    it('should generate FULL_CUT command', () => {
      const command = service.FULL_CUT();
      expect(command).toBe('\x1DV\x41\x00');
    });

    it('should generate PARTIAL_CUT command', () => {
      const command = service.PARTIAL_CUT();
      expect(command).toBe('\x1DV\x42\x00');
    });
  });

  describe('formatItems', () => {
    it('should format items for 58mm paper width', () => {
      const items: ReceiptItem[] = [
        {
          name: 'Paracetamol 500mg',
          quantity: 2,
          unitPrice: '15000.00',
          subtotal: '30000.00',
        },
      ];

      const formatted = service.formatItems(items, 58);

      // Item name is truncated for 58mm paper width (16 char limit for names)
      expect(formatted).toContain('Paracetamol');
      expect(formatted).toContain('2');
      expect(formatted).toContain('15000.00');
      expect(formatted).toContain('30000.00');
    });

    it('should format items for 80mm paper width', () => {
      const items: ReceiptItem[] = [
        {
          name: 'Obat Batuk Herbal',
          quantity: 1,
          unitPrice: '25000.00',
          subtotal: '25000.00',
        },
      ];

      const formatted = service.formatItems(items, 80);

      expect(formatted).toContain('Obat Batuk Herbal');
      expect(formatted).toContain('1');
      expect(formatted).toContain('25000.00');
    });

    it('should handle Indonesian text encoding correctly', () => {
      const items: ReceiptItem[] = [
        {
          name: 'Obat Batuk Herbal',
          quantity: 1,
          unitPrice: '15000.00',
          subtotal: '15000.00',
        },
      ];

      const formatted = service.formatItems(items, 58);

      // Should properly encode Indonesian characters (truncated for width constraints)
      expect(formatted).toContain('Obat');  // Truncated part with Indonesian characters
      expect(formatted).toContain('15000.00');  // Price is still correctly formatted
    });

    it('should handle multiple items', () => {
      const items: ReceiptItem[] = [
        {
          name: 'Item 1',
          quantity: 1,
          unitPrice: '10000.00',
          subtotal: '10000.00',
        },
        {
          name: 'Item 2',
          quantity: 2,
          unitPrice: '20000.00',
          subtotal: '40000.00',
        },
      ];

      const formatted = service.formatItems(items, 58);

      expect(formatted).toContain('Item 1');
      expect(formatted).toContain('Item 2');
    });
  });

  describe('formatPayment', () => {
    it('should format cash payment details', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.CASH,
        cashDetails: {
          change: '5000.00',
        },
      };

      const formatted = service.formatPayment(payment);

      expect(formatted).toContain('Tunai');
      expect(formatted).toContain('5000.00');
    });

    it('should format bank transfer payment details', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.TRANSFER,
        transferDetails: {
          accountName: 'Budi Santoso',
          referenceNumber: 'REF123456',
        },
      };

      const formatted = service.formatPayment(payment);

      expect(formatted).toContain('Transfer Bank');
      expect(formatted).toContain('Budi Santoso');
      expect(formatted).toContain('REF123456');
    });

    it('should format GoPay e-wallet payment details', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.E_WALLET,
        ewalletDetails: {
          walletType: EWalletType.GOPAY,
          confirmationInput: '08123456789',
        },
      };

      const formatted = service.formatPayment(payment);

      expect(formatted).toContain('GoPay');
      expect(formatted).toContain('08123456789');
    });

    it('should format OVO e-wallet payment details', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.E_WALLET,
        ewalletDetails: {
          walletType: EWalletType.OVO,
          confirmationInput: 'CONF123',
        },
      };

      const formatted = service.formatPayment(payment);

      expect(formatted).toContain('OVO');
      expect(formatted).toContain('CONF123');
    });

    it('should format Dana e-wallet payment details', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.E_WALLET,
        ewalletDetails: {
          walletType: EWalletType.DANA,
          confirmationInput: 'CONF456',
        },
      };

      const formatted = service.formatPayment(payment);

      expect(formatted).toContain('Dana');
      expect(formatted).toContain('CONF456');
    });

    it('should format ShopeePay e-wallet payment details', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.E_WALLET,
        ewalletDetails: {
          walletType: EWalletType.SHOPEE_PAY,
          confirmationInput: 'CONF789',
        },
      };

      const formatted = service.formatPayment(payment);

      expect(formatted).toContain('ShopeePay');
      expect(formatted).toContain('CONF789');
    });
  });

  describe('generateReceipt', () => {
    it('should generate ESC/POS receipt data buffer', () => {
      const buffer = service.generateReceipt(mockReceiptData);

      expect(buffer).toBeInstanceOf(Uint8Array);
      expect(buffer.length).toBeGreaterThan(0);
    });

    it('should include pharmacy name in receipt', () => {
      const buffer = service.generateReceipt(mockReceiptData);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      expect(decoded).toContain('Apotek Sehat');
    });

    it('should include transaction number in receipt', () => {
      const buffer = service.generateReceipt(mockReceiptData);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      expect(decoded).toContain('TRX-20260514-0001');
    });

    it('should include transaction date in receipt', () => {
      const buffer = service.generateReceipt(mockReceiptData);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      // Date is formatted in Indonesian locale
      expect(decoded).toContain('14/05/2026');
    });

    it('should include items in receipt', () => {
      const buffer = service.generateReceipt(mockReceiptData);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      // Item name may be truncated for width constraints
      expect(decoded).toContain('Paracetamol');
    });

    it('should include payment method in receipt', () => {
      const buffer = service.generateReceipt(mockReceiptData);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      expect(decoded).toContain('Tunai');
    });

    it('should include total amount in receipt', () => {
      const buffer = service.generateReceipt(mockReceiptData);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      expect(decoded).toContain('30000.00');
    });

    it('should include thank you message in Indonesian', () => {
      const buffer = service.generateReceipt(mockReceiptData);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      expect(decoded).toContain('Terima kasih');
    });

    it('should include ESC/POS cut command at the end', () => {
      const buffer = service.generateReceipt(mockReceiptData);

      // Last bytes should be the cut command
      const length = buffer.length;
      expect(buffer[length - 4]).toBe(0x1D); // GS
      expect(buffer[length - 3]).toBe(0x56); // V
      expect(buffer[length - 2]).toBe(0x41); // A
      expect(buffer[length - 1]).toBe(0x00); // nul
    });

    it('should support 58mm paper width', () => {
      const receiptData58mm = { ...mockReceiptData, paperWidth: 58 };
      const buffer = service.generateReceipt(receiptData58mm);

      expect(buffer).toBeInstanceOf(Uint8Array);
      expect(buffer.length).toBeGreaterThan(0);
    });

    it('should support 80mm paper width', () => {
      const receiptData80mm = { ...mockReceiptData, paperWidth: 80 };
      const buffer = service.generateReceipt(receiptData80mm);

      expect(buffer).toBeInstanceOf(Uint8Array);
      expect(buffer.length).toBeGreaterThan(0);
    });

    it('should support receipt with tax', () => {
      const receiptDataWithTax = {
        ...mockReceiptData,
        tax: '3000.00',
        total: '33000.00',
      };
      const buffer = service.generateReceipt(receiptDataWithTax);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      expect(decoded).toContain('3000.00');
      expect(decoded).toContain('33000.00');
    });
  });

  describe('Encoding', () => {
    it('should properly encode Indonesian characters', () => {
      const receiptDataWithIndonesian: ReceiptData = {
        ...mockReceiptData,
        pharmacyName: 'Apotek Sehat',
        pharmacyAddress: 'Jl. Sudirman No. 123, Jakarta',
        items: [
          {
            name: 'Obat Batuk Herbal',
            quantity: 1,
            unitPrice: '15000.00',
            subtotal: '15000.00',
          },
        ],
        subtotal: '15000.00',
        total: '15000.00',
        payment: {
          method: PaymentMethod.CASH,
          cashDetails: {
            change: '0.00',
          },
        },
        paperWidth: 58,
      };

      const buffer = service.generateReceipt(receiptDataWithIndonesian);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      expect(decoded).toContain('Apotek Sehat');
      expect(decoded).toContain('Jl. Sudirman');
      expect(decoded).toContain('Jakarta');
      // Item name may be truncated for width constraints
      expect(decoded).toContain('Obat');
    });

    it('should use UTF-8 encoding for receipt data', () => {
      const buffer = service.generateReceipt(mockReceiptData);

      // Verify buffer can be decoded with UTF-8
      const decoder = new TextDecoder('utf-8');
      const decoded = decoder.decode(buffer);

      expect(decoded).toBeTruthy();
      expect(decoded.length).toBeGreaterThan(0);
    });
  });

  describe('Receipt Layout', () => {
    it('should include header with pharmacy info', () => {
      const buffer = service.generateReceipt(mockReceiptData);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      // Pharmacy name and address should be at the beginning
      expect(decoded.indexOf('Apotek Sehat')).toBeLessThan(decoded.indexOf('Paracetamol'));
    });

    it('should include payment details after items', () => {
      const buffer = service.generateReceipt(mockReceiptData);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      const itemsIndex = decoded.indexOf('Paracetamol');
      const paymentIndex = decoded.indexOf('Tunai');

      expect(paymentIndex).toBeGreaterThan(itemsIndex);
    });

    it('should include total before payment', () => {
      const buffer = service.generateReceipt(mockReceiptData);
      const decoded = new TextDecoder('utf-8').decode(buffer);

      const totalIndex = decoded.indexOf('30000.00'); // total appears in items too
      const paymentIndex = decoded.indexOf('Tunai');

      // Payment should come after the last total
      const lastTotalIndex = decoded.lastIndexOf('30000.00');
      expect(paymentIndex).toBeGreaterThan(lastTotalIndex);
    });
  });
});
