# GBS POS — Backend API Guide (Golang)

This document is the **complete specification** for the Golang backend that powers the GBS POS Android application. The Android app is already built with dummy/local data and is ready to connect to a real API.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Tech Stack Recommendation](#2-tech-stack-recommendation)
3. [Project Structure](#3-project-structure)
4. [Database Schema](#4-database-schema)
5. [API Endpoints](#5-api-endpoints)
6. [Request/Response Examples](#6-requestresponse-examples)
7. [Authentication & Authorization](#7-authentication--authorization)
8. [Business Rules](#8-business-rules)
9. [Neurogine Card Payment Fields](#9-neurogine-card-payment-fields)
10. [Offline-First Sync Strategy](#10-offline-first-sync-strategy)
11. [Multi-Terminal Support](#11-multi-terminal-support)
12. [Error Response Format](#12-error-response-format)
13. [Seed Data](#13-seed-data)
14. [Docker Deployment](#14-docker-deployment)
15. [Environment Variables](#15-environment-variables)
16. [Testing Checklist](#16-testing-checklist)

---

## 1. Overview

The GBS POS Android app is a **Point of Sale** system for retail, F&B, and clothing stores. It runs on Sunmi D3 Pro dual-screen devices and supports:

- Product catalog with multi-store support (Retail, F&B, Outfit)
- Shopping cart with 10% PPN tax calculation
- Cash, Card (Neurogine SoftPOS), and QRIS payments
- Order history with void and settlement features
- Customer display on secondary screen
- USB thermal printer integration
- Offline-first architecture (works without internet)

**Current state**: The app uses Room (SQLite) for local storage and Retrofit for network calls. The base URL is currently `https://your-api.com/` (placeholder). Once the backend is ready, we will update the base URL and the app will sync data with your API.

**Key requirement**: The backend must support **idempotent operations** and **bulk sync** because the app works offline and queues data locally.

---

## 2. Tech Stack Recommendation

| Layer | Recommendation | Alternative |
|-------|---------------|-------------|
| Framework | **Gin** (`github.com/gin-gonic/gin`) | Echo, Fiber |
| Auth | JWT (`github.com/golang-jwt/jwt/v5`) | PASETO |
| Database | **PostgreSQL 15+** | MySQL 8+ |
| ORM | **GORM** or raw SQL with `pgx` | sqlx |
| Migration | `golang-migrate/migrate` | Goose |
| Config | `github.com/caarlos0/env/v10` | Viper |
| Logging | `github.com/rs/zerolog` | zap, slog |
| Validation | `github.com/go-playground/validator/v10` | — |
| Docker | Docker Compose (app + postgres) | Kubernetes |

---

## 3. Project Structure

```
gbs-pos-api/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Environment variables
│   ├── database/
│   │   └── database.go          # PostgreSQL connection + migration
│   ├── middleware/
│   │   ├── auth.go              # JWT verification
│   │   ├── logger.go            # Request logging
│   │   └── cors.go              # CORS headers
│   ├── handler/
│   │   ├── auth_handler.go      # POST /login
│   │   ├── product_handler.go   # CRUD /products
│   │   ├── order_handler.go     # CRUD /orders, void, settle
│   │   └── settlement_handler.go# GET /settlements
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── product_service.go
│   │   ├── order_service.go
│   │   └── settlement_service.go
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── product_repo.go
│   │   ├── order_repo.go
│   │   └── settlement_repo.go
│   └── model/
│       ├── user.go
│       ├── product.go
│       ├── order.go
│       └── settlement.go
├── migrations/
│   ├── 001_create_users.up.sql
│   ├── 001_create_users.down.sql
│   ├── 002_create_products.up.sql
│   ├── 002_create_products.down.sql
│   ├── 003_create_orders.up.sql
│   ├── 004_create_order_items.up.sql
│   ├── 005_create_settlements.up.sql
│   └── 006_seed_data.up.sql
├── pkg/
│   └── response/
│       └── response.go          # Standard API response wrapper
├── docker-compose.yml
├── Dockerfile
├── .env.example
├── go.mod
└── go.sum
```

---

## 4. Database Schema

### 4.1 `users`

```sql
CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    username      VARCHAR(50)  UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,  -- bcrypt
    name          VARCHAR(100),
    role          VARCHAR(20)  NOT NULL CHECK (role IN ('ADMIN', 'CASHIER')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_username ON users(username);
```

### 4.2 `products`

```sql
CREATE TABLE products (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(200)   NOT NULL,
    price       DECIMAL(12,2)  NOT NULL CHECK (price >= 0),
    category    VARCHAR(100)   NOT NULL,
    image_url   VARCHAR(500),
    store_type  VARCHAR(20)    NOT NULL CHECK (store_type IN ('RETAIL', 'FNB', 'OUTFIT')),
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_store_type ON products(store_type);
CREATE INDEX idx_products_category ON products(category);
```

### 4.3 `orders`

```sql
CREATE TABLE orders (
    id              VARCHAR(32)    PRIMARY KEY,  -- client-generated UUID8
    subtotal        DECIMAL(12,2)  NOT NULL,
    tax             DECIMAL(12,2)  NOT NULL,
    total           DECIMAL(12,2)  NOT NULL,
    payment_method  VARCHAR(20)    NOT NULL CHECK (payment_method IN ('CASH', 'CARD', 'QRIS')),
    cash_received   DECIMAL(12,2),
    change_amount   DECIMAL(12,2),
    timestamp       BIGINT         NOT NULL,     -- Unix milliseconds
    is_voided       BOOLEAN        NOT NULL DEFAULT FALSE,
    is_settled      BOOLEAN        NOT NULL DEFAULT FALSE,
    transaction_id  VARCHAR(100),                -- Neurogine
    approval_code   VARCHAR(50),                 -- Neurogine
    entry_mode      VARCHAR(20),                 -- Neurogine
    masked_account  VARCHAR(50),                 -- Neurogine
    acq_mid         VARCHAR(50),                 -- Neurogine
    acq_tid         VARCHAR(50),                 -- Neurogine
    pos_message_id  VARCHAR(100),                -- Neurogine
    bank_name       VARCHAR(50),                 -- e.g., "BCA"
    store_type      VARCHAR(20),                 -- RETAIL, FNB, OUTFIT
    terminal_id     VARCHAR(32),                 -- POS device identifier
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_timestamp ON orders(timestamp DESC);
CREATE INDEX idx_orders_is_settled ON orders(is_settled);
CREATE INDEX idx_orders_is_voided ON orders(is_voided);
CREATE INDEX idx_orders_store_type ON orders(store_type);
CREATE INDEX idx_orders_terminal_id ON orders(terminal_id);
CREATE INDEX idx_orders_transaction_id ON orders(transaction_id);
```

### 4.4 `order_items`

```sql
CREATE TABLE order_items (
    id             SERIAL PRIMARY KEY,
    order_id       VARCHAR(32)    NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id     INT            NOT NULL,
    product_name   VARCHAR(200)   NOT NULL,
    product_price  DECIMAL(12,2)  NOT NULL,
    qty            INT            NOT NULL CHECK (qty > 0),
    subtotal       DECIMAL(12,2)  NOT NULL
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
```

### 4.5 `settlements`

```sql
CREATE TABLE settlements (
    id          VARCHAR(64)    PRIMARY KEY,  -- "SETTLE-<timestamp>"
    timestamp   BIGINT         NOT NULL,
    batch_count INT            NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL,
    card_total  DECIMAL(12,2)  NOT NULL,
    qris_total  DECIMAL(12,2)  NOT NULL,
    cash_total  DECIMAL(12,2)  NOT NULL,
    status      VARCHAR(20)    NOT NULL CHECK (status IN ('SUCCESS', 'FAILED')),
    store_type  VARCHAR(20),
    terminal_id VARCHAR(32),
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_settlements_timestamp ON settlements(timestamp DESC);
```

---

## 5. API Endpoints

**Base URL**: `https://api.gbs-pos.com/v1`

All authenticated endpoints require:
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

### 5.1 Authentication

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/login` | No | Authenticate user, return JWT |

### 5.2 Products

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/products` | Yes | List all products (filter by `storeType`) |
| `POST` | `/products` | Yes | Create product (ADMIN only) |
| `PUT` | `/products/:id` | Yes | Update product (ADMIN only) |
| `DELETE` | `/products/:id` | Yes | Delete product (ADMIN only) |

### 5.3 Orders

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/orders` | Yes | List orders (with filters) |
| `GET` | `/orders/:id` | Yes | Get order detail |
| `POST` | `/orders` | Yes | Create order (idempotent) |
| `PATCH` | `/orders/:id/void` | Yes | Void order (ADMIN only) |
| `GET` | `/orders/unsettled/summary` | Yes | Batch stats for settlement |
| `POST` | `/orders/settle` | Yes | Run settlement (ADMIN only) |

### 5.4 Settlements

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/settlements` | Yes | List recent settlements |
| `GET` | `/settlements/:id` | Yes | Get settlement detail |

---

## 6. Request/Response Examples

### 6.1 `POST /login`

**Request:**
```http
POST /v1/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**Response 200 OK:**
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

**Response 401 Unauthorized:**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Username or password is incorrect"
  }
}
```

---

### 6.2 `GET /products`

**Request:**
```http
GET /v1/products?storeType=RETAIL
Authorization: Bearer <token>
```

**Response 200 OK:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1001,
      "name": "Chitato",
      "price": 11500.00,
      "category": "Snacks",
      "imageUrl": "https://images.unsplash.com/photo-1621939514649-28b12e81658b",
      "storeType": "RETAIL",
      "createdAt": "2025-01-15T10:00:00Z",
      "updatedAt": "2025-01-15T10:00:00Z"
    },
    {
      "id": 1002,
      "name": "Indomie Goreng",
      "price": 3500.00,
      "category": "Snacks",
      "imageUrl": "https://images.unsplash.com/photo-1612929633738-8fe44f7ec841",
      "storeType": "RETAIL",
      "createdAt": "2025-01-15T10:00:00Z",
      "updatedAt": "2025-01-15T10:00:00Z"
    }
  ]
}
```

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `storeType` | string | No | Filter: `RETAIL`, `FNB`, `OUTFIT`. Default: all |
| `category` | string | No | Filter by category name |

---

### 6.3 `POST /products`

**Request:**
```http
POST /v1/products
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Teh Botol",
  "price": 5000.00,
  "category": "Beverages",
  "imageUrl": "https://example.com/teh-botol.jpg",
  "storeType": "RETAIL"
}
```

**Response 201 Created:**
```json
{
  "success": true,
  "data": {
    "id": 1025,
    "name": "Teh Botol",
    "price": 5000.00,
    "category": "Beverages",
    "imageUrl": "https://example.com/teh-botol.jpg",
    "storeType": "RETAIL",
    "createdAt": "2025-01-15T12:00:00Z",
    "updatedAt": "2025-01-15T12:00:00Z"
  }
}
```

---

### 6.4 `PUT /products/:id`

**Request:**
```http
PUT /v1/products/1001
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Chitato Rasa Sapi Panggang",
  "price": 12000.00,
  "category": "Snacks",
  "imageUrl": "https://images.unsplash.com/photo-1621939514649-28b12e81658b",
  "storeType": "RETAIL"
}
```

**Response 200 OK:** Same as POST response.

**Response 404 Not Found:**
```json
{
  "success": false,
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product with ID 1001 not found"
  }
}
```

---

### 6.5 `DELETE /products/:id`

**Request:**
```http
DELETE /v1/products/1001
Authorization: Bearer <token>
```

**Response 204 No Content:** (empty body)

---

### 6.6 `POST /orders`

**Request:**
```http
POST /v1/orders
Authorization: Bearer <token>
Content-Type: application/json

{
  "id": "A1B2C3D4",
  "items": [
    {
      "productId": 1001,
      "productName": "Chitato",
      "productPrice": 11500.00,
      "qty": 2,
      "subtotal": 23000.00
    },
    {
      "productId": 1002,
      "productName": "Indomie Goreng",
      "productPrice": 3500.00,
      "qty": 3,
      "subtotal": 10500.00
    }
  ],
  "subtotal": 33500.00,
  "tax": 3350.00,
  "total": 36850.00,
  "paymentMethod": "CARD",
  "cashReceived": null,
  "changeAmount": null,
  "timestamp": 1716023456789,
  "storeType": "RETAIL",
  "terminalId": "POS-001",
  "transactionId": "TXN-20250115-001",
  "approvalCode": "123456",
  "entryMode": "CONTACTLESS",
  "maskedAccount": "************1234",
  "acqMid": "123456789012345",
  "acqTid": "TID001",
  "posMessageId": "MSG-ABC-123",
  "bankName": "BCA"
}
```

**Response 201 Created:**
```json
{
  "success": true,
  "data": {
    "id": "A1B2C3D4",
    "items": [
      {
        "productId": 1001,
        "productName": "Chitato",
        "productPrice": 11500.00,
        "qty": 2,
        "subtotal": 23000.00
      },
      {
        "productId": 1002,
        "productName": "Indomie Goreng",
        "productPrice": 3500.00,
        "qty": 3,
        "subtotal": 10500.00
      }
    ],
    "subtotal": 33500.00,
    "tax": 3350.00,
    "total": 36850.00,
    "paymentMethod": "CARD",
    "cashReceived": null,
    "changeAmount": null,
    "timestamp": 1716023456789,
    "isVoided": false,
    "isSettled": false,
    "storeType": "RETAIL",
    "terminalId": "POS-001",
    "transactionId": "TXN-20250115-001",
    "approvalCode": "123456",
    "entryMode": "CONTACTLESS",
    "maskedAccount": "************1234",
    "acqMid": "123456789012345",
    "acqTid": "TID001",
    "posMessageId": "MSG-ABC-123",
    "bankName": "BCA",
    "createdAt": "2025-01-15T12:00:00Z"
  }
}
```

**IMPORTANT — Idempotency:**
If an order with the same `id` already exists, return `200 OK` with the existing record. Do **not** create a duplicate. This is critical because the app may retry sending the same order if the network fails.

```json
{
  "success": true,
  "data": { ...existing order... },
  "idempotent": true
}
```

---

### 6.7 `GET /orders`

**Request:**
```http
GET /v1/orders?storeType=RETAIL&startDate=1716000000000&endDate=1716100000000&isVoided=false
Authorization: Bearer <token>
```

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `storeType` | string | No | Filter: `RETAIL`, `FNB`, `OUTFIT` |
| `startDate` | long | No | Unix millis, filter orders from this time |
| `endDate` | long | No | Unix millis, filter orders until this time |
| `isVoided` | boolean | No | Filter voided/non-voided orders |
| `isSettled` | boolean | No | Filter settled/non-settled orders |
| `paymentMethod` | string | No | Filter: `CASH`, `CARD`, `QRIS` |
| `terminalId` | string | No | Filter by POS terminal |

**Response 200 OK:**
```json
{
  "success": true,
  "data": [
    {
      "id": "A1B2C3D4",
      "items": [
        {
          "productId": 1001,
          "productName": "Chitato",
          "productPrice": 11500.00,
          "qty": 2,
          "subtotal": 23000.00
        }
      ],
      "subtotal": 33500.00,
      "tax": 3350.00,
      "total": 36850.00,
      "paymentMethod": "CARD",
      "cashReceived": null,
      "changeAmount": null,
      "timestamp": 1716023456789,
      "isVoided": false,
      "isSettled": false,
      "storeType": "RETAIL",
      "terminalId": "POS-001",
      "transactionId": "TXN-20250115-001",
      "approvalCode": "123456",
      "entryMode": "CONTACTLESS",
      "maskedAccount": "************1234",
      "acqMid": "123456789012345",
      "acqTid": "TID001",
      "posMessageId": "MSG-ABC-123",
      "bankName": "BCA",
      "createdAt": "2025-01-15T12:00:00Z"
    }
  ]
}
```

**Note:** Orders should be returned sorted by `timestamp DESC` (newest first).

---

### 6.8 `GET /orders/:id`

**Request:**
```http
GET /v1/orders/A1B2C3D4
Authorization: Bearer <token>
```

**Response 200 OK:** Same structure as a single order from the list endpoint.

**Response 404 Not Found:**
```json
{
  "success": false,
  "error": {
    "code": "ORDER_NOT_FOUND",
    "message": "Order with ID A1B2C3D4 not found"
  }
}
```

---

### 6.9 `PATCH /orders/:id/void`

**Request:**
```http
PATCH /v1/orders/A1B2C3D4/void
Authorization: Bearer <token>
Content-Type: application/json

{
  "reason": "Customer requested cancellation"
}
```

**Response 200 OK:**
```json
{
  "success": true,
  "data": {
    "id": "A1B2C3D4",
    "isVoided": true,
    "voidReason": "Customer requested cancellation",
    "voidedBy": "admin",
    "voidedAt": "2025-01-15T13:00:00Z",
    ...
  }
}
```

**Response 409 Conflict (already voided):**
```json
{
  "success": false,
  "error": {
    "code": "ORDER_ALREADY_VOIDED",
    "message": "Order A1B2C3D4 has already been voided"
  }
}
```

**Response 409 Conflict (already settled):**
```json
{
  "success": false,
  "error": {
    "code": "ORDER_ALREADY_SETTLED",
    "message": "Cannot void a settled order"
  }
}
```

**Response 403 Forbidden (cashier trying to void):**
```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_PERMISSIONS",
    "message": "Only ADMIN role can void orders"
  }
}
```

---

### 6.10 `GET /orders/unsettled/summary`

**Request:**
```http
GET /v1/orders/unsettled/summary?storeType=RETAIL&terminalId=POS-001
Authorization: Bearer <token>
```

**Response 200 OK:**
```json
{
  "success": true,
  "data": {
    "count": 15,
    "total": 450000.00,
    "paymentSummary": {
      "CASH": { "count": 5, "total": 150000.00 },
      "CARD": { "count": 7, "total": 250000.00 },
      "QRIS": { "count": 3, "total": 50000.00 }
    }
  }
}
```

**Logic:** `SUM(orders)` WHERE `is_settled = false` AND `is_voided = false`, grouped by `payment_method`.

---

### 6.11 `POST /orders/settle`

**Request:**
```http
POST /v1/orders/settle
Authorization: Bearer <token>
Content-Type: application/json

{
  "settlementId": "SETTLE-1716023456789",
  "timestamp": 1716023456789,
  "storeType": "RETAIL",
  "terminalId": "POS-001"
}
```

**Response 200 OK:**
```json
{
  "success": true,
  "data": {
    "id": "SETTLE-1716023456789",
    "timestamp": 1716023456789,
    "batchCount": 15,
    "totalAmount": 450000.00,
    "cardTotal": 250000.00,
    "qrisTotal": 50000.00,
    "cashTotal": 150000.00,
    "status": "SUCCESS",
    "storeType": "RETAIL",
    "terminalId": "POS-001",
    "createdAt": "2025-01-15T14:00:00Z"
  }
}
```

**Logic:**
1. Find all orders WHERE `is_settled = false` AND `is_voided = false` (optionally filtered by `storeType` and `terminalId`)
2. Calculate totals per payment method
3. Create settlement record
4. Mark all included orders as `is_settled = true`
5. Return settlement summary

**Response 409 Conflict (no unsettled orders):**
```json
{
  "success": false,
  "error": {
    "code": "NO_UNSETTLED_ORDERS",
    "message": "There are no unsettled orders to settle"
  }
}
```

---

### 6.12 `GET /settlements`

**Request:**
```http
GET /v1/settlements?limit=5&storeType=RETAIL
Authorization: Bearer <token>
```

**Response 200 OK:**
```json
{
  "success": true,
  "data": [
    {
      "id": "SETTLE-1716023456789",
      "timestamp": 1716023456789,
      "batchCount": 15,
      "totalAmount": 450000.00,
      "cardTotal": 250000.00,
      "qrisTotal": 50000.00,
      "cashTotal": 150000.00,
      "status": "SUCCESS",
      "storeType": "RETAIL",
      "terminalId": "POS-001",
      "createdAt": "2025-01-15T14:00:00Z"
    }
  ]
}
```

---

## 7. Authentication & Authorization

### 7.1 JWT Token

- **Algorithm**: HS256
- **Secret**: Store in environment variable `JWT_SECRET` (min 32 characters)
- **Expiry**: 24 hours
- **Payload**:
```json
{
  "sub": 1,
  "username": "admin",
  "role": "ADMIN",
  "iat": 1716023456,
  "exp": 1716109856
}
```

### 7.2 Roles & Permissions

| Endpoint | ADMIN | CASHIER |
|----------|-------|---------|
| `POST /login` | Yes | Yes |
| `GET /products` | Yes | Yes |
| `POST /products` | Yes | No (403) |
| `PUT /products/:id` | Yes | No (403) |
| `DELETE /products/:id` | Yes | No (403) |
| `GET /orders` | Yes | Yes |
| `GET /orders/:id` | Yes | Yes |
| `POST /orders` | Yes | Yes |
| `PATCH /orders/:id/void` | Yes | No (403) |
| `GET /orders/unsettled/summary` | Yes | Yes |
| `POST /orders/settle` | Yes | No (403) |
| `GET /settlements` | Yes | Yes |

### 7.3 Middleware Implementation (Gin example)

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, ErrorResponse("UNAUTHORIZED", "Missing authorization header"))
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv("JWT_SECRET")), nil
        })

        if err != nil || !token.Valid {
            c.JSON(401, ErrorResponse("INVALID_TOKEN", "Invalid or expired token"))
            c.Abort()
            return
        }

        claims := token.Claims.(jwt.MapClaims)
        c.Set("userID", claims["sub"])
        c.Set("username", claims["username"])
        c.Set("role", claims["role"])
        c.Next()
    }
}

func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role := c.GetString("role")
        for _, r := range roles {
            if role == r {
                c.Next()
                return
            }
        }
        c.JSON(403, ErrorResponse("INSUFFICIENT_PERMISSIONS", "You don't have permission to access this resource"))
        c.Abort()
    }
}
```

---

## 8. Business Rules

### 8.1 Tax Calculation

- **Tax rate**: 10% PPN (Indonesian VAT)
- **Formula**: `tax = subtotal × 0.10`
- **Total**: `total = subtotal + tax`
- The Android app calculates these values. The backend should **accept** the values sent by the app, not recalculate them (to avoid floating-point discrepancies).

### 8.2 Currency

- Currency: **Indonesian Rupiah (Rp)**
- Stored as `DECIMAL(12,2)` in the database
- No fractional rupiah in practice (always whole numbers), but use 2 decimal places for consistency

### 8.3 Order Lifecycle

```
Cart → Checkout → Order Created → Receipt → (optional: Void) → (optional: Settle)
```

1. **Cart**: Items added/removed, quantities adjusted. Stored locally on device.
2. **Checkout**: Payment method selected (CASH, CARD, QRIS).
3. **Order Created**: Order + items saved to local Room DB and sent to backend.
4. **Receipt**: Order displayed. If `auto_print` is enabled, thermal printer prints receipt.
5. **Void**: Only non-settled orders can be voided. Requires ADMIN role. Card payments should also trigger Neurogine void on the device.
6. **Settlement**: Batch-closes all unsettled, non-voided orders. Generates settlement record.

### 8.4 Settlement Rules

- Aggregates orders WHERE `is_settled = false` AND `is_voided = false`
- Breakdown by payment method: CARD, QRIS, CASH
- Settlement ID format: `SETTLE-<unix_timestamp_millis>`
- After success, all included orders are marked `is_settled = true`
- A failed settlement should NOT mark orders as settled

### 8.5 Void Rules

- Cash/QRIS orders: Can be voided directly (set `is_voided = true`)
- Card orders: The Android app handles Neurogine void separately. The backend just records the void status.
- **Cannot void settled orders** — return 409 Conflict
- **Cannot void already-voided orders** — return 409 Conflict

### 8.6 Multi-Store System

- Products are scoped to `store_type`: `RETAIL`, `FNB`, `OUTFIT`
- Each store type has different product catalogs and categories
- Orders should track `store_type` for reporting
- Settlements should be scoped to `store_type` and optionally `terminal_id`

### 8.7 Product Categories by Store Type

| Store Type | Categories |
|------------|-----------|
| **RETAIL** | Snacks, Beverages, Household, Personal Care |
| **FNB** | Food, Beverages, Desserts |
| **OUTFIT** | Tops, Bottoms, Outerwear, Accessories |

---

## 9. Neurogine Card Payment Fields

When `payment_method = "CARD"`, the Android app sends these fields from the Neurogine SoftPOS SDK:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `transactionId` | string | Transaction ID from Neurogine | `"TXN-20250115-001"` |
| `approvalCode` | string | Approval code from bank | `"123456"` |
| `entryMode` | string | How card was read | `"CONTACTLESS"`, `"CHIP"`, `"SWIPE"` |
| `maskedAccount` | string | Masked card number | `"************1234"` |
| `acqMid` | string | Acquirer Merchant ID | `"123456789012345"` |
| `acqTid` | string | Acquirer Terminal ID | `"TID001"` |
| `posMessageId` | string | POS message identifier | `"MSG-ABC-123"` |
| `bankName` | string | Selected bank | `"BCA"`, `"BNI"`, `"Mandiri"` |

All fields are nullable except `transactionId`, `approvalCode`, and `posMessageId` which should be present for every card transaction.

---

## 10. Offline-First Sync Strategy

The Android app is designed to work **without internet**. When connectivity returns, it syncs data with the backend.

### 10.1 Product Sync

1. App calls `GET /products?storeType=RETAIL`
2. If successful, app replaces local product cache with server data
3. If failed, app uses locally cached products (seeded on first launch)

**Recommendation**: Add `lastSync` header or query param for delta sync:
```http
GET /v1/products?storeType=RETAIL&lastSync=1716000000000
```
Return only products modified after `lastSync`. Include a `X-Last-Sync` response header.

### 10.2 Order Sync

1. App creates order locally (Room DB)
2. App sends `POST /orders` to backend
3. If network fails, order stays in local queue
4. When online, app retries sending queued orders

**Idempotency is critical**: Use the order `id` (client-generated UUID8) as the unique key. If the order already exists, return 200 with existing data.

**Future enhancement**: Add bulk sync endpoint:
```http
POST /v1/sync/orders
Content-Type: application/json

{
  "terminalId": "POS-001",
  "orders": [
    { ...order1... },
    { ...order2... },
    { ...order3... }
  ]
}
```

### 10.3 Conflict Resolution

- **Products**: Server is authoritative. If the app has a product that doesn't exist on the server, it will be overwritten on next sync.
- **Orders**: Client-generated IDs prevent conflicts. The first `POST` wins.
- **Settlements**: Only one settlement can run at a time per terminal. Use a database-level lock or optimistic concurrency.

---

## 11. Multi-Terminal Support

Multiple Android POS devices can connect to the same backend. Each device should identify itself with a `terminalId`.

### 11.1 Terminal Registration

The app should send `terminalId` with every order and settlement request. The backend should:
- Track which terminal created each order
- Scope settlements to specific terminals
- Allow filtering orders by terminal

### 11.2 Terminal ID Format

The Android app generates terminal IDs as:
```
POS-<device_serial_or_uuid>
```

For now, the app sends `terminalId` in the request body. In the future, we may add a terminal registration endpoint.

### 11.3 Concurrent Settlement Protection

If two terminals try to settle at the same time:
- Use a database transaction with `SELECT ... FOR UPDATE` on unsettled orders
- Or use a settlement lock per terminal: only one active settlement per `terminal_id`

---

## 12. Error Response Format

All error responses follow this format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `INVALID_TOKEN` | 401 | Token expired or malformed |
| `INSUFFICIENT_PERMISSIONS` | 403 | User role doesn't allow this action |
| `INVALID_CREDENTIALS` | 401 | Wrong username/password |
| `PRODUCT_NOT_FOUND` | 404 | Product ID doesn't exist |
| `ORDER_NOT_FOUND` | 404 | Order ID doesn't exist |
| `SETTLEMENT_NOT_FOUND` | 404 | Settlement ID doesn't exist |
| `ORDER_ALREADY_VOIDED` | 409 | Order was already voided |
| `ORDER_ALREADY_SETTLED` | 409 | Cannot void a settled order |
| `NO_UNSETTLED_ORDERS` | 409 | No orders to settle |
| `VALIDATION_ERROR` | 422 | Request body validation failed |
| `INTERNAL_SERVER_ERROR` | 500 | Unexpected server error |

### Validation Error Example (422)

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request body",
    "details": [
      {
        "field": "price",
        "message": "Price must be greater than 0"
      },
      {
        "field": "storeType",
        "message": "Store type must be one of: RETAIL, FNB, OUTFIT"
      }
    ]
  }
}
```

---

## 13. Seed Data

### 13.1 Default Users

```sql
-- Passwords are bcrypt hashed. Plaintext: "admin123" and "cashier123"
INSERT INTO users (username, password_hash, name, role) VALUES
('admin', '$2a$10$...', 'Admin User', 'ADMIN'),
('cashier', '$2a$10$...', 'Cashier User', 'CASHIER');
```

### 13.2 Sample Products

```sql
-- RETAIL products
INSERT INTO products (name, price, category, image_url, store_type) VALUES
('Chitato', 11500, 'Snacks', 'https://images.unsplash.com/photo-1621939514649-28b12e81658b', 'RETAIL'),
('Indomie Goreng', 3500, 'Snacks', 'https://images.unsplash.com/photo-1612929633738-8fe44f7ec841', 'RETAIL'),
('Teh Botol', 5000, 'Beverages', 'https://images.unsplash.com/photo-1556679343-c7306c1976bc', 'RETAIL'),
('Sabun Mandi', 8000, 'Personal Care', 'https://images.unsplash.com/photo-1556228578-0d85b1a4d571', 'RETAIL'),
('Pembersih Lantai', 15000, 'Household', 'https://images.unsplash.com/photo-1585421514284-efb74c2b69ba', 'RETAIL');

-- FNB products
INSERT INTO products (name, price, category, image_url, store_type) VALUES
('Nasi Goreng', 25000, 'Food', 'https://images.unsplash.com/photo-1512058564366-18510be2db19', 'FNB'),
('Es Teh Manis', 8000, 'Beverages', 'https://images.unsplash.com/photo-1556679343-c7306c1976bc', 'FNB'),
('Pisang Goreng', 12000, 'Desserts', 'https://images.unsplash.com/photo-1528975604071-b4dc52a2d18c', 'FNB');

-- OUTFIT products
INSERT INTO products (name, price, category, image_url, store_type) VALUES
('Kaos Polos', 75000, 'Tops', 'https://images.unsplash.com/photo-1521572163474-6864f9cf17ab', 'OUTFIT'),
('Celana Jeans', 250000, 'Bottoms', 'https://images.unsplash.com/photo-1542272604-787c3835535d', 'OUTFIT'),
('Jaket Hoodie', 185000, 'Outerwear', 'https://images.unsplash.com/photo-1556821840-3a63f95609a7', 'OUTFIT');
```

---

## 14. Docker Deployment

### 14.1 `docker-compose.yml`

```yaml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://gbspos:gbspos_secret@postgres:5432/gbs_pos?sslmode=disable
      - JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters
      - PORT=8080
      - ENV=production
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=gbs_pos
      - POSTGRES_USER=gbspos
      - POSTGRES_PASSWORD=gbspos_secret
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U gbspos"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  pgdata:
```

### 14.2 `Dockerfile`

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gbs-pos-api ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /gbs-pos-api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./gbs-pos-api"]
```

### 14.3 Running Locally

```bash
# Clone the repo
git clone <repo-url>
cd gbs-pos-api

# Copy environment file
cp .env.example .env

# Start services
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down
```

---

## 15. Environment Variables

```env
# Server
PORT=8080
ENV=development  # development, staging, production

# Database
DATABASE_URL=postgres://gbspos:gbspos_secret@localhost:5432/gbs_pos?sslmode=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters
JWT_EXPIRY_HOURS=24

# Logging
LOG_LEVEL=debug  # debug, info, warn, error
```

---

## 16. Testing Checklist

Before handing over to the Android team, verify:

- [ ] `POST /login` returns valid JWT for both admin and cashier users
- [ ] JWT middleware rejects requests without token (401)
- [ ] JWT middleware rejects expired tokens (401)
- [ ] `GET /products` returns all products
- [ ] `GET /products?storeType=RETAIL` returns only RETAIL products
- [ ] `POST /products` creates a new product (ADMIN only)
- [ ] `POST /products` returns 403 for CASHIER role
- [ ] `PUT /products/:id` updates an existing product
- [ ] `DELETE /products/:id` removes a product
- [ ] `POST /orders` creates an order with items
- [ ] `POST /orders` is idempotent (same ID returns existing order)
- [ ] `GET /orders` returns orders sorted by timestamp DESC
- [ ] `GET /orders?isVoided=false` filters correctly
- [ ] `PATCH /orders/:id/void` sets `is_voided = true`
- [ ] `PATCH /orders/:id/void` returns 409 for already-voided orders
- [ ] `PATCH /orders/:id/void` returns 409 for already-settled orders
- [ ] `PATCH /orders/:id/void` returns 403 for CASHIER role
- [ ] `GET /orders/unsettled/summary` returns correct counts and totals
- [ ] `POST /orders/settle` creates settlement and marks orders settled
- [ ] `POST /orders/settle` returns 409 when no unsettled orders exist
- [ ] `GET /settlements` returns recent settlements
- [ ] All timestamps are in Unix milliseconds (consistent with Android app)
- [ ] All monetary values use 2 decimal places
- [ ] Error responses follow the standard format (`success`, `error.code`, `error.message`)
- [ ] CORS headers allow requests from the Android app (if using WebView)

---

## 17. Android App Integration Notes

### 17.1 Current Retrofit Setup

The Android app uses:
- **Retrofit 3.0.0** with Gson converter
- **OkHttp 5.3.2** with logging interceptor
- **Base URL**: Currently `https://your-api.com/` (needs to be changed)

### 17.2 What the Android Team Will Do

Once your backend is ready, the Android team will:

1. Update `NetworkModule.kt` with the real base URL
2. Add JWT auth interceptor to OkHttp (reads token from DataStore)
3. Update repository implementations to call your API endpoints
4. Add offline queue for orders that fail to sync
5. Add sync-on-connect logic (detect network changes)

### 17.2 API Contract Changes

If you need to change any endpoint URL, request body, or response format, **please communicate with the Android team first**. The app's Retrofit interfaces are already defined, so any API change requires a matching update on the Android side.

### 17.3 Current API Interface (for reference)

```kotlin
interface ApiService {
    @POST("login")
    suspend fun login(@Body request: LoginRequest): LoginResponse

    @GET("products")
    suspend fun getProducts(): List<ProductResponse>

    @POST("products")
    suspend fun createProduct(@Body product: ProductRequest): ProductResponse

    @PUT("products/{id}")
    suspend fun updateProduct(@Path("id") id: Int, @Body product: ProductRequest): ProductResponse

    @DELETE("products/{id}")
    suspend fun deleteProduct(@Path("id") id: Int)

    @GET("orders")
    suspend fun getOrders(): List<OrderResponse>

    @GET("orders/{id}")
    suspend fun getOrder(@Path("id") id: String): OrderResponse

    @POST("orders")
    suspend fun createOrder(@Body order: OrderRequest): OrderResponse
}
```

---

## 17.5 Multi-Client Configuration (Android + Vue)

Since the backend serves **two different clients** — the Android POS app and the Vue CMS dashboard — there are specific configurations the Golang backend must handle.

### 17.5.1 CORS (Cross-Origin Resource Sharing)

**Who needs it?**
- ✅ **Vue dashboard** (browser) — **REQUIRES CORS**. The dashboard runs on `https://cms.gbs.com` and makes `fetch`/`axios` requests to `https://api.pos.gbs.com`. Without CORS headers, the browser blocks these requests.
- ❌ **Android native app** (OkHttp/Retrofit) — **Does NOT need CORS**. Native HTTP clients ignore CORS policies.

**Gin CORS middleware setup:**

```go
import "github.com/gin-contrib/cors"

func setupCORS() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"https://cms.gbs.com", "http://localhost:5173"}, // Vue dev + prod
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "X-Last-Sync"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}

// In main.go
r := gin.Default()
r.Use(setupCORS())
```

**Important:** Do NOT use `AllowOrigins: ["*"]` with `AllowCredentials: true` — browsers reject this combination. Explicitly list your Vue dashboard domains.

### 17.5.2 Request Headers Comparison

| Header | Android (OkHttp) | Vue (Axios) | Required? |
|--------|-----------------|-------------|-----------|
| `Authorization: Bearer <token>` | ✅ Sent automatically via interceptor | ✅ Sent via Axios interceptors | **Yes** |
| `Content-Type: application/json` | ✅ Gson converter adds this | ✅ Axios default | **Yes** |
| `Accept: application/json` | ✅ Retrofit adds this | ✅ Axios default | Recommended |
| `X-Client-Type: android` | Optional (you can add) | Optional (`web`) | Optional |
| `User-Agent` | `okhttp/5.3.2` | Browser UA | Debug only |

**Recommendation:** Add an optional `X-Client-Type` header so the backend can differentiate clients in logs:
```go
clientType := c.GetHeader("X-Client-Type") // "android" or "web"
```

### 17.5.3 Response Format — Must Be Identical

Both clients consume the **same JSON format**. Do NOT return different structures for Android vs Vue.

```json
{
  "success": true,
  "data": { ... },
  "meta": { "page": 1, "limit": 20 }
}
```

### 17.5.4 File Upload Differences

| Aspect | Vue Dashboard | Android POS |
|--------|--------------|-------------|
| **Upload** | Browser `<input type="file">` → `multipart/form-data` | ❌ Android does NOT upload videos |
| **Download** | ❌ Vue does NOT download videos | ✅ `GET /ads/download/:id` → save to cache |
| **Stream** | Preview in `<video>` tag | ✅ ExoPlayer streams or plays cached file |

**Vue upload example (browser):**
```javascript
const formData = new FormData()
formData.append('file', fileInput.files[0])
formData.append('name', 'Indomie Promo')
formData.append('storeTypes', 'RETAIL')

axios.post('/v1/ads/upload', formData, {
  headers: { 'Content-Type': 'multipart/form-data' }
})
```

**Android download example (OkHttp):**
```kotlin
val request = Request.Builder()
    .url("https://api.cms.gbs.com/v1/ads/download/1")
    .header("Authorization", "Bearer $token")
    .build()

okHttpClient.newCall(request).execute().use { response ->
    val file = File(cacheDir, "ads/indomie.mp4")
    response.body?.byteStream()?.use { input ->
        file.outputStream().use { output ->
            input.copyTo(output)
        }
    }
}
```

### 17.5.5 Video File Serving Headers

When serving video files (`GET /ads/download/:id` or `GET /ads/stream/:id`), the backend **must** set these headers correctly for both clients:

```go
func serveVideo(c *gin.Context, filePath string) {
    fileInfo, _ := os.Stat(filePath)
    
    c.Header("Content-Type", "video/mp4")
    c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
    c.Header("Accept-Ranges", "bytes")           // Required for video seeking/scrubbing
    c.Header("Cache-Control", "public, max-age=86400") // Cache for 24 hours
    c.Header("Content-Disposition", "inline; filename=\""+filepath.Base(filePath)+"\"")
    
    c.File(filePath)
}
```

**Why these headers matter:**
- **`Accept-Ranges: bytes`**: Required for ExoPlayer to seek (jump to any timestamp) and for browser `<video>` to support scrubbing.
- **`Content-Length`**: Required for ExoPlayer to calculate duration and show progress.
- **`Cache-Control`**: Allows Android OkHttp and browser to cache the response.
- **`Content-Disposition: inline`**: Tells browser to play in `<video>` tag instead of downloading.

### 17.5.6 Timeout Considerations

| Client | Network | Recommended Timeout |
|--------|---------|---------------------|
| Android POS | WiFi / 4G | 30s read, 30s write |
| Vue Dashboard | Office WiFi | 10s read, 60s upload |

**OkHttp client config (Android):**
```kotlin
val client = OkHttpClient.Builder()
    .connectTimeout(10, TimeUnit.SECONDS)
    .readTimeout(30, TimeUnit.SECONDS)
    .writeTimeout(30, TimeUnit.SECONDS)
    .build()
```

**Gin server timeout (Go):**
```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      r,
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

### 17.5.7 SSL / HTTPS

| Environment | Requirement |
|-------------|-------------|
| **Production** | **Mandatory HTTPS** for both Android and Vue. Browsers block mixed content. Google Play requires HTTPS for API calls. |
| **Development** | HTTP is fine. For Android emulator, use `http://10.0.2.2:8080` to reach localhost. |

**Android Network Security Config (for development):**
```xml
<!-- res/xml/network_security_config.xml -->
<network-security-config>
    <domain-config cleartextTrafficPermitted="true">
        <domain includeSubdomains="true">10.0.2.2</domain>
        <domain includeSubdomains="true">192.168.1.xxx</domain>
    </domain-config>
</network-security-config>
```

```xml
<!-- AndroidManifest.xml -->
<application
    android:networkSecurityConfig="@xml/network_security_config"
    ... >
```

### 17.5.8 API Versioning

Use URL versioning so both clients can evolve independently:
```
https://api.pos.gbs.com/v1/login
https://api.pos.gbs.com/v2/login  ← future version
```

The Vue dashboard and Android app can upgrade to `v2` at different times.

### 17.5.9 Summary: Backend Checklist for Multi-Client

- [ ] CORS enabled with explicit Vue dashboard origins
- [ ] Same JSON response format for all clients
- [ ] `Authorization: Bearer <token>` auth for both
- [ ] `multipart/form-data` upload endpoint for Vue
- [ ] Video download endpoint with `Accept-Ranges: bytes` for Android
- [ ] Proper `Content-Type`, `Content-Length`, `Cache-Control` headers for video files
- [ ] HTTPS in production
- [ ] URL versioning (`/v1/`, `/v2/`)
- [ ] `X-Client-Type` header logging (optional but useful)

---

## 18. Quick Start for Golang Team

```bash
# 1. Clone and set up
git clone <repo-url>
cd gbs-pos-api
cp .env.example .env

# 2. Run database
docker-compose up -d postgres

# 3. Run migrations
migrate -path migrations -database "postgres://gbspos:gbspos_secret@localhost:5432/gbs_pos?sslmode=disable" up

# 4. Run server
go run cmd/server/main.go

# 5. Test login
curl -X POST http://localhost:8080/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

---

## Contact

For questions or API contract changes, coordinate with the Android team before making changes to endpoint URLs, request bodies, or response formats.

---

# Part 2 — CMS Backend API Guide

This section specifies the **separate CMS backend** for managing video advertisements that play on the customer display.

---

## 19. Architecture Overview

The CMS runs as a **separate backend** from the POS API:

```
┌─────────────┐      ┌─────────────────┐
│ Android POS │ ───► │ POS Backend     │
│   App       │      │ api.pos.gbs.com │
│             │      │   /login        │
│             │      │   /products     │
│             │      │   /orders       │
└─────────────┘      └─────────────────┘
       │
       │             ┌─────────────────┐
       └───────────► │ CMS Backend     │
                     │ api.cms.gbs.com │
                     │   /ads/upload   │
                     │   /ads/active   │
                     │   /ads/playlist │
                     └─────────────────┘
                            │
                            ▼
                     ┌─────────────────┐
                     │ Vue Dashboard   │
                     │ cms.gbs.com     │
                     │   /login        │
                     │   /ads          │
                     │   /upload       │
                     └─────────────────┘
```

**Why separate?**
- POS backend stays lean and fast (critical for transactions)
- CMS handles heavy video uploads without impacting POS performance
- Marketing team can build/manage the Vue dashboard independently
- Independent scaling and deployment

---

## 20. CMS Tech Stack

| Layer | Recommendation |
|-------|---------------|
| Backend Framework | **Gin** or **Echo** |
| Database | **PostgreSQL** (can share the same Postgres instance, separate DB) |
| File Storage | Local filesystem (mounted volume) — MVP approach |
| Auth | JWT (same secret as POS, or separate) |
| Frontend Dashboard | **Vue 3** + **Vite** + **Pinia** + **Tailwind CSS** |
| Video Delivery | Static file serve (nginx or Go's `http.FileServer`) |

---

## 21. CMS Database Schema

### 21.1 `ads`

```sql
CREATE TABLE ads (
    id             SERIAL PRIMARY KEY,
    name           VARCHAR(200)   NOT NULL,
    filename       VARCHAR(255)   NOT NULL,    -- e.g., "indomie.mp4"
    storage_path   VARCHAR(500)   NOT NULL,    -- e.g., "/uploads/ads/indomie_1716023456.mp4"
    file_size      BIGINT         NOT NULL,    -- bytes
    mime_type      VARCHAR(50)    NOT NULL,    -- "video/mp4"
    duration_seconds INT,                      -- video length (optional)
    store_types    VARCHAR(20)[]  NOT NULL,    -- {"RETAIL"}, {"OUTFIT"}, {"RETAIL","FNB"}
    playlist_order INT            NOT NULL DEFAULT 0,
    is_active      BOOLEAN        NOT NULL DEFAULT TRUE,
    start_date     DATE,                       -- scheduling
    end_date       DATE,
    start_time     TIME,                       -- e.g., 09:00:00
    end_time       TIME,                       -- e.g., 21:00:00
    created_by     INT            NOT NULL REFERENCES users(id),
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ads_store_types ON ads USING GIN (store_types);
CREATE INDEX idx_ads_is_active ON ads(is_active);
CREATE INDEX idx_ads_dates ON ads(start_date, end_date);
```

### 21.2 `ad_play_logs` (optional analytics)

```sql
CREATE TABLE ad_play_logs (
    id          SERIAL PRIMARY KEY,
    ad_id       INT            NOT NULL REFERENCES ads(id),
    terminal_id VARCHAR(32),                   -- POS device identifier
    store_type  VARCHAR(20)    NOT NULL,
    played_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ad_play_logs_ad_id ON ad_play_logs(ad_id);
CREATE INDEX idx_ad_play_logs_played_at ON ad_play_logs(played_at);
```

### 21.3 `users` (shared with POS or separate)

Option A: **Share the same `users` table** from the POS database (use `pos_db` with CMS tables)
Option B: **Separate CMS database** with its own `users` table

For simplicity, use the same database but CMS tables are separate. Only ADMIN users can manage ads.

---

## 22. File Storage Strategy (Local Filesystem)

### 22.1 Upload Directory Structure

```
/uploads/
  └── ads/
      ├── indomie_1716023456789.mp4
      ├── outfit_1716023456790.mp4
      └── videodummy_1716023456791.mp4
```

### 22.2 Storage Rules

- Max file size: **50 MB**
- Allowed types: `video/mp4`, `video/webm`, `video/quicktime`
- Filename format: `<original_name>_<unix_timestamp>.<ext>`
- Files are stored on the server's local disk
- In production, mount a persistent volume (`/uploads`) in Docker

### 22.3 Static File Serving

Videos are served via a static file endpoint:

```
GET /uploads/ads/indomie_1716023456789.mp4
```

Use nginx in front of the Go server for efficient file serving, or serve directly with Go's `http.FileServer`.

---

## 23. CMS API Endpoints

**Base URL**: `https://api.cms.gbs.com/v1`

### 23.1 Authentication

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/login` | No | Same as POS — returns JWT |

Use the same JWT secret as POS, or a separate one. The POS app does NOT call CMS login — only the Vue dashboard does.

### 23.2 Ad Management

| Method | Endpoint | Auth | Role | Description |
|--------|----------|------|------|-------------|
| `POST` | `/ads/upload` | Bearer | ADMIN | Upload video (multipart/form-data) |
| `GET` | `/ads` | Bearer | ADMIN | List all ads |
| `GET` | `/ads/:id` | Bearer | ADMIN | Get ad detail |
| `PUT` | `/ads/:id` | Bearer | ADMIN | Update ad metadata |
| `DELETE` | `/ads/:id` | Bearer | ADMIN | Delete ad + file |
| `POST` | `/ads/:id/toggle` | Bearer | ADMIN | Toggle is_active |

### 23.3 POS App Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/ads/active` | Bearer | Get active playlist for a store type |
| `GET` | `/ads/download/:id` | Bearer | Download video file (for offline cache) |
| `POST` | `/ads/:id/play` | Bearer | Log that an ad was played (optional) |

---

## 24. Request/Response Examples

### 24.1 `POST /ads/upload`

**Request:**
```http
POST /v1/ads/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="indomie.mp4"
Content-Type: video/mp4

<binary data>
------WebKitFormBoundary
Content-Disposition: form-data; name="name"

Indomie Promo Januari
------WebKitFormBoundary
Content-Disposition: form-data; name="storeTypes"

RETAIL
------WebKitFormBoundary
Content-Disposition: form-data; name="startDate"

2025-01-01
------WebKitFormBoundary
Content-Disposition: form-data; name="endDate"

2025-01-31
------WebKitFormBoundary--
```

**Response 201 Created:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Indomie Promo Januari",
    "filename": "indomie.mp4",
    "storagePath": "/uploads/ads/indomie_1716023456789.mp4",
    "fileSize": 5242880,
    "mimeType": "video/mp4",
    "storeTypes": ["RETAIL"],
    "playlistOrder": 0,
    "isActive": true,
    "startDate": "2025-01-01",
    "endDate": "2025-01-31",
    "startTime": null,
    "endTime": null,
    "createdBy": 1,
    "createdAt": "2025-01-15T12:00:00Z"
  }
}
```

**Response 413 Payload Too Large:**
```json
{
  "success": false,
  "error": {
    "code": "FILE_TOO_LARGE",
    "message": "Maximum file size is 50MB"
  }
}
```

**Response 415 Unsupported Media Type:**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_FILE_TYPE",
    "message": "Only video/mp4, video/webm, video/quicktime files are allowed"
  }
}
```

---

### 24.2 `GET /ads`

**Request:**
```http
GET /v1/ads?page=1&limit=20
Authorization: Bearer <token>
```

**Response 200 OK:**
```json
{
  "success": true,
  "data": {
    "ads": [
      {
        "id": 1,
        "name": "Indomie Promo Januari",
        "filename": "indomie.mp4",
        "storagePath": "/uploads/ads/indomie_1716023456789.mp4",
        "fileSize": 5242880,
        "durationSeconds": 30,
        "storeTypes": ["RETAIL"],
        "playlistOrder": 0,
        "isActive": true,
        "startDate": "2025-01-01",
        "endDate": "2025-01-31",
        "startTime": null,
        "endTime": null,
        "createdBy": 1,
        "createdAt": "2025-01-15T12:00:00Z"
      },
      {
        "id": 2,
        "name": "Outfit Fashion Week",
        "filename": "outfit.mp4",
        "storagePath": "/uploads/ads/outfit_1716023456790.mp4",
        "fileSize": 8388608,
        "durationSeconds": 45,
        "storeTypes": ["OUTFIT"],
        "playlistOrder": 0,
        "isActive": true,
        "startDate": "2025-02-01",
        "endDate": "2025-02-28",
        "startTime": "09:00:00",
        "endTime": "21:00:00",
        "createdBy": 1,
        "createdAt": "2025-01-15T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 2,
      "totalPages": 1
    }
  }
}
```

---

### 24.3 `PUT /ads/:id`

**Request:**
```http
PUT /v1/ads/1
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Indomie Promo Februari",
  "storeTypes": ["RETAIL", "FNB"],
  "playlistOrder": 1,
  "isActive": true,
  "startDate": "2025-02-01",
  "endDate": "2025-02-28",
  "startTime": "08:00:00",
  "endTime": "22:00:00"
}
```

**Response 200 OK:** Returns updated ad object.

---

### 24.4 `DELETE /ads/:id`

**Request:**
```http
DELETE /v1/ads/1
Authorization: Bearer <token>
```

**Response 204 No Content:** (empty body)

**Behavior:** Delete the database record AND the file from `/uploads/ads/`.

---

### 24.5 `POST /ads/:id/toggle`

**Request:**
```http
POST /v1/ads/1/toggle
Authorization: Bearer <token>
```

**Response 200 OK:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "isActive": false,
    "updatedAt": "2025-01-15T14:00:00Z"
  }
}
```

---

### 24.6 `GET /ads/active` (Called by POS app)

**Request:**
```http
GET /v1/ads/active?storeType=RETAIL
Authorization: Bearer <token>
```

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `storeType` | string | Yes | `RETAIL`, `FNB`, or `OUTFIT` |
| `terminalId` | string | No | POS device identifier |

**Response 200 OK:**
```json
{
  "success": true,
  "data": {
    "playlist": [
      {
        "id": 1,
        "name": "Indomie Promo Januari",
        "filename": "indomie.mp4",
        "downloadUrl": "https://api.cms.gbs.com/v1/ads/download/1",
        "playlistOrder": 0,
        "durationSeconds": 30
      },
      {
        "id": 3,
        "name": "Diskon Retail Akhir Pekan",
        "filename": "videodummy.mp4",
        "downloadUrl": "https://api.cms.gbs.com/v1/ads/download/3",
        "playlistOrder": 1,
        "durationSeconds": 20
      }
    ],
    "updatedAt": "2025-01-15T12:00:00Z"
  }
}
```

**Scheduling Logic:** The backend filters ads where:
1. `is_active = true`
2. `storeTypes` contains the requested `storeType`
3. `start_date <= CURRENT_DATE` (or null)
4. `end_date >= CURRENT_DATE` (or null)
5. `start_time <= CURRENT_TIME` (or null)
6. `end_time >= CURRENT_TIME` (or null)

Returns playlist sorted by `playlist_order ASC`.

**Response 404 No Active Ads:**
```json
{
  "success": true,
  "data": {
    "playlist": [],
    "updatedAt": "2025-01-15T12:00:00Z"
  }
}
```

---

### 24.7 `GET /ads/download/:id` (Called by POS app)

**Request:**
```http
GET /v1/ads/download/1
Authorization: Bearer <token>
```

**Response:** Binary file stream with `Content-Type: video/mp4` and `Content-Disposition: attachment; filename="indomie.mp4"`

The POS app downloads this file and saves it to local cache for offline playback.

---

### 24.8 `POST /ads/:id/play` (Optional analytics)

**Request:**
```http
POST /v1/ads/1/play
Authorization: Bearer <token>
Content-Type: application/json

{
  "terminalId": "POS-001",
  "storeType": "RETAIL"
}
```

**Response 200 OK:**
```json
{
  "success": true,
  "data": {
    "logged": true
  }
}
```

---

## 25. Vue Dashboard Specification

### 25.1 Tech Stack

| Layer | Choice |
|-------|--------|
| Framework | Vue 3 (Composition API) |
| Build Tool | Vite |
| State Management | Pinia |
| Styling | Tailwind CSS |
| HTTP Client | Axios |
| Video Preview | HTML5 `<video>` |
| Icons | Heroicons |

### 25.2 Dashboard Pages

| Route | Page | Description |
|-------|------|-------------|
| `/login` | Login | Same credentials as POS (admin/admin123) |
| `/dashboard` | Dashboard | Overview: total ads, active ads, play stats |
| `/ads` | Ad List | Table of all ads with search, filter, pagination |
| `/ads/new` | Upload Ad | Upload form with drag-and-drop, store type selection, scheduling |
| `/ads/:id` | Edit Ad | Edit metadata, preview video, toggle active |
| `/ads/:id/analytics` | Analytics | Play logs, impressions per terminal (optional) |

### 25.3 Upload Form Fields

```
┌─────────────────────────────────────┐
│ Upload Video                        │
│ [Drag & Drop or Click to Browse]    │
│                                     │
│ Name: [Indomie Promo Januari    ]   │
│                                     │
│ Store Types:                        │
│ [x] Retail  [ ] F&B  [ ] Outfit     │
│                                     │
│ Playlist Order: [0]                 │
│                                     │
│ Schedule:                           │
│ Start Date: [2025-01-01]            │
│ End Date:   [2025-01-31]            │
│ Start Time: [09:00] (optional)      │
│ End Time:   [21:00] (optional)      │
│                                     │
│ [Save Ad]                           │
└─────────────────────────────────────┘
```

### 25.4 Ad List Columns

| Column | Description |
|--------|-------------|
| Thumbnail | Video preview frame |
| Name | Ad name |
| Store Types | RETAIL, FNB, OUTFIT badges |
| Schedule | Jan 1 - Jan 31, 09:00-21:00 |
| Status | Active / Inactive toggle |
| Order | Playlist order |
| Actions | Edit, Delete |

---

## 26. POS App Integration (Video Download & Cache)

### 26.1 Current Behavior

The Android app currently plays hardcoded videos from `res/raw/`:
- `videodummy.mp4` (fallback)
- `indomie.mp4` (Retail)
- `outfit.mp4` (Outfit)

### 26.2 New Behavior with CMS

```
1. On app startup OR periodic sync (every 15 minutes):
   a. Call GET /ads/active?storeType=RETAIL
   b. Compare response with local cache manifest
   c. Download any new/changed videos to app cache
   d. Delete videos no longer in the playlist

2. VideoAdPlayer reads from local cache:
   a. If cache miss → play fallback (res/raw/videodummy.mp4)
   b. If cache hit → play from /data/data/<pkg>/cache/ads/<filename>

3. If offline:
   a. Use last cached playlist
   b. Videos already downloaded continue to play
```

### 26.3 Android Cache Directory

```kotlin
// Internal cache directory for ads
val adsCacheDir = File(context.cacheDir, "ads")
if (!adsCacheDir.exists()) adsCacheDir.mkdirs()

// Downloaded file path
val videoFile = File(adsCacheDir, "indomie.mp4")
```

### 26.4 Sync Service (Pseudo-code)

```kotlin
class AdSyncService {
    suspend fun syncAds(storeType: StoreType) {
        try {
            val response = cmsApi.getActiveAds(storeType.name)
            val playlist = response.playlist

            // Download new videos
            for (ad in playlist) {
                val cachedFile = File(adsCacheDir, ad.filename)
                if (!cachedFile.exists() || isStale(cachedFile, response.updatedAt)) {
                    downloadVideo(ad.downloadUrl, cachedFile)
                }
            }

            // Delete videos no longer in playlist
            val playlistFilenames = playlist.map { it.filename }.toSet()
            adsCacheDir.listFiles()?.forEach { file ->
                if (file.name !in playlistFilenames) {
                    file.delete()
                }
            }

            // Save manifest
            saveManifest(playlist)
        } catch (e: Exception) {
            // Offline — use cached manifest
            Log.w("AdSync", "Failed to sync ads", e)
        }
    }
}
```

### 26.5 VideoAdPlayer Update

Instead of hardcoded resource URIs, `VideoAdPlayer` should:
1. Read the cached playlist manifest
2. Build `MediaItem` list from `File` URIs (`file://...`)
3. If no cached videos exist, fall back to `res/raw/videodummy.mp4`

```kotlin
val videoUri = if (cachedFile.exists()) {
    Uri.fromFile(cachedFile).toString()
} else {
    "android.resource://${context.packageName}/raw/videodummy"
}
```

---

## 27. CMS Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `FILE_TOO_LARGE` | 413 | Video exceeds 50MB |
| `INVALID_FILE_TYPE` | 415 | Not an allowed video format |
| `UPLOAD_FAILED` | 500 | Server error during upload |
| `AD_NOT_FOUND` | 404 | Ad ID doesn't exist |
| `INVALID_SCHEDULE` | 422 | Start date after end date |
| `NO_ACTIVE_ADS` | 404 | No ads match the current schedule |
| `INSUFFICIENT_PERMISSIONS` | 403 | Only ADMIN can manage ads |

---

## 28. Docker Compose (CMS + POS together)

```yaml
version: '3.8'

services:
  # POS Backend
  pos-api:
    build:
      context: ./gbs-pos-api
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://gbspos:gbspos_secret@postgres:5432/gbs_pos?sslmode=disable
      - JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters
      - PORT=8080
    depends_on:
      - postgres
    volumes:
      - pos-uploads:/uploads/pos
    restart: unless-stopped

  # CMS Backend
  cms-api:
    build:
      context: ./gbs-cms-api
      dockerfile: Dockerfile
    ports:
      - "8081:8081"
    environment:
      - DATABASE_URL=postgres://gbspos:gbspos_secret@postgres:5432/gbs_pos?sslmode=disable
      - JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters
      - PORT=8081
      - UPLOAD_DIR=/uploads/ads
    depends_on:
      - postgres
    volumes:
      - cms-uploads:/uploads/ads
    restart: unless-stopped

  # Vue Dashboard (served by nginx)
  cms-dashboard:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./gbs-cms-dashboard/dist:/usr/share/nginx/html:ro
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
    depends_on:
      - cms-api
    restart: unless-stopped

  # PostgreSQL (shared)
  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=gbs_pos
      - POSTGRES_USER=gbspos
      - POSTGRES_PASSWORD=gbspos_secret
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U gbspos"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  pgdata:
  pos-uploads:
  cms-uploads:
```

---

## 29. Summary: What Each Team Builds

### Golang Team (POS Backend)
Already specified in Part 1 of this guide. Handles transactions, products, orders, settlements.

### Golang Team (CMS Backend)
Builds a separate API on port `8081` with:
- File upload handler (multipart, 50MB limit)
- Ad CRUD with scheduling
- Active playlist endpoint
- Static file serving for videos

### Vue Team (CMS Dashboard)
Builds a web app deployed at `cms.gbs.com` with:
- Login page (shared JWT)
- Ad upload with drag-and-drop
- Ad list with filters and pagination
- Scheduling UI (dates + time ranges)
- Video preview

### Android Team (POS App)
Updates the existing app:
- Add `CmsApiService` (Retrofit) for CMS endpoints
- Add `AdSyncService` to download/cache videos
- Update `VideoAdPlayer` to read from cache
- Periodic background sync (every 15 minutes)

---

## 30. Testing Checklist (CMS)

Before the Android team integrates:

- [ ] `POST /ads/upload` accepts multipart video upload
- [ ] Upload rejects files > 50MB
- [ ] Upload rejects non-video files
- [ ] `GET /ads` lists all ads with pagination
- [ ] `PUT /ads/:id` updates metadata without re-uploading file
- [ ] `DELETE /ads/:id` deletes both DB record and file
- [ ] `GET /ads/active?storeType=RETAIL` returns only active, scheduled ads
- [ ] Scheduling works: start_date, end_date, start_time, end_time
- [ ] `GET /ads/download/:id` streams the video file
- [ ] Videos are served with correct `Content-Type`
- [ ] `POST /ads/:id/play` logs play event (optional)
- [ ] Only ADMIN role can access CMS management endpoints
- [ ] CORS headers allow requests from Vue dashboard

---

## 31. Contact

For questions or API contract changes, coordinate with the Android team before making changes to endpoint URLs, request bodies, or response formats.
