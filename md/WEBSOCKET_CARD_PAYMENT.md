# WEBSOCKET CARD PAYMENT ARCHITECTURE

## Overview

Revisi arsitektur card payment untuk menghilangkan mekanisme HTTP Polling pada Companion App. Implementasi menggunakan WebSocket persisten dua arah antara **Web POS ↔ Backend ↔ Companion App**.

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              Sunmi D3 Pro (Browser)                                  │
│  ┌───────────────────────────────────────────────────────────────────────────────┐  │
│  │                              Web POS                                            │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │  │
│  │  │   Product   │  │     Cart    │  │  Checkout   │  │  Payment Waiting     │ │  │
│  │  │    Grid     │  │    Panel    │  │   Screen    │  │  / Receipt Screen    │ │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │  │
│  │                                                                               │  │
│  │  ═══════════════════════════════════════════════════════════════════════════   │  │
│  │  WebSocket Client (Persistent Connection)                                      │  │
│  │  ws://host:8080/ws?terminal_id=POS-001                                        │  │
│  └───────────────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │ ▲
                                        │ │ WebSocket (JSON)
                                        ▼ │
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                           Backend (gbs-pos-api)                                       │
│                                                                                       │
│  ┌───────────────────────────────────────────────────────────────────────────────┐   │
│  │                         WebSocket Hub (Bidirectional)                            │   │
│  │  ┌────────────────┐  ┌────────────────┐  ┌────────────────────────────────┐  │   │
│  │  │  POS Clients   │  │ Companion      │  │  Message Router               │  │   │
│  │  │  (terminal_id) │  │ Clients        │  │  - PAYMENT_REQUEST → Companion │  │   │
│  │  └────────────────┘  │ (device_id)    │  │  - PAYMENT_STATUS → POS       │  │   │
│  │                      └────────────────┘  └────────────────────────────────┘  │   │
│  └───────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                           │
│  ┌─────────────────────────────────────▼───────────────────────────────────────┐   │
│  │                      Card Payment Service                                      │   │
│  │  - CreatePayment()      → buat payment record, broadcast ke Companion        │   │
│  │  - UpdateStatus()       → update DB, broadcast ke POS via WebSocket          │   │
│  │  - ExpirePayments()     → cleanup expired payments                          │   │
│  └───────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                           │
│  ┌─────────────────────────────────────▼───────────────────────────────────────┐   │
│  │                        Database: card_payments                                 │   │
│  │  id, order_id, amount, status, terminal_id, device_id, transaction_id,          │   │
│  │  card_brand, masked_card, auth_code, failure_reason, expires_at                │   │
│  └───────────────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │ ▲
                                        │ │ WebSocket (JSON)
                                        ▼ │
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                        Companion App (Android)                                        │
│                                                                                       │
│  ┌───────────────────────────────────────────────────────────────────────────────┐   │
│  │                      WebSocket Client (Persistent)                             │   │
│  │  ws://host:8080/ws?device_id=HP-KASIR-01                                     │   │
│  │                                                                               │   │
│  │  ═════════════════════════════════════════════════════════════════════════   │   │
│  │  Connection State: CONNECTED / DISCONNECTED / RECONNECTING                    │   │
│  │  Auto-reconnect with exponential backoff                                      │   │
│  └───────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                           │
│  ┌─────────────────────────────────────▼───────────────────────────────────────┐   │
│  │                      Payment State Machine                                     │   │
│  │                                                                               │   │
│  │  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌───────────┐  │   │
│  │  │WAITING_FOR_  │───▶│ PROCESSING   │───▶│   SUCCESS    │    │  FAILED   │  │   │
│  │  │    CARD      │    │              │    │              │    │           │  │   │
│  │  └──────────────┘    └──────────────┘    └──────────────┘    └───────────┘  │   │
│  │         │                   │                                           │      │   │
│  │         │                   └───────────────────────────────────────────┘      │   │
│  │         ▼                                                                ▼      │   │
│  │  ┌──────────────┐                                                  ┌─────────┐ │   │
│  │  │  CANCELLED   │                                                  │ EXPIRED │ │   │
│  │  └──────────────┘                                                  └─────────┘ │   │
│  └───────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                           │
│  ┌─────────────────────────────────────▼───────────────────────────────────────┐   │
│  │                      Neurogine SoftPOS SDK Integration                        │   │
│  │  - Initialize SDK on app start                                               │   │
│  │  - ProcessCardPayment(amount) → returns transaction result                   │   │
│  │  - Handle card tap, PIN entry, NFC communication                           │   │
│  └───────────────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Flow Diagram: Complete Payment Cycle

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              FLOW: CARD PAYMENT VIA WEBSOCKET                           │
└─────────────────────────────────────────────────────────────────────────────────────────┘

    Web POS                      Backend                          Companion App
    ───────                      ───────                          ──────────────
       │                            │                                   │
       │  1. User pilih "Kartu"     │                                   │
       │  POST /v1/card-payment/init│                                   │
       │──────────────────────────▶│                                   │
       │                            │                                   │
       │                            │  2. Create card_payment record     │
       │                            │     status = WAITING_FOR_CARD     │
       │                            │──────────────────────────────────▶│
       │                            │                                   │
       │                            │  3. PAYMENT_REQUEST event         │
       │                            │     via WebSocket                 │
       │                            │──────────────────────────────────▶│
       │                            │                                   │
       │                            │                                   │ 4. Receive event
       │                            │                                   │ 5. Show UI:
       │                            │                                   │    "Tap kartu untuk
       │                            │                                   │     bayar Rp XXX"
       │                            │                                   │
       │  6. PAYMENT_STATUS event   │                                   │
       │     (WAITING_FOR_CARD)     │                                   │
       │◀──────────────────────────│                                   │
       │                            │                                   │
       │                            │                                   │ 7. User tap kartu
       │                            │                                   │ 8. Call Neurogine SDK
       │                            │                                   │ 9. Process payment
       │                            │                                   │
       │                            │ 10. PAYMENT_STATUS update         │
       │                            │     (PROCESSING)                  │
       │                            │◀──────────────────────────────────│
       │  11. PAYMENT_STATUS event  │                                   │
       │◀───────────────────────────│                                   │
       │                            │                                   │
       │  Show "Memproses..."       │                                   │
       │                            │                                   │
       │                            │                                   │ 12. SDK returns
       │                            │                                   │     result
       │                            │                                   │
       │                            │ 13. PAYMENT_STATUS update         │
       │                            │     (SUCCESS/FAILED)              │
       │                            │◀──────────────────────────────────│
       │                            │                                   │
       │  14. PAYMENT_STATUS event  │                                   │
       │     (SUCCESS/FAILED)       │                                   │
       │◀───────────────────────────│                                   │
       │                            │                                   │
       │  15a. SUCCESS:             │                                   │
       │     Show Receipt Screen    │                                   │
       │                            │                                   │
       │  15b. FAILED:              │                                   │
       │     Show error + Retry     │                                   │
       │                            │                                   │
       │                            │                                   │


