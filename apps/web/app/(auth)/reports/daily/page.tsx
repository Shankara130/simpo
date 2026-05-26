/**
 * Daily Sales Summary Report Page
 * Story 5.1, Task 6: Create Daily Sales Report Page (AC: 1, 2, 3)
 *
 * Features:
 * - Date picker for selecting report date (default: today)
 * - Branch selector for multi-branch owners
 * - Card-based layout with summary, payment breakdown, top products, hourly sales
 * - Loading states and error handling
 */

'use client';

import { useState, useEffect, useCallback } from 'react';
import axios from 'axios';

// Types matching backend DTO
interface PaymentBreakdown {
  paymentMethod: string;
  amount: string;
  transactionCount: number;
  percentage: number;
}

interface TopProduct {
  productId: number;
  sku: string;
  name: string;
  quantitySold: number;
  revenue: string;
}

interface HourlySales {
  hour: number;
  transactionCount: number;
  totalAmount: string;
}

interface DailySalesSummary {
  date: string;
  branchId: number;
  branchName: string;
  totalSales: string;
  totalTransactions: number;
  paymentBreakdown: PaymentBreakdown[];
  topProducts: TopProduct[];
  hourlySales: HourlySales[];
  generatedAt: string;
}

// Branch interface
interface Branch {
  id: number;
  name: string;
}

