import type { Product, CardPayment } from './types'

const apiBase = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('gbs_pos_token')
  const response = await fetch(`${apiBase}${path}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}), ...options.headers },
  })
  if (!response.ok) throw new Error(await response.text())
  return response.json()
}

export async function login(username: string, password: string) {
  return request<{ data: { token: string } }>('/v1/login', { method: 'POST', body: JSON.stringify({ username, password }) })
}

export async function getProducts() {
  const response = await request<{ data: Product[] }>('/v1/products')
  return response.data
}

export async function createOrder(order: {
  id: string
  items: Array<{ productId: number; productName: string; productPrice: number; qty: number; subtotal: number }>
  subtotal: number
  total: number
  terminalId: string
}) {
  const response = await request<{ data: unknown }>('/v1/orders', {
    method: 'POST',
    body: JSON.stringify({ ...order, tax: 0, paymentMethod: 'CARD', timestamp: Date.now() }),
  })
  return response.data
}

export async function initCardPayment(orderId: string, amount: number, terminalId: string, deviceId: string) {
  const response = await request<{ data: CardPayment }>('/v1/card-payment/init', {
    method: 'POST',
    body: JSON.stringify({ orderId, amount, terminalId, deviceId }),
  })
  return response.data
}

export async function cancelCardPayment(paymentId: string) {
  return request<{ data: CardPayment }>(`/v1/card-payment/${paymentId}/cancel`, { method: 'POST' })
}

export function websocketUrl(terminalId: string) {
  const base = import.meta.env.VITE_WS_URL ?? 'ws://localhost:8080/ws'
  return `${base}?client_type=pos&terminal_id=${encodeURIComponent(terminalId)}&token=${encodeURIComponent(localStorage.getItem('gbs_pos_token') ?? '')}`
}