═══════════════════════════════════════════════════════════════════════════════════════════

                              TIMEOUT / EXPIRY FLOW

    Web POS                      Backend                          Companion App
       │                            │                                   │
       │  T.1 PAYMENT_REQUEST expired│                                   │
       │◀───────────────────────────│                                   │
       │                            │                                   │
       │  Show "Waktu habis"        │                                   │
       │                            │                                   │
```

## WebSocket Message Protocol

### Endpoint
```
ws://host:8080/ws?client_type=pos&terminal_id=POS-001
ws://host:8080/ws?client_type=companion&device_id=HP-KASIR-01
```

### Connection Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `client_type` | string | Yes | `pos` atau `companion` |
| `terminal_id` | string | Yes* | Terminal ID (untuk pos) |
| `device_id` | string | Yes* | Device ID (untuk companion) |

*Required based on `client_type`

---

### Message Types: Server → Client

#### PAYMENT_REQUEST (Server → Companion)
Dikirim saat ada payment request baru dari Web POS.

```json
{
  "type": "PAYMENT_REQUEST",
  "payment_id": "550e8400-e29b-41d4-a716-446655440000",
  "order_id": "ORD-20260806-001",
  "amount": 150000,
  "currency": "IDR",
  "expires_at": "2026-08-06T12:35:00Z",
  "merchant_name": "GBS Store",
  "terminal_id": "POS-001"
}
```

#### PAYMENT_STATUS (Server → POS)
Dikirim saat status payment berubah.

```json
{
  "type": "PAYMENT_STATUS",
  "payment_id": "550e8400-e29b-41d4-a716-446655440000",
  "order_id": "ORD-20260806-001",
  "status": "WAITING_FOR_CARD",
  "message": "Silakan tap kartu di Companion App",
  "amount": 150000,
  "transaction_id": null,
  "card_brand": null,
  "masked_card": null,
  "auth_code": null,
  "failure_reason": null,
  "updated_at": "2026-08-06T12:30:00Z"
}
```

Status Values:
- `WAITING_FOR_CARD` - Menunggu customer tap kartu
- `PROCESSING` - Sedang diproses oleh SDK
- `SUCCESS` - Payment berhasil
- `FAILED` - Payment gagal
- `CANCELLED` - Payment dibatalkan oleh kasir
- `EXPIRED` - Payment expired

#### CONNECTION_ACK (Server → Client)
Konfirmasi koneksi berhasil.

```json
{
  "type": "CONNECTION_ACK",
  "client_type": "pos",
  "client_id": "POS-001",
  "session_id": "session-uuid",
  "server_time": "2026-08-06T12:30:00Z"
}
```

#### PONG (Server → Client)
Response untuk ping keepalive.

```json
{
  "type": "PONG"
}
```

---

### Message Types: Client → Server

#### PAYMENT_STATUS_UPDATE (Companion → Server)
Update status dari Companion App setelah processing.

```json
{
  "type": "PAYMENT_STATUS_UPDATE",
  "payment_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "SUCCESS",
  "transaction_id": "NEU-20260806-123456",
  "card_brand": "VISA",
  "masked_card": "**** **** **** 1234",
  "auth_code": "AUTH123456",
  "failure_reason": null
}
```

#### PING (Client → Server)
Keepalive ping.

```json
{
  "type": "PING"
}
```

#### REGISTER_COMPANION (Companion → Server)
Registration untuk companion app (optional, bisa via query param).

```json
{
  "type": "REGISTER_COMPANION",
  "device_id": "HP-KASIR-01",
  "device_name": "Samsung Galaxy A54",
  "sdk_version": "1.2.3",
  "capabilities": ["NFC", "BLE"]
}
```

---

## REST API Endpoints

### Card Payment Initialization
```
POST /v1/card-payment/init
```

**Request:**
```json
{
  "order_id": "ORD-20260806-001",
  "amount": 150000
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "payment_id": "550e8400-e29b-41d4-a716-446655440000",
    "order_id": "ORD-20260806-001",
    "amount": 150000,
    "status": "WAITING_FOR_CARD",
    "expires_at": "2026-08-06T12:35:00Z"
  }
}
```

### Get Payment Status
```
GET /v1/card-payment/:id
```

**Response:**
```json
{
  "success": true,
  "data": {
    "payment_id": "550e8400-e29b-41d4-a716-446655440000",
    "order_id": "ORD-20260806-001",
    "amount": 150000,
    "status": "SUCCESS",
    "transaction_id": "NEU-20260806-123456",
    "card_brand": "VISA",
    "masked_card": "**** **** **** 1234",
    "auth_code": "AUTH123456",
    "created_at": "2026-08-06T12:30:00Z",
    "updated_at": "2026-08-06T12:31:00Z"
  }
}
```

### Cancel Payment
```
POST /v1/card-payment/:id/cancel
```

**Response:**
```json
{
  "success": true,
  "data": {
    "payment_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "CANCELLED"
  }
}
```

---

## Database Schema

```sql
CREATE TABLE card_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'WAITING_FOR_CARD',
    terminal_id VARCHAR(50) NOT NULL,
    device_id VARCHAR(50),
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
CREATE INDEX idx_card_payments_terminal ON card_payments(terminal_id);
CREATE INDEX idx_card_payments_device ON card_payments(device_id);
CREATE INDEX idx_card_payments_expires ON card_payments(expires_at);
```

---

## Status Flow

```
┌──────────────┐
│WAITING_FOR_  │
│    CARD      │
└──────┬───────┘
       │
       ├──────────────────────────────────┐
       │                                  │
       ▼                                  ▼
