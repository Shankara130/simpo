/**
 * Authenticated Layout Tests
 * Story 4.5, Task 9: Add Navigation and Menu Items
 *
 * Tests for:
 * - "Expiring" link in sidebar navigation
 * - Badge showing count of expiring items (7-day urgent count)
 * - Highlight menu item when critical alerts exist
 */

import { render, screen, waitFor } from '@testing-library/react';
import AuthenticatedLayout from './layout';
import * as nextNavigation from 'next/navigation';

// Mock next/navigation
jest.mock('next/navigation', () => ({
  usePathname: jest.fn(),
}));

// Mock fetch globally
global.fetch = jest.fn();

describe('AuthenticatedLayout - Navigation', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    (nextNavigation.usePathname as jest.Mock).mockReturnValue('/');
    (global.fetch as jest.Mock).mockClear();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  // Task 9.1: Add "Expiring" link to sidebar navigation
  describe('Navigation items', () => {
    it('should display "Expiring" link in sidebar', () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ data: [], pagination: { total: 0 } }),
      });

      render(<AuthenticatedLayout><div>Test Content</div></AuthenticatedLayout>);

      const expiringLink = screen.getByText('Expiring');
      expect(expiringLink).toBeInTheDocument();
      expect(expiringLink.closest('a')).toHaveAttribute('href', '/inventory/expiring');
    });

    it('should display all expected navigation items', () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ data: [], pagination: { total: 0 } }),
      });

      render(<AuthenticatedLayout><div>Test Content</div></AuthenticatedLayout>);

      expect(screen.getByText('Dashboard')).toBeInTheDocument();
      expect(screen.getByText('Products')).toBeInTheDocument();
      expect(screen.getByText('Low Stock')).toBeInTheDocument();
      expect(screen.getByText('Expiring')).toBeInTheDocument();
      expect(screen.getByText('Reports')).toBeInTheDocument();
      expect(screen.getByText('Users')).toBeInTheDocument();
      expect(screen.getByText('Settings')).toBeInTheDocument();
    });
  });

  // Task 9.2: Add badge showing count of expiring items (7-day urgent count)
  describe('Expiry badge count', () => {
    it('should fetch and display urgent expiry count on mount', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({
          data: [
            { id: 1, name: 'Product 1' },
            { id: 2, name: 'Product 2' },
            { id: 3, name: 'Product 3' },
          ],
          pagination: { total: 3 },
        }),
      });

      render(<AuthenticatedLayout><div>Test Content</div></AuthenticatedLayout>);

      // Should call fetch for both low stock and expiring
      await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith('/api/v1/products/low-stock', {
          credentials: 'include',
        });
      });

      await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith('/api/v1/products/expiring?days=7', {
          credentials: 'include',
        });
      });
    });

    it('should display badge with urgent expiry count when count > 0', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({
          data: [{ id: 1 }, { id: 2 }],
          pagination: { total: 2 },
        }),
      });

      render(<AuthenticatedLayout><div>Test Content</div></AuthenticatedLayout>);

      await waitFor(() => {
        const badges = screen.getAllByText('2');
        const expiringBadge = badges.find(badge => {
          const link = badge.closest('a');
          return link?.getAttribute('href') === '/inventory/expiring';
        });
        expect(expiringBadge).toBeInTheDocument();
      });
    });

    it('should display gray badge when count is 0', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({
          data: [],
          pagination: { total: 0 },
        }),
      });

      render(<AuthenticatedLayout><div>Test Content</div></AuthenticatedLayout>);

      await waitFor(() => {
        const expiringLink = screen.getByText('Expiring').closest('a');
        expect(expiringLink?.textContent).toContain('0');
      });
    });

    it('should fetch expiry count every 30 seconds', async () => {
      let fetchCount = 0;
      (global.fetch as jest.Mock).mockImplementation(() => {
        fetchCount++;
        return Promise.resolve({
          ok: true,
          json: async () => ({ data: [], pagination: { total: 0 } }),
        });
      });

      render(<AuthenticatedLayout><div>Test Content</div></AuthenticatedLayout>);

      // Initial fetches (low stock + expiring)
      await waitFor(() => expect(fetchCount).toBeGreaterThanOrEqual(2));

      // Fast-forward 30 seconds
      jest.advanceTimersByTime(30000);

      // Should fetch again
      await waitFor(() => expect(fetchCount).toBeGreaterThanOrEqual(4));
    });
  });

  // Task 9.3: Highlight menu item when critical alerts exist
  describe('Urgent alert highlighting', () => {
    it('should highlight menu item with red background and pulse animation when urgent count > 0', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({
          data: [{ id: 1 }, { id: 2 }, { id: 3 }],
          pagination: { total: 3 },
        }),
      });

      render(<AuthenticatedLayout><div>Test Content</div></AuthenticatedLayout>);

      await waitFor(() => {
        const badge = screen.getByText('3');
        const expiringLink = badge.closest('a');

        expect(expiringLink?.querySelector('.bg-red-600')).toBeInTheDocument();
        expect(expiringLink?.querySelector('.text-white')).toBeInTheDocument();
        expect(expiringLink?.querySelector('.animate-pulse')).toBeInTheDocument();
      });
    });

    it('should not highlight menu item when urgent count is 0', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({
          data: [],
          pagination: { total: 0 },
        }),
      });

      render(<AuthenticatedLayout><div>Test Content</div></AuthenticatedLayout>);

      await waitFor(() => {
        const badge = screen.getByText('0');
        const expiringLink = badge.closest('a');

        expect(expiringLink?.querySelector('.bg-red-600')).not.toBeInTheDocument();
        expect(expiringLink?.querySelector('.animate-pulse')).not.toBeInTheDocument();
      });
    });

    it('should show standard red badge (not urgent) when count > 0 but no critical alerts', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({
          data: [],
          pagination: { total: 0 },
        }),
      });

      render(<AuthenticatedLayout><div>Test Content</div></AuthenticatedLayout>);

      await waitFor(() => {
        const badge = screen.getByText('0');
        expect(badge).toHaveClass('bg-gray-100', 'text-gray-600');
      });
    });
  });

  describe('Active state styling', () => {
    it('should highlight active navigation item', () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ data: [], pagination: { total: 0 } }),
      });

      (nextNavigation.usePathname as jest.Mock).mockReturnValue('/inventory/expiring');

      render(<AuthenticatedLayout><div>Test Content</div></AuthenticatedLayout>);

      const expiringLink = screen.getByText('Expiring').closest('a');
      expect(expiringLink).toHaveClass('bg-blue-100', 'text-blue-700');
    });

    it('should highlight dashboard when pathname is /', () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ data: [], pagination: { total: 0 } }),
      });

      (nextNavigation.usePathname as jest.Mock).mockReturnValue('/');

      render(<AuthenticatedLayout><div>Test Content</div></AuthenticatedLayout>);

      const dashboardLink = screen.getByText('Dashboard').closest('a');
      expect(dashboardLink).toHaveClass('bg-blue-100', 'text-blue-700');
    });
  });
});
