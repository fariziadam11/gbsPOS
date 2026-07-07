# Swagger API Documentation - POS API

## Accessing Swagger UI

Once the POS API is running, access Swagger UI at:
```
http://localhost:8080/swagger/index.html
```

## Authentication for Testing

### Option 1: Demo Login (Local JWT)

If `ENABLE_DEMO_AUTH=true`, use the login endpoint to get a token:

```bash
curl -X POST http://localhost:8080/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

Response:
```json
{
  "success": true,
  "data": {
    "user": { "id": 1, "username": "admin", "name": "Admin User", "role": "ADMIN" },
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
  -d "client_id=gbs-pos-android" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "username=admin" \
  -d "password=YOUR_PASSWORD"
```

#### Step 2: Use Token in Swagger UI

Click **Authorize** and enter:
```
Bearer <access_token>
```

## API Endpoints

### Public Endpoints (No Auth Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/ads/active?terminalId=xxx` | Get active ads playlist |
| GET | `/v1/ads/download/:id` | Download ad video |
| GET | `/v1/display/cart?terminalId=xxx` | Get cart display JSON |

### Protected Endpoints (Bearer Token Required)

#### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/login` | Login and get JWT token |

#### Products
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/products` | List products |
| GET | `/v1/products/low-stock` | Low stock products |
| GET | `/v1/products/:id` | Get product by ID |
| POST | `/v1/products` | Create product (ADMIN) |
| PUT | `/v1/products/:id` | Update product (ADMIN) |
| DELETE | `/v1/products/:id` | Delete product (ADMIN) |
| POST | `/v1/products/:id/stock` | Adjust stock |
| GET | `/v1/products/:id/history` | Stock history |
| POST | `/v1/products/import` | Import CSV (ADMIN) |
| GET | `/v1/products/export` | Export CSV (ADMIN) |

#### Orders
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/orders` | List orders |
| GET | `/v1/orders/:id` | Get order by ID |
| POST | `/v1/orders` | Create order |
| POST | `/v1/orders/sync` | Bulk sync orders |
| PATCH | `/v1/orders/:id/void` | Void order (ADMIN) |
| GET | `/v1/orders/unsettled/summary` | Unsettled orders |
| POST | `/v1/orders/settle` | Settle orders (ADMIN) |

#### Settlements
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/settlements` | List settlements |
| GET | `/v1/settlements/:id` | Get settlement |

#### Customers
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/customers` | List customers |
| GET | `/v1/customers/:id` | Get customer |
| GET | `/v1/customers/:id/orders` | Customer orders |
| POST | `/v1/customers` | Create customer |
| PATCH | `/v1/customers/:id` | Update customer |
| GET | `/v1/customers/phone/:phone` | Get by phone |

#### Discounts
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/discounts` | List discounts |
| POST | `/v1/discounts` | Create discount (ADMIN) |
| POST | `/v1/discounts/calculate` | Calculate price |
| PATCH | `/v1/discounts/:id` | Update discount (ADMIN) |
| POST | `/v1/discounts/:id/stop` | Stop discount (ADMIN) |
| DELETE | `/v1/discounts/:id` | Delete discount (ADMIN) |

#### Variants
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/products/:id/variants` | List variants |
| POST | `/v1/products/:id/variants` | Create variant (ADMIN) |
| PATCH | `/v1/products/:id/variants/:id` | Update variant (ADMIN) |
| DELETE | `/v1/products/:id/variants/:id` | Delete variant (ADMIN) |

#### Holds
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/holds` | List hold sessions |
| GET | `/v1/holds/:id` | Get hold session |
| POST | `/v1/holds` | Create hold |
| PUT | `/v1/holds/:id/resume` | Resume hold |
| DELETE | `/v1/holds/:id` | Delete hold |

#### Dashboard
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/dashboard/summary` | Dashboard summary |
| GET | `/v1/dashboard/revenue` | Revenue trend |
| GET | `/v1/dashboard/top-products` | Top products |

#### Fuel (POS devices only)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/fuel/prices` | List fuel prices |
| PATCH | `/v1/fuel/prices/:code` | Update price |
| GET | `/v1/fuel/pumps` | List pumps |
| POST | `/v1/fuel/pumps` | Create pump |
| PATCH | `/v1/fuel/pumps/:id` | Update pump |
| DELETE | `/v1/fuel/pumps/:id` | Delete pump |
| GET | `/v1/fuel/nozzles` | List nozzles |
| POST | `/v1/fuel/nozzles` | Create nozzle |
| GET | `/v1/fuel/sales` | Sales report |
| POST | `/v1/fuel/sales` | Record sale |

#### Cart Display
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/display/cart` | Upload cart display JSON (protected) |
| GET | `/v1/display/cart` | Get cart display JSON (public) |
| DELETE | `/v1/display/cart/:terminalId` | Clear display |

## Response Format

All protected endpoints return:
```json
{
  "success": true,
  "data": { ... },
  "error": null
}
```

## Notes

- Swagger UI supports "Try it out" for all endpoints
- Use the **Authorize** button to set Bearer token globally
- Tokens expire - re-authenticate when needed
- POS API runs on port 8080
- CMS API runs on port 8081
