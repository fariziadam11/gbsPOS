import { defineStore } from 'pinia'
import type { User } from '../types/api'

interface AuthState {
  token: string | null
  user: User | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: localStorage.getItem('token'),
    user: null,
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
    isAdmin: (state) => state.user?.role === 'ADMIN',
    username: (state) => state.user?.username || '',
  },
  actions: {
    setToken(token: string) {
      this.token = token
      localStorage.setItem('token', token)
    },
    setUser(user: User) {
      this.user = user
      localStorage.setItem('user', JSON.stringify(user))
    },
    logout() {
      this.token = null
      this.user = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    },
    init() {
      const token = localStorage.getItem('token')
      const userJson = localStorage.getItem('user')
      if (token) {
        this.token = token
      }
      if (userJson) {
        try {
          this.user = JSON.parse(userJson) as User
        } catch {
          this.user = null
        }
      }
    },
  },
})