┌──────────────┐                  ┌──────────────┐
│ PROCESSING   │                  │  CANCELLED   │
└──────┬───────┘                  └──────────────┘
       │
       ├───────────────────────────┐
       │                           │
       ▼                           ▼
┌──────────────┐           ┌──────────────┐
│   SUCCESS    │           │    FAILED    │
└──────────────┘           └──────────────┘

Tiap status WAITING_FOR_CARD memiliki expiry timeout (default: 5 menit)
Jika expired → status berubah ke EXPIRED
```

---

## Companion App States

### State Machine
```kotlin
sealed class PaymentState {
    object Idle : PaymentState()
    data class WaitingForCard(val payment: PaymentRequest) : PaymentState()
    object Processing : PaymentState()
    data class Success(val transaction: TransactionResult) : PaymentState()
    data class Failed(val reason: String) : PaymentState()
    object Cancelled : PaymentState()
    object Expired : PaymentState()
}
```

### State Transitions
```
Idle ─────────────────────────────▶ WaitingForCard
    │                                     │
    │ (PAYMENT_REQUEST received)          │ (User tap card / timeout)
    │                                     ▼
    │                              Processing ───────▶ Success
    │                                     │              │
    │                                     │              │ (SDK result)
    │                                     │              │
    │                                     ▼              │
    │                              Failed ◀───────┘
    │                                │
    │                                │
    │◀───────────────────────────────┘
    │         (Error / Timeout)
    │
    │ (Cancel from POS)
    ▼
