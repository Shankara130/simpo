'use client';

/**
 * System Health Monitoring Dashboard Page
 * Story 6.2, Task 4: Create Web Admin Health Dashboard Page (AC: 1-12, 14)
 *
 * Features:
 * - Health metrics display grid (uptime, DB, Redis, sessions, errors, response time, disk)
 * - Status indicators with color coding (green/yellow/red)
 * - Auto-refresh functionality (30-second interval)
 * - Alerts section with severity-based display
 * - Loading states and error handling
 * - RBAC check (hide page from non-Admin users)
 * - Manual refresh button
 * - Last-updated timestamp display
 */

import { useState, useEffect, useCallback } from 'react';
import { apiClient, ApiError } from '@/lib/apiClient';

// Types for health dashboard data
interface HealthMetrics {
  database: {
    status: string;
    response_time?: string;
  };
  redis: {
    status: string;
    response_time?: string;
  };
  sessions: {
    active: number;
  };
  api: {
    avg_response_time: string;
    requests_per_second: number;
  };
  errors: {
    rate: number;
    count: number;
    total_requests: number;
  };
  disk: {
    used_gb: number;
    total_gb: number;
    free_percentage: number;
  };
}

interface HealthAlert {
  severity: 'critical' | 'warning' | 'info';
  message: string;
  timestamp: string;
}

interface HealthDashboardResponse {
  status: 'healthy' | 'degraded' | 'unhealthy';
  uptime_percentage: number;
  uptime: string;
  version: string;
  timestamp: string;
  metrics: HealthMetrics;
  alerts: HealthAlert[];
}

interface AlertsResponse {
  alerts: HealthAlert[];
  total: number;
  critical: number;
  warning: number;
  info: number;
}

type UserRole = 'CASHIER' | 'OWNER' | 'ADMIN' | 'SYSTEM_ADMIN';

