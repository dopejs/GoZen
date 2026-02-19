import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { settingsApi, bindingsApi, syncApi } from '@/lib/api'
import type { Settings, SyncConfig } from '@/types/api'

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: settingsApi.get,
  })
}

export function useUpdateSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (settings: Partial<Settings>) => settingsApi.update(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: ({ currentPassword, newPassword }: { currentPassword: string; newPassword: string }) =>
      settingsApi.changePassword(currentPassword, newPassword),
  })
}

export function useBindings() {
  return useQuery({
    queryKey: ['bindings'],
    queryFn: bindingsApi.list,
  })
}

export function useCreateBinding() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (binding: { path: string; profile: string }) => bindingsApi.create(binding),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bindings'] })
    },
  })
}

export function useDeleteBinding() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (path: string) => bindingsApi.delete(path),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bindings'] })
    },
  })
}

export function useSyncConfig() {
  return useQuery({
    queryKey: ['sync', 'config'],
    queryFn: syncApi.getConfig,
  })
}

export function useUpdateSyncConfig() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (config: Partial<SyncConfig>) => syncApi.updateConfig(config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sync'] })
    },
  })
}

export function useSyncStatus() {
  return useQuery({
    queryKey: ['sync', 'status'],
    queryFn: syncApi.status,
  })
}

export function useSyncPull() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: syncApi.pull,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sync'] })
      queryClient.invalidateQueries({ queryKey: ['providers'] })
      queryClient.invalidateQueries({ queryKey: ['profiles'] })
    },
  })
}

export function useSyncPush() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: syncApi.push,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sync'] })
    },
  })
}
