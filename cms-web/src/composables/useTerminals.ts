import { useQuery } from '@tanstack/vue-query'
import { getTerminals } from '../api/cartDisplay'

export function useTerminals() {
  return useQuery({
    queryKey: ['terminals'],
    queryFn: getTerminals,
    select: (res) => (res.success ? res.data : []),
    refetchInterval: 30000, // Refresh every 30 seconds
  })
}