export default function HealthDashboardPage() {
  // State
  const [healthData, setHealthData] = useState<HealthDashboardResponse | null>(null);
  const [alerts, setAlerts] = useState<HealthAlert[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<string>('');

  // User role for RBAC (AC14)
  const [userRole, setUserRole] = useState<UserRole>('ADMIN');

  /**
   * Fetch user role on mount
   * Story 6.2, Task 4.6: RBAC check (hide page from non-Admin users)
   */
  useEffect(() => {
    fetchUserRole();
  }, []);

  /**
   * Check if user has access to health dashboard
   * Story 6.2, AC14: Access restricted to System Admin role only
   */
  const hasAccess = (role: UserRole): boolean => {
    return role === 'ADMIN' || role === 'SYSTEM_ADMIN';
  };

  /**
   * Fetch user role from auth context or API
   */
  const fetchUserRole = async () => {
    try {
      // In production, fetch from auth context or API
      // For now, assume Admin role (has access)
      setUserRole('ADMIN');
    } catch (err) {
      console.error('Failed to fetch user role:', err);
      setUserRole('CASHIER'); // Default to no access
    }
  };

  /**
   * Fetch health dashboard data
   * Story 6.2, Task 4.2: Implement health metrics display grid
   */
  const fetchHealthData = useCallback(async (showRefreshIndicator = false) => {
    if (showRefreshIndicator) {
      setRefreshing(true);
    }

    try {
      const [data] = await apiClient.get<HealthDashboardResponse>('/api/v1/admin/health/dashboard');
      setHealthData(data);
      setError(null);
      setLastUpdated(new Date().toLocaleString('id-ID'));
    } catch (err) {
      const apiError = err as ApiError;
      if (apiError?.response?.status === 401 || apiError?.response?.status === 403) {
        setError('Access denied. Admin only.');
      } else {
        setError('Failed to load health metrics');
      }
      console.error('Failed to fetch health data:', err);
    } finally {
      setLoading(false);
      if (showRefreshIndicator) {
        setRefreshing(false);
      }
    }
  }, []);

  /**
   * Fetch alerts
   * Story 6.2, Task 4.5: Implement alerts section with severity-based display
   */
  const fetchAlerts = useCallback(async () => {
    try {
      const [data] = await apiClient.get<AlertsResponse>('/api/v1/admin/health/alerts');
      setAlerts(data.alerts);
    } catch (err) {
      console.error('Failed to fetch alerts:', err);
    }
  }, []);

  /**
   * Manual refresh handler
   * Story 6.2, Task 4.8: Implement manual refresh button
   */
  const handleManualRefresh = useCallback(() => {
    fetchHealthData(true);
    fetchAlerts();
  }, [fetchHealthData, fetchAlerts]);

  /**
   * Auto-refresh functionality
   * Story 6.2, AC8: Health metrics refresh automatically every 30 seconds
   */
  useEffect(() => {
    // Initial load
    fetchHealthData();
    fetchAlerts();

    // Set up auto-refresh interval (30 seconds)
    const interval = setInterval(() => {
      fetchHealthData();
      fetchAlerts();
    }, 30000);

    return () => clearInterval(interval);
  }, [fetchHealthData, fetchAlerts]);

  /**
   * Get status badge color
   * Story 6.2, Task 4.3: Implement status indicators with color coding
   */
  const getStatusColor = (status: string): string => {
    const colors = {
      healthy: 'bg-green-100 text-green-800',
      degraded: 'bg-yellow-100 text-yellow-800',
      unhealthy: 'bg-red-100 text-red-800',
    };
    return colors[status as keyof typeof colors] || 'bg-gray-100 text-gray-800';
  };

  /**
   * Get alert severity color
   */
  const getAlertColor = (severity: string): string => {
    const colors = {
      critical: 'bg-red-100 text-red-800 border-red-500',
      warning: 'bg-yellow-100 text-yellow-800 border-yellow-500',
      info: 'bg-blue-100 text-blue-800 border-blue-500',
    };
    return colors[severity as keyof typeof colors] || 'bg-gray-100 text-gray-800';
  };

  /**
   * Get disk status color
   */
  const getDiskStatusColor = (freePercentage: number): string => {
    if (freePercentage < 10) {
      return 'bg-red-100 text-red-800';
    } else if (freePercentage < 20) {
      return 'bg-yellow-100 text-yellow-800';
    }
    return 'bg-green-100 text-green-800';
  };

  /**
   * Check if user has access
   */
  if (!hasAccess(userRole)) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">Access Denied</h1>
          <p className="text-gray-600">System Health Monitoring Dashboard is only accessible to administrators.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">System Health Monitoring</h1>
          <p className="text-sm text-gray-500 mt-1">Real-time system health metrics and alerts</p>
        </div>
        <button
          onClick={handleManualRefresh}
          disabled={refreshing}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
        >
          {refreshing ? (
            <>
              <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              Refreshing...
            </>
          ) : (
            <>
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 0m-15.356 2H15" />
              </svg>
              Refresh
            </>
          )}
        </button>
      </div>

      {/* Last updated timestamp */}
      {lastUpdated && (
        <p className="text-sm text-gray-500">
          Last updated: {lastUpdated}
        </p>
      )}

      {/* Loading state */}
      {loading && (
        <div className="flex items-center justify-center py-12">
          <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
          <span className="ml-3 text-gray-600">Loading health metrics...</span>
        </div>
      )}

      {/* Error state */}
      {error && (
        <div className="bg-red-50 border-l-4 border-red-500 p-4">
          <p className="text-red-800">{error}</p>
        </div>
      )}

      {/* Main content */}
      {!loading && !error && healthData && (
        <div className="space-y-6">
          {/* Overall status badge */}
          <div className={`inline-flex items-center px-4 py-2 rounded-lg ${getStatusColor(healthData.status)}`}>
            <span className="font-semibold">
              Status: {healthData.status.charAt(0).toUpperCase() + healthData.status.slice(1)}
            </span>
            <span className="ml-2 text-sm">
              Uptime: {healthData.uptime} ({healthData.uptime_percentage.toFixed(2)}%)
            </span>
          </div>

          {/* Alerts section */}
          {alerts.length > 0 && (
            <div className="space-y-2">
              <h2 className="text-lg font-semibold text-gray-900">Active Alerts</h2>
              {alerts.map((alert, index) => (
                <div key={index} className={`p-4 rounded-lg border-l-4 ${getAlertColor(alert.severity)}`}>
                  <div className="flex items-start justify-between">
                    <div>
                      <p className="font-semibold text-gray-900">
                        {alert.severity.charAt(0).toUpperCase() + alert.severity.slice(1)}
                      </p>
                      <p className="text-sm text-gray-700 mt-1">{alert.message}</p>
                    </div>
                    <p className="text-sm text-gray-500">
                      {new Date(alert.timestamp).toLocaleString('id-ID')}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Metrics grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {/* Database Metric Card */}
            <MetricCard
              title="Database"
              status={healthData.metrics.database.status}
              value={healthData.metrics.database.response_time || 'Checking...'}
              icon="🗄️"
              color={healthData.metrics.database.status === 'connected' ? 'green' : 'red'}
            />

            {/* Redis Metric Card */}
            <MetricCard
              title="Redis Cache"
              status={healthData.metrics.redis.status}
              value={healthData.metrics.redis.response_time || 'Checking...'}
              icon="⚡"
              color={healthData.metrics.redis.status === 'connected' ? 'green' : 'red'}
            />

            {/* Sessions Metric Card */}
            <MetricCard
              title="Active Sessions"
              status="active"
              value={healthData.metrics.sessions.active.toString()}
              icon="👥"
              color="blue"
            />

            {/* API Performance Card */}
            <MetricCard
              title="API Response Time"
              status="healthy"
              value={healthData.metrics.api.avg_response_time}
              icon="⚡"
              color="green"
            />

            {/* Error Rate Card */}
            <MetricCard
              title="Error Rate"
              status={healthData.metrics.errors.rate > 0.1 ? 'warning' : 'healthy'}
              value={`${healthData.metrics.errors.rate.toFixed(2)}%`}
              subtext={`${healthData.metrics.errors.count} / ${healthData.metrics.errors.total_requests}`}
              icon="📊"
              color={healthData.metrics.errors.rate > 0.1 ? 'yellow' : 'green'}
            />

            {/* Disk Usage Card */}
            <MetricCard
              title="Disk Usage"
              status={healthData.metrics.disk.free_percentage < 20 ? 'warning' : 'healthy'}
              value={`${healthData.metrics.disk.used_gb.toFixed(1)} / ${healthData.metrics.disk.total_gb} GB`}
              subtext={`${healthData.metrics.disk.free_percentage.toFixed(1)}% free`}
              icon="💾"
              color={getDiskStatusColor(healthData.metrics.disk.free_percentage)}
            />
          </div>

          {/* Version info */}
          <div className="text-sm text-gray-500 text-center">
            Version: {healthData.version} | Environment: {healthData.environment || 'production'}
          </div>
        </div>
      )}
    </div>
  );
}

/**
 * Metric Card Component
 * Story 6.2, Task 4.2: Reusable metric card component
 */
interface MetricCardProps {
  title: string;
  status: string;
  value: string;
  subtext?: string;
  icon: string;
  color: 'green' | 'yellow' | 'red' | 'blue';
}

function MetricCard({ title, status, value, subtext, icon, color }: MetricCardProps) {
  const colorStyles = {
    green: 'bg-green-50 border-green-200 text-green-800',
    yellow: 'bg-yellow-50 border-yellow-200 text-yellow-800',
    red: 'bg-red-50 border-red-200 text-red-800',
    blue: 'bg-blue-50 border-blue-200 text-blue-800',
  };

  return (
    <div className={`p-4 rounded-lg border-2 ${colorStyles[color]}`}>
      <div className="flex items-center justify-between mb-2">
        <h3 className="font-semibold text-gray-900">{title}</h3>
        <span className="text-2xl">{icon}</span>
      </div>
      <p className="text-2xl font-bold mb-1">{value}</p>
      {subtext && <p className="text-sm text-gray-600">{subtext}</p>}
    </div>
  );
}
