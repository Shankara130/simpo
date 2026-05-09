'use client';

import { ReactNode } from 'react';
import { AuthProvider } from './AuthContext';

/**
 * AppProvider Component Props
 *
 * Wraps all context providers for the application.
 * Add additional providers here as the app grows.
 */
interface AppProviderProps {
  children: ReactNode;
}

/**
 * AppProvider Component
 *
 * Root provider wrapper that combines all context providers.
 * This should be used in the root layout to wrap the entire app.
 */
export function AppProvider({ children }: AppProviderProps) {
  return (
    <AuthProvider>
      {children}
      {/* Add more providers here as needed:
        - ThemeProvider
        - QueryClientProvider
        - NotificationProvider
        etc.
      */}
    </AuthProvider>
  );
}
