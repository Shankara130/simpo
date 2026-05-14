/**
 * Printer Status Component
 * Displays printer connection status and provides visual feedback
 */

import React from 'react';
import { View, Text, StyleSheet, ActivityIndicator } from 'react-native';
import { PrinterStatus } from '../hardware/printer';

/**
 * Props for PrinterStatus component
 */
export interface PrinterStatusProps {
  /** Current printer status */
  status: PrinterStatus;
  /** Printer name (optional) */
  printerName?: string;
  /** Custom error message (for ERROR status) */
  errorMessage?: string;
  /** Whether to show compact view (for tight spaces) */
  compact?: boolean;
  /** Custom test ID for testing */
  testID?: string;
}

/**
 * Get status display configuration
 */
function getStatusConfig(status: PrinterStatus): {
  icon: string;
  label: string;
  color: string;
  bgColor: string;
} {
  switch (status) {
    case PrinterStatus.CONNECTED:
      return {
        icon: '✓',
        label: 'Printer Terhubung',
        color: '#16A34A',  // green-600
        bgColor: '#DCFCE7',  // green-100
      };
    case PrinterStatus.DISCONNECTED:
      return {
        icon: '○',
        label: 'Printer Terputus',
        color: '#6B7280',  // gray-500
        bgColor: '#F3F4F6',  // gray-100
      };
    case PrinterStatus.CONNECTING:
      return {
        icon: '…',
        label: 'Menghubungkan...',
        color: '#2563EB',  // blue-600
        bgColor: '#DBEAFE',  // blue-100
      };
    case PrinterStatus.ERROR:
      return {
        icon: '!',
        label: 'Printer Error',
        color: '#DC2626',  // red-600
        bgColor: '#FEE2E2',  // red-100
      };
    case PrinterStatus.OUT_OF_PAPER:
      return {
        icon: '⚠',
        label: 'Kertas Habis',
        color: '#F59E0B',  // amber-600
        bgColor: '#FEF3C7',  // amber-100
      };
    default:
      return {
        icon: '?',
        label: 'Status Unknown',
        color: '#6B7280',
        bgColor: '#F3F4F6',
      };
  }
}

/**
 * Printer Status Component
 */
export const PrinterStatusComponent: React.FC<PrinterStatusProps> = ({
  status,
  printerName,
  errorMessage,
  compact = false,
  testID = 'printer-status',
}) => {
  const config = getStatusConfig(status);
  const isLoading = status === PrinterStatus.CONNECTING;

  if (compact) {
    return (
      <View
        style={[
          styles.compactContainer,
          { backgroundColor: config.bgColor },
        ]}
        testID={testID}
        accessibilityLabel={`Printer status: ${config.label}`}
        accessibilityRole="text"
      >
        {isLoading ? (
          <ActivityIndicator
            size="small"
            color={config.color}
            style={styles.compactIndicator}
          />
        ) : (
          <View
            style={[
              styles.compactIndicator,
              { backgroundColor: config.color },
            ]}
          />
        )}
        <Text style={[styles.compactText, { color: config.color }]}>
          {config.label}
        </Text>
      </View>
    );
  }

  return (
    <View
      style={[styles.container, { backgroundColor: config.bgColor }]}
      testID={testID}
      accessibilityLabel={`Printer status: ${config.label}`}
      accessibilityRole="text"
    >
      <View style={styles.iconContainer}>
        {isLoading ? (
          <ActivityIndicator size="small" color={config.color} />
        ) : (
          <Text style={[styles.icon, { color: config.color }]}>
            {config.icon}
          </Text>
        )}
      </View>

      <View style={styles.textContainer}>
        <Text style={[styles.label, { color: config.color }]}>
          {config.label}
        </Text>

        {printerName && status !== PrinterStatus.DISCONNECTED && (
          <Text style={styles.printerName}>{printerName}</Text>
        )}

        {status === PrinterStatus.ERROR && errorMessage && (
          <Text style={styles.errorMessage}>{errorMessage}</Text>
        )}
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 12,
    borderRadius: 8,
    marginVertical: 4,
  },
  compactContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 6,
    paddingHorizontal: 10,
    borderRadius: 6,
    alignSelf: 'flex-start',
  },
  iconContainer: {
    width: 24,
    height: 24,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 12,
  },
  icon: {
    fontSize: 18,
    fontWeight: 'bold',
  },
  compactIndicator: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: 8,
  },
  textContainer: {
    flex: 1,
  },
  label: {
    fontSize: 14,
    fontWeight: '600',
  },
  compactText: {
    fontSize: 12,
    fontWeight: '500',
  },
  printerName: {
    fontSize: 12,
    color: '#6B7280',
    marginTop: 2,
  },
  errorMessage: {
    fontSize: 12,
    color: '#DC2626',
    marginTop: 2,
  },
});

export default PrinterStatusComponent;
