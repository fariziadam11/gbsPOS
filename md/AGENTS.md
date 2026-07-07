# Agent Instructions for pos-cms

## Architecture

Three Go modules in a monorepo:

- `gbs-pos-api/` — POS REST API on :8080 (products, orders, settlements)
- `gbs-cms-api/` — CMS REST API on :8081 (ad upload, playlist, download)
- `gbs-common/` — Shared middleware/auth/response imported by both via `replace` directive

Both APIs share the same PostgreSQL database (`gbs_pos`) but keep separate table namespaces. Do not merge them into a single module; the spec requires independent deployment.

## Build & Test

From repo root:

```bash
make build        # builds bin/gbs-pos-api and bin/gbs-cms-api
make test         # runs tests in both modules
make run-pos      # go run ./cmd/server from gbs-pos-api
make run-cms      # go run ./cmd/server from gbs-cms-api
```

From inside a module:

```bash
cd gbs-pos-api && go test ./... -count=1
cd gbs-pos-api && go build ./cmd/server
```

## Environment & Config

Copy `.env.example` to `.env` in each module. Config is loaded with `github.com/caarlos0/env/v10` (`internal/config/config.go`).
The `.env` file is automatically loaded at startup via `github.com/joho/godotenv` in `cmd/server/main.go`.

**Critical env vars:**

- `JWT_SECRET` — mandatory, must be ≥32 characters (validated at startup)
- `DATABASE_URL` — PostgreSQL connection string
- `MIGRATIONS_PATH` — (POS API only) path to SQL migrations. Leave empty to use GORM `AutoMigrate` (default). Set to a migrations directory to use `golang-migrate/v4` instead.
- `UPLOAD_DIR` — CMS only, local filesystem path for uploaded videos

**Docker:**

```bash
docker-compose up -d   # starts postgres, pos-api (:8080), cms-api (:8081)
```

Docker Compose leaves `MIGRATIONS_PATH` empty so both APIs use GORM `AutoMigrate`; CMS still starts after POS.

## Database

- **Production:** PostgreSQL 15+ via `gorm.io/driver/postgres`
- **Tests:** SQLite in-memory via `github.com/glebarez/sqlite` (`internal/database/test_helper.go`)
- `database.NewTestDB()` creates the in-memory DB and auto-migrates all models — use it in every test package that touches the DB
- Connection pool tuned in `database.Connect`: MaxOpen=25, MaxIdle=10, ConnMaxLifetime=1h

## Auth Middleware

Do **not** call `os.Getenv("JWT_SECRET")` inside request handlers or middleware. The secret is injected via constructor:

```go
middleware.NewAuthMiddleware(cfg.JWTSecret)
```

This returns a `gin.HandlerFunc` that enforces:

- `WithValidMethods(["HS256"])`
- `WithExpirationRequired()`
- `WithLeeway(5 * time.Second)`

Use `middleware.RequireRole("ADMIN")` for admin-only routes. Claims stored in Gin context: `userID`, `username`, `role`.

## Shared Module (`gbs-common`)

Both API modules import `gbs-common` via `replace gbs-common => ../gbs-common` in their `go.mod`. If you add new exported code to `gbs-common`, run `go mod tidy` in **all three modules** before building.

Contents:

- `middleware/auth.go` — `NewAuthMiddleware`, `RequireRole`
- `middleware/cors.go` — `CORSMiddleware`
- `middleware/logger.go` — `LoggerMiddleware` (zerolog request logging)
- `pkg/response/response.go` — Standard JSON envelope: `{success, data, error, idempotent}`

## Testing Patterns

- Handler tests wire a real Gin router with real services against SQLite (`internal/handler/handler_test.go`)
- Service tests use `database.NewTestDB()` + repositories directly
- Tests set `jwtSecret := "test-secret-key-minimum-32-characters"` and pass it to `NewAuthService(userRepo, jwtSecret, 24)`
- `gin.SetMode(gin.TestMode)` in test setup functions

## Key Conventions

