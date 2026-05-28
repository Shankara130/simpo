/**
 * ScannerSettingsScreen Component Tests
 * Tests scanner configuration UI and functionality
 * Story 7.2: USB Barcode Scanner Integration
 */

import React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { Alert } from 'react-native';
import { NavigationContainer } from '@react-navigation/native';
import { ScannerSettingsScreen } from './ScannerSettingsScreen';
import { ScannerConfigService } from '../services/ScannerConfigService';
import { DEFAULT_SCANNER_CONFIG } from '../types/scanner.types';

// Mock ScannerConfigService
jest.mock('../services/ScannerConfigService', () => ({
  ScannerConfigService: {
    load: jest.fn(),
    save: jest.fn(),
    reset: jest.fn(),
    // Story 7.3: Bluetooth methods
    loadPairedDevices: jest.fn(),
    savePairedDevices: jest.fn(),
    loadLastConnectedDevice: jest.fn(),
    saveLastConnectedDevice: jest.fn(),
    clearLastConnectedDevice: jest.fn(),
    loadBluetoothConfig: jest.fn(),
    saveBluetoothConfig: jest.fn(),
  },
}));

const mockAlert = jest.spyOn(Alert, 'alert').mockImplementation(() => {});

const renderWithNavigation = (component: React.ReactElement) => {
  return render(<NavigationContainer>{component}</NavigationContainer>);
};

