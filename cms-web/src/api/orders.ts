import posApiClient from './pos-client'
import type { ApiResponse, Order } from '../types/api'

export interface OrderFilters {
  storeType?: string
  startDate?: number
  endDate?: number
  isVoided?: boolean
  isSettled?: boolean
  paymentMethod?: string
}

export async function getOrders(filters?: OrderFilters): Promise<ApiResponse<Order[]>> {
  const params: Record<string, any> = {}
  if (filters) {
    if (filters.storeType) params.storeType = filters.storeType
    if (filters.startDate) params.startDate = filters.startDate
    if (filters.endDate) params.endDate = filters.endDate
    if (filters.isVoided !== undefined) params.isVoided = filters.isVoided
    if (filters.isSettled !== undefined) params.isSettled = filters.isSettled
    if (filters.paymentMethod) params.paymentMethod = filters.paymentMethod
  }
  const response = await posApiClient.get<ApiResponse<Order[]>>('/v1/orders', { params })
  return response.data
}

export async function getOrder(id: string): Promise<ApiResponse<Order>> {
  const response = await posApiClient.get<ApiResponse<Order>>(`/v1/orders/${id}`)
  return response.data
}
