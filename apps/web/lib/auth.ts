/**
 * Authentication utility functions for simpo web dashboard
 */

import { setTokenCookie, clearTokenCookie } from './apiClient';
import type { User } from '@/types/api';

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface AuthState {
  isAuthenticated: boolean;
  user: LoginResponse['user'] | null;
}

/**
 * Login user with email and password
 *
 * @param credentials - User login credentials
 * @returns Promise with login response
 */
export async function login(credentials: LoginCredentials): Promise<LoginResponse> {
  // This will be implemented when backend auth endpoint is ready
  // For now, return mock response
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(credentials),
  });

  if (!response.ok) {
    // Provide specific error messages based on status code
    if (response.status === 401) {
      throw new Error('Invalid email or password');
    } else if (response.status === 500) {
      throw new Error('Server error. Please try again later');
    } else if (response.status === 0 || !navigator.onLine) {
      throw new Error('Network error. Please check your connection');
    } else {
      throw new Error(`Login failed: ${response.statusText}`);
    }
  }

  const data: LoginResponse = await response.json();

  // Store token in cookie
  setTokenCookie(data.token);

  return data;
}

/**
 * Logout user and clear token
 */
export function logout(): void {
  clearTokenCookie();

  // Redirect to login page
  if (typeof window !== 'undefined') {
    window.location.href = '/login';
  }
}

/**
 * Check if user is authenticated
 *
 * @returns true if user has valid token
 */
export function isAuthenticated(): boolean {
  if (typeof document === 'undefined') {
    return false;
  }

  const cookies = document.cookie.split(';');
  const tokenCookie = cookies.find(cookie =>
    cookie.trim().startsWith('token=')
  );

  return !!tokenCookie;
}

/**
 * Get current user from localStorage (for client-side state)
 *
 * @returns User data or null
 */
export function getCurrentUser(): AuthState['user'] | null {
  if (typeof window === 'undefined') {
    return null;
  }

  const userStr = localStorage.getItem('user');
  if (userStr) {
    try {
      return JSON.parse(userStr);
    } catch {
      // Clear corrupted data on parse failure
      localStorage.removeItem('user');
      return null;
    }
  }

  return null;
}

/**
 * Store current user in localStorage
 *
 * @param user - User data to store
 */
export function setCurrentUser(user: AuthState['user']): void {
  if (typeof window !== 'undefined') {
    localStorage.setItem('user', JSON.stringify(user));
  }
}

/**
 * Clear current user from localStorage
 */
export function clearCurrentUser(): void {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('user');
  }
}

// Re-export clearTokenCookie from apiClient for convenience
export { clearTokenCookie };
