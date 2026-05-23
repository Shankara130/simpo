'use client';

import { useState, useEffect } from 'react';
import { format } from 'date-fns';

// Types for profit/loss report
interface ProfitLossSummary {
  periodStart: string;
  periodEnd: string;
  branchId: number;
  branchName: string;
  revenue: string;
  costOfGoodsSold: string;
  grossProfit: string;
  grossProfitMargin: number;
  breakdownBy: string;
  breakdowns?: ProfitLossBreakdown[];
  branchBreakdowns?: BranchBreakdown[];
  paymentBreakdowns?: PaymentMethodBreakdown[];
  generatedAt: string;
}

interface ProfitLossBreakdown {
  category: string;
  revenue: string;
  costOfGoodsSold: string;
  grossProfit: string;
  marginPercentage: number;
}

interface BranchBreakdown {
  branchId: number;
  branchName: string;
  revenue: string;
  costOfGoodsSold: string;
  grossProfit: string;
  marginPercentage: number;
}

interface PaymentMethodBreakdown {
  paymentMethod: string;
  revenue: string;
  costOfGoodsSold: string;
  grossProfit: string;
  marginPercentage: number;
}

interface Branch {
  id: number;
  name: string;
}

export default function ProfitLossPage() {
  // Story 5.2, Task 6.2, 6.3, 6.4: State for date range, breakdown, and branch filters
  const [startDate, setStartDate] = useState(format(new Date(), 'yyyy-MM-dd'));
  const [endDate, setEndDate] = useState(format(new Date(), 'yyyy-MM-dd'));
  const [breakdownBy, setBreakdownBy] = useState('');
  const [selectedBranch, setSelectedBranch] = useState<number | undefined>(undefined);
  const [branches, setBranches] = useState<Branch[]>([]);

  // Story 5.2, Task 6.6, 6.7: Report data and loading states
  const [report, setReport] = useState<ProfitLossSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch branches on component mount
  useEffect(() => {
    fetchBranches();
  }, []);

  const fetchBranches = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/v1/branches', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      if (response.ok) {
        const data = await response.json();
        setBranches(data);
      }
    } catch (err) {
      console.error('Failed to fetch branches:', err);
    }
  };

  // Story 5.2, Task 7.4: Export to CSV function (interim solution for Excel export)
  const exportToCSV = () => {
    if (!report) return;

    // Create CSV content
    let csv = 'Profit/Loss Report\n';
    csv += `Period: ${report.periodStart} to ${report.periodEnd}\n`;
    csv += `Branch: ${report.branchName}\n\n`;

    // Summary
    csv += 'Summary\n';
    csv += 'Metric,Amount\n';
    csv += `Revenue,${report.revenue}\n`;
    csv += `Cost of Goods Sold,${report.costOfGoodsSold}\n`;
    csv += `Gross Profit,${report.grossProfit}\n`;
    csv += `Gross Margin,${report.grossProfitMargin.toFixed(2)}%\n\n`;

    // Breakdown data based on selected breakdown type
    if (breakdownBy === 'category' && report.breakdowns) {
      csv += 'Breakdown by Category\n';
      csv += 'Category,Revenue,Cost of Goods Sold,Gross Profit,Margin %\n';
      report.breakdowns.forEach((item) => {
        csv += `${item.category},${item.revenue},${item.costOfGoodsSold},${item.grossProfit},${item.marginPercentage.toFixed(2)}\n`;
      });
    } else if (breakdownBy === 'branch' && report.branchBreakdowns) {
      csv += 'Breakdown by Branch\n';
      csv += 'Branch,Revenue,Cost of Goods Sold,Gross Profit,Margin %\n';
      report.branchBreakdowns.forEach((item) => {
        csv += `${item.branchName},${item.revenue},${item.costOfGoodsSold},${item.grossProfit},${item.marginPercentage.toFixed(2)}\n`;
      });
    } else if (breakdownBy === 'payment_method' && report.paymentBreakdowns) {
      csv += 'Breakdown by Payment Method\n';
      csv += 'Payment Method,Revenue,Cost of Goods Sold,Gross Profit,Margin %\n';
      report.paymentBreakdowns.forEach((item) => {
        csv += `${item.paymentMethod},${item.revenue},${item.costOfGoodsSold},${item.grossProfit},${item.marginPercentage.toFixed(2)}\n`;
      });
    }

    // Story 5.2, Task 7.5: Include report metadata in export
    csv += `\nGenerated: ${new Date(report.generatedAt).toLocaleString()}\n`;

    // Create download link
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `profit-loss-report-${report.periodStart}-to-${report.periodEnd}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  };

  // Story 5.2, Task 6.5: Generate report function
  const generateReport = async () => {
    setLoading(true);
    setError(null);

    try {
      const token = localStorage.getItem('token');
      const params = new URLSearchParams({
        start_date: startDate,
        end_date: endDate,
      });

      if (breakdownBy) {
        params.append('breakdown_by', breakdownBy);
      }

      if (selectedBranch !== undefined) {
        params.append('branch_id', selectedBranch.toString());
      }

      const response = await fetch(`/api/v1/reports/profit-loss?${params}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.detail || 'Failed to generate report');
      }

      const data = await response.json();
      setReport(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  // Story 5.2, Task 6.6: Display report sections in card-based layout
  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">Profit/Loss Report</h1>

      {/* Story 5.2, Task 6.2, 6.3, 6.4: Date range picker, breakdown selector, branch selector */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
            <input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="w-full border rounded px-3 py-2"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">End Date</label>
            <input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className="w-full border rounded px-3 py-2"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Breakdown By</label>
            <select
              value={breakdownBy}
              onChange={(e) => setBreakdownBy(e.target.value)}
              className="w-full border rounded px-3 py-2"
            >
              <option value="">None</option>
              <option value="category">Product Category</option>
              <option value="branch">Branch Location</option>
              <option value="payment_method">Payment Method</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Branch</label>
            <select
              value={selectedBranch ?? ''}
              onChange={(e) => setSelectedBranch(e.target.value ? Number(e.target.value) : undefined)}
              className="w-full border rounded px-3 py-2"
            >
              <option value="">All Branches</option>
              {branches.map((branch) => (
                <option key={branch.id} value={branch.id}>
                  {branch.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Story 5.2, Task 6.5: Generate Report button */}
        <div className="mt-4">
          <button
            onClick={generateReport}
            disabled={loading}
            className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:bg-gray-400"
          >
            {loading ? 'Generating...' : 'Generate Report'}
          </button>
        </div>
      </div>

      {/* Story 5.2, Task 7.1, 7.2: Export buttons */}
      {report && !loading && (
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h3 className="text-lg font-semibold mb-3">Export Report</h3>
          <div className="flex gap-2">
            {/* Story 5.2, Task 7.3: Export PDF (placeholder with print dialog) */}
            <button
              onClick={() => window.print()}
              className="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700"
            >
              Export PDF
            </button>
            {/* Story 5.2, Task 7.4: Export Excel (CSV download as interim solution) */}
            <button
              onClick={exportToCSV}
              className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
            >
              Export Excel
            </button>
          </div>
        </div>
      )}

      {/* Story 5.2, Task 6.7: Loading state */}
      {loading && (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <p className="text-center text-gray-600">Generating report...</p>
        </div>
      )}

      {/* Story 5.2, Task 6.8: Error handling */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-red-600">{error}</p>
        </div>
      )}

      {/* Story 5.2, Task 6.6: Display report sections in card-based layout */}
      {report && !loading && (
        <div>
          {/* Summary card: Revenue, COGS, Gross Profit, Gross Margin */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-white rounded-lg shadow p-4">
              <h3 className="text-sm font-medium text-gray-500 mb-2">Revenue</h3>
              <p className="text-2xl font-bold text-gray-900">Rp {parseFloat(report.revenue).toLocaleString('id-ID')}</p>
            </div>

            <div className="bg-white rounded-lg shadow p-4">
              <h3 className="text-sm font-medium text-gray-500 mb-2">Cost of Goods Sold</h3>
              <p className="text-2xl font-bold text-gray-900">Rp {parseFloat(report.costOfGoodsSold).toLocaleString('id-ID')}</p>
            </div>

            <div className="bg-white rounded-lg shadow p-4">
              <h3 className="text-sm font-medium text-gray-500 mb-2">Gross Profit</h3>
              <p className="text-2xl font-bold text-green-600">Rp {parseFloat(report.grossProfit).toLocaleString('id-ID')}</p>
            </div>

            <div className="bg-white rounded-lg shadow p-4">
              <h3 className="text-sm font-medium text-gray-500 mb-2">Gross Margin</h3>
              <p className="text-2xl font-bold text-blue-600">{report.grossProfitMargin.toFixed(2)}%</p>
            </div>
          </div>

          {/* Breakdown cards: Based on selected breakdown_by parameter */}
          {breakdownBy === 'category' && report.breakdowns && report.breakdowns.length > 0 && (
            <div className="bg-white rounded-lg shadow p-6 mb-6">
              <h2 className="text-lg font-semibold mb-4">Breakdown by Category</h2>
              <div className="overflow-x-auto">
                <table className="min-w-full">
                  <thead>
                    <tr className="border-b">
                      <th className="text-left py-2 px-3">Category</th>
                      <th className="text-right py-2 px-3">Revenue</th>
                      <th className="text-right py-2 px-3">COGS</th>
                      <th className="text-right py-2 px-3">Gross Profit</th>
                      <th className="text-right py-2 px-3">Margin %</th>
                    </tr>
                  </thead>
                  <tbody>
                    {report.breakdowns.map((breakdown, index) => (
                      <tr key={index} className="border-b">
                        <td className="py-2 px-3">{breakdown.category}</td>
                        <td className="text-right py-2 px-3">Rp {parseFloat(breakdown.revenue).toLocaleString('id-ID')}</td>
                        <td className="text-right py-2 px-3">Rp {parseFloat(breakdown.costOfGoodsSold).toLocaleString('id-ID')}</td>
                        <td className="text-right py-2 px-3">Rp {parseFloat(breakdown.grossProfit).toLocaleString('id-ID')}</td>
                        <td className="text-right py-2 px-3">{breakdown.marginPercentage.toFixed(2)}%</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {breakdownBy === 'branch' && report.branchBreakdowns && report.branchBreakdowns.length > 0 && (
            <div className="bg-white rounded-lg shadow p-6 mb-6">
              <h2 className="text-lg font-semibold mb-4">Breakdown by Branch</h2>
              <div className="overflow-x-auto">
                <table className="min-w-full">
                  <thead>
                    <tr className="border-b">
                      <th className="text-left py-2 px-3">Branch</th>
                      <th className="text-right py-2 px-3">Revenue</th>
                      <th className="text-right py-2 px-3">COGS</th>
                      <th className="text-right py-2 px-3">Gross Profit</th>
                      <th className="text-right py-2 px-3">Margin %</th>
                    </tr>
                  </thead>
                  <tbody>
                    {report.branchBreakdowns.map((breakdown, index) => (
                      <tr key={index} className="border-b">
                        <td className="py-2 px-3">{breakdown.branchName}</td>
                        <td className="text-right py-2 px-3">Rp {parseFloat(breakdown.revenue).toLocaleString('id-ID')}</td>
                        <td className="text-right py-2 px-3">Rp {parseFloat(breakdown.costOfGoodsSold).toLocaleString('id-ID')}</td>
                        <td className="text-right py-2 px-3">Rp {parseFloat(breakdown.grossProfit).toLocaleString('id-ID')}</td>
                        <td className="text-right py-2 px-3">{breakdown.marginPercentage.toFixed(2)}%</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {breakdownBy === 'payment_method' && report.paymentBreakdowns && report.paymentBreakdowns.length > 0 && (
            <div className="bg-white rounded-lg shadow p-6 mb-6">
              <h2 className="text-lg font-semibold mb-4">Breakdown by Payment Method</h2>
              <div className="overflow-x-auto">
                <table className="min-w-full">
                  <thead>
                    <tr className="border-b">
                      <th className="text-left py-2 px-3">Payment Method</th>
                      <th className="text-right py-2 px-3">Revenue</th>
                      <th className="text-right py-2 px-3">COGS</th>
                      <th className="text-right py-2 px-3">Gross Profit</th>
                      <th className="text-right py-2 px-3">Margin %</th>
                    </tr>
                  </thead>
                  <tbody>
                    {report.paymentBreakdowns.map((breakdown, index) => (
                      <tr key={index} className="border-b">
                        <td className="py-2 px-3">{breakdown.paymentMethod}</td>
                        <td className="text-right py-2 px-3">Rp {parseFloat(breakdown.revenue).toLocaleString('id-ID')}</td>
                        <td className="text-right py-2 px-3">Rp {parseFloat(breakdown.costOfGoodsSold).toLocaleString('id-ID')}</td>
                        <td className="text-right py-2 px-3">Rp {parseFloat(breakdown.grossProfit).toLocaleString('id-ID')}</td>
                        <td className="text-right py-2 px-3">{breakdown.marginPercentage.toFixed(2)}%</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Report metadata */}
          <div className="text-sm text-gray-500 mt-4">
            Generated: {new Date(report.generatedAt).toLocaleString()}
          </div>
        </div>
      )}
    </div>
  );
}
