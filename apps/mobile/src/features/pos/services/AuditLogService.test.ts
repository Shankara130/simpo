/**
 * Audit Log Service Tests
 * Tests for audit trail logging service with offline queue support
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import {
  AuditLogService,
  getAuditLogService,
  AuditEventType,
} from './AuditLogService';

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
}));

// Mock fetch API
global.fetch = jest.fn() as jest.Mock;

describe('AuditLogService', () => {
  let service: AuditLogService;

  beforeEach(() => {
    jest.clearAllMocks();
    // Reset before creating new instance
    AuditLogService.resetInstance();
    service = getAuditLogService('https://api.example.com');
  });

  afterEach(() => {
    AuditLogService.resetInstance();
  });

  describe('Initialization', () => {
    it('should create singleton instance', () => {
      const instance1 = getAuditLogService();
      const instance2 = getAuditLogService();

      expect(instance1).toBe(instance2);
    });

    it('should create service with API base URL', () => {
      const serviceWithUrl = getAuditLogService('https://api.example.com');

      expect(serviceWithUrl).toBeDefined();
    });
  });

  describe('Cash Drawer Logging', () => {
    it('should log successful cash drawer open event', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      });

      await service.logCashDrawerOpen('TRX-123', 'user-1', 'success', {
        drawerPin: 0,
        pulseTiming: 100,
      });

      expect(AsyncStorage.setItem).toHaveBeenCalled();
      expect(global.fetch).toHaveBeenCalledWith(
        'https://api.example.com/api/v1/audit',
        expect.objectContaining({
          method: 'POST',
        })
      );
    });

    it('should log failed cash drawer open event', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      });

      await service.logCashDrawerOpen('TRX-123', 'user-1', 'failed', {
        error: 'Drawer disconnected',
      });

      expect(AsyncStorage.setItem).toHaveBeenCalled();
      expect(global.fetch).toHaveBeenCalledWith(
        'https://api.example.com/api/v1/audit',
        expect.objectContaining({
          method: 'POST',
        })
      );
    });

    it('should queue event on API failure', async () => {
      (AsyncStorage.getItem as jest.Mock)
        .mockResolvedValueOnce(null) // getLocalLogs
        .mockResolvedValueOnce('[]'); // getQueue
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);
      (global.fetch as jest.Mock).mockRejectedValueOnce(new Error('Network error'));

      await service.logCashDrawerOpen('TRX-123', 'user-1', 'success');

      // Should queue the event for later sync
      expect(AsyncStorage.setItem).toHaveBeenCalledWith(
        '@simpo_audit_log_queue',
        expect.stringContaining('TRX-123')
      );
    });

    it('should work without user ID', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      });

      await service.logCashDrawerOpen('TRX-123', undefined, 'success');

      expect(global.fetch).toHaveBeenCalled();
    });

    it('should work without metadata', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      });

      await service.logCashDrawerOpen('TRX-123', 'user-1', 'success');

      expect(global.fetch).toHaveBeenCalled();
    });
  });

  describe('Local Storage', () => {
    it('should save event locally', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      });

      await service.logCashDrawerOpen('TRX-123', 'user-1', 'success');

      expect(AsyncStorage.setItem).toHaveBeenCalledWith(
        '@simpo_audit_logs',
        expect.stringContaining('TRX-123')
      );
    });

    it('should retrieve local logs', async () => {
      const mockLogs = [
        {
          id: 'audit_123',
          eventType: AuditEventType.DRAWER_OPEN,
          transactionId: 'TRX-123',
          status: 'success',
          timestamp: '2026-05-28T10:00:00Z',
        },
      ];

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(JSON.stringify(mockLogs));

      const logs = await service.getLocalLogs();

      expect(logs).toEqual(mockLogs);
      expect(AsyncStorage.getItem).toHaveBeenCalledWith('@simpo_audit_logs');
    });

    it('should return empty array when no logs exist', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);

      const logs = await service.getLocalLogs();

      expect(logs).toEqual([]);
    });

    it('should limit local logs to MAX_LOCAL_LOGS', async () => {
      const existingLogs = Array(1000).fill(null).map((_, i) => ({
        id: `audit_${i}`,
        eventType: AuditEventType.DRAWER_OPEN,
        status: 'success',
        timestamp: '2026-05-28T10:00:00Z',
      }));

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(JSON.stringify(existingLogs));
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      });

      await service.logCashDrawerOpen('TRX-123', 'user-1', 'success');

      expect(AsyncStorage.setItem).toHaveBeenCalled();
      const savedLogs = JSON.parse((AsyncStorage.setItem as jest.Mock).mock.calls[0][1] as string);
      expect(savedLogs.length).toBe(1000);
    });
  });

  describe('Queue Management', () => {
    it.skip('should sync queued events', async () => {
      const queuedEvent = {
        id: 'audit_queued',
        eventType: AuditEventType.DRAWER_OPEN,
        transactionId: 'TRX-123',
        status: 'success' as const,
        timestamp: '2026-05-28T10:00:00Z',
      };

      // Set up mocks for the sync process
      let callCount = 0;
      (AsyncStorage.getItem as jest.Mock).mockImplementation((key: string) => {
        callCount++;
        if (key === '@simpo_audit_logs') {
          return Promise.resolve(null);
        } else if (key === '@simpo_audit_log_queue') {
          if (callCount === 2) {
            // First call to getQueue - return the queued event
            return Promise.resolve(JSON.stringify([queuedEvent]));
          } else {
            // Subsequent calls - return empty queue
            return Promise.resolve(JSON.stringify([]));
          }
        }
        return Promise.resolve(null);
      });

      (AsyncStorage.setItem as jest.Mock).mockResolvedValue(undefined);
      (AsyncStorage.removeItem as jest.Mock).mockResolvedValue(undefined);
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ success: true }),
      });

      const syncedCount = await service.syncQueuedEvents();

      expect(syncedCount).toBe(1);
      expect(global.fetch).toHaveBeenCalled();
    });

    it.skip('should return 0 when queue is empty', async () => {
      // Mock getItem to return empty queue
      (AsyncStorage.getItem as jest.Mock).mockImplementation((key: string) => {
        if (key === '@simpo_audit_log_queue') {
          return Promise.resolve(null); // Empty queue
        }
        return Promise.resolve(null);
      });

      const syncedCount = await service.syncQueuedEvents();

      expect(syncedCount).toBe(0);
    });

    it.skip('should stop syncing on error', async () => {
      const queuedEvents = [
        {
          id: 'audit_1',
          eventType: AuditEventType.DRAWER_OPEN,
          transactionId: 'TRX-123',
          status: 'success' as const,
          timestamp: '2026-05-28T10:00:00Z',
        },
        {
          id: 'audit_2',
          eventType: AuditEventType.DRAWER_OPEN,
          transactionId: 'TRX-124',
          status: 'success' as const,
          timestamp: '2026-05-28T10:00:00Z',
        },
      ];

      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(JSON.stringify(queuedEvents));
      (global.fetch as jest.Mock).mockRejectedValueOnce(new Error('Network error'));

      const syncedCount = await service.syncQueuedEvents();

      expect(syncedCount).toBe(0);
    });
  });

  describe('Clear Operations', () => {
    it('should clear local logs', async () => {
      (AsyncStorage.removeItem as jest.Mock).mockResolvedValueOnce(undefined);

      await service.clearLocalLogs();

      expect(AsyncStorage.removeItem).toHaveBeenCalledWith('@simpo_audit_logs');
    });

    it('should clear queue', async () => {
      (AsyncStorage.removeItem as jest.Mock).mockResolvedValueOnce(undefined);

      await service.clearQueue();

      expect(AsyncStorage.removeItem).toHaveBeenCalledWith('@simpo_audit_log_queue');
    });
  });

  describe('Event Structure', () => {
    it('should create event with correct structure for drawer open', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      });

      await service.logCashDrawerOpen('TRX-123', 'user-1', 'success', {
        drawerPin: 0,
        pulseTiming: 100,
      });

      const savedData = JSON.parse((AsyncStorage.setItem as jest.Mock).mock.calls[0][1] as string);
      const event = savedData[0];

      expect(event.eventType).toBe(AuditEventType.DRAWER_OPEN);
      expect(event.transactionId).toBe('TRX-123');
      expect(event.userId).toBe('user-1');
      expect(event.status).toBe('success');
      expect(event.timestamp).toBeDefined();
      expect(event.metadata).toEqual({
        drawerPin: 0,
        pulseTiming: 100,
      });
    });

    it('should create event with correct structure for drawer failed', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      });

      await service.logCashDrawerOpen('TRX-123', 'user-1', 'failed', {
        error: 'Drawer disconnected',
      });

      const savedData = JSON.parse((AsyncStorage.setItem as jest.Mock).mock.calls[0][1] as string);
      const event = savedData[0];

      expect(event.eventType).toBe(AuditEventType.DRAWER_FAILED);
      expect(event.status).toBe('failed');
      expect(event.metadata).toEqual({
        error: 'Drawer disconnected',
      });
    });

    it('should generate unique event IDs', async () => {
      (AsyncStorage.getItem as jest.Mock)
        .mockResolvedValueOnce(null) // First call - getLocalLogs
        .mockResolvedValueOnce(null) // Second call - getLocalLogs
        .mockResolvedValueOnce('[]') // Third call - getQueue
        .mockResolvedValueOnce('[]'); // Fourth call - getQueue
      (AsyncStorage.setItem as jest.Mock).mockResolvedValue(undefined);
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ success: true }),
      });

      await service.logCashDrawerOpen('TRX-123', 'user-1', 'success');
      await service.logCashDrawerOpen('TRX-124', 'user-1', 'success');

      const savedData1 = JSON.parse((AsyncStorage.setItem as jest.Mock).mock.calls[0][1] as string);
      const event1 = savedData1[0];
      const savedData2 = JSON.parse((AsyncStorage.setItem as jest.Mock).mock.calls[1][1] as string);
      const event2 = savedData2[0];

      expect(event1.id).not.toBe(event2.id);
    });
  });

  describe('API Integration', () => {
    it('should send event to backend API', async () => {
      (AsyncStorage.getItem as jest.Mock).mockResolvedValueOnce(null);
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      });

      await service.logCashDrawerOpen('TRX-123', 'user-1', 'success');

      expect(global.fetch).toHaveBeenCalledWith(
        'https://api.example.com/api/v1/audit',
        expect.objectContaining({
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
        })
      );

      const fetchBody = JSON.parse((global.fetch as jest.Mock).mock.calls[0][1].body as string);
      expect(fetchBody.transactionId).toBe('TRX-123');
      expect(fetchBody.eventType).toBe(AuditEventType.DRAWER_OPEN);
    });

    it('should handle API errors gracefully', async () => {
      (AsyncStorage.getItem as jest.Mock)
        .mockResolvedValueOnce(null)
        .mockResolvedValueOnce('[]');
      (AsyncStorage.setItem as jest.Mock).mockResolvedValueOnce(undefined);
      (global.fetch as jest.Mock).mockRejectedValueOnce(new Error('Network error'));

      await service.logCashDrawerOpen('TRX-123', 'user-1', 'success');

      // Should queue the event instead of throwing
      expect(AsyncStorage.setItem).toHaveBeenCalled();
    });
  });
});
