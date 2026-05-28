/**
 * Navigation type definitions for POS feature
 * Defines navigation params and screen names
 */

import type { NavigatorScreenParams } from '@react-navigation/native';

export type POSStackParamList = {
  POS: undefined;
  TransactionHistory: NavigatorScreenParams<TransactionHistoryParamList> | undefined;
  TransactionDetail: { transactionId: number };
  Settings: undefined;
  ScannerSettings: undefined; // Story 7.2: USB Barcode Scanner Integration
};

export type TransactionHistoryParamList = {
  TransactionHistoryList: undefined;
  TransactionDetail: { transactionId: number };
};

export type POSStackScreenName = keyof POSStackParamList;
