import apiClient from './client'
import type { ApiResponse, LoginRequest, LoginResponse } from '../types/api'

export async function login(credentials: LoginRequest): Promise<ApiResponse<LoginResponse>> {
  const response = await apiClient.post<ApiResponse<LoginResponse>>('/v1/login', credentials)
  return response.data
}
