import posApiClient from './pos-client'
import type { ApiResponse, Pump, CreatePumpRequest, UpdatePumpRequest } from '../types/api'

export async function getPumps(): Promise<ApiResponse<Pump[]>> {
  const response = await posApiClient.get<ApiResponse<Pump[]>>('/v1/pumps')
  return response.data
}

export async function createPump(data: CreatePumpRequest): Promise<ApiResponse<Pump>> {
  const response = await posApiClient.post<ApiResponse<Pump>>('/v1/pumps', data)
  return response.data
}

export async function updatePump(id: string, data: UpdatePumpRequest): Promise<ApiResponse<Pump>> {
  const response = await posApiClient.put<ApiResponse<Pump>>(`/v1/pumps/${id}`, data)
  return response.data
}

export async function deletePump(id: string): Promise<void> {
  await posApiClient.delete(`/v1/pumps/${id}`)
}
