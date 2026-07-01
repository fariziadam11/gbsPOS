import posApiClient from './pos-client'
import type { ApiResponse, Nozzle, CreateNozzleRequest, UpdateNozzleRequest } from '../types/api'

export async function getNozzles(): Promise<ApiResponse<Nozzle[]>> {
  const response = await posApiClient.get<ApiResponse<Nozzle[]>>('/v1/nozzles')
  return response.data
}

export async function createNozzle(data: CreateNozzleRequest): Promise<ApiResponse<Nozzle>> {
  const response = await posApiClient.post<ApiResponse<Nozzle>>('/v1/nozzles', data)
  return response.data
}

export async function updateNozzle(id: string, data: UpdateNozzleRequest): Promise<ApiResponse<Nozzle>> {
  const response = await posApiClient.put<ApiResponse<Nozzle>>(`/v1/nozzles/${id}`, data)
  return response.data
}

export async function deleteNozzle(id: string): Promise<void> {
  await posApiClient.delete(`/v1/nozzles/${id}`)
}