Cancelled

Any ──────────────────────────────▶ Expired
    │                                  ▲
    │ (Expiry timeout)                  │ (Payment expired)
    │                                  │
    └──────────────────────────────────┘
```

---

## Security Considerations

1. **WebSocket Authentication**
   - JWT token validation on connection
   - Token passed via query param: `?token=<jwt>`
   - Backend validates and extracts client identity

2. **Client Type Validation**
   - `client_type` determines routing rules
   - POS clients only receive PAYMENT_STATUS updates for their terminal_id
   - Companion clients only receive PAYMENT_REQUEST for their device_id

3. **Rate Limiting**
   - Max 10 connections per device_id
   - Max 100 messages per minute per client
   - Connection timeout: 30 seconds for handshake

4. **Payment Security**
   - Transaction signing by Neurogine SDK
   - Backend validates transaction_id format
   - No sensitive card data stored (only masked card)

---

## Reconnection Strategy

### Companion App
```
1. Initial connection attempt
2. On disconnect:
   - Increment retry count
   - Wait: min(30s, 2^retry * 1s) — exponential backoff, max 30s
   - Attempt reconnect
3. On reconnect:
   - Re-register device_id
   - Request pending payments via REST API
4. Max retries: unlimited (always try to reconnect)
```

### Web POS
```
1. Initial connection on page load
2. On disconnect:
   - Show "Connection lost" indicator
   - Retry every 3 seconds
   - After 5 retries, show "Refresh page" button
3. On reconnect:
   - Request current payment status via REST API
   - Resume normal operation
```

---

## Error Handling

### Backend Errors
| Error Code | Description | Action |
|------------|-------------|--------|
| `INVALID_TOKEN` | JWT validation failed | Disconnect, prompt re-login |
| `INVALID_CLIENT_TYPE` | Unknown client type | Disconnect |
| `PAYMENT_NOT_FOUND` | Payment ID not found | Log error, ignore |
| `PAYMENT_EXPIRED` | Payment already expired | Notify via WebSocket |
| `RATE_LIMITED` | Too many messages | Slow down, disconnect if persistent |

### SDK Errors (Companion App)
| Error Code | Description | Action |
|------------|-------------|--------|
| `CARD_DECLINED` | Card rejected by bank | Send FAILED status |
| `CARD_TIMEOUT` | No card detected | Show retry option |
| `INVALID_CARD` | Unreadable card | Show error message |
| `SDK_ERROR` | Neurogine SDK error | Send FAILED with reason |

---

## Monitoring & Logging

### Metrics to Track
- Active WebSocket connections (by type)
- Messages per second
- Payment success/fail rate
- Average payment processing time
- Connection failures per device

### Log Events
```
[WS] Connection opened: client_type=pos, terminal_id=POS-001
[WS] Connection closed: client_type=companion, device_id=HP-KASIR-01, reason=timeout
[PAYMENT] Created: payment_id=xxx, order_id=xxx, amount=xxx
[PAYMENT] Status changed: payment_id=xxx, old=WAITING_FOR_CARD, new=PROCESSING
[PAYMENT] Completed: payment_id=xxx, status=SUCCESS, duration=2.3s
[PAYMENT] Expired: payment_id=xxx
```

---

## Migration from Polling

### Old Flow (HTTP Polling)
```
Companion App                    Backend
     │                              │
     │  GET /v1/card-payment/pending
     │─────────────────────────────▶│
     │◀─────────────────────────────│ 200: { payments: [] }
     │  (poll every 3 seconds)      │
