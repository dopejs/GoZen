import { describe, it, expect } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useUsageSummary, useHourlyUsage, useBudget, useUpdateBudget, useBudgetStatus, useProviderHealthList, useProviderHealth } from './use-usage'

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

describe('useBudget', () => {
  it('fetches budget', async () => {
    const { result } = renderHook(() => useBudget(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.enabled).toBe(true)
    expect(result.current.data?.monthly_limit).toBe(200)
  })
})

describe('useUpdateBudget', () => {
  it('updates budget', async () => {
    const { result } = renderHook(() => useUpdateBudget(), { wrapper: createWrapper() })

    await act(async () => {
      const response = await result.current.mutateAsync({ daily_limit: 20, enabled: true })
      expect(response).toBeDefined()
    })
  })
})

describe('useBudgetStatus', () => {
  it('fetches budget status', async () => {
    const { result } = renderHook(() => useBudgetStatus(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.daily_used).toBe(1.5)
    expect(result.current.data?.daily_remaining).toBe(8.5)
  })
})

describe('useProviderHealthList', () => {
  it('fetches provider health list', async () => {
    const { result } = renderHook(() => useProviderHealthList(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toHaveLength(1)
    expect(result.current.data?.[0].name).toBe('anthropic')
    expect(result.current.data?.[0].status).toBe('healthy')
  })
})

describe('useProviderHealth', () => {
  it('fetches single provider health', async () => {
    const { result } = renderHook(() => useProviderHealth('anthropic'), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.name).toBe('anthropic')
    expect(result.current.data?.status).toBe('healthy')
  })
})
