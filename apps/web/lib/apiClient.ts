import axios, { AxiosError, AxiosResponse } from 'axios';

// API base URL from environment variable or default to localhost
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081/api/v1';

/**
 * RFC 7807 Error Detail Format
 */
interface ApiErrorDetail {
  type: string;
  title: string;
  status: number;
  detail: string;
  instance?: string;
}

/**
 * Custom API Error Class for RFC 7807 errors
 */
export class ApiError extends Error {
  public readonly type: string;
  public readonly title: string;
  public readonly status: number;
  public readonly instance?: string;

  constructor(detail: ApiErrorDetail) {
    super(detail.detail);
    this.name = 'ApiError';
    this.type = detail.type;
    this.title = detail.title;
    this.status = detail.status;
    this.instance = detail.instance;
  }
}

/**
 * Axios API Client configured for simpo backend
 */
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

/**
 * Request interceptor: Add JWT token from cookies
 */
apiClient.interceptors.request.use(
  (config) => {
    // For client-side, token will be stored in cookies
    // Cookie is automatically sent by browser
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

/**
 * Response interceptor: Handle RFC 7807 error responses
 */
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    // Return data directly for consistency
    return response.data;
  },
  (error: AxiosError) => {
    if (error.response?.data && typeof error.response.data === 'object') {
      const data = error.response.data as Record<string, unknown>;

      // Check if error follows RFC 7807 format with proper type validation
      if (
        typeof data.type === 'string' &&
        typeof data.title === 'string' &&
        typeof data.status === 'number' &&
        typeof data.detail === 'string'
      ) {
        throw new ApiError({
          type: data.type,
          title: data.title,
          status: data.status,
          detail: data.detail,
          instance: typeof data.instance === 'string' ? data.instance : undefined,
        });
      }
    }

    // Handle network errors or other errors
    if (error.code === 'ECONNABORTED') {
      return Promise.reject(new Error('Request timeout. Please check your connection.'));
    }

    if (!error.response) {
      return Promise.reject(new Error('Network error. Please check if the server is running.'));
    }

    return Promise.reject(error);
  }
);

export default apiClient;

/**
 * Helper function to get token from cookie (for client-side)
 * Note: In production, cookies should be httpOnly for security
 */
export function getTokenFromCookie(): string | null {
  if (typeof document === 'undefined') {
    return null;
  }

  const cookies = document.cookie.split(';');
  const tokenCookie = cookies.find(cookie => {
    const [name] = cookie.trim().split('=');
    return name === 'token';
  });

  if (!tokenCookie) return null;

  // Handle cookies with = in value by joining after first split
  const [, ...valueParts] = tokenCookie.split('=');
  return valueParts.join('=') || null;
}

/**
 * Helper function to set cookie (for client-side after login)
 * Note: In production, cookies should be set by backend with httpOnly flag
 * Client-side cannot set HttpOnly, but we add SameSite for CSRF protection
 */
export function setTokenCookie(token: string): void {
  if (typeof document !== 'undefined') {
    // Set cookie with security flags (Note: HttpOnly can only be set server-side)
    const isSecure = window.location.protocol === 'https:';
    document.cookie = `token=${token}; path=/; max-age=28800; SameSite=Strict${isSecure ? '; Secure' : ''}`;
  }
}

/**
 * Helper function to clear cookie (for logout)
 */
export function clearTokenCookie(): void {
  if (typeof document !== 'undefined') {
    document.cookie = 'token=; path=/; max-age=0';
  }
}
