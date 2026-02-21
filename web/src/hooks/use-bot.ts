import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { botApi } from '@/lib/api'
import type { BotConfig } from '@/types/api'

export function useBot() {
  return useQuery({
    queryKey: ['bot'],
    queryFn: botApi.get,
  })
}

export function useUpdateBot() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (bot: Partial<BotConfig>) => botApi.update(bot),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bot'] })
    },
  })
}
