import { describe, it, expect } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useSettings, useUpdateSettings, useBindings, useCreateBinding, useDeleteBinding, useSyncConfig, useSyncStatus } from './use-settings'

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('useSettings', () => {
  it('fetches settings', async () => {
    const { result } = renderHook(() => useSettings(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.default_profile).toBe('default')
    expect(result.current.data?.default_client).toBe('claude')
    expect(result.current.data?.clients).toContain('claude')
  })
})

describe('useUpdateSettings', () => {
  it('updates settings', async () => {
    const { result } = renderHook(() => useUpdateSettings(), { wrapper: createWrapper() })

    let response: unknown
    await act(async () => {
      response = await result.current.mutateAsync({ default_profile: 'work' })
    })

    expect(response).toBeDefined()
  })
})

describe('useBindings', () => {
  it('fetches bindings', async () => {
    const { result } = renderHook(() => useBindings(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.bindings).toHaveLength(1)
    expect(result.current.data?.bindings[0].path).toBe('/test/project')
    expect(result.current.data?.profiles).toContain('default')
  })
})

describe('useCreateBinding', () => {
  it('creates a binding', async () => {
    const { result } = renderHook(() => useCreateBinding(), { wrapper: createWrapper() })

    let response: unknown
    await act(async () => {
      response = await result.current.mutateAsync({ path: '/new/project', profile: 'default' })
    })

    expect(response).toBeDefined()
  })
})

describe('useDeleteBinding', () => {
  it('deletes a binding', async () => {
    const { result } = renderHook(() => useDeleteBinding(), { wrapper: createWrapper() })

    await act(async () => {
      await result.current.mutateAsync('/test/project')
    })

    expect(result.current.isSuccess).toBe(true)
  })
})

describe('useSyncConfig', () => {
  it('fetches sync config', async () => {
    const { result } = renderHook(() => useSyncConfig(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.configured).toBe(false)
  })
})

describe('useSyncStatus', () => {
  it('fetches sync status', async () => {
    const { result } = renderHook(() => useSyncStatus(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.configured).toBe(false)
  })
})
