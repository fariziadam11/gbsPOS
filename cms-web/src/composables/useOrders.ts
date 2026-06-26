import { useQuery } from '@tanstack/vue-query'
import { toValue } from 'vue'
import type { MaybeRefOrGetter } from 'vue'
import { getOrders, getOrder } from '../api/orders'
import type { OrderFilters } from '../api/orders'

export function useOrders(filters?: MaybeRefOrGetter<OrderFilters | undefined>) {
  return useQuery({
    queryKey: ['orders', filters],
    queryFn: () => getOrders(toValue(filters)),
    select: (res) => (res.success ? res.data : null),
  })
}

export function useOrder(id: MaybeRefOrGetter<string>) {
  return useQuery({
    queryKey: ['order', id],
    queryFn: () => getOrder(toValue(id)),
    select: (res) => (res.success ? res.data : null),
    enabled: () => (toValue(id)?.length || 0) > 0,
  })
}
