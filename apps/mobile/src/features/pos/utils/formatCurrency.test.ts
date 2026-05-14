/**
 * Tests for formatCurrency utility
 * Tests Indonesian Rupiah formatting with dot separator
 */

import { formatCurrency } from './formatCurrency';

describe('formatCurrency', () => {
  describe('small amounts', () => {
    it('should format 1000 as Rp 1.000', () => {
      expect(formatCurrency(1000)).toBe('Rp 1.000');
    });

    it('should format 5000 as Rp 5.000', () => {
      expect(formatCurrency(5000)).toBe('Rp 5.000');
    });

    it('should format 10000 as Rp 10.000', () => {
      expect(formatCurrency(10000)).toBe('Rp 10.000');
    });
  });

  describe('medium amounts', () => {
    it('should format 150000 as Rp 150.000', () => {
      expect(formatCurrency(150000)).toBe('Rp 150.000');
    });

    it('should format 500000 as Rp 500.000', () => {
      expect(formatCurrency(500000)).toBe('Rp 500.000');
    });
  });

  describe('large amounts', () => {
    it('should format 1000000 as Rp 1.000.000', () => {
      expect(formatCurrency(1000000)).toBe('Rp 1.000.000');
    });

    it('should format 1234567 as Rp 1.234.567', () => {
      expect(formatCurrency(1234567)).toBe('Rp 1.234.567');
    });

    it('should format 10000000 as Rp 10.000.000', () => {
      expect(formatCurrency(10000000)).toBe('Rp 10.000.000');
    });

    it('should format 1000000000 as Rp 1.000.000.000', () => {
      expect(formatCurrency(1000000000)).toBe('Rp 1.000.000.000');
    });
  });

  describe('edge cases', () => {
    it('should format 0 as Rp 0', () => {
      expect(formatCurrency(0)).toBe('Rp 0');
    });

    it('should format 1 as Rp 1', () => {
      expect(formatCurrency(1)).toBe('Rp 1');
    });

    it('should handle decimals by rounding', () => {
      expect(formatCurrency(1500.7)).toBe('Rp 1.501');
    });

    it('should handle very large numbers', () => {
      expect(formatCurrency(999999999999)).toBe('Rp 999.999.999.999');
    });
  });

  describe('string input handling', () => {
    it('should handle string number input', () => {
      expect(formatCurrency('150000' as any)).toBe('Rp 150.000');
    });

    it('should handle decimal string input', () => {
      expect(formatCurrency('150000.50' as any)).toBe('Rp 150.001');
    });
  });
});
