import apiClient from './client'
import type { ApiResponse, SettingsResponse, UpdateSettingsRequest } from '../types/api'

export async function getSettings(): Promise<ApiResponse<SettingsResponse>> {
  const response = await apiClient.get<ApiResponse<SettingsResponse>>('/v1/settings')
  return response.data
}

export async function updateSettings(data: UpdateSettingsRequest): Promise<ApiResponse<SettingsResponse>> {
  const response = await apiClient.put<ApiResponse<SettingsResponse>>('/v1/settings', data)
  return response.data
}
