# Implementation Plan: Web POS with Card Payment via Companion App

## Overview

Membangun Web POS yang berjalan di browser Sunmi D3 Pro dengan fitur:
- Same features seperti POS app (products, cart, checkout)
- Card payment via Companion App (device terpisah dengan NFC)
- Real-time communication via WebSocket

## Architecture Summary

```
┌─────────────────────────────────────────────────────────────────┐
│                     Sunmi D3 Pro (Browser)                       │
│  Web POS ──WebSocket──► Backend ──HTTP Polling──► Companion App │
└─────────────────────────────────────────────────────────────────┘
```

## Phases

---

## Phase 1: Backend - WebSocket Infrastructure

### 1.1 Add Dependencies
```go
// go.mod
github.com/gorilla/websocket v1.5.3
```

### 1.2 Create WebSocket Hub
**Files:**
- `gbs-pos-api/internal/websocket/hub.go`
- `gbs-pos-api/internal/websocket/client.go`
- `gbs-pos-api/internal/websocket/messages.go`

**Features:**
- Terminal-based client registration
- Broadcast to specific terminal
- Ping/pong keepalive
- Thread-safe operations

### 1.3 Create Card Payment Model
**File:** `gbs-pos-api/internal/model/card_payment.go`

**Schema:**
```go
type CardPayment struct {
    ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
    OrderID        string    `gorm:"size:50;not null"`
    Amount         float64   `gorm:"type:decimal(15,2);not null"`
    Status         string    `gorm:"size:20;default:'WAITING_FOR_TAP'"`
    DeviceID       string    `gorm:"size:50"`
    TerminalID     string    `gorm:"size:50"`
    TransactionID  string    `gorm:"size:100"`
    CardBrand      string    `gorm:"size:20"`
    MaskedCard     string    `gorm:"size:20"`
    AuthCode       string    `gorm:"size:20"`
    FailureReason  string    `gorm:"type:text"`
    ExpiresAt      time.Time
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

**Status Values:**
- `WAITING_FOR_TAP`
- `SUCCESS`
- `FAILED`
- `CANCELLED`
- `EXPIRED`

### 1.4 Create Card Payment Service
**File:** `gbs-pos-api/internal/service/card_payment_service.go`

**Methods:**
- `CreatePayment(orderID, amount, terminalID, deviceID)` → Create & return payment
- `ConfirmPayment(paymentID, result)` → Update status, broadcast via WS
- `CancelPayment(paymentID)` → Cancel & notify
- `GetPendingPayments(deviceID)` → For companion app polling
- `ExpirePayments()` → Background cleanup

### 1.5 Create Card Payment Handler
**File:** `gbs-pos-api/internal/handler/card_payment_handler.go`

**Endpoints:**
```
POST   /v1/card-payment/init         - Initialize payment
POST   /v1/card-payment/:id/confirm  - Confirm from companion app
POST   /v1/card-payment/:id/cancel   - Cancel payment
GET    /v1/card-payment/pending      - Get pending payments
GET    /v1/card-payment/:id          - Get payment status
```

### 1.6 Create Routes
**Files:**
- `gbs-pos-api/internal/router/websocket_route.go` → `GET /ws`
- `gbs-pos-api/internal/router/card_payment_route.go` → Card payment routes
- `gbs-pos-api/internal/router/router.go` → Register new routes

### 1.7 Integrate WebSocket Hub in Main
**File:** `gbs-pos-api/cmd/server/main.go`

**Changes:**
- Initialize WebSocket hub
- Register hub cleanup on shutdown
- Pass hub to handlers via dependency injection

---

## Phase 2: Backend - Database Migration

### 2.1 Create Migration
**File:** `migrations/005_create_card_payments.sql`

```sql
CREATE TABLE IF NOT EXISTS card_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'WAITING_FOR_TAP',
    device_id VARCHAR(50),
    terminal_id VARCHAR(50),
    transaction_id VARCHAR(100),
    card_brand VARCHAR(20),
    masked_card VARCHAR(20),
    auth_code VARCHAR(20),
    failure_reason TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_card_payments_status ON card_payments(status);
