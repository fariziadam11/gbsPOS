import { defineStore } from 'pinia'
import { userManager, extractRoles, parseTokenExpiry, getKeycloakLogoutUrl, clientId } from '../keycloak'
import type { User } from 'oidc-client-ts'

function storeUser(user: User) {
  const stored = parseUserFromKeycloak(user)
  localStorage.setItem(TOKEN_KEY, user.access_token)
  if (user.id_token) {
    localStorage.setItem(ID_TOKEN_KEY, user.id_token)
  }
  localStorage.setItem(USER_KEY, JSON.stringify(stored))
  const expiry = parseTokenExpiry(user.access_token)
  if (expiry) {
    localStorage.setItem(EXPIRES_AT_KEY, String(expiry))
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
      const session = storeUser(user)
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
    async login() {
      await userManager.signinRedirect()
    },
    async handleLoginCallback(url?: string) {
      const user = await userManager.signinRedirectCallback(url)
      this.setUserSession(user)
      return user
    },
    async logout() {
      const idToken = this.idToken
      this.clearSession()
      try {
        await userManager.removeUser()
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
      userManager.events.addUserLoaded((user) => {
        this.setUserSession(user)
      })
      userManager.events.addUserUnloaded(() => {
        this.clearSession()
      })
      userManager.events.addAccessTokenExpired(() => {
        this.clearSession()
      })
      userManager.events.addSilentRenewError(() => {
        this.clearSession()
      })
    },
    init() {
      this.restoreFromStorage()
      this.startTokenSync()
    },
  },
})
