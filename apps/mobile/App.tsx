/**
 * simpo POS Application
 * Point of Sale system for Indonesian SME pharmacies
 *
 * @format
 */

import React from 'react';
import { StatusBar, StyleSheet, useColorScheme, View, Text } from 'react-native';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { NavigationContainer } from '@react-navigation/native';
import { CartProvider } from './src/features/pos/context/CartContext';
import { POSNavigator } from './src/features/pos/navigation/POSNavigator';

// Error Boundary component for catching component errors
class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { hasError: boolean; error: Error | null }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): { hasError: boolean; error: Error | null } {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo): void {
    console.error('Error boundary caught an error:', error, errorInfo);
  }

  render(): React.ReactNode {
    if (this.state.hasError) {
      return (
        <View style={styles.errorContainer}>
          <Text style={styles.errorTitle}>Something went wrong</Text>
          <Text style={styles.errorMessage}>
            {this.state.error?.message || 'An unexpected error occurred'}
          </Text>
          <Text style={styles.errorNote}>Please restart the app</Text>
        </View>
      );
    }

    return this.props.children;
  }
}

function App(): React.JSX.Element {
  const isDarkMode = useColorScheme() === 'dark';

  return (
    <SafeAreaProvider>
      <StatusBar barStyle={isDarkMode ? 'light-content' : 'dark-content'} />
      <ErrorBoundary>
        <CartProvider>
          <NavigationContainer>
            <POSNavigator />
          </NavigationContainer>
        </CartProvider>
      </ErrorBoundary>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
    backgroundColor: '#FFF5F5',
  },
  errorTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#D32F2F',
    marginBottom: 8,
  },
  errorMessage: {
    fontSize: 16,
    color: '#212121',
    textAlign: 'center',
    marginBottom: 16,
  },
  errorNote: {
    fontSize: 14,
    color: '#757575',
    fontStyle: 'italic',
  },
});

export default App;