- **Idempotency:** Order create returns `201` on first create, `200` with `idempotent: true` on duplicate. Always check the boolean.
- **Product Update:** Partial updates via field guards (`if updates.Name != ""`); do not blindly assign all fields.
- **Settlement:** Runs inside `db.Transaction()` with `SELECT ... FOR UPDATE` on unsettled orders (`clause.Locking{Strength: "UPDATE"}`).
- **CORS:** Explicit origins (`https://cms.gbs.com`, `localhost:5173`), never `"*"` with `AllowCredentials: true`.
- **Video Serving:** `Content-Type`, `Content-Length`, `Accept-Ranges: bytes`, `Cache-Control`, `Content-Disposition: inline` — required for ExoPlayer/browser scrubbing.
- **Response Format:** Always use `response.Success()`, `response.Error()`, `response.ValidationError()`. Do not return raw `gin.H` to authenticated endpoints.

## Important Gotchas

- Schema changes are applied via GORM `AutoMigrate` by default. Set `MIGRATIONS_PATH` to a `golang-migrate` directory to use SQL migrations instead.
- `gbs-common` has its own `go.mod`. Its dependency versions may differ from the API modules. Tidy all three after changes.
- `gbs-pos-api/Dockerfile` does not copy SQL migrations by default; GORM `AutoMigrate` is used. Set `MIGRATIONS_PATH` and mount/copy migrations if you prefer `golang-migrate`.
- The `Ad.StoreTypes` field uses `gorm:"serializer:json"` (stored as JSON text), not PostgreSQL arrays. Query with `store_types LIKE '%RETAIL%'`.
- `MaxMultipartMemory = 32 << 20` is set on the Gin router. The CMS upload handler separately enforces a 50MB file limit.

## Deployment

- See `DEPLOYMENT.md` for the full production deployment guide (VPS + Docker Compose + Cloudflare Tunnel + GitHub Actions).
- Both APIs expose `/health` for Docker health checks (no auth required).
- Production uses `docker-compose.prod.yml` with secrets via `.env`, restart policies, and health checks.
- Images are built and pushed to GitHub Container Registry (`ghcr.io`) via GitHub Actions on every push to `main`.
- The `.env` file is auto-loaded by `godotenv` at startup; never commit `.env` to Git.
- Both APIs run GORM `AutoMigrate` for their own models against the shared DB. `migrations/` is optional and only used when `MIGRATIONS_PATH` is set.

## Project Memory (AI Dynamic Context)

This section is maintained by AI agents. It must be updated after each completed task.

### Last Updated

- 2026-06-12
- 2026-06-12 (context update): added Stock Management + POS Hold/Resume planning phase
- 2026-06-15
- 2026-06-17
- 2026-07-03

### Active Development Focus

- POS system enhancements
- CMS dashboard integration

### Recently Implemented

- Product module with storeType filtering (RETAIL / FNB / OUTFIT)
- Order repository with settlement summary (CASH / CARD / QRIS)
- Settlement locking using SELECT FOR UPDATE
- CMS Web unified dashboard (POS + CMS views)
- Discount engine expanded to support PRODUCT, TRANSACTION, and VOUCHER scopes with combined discount flow.
- Pricing calculation service applies PRODUCT -> TRANSACTION -> VOUCHER discounts and selects highest monetary TRANSACTION/VOUCHER discount per scope/code.
- CMS Web Discount management page centralizes PRODUCT / TRANSACTION / VOUCHER CRUD and product-page discount icons now deep-link to filtered discount management.
- Hold module reworked with lifecycle ACTIVE -> RESUMED.
- CMS Web auth now falls back to local username/password JWT login when frontend Keycloak env is empty.

### Known System Rules (Runtime)

- storeType is GLOBAL filter across POS system
- Orders are immutable after settlement
- DB changes are applied via GORM AutoMigrate by default; golang-migrate is optional
- Handler must not contain business logic
- Stock movement must be tracked as immutable ledger (IN / OUT / OPNAME)
- POS Hold must persist cart snapshot before checkout or discard

## Memory Update Rule

After completing ANY task:

- Update section "Project Memory"
- Only record meaningful changes:
    - new module
    - schema change
    - business rule change
- Do NOT log minor refactors
- Keep max 10 bullet points per section
