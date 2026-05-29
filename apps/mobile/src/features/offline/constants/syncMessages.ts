/**
 * Sync Messages - Indonesian Language Constants
 * Story 8.4: Implement Visual Sync Status Indicators
 *
 * All user-facing sync status messages in Indonesian language
 * Provides message templates and error translations
 */

/**
 * Main sync status messages
 */
export const SYNC_MESSAGES = {
  WAITING_FOR_INTERNET: 'Menunggu koneksi internet...',
  SYNCING: 'Sync dalam proses...',
  SYNCED: 'Semua data ter-sync',
  SYNC_FAILED: 'Gagal sync',
  PENDING_TRANSACTIONS: '{count} transaksi pending',
  RETRY_COUNTDOWN: 'Retry otomatis dalam {minutes} menit',
  SYNC_NOW: 'Sync Sekarang',
  CLOSE: 'Tutup',
  LAST_SYNC: 'Terakhir sync: {time}',
  SYNC_COMPLETE: 'Sync selesai - {count} transaksi berhasil',
  SYNC_PROGRESS: 'Sync: {pending} pending, {synced} synced',
} as const;

/**
 * Sync phase descriptions in Indonesian
 */
export const SYNC_PHASE_MESSAGES = {
  idle: null,
  uploading: 'Mengupload transaksi...',
  downloading_stock: 'Download data stok...',
  downloading_products: 'Download produk baru...',
  downloading_user: 'Download data pengguna...',
  synced: 'Semua data ter-sync',
  failed: 'Sync gagal',
} as const;

/**
 * Technical error message translations
 * Maps backend error codes/types to user-friendly Indonesian messages
 */
export const ERROR_MESSAGES = {
  NETWORK_ERROR: 'Error jaringan',
  SERVER_ERROR: 'Error server',
  INSUFFICIENT_STOCK: 'Stok tidak cukup',
  TIMEOUT_ERROR: 'Request timeout',
  VALIDATION_ERROR: 'Data tidak valid',
  CONFLICT_ERROR: 'Konflik data',
  UNKNOWN_ERROR: 'Error tidak diketahui',
} as const;

/**
 * Error message translation utility
 * Translates technical error types to user-friendly Indonesian messages
 */
export function translateError(errorType: string): string {
  const errorMap: Record<string, string> = {
    'network': ERROR_MESSAGES.NETWORK_ERROR,
    'server': ERROR_MESSAGES.SERVER_ERROR,
    'insufficient_stock': ERROR_MESSAGES.INSUFFICIENT_STOCK,
    'timeout': ERROR_MESSAGES.TIMEOUT_ERROR,
    'validation': ERROR_MESSAGES.VALIDATION_ERROR,
    'conflict': ERROR_MESSAGES.CONFLICT_ERROR,
  };

  const normalizedType = errorType.toLowerCase();
  return errorMap[normalizedType] || ERROR_MESSAGES.UNKNOWN_ERROR;
}

/**
 * Format timestamp for Indonesian locale
 * @param timestamp - ISO 8601 timestamp string
 * @returns Formatted time string (e.g., "10:30")
 */
export function formatTimestamp(timestamp: string | null): string {
  if (!timestamp) {
    return '-';
  }

  try {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('id-ID', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    });
  } catch (error) {
    console.warn('[syncMessages] Failed to format timestamp:', error);
    return '-';
  }
}

/**
 * Format countdown as Indonesian minutes string
 * @param minutes - Number of minutes
 * @returns Formatted countdown (e.g., "5 menit")
 */
export function formatCountdown(minutes: number): string {
  if (minutes <= 0) {
    return '0 menit';
  }
  return `${minutes} menit`;
}

/**
 * Format pending transaction count
 * @param count - Number of pending transactions
 * @returns Formatted count message (e.g., "5 transaksi pending")
 */
export function formatPendingCount(count: number): string {
  return SYNC_MESSAGES.PENDING_TRANSACTIONS.replace('{count}', String(count));
}

/**
 * Format retry countdown message
 * @param minutes - Number of minutes until retry
 * @returns Formatted retry message (e.g., "Retry otomatis dalam 5 menit")
 */
export function formatRetryCountdown(minutes: number): string {
  return SYNC_MESSAGES.RETRY_COUNTDOWN.replace('{minutes}', String(minutes));
}

/**
 * Format sync completion message
 * @param count - Number of synced transactions
 * @returns Formatted completion message (e.g., "Sync selesai - 5 transaksi berhasil")
 */
export function formatSyncComplete(count: number): string {
  return SYNC_MESSAGES.SYNC_COMPLETE.replace('{count}', String(count));
}

/**
 * Format sync progress message
 * @param pending - Number of pending transactions
 * @param synced - Number of synced transactions
 * @returns Formatted progress message (e.g., "Sync: 3 pending, 5 synced")
 */
export function formatSyncProgress(pending: number, synced: number): string {
  return SYNC_MESSAGES.SYNC_PROGRESS
    .replace('{pending}', String(pending))
    .replace('{synced}', String(synced));
}

/**
 * Get message template by key
 * @param key - Message key
 * @param params - Optional parameters for template interpolation
 * @returns Formatted message
 */
export function getMessage(key: keyof typeof SYNC_MESSAGES, params?: Record<string, string | number>): string {
  let message = SYNC_MESSAGES[key];

  if (params) {
    Object.entries(params).forEach(([paramKey, value]) => {
      message = message.replace(`{${paramKey}}`, String(value));
    });
  }

  return message;
}
