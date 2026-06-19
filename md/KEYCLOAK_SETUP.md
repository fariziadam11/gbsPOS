# Keycloak Authentication Setup for GBS POS

This guide explains how to set up Keycloak for the GBS POS Android app and backend APIs, and the planned integration approach.

---

## 1. Run Keycloak with Docker

### Option A — Dev mode (quick testing, H2 database)

Use `start-dev` when testing. It allows HTTP and uses an embedded H2 database.

```bash
docker run -d \
  --name keycloak \
  -p 8080:8080 \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=<STRONG_PASSWORD> \
  quay.io/keycloak/keycloak:26.1 \
  start-dev
```

Open the admin console:

```text
http://<your-server-ip>:8080/admin
```

### Option B — Production mode behind Cloudflare Tunnel (recommended)

In production mode, Keycloak requires an external database such as PostgreSQL. It also expects HTTPS, but because Cloudflare terminates TLS, you tell Keycloak to trust the reverse-proxy headers and allow HTTP on port 8080.

```bash
docker run -d \
  --name keycloak \
  --restart unless-stopped \
  -p 127.0.0.1:8083:8080 \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=<STRONG_PASSWORD> \
  -e KC_DB=postgres \
  -e KC_DB_URL_HOST=postgres \
  -e KC_DB_URL_DATABASE=keycloak \
  -e KC_DB_USERNAME=keycloak \
  -e KC_DB_PASSWORD=<DB_PASSWORD> \
  -e KC_HOSTNAME=https://auth.armmada.id \
  -e KC_PROXY_HEADERS=xforwarded \
  -e KC_HTTP_ENABLED=true \
  quay.io/keycloak/keycloak:26.1 \
  start
```

Important points:
- Replace `<STRONG_PASSWORD>` and `<DB_PASSWORD>`.
- Use `127.0.0.1:8083` instead of `8083` so Keycloak is not reachable directly over HTTP from the internet; only the Cloudflare Tunnel should access it.
- `KC_HOSTNAME=https://auth.armmada.id` tells Keycloak its public URL.
- `KC_PROXY_HEADERS=xforwarded` tells Keycloak to read `X-Forwarded-Proto` from Cloudflare so it knows the connection is HTTPS on the public side.
- `KC_HTTP_ENABLED=true` keeps port 8080 open for the tunnel.
- PostgreSQL must already exist and the `keycloak` database must be created.

---

## 2. Create a Realm

1. In the admin console, click the current realm name at the top-left.
2. Click **Create realm**.
3. Enter a realm name, for example: `gbs-pos`.
4. Click **Create**.

---

## 3. Create Realm Roles

1. Go to **Realm roles** in the left menu.
2. Click **Create role**.
3. Add the following roles:
   - `ADMIN`
   - `CASHIER`

These names match the roles already used by the POS backend and Android app.

---

## 4. Create Users

1. Go to **Users** → **Add user**.
2. Fill in the required fields:
   - **Username**
   - **Email** (optional)
   - **First name** / **Last name** (optional)
3. Click **Create**.
4. Go to the **Credentials** tab → **Set password**.
   - Enter a password.
   - Disable **Temporary** so the user does not have to change it on first login.
   - Click **Save**.
5. Go to the **Role mapping** tab → **Assign role**.
   - Assign `ADMIN` or `CASHIER`.

Repeat for every POS user.

---

## 5. Create the Android Client

1. Go to **Clients** → **Create client**.
2. Configure the client:
   - **Client ID**: `gbs-pos-android`
   - **Root URL**: leave empty (not used by native apps)
   - **Home URL**: leave empty, or set a support page if you want a link in the admin console
   - **Client authentication**: OFF (public client)
   - **Authentication flow**:
     - Enable **Standard flow** (Authorization Code with PKCE)
     - Optionally enable **Direct access grants** if you want a username/password fallback later.
