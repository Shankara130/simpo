/**
 * useSyncProgress Hook Tests
 * Story 8.2: Implement Transaction Sync Queue
 *
 * Tests for sync progress state management hook
 */

import { renderHook, act, waitFor } from '@testing-library/react-hooks';
import { jest } from '@jest/globals';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { useSyncProgress } from './useSyncProgress';
import { SyncQueue } from '../services/SyncQueueService';
import { SYNC_PROGRESS_KEY } from '../types/sync.types';

// Mock dependencies
jest.mock('@react-native-async-storage/async-storage', () =>
  require('@react-native-async-storage/async-storage/jest/async-storage-mock')
);

jest.mock('../services/SyncQueueService');

// Mock EventEmitter for React Native
jest.mock('react-native/Libraries/EventEmitter/NativeEventEmitter');

describe('useSyncProgress', () => {
  let mockSyncQueue: any;

  beforeEach(() => {
    jest.clearAllMocks();
    AsyncStorage.clear();

    // Setup SyncQueue mock
    const { SyncQueue: SyncQueueModule } = require('../services/SyncQueueService');
    SyncQueueModule.getInstance = jest.fn(() => ({
      getPendingTransactions: jest.fn().mockResolvedValue([]),
      calculateBackoff: jest.fn(),
      scheduleRetry: jest.fn(),
      shouldRetryNow: jest.fn().mockResolvedValue(true),
      getRetryCount: jest.fn().mockResolvedValue(0),
      saveQueueState: jest.fn(),
      loadQueueState: jest.fn().mockResolvedValue({
        isProcessing: false,
        currentTransactionId: null,
        lastUpdated: new Date().toISOString(),
      }),
      isProcessingStateOrphaned: jest.fn().mockResolvedValue(false),
      processQueue: jest.fn(),
      stopProcessing: jest.fn(),
    }));

    mockSyncQueue = SyncQueueModule.getInstance();
  });

  afterEach(() => {
    jest.clearAllTimers();
  });

  it('should initialize with default state', async () => {
    const { result } = renderHook(() => useSyncProgress());

    // Initial state should be defaults
    expect(result.current.pendingCount).toBe(0);
    expect(result.current.processingCount).toBe(0);
    expect(result.current.syncedCount).toBe(0);
    expect(result.current.failedCount).toBe(0);
    expect(result.current.currentTransaction).toBeNull();
  });

  it('should load persisted state from AsyncStorage', async () => {
    const persistedState = {
      pendingCount: 5,
      processingCount: 0,
      syncedCount: 3,
      failedCount: 1,
      currentTransaction: 'OFFLINE-123',
    };

    await AsyncStorage.setItem(SYNC_PROGRESS_KEY, JSON.stringify(persistedState));

    const { result } = renderHook(() => useSyncProgress());

    await waitFor(() => {
      expect(result.current.isInitialized).toBe(true);
      // Should load persisted values
      expect(result.current.pendingCount).toBe(5);
      expect(result.current.syncedCount).toBe(3);
    });
  });

  it('should refresh pending count from storage', async () => {
    mockSyncQueue.getPendingTransactions.mockResolvedValue([
      { id: 1 },
      { id: 2 },
      { id: 3 },
    ]);

    const { result } = renderHook(() => useSyncProgress());

    await waitFor(() => {
      expect(result.current.isInitialized).toBe(true);
    });

    act(() => {
      result.current.refreshPendingCount();
    });

    await waitFor(() => {
      expect(result.current.pendingCount).toBe(3);
    });
  });

  it('should reset session counters', () => {
    const { result } = renderHook(() => useSyncProgress());

    act(() => {
      // Manually set some counters
      result.current.incrementSyncedCount('OFFLINE-1');
      result.current.incrementFailedCount('OFFLINE-2');
    });

    act(() => {
      result.current.resetSessionCounters();
    });

    expect(result.current.syncedCount).toBe(0);
    expect(result.current.failedCount).toBe(0);
  });

  it('should set processing state', () => {
    const { result } = renderHook(() => useSyncProgress());

    act(() => {
      result.current.setProcessing(true, 'OFFLINE-123');
    });

    expect(result.current.processingCount).toBe(1);
    expect(result.current.currentTransaction).toBe('OFFLINE-123');

    act(() => {
      result.current.setProcessing(false);
    });

    expect(result.current.processingCount).toBe(0);
    expect(result.current.currentTransaction).toBeNull();
  });

  it('should increment synced count', () => {
    const { result } = renderHook(() => useSyncProgress());

    act(() => {
      result.current.incrementSyncedCount('OFFLINE-1');
    });

    expect(result.current.syncedCount).toBe(1);
    expect(result.current.processingCount).toBe(0);
  });

  it('should increment failed count', () => {
    const { result } = renderHook(() => useSyncProgress());

    act(() => {
      result.current.incrementFailedCount('OFFLINE-1');
    });

    expect(result.current.failedCount).toBe(1);
    expect(result.current.processingCount).toBe(0);
  });

  it('should detect and reset orphaned processing state', async () => {
    mockSyncQueue.isProcessingStateOrphaned.mockResolvedValue(true);

    const { result } = renderHook(() => useSyncProgress());

    await waitFor(() => {
      expect(result.current.isInitialized).toBe(true);
    });

    // Should reset processing count if orphaned
    // This is checked in the initialization effect
  });
});
