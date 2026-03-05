import { describe, it, expect } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useProviders, useProvider, useCreateProvider, useUpdateProvider, useDeleteProvider } from './use-providers'

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('useProviders', () => {
  it('fetches providers list', async () => {
    const { result } = renderHook(() => useProviders(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toHaveLength(2)
    expect(result.current.data?.[0].name).toBe('anthropic')
  })
})

describe('useProvider', () => {
  it('fetches a single provider', async () => {
    const { result } = renderHook(() => useProvider('anthropic'), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.name).toBe('anthropic')
    expect(result.current.data?.base_url).toBe('https://api.anthropic.com')
  })

  it('does not fetch when name is empty', () => {
    const { result } = renderHook(() => useProvider(''), { wrapper: createWrapper() })

    expect(result.current.isFetching).toBe(false)
  })
})

describe('useCreateProvider', () => {
  it('creates a provider', async () => {
    const { result } = renderHook(() => useCreateProvider(), { wrapper: createWrapper() })

    let response: unknown
    await act(async () => {
      response = await result.current.mutateAsync({
        name: 'new-provider',
        base_url: 'https://api.new.com',
        auth_token: 'token',
      })
    })

    expect(response).toBeDefined()
  })
})

describe('useUpdateProvider', () => {
  it('updates a provider', async () => {
    const { result } = renderHook(() => useUpdateProvider(), { wrapper: createWrapper() })

    let response: unknown
    await act(async () => {
      response = await result.current.mutateAsync({
        name: 'anthropic',
        provider: { base_url: 'https://updated.com' },
      })
    })

    expect(response).toBeDefined()
  })
})

describe('useDeleteProvider', () => {
  it('deletes a provider', async () => {
    const { result } = renderHook(() => useDeleteProvider(), { wrapper: createWrapper() })

    let response: unknown
    await act(async () => {
      response = await result.current.mutateAsync('anthropic')
    })

    expect(response).toBeDefined()
  })
})
