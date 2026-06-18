import { defineStore } from 'pinia'
import { userManager, extractRoles, parseTokenExpiry } from '../keycloak'
import type { User } from 'oidc-client-ts'

interface StoredUser {
  username: string
  name: string
  role: string
}

interface AuthState {
  token: string | null
  idToken: string | null
  user: StoredUser | null
  expiresAt: number | null
}

const TOKEN_KEY = 'token'
const ID_TOKEN_KEY = 'id_token'
const USER_KEY = 'user'
const EXPIRES_AT_KEY = 'expires_at'

function parseUserFromKeycloak(user: User): StoredUser {
  const username =
    user.profile.preferred_username ||
    user.profile.email ||
    user.profile.sub ||
    ''
  const name = (user.profile.name as string) || username
  const roles = extractRoles(user)
  const role = roles.includes('ADMIN')
    ? 'ADMIN'
    : roles.includes('CASHIER')
      ? 'CASHIER'
      : 'CASHIER'
  return { username, name, role }
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: localStorage.getItem(TOKEN_KEY),
    idToken: localStorage.getItem(ID_TOKEN_KEY),
    user: null,
    expiresAt: localStorage.getItem(EXPIRES_AT_KEY)
      ? Number(localStorage.getItem(EXPIRES_AT_KEY))
      : null,
  }),
  getters: {
    isAuthenticated: (state) => {
      if (!state.token) return false
      if (state.expiresAt && state.expiresAt < Date.now()) return false
      return true
    },
    isAdmin: (state) => state.user?.role === 'ADMIN',
    username: (state) => state.user?.name || state.user?.username || 'User',
  },
  actions: {
    setUserSession(user: User) {
      const stored = parseUserFromKeycloak(user)
      this.token = user.access_token
      this.idToken = user.id_token ?? null
      this.user = stored
      this.expiresAt = parseTokenExpiry(user.access_token)

      localStorage.setItem(TOKEN_KEY, user.access_token)
      if (user.id_token) {
        localStorage.setItem(ID_TOKEN_KEY, user.id_token)
      }
      localStorage.setItem(USER_KEY, JSON.stringify(stored))
      if (this.expiresAt) {
        localStorage.setItem(EXPIRES_AT_KEY, String(this.expiresAt))
      }
    },
    clearSession() {
      this.token = null
      this.idToken = null
      this.user = null
      this.expiresAt = null
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(ID_TOKEN_KEY)
      localStorage.removeItem(USER_KEY)
      localStorage.removeItem(EXPIRES_AT_KEY)
    },
    async login() {
      await userManager.signinRedirect()
    },
    async handleLoginCallback(url?: string) {
      const user = await userManager.signinRedirectCallback(url)
      this.setUserSession(user)
      return user
    },
    async logout() {
      try {
        await userManager.signoutRedirect({
          id_token_hint: this.idToken || undefined,
          post_logout_redirect_uri: window.location.origin + '/login',
        })
      } catch {
        await userManager.removeUser()
        this.clearSession()
      }
    },
    async restoreFromStorage() {
      const token = localStorage.getItem(TOKEN_KEY)
      const userJson = localStorage.getItem(USER_KEY)
      const expiresAt = localStorage.getItem(EXPIRES_AT_KEY)
      const idToken = localStorage.getItem(ID_TOKEN_KEY)

      if (!token) {
        this.clearSession()
        return
      }

      if (expiresAt && Number(expiresAt) < Date.now()) {
        this.clearSession()
        return
      }

      this.token = token
      this.idToken = idToken
      this.expiresAt = expiresAt ? Number(expiresAt) : null

      if (userJson) {
        try {
          this.user = JSON.parse(userJson) as StoredUser
        } catch {
          this.user = null
        }
      }
    },
    init() {
      this.restoreFromStorage()
    },
  },
})
