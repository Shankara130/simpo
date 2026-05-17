/**
 * POSNavigator - Stack Navigator for POS Flow
 * Configures navigation for POS screen, Transaction History, and Settings
 */

import React from 'react';
import { createStackNavigator } from '@react-navigation/stack';
import { POSScreen } from '../screens/POSScreen';
import { TransactionHistoryScreen } from '../screens/TransactionHistoryScreen';
import { TransactionDetailScreen } from '../screens/TransactionDetailScreen';
import { POSStackParamList } from '../types/navigation.types';

// Placeholder screens for future stories
const SettingsScreen = () => null;

const Stack = createStackNavigator<POSStackParamList>();

export const POSNavigator: React.FC = () => {
  return (
    <Stack.Navigator
      initialRouteName="POS"
      screenOptions={{
        headerShown: true,
        headerStyle: {
          backgroundColor: '#2196F3',
        },
        headerTintColor: '#FFFFFF',
        headerTitleStyle: {
          fontWeight: '600',
        },
      }}
    >
      <Stack.Screen
        name="POS"
        component={POSScreen}
        options={{
          title: 'Point of Sale',
          headerShown: false, // POSScreen has its own header
        }}
      />
      <Stack.Screen
        name="TransactionHistory"
        component={TransactionHistoryScreen}
        options={{
          title: 'Transaction History',
        }}
      />
      <Stack.Screen
        name="TransactionDetail"
        component={TransactionDetailScreen}
        options={{
          title: 'Transaction Detail',
          headerShown: false, // TransactionDetailScreen has its own header
        }}
      />
      <Stack.Screen
        name="Settings"
        component={SettingsScreen}
        options={{
          title: 'Settings',
        }}
      />
    </Stack.Navigator>
  );
};
