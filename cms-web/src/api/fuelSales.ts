import posApiClient from './pos-client'
import type { ApiResponse, FuelSalesReport } from '../types/api'

export async function getFuelSalesReport(from: string, to: string): Promise<ApiResponse<FuelSalesReport>> {
  const response = await posApiClient.get<ApiResponse<FuelSalesReport>>('/v1/fuel-sales/report', {
    params: { from, to },
  })
  return response.data
}