```

### New Flow (WebSocket)
```
Companion App                    Backend                    Web POS
     │                              │                          │
     │  WS connect                   │                          │
     │─────────────────────────────▶│                          │
     │◀─────────────────────────────│ CONNECTION_ACK           │
     │                              │                          │
     │                              │  POST /v1/card-payment/init
     │                              │◀─────────────────────────│
     │                              │  PAYMENT_REQUEST         │
     │  PAYMENT_REQUEST             │─────────────────────────▶│
     │◀─────────────────────────────│                          │
     │  (instant, no polling)        │                          │
```

---

## File Structure

```
gbs-pos-api/
├── internal/
│   ├── websocket/                 # WebSocket infrastructure
│   │   ├── hub.go                 # Central hub for all connections
│   │   ├── client.go              # Individual client connection
│   │   ├── messages.go            # Message type definitions
│   │   └── router.go              # Message routing logic
│   ├── cardpayment/               # Card payment module (BARU)
│   │   ├── model.go               # CardPayment model
│   │   ├── service.go             # CardPaymentService
│   │   ├── handler.go             # REST endpoints
│   │   └── repository.go          # Database operations
│   └── router/
│       ├── websocket_route.go    # WebSocket endpoint (BARU)
│       ├── card_payment_route.go  # Card payment routes (BARU)
│       └── router.go              # Updated to include new routes
├── cmd/server/
│   └── main.go                    # Updated to initialize WebSocket hub
└── migrations/
    └── 015_create_card_payments.sql (BARU)

pos-web/
├── src/
│   ├── hooks/
│   │   └── useWebSocket.ts        # WebSocket client hook (BARU)
│   ├── components/
│   │   └── PaymentWaiting.tsx     # Payment waiting screen (UPDATE)
│   └── stores/
│       └── paymentStore.ts        # Payment state (UPDATE)

pos-payment-app/ (Android)
├── app/src/main/java/com/gbs/pos/
│   ├── PaymentWebSocketClient.kt  # WebSocket client (BARU)
│   ├── PaymentStateMachine.kt     # State machine (BARU)
│   ├── NeurogineManager.kt        # SDK wrapper (UPDATE)
│   └── MainActivity.kt            # Main UI (UPDATE)
└── build.gradle.kts
```

---

## Implementation Checklist

- [ ] 1. Add gorilla/websocket dependency
- [ ] 2. Create WebSocket hub (hub.go)
- [ ] 3. Create WebSocket client (client.go)
- [ ] 4. Create message types (messages.go)
- [ ] 5. Create message router (router.go)
- [ ] 6. Create CardPayment model
- [ ] 7. Create CardPayment repository
- [ ] 8. Create CardPayment service
- [ ] 9. Create CardPayment handler
- [ ] 10. Create WebSocket route
- [ ] 11. Create card payment route
- [ ] 12. Update main.go with WebSocket integration
- [ ] 13. Create migration 015_create_card_payments.sql
- [ ] 14. Update Web POS useWebSocket hook
- [ ] 15. Update Web POS PaymentWaiting component
- [ ] 16. Create Companion App WebSocket client
- [ ] 17. Create Companion App state machine
- [ ] 18. Update Companion App UI
- [ ] 19. Test end-to-end flow
- [ ] 20. Update documentation

---

**Document Version:** 2.0  
**Last Updated:** August 2026  
**Status:** Ready for Implementation
