import { describe, it, expect } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useBot, useUpdateBot } from './use-bot'

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('useBot', () => {
  it('fetches bot config', async () => {
    const { result } = renderHook(() => useBot(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.enabled).toBe(true)
    expect(result.current.data?.profile).toBe('default')
    expect(result.current.data?.platforms?.telegram?.enabled).toBe(true)
  })
})

describe('useUpdateBot', () => {
  it('updates bot config', async () => {
    const { result } = renderHook(() => useUpdateBot(), { wrapper: createWrapper() })

    let response: unknown
    await act(async () => {
      response = await result.current.mutateAsync({ enabled: false })
    })

    expect(response).toBeDefined()
  })
})