3. Click **Save**.
4. On the client settings page, configure:
   - **Valid redirect URIs**: `com.gis.gbs.posapplication:/oauth2callback`
   - **Valid post logout redirect URIs**: `com.gis.gbs.posapplication:/logoutcallback`
   - **Web origins**: `*`
5. Click **Save**.

The redirect URI must match the Android intent filter used by AppAuth.

---

## 6. Create the CMS Web Client

1. Go to **Clients** → **Create client**.
2. Configure the client:
   - **Client ID**: `gbs-cms-web`
   - **Root URL**: leave empty
   - **Home URL**: leave empty
   - **Client authentication**: OFF (public client)
   - **Authentication flow**:
     - Enable **Standard flow** (Authorization Code with PKCE)
3. Click **Save**.
4. On the client settings page, configure:
   - **Valid redirect URIs**:
     - `https://cms.armmada.id/auth/callback`
     - `http://localhost:5173/auth/callback`
   - **Valid post logout redirect URIs**:
     - `https://cms.armmada.id/login`
     - `http://localhost:5173/login`
   - **Web origins**:
     - `https://cms.armmada.id`
     - `http://localhost:5173`
5. Click **Save**.

> If your CMS web origin is different, replace `cms.armmada.id` and `localhost:5173` accordingly.

---

## 7. Values Required by the Android App

After setup, you will have the following values:

| Value | Example | Description |
|-------|---------|-------------|
| `KEYCLOAK_BASE_URL` | `https://auth.armmada.id` or `http://<ip>:8080` | Keycloak server URL |
| `KEYCLOAK_REALM` | `gbs-pos` | Realm name |
| `KEYCLOAK_CLIENT_ID` | `gbs-pos-android` | Client ID |
| Redirect URI | `com.gis.gbs.posapplication:/oauth2callback` | Android deep link used by AppAuth |

Store these in `local.properties`:

```properties
KEYCLOAK_BASE_URL=https://auth.armmada.id
KEYCLOAK_REALM=gbs-pos
KEYCLOAK_CLIENT_ID=gbs-pos-android
```

These values will be exposed to the app through `BuildConfig`.

---

## 8. Offline Mode with Keycloak

Keycloak supports **offline tokens**, which allow a client to refresh access tokens even when the user's session has ended or the device has been offline for a while.

### Recommended offline strategy for POS devices

1. During login, request the `offline_access` scope so Keycloak returns an offline token.
2. Store locally in DataStore:
   - Access token
   - Refresh / offline token
   - Token expiry time
   - Username and role
3. While online:
   - Silently refresh the access token before it expires.
4. While offline:
   - Continue using the cached access token until it expires.
   - Allow local-only operations using the Room database.
   - Block server-dependent actions until the device is back online and the token is refreshed.

Offline tokens are useful for POS devices that may restart or lose network access, but sensitive operations should still require a valid access token.

---

## 9. Planned Integration

### Android App

1. Add the AppAuth dependency:
   - `net.openid:appauth-android`
2. Expose Keycloak config via `BuildConfig` from `local.properties`.
3. Replace the current username/password login with Authorization Code + PKCE using AppAuth:
   - User taps **Login with Keycloak**.
   - Chrome Custom Tab opens the Keycloak login page.
   - App receives an authorization code via the redirect URI.
   - App exchanges the code for access and refresh tokens.
   - Parse the ID token or access token to extract username and role.
4. Update `AuthDataStore` to persist:
   - Access token
   - Refresh / offline token
   - Expiry timestamp
   - Username
   - Role
5. Update `AuthInterceptor` to:
   - Attach `Authorization: Bearer <accessToken>`.
   - On `401`, attempt a silent refresh.
   - If refresh fails, clear the session and redirect to login.
6. Update `LoginScreen` to show a **Login with Keycloak** button and an optional **Demo / Offline Login** fallback.
7. Implement logout that clears local tokens and calls the Keycloak end-session endpoint.

### Backend (`gbs-pos-api` and `gbs-cms-api`)

