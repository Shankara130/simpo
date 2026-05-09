/**
 * AppProvider - Root Context Provider
 *
 * Wraps all React Context providers for the application.
 * Add additional providers here as the app grows.
 *
 * Current providers:
 * - AuthProvider: Authentication state management
 *
 * Future providers may include:
 * - CartProvider: Shopping cart for POS
 * - InventoryProvider: Real-time inventory state
 * - ThemeProvider: UI theme management
 */

import React from 'react';
import { AuthProvider } from './AuthContext';

export function AppProvider({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      {children}
    </AuthProvider>
  );
}
