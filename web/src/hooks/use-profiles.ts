import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { profilesApi } from '@/lib/api'
import type { Profile } from '@/types/api'

export function useProfiles() {
  return useQuery({
    queryKey: ['profiles'],
    queryFn: profilesApi.list,
  })
}

export function useProfile(name: string) {
  return useQuery({
    queryKey: ['profiles', name],
    queryFn: () => profilesApi.get(name),
    enabled: !!name,
  })
}

export function useCreateProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (profile: Profile) => profilesApi.create(profile),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profiles'] })
    },
  })
}

export function useUpdateProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ name, profile }: { name: string; profile: Partial<Profile> }) =>
      profilesApi.update(name, profile),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profiles'] })
    },
  })
}

export function useDeleteProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (name: string) => profilesApi.delete(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profiles'] })
    },
  })
}
