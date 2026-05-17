/**
 * TransactionFilterModal Component
 * Modal for filtering transaction history by date range and status
 * Story 3.7: Transaction History Filter Modal
 * AC5: Date Range and Status Filtering
 */

import React, { useState, useEffect } from 'react';
import {
  View,
  StyleSheet,
  Modal,
  TouchableOpacity,
  Text,
  SafeAreaView,
  ScrollView,
} from 'react-native';
import Icon from 'react-native-vector-icons/MaterialIcons';
import {
  TransactionFilters,
  DateRangePreset,
} from '../types/transactionHistory.types';

interface TransactionFilterModalProps {
  visible: boolean;
  onClose: () => void;
  onApply: (filters: TransactionFilters) => void;
  initialFilters: TransactionFilters;
}

/**
 * Get date range for preset
 */
const getDateRangeForPreset = (preset: DateRangePreset): {
  startDate: Date;
  endDate: Date | null;
} => {
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());

  switch (preset) {
    case 'today':
      return { startDate: today, endDate: null };

    case 'yesterday':
      const yesterday = new Date(today);
      yesterday.setDate(yesterday.getDate() - 1);
      return { startDate: yesterday, endDate: yesterday };

    case 'thisWeek':
      const firstDayOfWeek = new Date(today);
      const dayOfWeek = firstDayOfWeek.getDay();
      const diff = firstDayOfWeek.getDate() - dayOfWeek + (dayOfWeek === 0 ? -6 : 1);
      firstDayOfWeek.setDate(diff);
      return { startDate: firstDayOfWeek, endDate: null };

    case 'thisMonth':
      const firstDayOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
      return { startDate: firstDayOfMonth, endDate: null };

    case 'custom':
      return { startDate: today, endDate: null };

    default:
      return { startDate: today, endDate: null };
  }
};

/**
 * Format date to Indonesian locale
 */
const formatDate = (date: Date): string => {
  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
};

/**
 * Date Range Preset Options
 */
const DATE_RANGE_PRESETS: { value: DateRangePreset; label: string }[] = [
  { value: 'today', label: 'Hari Ini' },
  { value: 'yesterday', label: 'Kemarin' },
  { value: 'thisWeek', label: 'Minggu Ini' },
  { value: 'thisMonth', label: 'Bulan Ini' },
  { value: 'custom', label: 'Kustom' },
];

/**
 * Status Filter Options
 */
const STATUS_OPTIONS: { value: TransactionFilters['status']; label: string }[] = [
  { value: 'ALL', label: 'Semua Status' },
  { value: 'COMPLETED', label: 'Selesai' },
  { value: 'CANCELLED', label: 'Batal' },
  { value: 'PENDING', label: 'Tertunda' },
];

/**
 * Transaction Filter Modal Component
 */
