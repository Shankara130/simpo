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
};

export type TransactionHistoryParamList = {
  TransactionHistoryList: undefined;
  TransactionDetail: { transactionId: number };
};

export type POSStackScreenName = keyof POSStackParamList;
