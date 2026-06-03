import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { getSettings, updateSettings } from '../api/settings'

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
    select: (res) => (res.success ? res.data : null),
  })
}

export function useUpdateSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updateSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}
