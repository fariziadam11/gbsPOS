import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { getFuelPrices, updateFuelPrice } from '../api/fuelPrices'
import type { UpdateFuelPriceRequest } from '../types/api'

export function useFuelPrices() {
  return useQuery({
    queryKey: ['fuel-prices'],
    queryFn: getFuelPrices,
    select: (res) => (res.success ? res.data : []),
  })
}

export function useUpdateFuelPrice() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ code, data }: { code: string; data: UpdateFuelPriceRequest }) =>
      updateFuelPrice(code, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['fuel-prices'] })
    },
  })
}
