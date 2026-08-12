import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import { z } from 'zod'

const apiErrorSchema = z.object({
  success: z.literal(false),
  error: z.object({
    code: z.string(),
    message: z.string(),
    details: z.array(z.object({ field: z.string(), message: z.string() })).optional(),
  }),
})

const envelopeSchema = z.object({
  success: z.boolean(),
  data: z.unknown().optional(),
  idempotent: z.boolean().optional(),
})

export type ApiEnvelope<T> = { success: true; data: T; idempotent?: boolean }
export type ApiError = { code: string; message: string; details?: { field: string; message: string }[] }

export class BackendError extends Error {
  code: string
  details?: { field: string; message: string }[]

  constructor(error: ApiError) {
    super(error.message)
    this.name = 'BackendError'
    this.code = error.code
    this.details = error.details
  }
}

export function getErrorMessage(error: unknown): string {
  if (error instanceof BackendError) return `${error.code}: ${error.message}`
  if (axios.isAxiosError(error)) {
    const detail = error.response?.data?.error
    if (detail?.message) return `${detail.code ?? 'API_ERROR'}: ${detail.message}`
    if (error.code === 'ERR_NETWORK') return 'API tidak dapat dihubungi. Periksa koneksi atau konfigurasi CORS.'
    return error.message
  }
  if (error instanceof Error) return error.message
  if (typeof error === 'string') return error
  if (error && typeof error === 'object' && 'message' in error) return String(error.message)
  return error ? JSON.stringify(error) : 'Request gagal tanpa detail error.'
}

const token = () => localStorage.getItem('gbs-token')

function createClient(baseURL: string): AxiosInstance {
  const client = axios.create({ baseURL: baseURL.replace(/\/$/, ''), timeout: 30000 })
  client.interceptors.request.use((config) => {
    const value = token()
    if (value) config.headers.Authorization = `Bearer ${value}`
    return config
  })
  client.interceptors.response.use(undefined, (error) => {
    if (error.response?.status === 401) window.dispatchEvent(new Event('gbs:unauthorized'))
    return Promise.reject(error)
  })
  return client
}

const defaultCmsApi = import.meta.env.DEV ? 'http://localhost:8081/v1' : 'https://api-cms.armmada.id/v1'
const defaultPosApi = import.meta.env.DEV ? 'http://localhost:8080/v1' : 'https://api-pos.armmada.id/v1'

export const apiBaseUrls = {
  cms: import.meta.env.VITE_CMS_API_BASE_URL || defaultCmsApi,
  pos: import.meta.env.VITE_POS_API_BASE_URL || defaultPosApi,
}
export const cmsApi = createClient(apiBaseUrls.cms)
export const posApi = createClient(apiBaseUrls.pos)

export async function requestData<T>(client: AxiosInstance, config: AxiosRequestConfig, schema: z.ZodType<T>): Promise<T> {
  const response = await client.request(config)
  if (response.status === 204) return undefined as T
  const parsed = envelopeSchema.safeParse(response.data)
  if (!parsed.success) throw new Error('Invalid backend response envelope')
  if (!parsed.data.success) {
    const failure = apiErrorSchema.safeParse(response.data)
    throw new BackendError(failure.success ? failure.data.error : { code: 'API_ERROR', message: 'Request failed' })
  }
  const result = schema.safeParse(parsed.data.data)
  if (!result.success) throw new Error(`Invalid backend payload: ${result.error.message}`)
  return result.data
}

export async function requestVoid(client: AxiosInstance, config: AxiosRequestConfig): Promise<void> {
  await requestData(client, config, z.unknown().optional())
}

export const idSchema = z.union([z.string(), z.number()])
export const dateSchema = z.union([z.string(), z.date()]).nullable().optional()
