import { useQuery } from '@tanstack/vue-query'
import type { Ref } from 'vue'
import { getOrders, getOrder } from '../api/orders'
import type { OrderFilters } from '../api/orders'

export function useOrders(filters?: Ref<OrderFilters | undefined>) {
  return useQuery({
    queryKey: ['orders', filters?.value],
    queryFn: () => getOrders(filters?.value),
    select: (res) => (res.success ? res.data : null),
  })
}

export function useOrder(id: Ref<string>) {
  return useQuery({
    queryKey: ['order', id.value],
    queryFn: () => getOrder(id.value),
    select: (res) => (res.success ? res.data : null),
    enabled: () => id.value.length > 0,
  })
}
