/**
 * AuthContext - Authentication State Management
 *
 * Placeholder for future authentication implementation.
 * Will integrate with backend JWT tokens from Story 1.1 (GRAB boilerplate).
 *
 * Context will provide:
 * - User authentication state
 * - Login/logout actions
 * - JWT token management
 * - Session timeout handling
 */

import React, { createContext, useContext, useReducer, ReactNode } from 'react';

// Auth state types
export interface AuthState {
  isAuthenticated: boolean;
  user: null | {
    id: string;
    username: string;
    role: string;
  };
  token: null | string;
}

// Auth action types
export type AuthAction =
  | { type: 'LOGIN_SUCCESS'; payload: { user: AuthState['user']; token: string } }
  | { type: 'LOGOUT' }
  | { type: 'SESSION_EXPIRED' }
  | { type: 'RESTORE_SESSION'; payload: { token: string; user: AuthState['user'] } };

// Initial state
const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  token: null,
};

// Auth reducer
function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'LOGIN_SUCCESS':
      return {
        ...state,
        isAuthenticated: true,
        user: action.payload.user,
        token: action.payload.token,
      };
    case 'LOGOUT':
    case 'SESSION_EXPIRED':
      return initialState;
    case 'RESTORE_SESSION':
      return {
        ...state,
        isAuthenticated: true,
        user: action.payload.user,
        token: action.payload.token,
      };
    default:
      return state;
  }
}

// Context type
interface AuthContextType {
  state: AuthState;
  dispatch: React.Dispatch<AuthAction>;
}

// Create context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Provider component
export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(authReducer, initialState);

  return (
    <AuthContext.Provider value={{ state, dispatch }}>
      {children}
    </AuthContext.Provider>
  );
}

// Custom hook to use auth context
export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
