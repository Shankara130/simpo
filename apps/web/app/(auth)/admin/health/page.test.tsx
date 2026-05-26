import { render, screen } from '@testing-library/react'
import { HealthDashboardPage } from '../page'

// Mock the API client
jest.mock('@/lib/apiClient', () => ({
  apiClient: {
    get: jest.fn(),
  },
}))

describe('HealthDashboardPage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('renders health metrics dashboard', () => {
    const { getByText, getByTestId } = render(<HealthDashboardPage />)

    // Should render main title
    expect(getByText(/System Health/i)).toBeInTheDocument()

    // Should render loading state initially
    expect(getByText(/Loading health metrics/i)).toBeInTheDocument()
  })

  test('displays health metrics cards', async () => {
    // Mock successful API response
    const mockHealthData = {
      status: 'healthy',
      uptime_percentage: 99.8,
      uptime: '15d 4h 32m',
      version: '1.0.0',
      timestamp: '2026-05-27T00:00:00Z',
      metrics: {
        database: { status: 'connected', response_time: '5ms' },
        redis: { status: 'connected', response_time: '2ms' },
        sessions: { active: 15 },
        api: { avg_response_time: '45ms', requests_per_second: 12.5 },
        errors: { rate: 0.05, count: 23, total_requests: 46000 },
        disk: { used_gb: 45.2, total_gb: 100, free_percentage: 54.8 },
      },
      alerts: [],
    }

    const { getByText, findByTestId } = render(<HealthDashboardPage />)

    // Wait for data to load
    await waitFor(() => {
      expect(getByText(/Database.*Connected/i)).toBeInTheDocument()
      expect(getByText(/Redis.*Connected/i)).toBeInTheDocument()
    })
  })

  test('displays alerts when present', async () => {
    const mockHealthData = {
      status: 'degraded',
      alerts: [
        {
          severity: 'warning',
          message: 'Disk space below 20%',
          timestamp: '2026-05-27T00:00:00Z',
        },
      ],
    }

    const { getByText } = render(<HealthDashboardPage />)

    await waitFor(() => {
      expect(getByText(/Disk space below 20%/i)).toBeInTheDocument()
    })
  })

  test('auto-refreshes every 30 seconds', () => {
    jest.useFakeTimers()

    const { getByText } = render(<HealthDashboardPage />)

    // Fast-forward time
    jest.advanceTimersByTime(30000)

    // Should trigger data refresh
    await waitFor(() => {
      expect(getByText(/Refreshing.../i)).toBeInTheDocument()
    })

    jest.useRealTimers()
  })

  test('displays manual refresh button', () => {
    const { getByText } = render(<HealthDashboardPage />)

    expect(getByText(/Refresh/i)).toBeInTheDocument()
  })

  test('hides page from non-admin users', () => {
    // Mock user role as cashier
    const { queryByText } = render(<HealthDashboardPage />, {
      userRole: 'CASHIER',
    })

    // Should show access denied message
    expect(queryByText(/Access Denied/i)).toBeInTheDocument()
    expect(queryByText(/Admin only/i)).toBeInTheDocument()
  })
})

function waitFor<T>(callback: () => T): Promise<T> {
  return new Promise((resolve) => {
    setTimeout(() => resolve(callback()), 100)
  })
}