export const TransactionFilterModal: React.FC<TransactionFilterModalProps> = ({
  visible,
  onClose,
  onApply,
  initialFilters,
}) => {
  // Local state for filters
  const [filters, setFilters] = useState<TransactionFilters>(initialFilters);
  const [selectedPreset, setSelectedPreset] = useState<DateRangePreset>('today');

  /**
   * Update local state when initial filters change
   */
  useEffect(() => {
    setFilters(initialFilters);
  }, [initialFilters]);

  /**
   * Handle date range preset selection
   */
  const handlePresetSelect = (preset: DateRangePreset) => {
    setSelectedPreset(preset);
    const { startDate, endDate } = getDateRangeForPreset(preset);
    setFilters((prev) => ({ ...prev, startDate, endDate }));
  };

  /**
   * Handle status selection
   */
  const handleStatusSelect = (status: TransactionFilters['status']) => {
    setFilters((prev) => ({ ...prev, status }));
  };

  /**
   * Handle reset filters
   */
  const handleReset = () => {
    const defaultFilters: TransactionFilters = {
      startDate: new Date(),
      endDate: null,
      status: 'ALL',
    };
    setFilters(defaultFilters);
    setSelectedPreset('today');
  };

  /**
   * Handle apply filters
   */
  const handleApply = () => {
    onApply(filters);
    onClose();
  };

  return (
    <Modal
      visible={visible}
      animationType="slide"
      transparent
      onRequestClose={onClose}
    >
      <SafeAreaView style={styles.modalContainer}>
        <View style={styles.modalContent}>
          {/* Header */}
          <View style={styles.header}>
            <Text style={styles.headerTitle}>Filter Transaksi</Text>
            <TouchableOpacity onPress={onClose}>
              <Icon name="close" size={24} color="#212121" />
            </TouchableOpacity>
          </View>

          {/* Content */}
          <ScrollView style={styles.content}>
            {/* Date Range Section */}
            <View style={styles.section}>
              <Text style={styles.sectionTitle}>Rentang Tanggal</Text>
              {DATE_RANGE_PRESETS.map((preset) => (
                <TouchableOpacity
                  key={preset.value}
                  style={[
                    styles.optionButton,
                    selectedPreset === preset.value && styles.optionButtonActive,
                  ]}
                  onPress={() => handlePresetSelect(preset.value)}
                >
                  <View
                    style={[
                      styles.radioIcon,
                      selectedPreset === preset.value && styles.radioIconActive,
                    ]}
                  >
                    {selectedPreset === preset.value && <View style={styles.radioDot} />}
                  </View>
                  <Text
                    style={[
                      styles.optionLabel,
                      selectedPreset === preset.value && styles.optionLabelActive,
                    ]}
                  >
                    {preset.label}
                  </Text>
                </TouchableOpacity>
              ))}

              {/* Custom Date Display */}
              {filters.startDate && (
                <View style={styles.dateDisplay}>
                  <Text style={styles.dateLabel}>Dari:</Text>
                  <Text style={styles.dateValue}>{formatDate(filters.startDate)}</Text>
                </View>
              )}
              {filters.endDate && (
                <View style={styles.dateDisplay}>
                  <Text style={styles.dateLabel}>Sampai:</Text>
                  <Text style={styles.dateValue}>{formatDate(filters.endDate)}</Text>
                </View>
              )}
            </View>

            {/* Status Section */}
            <View style={styles.section}>
              <Text style={styles.sectionTitle}>Status</Text>
              {STATUS_OPTIONS.map((option) => (
                <TouchableOpacity
                  key={option.value}
                  style={[
                    styles.optionButton,
                    filters.status === option.value && styles.optionButtonActive,
                  ]}
                  onPress={() => handleStatusSelect(option.value)}
                >
                  <View
                    style={[
                      styles.radioIcon,
                      filters.status === option.value && styles.radioIconActive,
                    ]}
                  >
                    {filters.status === option.value && <View style={styles.radioDot} />}
                  </View>
                  <Text
                    style={[
                      styles.optionLabel,
                      filters.status === option.value && styles.optionLabelActive,
                    ]}
                  >
                    {option.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>
          </ScrollView>

          {/* Footer Buttons */}
          <View style={styles.footer}>
            <TouchableOpacity style={styles.resetButton} onPress={handleReset}>
              <Text style={styles.resetButtonText}>Reset</Text>
            </TouchableOpacity>
            <TouchableOpacity style={styles.applyButton} onPress={handleApply}>
              <Text style={styles.applyButtonText}>Terapkan</Text>
            </TouchableOpacity>
          </View>
        </View>
      </SafeAreaView>
    </Modal>
  );
};

/**
 * Styles
 */
const styles = StyleSheet.create({
  modalContainer: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'flex-end',
  },
  modalContent: {
    backgroundColor: '#FFFFFF',
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    maxHeight: '80%',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#212121',
  },
  content: {
    flex: 1,
  },
  section: {
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#F0F0F0',
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#757575',
    marginBottom: 12,
    textTransform: 'uppercase',
  },
  optionButton: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderRadius: 8,
    marginBottom: 8,
  },
  optionButtonActive: {
    backgroundColor: '#E3F2FD',
  },
  radioIcon: {
    width: 20,
    height: 20,
    borderRadius: 10,
    borderWidth: 2,
    borderColor: '#BDBDBD',
    marginRight: 12,
    justifyContent: 'center',
    alignItems: 'center',
  },
  radioIconActive: {
    borderColor: '#2196F3',
  },
  radioDot: {
    width: 10,
    height: 10,
    borderRadius: 5,
    backgroundColor: '#2196F3',
  },
  optionLabel: {
    fontSize: 14,
    color: '#212121',
  },
  optionLabelActive: {
    color: '#2196F3',
    fontWeight: '500',
  },
  dateDisplay: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
    paddingHorizontal: 16,
    marginLeft: 32,
  },
  dateLabel: {
    fontSize: 14,
    color: '#757575',
    marginRight: 8,
  },
  dateValue: {
    fontSize: 14,
    color: '#212121',
    fontWeight: '500',
  },
  footer: {
    flexDirection: 'row',
    padding: 16,
    borderTopWidth: 1,
    borderTopColor: '#E0E0E0',
    gap: 12,
  },
  resetButton: {
    flex: 1,
    paddingVertical: 12,
    borderWidth: 1,
    borderColor: '#2196F3',
    borderRadius: 8,
    alignItems: 'center',
  },
  resetButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#2196F3',
  },
  applyButton: {
    flex: 1,
    paddingVertical: 12,
    backgroundColor: '#2196F3',
    borderRadius: 8,
    alignItems: 'center',
  },
  applyButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFFFFF',
  },
});
