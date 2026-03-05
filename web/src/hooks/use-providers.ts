import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { providersApi } from '@/lib/api'
import type { Provider } from '@/types/api'

export function useProviders() {
  return useQuery({
    queryKey: ['providers'],
    queryFn: providersApi.list,
  })
}

export function useProvider(name: string) {
  return useQuery({
    queryKey: ['providers', name],
    queryFn: () => providersApi.get(name),
    enabled: !!name,
  })
}

export function useCreateProvider() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (provider: Provider) => providersApi.create(provider),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['providers'] })
    },
  })
}

export function useUpdateProvider() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ name, provider }: { name: string; provider: Partial<Provider> }) =>
      providersApi.update(name, provider),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['providers'] })
    },
  })
}

export function useDeleteProvider() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (name: string) => providersApi.delete(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['providers'] })
    },
  })
}
