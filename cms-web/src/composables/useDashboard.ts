import { ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import type { Ref } from 'vue'
import { getDashboardSummary, getRevenueTrend, getTopProducts } from '../api/dashboard'

export function useDashboardSummary(storeType?: Ref<string | undefined>) {
  return useQuery({
    queryKey: ['dashboard', 'summary', storeType?.value],
    queryFn: () => getDashboardSummary(storeType?.value),
    select: (res) => (res.success ? res.data : null),
  })
}

export function useRevenueTrend(days: Ref<number> = ref(7), storeType?: Ref<string | undefined>) {
  return useQuery({
    queryKey: ['dashboard', 'revenue', days.value, storeType?.value],
    queryFn: () => getRevenueTrend(days.value, storeType?.value),
    select: (res) => (res.success ? res.data : null),
  })
}

export function useTopProducts(limit: Ref<number> = ref(10), storeType?: Ref<string | undefined>) {
  return useQuery({
    queryKey: ['dashboard', 'top-products', limit.value, storeType?.value],
    queryFn: () => getTopProducts(limit.value, storeType?.value),
    select: (res) => (res.success ? res.data : null),
  })
}
