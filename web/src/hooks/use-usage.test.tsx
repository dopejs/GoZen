import { describe, it, expect } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useUsageSummary, useHourlyUsage } from './use-usage'

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('useUsageSummary', () => {
  it('fetches usage summary', async () => {
    const { result } = renderHook(() => useUsageSummary(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.total_input_tokens).toBe(10000)
    expect(result.current.data?.total_output_tokens).toBe(5000)
    expect(result.current.data?.total_cost).toBe(1.5)
    expect(result.current.data?.request_count).toBe(100)
  })

  it('fetches usage summary with params', async () => {
    const { result } = renderHook(
      () => useUsageSummary({ period: 'week' }),
      { wrapper: createWrapper() }
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.total_input_tokens).toBeDefined()
  })
})

describe('useHourlyUsage', () => {
  it('fetches hourly usage', async () => {
    const { result } = renderHook(() => useHourlyUsage(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toBeDefined()
  })

  it('fetches hourly usage with params', async () => {
    const { result } = renderHook(
      () => useHourlyUsage({ hours: 24 }),
      { wrapper: createWrapper() }
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
  })
})
