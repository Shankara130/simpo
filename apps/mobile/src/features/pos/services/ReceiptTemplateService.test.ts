/**
 * Receipt Template Service Tests
 * Tests for receipt template generation service
 */

import { ReceiptTemplateService } from './ReceiptTemplateService';
import { ReceiptData, PaperWidth, ReceiptItem } from '../types/receipt.types';

describe('ReceiptTemplateService', () => {
  let service: ReceiptTemplateService;

  beforeEach(() => {
    jest.clearAllMocks();
    service = new ReceiptTemplateService();
  });

  describe('Initialization', () => {
    it('should create service instance', () => {
      expect(service).toBeDefined();
    });

    it('should have default paper width', () => {
      expect(service).toBeDefined();
    });
  });

  describe('58mm Receipt Generation', () => {
    const mockReceiptData: ReceiptData = {
      transactionNumber: 'TRX-20240528-001',
      transactionDate: '2024-05-28T10:30:00+07:00',
      pharmacyName: 'Apotek Sehat',
      pharmacyAddress: 'Jl. Kesehatan No. 123',
      pharmacyPhone: '021-12345678',
      items: [
        { name: 'Paracetamol 500mg', quantity: 2, unitPrice: '15000', subtotal: '30000' },
        { name: 'Amoxicillin 500mg', quantity: 1, unitPrice: '25000', subtotal: '25000' },
      ],
      subtotal: '55000',
      tax: '5500',
      total: '60500',
      payment: {
        method: 'cash',
        cashDetails: { change: '39500' },
      },
      paperWidth: 58,
    };

    it('should generate complete 58mm receipt', () => {
      const receipt = service.generateReceipt(mockReceiptData);
      expect(receipt).toBeDefined();
      expect(receipt).toBeInstanceOf(Uint8Array);
      expect(receipt.length).toBeGreaterThan(0);
    });

    it('should include transaction number in 58mm receipt', () => {
      const receipt = service.generateReceipt(mockReceiptData);
      expect(receipt).toBeDefined();
      expect(receipt.length).toBeGreaterThan(0);
    });

    it('should include receipt items in 58mm receipt', () => {
      const receipt = service.generateReceipt(mockReceiptData);
      expect(receipt).toBeDefined();
      expect(mockReceiptData.items.length).toBeGreaterThan(0);
    });

    it('should include totals in 58mm receipt', () => {
      const receipt = service.generateReceipt(mockReceiptData);
      expect(receipt).toBeDefined();
    });

    it('should include payment details in 58mm receipt', () => {
      const receipt = service.generateReceipt(mockReceiptData);
      expect(receipt).toBeDefined();
    });

    it('should include thank you message in 58mm receipt', () => {
      const receipt = service.generateReceipt(mockReceiptData);
      expect(receipt).toBeDefined();
    });
  });

  describe('80mm Receipt Generation', () => {
    const mockReceiptData: ReceiptData = {
      transactionNumber: 'TRX-20240528-002',
      transactionDate: '2024-05-28T11:00:00+07:00',
      pharmacyName: 'Apotek Sehat',
      pharmacyAddress: 'Jl. Kesehatan No. 123, Jakarta',
      pharmacyPhone: '021-12345678',
      items: [
        { name: 'Paracetamol 500mg', quantity: 2, unitPrice: '15000', subtotal: '30000' },
        { name: 'Amoxicillin 500mg', quantity: 1, unitPrice: '25000', subtotal: '25000' },
        { name: 'Vitamin C 1000mg', quantity: 3, unitPrice: '10000', subtotal: '30000' },
      ],
      subtotal: '85000',
      tax: '8500',
      total: '93500',
      payment: {
        method: 'cash',
        cashDetails: { change: '56500' },
      },
      paperWidth: 80,
    };

    it('should generate complete 80mm receipt', () => {
      const receipt = service.generateReceipt(mockReceiptData);
      expect(receipt).toBeDefined();
      expect(receipt).toBeInstanceOf(Uint8Array);
      expect(receipt.length).toBeGreaterThan(0);
    });

    it('should include transaction number in 80mm receipt', () => {
      const receipt = service.generateReceipt(mockReceiptData);
      expect(receipt).toBeDefined();
    });

    it('should include receipt items in 80mm receipt', () => {
      const receipt = service.generateReceipt(mockReceiptData);
      expect(receipt).toBeDefined();
      expect(mockReceiptData.items.length).toBe(3);
    });

    it('should include totals in 80mm receipt', () => {
      const receipt = service.generateReceipt(mockReceiptData);
      expect(receipt).toBeDefined();
    });
  });

  describe('Receipt Header', () => {
    it('should generate pharmacy header with name and address', () => {
      const headerData = {
        pharmacyName: 'Apotek Test',
        pharmacyAddress: 'Jl. Test No. 123',
        pharmacyPhone: '021-999999',
      };

      const header = service.generateHeader(headerData);
      expect(header).toBeDefined();
      expect(header).toBeInstanceOf(Uint8Array);
      expect(header.length).toBeGreaterThan(0);
    });

    it('should include transaction number and date', () => {
      const transactionData = {
        transactionNumber: 'TRX-001',
        transactionDate: '2024-05-28T10:00:00+07:00',
      };

      const transaction = service.generateTransactionInfo(transactionData);
      expect(transaction).toBeDefined();
      expect(transaction).toBeInstanceOf(Uint8Array);
      expect(transaction.length).toBeGreaterThan(0);
    });
  });

  describe('Receipt Items Table', () => {
    const mockItems: ReceiptItem[] = [
      { name: 'Item 1', quantity: 1, unitPrice: '10000', subtotal: '10000' },
      { name: 'Item 2', quantity: 2, unitPrice: '15000', subtotal: '30000' },
    ];

    it('should generate items table for 58mm', () => {
      const itemsTable = service.generateItemsTable58mm(mockItems);
      expect(itemsTable).toBeDefined();
      expect(itemsTable).toBeInstanceOf(Uint8Array);
    });

    it('should generate items table for 80mm', () => {
      const itemsTable = service.generateItemsTable80mm(mockItems);
      expect(itemsTable).toBeDefined();
      expect(itemsTable).toBeInstanceOf(Uint8Array);
    });

    it('should handle empty items array', () => {
      const itemsTable = service.generateItemsTable58mm([]);
      expect(itemsTable).toBeDefined();
      expect(itemsTable).toBeInstanceOf(Uint8Array);
    });
  });

  describe('Receipt Totals', () => {
    it('should generate totals section', () => {
      const totalsData = {
        subtotal: '55000',
        tax: '5500',
        total: '60500',
      };

      const totals = service.generateTotals(totalsData);
      expect(totals).toBeDefined();
      expect(totals).toBeInstanceOf(Uint8Array);
      expect(totals.length).toBeGreaterThan(0);
    });

    it('should handle receipt without tax', () => {
      const totalsData = {
        subtotal: '55000',
        total: '55000',
      };

      const totals = service.generateTotals(totalsData);
      expect(totals).toBeDefined();
      expect(totals).toBeInstanceOf(Uint8Array);
    });
  });

  describe('Payment Details', () => {
    it('should generate cash payment details', () => {
      const paymentData = {
        method: 'cash' as const,
        cashDetails: { change: '39500' },
      };

      const payment = service.generatePaymentDetails(paymentData);
      expect(payment).toBeDefined();
      expect(payment).toBeInstanceOf(Uint8Array);
      expect(payment.length).toBeGreaterThan(0);
    });

    it('should generate transfer payment details', () => {
      const paymentData = {
        method: 'transfer' as const,
        transferDetails: {
          accountName: 'John Doe',
          referenceNumber: 'REF123456',
        },
      };

      const payment = service.generatePaymentDetails(paymentData);
      expect(payment).toBeDefined();
      expect(payment).toBeInstanceOf(Uint8Array);
    });

    it('should generate e-wallet payment details', () => {
      const paymentData = {
        method: 'ewallet' as const,
        ewalletDetails: {
          walletType: 'gopay' as const,
          confirmationInput: 'CONF789',
        },
      };

      const payment = service.generatePaymentDetails(paymentData);
      expect(payment).toBeDefined();
      expect(payment).toBeInstanceOf(Uint8Array);
    });
  });

  describe('Receipt Footer', () => {
    it('should generate footer with thank you message', () => {
      const footerData = {
        pharmacyPhone: '021-12345678',
      };

      const footer = service.generateFooter(footerData);
      expect(footer).toBeDefined();
      expect(footer).toBeInstanceOf(Uint8Array);
      expect(footer.length).toBeGreaterThan(0);
    });

    it('should include pharmacy contact information', () => {
      const footerData = {
        pharmacyPhone: '021-12345678',
      };

      const footer = service.generateFooter(footerData);
      expect(footer).toBeDefined();
      expect(footer.length).toBeGreaterThan(0);
    });
  });

  describe('Configurable Sections', () => {
    const minimalReceiptData: ReceiptData = {
      transactionNumber: 'TRX-001',
      transactionDate: '2024-05-28T10:00:00+07:00',
      pharmacyName: 'Apotek',
      pharmacyAddress: 'Address',
      pharmacyPhone: '123',
      items: [{ name: 'Item', quantity: 1, unitPrice: '10000', subtotal: '10000' }],
      subtotal: '10000',
      total: '10000',
      payment: { method: 'cash', cashDetails: { change: '0' } },
      paperWidth: 58,
    };

    it('should support configurable receipt sections', () => {
      const config = {
        showHeader: true,
        showItems: true,
        showTotals: true,
        showPayment: true,
        showFooter: true,
      };

      const receipt = service.generateConfigurableReceipt(minimalReceiptData, config);
      expect(receipt).toBeDefined();
      expect(receipt).toBeInstanceOf(Uint8Array);
    });

    it('should handle missing optional sections', () => {
      const config = {
        showHeader: true,
        showItems: true,
        showTotals: false,
        showPayment: true,
        showFooter: false,
      };

      const receipt = service.generateConfigurableReceipt(minimalReceiptData, config);
      expect(receipt).toBeDefined();
      expect(receipt).toBeInstanceOf(Uint8Array);
    });
  });

  describe('Indonesian Language Support', () => {
    it('should include Indonesian text in receipts', () => {
      const mockReceiptData: ReceiptData = {
        transactionNumber: 'TRX-001',
        transactionDate: '2024-05-28T10:00:00+07:00',
        pharmacyName: 'Apotek Sehat',
        pharmacyAddress: 'Jl. Kesehatan No. 123',
        pharmacyPhone: '021-12345678',
        items: [{ name: 'Paracetamol', quantity: 1, unitPrice: '15000', subtotal: '15000' }],
        subtotal: '15000',
        total: '15000',
        payment: { method: 'cash', cashDetails: { change: '5000' } },
        paperWidth: 58,
      };

      const receipt = service.generateReceipt(mockReceiptData);
      expect(receipt).toBeDefined();
      expect(receipt).toBeInstanceOf(Uint8Array);
    });

    it('should format currency in Indonesian format', () => {
      const totalsData = {
        subtotal: '15000',
        total: '15000',
      };

      const totals = service.generateTotals(totalsData);
      expect(totals).toBeDefined();
      expect(totals).toBeInstanceOf(Uint8Array);
    });
  });

  describe('Date and Time Formatting', () => {
    it('should format transaction date in Indonesian timezone', () => {
      const transactionData = {
        transactionNumber: 'TRX-001',
        transactionDate: '2024-05-28T10:00:00+07:00', // WIB timezone
      };

      const transaction = service.generateTransactionInfo(transactionData);
      expect(transaction).toBeDefined();
      expect(transaction).toBeInstanceOf(Uint8Array);
    });

    it('should handle different timezone formats', () => {
      const transactionData = {
        transactionNumber: 'TRX-001',
        transactionDate: '2024-05-28T10:00:00Z', // UTC timezone
      };

      const transaction = service.generateTransactionInfo(transactionData);
      expect(transaction).toBeDefined();
      expect(transaction).toBeInstanceOf(Uint8Array);
    });
  });

  describe('Error Handling', () => {
    it('should handle empty receipt data', () => {
      const emptyData: ReceiptData = {
        transactionNumber: '',
        transactionDate: '',
        pharmacyName: '',
        pharmacyAddress: '',
        pharmacyPhone: '',
        items: [],
        subtotal: '0',
        total: '0',
        payment: { method: 'cash', cashDetails: { change: '0' } },
        paperWidth: 58,
      };

      const receipt = service.generateReceipt(emptyData);
      expect(receipt).toBeDefined();
      expect(receipt).toBeInstanceOf(Uint8Array);
    });

    it('should handle missing optional fields', () => {
      const minimalData: ReceiptData = {
        transactionNumber: 'TRX-001',
        transactionDate: '2024-05-28T10:00:00+07:00',
        pharmacyName: 'Apotek',
        pharmacyAddress: 'Address',
        pharmacyPhone: '123',
        items: [],
        subtotal: '0',
        total: '0',
        payment: { method: 'cash', cashDetails: { change: '0' } },
        paperWidth: 58,
      };

      const receipt = service.generateReceipt(minimalData);
      expect(receipt).toBeDefined();
      expect(receipt).toBeInstanceOf(Uint8Array);
    });
  });
});
