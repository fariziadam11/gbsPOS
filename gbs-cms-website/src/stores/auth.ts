import { defineStore } from 'pinia'
import { z } from 'zod'
import { cmsApi, requestData } from '../lib/api'

const userSchema = z.object({ id: z.number(), username: z.string(), name: z.string(), role: z.string() })
const loginSchema = z.object({ user: userSchema, token: z.string() })
export type AuthUser = z.infer<typeof userSchema>

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('gbs-token'),
    user: JSON.parse(localStorage.getItem('gbs-user') ?? 'null') as AuthUser | null,
  }),
  getters: { isAuthenticated: (state) => Boolean(state.token && state.user) },
  actions: {
    async login(username: string, password: string) {
      const result = await requestData(cmsApi, { method: 'POST', url: '/login', data: { username, password } }, loginSchema)
      this.token = result.token
      this.user = result.user
      localStorage.setItem('gbs-token', result.token)
      localStorage.setItem('gbs-user', JSON.stringify(result.user))
    },
    logout() {
      this.token = null
      this.user = null
      localStorage.removeItem('gbs-token')
      localStorage.removeItem('gbs-user')
    },
  },
})
