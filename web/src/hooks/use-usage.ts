import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { usageApi, budgetApi, providerHealthApi } from '@/lib/api'
import type { Budget } from '@/types/api'

export function useUsageSummary(params?: {
  period?: 'today' | 'week' | 'month'
  since?: string
  until?: string
  project?: string
}) {
  return useQuery({
    queryKey: ['usage', 'summary', params],
    queryFn: () => usageApi.summary(params),
  })
}

export function useHourlyUsage(params?: {
  hours?: number
  since?: string
  until?: string
  groupBy?: 'provider' | 'model'
}) {
  return useQuery({
    queryKey: ['usage', 'hourly', params],
    queryFn: () => usageApi.hourly(params) as Promise<unknown>,
  })
}

export function useBudget() {
  return useQuery({
    queryKey: ['budget'],
    queryFn: budgetApi.get,
  })
}

export function useUpdateBudget() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (budget: Partial<Budget>) => budgetApi.update(budget),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budget'] })
    },
  })
}

export function useBudgetStatus() {
  return useQuery({
    queryKey: ['budget', 'status'],
    queryFn: budgetApi.status,
  })
}

export function useProviderHealthList() {
  return useQuery({
    queryKey: ['health', 'providers'],
    queryFn: providerHealthApi.list,
    refetchInterval: 30000, // Refresh every 30 seconds
  })
}

export function useProviderHealth(name: string) {
  return useQuery({
    queryKey: ['health', 'providers', name],
    queryFn: () => providerHealthApi.get(name),
    enabled: !!name,
  })
}
