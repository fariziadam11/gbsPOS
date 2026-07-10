import apiClient from './client'
import type { ApiResponse } from '../types/api'

export interface DeviceInfo {
  deviceModel: string | null
  deviceManufacturer: string | null
  deviceBrand: string | null
  androidVersion: string | null
  sdkInt: number | null
  appVersion: string | null
  appVersionCode: number | null
}

export interface Terminal {
  terminalId: string
  payload: string
  deviceModel: string | null
  deviceManufacturer: string | null
  deviceBrand: string | null
  androidVersion: string | null
  sdkInt: number | null
  appVersion: string | null
  appVersionCode: number | null
  updatedAt: string
}

export async function getTerminals(): Promise<ApiResponse<Terminal[]>> {
  const response = await apiClient.get<ApiResponse<Terminal[]>>('/v1/display/terminals')
  return response.data
}
