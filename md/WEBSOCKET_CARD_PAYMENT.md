# WebSocket & Card Payment Architecture

## Overview

Web POS (browser di Sunmi D3 Pro) memerlukan komunikasi real-time untuk card payment flow. Payment dilakukan di device terpisah (HP Android dengan NFC).

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Sunmi D3 Pro                             │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    Web POS (Browser)                         ││
│  │  - Product Grid                                              ││
│  │  - Cart Panel                                                ││
│  │  - Checkout                                                  ││
│  │  - Payment Card Waiting Screen                              ││
│  │  - Receipt Screen                                           ││
│  │                                                              ││
│  │  WebSocket: ws://host/ws?terminal_id=POS-001                ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                    HTTP POST /v1/card-payment/init
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Backend (gbs-pos-api)                         │
│                                                                  │
│  ┌──────────────────┐    ┌──────────────────────────────────────┐│
│  │  WebSocket Hub   │    │       Card Payment Service           ││
│  │  - Register      │    │                                      ││
│  │  - Unregister    │    │  - Create payment request            ││
│  │  - Broadcast     │    │  - Confirm payment                   ││
│  │  - Send to       │    │  - Cancel payment                    ││
│  │    terminal      │    │  - Handle Neurogine callback         ││
│  └──────────────────┘    └──────────────────────────────────────┘│
│                                                                  │
│  ┌──────────────────────────────────────────────────────────────┐│
│  │                    Database: card_payments                    ││
│  │  id, order_id, amount, status, device_id, transaction_id      ││
│  └──────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                    HTTP GET /v1/card-payment/pending
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Companion App (HP Android)                       │
│                                                                  │
│  - Polls pending payments                                       │
│  - Shows: "Pembayaran Rp XXX - Tap Kartu"                       │
│  - Calls Neurogine SoftPOS SDK                                  │
│  - Sends callback on completion                                 │
└─────────────────────────────────────────────────────────────────┘
```

## WebSocket Protocol

### Endpoint
```
ws://host:8080/ws?terminal_id=POS-001
```

### Message Format (JSON)

#### Server → Client

**Payment Ready (waiting for tap)**
```json
{
  "type": "PAYMENT_READY",
  "payment_id": "uuid",
  "order_id": "ORD-123",
  "amount": 150000,
  "expires_at": "2024-01-01T12:30:00Z"
}
```

**Payment Completed**
```json
{
  "type": "PAYMENT_COMPLETED",
  "payment_id": "uuid",
  "status": "SUCCESS",
  "transaction_id": "NEU-xxx",
  "card_brand": "VISA",
  "masked_card": "**** **** **** 1234"
}
```

**Payment Failed**
```json
{
  "type": "PAYMENT_FAILED",
  "payment_id": "uuid",
  "status": "FAILED",
  "reason": "Card declined"
}
```

#### Client → Server

**Ping (keepalive)**
```json
{
  "type": "PING"
}
```

**Pong (response)**
```json
{
  "type": "PONG"
}
```

## REST API Endpoints

### 1. Initialize Payment
```
POST /v1/card-payment/init
```

**Request:**
```json
{
  "order_id": "ORD-123",
  "amount": 150000,
  "device_id": "HP-KASIR-01"  // optional: specific device
}
```

**Response:**
```json
{
  "payment_id": "uuid",
  "status": "WAITING_FOR_TAP",
  "amount": 150000,
  "expires_at": "2024-01-01T12:30:00Z"
}
```

### 2. Confirm Payment (from Companion App)
```
POST /v1/card-payment/:id/confirm
```

**Request:**
```json
{
  "status": "SUCCESS",
  "transaction_id": "NEU-123456",
  "card_brand": "VISA",
  "masked_card": "**** **** **** 1234",
  "auth_code": "AUTH123"
}
```

**Response:**
```json
{
  "payment_id": "uuid",
  "status": "SUCCESS",
  "message": "Payment confirmed"
}
```

### 3. Get Pending Payments (for Companion App polling)
```
GET /v1/card-payment/pending?device_id=HP-KASIR-01
```

**Response:**
```json
{
  "payments": [
    {
      "payment_id": "uuid",
      "order_id": "ORD-123",
      "amount": 150000,
      "created_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

### 4. Cancel Payment
```
POST /v1/card-payment/:id/cancel
```

**Response:**
```json
{
  "payment_id": "uuid",
  "status": "CANCELLED"
}
```

## Database Schema

```sql
CREATE TABLE card_payments (
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
```

## Status Flow

```
WAITING_FOR_TAP → SUCCESS    (card tapped successfully)
WAITING_FOR_TAP → FAILED     (card declined/timeout)
WAITING_FOR_TAP → CANCELLED (cashier cancelled)
```

## Payment Expiry

- Default expiry: 5 minutes
- Background job cleans up expired payments
- On expiry: status → EXPIRED, notify via WebSocket

## Companion App Integration

### Polling Strategy
```
1. App opens → GET /v1/card-payment/pending
2. If empty → Poll every 3 seconds
3. If payment found → Show notification
4. User taps card → Call Neurogine SDK
5. On result → POST /v1/card-payment/:id/confirm
```

### Alternative: Push Notification (FCM)
```
1. Backend receives payment request
2. Send FCM push to device_id
3. App wakes up, shows payment
4. User taps card...
```

## Security

1. **Terminal Authentication**: WebSocket requires valid JWT token
2. **Device Registration**: Companion app registered with device_id
3. **Rate Limiting**: Prevent polling abuse
4. **Payment Signing**: Transaction ID signed by Neurogine

## Implementation Files

### Backend
| File | Purpose |
|------|---------|
| `internal/websocket/hub.go` | WebSocket connection hub |
| `internal/websocket/client.go` | Individual client connection |
| `internal/websocket/messages.go` | Message types |
| `internal/model/card_payment.go` | CardPayment model |
| `internal/handler/card_payment_handler.go` | REST endpoints |
| `internal/service/card_payment_service.go` | Business logic |
| `internal/router/card_payment_route.go` | Route definitions |
| `internal/router/websocket_route.go` | WebSocket upgrade route |

### Frontend (Web POS)
| File | Purpose |
|------|---------|
| `hooks/useWebSocket.ts` | WebSocket connection hook |
| `hooks/useCardPayment.ts` | Card payment state machine |
| `components/PaymentWaiting.tsx` | Waiting screen |
| `pages/POSPage.tsx` | Main POS with WebSocket integration |

### Companion App
| File | Purpose |
|------|---------|
| `PaymentService` | API calls to backend |
| `NeurogineManager` | SDK wrapper |
| `MainActivity` | Payment polling & UI |
