'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode, useCallback, useRef } from 'react';
import { User, LoginResponse } from '@/types/api';
import * as auth from '@/lib/auth';

/**
 * Authentication context state
 */
interface AuthContextType {
  isAuthenticated: boolean;
  user: User | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  loading: boolean;
}

/**
 * Authentication Context
 *
 * Provides authentication state and actions throughout the app.
 * Uses React Context for client-side state management (Next.js Server Components compatible).
 */
const AuthContext = createContext<AuthContextType | undefined>(undefined);

/**
 * Custom hook to use auth context
 */
export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}

/**
 * AuthProvider Component Props
 */
interface AuthProviderProps {
  children: ReactNode;
}

/**
 * AuthProvider Component
 *
 * Wraps the app to provide authentication context.
 * This is a client component (marked with 'use client' directive).
 */
export function AuthProvider({ children }: AuthProviderProps) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [isLoggingIn, setIsLoggingIn] = useState(false);

  /**
   * Initialize auth state on mount with validation
   */
  useEffect(() => {
    const initAuth = () => {
      // Check if user is authenticated via token cookie
      const authenticated = auth.isAuthenticated();
      const currentUser = auth.getCurrentUser();

      // Validate token and user data together - clear both if inconsistent
      if (authenticated && !currentUser) {
        // Token exists but no user data - clear token
        auth.clearTokenCookie();
        setIsAuthenticated(false);
        setUser(null);
      } else if (!authenticated && currentUser) {
        // User data exists but no token - clear user data
        auth.clearCurrentUser();
        setUser(null);
        setIsAuthenticated(false);
      } else {
        // Both consistent or both empty
        setIsAuthenticated(authenticated);
        setUser(currentUser);
      }

      setLoading(false);
    };

    initAuth();
  }, []);

  /**
   * Login function with race condition guard
   */
  const login = useCallback(async (email: string, password: string) => {
    // Guard clause to prevent concurrent login attempts
    if (isLoggingIn) {
      return;
    }

    setIsLoggingIn(true);
    setLoading(true);

    try {
      // Call login API
      const response: LoginResponse = await auth.login({ email, password });

      // Update state
      setIsAuthenticated(true);
      setUser(response.user);

      // Store user in localStorage
      auth.setCurrentUser(response.user);
    } catch (error) {
      throw error;
    } finally {
      setLoading(false);
      setIsLoggingIn(false);
    }
  }, [isLoggingIn]);

  /**
   * Logout function
   */
  const logout = useCallback(() => {
    auth.logout();
    setIsAuthenticated(false);
    setUser(null);
    auth.clearCurrentUser();
  }, []);

  const value: AuthContextType = {
    isAuthenticated,
    user,
    login,
    logout,
    loading,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

/**
 * Higher-order component to protect authenticated routes with proper redirect handling
 */
export function withAuth<P extends object>(
  Component: React.ComponentType<P>
): React.ComponentType<P> {
  return function AuthenticatedComponent(props: P) {
    const { isAuthenticated, loading } = useAuth();
    const redirectAttempted = useRef(false);

    // Handle redirect with proper cleanup to avoid memory leaks
    useEffect(() => {
      if (!loading && !isAuthenticated && !redirectAttempted.current) {
        redirectAttempted.current = true;
        // Use window.location.href for redirect (non-SPA behavior for auth)
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
      }
    }, [loading, isAuthenticated]);

    if (loading) {
      return (
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Loading...</p>
          </div>
        </div>
      );
    }

    if (!isAuthenticated) {
      // Return loading state during redirect
      return (
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
          </div>
        </div>
      );
    }

    return <Component {...props} />;
  };
}
