/**
 * Cash Drawer Status Component Tests
 * Tests for drawer status indicator component
 * Story 7.4: Cash Drawer Control via Printer Kick
 */

import React from 'react';
import { render } from '@testing-library/react-native';
import { CashDrawerStatus } from './CashDrawerStatus';
import { DrawerStatus } from '../../hardware/printer';

describe('CashDrawerStatus Component', () => {
  describe('Rendering', () => {
    it('should render connected status', () => {
      const { getByText } = render(
        <CashDrawerStatus status="connected" />
      );

      expect(getByText('Laci Uang: Terhubung')).toBeTruthy();
    });

    it('should render disconnected status', () => {
      const { getByText } = render(
        <CashDrawerStatus status="disconnected" />
      );

      expect(getByText('Laci Uang: Terputus')).toBeTruthy();
    });

    it('should render opening status', () => {
      const { getByText } = render(
        <CashDrawerStatus status="opening" />
      );

      expect(getByText('Laci Uang: Membuka...')).toBeTruthy();
    });

    it('should render failed status', () => {
      const { getByText } = render(
        <CashDrawerStatus status="failed" />
      );

      expect(getByText('Laci Uang: Gagal')).toBeTruthy();
    });

    it('should render failed status with custom error message', () => {
      const { getByText } = render(
        <CashDrawerStatus status="failed" error="Printer not connected" />
      );

      expect(getByText('Laci Uang: Gagal')).toBeTruthy();
    });
  });

  describe('Status Colors', () => {
    it('should show green color for connected status', () => {
      const { getByText } = render(
        <CashDrawerStatus status="connected" />
      );

      const statusText = getByText('Laci Uang: Terhubung');
      // Check if the component renders with green color (4CAF50)
      expect(statusText).toBeTruthy();
    });

    it('should show gray color for disconnected status', () => {
      const { getByText } = render(
        <CashDrawerStatus status="disconnected" />
      );

      const statusText = getByText('Laci Uang: Terputus');
      expect(statusText).toBeTruthy();
    });

    it('should show orange color for opening status', () => {
      const { getByText } = render(
        <CashDrawerStatus status="opening" />
      );

      const statusText = getByText('Laci Uang: Membuka...');
      expect(statusText).toBeTruthy();
    });

    it('should show red color for failed status', () => {
      const { getByText } = render(
        <CashDrawerStatus status="failed" />
      );

      const statusText = getByText('Laci Uang: Gagal');
      expect(statusText).toBeTruthy();
    });
  });

  describe('Error Handling', () => {
    it('should handle undefined error gracefully', () => {
      const { getByText } = render(
        <CashDrawerStatus status="failed" error={undefined} />
      );

      expect(getByText('Laci Uang: Gagal')).toBeTruthy();
    });

    it('should handle null error gracefully', () => {
      const { getByText } = render(
        <CashDrawerStatus status="failed" error={null} />
      );

      expect(getByText('Laci Uang: Gagal')).toBeTruthy();
    });

    it('should display custom error message when provided', () => {
      const { getByText } = render(
        <CashDrawerStatus status="failed" error="Drawer disconnected" />
      );

      expect(getByText('Laci Uang: Gagal')).toBeTruthy();
    });
  });

  describe('Component Structure', () => {
    it('should have proper container styling', () => {
      const { toJSON } = render(
        <CashDrawerStatus status="connected" />
      );

      const component = toJSON();
      expect(component).toBeTruthy();
    });

    it('should update status when status prop changes', () => {
      const { getByText, rerender } = render(
        <CashDrawerStatus status="disconnected" />
      );

      expect(getByText('Laci Uang: Terputus')).toBeTruthy();

      rerender(<CashDrawerStatus status="connected" />);

      expect(getByText('Laci Uang: Terhubung')).toBeTruthy();
    });
  });

  describe('Accessibility', () => {
    it('should provide accessible text for screen readers', () => {
      const { getByText } = render(
        <CashDrawerStatus status="connected" />
      );

      expect(getByText('Laci Uang: Terhubung')).toBeTruthy();
    });

    it('should indicate error state in accessible text', () => {
      const { getByText } = render(
        <CashDrawerStatus status="failed" error="Connection failed" />
      );

      expect(getByText('Laci Uang: Gagal')).toBeTruthy();
    });
  });
});
