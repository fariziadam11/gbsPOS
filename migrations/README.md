# Shared database migrations

SQL migrations for the shared PostgreSQL database `gbs_pos` (POS + CMS tables).

- **Runner:** only `gbs-pos-api` runs `golang-migrate` on startup.
- **CMS API** connects after migrations; it does not run schema changes.
- **Tests** use GORM `AutoMigrate` via `internal/database/test_helper.go` (SQLite in-memory).

Do not split into per-module folders or move to `gbs-common` — one database needs one ordered migration chain.

## Create a new migration

```bash
migrate create -ext sql -dir migrations -seq add_something
# edit migrations/0000XX_add_something.up.sql and .down.sql
```

## Apply migrations

**You do not need the `migrate` CLI to apply** — `gbs-pos-api` runs `migrate.Up()` on every startup.

| Environment | How to apply |
|-------------|--------------|
| Docker Compose (dev) | `./migrations` is bind-mounted into the container. After editing SQL: `docker compose up -d pos-api` (or restart pos-api). No rebuild required. |
| Docker (production image) | Rebuild and redeploy pos-api so new SQL files are copied into the image. |
| Local `go run` | Start pos-api once; `MIGRATIONS_PATH=../migrations` in `.env`. |

Only **pending** versions run (tracked in `schema_migrations`). Existing data is kept unless you use `docker compose down -v` (wipes Postgres volume).

Verify:

```bash
docker compose exec postgres psql -U postgres -d gbs_pos -c "SELECT * FROM schema_migrations;"
```

Manual CLI (optional, same result as startup):

```bash
migrate -path migrations -database "postgres://postgres:123@localhost:5433/gbs_pos?sslmode=disable" up
```
