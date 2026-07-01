import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { getPumps, createPump, updatePump, deletePump } from '../api/pumps'
import type { CreatePumpRequest, UpdatePumpRequest } from '../types/api'

export function usePumps() {
  return useQuery({
    queryKey: ['pumps'],
    queryFn: getPumps,
    select: (res) => (res.success ? res.data : []),
  })
}

export function useCreatePump() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreatePumpRequest) => createPump(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pumps'] })
    },
  })
}

export function useUpdatePump() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdatePumpRequest }) => updatePump(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pumps'] })
    },
  })
}

export function useDeletePump() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deletePump(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pumps'] })
    },
  })
}
