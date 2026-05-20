'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

export default function AuthenticatedLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const pathname = usePathname();
  const [lowStockCount, setLowStockCount] = useState(0);

  // Fetch low stock count periodically
  useEffect(() => {
    const fetchLowStockCount = async () => {
      try {
        const response = await fetch('/api/v1/products/low-stock', {
          credentials: 'include',
        });

        if (response.ok) {
          const data = await response.json();
          setLowStockCount(data.data?.length || 0);
        }
      } catch (error) {
        console.error('Failed to fetch low stock count:', error);
      }
    };

    // Fetch immediately
    fetchLowStockCount();

    // Fetch every 30 seconds
    const interval = setInterval(fetchLowStockCount, 30000);

    return () => clearInterval(interval);
  }, []);

  const navItems = [
    { href: '/dashboard', label: 'Dashboard' },
    { href: '/products', label: 'Products' },
    { href: '/inventory/low-stock', label: 'Low Stock', showBadge: true, badgeCount: lowStockCount },
    { href: '/reports', label: 'Reports' },
    { href: '/users', label: 'Users' },
    { href: '/settings', label: 'Settings' },
  ];

  const isActive = (href: string) => {
    if (href === '/dashboard') {
      return pathname === '/' || pathname === '/dashboard';
    }
    return pathname.startsWith(href);
  };

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header will be added here */}
      <header className="border-b bg-white">
        <div className="container mx-auto px-4 py-4">
          <h1 className="text-xl font-bold">simpo Admin Dashboard</h1>
        </div>
      </header>

      <div className="flex flex-1">
        {/* Sidebar will be added here */}
        <aside className="w-64 border-r bg-gray-50 p-4">
          <nav className="space-y-2">
            {navItems.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className={`relative block py-2 px-4 rounded hover:bg-gray-200 transition-colors ${
                  isActive(item.href) ? 'bg-blue-100 text-blue-700 font-medium' : ''
                }`}
              >
                {item.label}
                {item.showBadge && item.badgeCount > 0 && (
                  <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
                    {item.badgeCount}
                  </span>
                )}
                {item.showBadge && item.badgeCount === 0 && (
                  <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600">
                    0
                  </span>
                )}
              </Link>
            ))}
          </nav>
        </aside>

        {/* Main content */}
        <main className="flex-1 p-6">
          {children}
        </main>
      </div>

      {/* Footer will be added here */}
      <footer className="border-t bg-white py-4">
        <div className="container mx-auto px-4 text-center text-sm text-gray-600">
          © 2026 simpo. Pharmacy Management System.
        </div>
      </footer>
    </div>
  );
}
