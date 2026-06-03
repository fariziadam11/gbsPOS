import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import type { Ref } from 'vue'
import { getCustomers, getCustomer, createCustomer, updateCustomer } from '../api/customers'
import type { UpdateCustomerRequest } from '../types/api'

export function useCustomers(query?: Ref<string | undefined>) {
  return useQuery({
    queryKey: ['customers', query?.value],
    queryFn: () => getCustomers(query?.value),
    select: (res) => (res.success ? res.data : null),
  })
}

export function useCustomer(id: Ref<number>) {
  return useQuery({
    queryKey: ['customer', id.value],
    queryFn: () => getCustomer(id.value),
    select: (res) => (res.success ? res.data : null),
    enabled: () => id.value > 0,
  })
}

export function useCreateCustomer() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createCustomer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] })
    },
  })
}

export function useUpdateCustomer() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateCustomerRequest }) => updateCustomer(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] })
    },
  })
}
