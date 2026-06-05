import posApiClient from './pos-client'
import type { ApiResponse, Customer, CustomerDetail, CreateCustomerRequest, UpdateCustomerRequest } from '../types/api'

export async function getCustomers(query?: string): Promise<ApiResponse<Customer[]>> {
  const response = await posApiClient.get<ApiResponse<Customer[]>>('/v1/customers', {
    params: query ? { q: query } : {},
  })
  return response.data
}

export async function getCustomer(id: number): Promise<ApiResponse<CustomerDetail>> {
  const response = await posApiClient.get<ApiResponse<CustomerDetail>>(`/v1/customers/${id}`)
  return response.data
}

export async function createCustomer(data: CreateCustomerRequest): Promise<ApiResponse<Customer>> {
  const response = await posApiClient.post<ApiResponse<Customer>>('/v1/customers', data)
  return response.data
}

export async function updateCustomer(id: number, data: UpdateCustomerRequest): Promise<ApiResponse<Customer>> {
  const response = await posApiClient.put<ApiResponse<Customer>>(`/v1/customers/${id}`, data)
  return response.data
}
