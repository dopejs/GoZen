import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { middlewareApi } from '@/lib/api'
import type { MiddlewareConfig } from '@/types/api'

export function useMiddleware() {
  return useQuery({
    queryKey: ['middleware'],
    queryFn: middlewareApi.get,
  })
}

export function useUpdateMiddleware() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (config: MiddlewareConfig) => middlewareApi.update(config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['middleware'] })
    },
  })
}

export function useEnableMiddleware() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (name: string) => middlewareApi.enable(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['middleware'] })
    },
  })
}

export function useDisableMiddleware() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (name: string) => middlewareApi.disable(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['middleware'] })
    },
  })
}

export function useReloadMiddleware() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => middlewareApi.reload(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['middleware'] })
    },
  })
}