describe('ScannerSettingsScreen', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  afterEach(() => {
    mockAlert.mockClear();
  });

  describe('initialization', () => {
    it('should render loading state initially', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);
      (ScannerConfigService.loadBluetoothConfig as jest.Mock).mockResolvedValue({ autoReconnect: true, maxReconnectAttempts: 5, reconnectDelays: [1000, 2000, 4000, 8000], connectionTimeout: 10000 });
      (ScannerConfigService.loadPairedDevices as jest.Mock).mockResolvedValue([]);

      const { getByText } = renderWithNavigation(<ScannerSettingsScreen />);

      expect(getByText('Memuat pengaturan...')).toBeTruthy();
    });

    it('should load settings on mount', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);
      (ScannerConfigService.loadBluetoothConfig as jest.Mock).mockResolvedValue({ autoReconnect: true, maxReconnectAttempts: 5, reconnectDelays: [1000, 2000, 4000, 8000], connectionTimeout: 10000 });
      (ScannerConfigService.loadPairedDevices as jest.Mock).mockResolvedValue([]);

      const { getByText, findByText, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      // Wait for loading to complete
      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      expect(ScannerConfigService.load).toHaveBeenCalled();
      expect(getByText('Pengaturan Scanner')).toBeTruthy();
    });

    it('should render all setting sections', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);

      const { getByText, findByText, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      expect(getByText('Debounce Scan')).toBeTruthy();
      expect(getByText('Panjang Barcode')).toBeTruthy();
      expect(getByText('Feedback')).toBeTruthy();
      expect(getByText('Waktu Scan Maksimum')).toBeTruthy();
    });
  });

  describe('debounce settings', () => {
    it('should display current debounce value', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const debounceInput = await findByTestId('debounce-input');
      expect(debounceInput.props.value).toBe('500');
    });

    it('should allow changing debounce value within range', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const debounceInput = await findByTestId('debounce-input');

      fireEvent.changeText(debounceInput, '1000');
      expect(debounceInput.props.value).toBe('1000');
    });

    it('should reject debounce values outside range', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const debounceInput = await findByTestId('debounce-input');

      // Value outside range should not update
      fireEvent.changeText(debounceInput, '5000');
      expect(debounceInput.props.value).toBe('500'); // Should remain unchanged
    });
  });

  describe('barcode length settings', () => {
    it('should display min and max barcode length values', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const minLengthInput = await findByTestId('min-length-input');
      const maxLengthInput = await findByTestId('max-length-input');

      expect(minLengthInput.props.value).toBe('8');
      expect(maxLengthInput.props.value).toBe('13');
    });

    it('should allow changing min barcode length', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const minLengthInput = await findByTestId('min-length-input');

      fireEvent.changeText(minLengthInput, '10');
      expect(minLengthInput.props.value).toBe('10');
    });

    it('should allow changing max barcode length', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const maxLengthInput = await findByTestId('max-length-input');

      fireEvent.changeText(maxLengthInput, '20');
      expect(maxLengthInput.props.value).toBe('20');
    });
  });

  describe('feedback settings', () => {
    it('should display feedback enabled toggle', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue({
        ...DEFAULT_SCANNER_CONFIG,
        feedbackEnabled: true,
      });

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const feedbackSwitch = await findByTestId('feedback-switch');
      expect(feedbackSwitch.props.value).toBe(true);
    });

    it('should allow toggling feedback', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const feedbackSwitch = await findByTestId('feedback-switch');

      fireEvent(feedbackSwitch, 'onValueChange', false);
      expect(feedbackSwitch.props.value).toBe(false);
    });
  });

  describe('scan time settings', () => {
    it('should display max scan time value', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const maxScanTimeInput = await findByTestId('max-scan-time-input');
      expect(maxScanTimeInput.props.value).toBe('100');
    });

    it('should allow changing max scan time', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const maxScanTimeInput = await findByTestId('max-scan-time-input');

      fireEvent.changeText(maxScanTimeInput, '150');
      expect(maxScanTimeInput.props.value).toBe('150');
    });
  });

  describe('save functionality', () => {
    it('should save configuration when save button pressed', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);
      (ScannerConfigService.save as jest.Mock).mockResolvedValue();

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const saveButton = await findByTestId('save-button');
      fireEvent.press(saveButton);

      await waitFor(() => {
        expect(ScannerConfigService.save).toHaveBeenCalledWith(DEFAULT_SCANNER_CONFIG);
      });
    });

    it('should validate debounce range before saving', async () => {
      // Load config with invalid debounce
      const invalidConfig = {
        ...DEFAULT_SCANNER_CONFIG,
        debounceMs: 50, // Below minimum
      };
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(invalidConfig);
      (ScannerConfigService.save as jest.Mock).mockResolvedValue();

      const { findByTestId, findByText, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const saveButton = await findByTestId('save-button');
      fireEvent.press(saveButton);

      expect(mockAlert).toHaveBeenCalledWith(
        'Invalid Range',
        'Debounce interval harus antara 100ms - 2000ms'
      );
    });

    it('should validate barcode length range before saving', async () => {
      // Load config with invalid barcode length
      const invalidConfig = {
        ...DEFAULT_SCANNER_CONFIG,
        minBarcodeLength: 25, // Above maximum
      };
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(invalidConfig);
      (ScannerConfigService.save as jest.Mock).mockResolvedValue();

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const saveButton = await findByTestId('save-button');
      fireEvent.press(saveButton);

      expect(mockAlert).toHaveBeenCalledWith(
        'Invalid Range',
        'Min barcode length harus antara 1 - 20 karakter'
      );
    });

    it('should validate min < max barcode length', async () => {
      const invalidConfig = {
        ...DEFAULT_SCANNER_CONFIG,
        minBarcodeLength: 15,
        maxBarcodeLength: 10, // Less than min
      };
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(invalidConfig);
      (ScannerConfigService.save as jest.Mock).mockResolvedValue();

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const saveButton = await findByTestId('save-button');
      fireEvent.press(saveButton);

      expect(mockAlert).toHaveBeenCalledWith(
        'Invalid Range',
        'Min barcode length harus kurang dari max barcode length'
      );
    });
  });

  describe('reset functionality', () => {
    it('should reset to defaults when reset button pressed', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);
      (ScannerConfigService.reset as jest.Mock).mockResolvedValue();

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const resetButton = await findByTestId('reset-button');
      fireEvent.press(resetButton);

      // Reset confirmation dialog
      expect(mockAlert).toHaveBeenCalled();
    });
  });

  describe('test scan button', () => {
    it('should render test scan button', async () => {
      (ScannerConfigService.load as jest.Mock).mockResolvedValue(DEFAULT_SCANNER_CONFIG);

      const { findByTestId, queryByText } = renderWithNavigation(
        <ScannerSettingsScreen />
      );

      await waitFor(() => {
        expect(queryByText('Memuat pengaturan...')).toBeNull();
      });

      const testButton = await findByTestId('test-scan-button');
      expect(testButton).toBeTruthy();
      expect(queryByText('Test Scanner')).toBeTruthy();
    });
  });
});
