import { describe, it, expect } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useProfiles, useProfile, useCreateProfile, useUpdateProfile, useDeleteProfile } from './use-profiles'

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('useProfiles', () => {
  it('fetches profiles list', async () => {
    const { result } = renderHook(() => useProfiles(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toHaveLength(2)
    expect(result.current.data?.[0].name).toBe('default')
    expect(result.current.data?.[0].providers).toContain('anthropic')
  })
})

describe('useProfile', () => {
  it('fetches a single profile', async () => {
    const { result } = renderHook(() => useProfile('default'), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.name).toBe('default')
    expect(result.current.data?.providers).toContain('anthropic')
  })

  it('does not fetch when name is empty', () => {
    const { result } = renderHook(() => useProfile(''), { wrapper: createWrapper() })

    expect(result.current.isFetching).toBe(false)
  })
})

describe('useCreateProfile', () => {
  it('creates a profile', async () => {
    const { result } = renderHook(() => useCreateProfile(), { wrapper: createWrapper() })

    let response: unknown
    await act(async () => {
      response = await result.current.mutateAsync({
        name: 'new-profile',
        providers: ['anthropic'],
      })
    })

    expect(response).toBeDefined()
  })
})

describe('useUpdateProfile', () => {
  it('updates a profile', async () => {
    const { result } = renderHook(() => useUpdateProfile(), { wrapper: createWrapper() })

    let response: unknown
    await act(async () => {
      response = await result.current.mutateAsync({
        name: 'default',
        profile: { providers: ['anthropic'] },
      })
    })

    expect(response).toBeDefined()
  })
})

describe('useDeleteProfile', () => {
  it('deletes a profile', async () => {
    const { result } = renderHook(() => useDeleteProfile(), { wrapper: createWrapper() })

    let response: unknown
    await act(async () => {
      response = await result.current.mutateAsync('work')
    })

    expect(response).toBeDefined()
  })
})
