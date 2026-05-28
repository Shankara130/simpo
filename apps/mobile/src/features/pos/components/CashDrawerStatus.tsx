/**
 * Cash Drawer Status Component
 * Displays cash drawer connection status in POSScreen
 * Story 7.4: Cash Drawer Control via Printer Kick
 */

import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { DrawerStatus } from '../../hardware/printer';

interface CashDrawerStatusProps {
  status: DrawerStatus;
  error?: string | null;
}

export const CashDrawerStatus: React.FC<CashDrawerStatusProps> = ({ status, error }) => {
  const getStatusText = (): string => {
    switch (status) {
      case 'connected':
        return 'Laci Uang: Terhubung';
      case 'disconnected':
        return 'Laci Uang: Terputus';
      case 'opening':
        return 'Laci Uang: Membuka...';
      case 'failed':
        return error ? 'Laci Uang: Gagal' : 'Laci Uang: Gagal';
      default:
        return 'Laci Uang: Tidak Diketahui';
    }
  };

  const getStatusColor = (): string => {
    switch (status) {
      case 'connected':
        return '#4CAF50'; // Green
      case 'disconnected':
        return '#9E9E9E'; // Gray
      case 'opening':
        return '#FF9800'; // Orange
      case 'failed':
        return '#F44336'; // Red
      default:
        return '#9E9E9E'; // Gray
    }
  };

  return (
    <View style={styles.container}>
      <Text style={[styles.statusText, { color: getStatusColor() }]}>
        {getStatusText()}
      </Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
    backgroundColor: '#FFFFFF',
  },
  statusText: {
    fontSize: 12,
    fontWeight: '600',
  },
});
