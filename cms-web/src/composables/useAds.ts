import { ref } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import type { Ref } from 'vue'
import {
  getAds,
  getAd,
  createAd,
  updateAd,
  deleteAd,
  toggleAd,
} from '../api/ads'
import type { UpdateAdRequest } from '../types/api'

export function useAds(page: Ref<number> = ref(1), limit: Ref<number> = ref(20)) {
  return useQuery({
    queryKey: ['ads', page, limit],
    queryFn: () => getAds(page.value, limit.value),
    select: (res) => (res.success ? res.data : null),
  })
}

export function useAd(id: number) {
  return useQuery({
    queryKey: ['ad', id],
    queryFn: () => getAd(id),
    select: (res) => (res.success ? res.data : null),
    enabled: id > 0,
  })
}

export function useCreateAd() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createAd,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ads'] })
    },
  })
}

export function useUpdateAd() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateAdRequest }) =>
      updateAd(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['ads'] })
      queryClient.invalidateQueries({ queryKey: ['ad', variables.id] })
    },
  })
}

export function useDeleteAd() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteAd,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ads'] })
    },
  })
}

export function useToggleAd() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: toggleAd,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ads'] })
    },
  })
}
