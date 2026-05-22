/**
 * Product type definitions for POS feature
 * Matches backend API response structure from Epic 2
 */

export interface Product {
  id: number;
  sku: string;
  name: string;
  description?: string;
  stockQty: number;
  price: string; // Decimal as string for precision
  costPrice?: string;
  expiryDate?: string;
  branchId: number;
  reorderThreshold: number;
  category?: string;
  createdAt: string;
  updatedAt: string;
  isExpired?: boolean; // Story 4.6, AC2: Expired status indicator from backend
}

export interface ProductListResponse {
  data: Product[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export interface ProductSearchParams {
  page?: number;
  limit?: number;
  sku?: string;
  search?: string;
}
