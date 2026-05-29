/**
 * Mock for expo-sqlite
 * Used in tests for offline storage functionality
 */

export const openDatabaseAsync = jest.fn().mockResolvedValue({
  execAsync: jest.fn().mockResolvedValue(undefined),
  runAsync: jest.fn().mockResolvedValue({ lastInsertRowId: 1, changes: 1 }),
  getAsync: jest.fn().mockResolvedValue({ user_version: 1 }),
  getAllAsync: jest.fn().mockResolvedValue([]),
  getFirstAsync: jest.fn().mockResolvedValue(undefined),
  closeAsync: jest.fn().mockResolvedValue(undefined),
});

export const openDatabaseSync = jest.fn();

export default {
  openDatabaseAsync,
  openDatabaseSync,
};
