# Shared database migrations

SQL migrations for the shared PostgreSQL database `gbs_pos` (POS + CMS tables).

- **Runner:** only `gbs-pos-api` runs `golang-migrate` on startup.
- **CMS API** connects after migrations; it does not run schema changes.
- **Tests** use GORM `AutoMigrate` via `internal/database/test_helper.go` (SQLite in-memory).

Do not split into per-module folders or move to `gbs-common` — one database needs one ordered migration chain.