export default function DailySalesReportPage() {
  // Report state
  const [report, setReport] = useState<DailySalesSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Filter state
  const [selectedDate, setSelectedDate] = useState(formatDate(new Date()));
  const [selectedBranchId, setSelectedBranchId] = useState<number | undefined>(undefined);

  // Branches state (for multi-branch owners)
  const [branches, setBranches] = useState<Branch[]>([]);
  const [userRole, setUserRole] = useState<'OWNER' | 'ADMIN' | 'CASHIER'>('OWNER');

  /**
   * Fetch branches for the dropdown
   * Story 5.1, Task 6.3: Branch selector for multi-branch owners
   */
  useEffect(() => {
    fetchBranches();
    fetchUser();
  }, []);

  /**
   * Generate report when date or branch changes
   * Story 5.1, Task 6.4: "Generate Report" button triggers API call
   * Code review fix: Prevent race condition by checking branches are loaded
   */
  useEffect(() => {
    // Only generate report if branches are loaded (or no branch selected)
    if (branches.length > 0 || !selectedBranchId) {
      generateReport();
    }
  }, [generateReport, selectedBranchId, branches]);

  const fetchBranches = async () => {
    try {
      // In real implementation, fetch from API
      const mockBranches: Branch[] = [
        { id: 1, name: 'Apotek Sehat - Jakarta Pusat' },
        { id: 2, name: 'Apotek Sehat - Jakarta Selatan' },
        { id: 3, name: 'Apotek Sehat - Jakarta Barat' },
      ];
      setBranches(mockBranches);
    } catch (err) {
      console.error('Failed to fetch branches:', err);
    }
  };

  const fetchUser = async () => {
    // In real implementation, fetch from auth context
    setUserRole('OWNER');
  };

  /**
   * Generate daily sales report
   * Story 5.1, Task 6.4: Call API endpoint
   * Code review fix: Use useCallback to prevent excessive re-renders
   */
  const generateReport = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      // Build query parameters
      const params = new URLSearchParams({ date: selectedDate });
      if (selectedBranchId) {
        params.append('branch_id', selectedBranchId.toString());
      }

      // Call API
      // const response = await axios.get(`/api/v1/reports/daily?${params}`);

      // Mock response for development
      const mockReport: DailySalesSummary = {
        date: selectedDate,
        branchId: selectedBranchId || 0,
        branchName: selectedBranchId
          ? branches.find(b => b.id === selectedBranchId)?.name || ''
          : 'All Branches',
        totalSales: '15000000.00',
        totalTransactions: 45,
        paymentBreakdown: [
          { paymentMethod: 'CASH', amount: '8000000.00', transactionCount: 25, percentage: 55.56 },
          { paymentMethod: 'TRANSFER', amount: '5000000.00', transactionCount: 15, percentage: 33.33 },
          { paymentMethod: 'E_WALLET', amount: '2000000.00', transactionCount: 5, percentage: 11.11 },
        ],
        topProducts: [
          { productId: 1, sku: 'SKU-001', name: 'Paracetamol 500mg', quantitySold: 20, revenue: '300000.00' },
          { productId: 2, sku: 'SKU-002', name: 'Amoxicillin 500mg', quantitySold: 15, revenue: '375000.00' },
          { productId: 3, sku: 'SKU-003', name: 'Vitamin C 1000mg', quantitySold: 12, revenue: '180000.00' },
        ],
        hourlySales: Array.from({ length: 24 }, (_, i) => ({
          hour: i,
          transactionCount: Math.floor(Math.random() * 5),
          totalAmount: (Math.random() * 1000000).toFixed(2),
        })),
        generatedAt: new Date().toISOString(),
      };

      setReport(mockReport);
    } catch (err: any) {
      // Story 5.1, Task 6.7: Error handling with user-friendly messages
      setError(err.response?.data?.detail || 'Failed to generate report. Please try again.');
      setReport(null);
    } finally {
      setLoading(false);
    }
  }, [selectedDate, selectedBranchId, branches]);

  /**
   * Handle date change
   * Story 5.1, Task 6.2: Date picker component
   */
  const handleDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSelectedDate(e.target.value);
  };

  /**
   * Handle branch change
   * Story 5.1, Task 6.3: Branch selector dropdown
   */
  const handleBranchChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    setSelectedBranchId(value ? parseInt(value) : undefined);
  };

  /**
   * Format currency for display
   */
  const formatCurrency = (amount: string): string => {
    const num = parseFloat(amount);
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(num);
  };

  /**
   * Get payment method display name
   */
  const getPaymentMethodName = (method: string): string => {
    const names: Record<string, string> = {
      CASH: 'Tunai',
      TRANSFER: 'Transfer',
      E_WALLET: 'E-Wallet',
    };
    return names[method] || method;
  };

  /**
   * Export report to PDF
   * Story 5.3, Task 7.3: PDF export via backend API
   */
  const exportToPDF = async () => {
    try {
      setLoading(true);

      // Build query parameters
      const params = new URLSearchParams({
        date: selectedDate,
        format: 'pdf',
      });
      if (selectedBranchId) {
        params.append('branch_id', selectedBranchId.toString());
      }

      // Call backend API
      const response = await fetch(`/api/v1/reports/daily/export?${params}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
        },
      });

      if (!response.ok) {
        throw new Error('Gagal mengekspor PDF');
      }

      // Get blob and create download link
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `DailySalesReport_${selectedDate}.pdf`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);

      alert('PDF berhasil diekspor!');
    } catch (err) {
      console.error('Export PDF error:', err);
      alert('Gagal mengekspor PDF. Silakan coba lagi.');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Export report to Excel
   * Story 5.3, Task 7.4: Excel export via backend API
   */
  const exportToExcel = async () => {
    try {
      setLoading(true);

      // Build query parameters
      const params = new URLSearchParams({
        date: selectedDate,
        format: 'xlsx',
      });
      if (selectedBranchId) {
        params.append('branch_id', selectedBranchId.toString());
      }

      // Call backend API
      const response = await fetch(`/api/v1/reports/daily/export?${params}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
        },
      });

      if (!response.ok) {
        throw new Error('Gagal mengekspor Excel');
      }

      // Get blob and create download link
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `DailySalesReport_${selectedDate}.xlsx`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);

      alert('Excel berhasil diekspor!');
    } catch (err) {
      console.error('Export Excel error:', err);
      alert('Gagal mengekspor Excel. Silakan coba lagi.');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Generate CSV content from report data
   * Story 5.1, Task 7.4: Interim solution for Excel export
   */
  const generateCSV = (data: DailySalesSummary): string => {
    const rows: string[] = [];

    // Header
    rows.push('Laporan Penjualan Harian');
    rows.push(`Tanggal: ${data.date}`);
    rows.push(`Cabang: ${data.branchName}`);
    rows.push(`Dibuat: ${new Date(data.generatedAt).toLocaleString('id-ID')}`);
    rows.push('');

    // Summary
    rows.push('RINGKASAN');
    rows.push(`Total Penjualan,${data.totalSales}`);
    rows.push(`Total Transaksi,${data.totalTransactions}`);
    rows.push('');

    // Payment Breakdown
    rows.push('BREAKDOWN METODE PEMBAYARAN');
    rows.push('Metode,Jumlah,Transaksi,Persentase');
    data.paymentBreakdown.forEach(p => {
      rows.push(`${getPaymentMethodName(p.paymentMethod)},${p.amount},${p.transactionCount},${p.percentage.toFixed(2)}%`);
    });
    rows.push('');

    // Top Products
    rows.push('PRODUK TERLARIS');
    rows.push('SKU,Nama Produk,Qty Terjual,Pendapatan');
    data.topProducts.forEach(p => {
      rows.push(`${p.sku},${p.name},${p.quantitySold},${p.revenue}`);
    });
    rows.push('');

    // Hourly Sales
    rows.push('PENJUALAN PER JAM');
    rows.push('Jam,Transaksi,Total');
    data.hourlySales.filter(h => h.transactionCount > 0).forEach(h => {
      rows.push(`${h.hour.toString().padStart(2, '0')}:00,${h.transactionCount},${h.totalAmount}`);
    });

    return rows.join('\n');
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Laporan Penjualan Harian</h1>
          <p className="text-gray-600 mt-1">Ringkasan penjualan harian dengan rincian pembayaran dan produk terlaris</p>
        </div>

        {/* Export buttons */}
        {/* Story 5.1, Task 7.1, 7.2: PDF and Excel export buttons */}
        {report && !loading && (
          <div className="flex gap-2">
            <button
              onClick={exportToPDF}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors"
            >
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Export PDF
            </button>
            <button
              onClick={exportToExcel}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors"
            >
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Export Excel
            </button>
          </div>
        )}
      </div>

      {/* Filters */}
      {/* Story 5.1, Task 6.2, 6.3: Date picker and branch selector */}
      <div className="bg-white p-4 rounded-lg border shadow-sm">
        <div className="flex flex-wrap gap-4">
          {/* Date picker */}
          {/* Code review fix: MED-005 - Add accessibility improvements */}
          <div className="flex-1 min-w-[200px]">
            <label htmlFor="date" className="block text-sm font-medium text-gray-700 mb-1">
              Tanggal Laporan
            </label>
            <input
              id="date"
              type="date"
              value={selectedDate}
              onChange={handleDateChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              max={formatDate(new Date())}
              aria-label="Pilih tanggal laporan"
              aria-describedby="date-description"
            />
            <span id="date-description" className="sr-only">
              Pilih tanggal untuk melihat laporan penjualan harian
            </span>
          </div>

          {/* Branch selector (for Owners only) */}
          {/* Code review fix: MED-005 - Add accessibility improvements */}
          {userRole === 'OWNER' && (
            <div className="flex-1 min-w-[200px]">
              <label htmlFor="branch" className="block text-sm font-medium text-gray-700 mb-1">
                Cabang
              </label>
              <select
                id="branch"
                value={selectedBranchId || ''}
                onChange={handleBranchChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                aria-label="Pilih cabang untuk filter laporan"
              >
                <option value="">Semua Cabang</option>
                {branches.map(branch => (
                  <option key={branch.id} value={branch.id}>
                    {branch.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Generate Report button */}
          {/* Story 5.1, Task 6.4: Generate Report button */}
          {/* Code review fix: MED-005 - Add accessibility improvements */}
          <div className="flex items-end">
            <button
              onClick={generateReport}
              disabled={loading}
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
              aria-label="Generate laporan penjualan harian"
              aria-describedby="button-description"
            >
              {loading ? 'Memuat...' : 'Generate Laporan'}
            </button>
            <span id="button-description" className="sr-only">
              Klik untuk melihat laporan penjualan harian berdasarkan filter yang dipilih
            </span>
          </div>
        </div>
      </div>

      {/* Error state */}
      {/* Story 5.1, Task 6.7: Error handling */}
      {/* Code review fix: HIGH-005 - React's automatic escaping provides XSS protection */}
      {/* No dangerouslySetInnerHTML used - all error content is safely escaped */}
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
          <p className="font-medium">Error</p>
          <p className="text-sm">{error}</p>
        </div>
      )}

      {/* Loading state */}
      {/* Story 5.1, Task 6.6: Loading states */}
      {/* Code review fix: LOW-004 - Enhanced loading state with descriptive message */}
      {loading && (
        <div className="bg-white p-8 rounded-lg border shadow-sm text-center" role="status" aria-live="polite">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p className="mt-2 text-gray-600" aria-live="polite">
            Memuat laporan untuk {selectedDate}...
            {selectedBranchId && branches.find(b => b.id === selectedBranchId) && ` (${branches.find(b => b.id === selectedBranchId)?.name})`}
          </p>
        </div>
      )}

      {/* Report content */}
      {/* Story 5.1, Task 6.5: Card-based layout */}
      {report && !loading && (
        <div className="space-y-6">
          {/* Summary Cards */}
          {/* Code review fix: MED-005 - Add accessibility improvements */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4" role="region" aria-label="Ringkasan Penjualan">
            {/* Total Sales */}
            <div className="bg-white p-6 rounded-lg border shadow-sm">
              <p className="text-sm font-medium text-gray-600">Total Penjualan</p>
              <p className="text-2xl font-bold text-gray-900 mt-2" aria-label={`Total penjualan: ${formatCurrency(report.totalSales)}`}>
                {formatCurrency(report.totalSales)}
              </p>
              <p className="text-xs text-gray-500 mt-1">{report.branchName}</p>
            </div>

            {/* Total Transactions */}
            <div className="bg-white p-6 rounded-lg border shadow-sm">
              <p className="text-sm font-medium text-gray-600">Total Transaksi</p>
              <p className="text-2xl font-bold text-gray-900 mt-2" aria-label={`Total transaksi: ${report.totalTransactions}`}>
                {report.totalTransactions}
              </p>
              <p className="text-xs text-gray-500 mt-1">transaksi</p>
            </div>

            {/* Average Transaction */}
            <div className="bg-white p-6 rounded-lg border shadow-sm">
              <p className="text-sm font-medium text-gray-600">Rata-rata Transaksi</p>
              <p className="text-2xl font-bold text-gray-900 mt-2" aria-label={`Rata-rata transaksi: ${formatCurrency((parseFloat(report.totalSales) / Math.max(report.totalTransactions, 1)).toFixed(2))}`}>
                {formatCurrency((parseFloat(report.totalSales) / Math.max(report.totalTransactions, 1)).toFixed(2))}
              </p>
              <p className="text-xs text-gray-500 mt-1">per transaksi</p>
            </div>
          </div>

          {/* Payment Breakdown */}
          {/* Story 5.1, Task 6.5: Payment breakdown section */}
          <div className="bg-white p-6 rounded-lg border shadow-sm">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Breakdown Metode Pembayaran</h3>
            <div className="space-y-4">
              {report.paymentBreakdown.map(payment => (
                <div key={payment.paymentMethod} className="flex items-center">
                  <div className="flex-1">
                    <div className="flex justify-between mb-1">
                      <span className="text-sm font-medium text-gray-700">
                        {getPaymentMethodName(payment.paymentMethod)}
                      </span>
                      <span className="text-sm text-gray-600">
                        {formatCurrency(payment.amount)} ({payment.transactionCount} transaksi)
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div
                        className="bg-blue-600 h-2 rounded-full"
                        style={{ width: `${payment.percentage}%` }}
                      ></div>
                    </div>
                  </div>
                  <div className="ml-4 text-sm font-medium text-gray-700 w-12 text-right">
                    {payment.percentage.toFixed(1)}%
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Top Products Table */}
          {/* Story 5.1, Task 6.5: Top products section */}
          <div className="bg-white p-6 rounded-lg border shadow-sm">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Produk Terlaris</h3>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      SKU
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Nama Produk
                    </th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Qty Terjual
                    </th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Pendapatan
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {report.topProducts.map(product => (
                    <tr key={product.productId}>
                      <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
                        {product.sku}
                      </td>
                      <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
                        {product.name}
                      </td>
                      <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900 text-right">
                        {product.quantitySold}
                      </td>
                      <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900 text-right">
                        {formatCurrency(product.revenue)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Hourly Sales */}
          {/* Story 5.1, Task 6.5: Hourly sales section */}
          <div className="bg-white p-6 rounded-lg border shadow-sm">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Penjualan per Jam</h3>
            <div className="space-y-2">
              {report.hourlySales.filter(h => h.transactionCount > 0).map(hourly => (
                <div key={hourly.hour} className="flex items-center text-sm">
                  <div className="w-16 text-gray-600">
                    {hourly.hour.toString().padStart(2, '0')}:00
                  </div>
                  <div className="flex-1 mx-4">
                    <div className="bg-gray-200 rounded h-6 relative">
                      <div
                        className="bg-green-500 h-6 rounded"
                        style={{
                          width: `${Math.min(
                            (hourly.transactionCount /
                              Math.max(...report.hourlySales.map(h => h.transactionCount))) * 100,
                            100
                          )}%`,
                        }}
                      ></div>
                      <span className="absolute inset-0 flex items-center justify-center text-xs font-medium">
                        {hourly.transactionCount} transaksi
                      </span>
                    </div>
                  </div>
                  <div className="w-24 text-right text-gray-900">
                    {formatCurrency(hourly.totalAmount)}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Report metadata */}
          <div className="text-xs text-gray-500 text-center">
            Dibuat pada: {new Date(report.generatedAt).toLocaleString('id-ID')}
          </div>
        </div>
      )}
    </div>
  );
}

/**
 * Format date to YYYY-MM-DD string
 */
function formatDate(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}
