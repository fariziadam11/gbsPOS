import posApiClient from './pos-client'
import type { ApiResponse, CreateDiscountRequest, Discount, UpdateDiscountRequest } from '../types/api'

export async function getDiscounts(productId?: number): Promise<ApiResponse<Discount[]>> {
  const response = await posApiClient.get<ApiResponse<Discount[]>>('/v1/discounts', {
    params: productId ? { productId } : {},
  })
  return response.data
}

export async function getDiscount(id: number): Promise<ApiResponse<Discount>> {
  const response = await posApiClient.get<ApiResponse<Discount>>(`/v1/discounts/${id}`)
  return response.data
}

export async function createDiscount(data: CreateDiscountRequest): Promise<ApiResponse<Discount>> {
  const response = await posApiClient.post<ApiResponse<Discount>>('/v1/discounts', data)
  return response.data
}

export async function updateDiscount(id: number, data: UpdateDiscountRequest): Promise<ApiResponse<Discount>> {
  const response = await posApiClient.put<ApiResponse<Discount>>(`/v1/discounts/${id}`, data)
  return response.data
}

export async function stopDiscount(id: number): Promise<ApiResponse<Discount>> {
  const response = await posApiClient.patch<ApiResponse<Discount>>(`/v1/discounts/${id}/stop`)
  return response.data
}

export async function cancelDiscount(id: number): Promise<ApiResponse<Discount>> {
  const response = await posApiClient.patch<ApiResponse<Discount>>(`/v1/discounts/${id}/cancel`)
  return response.data
}

export async function deleteDiscount(id: number): Promise<void> {
  await posApiClient.delete(`/v1/discounts/${id}`)
}
