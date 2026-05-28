/**
 * Mock for expo Permissions
 * Used for testing permission requests
 */

export const Permissions = {
  askAsync: jest.fn().mockResolvedValue({ status: 'granted' }),
};

export const Constants = {
  platform: {
    android: {},
    ios: {},
  },
};
