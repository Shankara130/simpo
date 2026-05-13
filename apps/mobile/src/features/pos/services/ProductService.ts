/**
 * ProductService - API integration for Product data
 * Handles product search, listing, and API communication
 */

import axios, { AxiosError } from 'axios';
import { Product, ProductListResponse, ProductSearchParams } from '../types/product.types';

// API base URL - configure for different environments
const API_BASE_URL = __DEV__
  ? 'http://localhost:8080' // Development - local backend
  : 'https://api.simpo.id';  // Production - actual API endpoint

// API version prefix
const API_VERSION = '/api/v1';

// Full API base URL
const FULL_API_URL = `${API_BASE_URL}${API_VERSION}`;

/**
 * API Error class for handling service-specific errors
 */
export class ProductServiceError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public originalError?: any
  ) {
    super(message);
    this.name = 'ProductServiceError';
  }
}

/**
 * ProductService - Product API operations
 */
export const ProductService = {
  /**
   * Fetch list of products with pagination and optional filters
   */
  getProducts: async (params: ProductSearchParams = {}): Promise<ProductListResponse> => {
    try {
      let { page = 1, limit = 20, sku, search } = params;

      // Validate and sanitize pagination parameters
      if (page < 1 || !Number.isFinite(page)) {
        page = 1;
      }

      if (limit < 1 || !Number.isFinite(limit)) {
        limit = 20;
      }

      // Enforce reasonable maximum limit
      if (limit > 100) {
        limit = 100;
      }

      const response = await axios.get<ProductListResponse>(`${FULL_API_URL}/products`, {
        params: {
          page,
          limit,
          sku,
          search,
        },
        timeout: 10000, // 10 second timeout
      });

      return response.data;
    } catch (error) {
      throw ProductService._handleError(error, 'Failed to fetch products');
    }
  },

  /**
   * Get a single product by SKU (barcode lookup)
   */
  getProductBySKU: async (sku: string): Promise<Product> => {
    try {
      if (!sku || sku.trim().length === 0) {
        throw new ProductServiceError('SKU cannot be empty', 400);
      }

      const response = await axios.get<{ data: Product[] }>(`${FULL_API_URL}/products`, {
        params: { sku: sku.trim() },
        timeout: 10000,
      });

      // Validate response structure
      if (!response.data || !response.data.data) {
        throw new ProductServiceError('Invalid API response structure', 500);
      }

      const products = response.data.data;

      // Validate it's an array
      if (!Array.isArray(products)) {
        throw new ProductServiceError('API response data is not an array', 500);
      }

      if (products.length === 0) {
        throw new ProductServiceError(`Product with SKU ${sku} not found`, 404);
      }

      return products[0];
    } catch (error) {
      if (error instanceof ProductServiceError) {
        throw error;
      }
      throw ProductService._handleError(error, `Failed to fetch product with SKU: ${sku}`);
    }
  },

  /**
   * Get a single product by barcode (alias for getProductBySKU)
   * This method provides semantic clarity for barcode scanning use case
   */
  getProductByBarcode: async (barcode: string): Promise<Product> => {
    // Barcode is stored as SKU in the system, so we just alias to getProductBySKU
    return ProductService.getProductBySKU(barcode);
  },

  /**
   * Search products by name or SKU
   */
  searchProducts: async (query: string, page = 1, limit = 20): Promise<ProductListResponse> => {
    try {
      // Validate query parameter
      if (!query || typeof query !== 'string') {
        return ProductService.getProducts({ page, limit });
      }

      const trimmedQuery = query.trim();

      if (trimmedQuery.length === 0) {
        return ProductService.getProducts({ page, limit });
      }

      return await ProductService.getProducts({
        search: trimmedQuery,
        page,
        limit,
      });
    } catch (error) {
      throw ProductService._handleError(error, 'Failed to search products');
    }
  },

  /**
   * Handle API errors and convert to ProductServiceError
   * @private
   */
  _handleError: (error: unknown, defaultMessage: string) => ProductServiceError => {
    if (error instanceof ProductServiceError) {
      return error;
    }

    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;

      // Network error (no response)
      if (!axiosError.response) {
        return new ProductServiceError(
          'Network error. Please check your connection.',
          undefined,
          axiosError
        );
      }

      // Server responded with error status
      const status = axiosError.response.status;
      const data = axiosError.response.data as any;

      switch (status) {
        case 400:
          return new ProductServiceError(
            data?.message || 'Invalid request parameters',
            status,
            axiosError
          );
        case 401:
          return new ProductServiceError(
            'Unauthorized. Please log in again.',
            status,
            axiosError
          );
        case 403:
          return new ProductServiceError(
            'Access denied. You do not have permission to access this resource.',
            status,
            axiosError
          );
        case 404:
          return new ProductServiceError(
            'Resource not found.',
            status,
            axiosError
          );
        case 500:
          return new ProductServiceError(
            'Server error. Please try again later.',
            status,
            axiosError
          );
        default:
          return new ProductServiceError(
            data?.message || defaultMessage,
            status,
            axiosError
          );
      }
    }

    // Unknown error
    return new ProductServiceError(defaultMessage, undefined, error);
  },
};

/**
 * Product Hooks - React hooks for product operations
 * These can be used in components to fetch product data
 */
export const useProductService = () => {
  return {
    getProducts: ProductService.getProducts,
    getProductBySKU: ProductService.getProductBySKU,
    searchProducts: ProductService.searchProducts,
  };
};
