# Database Migrations

## Running Migrations

To apply the migrations to your database, run:

```bash
psql -U lojagtec -d lojagtec -f scripts/migrations/001_create_admin_and_products.sql
```

Or from the scripts directory:

```bash
cd scripts/migrations
psql -U lojagtec -d lojagtec -f 001_create_admin_and_products.sql
```

## Default Admin Credentials

After running the migration, you can log in with:
- **Username**: admin
- **Password**: admin123

**IMPORTANT**: Change this password immediately after first login in a production environment!