CREATE INDEX idx_card_payments_device ON card_payments(device_id);
CREATE INDEX idx_card_payments_terminal ON card_payments(terminal_id);
CREATE INDEX idx_card_payments_expires ON card_payments(expires_at);
```

---

## Phase 3: Web POS Frontend - Project Setup

### 3.1 Create Project
```
pos-web/
├── package.json
├── vite.config.ts
├── tsconfig.json
├── tailwind.config.js
├── index.html
└── src/
    ├── main.tsx
    ├── App.tsx
    ├── api/
    │   ├── client.ts
    │   ├── payment.ts
    │   └── websocket.ts
    ├── components/
    │   ├── ProductGrid.tsx
    │   ├── ProductCard.tsx
    │   ├── CategoryTabs.tsx
    │   ├── CartPanel.tsx
    │   ├── CartItem.tsx
    │   ├── CheckoutPanel.tsx
    │   ├── PaymentMethod.tsx
    │   ├── PaymentWaiting.tsx      # NEW: Card payment waiting
    │   ├── ReceiptPanel.tsx
    │   └── ReceiptPrinter.tsx
    ├── hooks/
    │   ├── useWebSocket.ts         # NEW: WebSocket connection
    │   ├── useCart.ts
    │   ├── useProducts.ts
    │   └── useAuth.ts
    ├── stores/
    │   ├── cartStore.ts
    │   ├── productStore.ts
    │   └── paymentStore.ts         # NEW: Payment state
    ├── pages/
    │   ├── LoginPage.tsx
    │   ├── POSPage.tsx
    │   └── ReceiptPage.tsx
    └── types/
        ├── product.ts
        ├── cart.ts
        └── payment.ts
