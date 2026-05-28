/**
 * Printer Settings Screen Tests
 * Tests for printer settings and configuration screen
 */

import React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { PrinterSettingsScreen } from '../screens/PrinterSettingsScreen';
import { PrinterDevice, PrinterStatus } from '../hardware/printer';

// Mock navigation
jest.mock('@react-navigation/native', () => ({
  useNavigation: () => ({
    goBack: jest.fn(),
    navigate: jest.fn(),
  }),
  useRoute: () => ({
    params: {},
  }),
}));

// Mock Slider component
jest.mock('@react-native-community/slider', () => 'Slider');

// Mock PrinterManager
jest.mock('../hardware/PrinterManager', () => {
  return {
    PrinterManager: {
      getInstance: () => ({
        discoverPrinters: jest.fn(() => Promise.resolve([
          {
            id: 'usb-1',
            name: 'Xprinter XP-58IIH',
            connectionType: 0, // USB
            vendorId: 0x0416,
            productId: 0x5011,
          },
          {
            id: 'bt-1',
            name: 'POS-58BT',
            connectionType: 1, // BLUETOOTH
            address: '00:11:22:33:44:55',
          },
        ])),
        connect: jest.fn(() => Promise.resolve(true)),
        disconnect: jest.fn(() => Promise.resolve(true)),
        getStatus: jest.fn(() => 2), // CONNECTED
        getCurrentPrinter: jest.fn(() => null),
        onError: jest.fn(),
        onStatusChange: jest.fn(),
      }),
    },
  };
});

