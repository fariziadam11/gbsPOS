import { useQuery } from '@tanstack/vue-query'
import { getFuelSalesReport } from '../api/fuelSales'
import type { FuelSalesReport } from '../types/api'
import type { Ref, ComputedRef } from 'vue'

export function useFuelSalesReport(from: Ref<string> | ComputedRef<string>, to: Ref<string> | ComputedRef<string>, enabled: Ref<boolean> | ComputedRef<boolean>) {
  return useQuery<FuelSalesReport | null>({
    queryKey: ['fuel-sales-report', from, to],
    queryFn: async () => {
      const res = await getFuelSalesReport(from.value, to.value)
      return res.success ? res.data : null
    },
    enabled: () => enabled.value && !!from.value && !!to.value,
  })
}
