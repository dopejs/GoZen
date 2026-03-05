import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@/test/utils'
import userEvent from '@testing-library/user-event'
import { MonitoringPage } from './index'

// Mock requestsApi.get for detail modal
vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual('@/lib/api')
  return {
    ...actual,
    requestsApi: {
      ...(actual as any).requestsApi,
      get: vi.fn().mockResolvedValue({
        id: 'req-123',
        timestamp: '2025-01-01T00:00:00Z',
        provider: 'anthropic',
        model: 'claude-sonnet-4',
        status_code: 200,
        duration_ms: 1500,
        input_tokens: 100,
        output_tokens: 50,
        cost_usd: 0.005,
        session_id: 'sess-123',
        client_type: 'claude',
        request_format: 'anthropic',
        request_size: 1024,
        failover_chain: [],
      }),
    },
  }
})

describe('MonitoringPage', () => {
  it('renders the monitoring page title', async () => {
    render(<MonitoringPage />)
    expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument()
  })

  it('renders request table with data from MSW', async () => {
    render(<MonitoringPage />)

    await waitFor(() => {
      expect(screen.getByText('anthropic')).toBeInTheDocument()
    })

    // Verify table columns are present
    expect(screen.getByText('200')).toBeInTheDocument()
  })

  it('renders refresh button', () => {
    render(<MonitoringPage />)
    expect(screen.getByRole('button', { name: /refresh/i })).toBeInTheDocument()
  })

  it('renders filter controls', async () => {
    render(<MonitoringPage />)

    await waitFor(() => {
      // Auto-refresh toggle
      expect(screen.getByRole('switch')).toBeInTheDocument()
    })
  })

  it('opens detail modal when clicking a request row', async () => {
    const user = userEvent.setup()
    render(<MonitoringPage />)

    await waitFor(() => {
      expect(screen.getByText('anthropic')).toBeInTheDocument()
    })

    // Click the row
    const row = screen.getByText('anthropic').closest('tr')
    if (row) {
      await user.click(row)
    }

    // Detail modal should appear
    await waitFor(() => {
      expect(screen.getByText('req-123')).toBeInTheDocument()
    })
  })
})
