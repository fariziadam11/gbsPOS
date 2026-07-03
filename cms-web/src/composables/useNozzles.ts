import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { getNozzles, createNozzle, updateNozzle, deleteNozzle } from '../api/nozzles'
import type { CreateNozzleRequest, UpdateNozzleRequest } from '../types/api'

export function useNozzles() {
  return useQuery({
    queryKey: ['nozzles'],
    queryFn: getNozzles,
    select: (res) => (res.success ? res.data : []),
  })
}

export function useCreateNozzle() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateNozzleRequest) => createNozzle(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nozzles'] })
    },
  })
}

export function useUpdateNozzle() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateNozzleRequest }) => updateNozzle(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nozzles'] })
    },
  })
}

export function useDeleteNozzle() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteNozzle(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nozzles'] })
    },
  })
}
