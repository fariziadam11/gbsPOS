# Dual Authentication - JWT & Keycloak

Dokumentasi ini menjelaskan bagaimana mengkonfigurasi sistem untuk mendukung dual authentication (JWT dan Keycloak) secara bersamaan.

## Arsitektur Authentication

```
┌─────────────────────────────────────────────────────────────────┐
│                        API Request                               │
│                   (Authorization: Bearer <token>)                │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Composite Auth Middleware                       │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  1. Parse token header untuk detect algorithm            │  │
│  │     - RS256 → Keycloak token                             │  │
│  │     - HS256 → JWT token (legacy/local)                   │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
            ┌───────────────────┴───────────────────┐
            │                                       │
            ▼                                       ▼
┌───────────────────────────┐       ┌───────────────────────────┐
│    JWT Validation         │       │   Keycloak Validation     │
│    (HS256 with secret)    │       │   (RS256 with JWKS)       │
│                           │       │                           │
│  - Validate signature     │       │  - Validate signature     │
│  - Check expiration       │       │    against Keycloak JWKS   │
│  - Extract claims         │       │  - Check expiration       │
│  - Set user context       │       │  - Extract realm_access    │
│                           │       │    roles (ADMIN/CASHIER) │
│                           │       │  - Set user context       │
└───────────────────────────┘       └───────────────────────────┘
            │                                       │
            └───────────────────┬───────────────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │   Role-based Access  │
                    │   - ADMIN            │
                    │   - CASHIER          │
                    └───────────────────────┘
```

## Environment Variables

### Required Configuration

```bash
# ===================
# Dual Auth Configuration
# ===================

# JWT Secret (for local JWT tokens - HS256)
# Wajib min 32 karakter
JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters

# Keycloak Configuration (for Keycloak tokens - RS256)
KEYCLOAK_BASE_URL=https://auth.armmada.id
KEYCLOAK_REALM=gbs-pos

# Dual Auth Mode - accept BOTH JWT and Keycloak tokens
ENABLE_DEMO_AUTH=true

# ===================
# Server Configuration
# ===================
PORT=8080  # POS API
PORT=8081  # CMS API
```

### Configuration Modes

| Mode | JWT_SECRET | KEYCLOAK_* | ENABLE_DEMO_AUTH | Behavior |
|------|------------|------------|------------------|----------|
| JWT Only | ✅ Set | ❌ Empty | Any | Accepts HS256 tokens only |
| Keycloak Only | Any | ✅ Set | `false` | Accepts RS256 tokens only |
| **Dual Auth** | ✅ Set | ✅ Set | `true` | Accepts both HS256 and RS256 |

## Obtaining Tokens

### Option 1: Local JWT Token (Demo Mode)

```bash
# Login via API
curl -X POST https://api-pos.armmada.id/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

Response:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "username": "admin",
      "name": "Admin User",
      "role": "ADMIN"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### Option 2: Keycloak Token (Production)

```bash
# Get token from Keycloak
curl -X POST https://auth.armmada.id/realms/gbs-pos/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=gbs-pos" \
  -d "username=admin" \
  -d "password=admin123"
```

Response:
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 300,
  "refresh_token": "...",
  "scope": "openid profile email"
}
```

## Using Tokens

### JWT Token (HS256)
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Keycloak Token (RS256)
```http
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

## JWT Token Claims (Local)

```json
{
  "sub": "1",
  "username": "admin",
  "role": "ADMIN",
  "exp": 1735689600,
  "iat": 1735603200
}
```

## Keycloak Token Claims

```json
{
  "sub": "user-uuid-from-keycloak",
  "preferred_username": "admin",
  "email": "admin@example.com",
  "name": "Admin User",
  "realm_access": {
    "roles": ["ADMIN", "USER"]
  },
  "azp": "gbs-pos",
  "exp": 1735689600,
  "iat": 1735603200
}
```

## Roles

| Role | Access Level |
|------|-------------|
| **ADMIN** | Full access to all endpoints |
| **CASHIER** | Limited to POS operations |

## Error Responses

### 401 Unauthorized

```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing authorization header"
  }
}
```

### 401 Invalid Token (JWT)

```json
{
  "success": false,
  "error": {
    "code": "INVALID_TOKEN",
    "message": "Invalid or expired token"
  }
}
```

### 401 Invalid Token (Keycloak)

```json
{
  "success": false,
  "error": {
    "code": "INVALID_TOKEN",
    "message": "Invalid or expired Keycloak token"
  }
}
```

### 403 Forbidden (Insufficient Role)

```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_PERMISSIONS",
    "message": "You don't have permission to access this resource"
  }
}
```

## Keycloak Realm Configuration

### Required Roles

Create these roles in Keycloak realm:

1. **ADMIN** - Administrator role
2. **CASHIER** - Cashier role

### Role Assignment

Assign roles to users via:
1. Keycloak Admin Console → Users → Select User → Role Mappings
2. Or via Keycloak API

### Client Configuration

For client `gbs-pos` or `gbs-cms`:

1. Set Access Type: `confidential` or `public`
2. Enable Standard Flow (for web apps)
3. Enable Direct Access Grants (for API auth)
4. Configure Valid Redirect URIs

## Production Checklist

- [ ] Set secure `JWT_SECRET` (min 32 characters)
- [ ] Configure `KEYCLOAK_BASE_URL` and `KEYCLOAK_REALM`
- [ ] Set `ENABLE_DEMO_AUTH=true` for dual auth
- [ ] Create ADMIN and CASHIER roles in Keycloak
- [ ] Assign roles to Keycloak users
- [ ] Update API host in Swagger docs
- [ ] Test token flow with both JWT and Keycloak
- [ ] Set `LOG_LEVEL=info` in production

## Docker Compose Configuration

```yaml
services:
  gbs-pos-api:
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - KEYCLOAK_BASE_URL=${KEYCLOAK_BASE_URL}
      - KEYCLOAK_REALM=${KEYCLOAK_REALM}
      - ENABLE_DEMO_AUTH=true
      - ENV=production
```

## Troubleshooting

### Token Expired
```
"Invalid or expired token"
```
Solution: Get a new token using the login endpoint

### Keycloak JWKS Error
```
"failed to create JWKS keyfunc"
```
Solution: Verify `KEYCLOAK_BASE_URL` is correct and accessible

### Role Not Found
```
"No ADMIN/CASHIER role found in Keycloak token"
```
Solution: Assign role to user in Keycloak Admin Console

### Algorithm Mismatch
```
"Unsupported token algorithm: none"
```
Solution: Token must be signed (RS256 or HS256)
