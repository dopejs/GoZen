import { useQuery } from '@tanstack/react-query'
import { requestsApi } from '@/lib/api'

export interface RequestsParams {
  provider?: string
  session?: string
  model?: string
  status_min?: number
  status_max?: number
  limit?: number
}

export function useRequests(params?: RequestsParams, refetchInterval?: number) {
  return useQuery({
    queryKey: ['requests', params],
    queryFn: () => requestsApi.list(params),
    refetchInterval,
  })
}
