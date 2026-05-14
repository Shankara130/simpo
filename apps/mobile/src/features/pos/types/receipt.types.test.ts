/**
 * Receipt Types Tests
 * Test receipt type definitions and discriminated unions
 */

import { PaymentMethod, EWalletType } from './payment.types';
import {
  PaperWidth,
  ReceiptItem,
  PaymentDetails,
  ReceiptData,
  ReceiptConfig,
} from './receipt.types';

describe('Receipt Types', () => {
  describe('PaperWidth', () => {
    it('should accept 58mm as valid paper width', () => {
      const width: PaperWidth = 58;
      expect(width).toBe(58);
    });

    it('should accept 80mm as valid paper width', () => {
      const width: PaperWidth = 80;
      expect(width).toBe(80);
    });
  });

  describe('ReceiptItem', () => {
    it('should create valid receipt item with all required fields', () => {
      const item: ReceiptItem = {
        name: 'Paracetamol 500mg',
        quantity: 2,
        unitPrice: '15000.00',
        subtotal: '30000.00',
      };

      expect(item.name).toBe('Paracetamol 500mg');
      expect(item.quantity).toBe(2);
      expect(item.unitPrice).toBe('15000.00');
      expect(item.subtotal).toBe('30000.00');
    });

    it('should handle Indonesian product names correctly', () => {
      const item: ReceiptItem = {
        name: 'Obat Batuk Herbal',
        quantity: 1,
        unitPrice: '25000.00',
        subtotal: '25000.00',
      };

      expect(item.name).toBe('Obat Batuk Herbal');
    });
  });

  describe('PaymentDetails', () => {
    it('should create cash payment details', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.CASH,
        cashDetails: {
          change: '5000.00',
        },
      };

      expect(payment.method).toBe(PaymentMethod.CASH);
      expect(payment.cashDetails).toBeDefined();
      expect(payment.cashDetails?.change).toBe('5000.00');
    });

    it('should create bank transfer payment details', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.TRANSFER,
        transferDetails: {
          accountName: 'Budi Santoso',
          referenceNumber: 'REF123456',
        },
      };

      expect(payment.method).toBe(PaymentMethod.TRANSFER);
      expect(payment.transferDetails).toBeDefined();
      expect(payment.transferDetails?.accountName).toBe('Budi Santoso');
      expect(payment.transferDetails?.referenceNumber).toBe('REF123456');
    });

    it('should create e-wallet payment details', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.E_WALLET,
        ewalletDetails: {
          walletType: EWalletType.GOPAY,
          confirmationInput: '08123456789',
        },
      };

      expect(payment.method).toBe(PaymentMethod.E_WALLET);
      expect(payment.ewalletDetails).toBeDefined();
      expect(payment.ewalletDetails?.walletType).toBe(EWalletType.GOPAY);
      expect(payment.ewalletDetails?.confirmationInput).toBe('08123456789');
    });

    it('should support OVO e-wallet type', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.E_WALLET,
        ewalletDetails: {
          walletType: EWalletType.OVO,
          confirmationInput: 'CONF123',
        },
      };

      expect(payment.ewalletDetails?.walletType).toBe(EWalletType.OVO);
    });

    it('should support Dana e-wallet type', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.E_WALLET,
        ewalletDetails: {
          walletType: EWalletType.DANA,
          confirmationInput: 'CONF456',
        },
      };

      expect(payment.ewalletDetails?.walletType).toBe(EWalletType.DANA);
    });

    it('should support ShopeePay e-wallet type', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.E_WALLET,
        ewalletDetails: {
          walletType: EWalletType.SHOPEE_PAY,
          confirmationInput: 'CONF789',
        },
      };

      expect(payment.ewalletDetails?.walletType).toBe(EWalletType.SHOPEE_PAY);
    });
  });

  describe('ReceiptData', () => {
    it('should create complete receipt data with all required fields', () => {
      const receiptData: ReceiptData = {
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

      expect(receiptData.transactionNumber).toBe('TRX-20260514-0001');
      expect(receiptData.pharmacyName).toBe('Apotek Sehat');
      expect(receiptData.items).toHaveLength(1);
      expect(receiptData.payment.method).toBe(PaymentMethod.CASH);
      expect(receiptData.paperWidth).toBe(58);
    });

    it('should support receipt with tax', () => {
      const receiptData: ReceiptData = {
        transactionNumber: 'TRX-20260514-0002',
        transactionDate: '2026-05-14T15:30:00+07:00',
        pharmacyName: 'Apotek Sehat',
        pharmacyAddress: 'Jl. Sudirman No. 123, Jakarta',
        pharmacyPhone: '021-1234567',
        items: [],
        subtotal: '30000.00',
        tax: '3000.00',
        total: '33000.00',
        payment: {
          method: PaymentMethod.TRANSFER,
          transferDetails: {
            accountName: 'Budi Santoso',
            referenceNumber: 'REF123',
          },
        },
        paperWidth: 80,
      };

      expect(receiptData.tax).toBe('3000.00');
      expect(receiptData.paperWidth).toBe(80);
    });

    it('should support 58mm paper width', () => {
      const receiptData: ReceiptData = {
        transactionNumber: 'TRX-001',
        transactionDate: '2026-05-14T15:30:00+07:00',
        pharmacyName: 'Apotek',
        pharmacyAddress: 'Jl. Test',
        pharmacyPhone: '123',
        items: [],
        subtotal: '0',
        total: '0',
        payment: {
          method: PaymentMethod.CASH,
          cashDetails: {
            change: '0',
          },
        },
        paperWidth: 58,
      };

      expect(receiptData.paperWidth).toBe(58);
    });

    it('should support 80mm paper width', () => {
      const receiptData: ReceiptData = {
        transactionNumber: 'TRX-001',
        transactionDate: '2026-05-14T15:30:00+07:00',
        pharmacyName: 'Apotek',
        pharmacyAddress: 'Jl. Test',
        pharmacyPhone: '123',
        items: [],
        subtotal: '0',
        total: '0',
        payment: {
          method: PaymentMethod.CASH,
          cashDetails: {
            change: '0',
          },
        },
        paperWidth: 80,
      };

      expect(receiptData.paperWidth).toBe(80);
    });
  });

  describe('ReceiptConfig', () => {
    it('should create pharmacy configuration', () => {
      const config: ReceiptConfig = {
        pharmacyName: 'Apotek Sehat',
        pharmacyAddress: 'Jl. Sudirman No. 123, Jakarta',
        pharmacyPhone: '021-1234567',
        defaultPaperWidth: 58,
      };

      expect(config.pharmacyName).toBe('Apotek Sehat');
      expect(config.pharmacyAddress).toBe('Jl. Sudirman No. 123, Jakarta');
      expect(config.pharmacyPhone).toBe('021-1234567');
      expect(config.defaultPaperWidth).toBe(58);
    });
  });

  describe('Type Guards and Discriminated Unions', () => {
    it('should narrow payment type using method discriminator', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.TRANSFER,
        transferDetails: {
          accountName: 'Test Account',
          referenceNumber: 'REF123',
        },
      };

      if (payment.method === PaymentMethod.TRANSFER) {
        // TypeScript should know transferDetails exists here
        expect(payment.transferDetails?.accountName).toBe('Test Account');
      } else {
        fail('Payment should be TRANSFER type');
      }
    });

    it('should correctly identify cash payments', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.CASH,
        cashDetails: {
          change: '1000.00',
        },
      };

      if (payment.method === PaymentMethod.CASH) {
        expect(payment.cashDetails).toBeDefined();
      } else {
        fail('Payment should be CASH type');
      }
    });

    it('should correctly identify e-wallet payments', () => {
      const payment: PaymentDetails = {
        method: PaymentMethod.E_WALLET,
        ewalletDetails: {
          walletType: EWalletType.DANA,
          confirmationInput: '08123456789',
        },
      };

      if (payment.method === PaymentMethod.E_WALLET) {
        expect(payment.ewalletDetails).toBeDefined();
        expect(payment.ewalletDetails?.walletType).toBe(EWalletType.DANA);
      } else {
        fail('Payment should be E_WALLET type');
      }
    });
  });
});
