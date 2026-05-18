/**
 * InventoryService - API integration for Inventory management
 * Story 4.1: Product list with search, filters, and pagination
 */

import axios, { AxiosError } from 'axios';

// API base URL
const API_BASE_URL = __DEV__
  ? 'http://localhost:8080'
  : 'https://api.simpo.id';

const API_VERSION = '/api/v1';
const FULL_API_URL = `${API_BASE_URL}${API_VERSION}`;

/**
 * Product Item in list view
 * Story 4.1, AC4: Product display fields
 */
export interface ProductListItem {
  id: number;
  sku: string;
  name: string;
  description?: string;
  stockQty: number;
  price: string;
  expiryDate?: string;
  branchId: number;
  category?: string;
  reorderThreshold: number;
  isLowStock: boolean;    // Story 4.1, AC5: Low stock indicator
  isExpired: boolean;     // Story 4.1, AC6: Expired indicator
  createdAt: string;
  updatedAt: string;
}

/**
 * Pagination metadata
 * Story 4.1, AC7: Pagination support
 */
export interface PaginationMetadata {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

/**
 * Product list response
 * Story 4.1, AC1, AC7: List with pagination
 */
export interface ProductListResponse {
  data: ProductListItem[];
  pagination: PaginationMetadata;
}

/**
 * Product list request parameters
 * Story 4.1, AC2, AC3: Search and filter parameters
 */
export interface ProductListParams {
  search?: string;      // Search by name or SKU (AC2)
  category?: string;    // Filter by category (AC3)
  branch_id?: number;   // Filter by branch (AC3 - Owner only)
  low_stock?: boolean;  // Filter for low stock items
  expired?: boolean;    // Filter for expired items
  page?: number;        // Page number (AC7)
  limit?: number;       // Items per page (AC7)
  sort_by?: string;     // Field to sort by
  sort_order?: string;  // "asc" or "desc"
}

/**
 * API Error class for inventory service errors
 */
export class InventoryServiceError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public originalError?: any
  ) {
    super(message);
    this.name = 'InventoryServiceError';
  }
}

/**
 * InventoryService - Inventory API operations
 * Story 4.1: Product list management
 */
export const InventoryService = {
  /**
   * Fetch product list with search, filters, and pagination
   * Story 4.1, AC1, AC2, AC3, AC7
   */
  listProducts: async (params: ProductListParams = {}): Promise<ProductListResponse> => {
    try {
      const {
        page = 1,
        limit = 20,
        search,
        category,
        branch_id,
        low_stock,
        expired,
        sort_by = 'created_at',
        sort_order = 'desc'
      } = params;

      // Build query parameters
      const queryParams: Record<string, any> = {
        page,
        limit,
        sort_by,
        sort_order,
      };

      if (search) {
        queryParams.search = search;
      }

      if (category) {
        queryParams.category = category;
      }

      if (branch_id !== undefined) {
        queryParams.branch_id = branch_id;
      }

      if (low_stock !== undefined) {
        queryParams.low_stock = low_stock;
      }

      if (expired !== undefined) {
        queryParams.expired = expired;
      }

      const response = await axios.get<ProductListResponse>(
        `${FULL_API_URL}/products`,
        {
          params: queryParams,
          timeout: 10000, // 10 second timeout (NFR-PERF-005: <3s dashboard load)
        }
      );

      return response.data;
    } catch (error) {
      throw InventoryService._handleError(error, 'Failed to fetch products');
    }
  },

  /**
   * Search products by name or SKU
   * Story 4.1, AC2: Search functionality
   */
  searchProducts: async (query: string, page = 1, limit = 20): Promise<ProductListResponse> => {
    if (!query || query.trim().length === 0) {
      return InventoryService.listProducts({ page, limit });
    }

    return InventoryService.listProducts({
      search: query.trim(),
      page,
      limit,
    });
  },

  /**
   * Get low stock products
   * Story 4.1, AC5: Low stock indicator
   */
  getLowStockProducts: async (page = 1, limit = 20): Promise<ProductListResponse> => {
    return InventoryService.listProducts({
      low_stock: true,
      page,
      limit,
    });
  },

  /**
   * Handle API errors and convert to InventoryServiceError
   * @private
   */
  _handleError: (error: unknown, defaultMessage: string) => InventoryServiceError => {
    if (error instanceof InventoryServiceError) {
      return error;
    }

    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;

      // Network error
      if (!axiosError.response) {
        return new InventoryServiceError(
          'Network error. Please check your connection.',
          undefined,
          axiosError
        );
      }

      const status = axiosError.response.status;
      const data = axiosError.response.data as any;

      switch (status) {
        case 400:
          return new InventoryServiceError(
            data?.message || 'Invalid request parameters',
            status,
            axiosError
          );
        case 401:
          return new InventoryServiceError(
            'Unauthorized. Please log in again.',
            status,
            axiosError
          );
        case 403:
          return new InventoryServiceError(
            'Access denied. Insufficient permissions.',
            status,
            axiosError
          );
        case 404:
          return new InventoryServiceError(
            'Resource not found.',
            status,
            axiosError
          );
        case 500:
          return new InventoryServiceError(
            'Server error. Please try again later.',
            status,
            axiosError
          );
        default:
          return new InventoryServiceError(
            data?.message || defaultMessage,
            status,
            axiosError
          );
      }
    }

    return new InventoryServiceError(defaultMessage, undefined, error);
  },
};

/**
 * Inventory Hooks - React hooks for inventory operations
 */
export const useInventoryService = () => {
  return {
    listProducts: InventoryService.listProducts,
    searchProducts: InventoryService.searchProducts,
    getLowStockProducts: InventoryService.getLowStockProducts,
  };
};