1. Add Keycloak configuration:
   - Keycloak base URL
   - Realm
   - JWKS endpoint: `{base}/realms/{realm}/protocol/openid-connect/certs`
2. Replace the custom HS256 JWT middleware with middleware that validates Keycloak RS256 access tokens.
3. Extract the `role` claim from the token and set it in the Gin context.
4. Keep the existing `/v1/login` endpoint as a **fallback/demo** route, enabled only when a demo-auth feature flag is on.
   - This preserves the local `admin/admin` and `cashier/1234` accounts for development and emergencies.

### CMS Web

1. Install an OIDC client library (`oidc-client-ts`).
2. Add Keycloak environment variables:
   - `VITE_KEYCLOAK_BASE_URL`
   - `VITE_KEYCLOAK_REALM`
   - `VITE_KEYCLOAK_CLIENT_ID`
3. Replace the username/password login form with a redirect to Keycloak (`signinRedirect`).
4. Add an `/auth/callback` route to handle the authorization code exchange (`signinRedirectCallback`).
5. Store the Keycloak access token in `localStorage` and attach it as a Bearer token to CMS/POS API requests.
6. Parse the access token to extract the user's role (`ADMIN` / `CASHIER`).
7. Update logout to redirect to Keycloak's end-session endpoint.

---

## 10. Notes and Best Practices

- Always use HTTPS in production.
- Do not embed Keycloak secrets in the Android app because it is a public client.
- Use PKCE (Proof Key for Code Exchange) for the Authorization Code flow.
- Use `offline_access` scope only if the device needs long-lived offline access.
- Keep the demo login fallback behind a feature flag; disable it in production builds.
- Back up the Keycloak database if you run it in production.

---

## 11. Configure the Android App

Add the following to `local.properties` (do not commit this file):

```properties
keycloak.base.url=https://auth.armmada.id
keycloak.realm=gbs-pos
keycloak.client.id=gbs-pos-android
keycloak.redirect.uri=com.gis.gbs.posapplication:/oauth2callback
keycloak.redirect.scheme=com.gis.gbs.posapplication
enable.demo.auth=false
```

Then rebuild:

```bash
./gradlew assembleDebug
```

## 12. Configure the CMS Web

Add the following to `pos-cms/cms-web/.env` (do not commit this file):

```properties
VITE_API_BASE_URL=https://api-cms.armmada.id
VITE_POS_API_BASE_URL=https://api-pos.armmada.id

VITE_KEYCLOAK_BASE_URL=https://auth.armmada.id
VITE_KEYCLOAK_REALM=gbs-pos
VITE_KEYCLOAK_CLIENT_ID=gbs-cms-web
```

Then rebuild:

```bash
cd pos-cms/cms-web
bun install
bun run build
```

## 13. Configure the Backend

Set environment variables for `gbs-pos-api` and `gbs-cms-api`:

```bash
KEYCLOAK_BASE_URL=https://auth.armmada.id
KEYCLOAK_REALM=gbs-pos
# Optional: keep /v1/login for local demo accounts
ENABLE_DEMO_AUTH=false
JWT_SECRET=<only required when ENABLE_DEMO_AUTH=true>
```

When `KEYCLOAK_BASE_URL` and `KEYCLOAK_REALM` are set, the backend validates RS256 Keycloak access tokens via JWKS and falls back to the legacy HS256 middleware only when `ENABLE_DEMO_AUTH=true`.

## 14. Next Steps

1. Ensure Keycloak is running and reachable at `https://auth.armmada.id`.
2. Verify the `gbs-pos-android` client redirect URI matches `com.gis.gbs.posapplication:/oauth2callback`.
3. Verify the `gbs-cms-web` client redirect URIs include `https://cms.armmada.id/auth/callback` and `http://localhost:5173/auth/callback`.
4. Add the Keycloak config to `local.properties`, backend env files, and `pos-cms/cms-web/.env`.
5. Build and deploy Android, backend, and CMS web.
6. Test login from the Android app and the CMS web.
