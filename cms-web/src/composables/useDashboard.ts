import { toValue } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import type { MaybeRefOrGetter } from 'vue'
import { getDashboardSummary, getRevenueTrend, getTopProducts } from '../api/dashboard'
import type { DashboardDateRange } from '../api/dashboard'

export function useDashboardSummary(
  storeType?: MaybeRefOrGetter<string | undefined>,
  dateRange?: MaybeRefOrGetter<DashboardDateRange | undefined>,
) {
  return useQuery({
    queryKey: ['dashboard', 'summary', storeType, dateRange],
    queryFn: () => getDashboardSummary(toValue(storeType), toValue(dateRange)),
    select: (res) => (res.success ? res.data : null),
  })
}

export function useRevenueTrend(
  dateRange?: MaybeRefOrGetter<DashboardDateRange | undefined>,
  storeType?: MaybeRefOrGetter<string | undefined>,
) {
  return useQuery({
    queryKey: ['dashboard', 'revenue', dateRange, storeType],
    queryFn: () => getRevenueTrend(toValue(dateRange), toValue(storeType)),
    select: (res) => (res.success ? res.data : null),
  })
}

export function useTopProducts(limit: MaybeRefOrGetter<number> = 10, storeType?: MaybeRefOrGetter<string | undefined>) {
  return useQuery({
    queryKey: ['dashboard', 'top-products', limit, storeType],
    queryFn: () => getTopProducts(toValue(limit), toValue(storeType)),
    select: (res) => (res.success ? res.data : null),
  })
}

