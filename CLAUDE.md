# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GBS POS-CMS is a Go monorepo for a Point of Sale and Content Management System. It consists of:
- **gbs-pos-api** (port 8080) - POS REST API for retail/F&B/fuel operations
- **gbs-cms-api** (port 8081) - CMS REST API for ads, users, settings
- **gbs-common** - Shared middleware (JWT, Keycloak, CORS, logging)
- **cms-web** - Vue 3 Admin Panel
- **migrations** - SQL migrations shared between APIs

## Development Commands

### Build & Run
```bash
make build        # Build both binaries
make build-pos    # Build POS API only
make build-cms    # Build CMS API only

make run-pos      # Run POS API on port 8080
make run-cms      # Run CMS API on port 8081
```

### Testing
```bash
make test         # Run all tests
make test-pos     # POS API tests only
make test-cms     # CMS API tests only

# With coverage
cd gbs-pos-api && go test -race -coverprofile=coverage.out ./...
cd gbs-cms-api && go test -race -coverprofile=coverage.out ./...
```

### Linting
```bash
golangci-lint run ./...   # From repo root, lints all packages
```

### Docker Development
```bash
docker-compose up -d      # Start all services (POS, CMS, Postgres, KrakenD)
docker-compose logs -f     # Follow logs
docker-compose down        # Stop all services

# Fix dirty migration state
docker exec -it gbs-pos-cms-api-postgres-1 psql -U postgres -d gbs_pos \
  -c "UPDATE schema_migrations SET dirty = false WHERE dirty = true;"
docker restart gbs-pos-cms-api-pos-api-1
```

## Architecture

### Layered Structure (per API)
Each API follows the pattern: `handler → service → repository`

```
cmd/server/main.go
└── internal/
    ├── config/      # Environment variables → Config struct
    ├── database/    # DB connection, migrations, seed data
    ├── dto/         # Request/Response structs
    ├── handler/     # HTTP handlers (Parse, Validate, Call Service, Respond)
    ├── model/       # GORM database models
    ├── repository/  # Database queries (Create, Find, Update, Delete)
    ├── router/      # Route definitions with middleware
    └── service/     # Business logic, orchestrates repositories
```

### Key Design Patterns

1. **Dependency Injection**: All repositories/services are instantiated in `main.go` and injected
2. **Shared Models**: GORM models in each API's `model/` package; some are shared
3. **Dual Auth**: Middleware auto-detects HS256 (local JWT) vs RS256 (Keycloak) tokens
4. **Pricing Engine**: Discount service calculates final prices, injected into product service

### Authentication Flow
```
Request → CORSMiddleware → AuthMiddleware (HS256 or RS256) → RequireRole("ADMIN") → Handler
```
Token algorithm detection: RS256 → Keycloak JWKS validation, HS256 → JWT secret validation

### Database
- PostgreSQL 15, database `gbs_pos`
- GORM AutoMigrate by default (unless `MIGRATIONS_PATH` is set)
- SQL migrations in `/migrations/` for schema changes
- SQLite in-memory for tests

### Default Credentials (after seed)
- Username: `admin`, Password: `admin123` (ADMIN role)
- Username: `cashier`, Password: `cashier123` (CASHIER role)

## Key Configuration

### POS API (.env)
```env
DATABASE_URL=postgres://user:pass@host:5432/gbs_pos?sslmode=disable
JWT_SECRET=<min-32-chars>
PORT=8080
ENABLE_DEMO_AUTH=true        # Enable local /v1/login endpoint
KEYCLOAK_BASE_URL=https://auth.armmada.id
KEYCLOAK_REALM=gbs-pos
```

### CMS API (.env)
```env
DATABASE_URL=<same as POS>
JWT_SECRET=<same as POS>
PORT=8081
UPLOAD_DIR=./uploads
ENABLE_DEMO_AUTH=true
```

### CMS Web (.env)
```env
VITE_API_BASE_URL=http://localhost:8081
VITE_POS_API_BASE_URL=http://localhost:8080
VITE_KEYCLOAK_BASE_URL=https://auth.armmada.id
VITE_KEYCLOAK_REALM=gbs-pos
VITE_KEYCLOAK_CLIENT_ID=gbs-cms-web
```

## Important Notes

1. **Shared Database**: Both APIs share the same PostgreSQL database (`gbs_pos`)
2. **Keycloak Optional**: Set `KEYCLOAK_BASE_URL` empty to use local JWT auth only
3. **GORM AutoMigrate**: Models are auto-migrated on startup; migrations folder is for schema snapshots
4. **CMS Web Auth**: Uses Keycloak OIDC PKCE flow when `VITE_KEYCLOAK_*` vars are set
5. **Multi-store Support**: Products and ads have `store_type` field (RETAIL, FB, FUEL, ALL)

## CI/CD

- **develop branch**: Auto-deploys to staging on push
- **main branch**: Manual deploy to production via GitHub Actions
- Pipeline: lint → test → security scan (Trivy, gosec) → Docker build → deploy
