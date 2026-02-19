import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { usageApi, budgetApi, providerHealthApi } from '@/lib/api'
import type { Budget } from '@/types/api'

export function useUsageSummary(period?: 'today' | 'week' | 'month') {
  return useQuery({
    queryKey: ['usage', 'summary', period],
    queryFn: () => usageApi.summary(period),
  })
}

export function useHourlyUsage(date?: string) {
  return useQuery({
    queryKey: ['usage', 'hourly', date],
    queryFn: () => usageApi.hourly(date),
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
