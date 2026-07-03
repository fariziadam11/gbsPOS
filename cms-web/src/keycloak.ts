import { UserManager, WebStorageStateStore, type User } from 'oidc-client-ts'

const baseUrl = String(import.meta.env.VITE_KEYCLOAK_BASE_URL || '').trim()
const realm = String(import.meta.env.VITE_KEYCLOAK_REALM || '').trim()
export const clientId = String(import.meta.env.VITE_KEYCLOAK_CLIENT_ID || '').trim()

export type AuthMode = 'keycloak' | 'basic'

export const isKeycloakEnabled = baseUrl !== '' && realm !== '' && clientId !== ''
export const authMode: AuthMode = isKeycloakEnabled ? 'keycloak' : 'basic'
export const keycloakAuthority = isKeycloakEnabled
  ? `${baseUrl.replace(/\/+$/, '')}/realms/${realm}`
  : ''

let userManager: UserManager | null = null

export function getUserManager(): UserManager {
  if (!isKeycloakEnabled) {
    throw new Error('Keycloak authentication is not configured')
  }

  if (!userManager) {
    userManager = new UserManager({
      authority: keycloakAuthority,
      client_id: clientId,
      redirect_uri: `${window.location.origin}/auth/callback`,
      response_type: 'code',
      scope: 'openid profile email',
      loadUserInfo: true,
      automaticSilentRenew: true,
      silent_redirect_uri: `${window.location.origin}/auth/callback`,
      userStore: new WebStorageStateStore({ store: window.localStorage }),
    })
  }

  return userManager
}

export function getKeycloakLoginUrl(): string {
  if (!isKeycloakEnabled) {
    throw new Error('Keycloak authentication is not configured')
  }
  return keycloakAuthority + '/protocol/openid-connect/auth'
}

export function getKeycloakLogoutUrl(): string {
  if (!isKeycloakEnabled) {
    throw new Error('Keycloak authentication is not configured')
  }
  return keycloakAuthority + '/protocol/openid-connect/logout'
}

function decodeTokenPayload(token: string): Record<string, unknown> | null {
  const payload = token.split('.')[1]
  if (!payload) return null

  try {
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')
    return JSON.parse(atob(padded)) as Record<string, unknown>
  } catch {
    return null
  }
}

export function extractRoles(user: User | null): string[] {
  if (!user?.access_token) return []
  const payload = decodeTokenPayload(user.access_token)
  const realmAccess = payload?.realm_access as { roles?: string[] } | undefined
  return realmAccess?.roles ?? []
}

export function parseTokenExpiry(token: string | null): number | null {
  if (!token) return null
  const payload = decodeTokenPayload(token)
  return typeof payload?.exp === 'number' ? payload.exp * 1000 : null
}
