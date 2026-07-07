# Swagger API Documentation

## Accessing Swagger UI

Once the CMS API is running, access Swagger UI at:
```
http://localhost:8081/swagger/index.html
```

## Authentication for Testing

### Option 1: Demo Login (Local JWT)

If `ENABLE_DEMO_AUTH=true`, use the login endpoint to get a token:

```bash
curl -X POST http://localhost:8081/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
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

Copy the `token` value and click the **Authorize** button in Swagger UI, then enter:
```
Bearer <your-token-here>
```

### Option 2: Keycloak Token

If using Keycloak authentication:

#### Step 1: Get Keycloak Token

```bash
# Replace with your Keycloak configuration
curl -X POST "https://auth.armmada.id/realms/gbs-pos/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=gbs-cms-web" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "username=admin" \
  -d "password=YOUR_PASSWORD"
```

Response:
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 300
}
```

#### Step 2: Use Token in Swagger UI

Copy the `access_token` value and click **Authorize** in Swagger UI:
```
Bearer <access_token-here>
```

## API Endpoints

### Public Endpoints (No Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/ads/active?storeType=RETAIL` | Get active ads playlist |
| GET | `/v1/ads/download/{id}` | Download ad video file |
| GET | `/v1/display/cart?terminalId=xxx` | Get cart display JSON |

### Protected Endpoints (Bearer Token Required)

#### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/login` | Login and get JWT token |

#### Ads Management (ADMIN only)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/ads` | List all ads |
| POST | `/v1/ads/upload` | Upload new ad |
| GET | `/v1/ads/{id}` | Get ad by ID |
| PUT | `/v1/ads/{id}` | Update ad |
| DELETE | `/v1/ads/{id}` | Delete ad |
| POST | `/v1/ads/{id}/toggle` | Toggle ad status |
| POST | `/v1/ads/{id}/play` | Log ad play |

#### Users Management (ADMIN only)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/users` | List all users |
| GET | `/v1/users/{id}` | Get user by ID |
| POST | `/v1/users` | Create user |
| PUT | `/v1/users/{id}` | Update user |
| DELETE | `/v1/users/{id}` | Delete user |

#### Settings (ADMIN only)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/settings` | Get all settings |
| PUT | `/v1/settings` | Update settings |

#### Cart Display
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/display/cart` | Upload cart display JSON (protected) |
| GET | `/v1/display/cart?terminalId=xxx` | Get cart display JSON (public) |

## Testing the Cart Display Endpoint

### From Android POS (Protected)

Android sends cart display data with Bearer token:

```bash
curl -X POST http://localhost:8081/v1/display/cart \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "terminalId": "android_device_id",
    "Initial": {
      "NamaKasir": "DOMAR",
      "KodeToko": "T14AB",
      "NamaToko": "Indomaret Pusat",
      "JenisToko": "POINT"
    },
    "DaftarBelanja": [
      {
        "Deskripsi": "INDOMIE,MIE INSTANT AYAM SPECIAL 68g PCK",
        "Harga": "3.100",
        "Qty": "2",
        "Total": "6.200"
      }
    ],
    "Summary": {
      "Hemat": "0",
      "Total": "6.200",
      "Bayar": "0",
      "Kembali": "0"
    },
    "TeksSelesai": "Transaksi"
  }'
```

### From Browser Display (Public)

Browser/customer display polls without authentication:

```bash
curl "http://localhost:8081/v1/display/cart?terminalId=android_device_id"
```

Response:
```json
{
  "Initial": {
    "NamaKasir": "DOMAR",
    "KodeToko": "T14AB",
    "NamaToko": "Indomaret Pusat",
    "JenisToko": "POINT"
  },
  "DaftarBelanja": [...],
  "Summary": {...},
  "TeksSelesai": "Transaksi"
}
```

## Response Format

All protected endpoints return:

```json
{
  "success": true,
  "data": { ... },
  "error": null
}
```

On error:
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message"
  }
}
```

## Role-Based Access

- **ADMIN**: Full access to all endpoints
- **CASHIER**: Read-only access to public endpoints

## Notes

- Swagger UI supports "Try it out" for all endpoints
- Use the **Authorize** button to set your Bearer token globally
- Tokens expire - re-authenticate when needed
- The cart display endpoint accepts raw JSON from Android
