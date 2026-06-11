import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed } from 'vue'
import type { Ref } from 'vue'
import {
  cancelDiscount,
  createDiscount,
  deleteDiscount,
  getDiscount,
  getDiscounts,
  stopDiscount,
  updateDiscount,
} from '../api/discounts'
import type { UpdateDiscountRequest } from '../types/api'

export function useDiscounts(productId?: Ref<number | null>) {
  return useQuery({
    queryKey: computed(() => ['discounts', productId?.value]),
    queryFn: () => getDiscounts(productId?.value || undefined),
    select: (res) => (res.success ? res.data : null),
    enabled: () => !productId || !!productId.value,
  })
}

export function useDiscount(id: Ref<number>) {
  return useQuery({
    queryKey: computed(() => ['discount', id.value]),
    queryFn: () => getDiscount(id.value),
    select: (res) => (res.success ? res.data : null),
    enabled: () => id.value > 0,
  })
}

export function useCreateDiscount() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createDiscount,
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['discounts', variables.productId] })
      queryClient.invalidateQueries({ queryKey: ['discounts'] })
      queryClient.invalidateQueries({ queryKey: ['products'] })
    },
  })
}

export function useUpdateDiscount() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateDiscountRequest }) => updateDiscount(id, data),
    onSuccess: (_data, variables) => {
      if (variables.data.productId) {
        queryClient.invalidateQueries({ queryKey: ['discounts', variables.data.productId] })
      }
      queryClient.invalidateQueries({ queryKey: ['discounts'] })
      queryClient.invalidateQueries({ queryKey: ['discount', variables.id] })
      queryClient.invalidateQueries({ queryKey: ['products'] })
    },
  })
}

export function useStopDiscount() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: stopDiscount,
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['discounts'] })
      queryClient.invalidateQueries({ queryKey: ['discount', id] })
      queryClient.invalidateQueries({ queryKey: ['products'] })
    },
  })
}

export function useCancelDiscount() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: cancelDiscount,
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['discounts'] })
      queryClient.invalidateQueries({ queryKey: ['discount', id] })
      queryClient.invalidateQueries({ queryKey: ['products'] })
    },
  })
}

export function useDeleteDiscount() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteDiscount,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['discounts'] })
      queryClient.invalidateQueries({ queryKey: ['products'] })
    },
  })
}
