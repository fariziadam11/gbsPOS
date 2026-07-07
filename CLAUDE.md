# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GBS POS & CMS is a multi-service system for managing a fuel/gas station retail business. It consists of:
- **gbs-pos-api**: Point of Sale backend (orders, products, settlements, fuel pumps, customers, dashboard)
- **gbs-cms-api**: Content Management backend (ads, settings, user management)
- **gbs-common**: Shared Go middleware (auth, CORS, logging)
- **cms-web**: Vue 3 admin dashboard for managing both APIs

## Tech Stack

| Component | Language/Framework |
|-----------|-------------------|
| APIs | Go 1.26.1 with Gin |
| Database | PostgreSQL with GORM |
| Authentication | JWT (HS256) with optional Keycloak (RS256) |
| Frontend | Vue 3, PrimeVue 4, Pinia, Vue Router |
| Build Tool | Vite |
| Containerization | Docker & Docker Compose |

## Common Commands

### Go APIs

```bash
# Run tests
make test              # Run all tests (POS + CMS)
make test-pos          # Run POS API tests only
make test-cms          # Run CMS API tests only
go test ./... -count=1 # Run with no caching

# Build binaries
make build             # Build both APIs
make build-pos         # Build POS API to ./bin/gbs-pos-api
make build-cms         # Build CMS API to ./bin/gbs-cms-api

# Run locally (from project root, loads .env files)
make run-pos           # Run POS API on port 8080
make run-cms           # Run CMS API on port 8081

# Direct go commands
cd gbs-pos-api && go run ./cmd/server
cd gbs-cms-api && go run ./cmd/server
```

### Docker

```bash
# Development environment (from project root)
docker compose up          # Start all services
docker compose up -d       # Run in background
docker compose down        # Stop all services

# Production (uses docker-compose.prod.yml with env vars from /opt/gbs/.env)
```

### Frontend (cms-web)

```bash
cd cms-web
npm install
npm run dev          # Development server on port 5173
npm run build        # Production build
npm run preview       # Preview production build
```

## Architecture

### API Structure (Go)

Each API follows layered architecture:
```
cmd/server/main.go      # Entry point, dependency injection
internal/
  config/               # Environment variable loading
  database/             # DB connection, migrations, seeding
  model/                # GORM models
  repository/            # Data access layer
  service/              # Business logic
  handler/              # HTTP handlers
  router/               # Route definitions
  dto/                  # Request/response structures
```

### Database

Both APIs share the same PostgreSQL database (`gbs_pos`). Migrations can be:
- **AutoMigrate**: Default (leave `MIGRATIONS_PATH` empty)
- **golang-migrate**: Optional (set `MIGRATIONS_PATH=/path/to/migrations`)

### Authentication

The system supports dual auth modes (determined by environment):
1. **JWT Legacy** (`ENABLE_DEMO_AUTH=true`): Local JWT with HS256
2. **Keycloak** (`KEYCLOAK_BASE_URL` + `KEYCLOAK_REALM` set): OIDC with RS256

The `gbs-common/middleware/auth.go` `NewCompositeAuthMiddleware` detects token type automatically.

### CMS Web

The Vue frontend connects to both backend APIs:
- `VITE_API_BASE_URL`: Points to gbs-cms-api
- `VITE_POS_API_BASE_URL`: Points to gbs-pos-api

It uses Pinia for state management and TanStack Vue Query for API calls.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | 8080/8081 |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://gbspos:gbspos_secret@localhost:5432/gbs_pos?sslmode=disable` |
| `JWT_SECRET` | JWT signing key (min 32 chars) | - |
| `KEYCLOAK_BASE_URL` | Keycloak server URL | - |
| `KEYCLOAK_REALM` | Keycloak realm name | - |
| `ENABLE_DEMO_AUTH` | Enable local JWT auth | false |
| `LOG_LEVEL` | Logging level | debug |
| `UPLOAD_DIR` | CMS file upload directory | `./uploads/ads` |

## Testing

Tests use `github.com/stretchr/testify` and SQLite in-memory for isolation:
- Repository tests in `internal/repository/*_test.go`
- Service tests in `internal/service/*_test.go`
- Handler tests in `internal/handler/*_test.go`

Run specific test files:
```bash
go test ./internal/repository/... -v
go test ./internal/service/order_service_test.go -v
```

## Key Files

- `gbs-pos-api/cmd/server/main.go`: POS API entry point
- `gbs-cms-api/cmd/server/main.go`: CMS API entry point (handles routing inline)
- `gbs-common/middleware/auth.go`: JWT/Keycloak authentication
- `docker-compose.yml`: Development environment
- `docker-compose.prod.yml`: Production deployment
- `Makefile`: Common development commands
