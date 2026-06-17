# GBS POS & CMS API — Endpoint Documentation

Complete reference for all REST API endpoints. Base URLs:

- **POS API**: `http://localhost:8080`
- **CMS API**: `http://localhost:8081`

All authenticated endpoints require: `Authorization: Bearer <jwt_token>`

---

## Table of Contents

- [POS API](#pos-api)
  - [Authentication](#pos-authentication)
  - [Products](#pos-products)
  - [Orders](#pos-orders)
  - [Settlements](#pos-settlements)
- [CMS API](#cms-api)
  - [Authentication](#cms-authentication)
  - [Ads](#cms-ads)
- [Response Format](#response-format)
- [Error Codes](#error-codes)

---

## POS API

### Authentication

#### `POST /v1/login`

Authenticate a user and receive a JWT token.

**Handler**: `AuthHandler.Login`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | No |
| **Request Body** | `{"username": "admin", "password": "admin123"}` |
| **Response 200** | `{"success": true, "data": {"user": {...}, "token": "eyJ..."}}` |
| **Response 401** | `INVALID_CREDENTIALS` |
| **Response 422** | `VALIDATION_ERROR` — missing username/password |

**What it does:**
1. Validates the request body (both fields required via Gin binding)
2. Calls `AuthService.Login` which:
   - Finds user by username
   - Compares bcrypt password hash
   - Generates JWT with `sub` (user ID), `username`, `role`, `iat`, `exp`
3. Returns user object + token

**Default users:**
- `admin` / `admin123` → role: `ADMIN`
- `cashier` / `cashier123` → role: `CASHIER`

---

### Products

#### `GET /v1/products`

List all products with optional filters.

**Handler**: `ProductHandler.List`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes |
| **Query Params** | `storeType` (RETAIL/FNB/OUTFIT), `category`, `lastSync` (unix millis) |
| **Response Header** | `X-Last-Sync: <unix_millis>` |
| **Response 200** | `{"success": true, "data": [Product]}` |

**What it does:**
- Passes filters to `ProductService.List` which queries the repository
- `lastSync` returns only products with `updated_at > lastSync` (delta sync for offline-first)
- Sets `X-Last-Sync` header to current server time

---

#### `POST /v1/products`

Create a new product.

**Handler**: `ProductHandler.Create`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes + ADMIN role |
| **Request Body** | `{"name": "...", "price": 5000.00, "category": "...", "imageUrl": "...", "storeType": "RETAIL"}` |
| **Response 201** | `{"success": true, "data": Product}` |
| **Response 403** | `INSUFFICIENT_PERMISSIONS` (cashier role) |
| **Response 422** | `VALIDATION_ERROR` |

**What it does:**
1. Binds JSON body to `model.Product` (Gin validation: all fields required)
2. Calls `ProductService.Create` → repository inserts into DB
3. Returns 201 Created with the new product (including auto-generated `id`)

---

#### `PUT /v1/products/:id`

Update an existing product.

**Handler**: `ProductHandler.Update`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes + ADMIN role |
| **Path Param** | `id` — product ID (uint) |
| **Request Body** | Partial update — any subset of product fields |
| **Response 200** | `{"success": true, "data": Product}` |
| **Response 404** | `PRODUCT_NOT_FOUND` |

**What it does:**
1. Parses `id` from URL path
2. Validates request body JSON
3. Calls `ProductService.Update` which:
   - Finds existing product by ID
   - Applies field guards: only updates non-empty / non-zero fields
   - **Does NOT clear omitted fields** (e.g., if you only send `name`, `price` stays unchanged)
4. Returns updated product

---

#### `DELETE /v1/products/:id`

Delete a product.

**Handler**: `ProductHandler.Delete`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes + ADMIN role |
| **Path Param** | `id` — product ID (uint) |
| **Response 204** | Empty body |
| **Response 404** | `PRODUCT_NOT_FOUND` |

---

### Orders

#### `GET /v1/orders`

List orders with filters.

**Handler**: `OrderHandler.List`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes |
| **Query Params** | `storeType`, `startDate` (unix millis), `endDate` (unix millis), `isVoided` (bool), `isSettled` (bool), `paymentMethod` (CASH/CARD/QRIS), `terminalId` |
| **Response 200** | `{"success": true, "data": [Order]}` |

**What it does:**
- Builds a dynamic GORM query based on which filters are provided
- Orders are always sorted by `timestamp DESC` (newest first)
- `isVoided` and `isSettled` use boolean pointers (nil = no filter)

---

#### `GET /v1/orders/:id`

Get a single order by ID.

**Handler**: `OrderHandler.Get`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes |
| **Path Param** | `id` — order ID (string, client-generated UUID) |
| **Response 200** | `{"success": true, "data": Order}` |
| **Response 404** | `ORDER_NOT_FOUND` |

**What it does:**
- Calls `OrderService.Get` → `FindByIDWithItems`
- Returns order with all `items` pre-loaded (via GORM `Preload`)

---

#### `POST /v1/orders`

Create an order. **Idempotent** — safe to retry.

**Handler**: `OrderHandler.Create`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes |
| **Request Body** | Full order with `items` array, `paymentMethod`, Neurogine card fields, `storeType`, `terminalId` |
| **Response 201** | First create — `{"success": true, "data": Order}` |
| **Response 200** | Duplicate — `{"success": true, "data": Order, "idempotent": true}` |
| **Response 422** | `VALIDATION_ERROR` — enum validation, missing items, negative values |

**What it does:**
1. **Validates request body** via `service.ValidateOrder`:
   - `id` is required
   - `items` cannot be empty, each item `qty > 0`
   - `subtotal`, `tax`, `total` must be >= 0
   - `paymentMethod` must be CASH, CARD, or QRIS
   - `storeType` must be RETAIL, FNB, or OUTFIT (if provided)
2. **Idempotency check** in `OrderService.Create`:
   - Queries DB for existing order with same `id`
   - If found → returns existing order + `idempotent: true`
   - If DB error (not RecordNotFound) → returns 500
   - Otherwise → inserts new order
3. Returns `201` on first create, `200` with `idempotent: true` on duplicate

---

#### `POST /v1/sync/orders`

Bulk sync orders (offline-first queue flush).

**Handler**: `OrderHandler.BulkSync`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes |
| **Request Body** | `{"terminalId": "POS-001", "orders": [Order, Order, ...]}` |
| **Response 200** | `{"success": true, "data": {"orders": [...], "created": N, "existing": N, "failed": N, "idempotent": bool}}` |

**What it does:**
1. Iterates through `orders` array
2. For each order, injects `terminalId` if missing
3. Calls `OrderService.Create` for each (handles idempotency automatically)
4. Returns summary counts: created, existing (idempotent), failed

---

#### `PATCH /v1/orders/:id/void`

Void (cancel) an order.

**Handler**: `OrderHandler.Void`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes + ADMIN role |
| **Path Param** | `id` — order ID |
| **Request Body** | `{"reason": "Customer requested cancellation"}` |
| **Response 200** | `{"success": true, "data": Order}` |
| **Response 403** | `INSUFFICIENT_PERMISSIONS` |
| **Response 409** | `ORDER_ALREADY_VOIDED` or `ORDER_ALREADY_SETTLED` |
| **Response 404** | `ORDER_NOT_FOUND` |

**What it does:**
1. Extracts `voidedBy` from JWT claims (username)
2. Calls `OrderService.Void` which enforces business rules:
   - Cannot void settled orders → 409
   - Cannot void already-voided orders → 409
   - Sets `is_voided = true`, `void_reason`, `voided_by`, `voided_at = NOW()`
3. Reloads order with items for consistent response

---

#### `GET /v1/orders/unsettled/summary`

Get batch statistics for settlement (count and totals by payment method).

**Handler**: `OrderHandler.UnsettledSummary`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes |
| **Query Params** | `storeType`, `terminalId` |
| **Response 200** | `{"success": true, "data": {"count": 15, "total": 450000.00, "paymentSummary": {"CASH": {...}, "CARD": {...}, "QRIS": {...}}}}` |

**What it does:**
- Aggregates `SUM(orders)` WHERE `is_settled = false` AND `is_voided = false`
- Groups by `payment_method`
- Returns counts and totals for all three payment methods (0 if none)

---

#### `POST /v1/orders/settle`

Run settlement batch — atomically closes all unsettled orders.

**Handler**: `OrderHandler.Settle`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes + ADMIN role |
| **Request Body** | `{"settlementId": "SETTLE-1716023456789", "timestamp": 1716023456789, "storeType": "RETAIL", "terminalId": "POS-001"}` |
| **Response 200** | `{"success": true, "data": Settlement}` |
| **Response 409** | `NO_UNSETTLED_ORDERS` |
| **Response 403** | `INSUFFICIENT_PERMISSIONS` |

**What it does:**
1. Calls `SettlementService.Settle` inside a **database transaction**
2. Within the transaction:
   - `FindUnsettledOrders` with `SELECT ... FOR UPDATE` (prevents concurrent settlement)
   - If no orders → rollback with `NO_UNSETTLED_ORDERS`
   - Calculates totals per payment method (CASH, CARD, QRIS)
   - Creates settlement record with `status = "SUCCESS"`
   - Marks all included orders as `is_settled = true`
3. Commits transaction
4. Returns settlement summary

---

### Settlements

#### `GET /v1/settlements`

List recent settlements.

**Handler**: `SettlementHandler.List`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes |
| **Query Params** | `limit` (default 20), `storeType` |
| **Response 200** | `{"success": true, "data": [Settlement]}` |

---

#### `GET /v1/settlements/:id`

Get settlement detail by ID.

**Handler**: `SettlementHandler.Get`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes |
| **Path Param** | `id` — settlement ID |
| **Response 200** | `{"success": true, "data": Settlement}` |
| **Response 404** | `SETTLEMENT_NOT_FOUND` |

---

## CMS API

### Authentication

#### `POST /v1/login`

Authenticate and receive JWT. **Same logic as POS login**, separate instance.

**Handler**: `AuthHandler.Login` (CMS)

| Aspect | Detail |
|--------|--------|
| **Auth Required** | No |
| **Credentials** | Same users table as POS (admin/admin123, cashier/cashier123) |
| **Response** | Same format as POS login |

---

### Ads

#### `POST /v1/ads/upload`

Upload a video ad (multipart/form-data).

**Handler**: `CMSHandler.UploadAd`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes + ADMIN role |
| **Content-Type** | `multipart/form-data` |
| **Form Fields** | `file` (required), `name` (required), `storeTypes` (array), `playlistOrder`, `startDate`, `endDate`, `startTime`, `endTime` |
| **Response 201** | `{"success": true, "data": Ad}` |
| **Response 413** | `FILE_TOO_LARGE` (> 50MB) |
| **Response 415** | `INVALID_FILE_TYPE` (must be .mp4, .webm, .mov) |
| **Response 422** | `INVALID_SCHEDULE` (startDate > endDate) |

**What it does:**
1. Validates multipart file (size, extension)
2. Parses scheduling fields (date/time as optional pointers)
3. Creates DB record first (so we know the storage path)
4. Saves file to disk at `<uploadDir>/<name>_<timestamp>.<ext>`
5. **Cleanup**: if file save fails, deletes the orphaned DB record
6. Returns 201 with full ad object

---

#### `GET /v1/ads`

List all ads with pagination.

**Handler**: `CMSHandler.ListAds`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes + ADMIN role |
| **Query Params** | `page` (default 1), `limit` (default 20) |
| **Response 200** | `{"success": true, "data": {"ads": [...], "pagination": {"page": 1, "limit": 20, "total": N, "totalPages": N}}}` |

---

#### `GET /v1/ads/:id`

Get ad detail by ID.

**Handler**: `CMSHandler.GetAd`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes + ADMIN role |
| **Response 200** | `{"success": true, "data": Ad}` |
| **Response 404** | `AD_NOT_FOUND` |

---

#### `PUT /v1/ads/:id`

Update ad metadata (no file re-upload).

**Handler**: `CMSHandler.UpdateAd`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes + ADMIN role |
| **Request Body** | Partial ad fields (name, storeTypes, playlistOrder, isActive, dates, times) |
| **Response 200** | `{"success": true, "data": Ad}` |
| **Response 404** | `AD_NOT_FOUND` |
| **Response 422** | `INVALID_SCHEDULE` |

**What it does:**
- Partial update with field guards (similar to products)
- Validates schedule consistency (startDate <= endDate)
- Does NOT modify the uploaded file

---

#### `DELETE /v1/ads/:id`

Delete ad + associated file.

**Handler**: `CMSHandler.DeleteAd`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes + ADMIN role |
| **Response 204** | Empty body |
| **Response 404** | `AD_NOT_FOUND` |

**What it does:**
1. Finds ad in DB
2. Deletes DB record
3. Deletes file from disk (silently ignores if file already gone)
4. Returns 204

---

#### `POST /v1/ads/:id/toggle`

Toggle ad `isActive` status.

**Handler**: `CMSHandler.ToggleAd`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes + ADMIN role |
| **Response 200** | `{"success": true, "data": {"id": 1, "isActive": false, "updatedAt": "..."}}` |
| **Response 404** | `AD_NOT_FOUND` |

---

#### `GET /v1/ads/active`

Get active playlist for a store type (called by Android POS).

**Handler**: `CMSHandler.ActivePlaylist`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes (any role) |
| **Query Param** | `storeType` (required: RETAIL/FNB/OUTFIT) |
| **Response 200** | `{"success": true, "data": {"playlist": [...], "updatedAt": "..."}}` |

**What it does:**
1. Validates `storeType` is provided
2. Calls `CMSService.GetActivePlaylist` which filters:
   - `is_active = true`
   - `store_types LIKE '%storeType%'`
   - `start_date <= CURRENT_DATE` (or null)
   - `end_date >= CURRENT_DATE` (or null)
   - `start_time <= CURRENT_TIME` (or null)
   - `end_time >= CURRENT_TIME` (or null)
3. Sorts by `playlist_order ASC`
4. Returns playlist items with `downloadUrl` pre-constructed

---

#### `GET /v1/ads/download/:id`

Download video file (binary stream).

**Handler**: `CMSHandler.DownloadAd`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes (any role) |
| **Response** | Binary file stream |
| **Headers** | `Content-Type`, `Content-Length`, `Accept-Ranges: bytes`, `Cache-Control: public, max-age=86400`, `Content-Disposition: inline` |
| **Response 404** | `AD_NOT_FOUND` (or file not found) |

**What it does:**
- Reads file from stored `storagePath`
- Sets all required video headers for ExoPlayer/browser scrubbing
- Uses `c.File()` for efficient file serving

---

#### `POST /v1/ads/:id/play`

Log that an ad was played (optional analytics).

**Handler**: `CMSHandler.LogPlay`

| Aspect | Detail |
|--------|--------|
| **Auth Required** | Yes (any role) |
| **Request Body** | `{"terminalId": "POS-001", "storeType": "RETAIL"}` |
| **Response 200** | `{"success": true, "data": {"logged": true}}` |

---

## Response Format

All endpoints return the same envelope:

```json
{
  "success": true,
  "data": { ... },
  "idempotent": true
}
```

Or on error:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "details": [
      {"field": "price", "message": "Price must be greater than 0"}
    ]
  }
}
```

## Error Codes

| Code | HTTP | When |
|------|------|------|
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `INVALID_TOKEN` | 401 | Token expired or malformed |
| `INVALID_CREDENTIALS` | 401 | Wrong username/password |
| `INSUFFICIENT_PERMISSIONS` | 403 | Cashier trying admin action |
| `PRODUCT_NOT_FOUND` | 404 | Product ID doesn't exist |
| `ORDER_NOT_FOUND` | 404 | Order ID doesn't exist |
| `SETTLEMENT_NOT_FOUND` | 404 | Settlement ID doesn't exist |
| `AD_NOT_FOUND` | 404 | Ad ID doesn't exist |
| `ORDER_ALREADY_VOIDED` | 409 | Cannot void again |
| `ORDER_ALREADY_SETTLED` | 409 | Cannot void settled order |
| `NO_UNSETTLED_ORDERS` | 409 | Nothing to settle |
| `FILE_TOO_LARGE` | 413 | Video > 50MB |
| `INVALID_FILE_TYPE` | 415 | Not .mp4/.webm/.mov |
| `INVALID_SCHEDULE` | 422 | startDate > endDate |
| `VALIDATION_ERROR` | 422 | Request body validation failed |
| `INTERNAL_SERVER_ERROR` | 500 | Unexpected error |
