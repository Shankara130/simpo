/**
 * Printer Status Component Tests
 * Test printer status display and visual feedback
 */

import React from 'react';
import { render } from '@testing-library/react-native';
import PrinterStatusComponent, { PrinterStatusProps } from './PrinterStatus';
import { PrinterStatus } from '../hardware/printer';

describe('PrinterStatus Component', () => {
  const defaultProps: PrinterStatusProps = {
    status: PrinterStatus.DISCONNECTED,
  };

  describe('Rendering', () => {
    it('should render DISCONNECTED status correctly', () => {
      const { getByTestId, getByText } = render(
        <PrinterStatusComponent {...defaultProps} status={PrinterStatus.DISCONNECTED} />
      );

      expect(getByTestId('printer-status')).toBeDefined();
      expect(getByText('Printer Terputus')).toBeDefined();
    });

    it('should render CONNECTED status correctly', () => {
      const { getByTestId, getByText } = render(
        <PrinterStatusComponent {...defaultProps} status={PrinterStatus.CONNECTED} />
      );

      expect(getByTestId('printer-status')).toBeDefined();
      expect(getByText('Printer Terhubung')).toBeDefined();
    });

    it('should render CONNECTING status correctly', () => {
      const { getByTestId, getByText } = render(
        <PrinterStatusComponent {...defaultProps} status={PrinterStatus.CONNECTING} />
      );

      expect(getByTestId('printer-status')).toBeDefined();
      expect(getByText('Menghubungkan...')).toBeDefined();
    });

    it('should render ERROR status correctly', () => {
      const { getByTestId, getByText } = render(
        <PrinterStatusComponent {...defaultProps} status={PrinterStatus.ERROR} />
      );

      expect(getByTestId('printer-status')).toBeDefined();
      expect(getByText('Printer Error')).toBeDefined();
    });

    it('should render OUT_OF_PAPER status correctly', () => {
      const { getByTestId, getByText } = render(
        <PrinterStatusComponent {...defaultProps} status={PrinterStatus.OUT_OF_PAPER} />
      );

      expect(getByTestId('printer-status')).toBeDefined();
      expect(getByText('Kertas Habis')).toBeDefined();
    });
  });

  describe('Printer Name Display', () => {
    it('should display printer name when connected', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.CONNECTED}
          printerName="PT-210 Thermal Printer"
        />
      );

      expect(getByText('PT-210 Thermal Printer')).toBeDefined();
    });

    it('should display printer name when connecting', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.CONNECTING}
          printerName="POS-58BT"
        />
      );

      expect(getByText('POS-58BT')).toBeDefined();
    });

    it('should not display printer name when disconnected', () => {
      const { queryByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.DISCONNECTED}
          printerName="PT-210"
        />
      );

      expect(queryByText('PT-210')).toBeNull();
    });
  });

  describe('Error Message Display', () => {
    it('should display error message when status is ERROR', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.ERROR}
          errorMessage="Connection failed"
        />
      );

      expect(getByText('Connection failed')).toBeDefined();
    });

    it('should not display error message when status is not ERROR', () => {
      const { queryByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.CONNECTED}
          errorMessage="This should not show"
        />
      );

      expect(queryByText('This should not show')).toBeNull();
    });
  });

  describe('Compact Mode', () => {
    it('should render in compact mode when compact prop is true', () => {
      const { getByTestId, getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.CONNECTED}
          compact={true}
        />
      );

      expect(getByTestId('printer-status')).toBeDefined();
      expect(getByText('Printer Terhubung')).toBeDefined();
    });

    it('should display compact label correctly', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.DISCONNECTED}
          compact={true}
        />
      );

      expect(getByText('Printer Terputus')).toBeDefined();
    });
  });

  describe('Accessibility', () => {
    it('should have accessibility label for status', () => {
      const { getByLabelText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.CONNECTED}
        />
      );

      expect(getByLabelText('Printer status: Printer Terhubung')).toBeDefined();
    });

    it('should have accessibility role of text', () => {
      const { getAllByRole } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.CONNECTED}
        />
      );

      expect(getAllByRole('text').length).toBeGreaterThan(0);
    });
  });

  describe('Custom Test ID', () => {
    it('should use custom test ID when provided', () => {
      const { getByTestId } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.CONNECTED}
          testID="custom-printer-status"
        />
      );

      expect(getByTestId('custom-printer-status')).toBeDefined();
    });

    it('should use default test ID when not provided', () => {
      const { getByTestId } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.CONNECTED}
        />
      );

      expect(getByTestId('printer-status')).toBeDefined();
    });
  });

  describe('Indonesian Language', () => {
    it('should display Indonesian label for CONNECTED status', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.CONNECTED}
        />
      );

      expect(getByText('Printer Terhubung')).toBeDefined();
    });

    it('should display Indonesian label for DISCONNECTED status', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.DISCONNECTED}
        />
      );

      expect(getByText('Printer Terputus')).toBeDefined();
    });

    it('should display Indonesian label for CONNECTING status', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.CONNECTING}
        />
      );

      expect(getByText('Menghubungkan...')).toBeDefined();
    });

    it('should display Indonesian label for OUT_OF_PAPER status', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.OUT_OF_PAPER}
        />
      );

      expect(getByText('Kertas Habis')).toBeDefined();
    });
  });

  describe('Status Indicators', () => {
    it('should show checkmark icon for CONNECTED status', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.CONNECTED}
        />
      );

      expect(getByText('✓')).toBeDefined();
    });

    it('should show circle icon for DISCONNECTED status', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.DISCONNECTED}
        />
      );

      expect(getByText('○')).toBeDefined();
    });

    it('should show exclamation icon for ERROR status', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.ERROR}
        />
      );

      expect(getByText('!')).toBeDefined();
    });

    it('should show warning icon for OUT_OF_PAPER status', () => {
      const { getByText } = render(
        <PrinterStatusComponent
          {...defaultProps}
          status={PrinterStatus.OUT_OF_PAPER}
        />
      );

      expect(getByText('⚠')).toBeDefined();
    });
  });
});
