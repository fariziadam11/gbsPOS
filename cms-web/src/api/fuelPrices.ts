import posApiClient from './pos-client'
import type { ApiResponse, FuelPrice, UpdateFuelPriceRequest } from '../types/api'

export async function getFuelPrices(): Promise<ApiResponse<FuelPrice[]>> {
  const response = await posApiClient.get<ApiResponse<FuelPrice[]>>('/v1/fuel-prices')
  return response.data
}

export async function updateFuelPrice(code: string, data: UpdateFuelPriceRequest): Promise<ApiResponse<FuelPrice>> {
  const response = await posApiClient.put<ApiResponse<FuelPrice>>(`/v1/fuel-prices/${code}`, data)
  return response.data
}
