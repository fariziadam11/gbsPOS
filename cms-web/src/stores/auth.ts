import { defineStore } from 'pinia'
import {
  authMode,
  clientId,
  extractRoles,
  getKeycloakLogoutUrl,
  getUserManager,
  isKeycloakEnabled,
  parseTokenExpiry,
} from '../keycloak'
import { login as basicAuthLogin } from '../api/auth'
import type { LoginRequest, LoginResponse } from '../types/api'
import type { User } from 'oidc-client-ts'

function storeKeycloakUser(user: User) {
  const stored = parseUserFromKeycloak(user)
  localStorage.setItem(TOKEN_KEY, user.access_token)
  if (user.id_token) {
    localStorage.setItem(ID_TOKEN_KEY, user.id_token)
  } else {
    localStorage.removeItem(ID_TOKEN_KEY)
  }
  localStorage.setItem(USER_KEY, JSON.stringify(stored))
  const expiry = parseTokenExpiry(user.access_token)
  if (expiry) {
    localStorage.setItem(EXPIRES_AT_KEY, String(expiry))
  } else {
    localStorage.removeItem(EXPIRES_AT_KEY)
  }
  return { token: user.access_token, idToken: user.id_token ?? null, user: stored, expiresAt: expiry }
}

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

function storeBasicUser(result: LoginResponse) {
  const stored: StoredUser = {
    username: result.user.username,
    name: result.user.name,
    role: result.user.role,
  }
  const expiry = parseTokenExpiry(result.token)

  localStorage.setItem(TOKEN_KEY, result.token)
  localStorage.removeItem(ID_TOKEN_KEY)
  localStorage.setItem(USER_KEY, JSON.stringify(stored))
  if (expiry) {
    localStorage.setItem(EXPIRES_AT_KEY, String(expiry))
  } else {
    localStorage.removeItem(EXPIRES_AT_KEY)
  }

  return { token: result.token, idToken: null, user: stored, expiresAt: expiry }
}

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
    username: (state) => state.user?.username || state.user?.name || 'User',
  },
  actions: {
    setUserSession(user: User) {
      const session = storeKeycloakUser(user)
      this.token = session.token
      this.idToken = session.idToken
      this.user = session.user
      this.expiresAt = session.expiresAt
    },
    setBasicSession(result: LoginResponse) {
      const session = storeBasicUser(result)
      this.token = session.token
      this.idToken = session.idToken
      this.user = session.user
      this.expiresAt = session.expiresAt
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
    async login(credentials?: LoginRequest) {
      if (authMode === 'keycloak') {
        await getUserManager().signinRedirect()
        return
      }

      if (!credentials) {
        throw new Error('Username and password are required')
      }

      const response = await basicAuthLogin(credentials)
      this.setBasicSession(response.data)
    },
    async handleLoginCallback(url?: string) {
      if (!isKeycloakEnabled) {
        throw new Error('Keycloak authentication is not configured')
      }

      const user = await getUserManager().signinRedirectCallback(url)
      this.setUserSession(user)
      return user
    },
    async logout() {
      if (authMode === 'basic') {
        this.clearSession()
        window.location.href = '/login'
        return
      }

      const idToken = this.idToken
      this.clearSession()
      try {
        await getUserManager().removeUser()
      } catch {
        // ignore
      }

      // Redirect to Keycloak end-session endpoint manually so we control
      // the exact URL and always clear local state first.
      const logoutUrl = new URL(getKeycloakLogoutUrl())
      logoutUrl.searchParams.set('post_logout_redirect_uri', window.location.origin + '/login')
      logoutUrl.searchParams.set('client_id', clientId)
      if (idToken) {
        logoutUrl.searchParams.set('id_token_hint', idToken)
      }
      window.location.href = logoutUrl.toString()
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
    startTokenSync() {
      if (!isKeycloakEnabled) {
        return
      }

      const manager = getUserManager()
      manager.events.addUserLoaded((user) => {
        this.setUserSession(user)
      })
      manager.events.addUserUnloaded(() => {
        this.clearSession()
      })
      manager.events.addAccessTokenExpired(() => {
        this.clearSession()
      })
      manager.events.addSilentRenewError(() => {
        this.clearSession()
      })
    },
    init() {
      this.restoreFromStorage()
      this.startTokenSync()
    },
  },
})
