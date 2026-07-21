# GBS POS & CMS API - Postman Collections

This folder contains Postman collections for testing the GBS POS and CMS APIs with dual authentication support (JWT and Keycloak).

## Files

- `GBS_POS_API.postman_collection.json` - Collection for POS API endpoints
- `GBS_CMS_API.postman_collection.json` - Collection for CMS API endpoints
- `GBS_API_Environment.postman_environment.json` - Environment with API URLs and variables

## Quick Start

### 1. Import Collections

1. Open Postman
2. Click **Import** button
3. Select all JSON files from this folder

### 2. Configure Environment

1. Go to **Environments** in Postman
2. Select "GBS APIs - Development"
3. Update the following variables:
   - `posBaseUrl`: `https://api-pos.armmada.id/v1`
   - `cmsBaseUrl`: `https://api-cms.armmada.id/v1`
   - `keycloakBaseUrl`: Your Keycloak server URL (e.g., `https://auth.armmada.id`)
   - `keycloakRealm`: Your Keycloak realm name (e.g., `gbs-pos`)

### 3. Authentication

#### Option A: Using Local JWT (Demo Mode)

When `ENABLE_DEMO_AUTH=true`:

1. Open the **Authentication** folder in either collection
2. Run **Login (JWT)** request
3. The token will be automatically saved to `jwtToken` variable
4. Use the collection for authenticated requests

#### Option B: Using Keycloak

1. Open the **Authentication** folder in either collection
2. Run **Login (Keycloak - Get Token)** request
3. The token will be automatically saved to `keycloakToken` variable
4. Update the collection auth to use the Keycloak token

#### Option C: Dual Auth Mode

When both Keycloak and JWT are configured (`ENABLE_DEMO_AUTH=true` and Keycloak is set):

1. You can use either JWT or Keycloak tokens
2. The API automatically detects the token type based on the algorithm:
   - `RS256` → Keycloak token
   - `HS256` → Local JWT token

## API Endpoints Overview

### POS API (https://api-pos.armmada.id/v1)

| Category | Endpoints |
|----------|-----------|
| **Health** | GET /health |
| **Authentication** | POST /login |
| **Products** | GET/POST /products, GET/PUT/DELETE /products/:id |
| **Products - Stock** | POST /products/:id/stock, GET /products/low-stock |
| **Discounts** | GET/POST /discounts, POST /discounts/calculate |
| **Orders** | GET/POST /orders, GET /orders/:id, PATCH /orders/:id/void |
| **Orders - Sync** | POST /sync/orders, GET /orders/unsettled/summary, POST /orders/settle |
| **Settlements** | GET /settlements, GET /settlements/:id |
| **Customers** | GET/POST /customers, GET /customers/phone/:phone |
| **Dashboard** | GET /dashboard/summary, /revenue, /top-products |
| **Fuel** | GET /fuel/prices, PATCH /fuel/prices/:code, GET/POST /fuel/pumps |
| **Fuel - Sales** | POST /fuel/sales, GET /fuel/report |
| **QRIS** | POST /qris/payments, GET /qris/payments/:orderId/status |
| **Hold Sessions** | GET/POST /holds, PUT /holds/:id/resume, DELETE /holds/:id |

### CMS API (https://api-cms.armmada.id/v1)

| Category | Endpoints |
|----------|-----------|
| **Health** | GET /health |
| **Authentication** | POST /login |
| **Ads** | GET /ads, POST /ads/upload, GET/PUT/DELETE /ads/:id |
| **Ads - Playlist** | GET /ads/active, GET /ads/download/:id, POST /ads/:id/play |
| **Ads - Toggle** | POST /ads/:id/toggle |
| **Users** | GET/POST /users, GET/PUT/DELETE /users/:id |
| **Settings** | GET/PUT /settings |
| **Cart Display** | POST /display/cart, GET /display/cart, GET /display/terminals |

## Authentication Flow

### JWT Token (Local Auth)

```http
POST /v1/login
Content-Type: application/json

{
    "username": "admin",
    "password": "admin123"
}
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

Use the token:
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Keycloak Token (OIDC)

```http
POST /v1/auth/realms/{realm}/protocol/openid-connect/token
Content-Type: application/x-www-form-urlencoded

grant_type=password&client_id=gbs-pos&username=admin&password=admin123
```

Response:
```json
{
    "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 300,
    ...
}
```

Use the token:
```http
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Role-Based Access

| Role | Access |
|------|--------|
| **ADMIN** | Full access to all endpoints |
| **CASHIER** | Limited to POS operations (products, orders, etc.) |

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

### 403 Forbidden
```json
{
    "success": false,
    "error": {
        "code": "INSUFFICIENT_PERMISSIONS",
        "message": "You don't have permission to access this resource"
    }
}
```

### 404 Not Found
```json
{
    "success": false,
    "error": {
        "code": "PRODUCT_NOT_FOUND",
        "message": "Product not found"
    }
}
```

## Common Test Scenarios

### 1. Complete Sale Flow

1. Login to get token
2. List products → Select product by ID
3. Create order with items
4. View order confirmation

### 2. Admin Product Management

1. Login as admin
2. Create new product
3. Adjust stock (add inventory)
4. Update product price
5. Verify product appears in list

### 3. CMS Advertisement Management

1. Login to CMS API
2. Upload video advertisement
3. List all ads
4. Toggle ad status (active/inactive)
5. Get active playlist for store type

## Notes

- All timestamps are Unix milliseconds
- Store types: `RETAIL`, `FOOD`, `OUTFIT`
- Payment methods: `CASH`, `QRIS`, `DEBIT`, `CREDIT`
- Discount types: `PERCENT`, `FIXED`, `VOUCHER`