```

### 3.2 Dependencies
```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.0",
    "zustand": "^4.4.0",
    "axios": "^1.6.0",
    "lucide-react": "^0.294.0"
  },
  "devDependencies": {
    "vite": "^5.0.0",
    "typescript": "^5.3.0",
    "tailwindcss": "^3.3.0",
    "@vitejs/plugin-react": "^4.2.0"
  }
}
```

---

## Phase 4: Web POS Frontend - Core Features

### 4.1 Authentication
- `LoginPage.tsx` → Username/password login
- Store JWT token in localStorage
- Auto-redirect to POS on valid token

### 4.2 Product Grid
- Fetch products from `/v1/products`
- Category tabs filter
- Search functionality
- Product card with quick-add to cart

### 4.3 Cart Management
- Add/remove/update quantity
- Show subtotal, tax, discount
- Clear cart option

### 4.4 Checkout Panel
- Order summary
- Payment method selection:
  - **Tunai** (Cash) - existing
  - **QRIS** - existing
  - **Kartu** (Card) - **NEW**

---

## Phase 5: Web POS Frontend - Card Payment (NEW)

### 5.1 WebSocket Hook
**File:** `src/hooks/useWebSocket.ts`

**Features:**
- Connect on mount with terminal_id
- Auto-reconnect on disconnect
- Handle message types:
  - `PAYMENT_READY` → Show waiting screen
  - `PAYMENT_COMPLETED` → Show receipt
  - `PAYMENT_FAILED` → Show error, allow retry
- Ping/pong keepalive

### 5.2 Payment Store
**File:** `src/stores/paymentStore.ts`

**State:**
```typescript
interface PaymentState {
  status: 'idle' | 'waiting_for_tap' | 'completed' | 'failed';
  paymentId: string | null;
  orderId: string | null;
  amount: number;
  transactionId: string | null;
  error: string | null;
}
```

### 5.3 Card Payment Flow
**File:** `src/components/PaymentWaiting.tsx`

**UI States:**
1. **Initializing** → "Menghubungkan..."
2. **Waiting** → "Silakan tap kartu di HP Kasir" + animated icon
3. **Processing** → "Memproses pembayaran..."
4. **Success** → Auto-redirect to receipt
5. **Failed** → Show error + "Coba Lagi" button
6. **Expired** → "Waktu habis" + "Buat Ulang"

### 5.4 Integration with Checkout
**File:** `src/components/PaymentMethod.tsx`

**Flow:**
1. User selects "Kartu"
2. `POST /v1/card-payment/init` with order details
3. Response: `{ payment_id, status: "WAITING_FOR_TAP" }`
4. Switch to `PaymentWaiting` component
5. Listen WebSocket for `PAYMENT_COMPLETED`
6. On success → Navigate to receipt

---

## Phase 6: Companion App - Android

### 6.1 Project Structure
```
pos-payment-app/
├── app/
│   └── src/main/
│       ├── java/com/gis/gbs/payment/
│       │   ├── MainActivity.kt
│       │   ├── PaymentService.kt      # API calls
│       │   ├── NeurogineManager.kt     # SDK wrapper
│       │   └── model/
│       │       └── Payment.kt
│       └── AndroidManifest.xml
└── build.gradle.kts
```

### 6.2 Features
- **Device Registration** → POST `/v1/device/register`
- **Payment Polling** → GET `/v1/card-payment/pending?device_id=xxx`
- **Payment Notification** → Show pending payments
- **Neurogine Integration** → Call SDK on card tap
- **Result Callback** → POST `/v1/card-payment/:id/confirm`

### 6.3 UI Flow
```
1. App starts → Auto-register device
2. Main screen shows: "Menunggu pembayaran..."
3. Polling hits → Show: "Pembayaran Rp 150.000"
4. User taps card → Call Neurogine SDK
5. Success/Fail → Send callback
6. Return to waiting state
```

---

## Phase 7: Integration & Testing

### 7.1 Backend Testing
- WebSocket connection & messaging
- Card payment CRUD
- Terminal registration
- Payment expiry job

### 7.2 Frontend Testing
- Product catalog display
- Cart operations
- Checkout flow
- Card payment waiting screen
- Receipt display
- WebSocket reconnection

### 7.3 Companion App Testing
- Device registration
- Payment polling
- Neurogine SDK integration
- Result callback

### 7.4 E2E Testing
- Full flow: Add to cart → Card payment → Receipt

---

## File Summary

### Backend (`gbs-pos-api/`)
| File | Phase | Priority |
|------|-------|----------|
| `go.mod` (add websocket) | 1.1 | High |
| `internal/websocket/hub.go` | 1.2 | High |
| `internal/websocket/client.go` | 1.2 | High |
| `internal/websocket/messages.go` | 1.2 | High |
| `internal/model/card_payment.go` | 1.3 | High |
| `internal/service/card_payment_service.go` | 1.4 | High |
| `internal/handler/card_payment_handler.go` | 1.5 | High |
| `internal/router/websocket_route.go` | 1.6 | High |
| `internal/router/card_payment_route.go` | 1.6 | High |
| `internal/router/router.go` (modify) | 1.6 | High |
| `cmd/server/main.go` (modify) | 1.7 | High |
| `migrations/005_create_card_payments.sql` | 2.1 | High |

### Frontend (`pos-web/`)
| File | Phase | Priority |
|------|-------|----------|
| `package.json`, `vite.config.ts` | 3.1 | High |
| `src/hooks/useWebSocket.ts` | 5.1 | High |
| `src/stores/paymentStore.ts` | 5.2 | High |
| `src/components/PaymentWaiting.tsx` | 5.3 | High |
| `src/components/PaymentMethod.tsx` (modify) | 5.4 | High |
| `src/pages/POSPage.tsx` (modify) | 5.4 | High |
| `src/api/client.ts` | 4.1 | Medium |
| `src/components/ProductGrid.tsx` | 4.2 | Medium |
| `src/components/CartPanel.tsx` | 4.3 | Medium |
| `src/components/CheckoutPanel.tsx` | 4.4 | Medium |
| `src/components/ReceiptPanel.tsx` | 4.4 | Medium |

### Companion App (`pos-payment-app/`)
| File | Phase | Priority |
|------|-------|----------|
| `PaymentService.kt` | 6.2 | High |
| `NeurogineManager.kt` | 6.2 | High |
| `MainActivity.kt` | 6.3 | High |
| `AndroidManifest.xml` | 6.1 | Medium |

---

## Timeline Estimation

| Phase | Task | Estimation |
|-------|------|------------|
| 1 | Backend WebSocket | 2-3 days |
| 2 | Database Migration | 0.5 day |
| 3 | Frontend Setup | 1 day |
| 4 | Frontend Core | 2-3 days |
| 5 | Frontend Card Payment | 1-2 days |
| 6 | Companion App | 2-3 days |
| 7 | Integration & Testing | 2-3 days |
| **Total** | | **10-15 days** |

---

## Next Steps

1. ✅ Architecture design (done)
2. ⬜ Backend implementation (Phase 1-2)
3. ⬜ Frontend implementation (Phase 3-5)
4. ⬜ Companion App (Phase 6)
5. ⬜ Integration & Testing (Phase 7)

---

**Approved by:** _________________**Date:** __________________