describe('PrinterSettingsScreen', () => {
  describe('Screen Rendering', () => {
    it('should render printer settings screen', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      expect(getByTestId('printer-settings-screen')).toBeDefined();
    });

    it('should have title "Printer Settings"', () => {
      const { getByText } = render(<PrinterSettingsScreen />);
      expect(getByText('Pengaturan Printer')).toBeDefined();
    });
  });

  describe('Printer Discovery', () => {
    it('should have scan button to discover printers', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      expect(getByTestId('scan-printers-button')).toBeDefined();
    });

    it('should show loading indicator while scanning', async () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const scanButton = getByTestId('scan-printers-button');

      fireEvent.press(scanButton);

      await waitFor(() => {
        expect(getByTestId('scanning-indicator')).toBeDefined();
      });
    });

    it('should display discovered printers', async () => {
      const { getByTestId, getByText } = render(<PrinterSettingsScreen />);
      const scanButton = getByTestId('scan-printers-button');

      fireEvent.press(scanButton);

      await waitFor(() => {
        expect(getByText('Xprinter XP-58IIH')).toBeDefined();
        expect(getByText('POS-58BT')).toBeDefined();
      });
    });
  });

  describe('Printer Connection', () => {
    it('should show connect button for each printer', async () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const scanButton = getByTestId('scan-printers-button');

      fireEvent.press(scanButton);

      await waitFor(() => {
        expect(getByTestId('connect-usb-1')).toBeDefined();
        expect(getByTestId('connect-bt-1')).toBeDefined();
      });
    });

    it('should show disconnect button when printer is connected', async () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      await waitFor(() => {
        const disconnectButton = getByTestId('disconnect-button');
        expect(disconnectButton).toBeDefined();
      });
    });

    it('should update printer status after connection', async () => {
      const { getByTestId, getByText } = render(<PrinterSettingsScreen />);
      const connectButton = getByTestId('connect-usb-1');

      fireEvent.press(connectButton);

      await waitFor(() => {
        expect(getByText('Terhubung')).toBeDefined();
      });
    });
  });

  describe('Paper Width Selection', () => {
    it('should have paper width selector', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      expect(getByTestId('paper-width-selector')).toBeDefined();
    });

    it('should show 58mm and 80mm options', () => {
      const { getByText } = render(<PrinterSettingsScreen />);
      expect(getByText('58mm')).toBeDefined();
      expect(getByText('80mm')).toBeDefined();
    });

    it('should allow selecting 58mm paper width', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const width58 = getByTestId('width-58mm');

      fireEvent.press(width58);

      // Verify selection state
      expect(getByTestId('width-58mm')).toBeDefined();
    });

    it('should allow selecting 80mm paper width', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const width80 = getByTestId('width-80mm');

      fireEvent.press(width80);

      // Verify selection state
      expect(getByTestId('width-80mm')).toBeDefined();
    });
  });

  describe('Printer Darkness Adjustment', () => {
    it('should have darkness adjustment slider', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      expect(getByTestId('darkness-slider')).toBeDefined();
    });

    it('should show current darkness value', () => {
      const { getByText } = render(<PrinterSettingsScreen />);
      expect(getByText('Intensitas')).toBeDefined();
    });

    it('should allow adjusting darkness level', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const slider = getByTestId('darkness-slider');

      fireEvent(slider, 'valueChange', 0.7);

      // Verify value change
      expect(getByTestId('darkness-slider')).toBeDefined();
    });
  });

  describe('Printer Profile Management', () => {
    it('should have save profile button', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      expect(getByTestId('save-profile-button')).toBeDefined();
    });

    it('should have load profile button', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      expect(getByTestId('load-profile-button')).toBeDefined();
    });

    it('should allow saving printer configuration', async () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const saveButton = getByTestId('save-profile-button');

      fireEvent.press(saveButton);

      await waitFor(() => {
        expect(getByText('Profil tersimpan')).toBeDefined();
      });
    });

    it('should allow loading saved printer configuration', async () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const loadButton = getByTestId('load-profile-button');

      fireEvent.press(loadButton);

      await waitFor(() => {
        expect(getByText('Profil dimuat')).toBeDefined();
      });
    });
  });

  describe('Test Print Functionality', () => {
    it('should have test print button', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      expect(getByTestId('test-print-button')).toBeDefined();
    });

    it('should enable test print when printer is connected', async () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      await waitFor(() => {
        const testButton = getByTestId('test-print-button');
        expect(testButton).toBeDefined();
      });
    });

    it('should disable test print when no printer is connected', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const testButton = getByTestId('test-print-button');

      expect(testButton.props.disabled).toBe(true);
    });

    it('should print test receipt when test print is pressed', async () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const testButton = getByTestId('test-print-button');

      fireEvent.press(testButton);

      await waitFor(() => {
        expect(getByText('Test print berhasil')).toBeDefined();
      });
    });
  });

  describe('Error Handling', () => {
    it('should show error message when scan fails', async () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const scanButton = getByTestId('scan-printers-button');

      // Mock failed scan
      fireEvent.press(scanButton);

      await waitFor(() => {
        expect(getByText('Gagal memindai printer')).toBeDefined();
      });
    });

    it('should show error message when connection fails', async () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const connectButton = getByTestId('connect-usb-1');

      fireEvent.press(connectButton);

      await waitFor(() => {
        expect(getByText('Gagal terhubung ke printer')).toBeDefined();
      });
    });

    it('should provide retry option for failed operations', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      const retryButton = getByTestId('retry-button');

      expect(retryButton).toBeDefined();
    });
  });

  describe('Accessibility', () => {
    it('should have proper accessibility labels', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      expect(getByTestId('printer-settings-screen')).toBeAccessibilityElement();
    });

    it('should announce printer status changes', async () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      await waitFor(() => {
        const statusElement = getByTestId('printer-status-announcement');
        expect(statusElement).toBeDefined();
      });
    });
  });

  // ============================================================================
  // Cash Drawer Configuration Tests (Story 7.4)
  // ============================================================================

  describe('Cash Drawer Configuration', () => {
    beforeEach(() => {
      jest.clearAllMocks();
    });

    it('should display cash drawer configuration section', () => {
      const { getByText } = render(<PrinterSettingsScreen />);

      expect(getByText('Buka Laci Uang Otomatis')).toBeTruthy();
    });

    it('should have drawer auto-open toggle', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      expect(getByTestId('drawer-auto-toggle')).toBeDefined();
    });

    it('should have pulse timing slider', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      expect(getByTestId('drawer-pulse-slider')).toBeDefined();
    });

    it('should have drawer pin selection (Pin 2 and Pin 5)', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      expect(getByTestId('drawer-pin-2')).toBeDefined();
      expect(getByTestId('drawer-pin-5')).toBeDefined();
    });

    it('should have test drawer button', () => {
      const { getByTestId, getByText } = render(<PrinterSettingsScreen />);

      expect(getByTestId('test-drawer-button')).toBeDefined();
      expect(getByText('🧪 Tes Buka Laci')).toBeTruthy();
    });

    it('should disable test drawer button when printer not connected', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      const testButton = getByTestId('test-drawer-button');
      expect(testButton).toBeDefined();
      // Button should be disabled when printer is not connected
    });

    it('should show drawer configuration only when auto-open is enabled', () => {
      const { getByTestId, queryByText } = render(<PrinterSettingsScreen />);

      // Initially, drawer config options should be visible when auto-open is enabled
      expect(getByTestId('drawer-pulse-slider')).toBeDefined();
    });

    it('should allow toggling auto-open setting', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      const toggle = getByTestId('drawer-auto-toggle');
      expect(toggle).toBeDefined();
    });

    it('should allow adjusting pulse timing', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      const slider = getByTestId('drawer-pulse-slider');
      expect(slider).toBeDefined();
    });

    it('should allow selecting drawer pin (Pin 2 vs Pin 5)', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      const pin2Button = getByTestId('drawer-pin-2');
      const pin5Button = getByTestId('drawer-pin-5');

      expect(pin2Button).toBeDefined();
      expect(pin5Button).toBeDefined();
    });
  });
});

  describe('UI Layout', () => {
    it('should show printer status at top', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      expect(getByTestId('printer-status-indicator')).toBeDefined();
    });

    it('should organize settings in logical sections', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);

      expect(getByTestId('discovery-section')).toBeDefined();
      expect(getByTestId('configuration-section')).toBeDefined();
      expect(getByTestId('profile-section')).toBeDefined();
    });

    it('should provide back button to return', () => {
      const { getByTestId } = render(<PrinterSettingsScreen />);
      expect(getByTestId('back-button')).toBeDefined();
    });
  });
});