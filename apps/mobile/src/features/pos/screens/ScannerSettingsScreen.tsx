/**
 * ScannerSettingsScreen Component
 * Configuration UI for USB and Bluetooth barcode scanner settings
 * Story 7.2: USB Barcode Scanner Integration
 * Story 7.3: Bluetooth Barcode Scanner Support
 */

import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TextInput,
  Switch,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import Icon from 'react-native-vector-icons/MaterialIcons';
import { ScannerConfig, DEFAULT_SCANNER_CONFIG, BluetoothDevice, BluetoothConnectionState, BluetoothConfig, DEFAULT_BLUETOOTH_CONFIG } from '../types/scanner.types';
import { ScannerConfigService } from '../services/ScannerConfigService';

type ScannerSettingsNavigationProp = StackNavigationProp<any, 'ScannerSettings'>;

interface ScannerSettingsScreenProps {
  // Props can be added for navigation parameters if needed
}

export const ScannerSettingsScreen: React.FC<ScannerSettingsScreenProps> = () => {
  const navigation = useNavigation<ScannerSettingsNavigationProp>();

  // Local state for settings
  const [config, setConfig] = useState<ScannerConfig>(DEFAULT_SCANNER_CONFIG);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  // Story 7.3: Bluetooth scanner state
  const [bluetoothConfig, setBluetoothConfig] = useState<BluetoothConfig>(DEFAULT_BLUETOOTH_CONFIG);
  const [pairedDevices, setPairedDevices] = useState<BluetoothDevice[]>([]);
  const [isScanning, setIsScanning] = useState(false);
  const [scanInProgress, setScanInProgress] = useState(false);

  // Load settings on mount
  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      setIsLoading(true);
      const savedConfig = await ScannerConfigService.load();
      setConfig(savedConfig);

      // Story 7.3: Load Bluetooth settings
      const savedBluetoothConfig = await ScannerConfigService.loadBluetoothConfig();
      setBluetoothConfig(savedBluetoothConfig);

      const savedPairedDevices = await ScannerConfigService.loadPairedDevices();
      setPairedDevices(savedPairedDevices);
    } catch (error) {
      console.error('Failed to load scanner config:', error);
      // Use defaults on error
      setConfig(DEFAULT_SCANNER_CONFIG);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      setIsSaving(true);

      // Validate debounce range (100ms - 2000ms)
      if (config.debounceMs < 100 || config.debounceMs > 2000) {
        Alert.alert('Invalid Range', 'Debounce interval harus antara 100ms - 2000ms');
        return;
      }

      // Validate barcode length ranges
      if (config.minBarcodeLength < 1 || config.minBarcodeLength > 20) {
        Alert.alert('Invalid Range', 'Min barcode length harus antara 1 - 20 karakter');
        return;
      }

      if (config.maxBarcodeLength < 8 || config.maxBarcodeLength > 50) {
        Alert.alert('Invalid Range', 'Max barcode length harus antara 8 - 50 karakter');
        return;
      }

      // Ensure min < max
      if (config.minBarcodeLength >= config.maxBarcodeLength) {
        Alert.alert('Invalid Range', 'Min barcode length harus kurang dari max barcode length');
        return;
      }

      // Save configuration
      await ScannerConfigService.save(config);

      Alert.alert(
        'Berhasil',
        'Pengaturan scanner berhasil disimpan',
        [
          {
            text: 'OK',
            onPress: () => navigation.goBack(),
          },
        ]
      );
    } catch (error) {
      console.error('Failed to save scanner config:', error);
      Alert.alert('Gagal', 'Gagal menyimpan pengaturan scanner');
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    Alert.alert(
      'Reset Pengaturan',
      'Apakah Anda yakin ingin mereset pengaturan scanner ke default?',
      [
        {
          text: 'Batal',
          style: 'cancel',
        },
        {
          text: 'Reset',
          style: 'destructive',
          onPress: async () => {
            try {
              await ScannerConfigService.reset();
              setConfig(DEFAULT_SCANNER_CONFIG);
              Alert.alert('Berhasil', 'Pengaturan scanner direset ke default');
            } catch (error) {
              Alert.alert('Gagal', 'Gagal mereset pengaturan scanner');
            }
          },
        },
      ]
    );
  };

  const handleTestScan = () => {
    Alert.alert(
      'Test Scan',
      'Arahkan kursor ke input field dan scan barcode untuk menguji scanner.\n\nPastikan:',
      [
        {
          text: 'Batal',
          style: 'cancel',
        },
        {
          text: 'OK',
          onPress: () => {
            // Navigate to POSScreen for testing
            navigation.navigate('POSScreen');
          },
        },
      ]
    );
  };

  // Story 7.3: Bluetooth scanner handlers
  const handleStartDiscovery = async () => {
    setScanInProgress(true);
    // Simulate device discovery (in production, would use BluetoothManager)
    setTimeout(() => {
      // Mock discovered devices
      setScanInProgress(false);
      Alert.alert('Discovery Complete', 'No new devices found (demo mode)');
    }, 3000);
  };

  const handleConnectDevice = async (device: BluetoothDevice) => {
    Alert.alert('Connect Device', `Connect to ${device.name}?`, [
      {
        text: 'Batal',
        style: 'cancel',
      },
      {
        text: 'Connect',
        onPress: async () => {
          // In production, would use BluetoothManager.connectToDevice()
          Alert.alert('Connected', `Connected to ${device.name}`);
        },
      },
    ]);
  };

  const handleDisconnectDevice = async (device: BluetoothDevice) => {
    Alert.alert('Disconnect Device', `Disconnect from ${device.name}?`, [
      {
        text: 'Batal',
        style: 'cancel',
      },
      {
        text: 'Disconnect',
        onPress: async () => {
          // In production, would use BluetoothManager.disconnectDevice()
          Alert.alert('Disconnected', `Disconnected from ${device.name}`);
        },
      },
    ]);
  };

  const handleForgetDevice = async (device: BluetoothDevice) => {
    Alert.alert('Forget Device', `Remove ${device.name} from paired devices?`, [
      {
        text: 'Batal',
        style: 'cancel',
      },
      {
        text: 'Forget',
        style: 'destructive',
        onPress: async () => {
          try {
            const updatedDevices = pairedDevices.filter((d) => d.id !== device.id);
            setPairedDevices(updatedDevices);
            await ScannerConfigService.savePairedDevices(updatedDevices);
            Alert.alert('Removed', `${device.name} has been removed`);
          } catch (error) {
            Alert.alert('Gagal', 'Failed to remove device');
          }
        },
      },
    ]);
  };

  const handleToggleAutoReconnect = async (value: boolean) => {
    try {
      const updatedConfig = { ...bluetoothConfig, autoReconnect: value };
      setBluetoothConfig(updatedConfig);
      await ScannerConfigService.saveBluetoothConfig(updatedConfig);
    } catch (error) {
      Alert.alert('Gagal', 'Failed to save Bluetooth settings');
    }
  };

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#4CAF50" />
        <Text style={styles.loadingText}>Memuat pengaturan...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity
          style={styles.backButton}
          onPress={() => navigation.goBack()}
          testID="back-button"
        >
          <Icon name="arrow-back" size={24} color="#333" />
        </TouchableOpacity>
        <Text style={styles.headerTitle}>Pengaturan Scanner</Text>
      </View>

      {/* Settings Form */}
      <ScrollView style={styles.content}>
        {/* Debounce Interval */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Debounce Scan</Text>
          <Text style={styles.sectionDescription}>
            Interval untuk mencegah scan ganda (millidetik)
          </Text>

          <View style={styles.inputContainer}>
            <Text style={styles.inputLabel}>Debounce (ms):</Text>
            <TextInput
              style={styles.input}
              value={config.debounceMs.toString()}
              onChangeText={(text) => {
                const value = parseInt(text, 10);
                if (!isNaN(value) && value >= 100 && value <= 2000) {
                  setConfig({ ...config, debounceMs: value });
                }
              }}
              keyboardType="number-pad"
              testID="debounce-input"
            />
            <Text style={styles.inputHint}>
              Range: 100 - 2000 ms (Default: {DEFAULT_SCANNER_CONFIG.debounceMs} ms)
            </Text>
          </View>
        </View>

        {/* Barcode Length Settings */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Panjang Barcode</Text>
          <Text style={styles.sectionDescription}>
            Konfigurasi panjang barcode minimum dan maksimum
          </Text>

          <View style={styles.inputContainer}>
            <Text style={styles.inputLabel}>Minimum Karakter:</Text>
            <TextInput
              style={styles.input}
              value={config.minBarcodeLength.toString()}
              onChangeText={(text) => {
                const value = parseInt(text, 10);
                if (!isNaN(value) && value >= 1 && value <= 20) {
                  setConfig({ ...config, minBarcodeLength: value });
                }
              }}
              keyboardType="number-pad"
              testID="min-length-input"
            />
            <Text style={styles.inputHint}>
              Range: 1 - 20 (Default: {DEFAULT_SCANNER_CONFIG.minBarcodeLength})
            </Text>
          </View>

          <View style={styles.inputContainer}>
            <Text style={styles.inputLabel}>Maksimum Karakter:</Text>
            <TextInput
              style={styles.input}
              value={config.maxBarcodeLength.toString()}
              onChangeText={(text) => {
                const value = parseInt(text, 10);
                if (!isNaN(value) && value >= 8 && value <= 50) {
                  setConfig({ ...config, maxBarcodeLength: value });
                }
              }}
              keyboardType="number-pad"
              testID="max-length-input"
            />
            <Text style={styles.inputHint}>
              Range: 8 - 50 (Default: {DEFAULT_SCANNER_CONFIG.maxBarcodeLength})
            </Text>
          </View>
        </View>

        {/* Feedback Settings */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Feedback</Text>
          <Text style={styles.sectionDescription}>
            Konfigurasi feedback visual dan haptic saat scan
          </Text>

          <View style={styles.switchContainer}>
            <Text style={styles.switchLabel}>Getar & Suara</Text>
            <Switch
              value={config.feedbackEnabled}
              onValueChange={(value) => setConfig({ ...config, feedbackEnabled: value })}
              testID="feedback-switch"
            />
            <Text style={styles.switchHint}>
              Aktifkan getaran dan feedback saat scan berhasil
            </Text>
          </View>
        </View>

        {/* Scan Time Settings */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Waktu Scan Maksimum</Text>
          <Text style={styles.sectionDescription}>
            Waktu maksimum untuk deteksi scanner vs keyboard (ms)
          </Text>

          <View style={styles.inputContainer}>
            <Text style={styles.inputLabel}>Max Scan Time (ms):</Text>
            <TextInput
              style={styles.input}
              value={config.maxScanTimeMs.toString()}
              onChangeText={(text) => {
                const value = parseInt(text, 10);
                if (!isNaN(value) && value >= 50 && value <= 500) {
                  setConfig({ ...config, maxScanTimeMs: value });
                }
              }}
              keyboardType="number-pad"
              testID="max-scan-time-input"
            />
            <Text style={styles.inputHint}>
              Range: 50 - 500 ms (Default: {DEFAULT_SCANNER_CONFIG.maxScanTimeMs} ms)
            </Text>
          </View>
        </View>

        {/* Story 7.3: Bluetooth Scanner Section */}
        <View style={styles.section}>
          <View style={styles.sectionHeader}>
            <Icon name="bluetooth" size={24} color="#4CAF50" />
            <View style={styles.sectionHeaderContent}>
              <Text style={styles.sectionTitle}>Bluetooth Scanner</Text>
              <Text style={styles.sectionDescription}>
                Kelola perangkat scanner Bluetooth yang terhubung
              </Text>
            </View>
          </View>

          {/* Auto-reconnect toggle */}
          <View style={styles.switchContainer}>
            <View style={styles.switchContent}>
              <Text style={styles.switchLabel}>Auto-reconnect</Text>
              <Text style={styles.switchHint}>
                Otomatis reconnect ke scanner terakhir saat aplikasi dibuka
              </Text>
            </View>
            <Switch
              value={bluetoothConfig.autoReconnect}
              onValueChange={handleToggleAutoReconnect}
              testID="bluetooth-auto-reconnect-switch"
            />
          </View>

          {/* Device discovery button */}
          <TouchableOpacity
            style={[styles.actionButton, styles.discoveryButton]}
            onPress={handleStartDiscovery}
            disabled={scanInProgress}
            testID="bluetooth-discovery-button"
          >
            {scanInProgress ? (
              <>
                <ActivityIndicator size="small" color="#4CAF50" />
                <Text style={styles.actionButtonText}>Scanning...</Text>
              </>
            ) : (
              <>
                <Icon name="search" size={20} color="#4CAF50" />
                <Text style={styles.actionButtonText}>Cari Perangkat Baru</Text>
              </>
            )}
          </TouchableOpacity>

          {/* Paired devices list */}
          {pairedDevices.length > 0 && (
            <View style={styles.deviceListContainer}>
              <Text style={styles.deviceListTitle}>Perangkat Terhubung</Text>
              {pairedDevices.map((device) => (
                <View key={device.id} style={styles.deviceItem}>
                  <View style={styles.deviceInfo}>
                    <Icon
                      name={device.connected ? 'bluetooth-connected' : 'bluetooth'}
                      size={24}
                      color={device.connected ? '#4CAF50' : '#999'}
                    />
                    <View style={styles.deviceDetails}>
                      <Text style={styles.deviceName}>{device.name}</Text>
                      <Text style={styles.deviceStatus}>
                        {device.connected ? 'Terhubung' : 'Tidak Terhubung'}
                      </Text>
                    </View>
                  </View>
                  <View style={styles.deviceActions}>
                    {device.connected ? (
                      <TouchableOpacity
                        style={styles.deviceActionBtn}
                        onPress={() => handleDisconnectDevice(device)}
                        testID={`disconnect-${device.id}`}
                      >
                        <Icon name="bluetooth-disabled" size={20} color="#F44336" />
                      </TouchableOpacity>
                    ) : (
                      <TouchableOpacity
                        style={styles.deviceActionBtn}
                        onPress={() => handleConnectDevice(device)}
                        testID={`connect-${device.id}`}
                      >
                        <Icon name="bluetooth" size={20} color="#4CAF50" />
                      </TouchableOpacity>
                    )}
                    <TouchableOpacity
                      style={styles.deviceActionBtn}
                      onPress={() => handleForgetDevice(device)}
                      testID={`forget-${device.id}`}
                    >
                      <Icon name="delete" size={20} color="#999" />
                    </TouchableOpacity>
                  </View>
                </View>
              ))}
            </View>
          )}

          {/* Troubleshooting guide */}
          <TouchableOpacity
            style={styles.troubleshootButton}
            onPress={() => {
              Alert.alert(
                'Troubleshooting Bluetooth',
                'Jika scanner Bluetooth tidak terdeteksi:\n\n' +
                '1. Pastikan Bluetooth aktif di perangkat\n' +
                '2. Pastikan scanner dalam mode pairing\n' +
                '3. Coba restart scanner\n' +
                '4. Clear cache Bluetooth di pengaturan sistem\n' +
                '5. Pastikan izin Bluetooth diberikan'
              );
            }}
          >
            <Icon name="help-outline" size={18} color="#4CAF50" />
            <Text style={styles.troubleshootText}>Troubleshooting Bluetooth</Text>
          </TouchableOpacity>
        </View>

        {/* Test Scan Button */}
        <View style={styles.section}>
          <TouchableOpacity
            style={styles.testButton}
            onPress={handleTestScan}
            testID="test-scan-button"
          >
            <Icon name="qr-code-scanner" size={24} color="#4CAF50" />
            <Text style={styles.testButtonText}>Test Scanner</Text>
          </TouchableOpacity>
          <Text style={styles.testButtonHint}>
            Buka layar POS untuk menguji scanner barcode
          </Text>
        </View>
      </ScrollView>

      {/* Footer Actions */}
      <View style={styles.footer}>
        <TouchableOpacity
          style={[styles.footerButton, styles.resetButton]}
          onPress={handleReset}
          testID="reset-button"
        >
          <Text style={styles.resetButtonText}>Reset ke Default</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.footerButton, styles.saveButton]}
          onPress={handleSave}
          disabled={isSaving}
          testID="save-button"
        >
          {isSaving ? (
            <ActivityIndicator size="small" color="#FFF" />
          ) : (
            <Text style={styles.saveButtonText}>Simpan Pengaturan</Text>
          )}
        </TouchableOpacity>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },

  // Loading state
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#F5F5F5',
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
    color: '#666',
  },

  // Header
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    backgroundColor: '#FFF',
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
  },
  backButton: {
    padding: 8,
    marginRight: 8,
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
  },

  // Content
  content: {
    flex: 1,
    padding: 16,
  },

  // Sections
  section: {
    backgroundColor: '#FFF',
    borderRadius: 8,
    padding: 16,
    marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginBottom: 4,
  },
  sectionDescription: {
    fontSize: 14,
    color: '#666',
    marginBottom: 16,
  },

  // Input
  inputContainer: {
    marginBottom: 16,
  },
  inputLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#333',
    marginBottom: 8,
  },
  input: {
    borderWidth: 1,
    borderColor: '#E0E0E0',
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
    backgroundColor: '#FFF',
  },
  inputHint: {
    fontSize: 12,
    color: '#999',
    marginTop: 4,
  },

  // Switch
  switchContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  switchLabel: {
    fontSize: 16,
    color: '#333',
  },
  switchHint: {
    fontSize: 12,
    color: '#999',
    marginTop: 8,
    flex: 1,
    marginLeft: 16,
  },

  // Test button
  testButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#FFF',
    borderWidth: 2,
    borderColor: '#4CAF50',
    borderRadius: 8,
    padding: 16,
  },

  // Story 7.3: Bluetooth styles
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
  },
  sectionHeaderContent: {
    marginLeft: 12,
    flex: 1,
  },
  switchContent: {
    flex: 1,
  },
  actionButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: 8,
    padding: 12,
    marginTop: 12,
    borderWidth: 1,
  },
  discoveryButton: {
    backgroundColor: '#FFF',
    borderColor: '#4CAF50',
  },
  actionButtonText: {
    fontSize: 14,
    fontWeight: '600',
    marginLeft: 8,
  },
  deviceListContainer: {
    marginTop: 16,
  },
  deviceListTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#333',
    marginBottom: 8,
  },
  deviceItem: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: '#F5F5F5',
    borderRadius: 8,
    padding: 12,
    marginBottom: 8,
  },
  deviceInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
  },
  deviceDetails: {
    marginLeft: 12,
  },
  deviceName: {
    fontSize: 14,
    fontWeight: '500',
    color: '#333',
  },
  deviceStatus: {
    fontSize: 12,
    color: '#666',
    marginTop: 2,
  },
  deviceActions: {
    flexDirection: 'row',
    gap: 8,
  },
  deviceActionBtn: {
    padding: 8,
  },
  troubleshootButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: 12,
    paddingVertical: 8,
  },
  troubleshootText: {
    fontSize: 12,
    color: '#4CAF50',
    marginLeft: 4,
  },
  testButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#4CAF50',
    marginLeft: 8,
  },
  testButtonHint: {
    fontSize: 12,
    color: '#666',
    marginTop: 8,
    textAlign: 'center',
  },

  // Footer
  footer: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    paddingVertical: 12,
    backgroundColor: '#FFF',
    borderTopWidth: 1,
    borderTopColor: '#E0E0E0',
    gap: 8,
  },
  footerButton: {
    flex: 1,
    paddingVertical: 12,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
  },
  resetButton: {
    backgroundColor: '#FFF',
    borderWidth: 1,
    borderColor: '#E0E0E0',
  },
  resetButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#666',
  },
  saveButton: {
    backgroundColor: '#4CAF50',
  },
  saveButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFF',
  },
});
