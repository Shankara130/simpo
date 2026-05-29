module.exports = {
  preset: '@react-native/jest-preset',
  setupFilesAfterEnv: ['@testing-library/jest-native/extend-expect'],
  transformIgnorePatterns: [
    'node_modules/(?!(@react-native|react-native|@react-navigation|@react-native-community|@testing-library|@finan-me|react-native-fs|uuid|expo|expo-sqlite|expo-modules-core))',
  ],
  moduleNameMapper: {
    '^@react-native-async-storage/async-storage$': '<rootDir>/src/__mocks__/@react-native-async-storage/async-storage',
    '^react-native-vector-icons/(.*)$': '<rootDir>/src/__mocks__/react-native-vector-icons/$1',
    '^@finan-me/react-native-thermal-printer$': '<rootDir>/src/__mocks__/@finan-me/react-native-thermal-printer',
    '^expo-sqlite$': '<rootDir>/src/__mocks__/expo-sqlite',
    '^./OfflineStorageService$': '<rootDir>/src/features/offline/services/OfflineStorageService.mock',
    '^./SyncAPI$': '<rootDir>/src/features/offline/services/SyncAPI.mock',
  },
  // Note: @testing-library/jest-native is deprecated but kept for compatibility
  // Future: migrate to built-in Jest matchers in @testing-library/react-native v12.4+
  testMatch: [
    '**/__tests__/**/*.ts?(x)',
    '**/?(*.)+(spec|test).ts?(x)',
  ],
  testPathIgnorePatterns: [
    '/node_modules/',
    '/android/',
    '/ios/',
  ],
};
