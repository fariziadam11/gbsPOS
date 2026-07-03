# GBS POS-CMS API

Backend monorepo untuk sistem **Point of Sale (POS)** dan **Content Management System (CMS)** GBS,
dibangun dengan Go. Mendukung Android POS Sunmi D3 Pro dengan offline-first, multi-toko, diskon,
manajemen stok, dan streaming video iklan.

---

## Daftar Isi

- [Status Proyek](#status-proyek)
- [Apa yang Baru](#apa-yang-baru)
- [Arsitektur](#arsitektur)
- [Tech Stack](#tech-stack)
- [Struktur Monorepo](#struktur-monorepo)
- [Database Schema](#database-schema)
- [API Endpoints](#api-endpoints)
- [Cara Menjalankan](#cara-menjalankan)
- [Environment Variables](#environment-variables)
- [Akun Default](#akun-default)
- [Keycloak Authentication](#keycloak-authentication)
- [Testing](#testing)
- [CI/CD Pipeline](#cicd-pipeline)
- [Branch & Kontributor](#branch--kontributor)

---

## Status Proyek

| Komponen | Status | Keterangan |
|----------|--------|------------|
| **POS API** | ✅ Running | Port 8080 |
| **CMS API** | ✅ Running | Port 8081 |
| **CMS Web (Vue 3)** | ✅ Built | Admin panel lengkap |
| **PostgreSQL** | ✅ Running | Port 5433 |
| **Database Migrations** | ✅ 14 migrations | Termasuk stok, diskon, varian, hold |
| **Keycloak Auth** | ✅ Configured | RS256 + fallback demo HS256 |
| **Unit Tests** | ✅ Tersedia | SQLite in-memory |
| **Docker Compose** | ✅ Configured | Multi-container |
| **CI/CD Pipeline** | ✅ Configured | GitHub Actions + GitLab CI |

> **Go**: 1.26.1 | **PostgreSQL**: 15 | **Vue**: 3.5.34

---

## Apa yang Baru

Update terbaru dari tim (branch `main`) menambahkan fitur-fitur besar:

### POS API — Fitur Baru
| Fitur | Deskripsi |
|-------|-----------|
| **Manajemen Pelanggan** | CRUD customers, loyalty points, link ke orders |
| **Manajemen Stok** | `stock_quantity`, `low_stock_threshold`, audit trail `stock_movements` |
| **Diskon** | Diskon per produk (PERCENTAGE/FIXED), scheduling start/end date, status engine |
| **Varian Produk** | Multi-dimensi dengan `attributes` JSONB, SKU, harga & stok per varian |
| **Hold Session** | Simpan cart sementara ke `pos_hold_sessions` (UUID, JSONB payload) |
| **Pricing Engine** | Kalkulasi harga akhir dengan diskon aktif |
| **Dashboard** | Summary statistik: revenue, orders, produk terlaris |
| **Order Discount** | Field `discount_type`, `discount_value`, `discount_amount` di tabel orders |
| **Keycloak Auth** | Middleware RS256 + `ENABLE_DEMO_AUTH` fallback |

### CMS API — Fitur Baru
| Fitur | Deskripsi |
|-------|-----------|
| **User Management** | CRUD users via CMS (`/v1/users`) |
| **Settings** | Endpoint konfigurasi sistem (`/v1/settings`) |
| **Keycloak Integration** | Middleware `keycloak.go` di `gbs-common` |

### CMS Web — Fitur Baru
| Halaman | Deskripsi |
|---------|-----------|
| `ProductsView.vue` | CRUD produk dengan filter storeType |
| `OrdersView.vue` | Riwayat order dengan filter lengkap |
| `CustomersView.vue` | Manajemen pelanggan |
| `DiscountsView.vue` | Manajemen diskon per produk |
| `UsersView.vue` | Manajemen pengguna |
| `SettingsView.vue` | Konfigurasi sistem |
| `AdsView.vue` | Manajemen iklan (dedicated page) |
| `AuthCallbackView.vue` | Handler OIDC callback dari Keycloak |

### Migrations Baru
| Migration | Isi |
|-----------|-----|
| `009_seed_more_products` | Seed produk tambahan |
| `010_add_stock_and_customers` | Stok produk + tabel customers |
| `011_create_discounts` | Tabel discounts |
| `012_add_variants_discounts` | Tabel product_variants + kolom diskon di orders |
| `013_create_pos_hold_sessions` | Tabel hold sessions |
| `014_expand_discounts_scope` | Perluasan scope diskon |

---

## Arsitektur

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                             │
│  📱 Android POS App      🌐 CMS Web (Vue 3)                      │
└──────────────┬──────────────────────┬───────────────────────────┘
               │                      │
        ┌──────▼──────┐        ┌──────▼──────┐
        │  POS API    │        │  CMS API    │
        │  :8080      │        │  :8081      │
        └──────┬──────┘        └──────┬──────┘
               │                      │
        ┌──────▼──────────────────────▼──────┐
        │         gbs-common (shared)         │
        │  JWT/Keycloak · CORS · Logger       │
        └───────────────────┬────────────────┘
                            │
               ┌────────────▼────────────┐
               │     PostgreSQL 15        │
               │     DB: gbs_pos          │
               └─────────────────────────┘
                            │
               ┌────────────▼────────────┐
               │  Keycloak (optional)     │
               │  auth.armmada.id         │
               └─────────────────────────┘
```

**Monorepo terdiri dari:**
- `gbs-pos-api/` — POS REST API (port 8080)
- `gbs-cms-api/` — CMS REST API (port 8081)
- `gbs-common/` — Shared middleware & utilities
- `cms-web/` — Vue 3 Admin Panel (port 5173 dev)
- `migrations/` — SQL migrations (shared)

---

## Tech Stack

### Backend (Go)
| Library | Versi | Fungsi |
|---------|-------|--------|
| `gin-gonic/gin` | v1.12.0 | HTTP Framework |
| `gorm.io/gorm` | v1.31.1 | ORM |
| `golang-jwt/jwt/v5` | v5.3.1 | JWT Auth (HS256 demo) |
| `golang-migrate/migrate/v4` | v4.19.1 | DB Migration |
| `caarlos0/env/v10` | v10.0.0 | Config dari env |
| `rs/zerolog` | v1.35.1 | Structured Logging |
| `go-playground/validator/v10` | v10.30.1 | Input Validation |
| `golang.org/x/crypto` | v0.51.0 | bcrypt passwords |
| `glebarez/sqlite` | v1.11.0 | SQLite untuk test |

### Frontend (CMS Web)
| Library | Versi |
|---------|-------|
| Vue 3 | v3.5.34 |
| PrimeVue | v4.5.5 |
| Vue Router | v4 |
| Pinia | v3.0.4 |
| TanStack Query | v5.100.11 |
| Axios | v1.16.1 |
| Vite | v8.0.12 |
| TypeScript | v6.0.2 |

### Auth
| Komponen | Keterangan |
|----------|------------|
| **Keycloak** | RS256, PKCE, realm `gbs-pos` |
| **Demo fallback** | HS256 JWT via `ENABLE_DEMO_AUTH=true` |
| **Android** | AppAuth + PKCE flow |
| **CMS Web** | `keycloak.ts` OIDC client |

---

## Struktur Monorepo

```
gbs-pos-cms-api/
│
├── gbs-pos-api/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/           # Env config
│   │   ├── database/         # Connect, seed, test helper
│   │   ├── dto/              # 10+ DTO structs
│   │   │   ├── customer_dto.go
│   │   │   ├── dashboard_dto.go
│   │   │   ├── discount_dto.go
│   │   │   ├── hold_dto.go
│   │   │   ├── pricing_dto.go
│   │   │   ├── product_dto.go
│   │   │   └── variant_dto.go
│   │   ├── handler/          # HTTP handlers
│   │   │   ├── customer_handler.go   (BARU)
│   │   │   ├── dashboard_handler.go  (BARU)
│   │   │   ├── discount_handler.go   (BARU)
│   │   │   ├── hold_handler.go       (BARU)
│   │   │   └── product_variant_handler.go (BARU)
│   │   ├── model/model.go    # GORM models (diperluas)
│   │   ├── repository/       # DB queries
│   │   │   ├── customer_repo.go      (BARU)
│   │   │   ├── dashboard_repo.go     (BARU)
│   │   │   ├── discount_repo.go      (BARU)
│   │   │   ├── hold_repo.go          (BARU)
│   │   │   ├── product_variant_repo.go (BARU)
│   │   │   └── stock_movement_repo.go  (BARU)
│   │   ├── router/           # Routes
│   │   │   ├── customer_route.go     (BARU)
│   │   │   ├── dashboard_route.go    (BARU)
│   │   │   ├── discount_route.go     (BARU)
│   │   │   ├── hold_route.go         (BARU)
│   │   │   └── variant_route.go      (BARU)
│   │   └── service/          # Business logic
│   │       ├── customer_service.go   (BARU)
│   │       ├── dashboard_service.go  (BARU)
│   │       ├── discount_service.go   (BARU)
│   │       ├── hold_service.go       (BARU)
│   │       ├── pricing_service.go    (BARU)
│   │       └── product_variant_service.go (BARU)
│   └── go.mod
│
├── gbs-cms-api/
│   ├── internal/
│   │   ├── dto/
│   │   │   ├── settings_dto.go  (BARU)
│   │   │   └── user_dto.go      (BARU)
│   │   ├── handler/
│   │   │   ├── settings_handler.go (BARU)
│   │   │   └── user_handler.go     (BARU)
│   │   ├── model/settings.go       (BARU)
│   │   ├── repository/
│   │   │   └── settings_repo.go    (BARU)
│   │   └── service/
│   │       ├── settings_service.go (BARU)
│   │       └── user_service.go     (BARU)
│   └── go.mod
│
├── gbs-common/
│   └── middleware/
│       ├── auth.go           # JWT HS256
│       ├── keycloak.go       # Keycloak RS256 (BARU)
│       ├── cors.go
│       └── logger.go
│
├── cms-web/
│   ├── src/
│   │   ├── api/              # API clients (8 files)
│   │   ├── composables/      # 7 composables
│   │   ├── stores/auth.ts    # Pinia + Keycloak
│   │   ├── keycloak.ts       # OIDC config (BARU)
│   │   └── views/            # 12 views
│   ├── Dockerfile            (BARU)
│   └── nginx.conf            (BARU)
│
├── migrations/               # 14 SQL migrations
├── md/                       # Dokumentasi dipindah ke sini
│   ├── KEYCLOAK_SETUP.md     (BARU)
│   ├── WEBSITE_DEPLOYMENT.md (BARU)
│   └── AGENTS.md, API_ENDPOINTS.md, dst.
├── docker-compose.yml
├── docker-compose.prod.yml
└── Makefile
```

---

## Database Schema

### Tabel Lengkap (14 migrations)

```
users                 products               orders
─────────────         ──────────────────     ─────────────────────
id                    id                     id (string, UUID8)
username              name                   subtotal / tax / total
password_hash         price                  payment_method
name                  category               cash_received / change_amount
role                  image_url              timestamp
gender                store_type             is_voided / is_settled
                      stock_quantity  ←NEW   store_type / terminal_id
                      low_stock_threshold←NEW customer_id    ←NEW
                                             loyalty_points_earned ←NEW
                                             discount_type  ←NEW
                                             discount_value ←NEW
                                             discount_amount←NEW
                                             transaction_id (Neurogine)

order_items           customers              discounts
────────────          ────────────           ──────────────────
id                    id                     id
order_id (FK)         name                   product_id (FK)
product_id            phone (unique)         name
product_name          email                  type (PERCENTAGE/FIXED)
product_price         address                value
qty / subtotal        loyalty_points         start_date / end_date
variant_id   ←NEW     created_at             status (PENDING/ACTIVE/...)
variant_name ←NEW
sku          ←NEW

product_variants      stock_movements        pos_hold_sessions
─────────────────     ───────────────        ─────────────────────
id                    id                     id (UUID PK)
product_id (FK)       product_id (FK)        store_type
sku                   type (IN/OUT/ADJ)      terminal_id
name                  quantity               payload (JSONB)
attributes (JSONB)    reason                 total
price                 reference_id           status (ACTIVE/RESUMED/EXPIRED)
stock_quantity        created_by
is_active             created_at
sort_order

settlements           ads                    ad_play_logs
───────────           ────────────────       ────────────
id                    id                     id
timestamp             name / file_path       ad_id (FK)
batch_count           file_size / duration   terminal_id
total/card/qris/cash  store_types (JSON)     played_at
status                playlist_order
store_type            start/end date+time
terminal_id           is_active / play_count
```

### Migrations History

| # | Migration | Novum |
|---|-----------|-------|
| 001 | create_users | Tabel users |
| 002 | create_products | Tabel products |
| 003 | create_orders | Tabel orders |
| 004 | create_order_items | Tabel order items |
| 005 | create_settlements | Tabel settlements |
| 006 | create_ads | Tabel ads |
| 007 | create_ad_play_logs | Analytics ads |
| 008 | seed_data | Admin + Cashier + produk |
| 009 | seed_more_products | Produk tambahan |
| **010** | **add_stock_and_customers** | **Stok produk + tabel customers** |
| **011** | **create_discounts** | **Tabel discounts** |
| **012** | **add_variants_discounts** | **Varian produk + diskon di orders** |
| **013** | **create_pos_hold_sessions** | **Hold session cart** |
| **014** | **expand_discounts_scope** | **Perluasan scope diskon** |

---

## API Endpoints

### POS API (`http://localhost:8080`)

#### Auth
| Method | Endpoint | Auth | Role |
|--------|----------|------|------|
| `POST` | `/v1/login` | No | All |

#### Products + Variants
| Method | Endpoint | Role |
|--------|----------|------|
| `GET` | `/v1/products` | All |
| `POST` | `/v1/products` | ADMIN |
| `PUT` | `/v1/products/:id` | ADMIN |
| `DELETE` | `/v1/products/:id` | ADMIN |
| `GET` | `/v1/products/:id/variants` | All |
| `POST` | `/v1/products/:id/variants` | ADMIN |
| `PUT` | `/v1/products/:id/variants/:vid` | ADMIN |
| `DELETE` | `/v1/products/:id/variants/:vid` | ADMIN |

#### Orders
| Method | Endpoint | Role |
|--------|----------|------|
| `GET` | `/v1/orders` | All |
| `GET` | `/v1/orders/:id` | All |
| `POST` | `/v1/orders` | All |
| `POST` | `/v1/sync/orders` | All |
| `PATCH` | `/v1/orders/:id/void` | ADMIN |
| `GET` | `/v1/orders/unsettled/summary` | All |
| `POST` | `/v1/orders/settle` | ADMIN |

#### Settlements
| Method | Endpoint | Role |
|--------|----------|------|
| `GET` | `/v1/settlements` | All |
| `GET` | `/v1/settlements/:id` | All |

#### Customers *(BARU)*
| Method | Endpoint | Role |
|--------|----------|------|
| `GET` | `/v1/customers` | All |
| `POST` | `/v1/customers` | ADMIN |
| `PUT` | `/v1/customers/:id` | ADMIN |
| `DELETE` | `/v1/customers/:id` | ADMIN |

#### Discounts *(BARU)*
| Method | Endpoint | Role |
|--------|----------|------|
| `GET` | `/v1/discounts` | All |
| `POST` | `/v1/discounts` | ADMIN |
| `PUT` | `/v1/discounts/:id` | ADMIN |
| `DELETE` | `/v1/discounts/:id` | ADMIN |
| `POST` | `/v1/discounts/:id/stop` | ADMIN |
| `GET` | `/v1/discounts/active` | All |

#### Hold Sessions *(BARU)*
| Method | Endpoint | Role |
|--------|----------|------|
| `GET` | `/v1/holds` | All |
| `POST` | `/v1/holds` | All |
| `GET` | `/v1/holds/:id` | All |
| `PUT` | `/v1/holds/:id/resume` | All |
| `DELETE` | `/v1/holds/:id` | All |

#### Dashboard *(BARU)*
| Method | Endpoint | Role |
|--------|----------|------|
| `GET` | `/v1/dashboard` | ADMIN |

### CMS API (`http://localhost:8081`)

#### Auth & Ads
| Method | Endpoint | Role |
|--------|----------|------|
| `POST` | `/v1/login` | All |
| `POST` | `/v1/ads/upload` | ADMIN |
| `GET` | `/v1/ads` | ADMIN |
| `GET/PUT/DELETE` | `/v1/ads/:id` | ADMIN |
| `POST` | `/v1/ads/:id/toggle` | ADMIN |
| `GET` | `/v1/ads/active` | All |
| `GET` | `/v1/ads/download/:id` | All |
| `POST` | `/v1/ads/:id/play` | All |

#### Users *(BARU)*
| Method | Endpoint | Role |
|--------|----------|------|
| `GET` | `/v1/users` | ADMIN |
| `POST` | `/v1/users` | ADMIN |
| `PUT` | `/v1/users/:id` | ADMIN |
| `DELETE` | `/v1/users/:id` | ADMIN |

#### Settings *(BARU)*
| Method | Endpoint | Role |
|--------|----------|------|
| `GET` | `/v1/settings` | ADMIN |
| `PUT` | `/v1/settings` | ADMIN |

> Dokumentasi endpoint lengkap: [`md/API_ENDPOINTS.md`](md/API_ENDPOINTS.md)

---

## Cara Menjalankan

### Prasyarat
- [Docker Desktop](https://www.docker.com/products/docker-desktop) v24+
- [Go](https://go.dev/dl/) 1.26+
- [Node.js](https://nodejs.org/) 18+ / [Bun](https://bun.sh/)

### 1. Clone & Checkout

```bash
git clone https://github.com/fariziadam11/gbs-pos-cms-api.git
cd gbs-pos-cms-api

# Gunakan branch main (update terbaru dari tim)
git checkout main
```

### 2. Jalankan dengan Docker Compose

```bash
docker-compose up -d

# Cek status
docker-compose ps

# Lihat logs
docker-compose logs -f
```

Jika error `Dirty database version`:

```bash
docker exec -it gbs-pos-cms-api-postgres-1 psql -U postgres -d gbs_pos \
  -c "UPDATE schema_migrations SET dirty = false WHERE dirty = true;"
docker restart gbs-pos-cms-api-pos-api-1
```

### 3. Verifikasi

```bash
# POS API
curl -X POST http://localhost:8080/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# CMS API
curl -X POST http://localhost:8081/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 4. CMS Web (Development)

```bash
cd cms-web
npm install   # atau: bun install
npm run dev   # atau: bun run dev
# Buka http://localhost:5173
```

### 5. Jalankan Lokal (tanpa Docker)

```bash
cp gbs-pos-api/.env.example gbs-pos-api/.env
cp gbs-cms-api/.env.example gbs-cms-api/.env
# Edit DATABASE_URL di masing-masing .env

make run-pos   # Terminal 1
make run-cms   # Terminal 2
```

### Perintah Make

```bash
make build        # Build kedua binary
make test         # Semua tests
make test-pos     # Tests POS API
make test-cms     # Tests CMS API
make run-pos      # Run POS API
make run-cms      # Run CMS API
```

---

## Environment Variables

### POS API

| Variable | Wajib | Keterangan |
|----------|-------|------------|
| `DATABASE_URL` | Ya | PostgreSQL connection string |
| `JWT_SECRET` | Jika demo auth | Min 32 karakter |
| `PORT` | Tidak | Default `8080` |
| `ENV` | Tidak | `development` / `production` |
| `LOG_LEVEL` | Tidak | `debug` / `info` / `warn` |
| `MIGRATIONS_PATH` | Tidak | Kosong = GORM AutoMigrate |
| `KEYCLOAK_BASE_URL` | Jika Keycloak | URL Keycloak |
| `KEYCLOAK_REALM` | Jika Keycloak | Nama realm |
| `ENABLE_DEMO_AUTH` | Tidak | `true` = aktifkan `/v1/login` lokal |

### CMS API

| Variable | Wajib | Keterangan |
|----------|-------|------------|
| `DATABASE_URL` | Ya | Sama dengan POS |
| `JWT_SECRET` | Jika demo auth | |
| `PORT` | Tidak | Default `8081` |
| `UPLOAD_DIR` | Tidak | Default `./uploads` |
| `KEYCLOAK_BASE_URL` | Jika Keycloak | |
| `KEYCLOAK_REALM` | Jika Keycloak | |
| `ENABLE_DEMO_AUTH` | Tidak | |

### CMS Web (`.env`)

```env
VITE_API_BASE_URL=https://api-cms.armmada.id
VITE_POS_API_BASE_URL=https://api-pos.armmada.id
VITE_KEYCLOAK_BASE_URL=https://auth.armmada.id
VITE_KEYCLOAK_REALM=gbs-pos
VITE_KEYCLOAK_CLIENT_ID=gbs-cms-web
```

CMS Web only enables Keycloak when all three `VITE_KEYCLOAK_*` values are non-empty.
Leave them empty to use the local username/password login against `VITE_API_BASE_URL/v1/login`.

---

## Akun Default

Tersedia setelah migration `008_seed_data`:

| Username | Password | Role |
|----------|----------|------|
| `admin` | `admin123` | ADMIN |
| `cashier` | `cashier123` | CASHIER |

> Akun ini bisa digunakan langsung via `/v1/login` ketika backend Keycloak env kosong, atau ketika Keycloak aktif dan `ENABLE_DEMO_AUTH=true`.
> Dalam production dengan Keycloak, buat user di realm Keycloak.

---

## Keycloak Authentication

Proyek ini mendukung **Keycloak** sebagai identity provider enterprise dengan fallback demo auth.

### Quick Setup

```bash
# Jalankan Keycloak (dev mode)
docker run -d --name keycloak -p 8082:8080 \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=admin \
  quay.io/keycloak/keycloak:26.1 start-dev
```

Kemudian:
1. Buka `http://localhost:8082/admin`
2. Buat realm `gbs-pos`
3. Buat roles: `ADMIN`, `CASHIER`
4. Buat client `gbs-pos-android` dan `gbs-cms-web`
5. Buat users dan assign roles

### Flow
```
Android / CMS Web → Keycloak (PKCE) → RS256 Token → Backend API
                                              ↓
                                    Validasi via JWKS
```

### Config Backend

```env
KEYCLOAK_BASE_URL=https://auth.armmada.id
KEYCLOAK_REALM=gbs-pos
ENABLE_DEMO_AUTH=false
```

> Panduan lengkap: [`md/KEYCLOAK_SETUP.md`](md/KEYCLOAK_SETUP.md)

---

## Testing

### Jalankan Tests

```bash
make test

# Dengan coverage
cd gbs-pos-api
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test coverage includes:
- Auth service (JWT issue & validation)
- Product & variant service
- Order service (idempotency, void logic)
- Settlement service (transactional batch close)
- **Discount service** *(BARU — termasuk pricing engine)*
- **Hold service** *(BARU)*
- **Dashboard repository** *(BARU)*
- Handler integration tests (SQLite in-memory)

---

## CI/CD Pipeline

```
Push ke develop → Lint → Test → Security Scan → Build Images → Deploy Staging
Push ke main    → Lint → Test → Security Scan → Build Images → Deploy Production (manual)
```

| Stage | Tool |
|-------|------|
| Lint | golangci-lint |
| Test | go test -race |
| Security | Trivy + gosec |
| Build | Docker Buildx (amd64, arm64) |
| Deploy | SSH + docker-compose |

Setup: [`md/CI_CD_DOCUMENTATION.md`](md/CI_CD_DOCUMENTATION.md)

---

## Branch & Kontributor

| Branch | Keterangan |
|--------|------------|
| `main` | Production-ready, update terbaru tim |
| `develop` | Staging, auto-deploy |
| `adam` | Branch pengembangan individual |

### Kontribusi

```bash
git checkout main
git pull origin main
git checkout -b feature/nama-fitur

# ... buat perubahan ...

git add .
git commit -m "feat: deskripsi fitur"
git push origin feature/nama-fitur
# Buat Pull Request ke main
```

---

## Dokumentasi

| File | Keterangan |
|------|------------|
| [`md/API_ENDPOINTS.md`](md/API_ENDPOINTS.md) | Referensi lengkap semua endpoint |
| [`md/BACKEND_API_GUIDE.md`](md/BACKEND_API_GUIDE.md) | Spesifikasi backend |
| [`md/KEYCLOAK_SETUP.md`](md/KEYCLOAK_SETUP.md) | Setup Keycloak lengkap |
| [`md/WEBSITE_DEPLOYMENT.md`](md/WEBSITE_DEPLOYMENT.md) | Deployment web |
| [`md/ARCHITECTURE_DIAGRAM.md`](md/ARCHITECTURE_DIAGRAM.md) | Diagram arsitektur |
| [`md/DEPLOYMENT.md`](md/DEPLOYMENT.md) | Panduan deployment |
| [`dbdiagram.dbml`](dbdiagram.dbml) | Schema database |
| [`GBS_POS_CMS_API.postman_collection.json`](GBS_POS_CMS_API.postman_collection.json) | Postman collection |

---

*Last updated: June 2026 — Branch: `main`*
