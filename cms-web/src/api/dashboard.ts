import posApiClient from './pos-client'
import type { ApiResponse, DashboardSummary, RevenuePoint, TopProduct } from '../types/api'

export async function getDashboardSummary(storeType?: string): Promise<ApiResponse<DashboardSummary>> {
  const response = await posApiClient.get<ApiResponse<DashboardSummary>>('/v1/dashboard/summary', {
    params: storeType ? { storeType } : {},
  })
  return response.data
}

export async function getRevenueTrend(days: number = 7, storeType?: string): Promise<ApiResponse<RevenuePoint[]>> {
  const response = await posApiClient.get<ApiResponse<RevenuePoint[]>>('/v1/dashboard/revenue', {
    params: { days, ...(storeType ? { storeType } : {}) },
  })
  return response.data
}

export async function getTopProducts(limit: number = 10, storeType?: string): Promise<ApiResponse<TopProduct[]>> {
  const response = await posApiClient.get<ApiResponse<TopProduct[]>>('/v1/dashboard/top-products', {
    params: { limit, ...(storeType ? { storeType } : {}) },
  })
  return response.data
}
