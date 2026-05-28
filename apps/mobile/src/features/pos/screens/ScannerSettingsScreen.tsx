/**
 * ScannerSettingsScreen Component
 * Configuration UI for USB barcode scanner settings
 * Story 7.2: USB Barcode Scanner Integration
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
import { ScannerConfig, DEFAULT_SCANNER_CONFIG } from '../types/scanner.types';
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

  // Load settings on mount
  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      setIsLoading(true);
      const savedConfig = await ScannerConfigService.load();
      setConfig(savedConfig);
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
