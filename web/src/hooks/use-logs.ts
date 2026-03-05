import { useQuery } from '@tanstack/react-query'
import { logsApi } from '@/lib/api'

export interface LogsParams {
  provider?: string
  session_id?: string
  client_type?: string
  errors_only?: boolean
  limit?: number
}

export function useLogs(params?: LogsParams, refetchInterval?: number) {
  return useQuery({
    queryKey: ['logs', params],
    queryFn: () => logsApi.list(params),
    refetchInterval,
  })
}
