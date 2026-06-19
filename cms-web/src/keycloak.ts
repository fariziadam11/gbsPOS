import { UserManager, WebStorageStateStore, type User } from 'oidc-client-ts'

const baseUrl = import.meta.env.VITE_KEYCLOAK_BASE_URL || 'https://auth.armmada.id'
const realm = import.meta.env.VITE_KEYCLOAK_REALM || 'gbs-pos'
export const clientId = import.meta.env.VITE_KEYCLOAK_CLIENT_ID || 'gbs-cms-web'

export const keycloakAuthority = `${baseUrl}/realms/${realm}`

export const userManager = new UserManager({
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

export function getKeycloakLoginUrl(): string {
  return keycloakAuthority + '/protocol/openid-connect/auth'
}

export function getKeycloakLogoutUrl(): string {
  return keycloakAuthority + '/protocol/openid-connect/logout'
}

export function extractRoles(user: User | null): string[] {
  if (!user?.access_token) return []
  try {
    const payload = JSON.parse(atob(user.access_token.split('.')[1]))
    const realmAccess = payload.realm_access as { roles?: string[] } | undefined
    return realmAccess?.roles ?? []
  } catch {
    return []
  }
}

export function parseTokenExpiry(token: string | null): number | null {
  if (!token) return null
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.exp ? payload.exp * 1000 : null
  } catch {
    return null
  }
}
