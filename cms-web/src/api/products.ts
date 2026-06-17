import posApiClient from './pos-client'
import type { ApiResponse, Product, CreateProductRequest, UpdateProductRequest, ImportResult } from '../types/api'

export async function getProducts(storeType?: string): Promise<ApiResponse<Product[]>> {
  const response = await posApiClient.get<ApiResponse<Product[]>>('/v1/products', {
    params: storeType ? { storeType } : {},
  })
  return response.data
}

export async function getProduct(id: number): Promise<ApiResponse<Product>> {
  const response = await posApiClient.get<ApiResponse<Product>>(`/v1/products/${id}`)
  return response.data
}

export async function createProduct(data: CreateProductRequest): Promise<ApiResponse<Product>> {
  const response = await posApiClient.post<ApiResponse<Product>>('/v1/products', data)
  return response.data
}

export async function updateProduct(id: number, data: UpdateProductRequest): Promise<ApiResponse<Product>> {
  const response = await posApiClient.put<ApiResponse<Product>>(`/v1/products/${id}`, data)
  return response.data
}

export async function deleteProduct(id: number): Promise<void> {
  await posApiClient.delete(`/v1/products/${id}`)
}

export async function importProducts(file: File, storeType?: string): Promise<ApiResponse<ImportResult>> {
  const formData = new FormData()
  formData.append('file', file)
  if (storeType) formData.append('storeType', storeType)
  const response = await posApiClient.post<ApiResponse<ImportResult>>('/v1/products/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return response.data
}

export function getExportUrl(storeType?: string): string {
  const baseUrl = import.meta.env.VITE_POS_API_BASE_URL
  const token = localStorage.getItem('token')
  const params = storeType ? `?storeType=${encodeURIComponent(storeType)}` : ''
  return `${baseUrl}/v1/products/export${params}&token=${encodeURIComponent(token || '')}`
}

// Variants
export async function getVariants(productId: number): Promise<ApiResponse<VariantItem[]>> {
  const response = await posApiClient.get<ApiResponse<VariantItem[]>>(`/v1/products/${productId}/variants`)
  return response.data
}

export async function createVariant(productId: number, data: CreateVariantReq): Promise<ApiResponse<VariantItem>> {
  const response = await posApiClient.post<ApiResponse<VariantItem>>(`/v1/products/${productId}/variants`, data)
  return response.data
}

export async function updateVariant(id: number, data: CreateVariantReq): Promise<ApiResponse<VariantItem>> {
  const response = await posApiClient.put<ApiResponse<VariantItem>>(`/v1/variants/${id}`, data)
  return response.data
}

export async function deleteVariant(id: number): Promise<void> {
  await posApiClient.delete(`/v1/variants/${id}`)
}

export interface VariantItem {
  id: number
  productId: number
  sku: string | null
  attributes: Record<string, string>
  price: number | null
  stockQuantity: number
  lowStockThreshold: number | null
  isActive: boolean
  sortOrder: number
}

export interface CreateVariantReq {
  sku?: string | null
  attributes: Record<string, string>
  price?: number | null
  stockQuantity?: number
  lowStockThreshold?: number | null
  sortOrder?: number
}
