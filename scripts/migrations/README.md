# Database Migrations

Migrations are SQL files in this directory named with an ascending integer prefix:

```
1_baseline.sql
2_add_feature.sql
3_fix_index.sql
```

The application automatically applies missing migrations on startup via `internal/database/migrations.go`.

## Manual Execution (if needed)

You can still run a migration manually with `psql`:

```bash
psql -U lojagtec -d lojagtec -f scripts/migrations/1_baseline.sql
```

## Creating a New Migration

1. Pick the next integer (e.g., `4`).
2. Create `4_description.sql`.
3. Write idempotent SQL where possible (`CREATE TABLE IF NOT EXISTS`, `ADD COLUMN IF NOT EXISTS`, etc.).
4. Restart the server; the migration will be applied automatically.

## Development Setup

In `development` mode, the server will also run `scripts/schema/schema.sql` and `scripts/schema/seed.sql` if the database is completely empty. In production, only migrations run — `schema.sql` and `seed.sql` are never applied automatically.
