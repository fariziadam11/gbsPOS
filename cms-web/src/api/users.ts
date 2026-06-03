import apiClient from './client'
import type { ApiResponse, UserListItem, CreateUserRequest, UpdateUserRequest } from '../types/api'

export async function getUsers(): Promise<ApiResponse<UserListItem[]>> {
  const response = await apiClient.get<ApiResponse<UserListItem[]>>('/v1/users')
  return response.data
}

export async function getUser(id: number): Promise<ApiResponse<UserListItem>> {
  const response = await apiClient.get<ApiResponse<UserListItem>>(`/v1/users/${id}`)
  return response.data
}

export async function createUser(data: CreateUserRequest): Promise<ApiResponse<UserListItem>> {
  const response = await apiClient.post<ApiResponse<UserListItem>>('/v1/users', data)
  return response.data
}

export async function updateUser(id: number, data: UpdateUserRequest): Promise<ApiResponse<UserListItem>> {
  const response = await apiClient.put<ApiResponse<UserListItem>>(`/v1/users/${id}`, data)
  return response.data
}

export async function deleteUser(id: number): Promise<void> {
  await apiClient.delete(`/v1/users/${id}`)
}
