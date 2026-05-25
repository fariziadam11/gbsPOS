import axios from 'axios'
import type { ApiErrorResponse } from '../types/api'

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const isLoginRequest = error.config?.url?.endsWith('/login')
    if (error.response?.status === 401 && !isLoginRequest) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export function getApiError(err: unknown): ApiErrorResponse | null {
  if (axios.isAxiosError(err) && err.response?.data) {
    const data = err.response.data as ApiErrorResponse
    if (data.success === false && data.error) {
      return data
    }
  }
  return null
}

export function getErrorMessage(err: unknown): string {
  const apiError = getApiError(err)
  if (apiError) {
    return apiError.error.message
  }
  if (err instanceof Error) {
    return err.message
  }
  return 'An unexpected error occurred'
}

export default apiClient
