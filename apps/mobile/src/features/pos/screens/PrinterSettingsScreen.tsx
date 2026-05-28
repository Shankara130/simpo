/**
 * Printer Settings Screen
 * Allows users to discover, connect, and configure thermal printers
 * Story 7.1 Task 5: Create Printer Settings Screen (AC: 6)
 */

import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  SafeAreaView,
} from 'react-native';
import Slider from '@react-native-community/slider';
import { useNavigation } from '@react-navigation/native';
import { PrinterManager } from '../hardware/PrinterManager';
import {
  PrinterDevice,
  PrinterConnectionType,
  PrinterStatus,
} from '../hardware/printer';
import { PrinterStatusComponent } from '../components/PrinterStatus';

/**
 * Printer Settings Props
 */
interface PrinterSettingsProps {
  onPrinterConnected?: (device: PrinterDevice) => void;
  onPrinterDisconnected?: () => void;
}

/**
 * Printer Settings Screen Component
 */
export const PrinterSettingsScreen: React.FC<PrinterSettingsProps> = ({
  onPrinterConnected,
  onPrinterDisconnected,
}) => {
  const navigation = useNavigation();
  const printerManager = PrinterManager.getInstance();

  // State
  const [isScanning, setIsScanning] = useState(false);
  const [discoveredPrinters, setDiscoveredPrinters] = useState<PrinterDevice[]>([]);
  const [currentPrinter, setCurrentPrinter] = useState<PrinterDevice | null>(null);
  const [printerStatus, setPrinterStatus] = useState<PrinterStatus>(PrinterStatus.DISCONNECTED);
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [selectedPaperWidth, setSelectedPaperWidth] = useState<58 | 80>(58);
  const [printerDarkness, setPrinterDarkness] = useState(0.5);
  const [isTestPrinting, setIsTestPrinting] = useState(false);

  /**
   * Scan for available printers
   */
  const handleScanPrinters = useCallback(async () => {
    setIsScanning(true);
    setErrorMessage('');

    try {
      const printers = await printerManager.discoverPrinters();
      setDiscoveredPrinters(printers);

      if (printers.length === 0) {
        Alert.alert('Info', 'Tidak ada printer yang ditemukan. Pastikan printer menyala dan terhubung.');
      }
    } catch (error) {
      setErrorMessage('Gagal memindai printer');
      Alert.alert('Error', 'Gagal memindai printer. Silakan coba lagi.');
    } finally {
      setIsScanning(false);
    }
  }, [printerManager]);

  /**
   * Connect to printer
   */
  const handleConnectPrinter = useCallback(async (device: PrinterDevice) => {
    setErrorMessage('');

    try {
      const connected = await printerManager.connect(device);

      if (connected) {
        setCurrentPrinter(device);
        onPrinterConnected?.(device);
        Alert.alert('Sukses', `Berhasil terhubung ke ${device.name}`);
      } else {
        setErrorMessage('Gagal terhubung ke printer');
        Alert.alert('Error', 'Gagal terhubung ke printer. Silakan coba lagi.');
      }
    } catch (error) {
      setErrorMessage('Gagal terhubung ke printer');
      Alert.alert('Error', 'Gagal terhubung ke printer. Silakan coba lagi.');
    }
  }, [printerManager, onPrinterConnected]);

  /**
   * Disconnect from printer
   */
  const handleDisconnectPrinter = useCallback(async () => {
    try {
      const disconnected = await printerManager.disconnect();

      if (disconnected) {
        setCurrentPrinter(null);
        onPrinterDisconnected?.();
        Alert.alert('Sukses', 'Berhasil diputus dari printer');
      }
    } catch (error) {
      Alert.alert('Error', 'Gagal memutus dari printer');
    }
  }, [printerManager, onPrinterDisconnected]);

  /**
   * Test print functionality
   */
  const handleTestPrint = useCallback(async () => {
    if (!currentPrinter) {
      Alert.alert('Info', 'Silakan hubungkan ke printer terlebih dahulu');
      return;
    }

    setIsTestPrinting(true);

    try {
      // Generate test receipt using ReceiptTemplateService
      const { ReceiptTemplateService } = await import('../services/ReceiptTemplateService');
      const templateService = new ReceiptTemplateService();

      const testReceiptData = {
        transactionNumber: 'TEST-001',
        transactionDate: new Date().toISOString(),
        pharmacyName: 'APOTEK TEST',
        pharmacyAddress: 'Jl. Test No. 123',
        pharmacyPhone: '021-123456',
        items: [
          {
            name: 'Item Test',
            quantity: 1,
            unitPrice: '15000',
            subtotal: '15000',
          },
        ],
        subtotal: '15000',
        total: '15000',
        payment: {
          method: 'cash',
          cashDetails: { change: '5000' },
        },
        paperWidth: selectedPaperWidth,
      };

      const receiptData = templateService.generateReceipt(testReceiptData);
      const printSuccess = await printerManager.print(receiptData);

      if (printSuccess) {
        Alert.alert('Sukses', 'Test print berhasil');
      } else {
        Alert.alert('Gagal', 'Test print gagal. Silakan coba lagi.');
      }
    } catch (error) {
      Alert.alert('Error', 'Gagal melakukan test print');
    } finally {
      setIsTestPrinting(false);
    }
  }, [currentPrinter, printerManager, selectedPaperWidth, onPrinterConnected, onPrinterDisconnected]);

  /**
   * Save printer profile
   */
  const handleSaveProfile = useCallback(() => {
    // Save printer configuration to AsyncStorage or similar
    const profile = {
      deviceId: currentPrinter?.id,
      paperWidth: selectedPaperWidth,
      darkness: printerDarkness,
      savedAt: new Date().toISOString(),
    };

    // TODO: Implement actual storage saving
    Alert.alert('Sukses', 'Profil printer tersimpan');
  }, [currentPrinter, selectedPaperWidth, printerDarkness]);

  /**
   * Load printer profile
   */
  const handleLoadProfile = useCallback(() => {
    // Load printer configuration from AsyncStorage or similar
    // TODO: Implement actual storage loading
    Alert.alert('Info', 'Profil printer dimuat');
  }, []);

  /**
   * Handle printer status changes
   */
  useEffect(() => {
    const statusHandler = (status: PrinterStatus) => {
      setPrinterStatus(status);
    };

    printerManager.onStatusChange(statusHandler);

    // Get initial status
    setPrinterStatus(printerManager.getStatus());
    setCurrentPrinter(printerManager.getCurrentPrinter());

    return () => {
      // Cleanup: Remove status handler to prevent memory leaks
      printerManager.clearStatusChangeHandler();
    };
  }, [printerManager]);

  /**
   * Get connection type icon
   */
  const getConnectionIcon = (type: PrinterConnectionType): string => {
    switch (type) {
      case PrinterConnectionType.USB:
        return '🔌';
      case PrinterConnectionType.BLUETOOTH:
        return '📶';
      case PrinterConnectionType.NETWORK:
        return '🌐';
      default:
        return '🖨️';
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <TouchableOpacity
          onPress={() => navigation.goBack()}
          style={styles.backButton}
        >
          <Text style={styles.backButtonText}>←</Text>
        </TouchableOpacity>
        <Text style={styles.title}>Pengaturan Printer</Text>
      </View>

      <ScrollView style={styles.content}>
        {/* Printer Status Section */}
        <View style={styles.section} testID="printer-status-indicator">
          <Text style={styles.sectionTitle}>Status Printer</Text>
          <PrinterStatusComponent
            status={printerStatus}
            printerName={currentPrinter?.name}
            errorMessage={errorMessage}
          />
        </View>

        {/* Printer Discovery Section */}
        <View style={styles.section} testID="discovery-section">
          <Text style={styles.sectionTitle}>Temukan Printer</Text>
          <TouchableOpacity
            style={styles.scanButton}
            onPress={handleScanPrinters}
            disabled={isScanning}
            testID="scan-printers-button"
          >
            {isScanning ? (
              <ActivityIndicator size="small" color="#fff" />
            ) : (
              <Text style={styles.scanButtonText}>🔍 Scan Printer</Text>
            )}
          </TouchableOpacity>

          {isScanning && (
            <View style={styles.loadingContainer}>
              <ActivityIndicator size="small" color="#2563EB" />
              <Text style={styles.loadingText}>Memindai printer...</Text>
            </View>
          )}

          {discoveredPrinters.length > 0 && (
            <View style={styles.printerList}>
              {discoveredPrinters.map((printer) => (
                <View
                  key={printer.id}
                  style={styles.printerItem}
                  testID={`printer-item-${printer.id}`}
                >
                  <View style={styles.printerInfo}>
                    <Text style={styles.printerIcon}>
                      {getConnectionIcon(printer.connectionType)}
                    </Text>
                    <View style={styles.printerDetails}>
                      <Text style={styles.printerName}>{printer.name}</Text>
                      <Text style={styles.printerType}>
                        {printer.connectionType.toUpperCase()}
                        {printer.address && ` • ${printer.address}`}
                      </Text>
                    </View>
                  </View>
                  <TouchableOpacity
                    style={styles.connectButton}
                    onPress={() => handleConnectPrinter(printer)}
                    testID={`connect-${printer.id}`}
                  >
                    <Text style={styles.connectButtonText}>
                      {currentPrinter?.id === printer.id ? 'Terhubung' : 'Hubungkan'}
                    </Text>
                  </TouchableOpacity>
                </View>
              ))}
            </View>
          )}
        </View>

        {/* Configuration Section */}
        <View style={styles.section} testID="configuration-section">
          <Text style={styles.sectionTitle}>Konfigurasi</Text>

          {/* Paper Width Selection */}
          <View style={styles.configItem}>
            <Text style={styles.configLabel}>Ukuran Kertas</Text>
            <View style={styles.paperWidthSelector}>
              <TouchableOpacity
                style={[
                  styles.paperWidthOption,
                  selectedPaperWidth === 58 && styles.selectedOption,
                ]}
                onPress={() => setSelectedPaperWidth(58)}
                testID="width-58mm"
              >
                <Text
                  style={[
                    styles.paperWidthText,
                    selectedPaperWidth === 58 && styles.selectedText,
                  ]}
                >
                  58mm
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[
                  styles.paperWidthOption,
                  selectedPaperWidth === 80 && styles.selectedOption,
                ]}
                onPress={() => setSelectedPaperWidth(80)}
                testID="width-80mm"
              >
                <Text
                  style={[
                    styles.paperWidthText,
                    selectedPaperWidth === 80 && styles.selectedText,
                  ]}
                >
                  80mm
                </Text>
              </TouchableOpacity>
            </View>
          </View>

          {/* Printer Darkness Adjustment */}
          <View style={styles.configItem}>
            <Text style={styles.configLabel}>Intensitas</Text>
            <View style={styles.darknessContainer}>
              <Text style={styles.darknessValue}>
                {Math.round(printerDarkness * 100)}%
              </Text>
              <Slider
                style={styles.darknessSlider}
                minimumValue={0}
                maximumValue={1}
                step={0.1}
                value={printerDarkness}
                onValueChange={setPrinterDarkness}
                testID="darkness-slider"
              />
            </View>
          </View>
        </View>

        {/* Profile Management Section */}
        <View style={styles.section} testID="profile-section">
          <Text style={styles.sectionTitle}>Manajemen Profil</Text>
          <View style={styles.profileButtons}>
            <TouchableOpacity
              style={styles.profileButton}
              onPress={handleSaveProfile}
              testID="save-profile-button"
            >
              <Text style={styles.profileButtonText}>💾 Simpan Profil</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={styles.profileButton}
              onPress={handleLoadProfile}
              testID="load-profile-button"
            >
              <Text style={styles.profileButtonText}>📂 Muat Profil</Text>
            </TouchableOpacity>
          </View>
        </View>

        {/* Test Print Section */}
        <View style={styles.section}>
          <TouchableOpacity
            style={[styles.testPrintButton, !currentPrinter && styles.disabledButton]}
            onPress={handleTestPrint}
            disabled={!currentPrinter || isTestPrinting}
            testID="test-print-button"
          >
            {isTestPrinting ? (
              <ActivityIndicator size="small" color="#fff" />
            ) : (
              <Text style={styles.testPrintButtonText}>🧪 Test Print</Text>
            )}
          </TouchableOpacity>
          {!currentPrinter && (
            <Text style={styles.hintText}>
              Hubungkan ke printer untuk test print
            </Text>
          )}
        </View>
      </ScrollView>

      {/* Disconnect Button */}
      {currentPrinter && (
        <View style={styles.footer}>
          <TouchableOpacity
            style={styles.disconnectButton}
            onPress={handleDisconnectPrinter}
            testID="disconnect-button"
          >
            <Text style={styles.disconnectButtonText}>
              Putus dari {currentPrinter.name}
            </Text>
          </TouchableOpacity>
        </View>
      )}
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F9FAFB',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    backgroundColor: '#fff',
    borderBottomWidth: 1,
    borderBottomColor: '#E5E7EB',
  },
  backButton: {
    width: 40,
    height: 40,
    justifyContent: 'center',
    marginRight: 12,
  },
  backButtonText: {
    fontSize: 24,
    color: '#374151',
  },
  title: {
    flex: 1,
    fontSize: 18,
    fontWeight: 'bold',
    color: '#111827',
  },
  content: {
    flex: 1,
    padding: 16,
  },
  section: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 16,
    borderWidth: 1,
    borderColor: '#E5E7EB',
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 12,
  },
  scanButton: {
    backgroundColor: '#2563EB',
    borderRadius: 8,
    padding: 12,
    alignItems: 'center',
  },
  scanButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  loadingContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: 12,
  },
  loadingText: {
    marginLeft: 8,
    color: '#6B7280',
    fontSize: 14,
  },
  printerList: {
    marginTop: 12,
  },
  printerItem: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#F3F4F6',
  },
  printerInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
  },
  printerIcon: {
    fontSize: 24,
    marginRight: 12,
  },
  printerDetails: {
    flex: 1,
  },
  printerName: {
    fontSize: 14,
    fontWeight: '600',
    color: '#111827',
  },
  printerType: {
    fontSize: 12,
    color: '#6B7280',
  },
  connectButton: {
    backgroundColor: '#10B981',
    borderRadius: 6,
    paddingVertical: 6,
    paddingHorizontal: 12,
  },
  connectButtonText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '600',
  },
  configItem: {
    marginBottom: 16,
  },
  configLabel: {
    fontSize: 14,
    fontWeight: '600',
    color: '#374151',
    marginBottom: 8,
  },
  paperWidthSelector: {
    flexDirection: 'row',
    gap: 8,
  },
  paperWidthOption: {
    flex: 1,
    padding: 12,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#D1D5DB',
    alignItems: 'center',
  },
  selectedOption: {
    backgroundColor: '#DBEAFE',
    borderColor: '#2563EB',
  },
  paperWidthText: {
    fontSize: 14,
    color: '#6B7280',
  },
  selectedText: {
    color: '#1E40AF',
    fontWeight: '600',
  },
  darknessContainer: {
    alignItems: 'center',
  },
  darknessValue: {
    fontSize: 14,
    color: '#374151',
    marginBottom: 8,
  },
  darknessSlider: {
    width: '100%',
  },
  profileButtons: {
    flexDirection: 'row',
    gap: 8,
  },
  profileButton: {
    flex: 1,
    backgroundColor: '#8B5CF6',
    borderRadius: 8,
    padding: 12,
    alignItems: 'center',
  },
  profileButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  testPrintButton: {
    backgroundColor: '#059669',
    borderRadius: 8,
    padding: 16,
    alignItems: 'center',
  },
  testPrintButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  disabledButton: {
    backgroundColor: '#D1D5DB',
  },
  hintText: {
    fontSize: 12,
    color: '#6B7280',
    textAlign: 'center',
    marginTop: 8,
  },
  footer: {
    padding: 16,
    backgroundColor: '#fff',
    borderTopWidth: 1,
    borderTopColor: '#E5E7EB',
  },
  disconnectButton: {
    backgroundColor: '#DC2626',
    borderRadius: 8,
    padding: 16,
    alignItems: 'center',
  },
  disconnectButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});

export default PrinterSettingsScreen;
