/**
 * useBidirectionalSync Hook Tests
 * Story 8.3: Implement Bidirectional Data Synchronization
 *
 * Test coverage for bidirectional sync state management
 */

import { renderHook, act, waitFor } from '@testing-library/react-native';
import { useBidirectionalSync } from './useBidirectionalSync';
import AsyncStorage from '@react-native-async-storage/async-storage';

// Mock dependencies
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
}));

jest.mock('../services/SyncOrchestrator', () => ({
  SyncOrchestrator: {
    getInstance: jest.fn(() => mockOrchestrator),
  },
}));

const mockOrchestrator = {
  sync: jest.fn(),
  stopSync: jest.fn(),
  getSyncState: jest.fn(),
  updatePhase: jest.fn(),
  calculateBackoff: jest.fn(),
  scheduleRetry: jest.fn(),
};

describe('useBidirectionalSync', () => {
  const mockGetItem = AsyncStorage.getItem as jest.Mock;
  const mockSetItem = AsyncStorage.setItem as jest.Mock;

  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();

    // Default AsyncStorage behavior
    mockGetItem.mockResolvedValue(null);
    mockSetItem.mockResolvedValue(undefined);

    // Default orchestrator behavior
    mockOrchestrator.sync.mockResolvedValue({
      success: true,
      phase: 'synced',
      uploaded: 1,
      downloadedStock: 0,
      downloadedProducts: 0,
      userUpdated: false,
      duration: 1000,
    });

    mockOrchestrator.getSyncState.mockResolvedValue({
      status: 'idle',
      phase: 'idle',
      pendingCount: 0,
      processingCount: 0,
      syncedCount: 0,
      failedCount: 0,
      currentPhase: null,
      lastSyncTime: null,
    });
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('should initialize with default state', async () => {
    const { result } = renderHook(() => useBidirectionalSync());

    expect(result.current.status).toBe('idle');
    expect(result.current.phase).toBe('idle');
    expect(result.current.isInitialized).toBe(false);
  });

  it('should load persisted state on mount', async () => {
    const savedState = {
      status: 'syncing' as const,
      phase: 'uploading' as const,
      pendingCount: 5,
      processingCount: 1,
      syncedCount: 3,
      failedCount: 1,
      currentPhase: 'Uploading pending transactions',
      lastSyncTime: '2026-05-29T10:00:00Z',
    };

    mockGetItem.mockResolvedValue(JSON.stringify(savedState));

    const { result, waitForNextUpdate } = renderHook(() => useBidirectionalSync());

    // Wait for initialization
    await waitForNextUpdate();

    expect(result.current.status).toBe('syncing');
    expect(result.current.phase).toBe('uploading');
  });

  it('should start sync and update state', async () => {
    mockOrchestrator.sync.mockResolvedValue({
      success: true,
      phase: 'synced',
      uploaded: 2,
      downloadedStock: 10,
      downloadedProducts: 5,
      userUpdated: true,
      duration: 5000,
    });

    const { result } = renderHook(() => useBidirectionalSync());

    await act(async () => {
      await result.current.startSync();
    });

    expect(mockOrchestrator.sync).toHaveBeenCalledTimes(1);
    expect(result.current.status).toBe('synced');
    expect(result.current.syncedCount).toBe(2);
  });

  it('should handle sync failure', async () => {
    mockOrchestrator.sync.mockResolvedValue({
      success: false,
      phase: 'failed',
      uploaded: 0,
      downloadedStock: 0,
      downloadedProducts: 0,
      userUpdated: false,
      duration: 1000,
      error: 'Network error',
    });

    const { result } = renderHook(() => useBidirectionalSync());

    await act(async () => {
      await result.current.startSync();
    });

    expect(result.current.status).toBe('failed');
    expect(result.current.phase).toBe('failed');
    expect(result.current.error).toBe('Network error');
  });

  it('should stop sync', () => {
    const { result } = renderHook(() => useBidirectionalSync());

    act(() => {
      result.current.stopSync();
    });

    expect(mockOrchestrator.stopSync).toHaveBeenCalledTimes(1);
  });
});
