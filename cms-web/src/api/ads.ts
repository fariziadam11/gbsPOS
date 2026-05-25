import apiClient from './client'
import type {
  Ad,
  AdListResponse,
  ApiResponse,
  CreateAdRequest,
  UpdateAdRequest,
  PlaylistResponse,
  ToggleAdResponse,
} from '../types/api'

export async function getAds(page = 1, limit = 20): Promise<ApiResponse<AdListResponse>> {
  const response = await apiClient.get<ApiResponse<AdListResponse>>('/v1/ads', {
    params: { page, limit },
  })
  return response.data
}

export async function getAd(id: number): Promise<ApiResponse<Ad>> {
  const response = await apiClient.get<ApiResponse<Ad>>(`/v1/ads/${id}`)
  return response.data
}

export async function createAd(data: CreateAdRequest): Promise<ApiResponse<Ad>> {
  const formData = new FormData()
  formData.append('file', data.file)
  formData.append('name', data.name)
  data.storeTypes.forEach((st) => formData.append('storeTypes', st))
  if (data.playlistOrder !== undefined) {
    formData.append('playlistOrder', String(data.playlistOrder))
  }
  if (data.startDate) formData.append('startDate', data.startDate)
  if (data.endDate) formData.append('endDate', data.endDate)
  if (data.startTime) formData.append('startTime', data.startTime)
  if (data.endTime) formData.append('endTime', data.endTime)

  const response = await apiClient.post<ApiResponse<Ad>>('/v1/ads/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return response.data
}

export async function updateAd(id: number, data: UpdateAdRequest): Promise<ApiResponse<Ad>> {
  const response = await apiClient.put<ApiResponse<Ad>>(`/v1/ads/${id}`, data)
  return response.data
}

export async function deleteAd(id: number): Promise<void> {
  await apiClient.delete(`/v1/ads/${id}`)
}

export async function toggleAd(id: number): Promise<ApiResponse<ToggleAdResponse>> {
  const response = await apiClient.post<ApiResponse<ToggleAdResponse>>(`/v1/ads/${id}/toggle`)
  return response.data
}

export async function getActivePlaylist(storeType: string): Promise<ApiResponse<PlaylistResponse>> {
  const response = await apiClient.get<ApiResponse<PlaylistResponse>>('/v1/ads/active', {
    params: { storeType },
  })
  return response.data
}

export function getDownloadUrl(id: number): string {
  return `${import.meta.env.VITE_API_BASE_URL}/v1/ads/download/${id}`
}
