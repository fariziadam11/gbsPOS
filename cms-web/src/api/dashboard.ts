import posApiClient from './pos-client'
import type { ApiResponse, DashboardSummary, RevenuePoint, TopProduct } from '../types/api'

export interface DashboardDateRange {
  startDate?: string
  endDate?: string
}

function cleanParams(params: Record<string, any>): Record<string, any> {
  return Object.fromEntries(Object.entries(params).filter(([, v]) => v !== undefined && v !== null && v !== ''))
}

export async function getDashboardSummary(
  storeType?: string,
  dateRange?: DashboardDateRange,
): Promise<ApiResponse<DashboardSummary>> {
  const response = await posApiClient.get<ApiResponse<DashboardSummary>>('/v1/dashboard/summary', {
    params: cleanParams({
      ...(storeType ? { storeType } : {}),
      ...(dateRange?.startDate ? { startDate: dateRange.startDate } : {}),
      ...(dateRange?.endDate ? { endDate: dateRange.endDate } : {}),
    }),
  })
  return response.data
}

export async function getRevenueTrend(
  dateRange?: DashboardDateRange,
  storeType?: string,
): Promise<ApiResponse<RevenuePoint[]>> {
  const response = await posApiClient.get<ApiResponse<RevenuePoint[]>>('/v1/dashboard/revenue', {
    params: cleanParams({
      ...(dateRange?.startDate ? { startDate: dateRange.startDate } : {}),
      ...(dateRange?.endDate ? { endDate: dateRange.endDate } : {}),
      ...(storeType ? { storeType } : {}),
    }),
  })
  return response.data
}

export async function getTopProducts(limit: number = 10, storeType?: string): Promise<ApiResponse<TopProduct[]>> {
  const response = await posApiClient.get<ApiResponse<TopProduct[]>>('/v1/dashboard/top-products', {
    params: cleanParams({ limit, ...(storeType ? { storeType } : {}) }),
  })
  return response.data
}
