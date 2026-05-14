/**
 * Payment Types Tests
 * Test payment method type definitions and discriminated union patterns
 */

import {
  PaymentMethod,
  EWalletType,
  CashPaymentData,
  BankTransferPaymentData,
  EWalletPaymentData,
  PaymentData,
  PaymentSelection,
} from './payment.types';

describe('Payment Types', () => {
  describe('PaymentMethod Enum', () => {
    it('should have CASH value equal to "CASH"', () => {
      expect(PaymentMethod.CASH).toBe('CASH');
    });

    it('should have TRANSFER value equal to "TRANSFER"', () => {
      expect(PaymentMethod.TRANSFER).toBe('TRANSFER');
    });

    it('should have E_WALLET value equal to "E_WALLET"', () => {
      expect(PaymentMethod.E_WALLET).toBe('E_WALLET');
    });
  });

  describe('EWalletType Enum', () => {
    it('should have GOPAY value equal to "GOPAY"', () => {
      expect(EWalletType.GOPAY).toBe('GOPAY');
    });

    it('should have OVO value equal to "OVO"', () => {
      expect(EWalletType.OVO).toBe('OVO');
    });

    it('should have DANA value equal to "DANA"', () => {
      expect(EWalletType.DANA).toBe('DANA');
    });

    it('should have SHOPEE_PAY value equal to "SHOPEE_PAY"', () => {
      expect(EWalletType.SHOPEE_PAY).toBe('SHOPEE_PAY');
    });
  });

  describe('CashPaymentData Interface', () => {
    it('should create valid cash payment data', () => {
      const cashData: CashPaymentData = {
        method: PaymentMethod.CASH,
      };

      expect(cashData.method).toBe(PaymentMethod.CASH);
    });
  });

  describe('BankTransferPaymentData Interface', () => {
    it('should create valid bank transfer payment data', () => {
      const transferData: BankTransferPaymentData = {
        method: PaymentMethod.TRANSFER,
        accountName: 'John Doe',
        referenceNumber: 'REF123',
      };

      expect(transferData.method).toBe(PaymentMethod.TRANSFER);
      expect(transferData.accountName).toBe('John Doe');
      expect(transferData.referenceNumber).toBe('REF123');
    });
  });

  describe('EWalletPaymentData Interface', () => {
    it('should create valid e-wallet payment data', () => {
      const ewalletData: EWalletPaymentData = {
        method: PaymentMethod.E_WALLET,
        walletType: EWalletType.GOPAY,
        confirmationInput: '081234567890',
      };

      expect(ewalletData.method).toBe(PaymentMethod.E_WALLET);
      expect(ewalletData.walletType).toBe(EWalletType.GOPAY);
      expect(ewalletData.confirmationInput).toBe('081234567890');
    });
  });

  describe('PaymentData Union Type', () => {
    it('should accept cash payment data', () => {
      const paymentData: PaymentData = {
        method: PaymentMethod.CASH,
      };

      expect(paymentData.method).toBe(PaymentMethod.CASH);
    });

    it('should accept bank transfer payment data', () => {
      const paymentData: PaymentData = {
        method: PaymentMethod.TRANSFER,
        accountName: 'Jane Smith',
        referenceNumber: 'TRX456',
      };

      expect(paymentData.method).toBe(PaymentMethod.TRANSFER);
    });

    it('should accept e-wallet payment data', () => {
      const paymentData: PaymentData = {
        method: PaymentMethod.E_WALLET,
        walletType: EWalletType.OVO,
        confirmationInput: 'CONF789',
      };

      expect(paymentData.method).toBe(PaymentMethod.E_WALLET);
    });
  });

  describe('PaymentSelection Interface', () => {
    it('should create valid payment selection with cash', () => {
      const selection: PaymentSelection = {
        paymentData: {
          method: PaymentMethod.CASH,
        },
        isValid: true,
      };

      expect(selection.paymentData.method).toBe(PaymentMethod.CASH);
      expect(selection.isValid).toBe(true);
    });

    it('should create valid payment selection with transfer', () => {
      const selection: PaymentSelection = {
        paymentData: {
          method: PaymentMethod.TRANSFER,
          accountName: 'Test Account',
          referenceNumber: 'REF001',
        },
        isValid: true,
      };

      expect(selection.paymentData.method).toBe(PaymentMethod.TRANSFER);
      expect(selection.isValid).toBe(true);
    });
  });

  describe('Type Discrimination with PaymentData', () => {
    it('should narrow type correctly for CASH payment', () => {
      const paymentData: PaymentData = {
        method: PaymentMethod.CASH,
      };

      if (paymentData.method === PaymentMethod.CASH) {
        // TypeScript should know this is CashPaymentData
        expect(paymentData.method).toBe(PaymentMethod.CASH);
        expect(Object.keys(paymentData)).toEqual(['method']);
      }
    });

    it('should narrow type correctly for TRANSFER payment', () => {
      const paymentData: PaymentData = {
        method: PaymentMethod.TRANSFER,
        accountName: 'Bank Account',
        referenceNumber: 'REF123',
      };

      if (paymentData.method === PaymentMethod.TRANSFER) {
        // TypeScript should know this is BankTransferPaymentData
        expect(paymentData.method).toBe(PaymentMethod.TRANSFER);
        expect(paymentData.accountName).toBe('Bank Account');
        expect(paymentData.referenceNumber).toBe('REF123');
      }
    });

    it('should narrow type correctly for E_WALLET payment', () => {
      const paymentData: PaymentData = {
        method: PaymentMethod.E_WALLET,
        walletType: EWalletType.DANA,
        confirmationInput: '123456',
      };

      if (paymentData.method === PaymentMethod.E_WALLET) {
        // TypeScript should know this is EWalletPaymentData
        expect(paymentData.method).toBe(PaymentMethod.E_WALLET);
        expect(paymentData.walletType).toBe(EWalletType.DANA);
        expect(paymentData.confirmationInput).toBe('123456');
      }
    });
  });

  describe('Type Guards for Payment Validation', () => {
    it('should identify cash payment correctly', () => {
      const paymentData: PaymentData = {
        method: PaymentMethod.CASH,
      };

      const isCash = paymentData.method === PaymentMethod.CASH;
      expect(isCash).toBe(true);
    });

    it('should identify transfer payment correctly', () => {
      const paymentData: PaymentData = {
        method: PaymentMethod.TRANSFER,
        accountName: 'Account',
        referenceNumber: 'REF',
      };

      const isTransfer = paymentData.method === PaymentMethod.TRANSFER;
      expect(isTransfer).toBe(true);
    });

    it('should identify e-wallet payment correctly', () => {
      const paymentData: PaymentData = {
        method: PaymentMethod.E_WALLET,
        walletType: EWalletType.SHOPEE_PAY,
        confirmationInput: 'CONF',
      };

      const isEWallet = paymentData.method === PaymentMethod.E_WALLET;
      expect(isEWallet).toBe(true);
    });
  });
});
