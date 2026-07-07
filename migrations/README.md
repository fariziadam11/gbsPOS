# Shared database migrations (optional)

SQL migrations for the shared PostgreSQL database `gbs_pos` (POS + CMS tables).

By default both APIs use GORM `AutoMigrate`. Set `MIGRATIONS_PATH` to this directory
to use `golang-migrate` instead.

- **Runner:** only `gbs-pos-api` runs `golang-migrate` when `MIGRATIONS_PATH` is set.
- **CMS API** runs GORM `AutoMigrate` for its own models on startup.
- **Tests** use GORM `AutoMigrate` via `internal/database/test_helper.go` (SQLite in-memory).

Do not split into per-module folders or move to `gbs-common` — one database needs one ordered migration chain.

## Create a new migration

```bash
migrate create -ext sql -dir migrations -seq add_something
# edit migrations/0000XX_add_something.up.sql and .down.sql
```

## Apply migrations

**You do not need the `migrate` CLI to apply** — `gbs-pos-api` runs `migrate.Up()` on startup when `MIGRATIONS_PATH` is configured.

| Environment | How to apply |
|-------------|--------------|
| Docker Compose (dev) | Set `MIGRATIONS_PATH=/app/migrations` and bind-mount `./migrations`. After editing SQL: `docker compose up -d pos-api` (or restart pos-api). No rebuild required. |
| Docker (production image) | Set `MIGRATIONS_PATH`, copy/mount migrations, then rebuild and redeploy pos-api. |
| Local `go run` | Start pos-api once with `MIGRATIONS_PATH=../migrations` in `.env`. |

Only **pending** versions run (tracked in `schema_migrations`). Existing data is kept unless you use `docker compose down -v` (wipes Postgres volume).

Verify:

```bash
docker compose exec postgres psql -U postgres -d gbs_pos -c "SELECT * FROM schema_migrations;"
```

Manual CLI (optional, same result as startup):

```bash
migrate -path migrations -database "postgres://postgres:123@localhost:5433/gbs_pos?sslmode=disable" up
```
