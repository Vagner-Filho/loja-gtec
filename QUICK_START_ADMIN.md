# Quick Start - Admin Panel

## TL;DR

1. Run the database migration:
   ```bash
   PGPASSWORD=postgres psql -U lojagtec -d lojagtec -f scripts/migrations/001_create_admin_and_products.sql
   ```

2. Build and run:
   ```bash
   go build -o lojagtec cmd/server/main.go
   ./lojagtec
   ```

3. Access admin panel:
   - URL: http://localhost:8080/admin/login
   - Username: `admin`
   - Password: `admin123`

## What's New

### Routes Created
- `/admin/login` - Admin login page (GET/POST)
- `/admin` - Admin dashboard (requires auth)
- `/admin/logout` - Logout endpoint
- `/api/admin/products` - Product API (GET, POST)
- `/api/admin/products/{id}` - Product API (PUT, DELETE)

### Files Created
- `internal/admin/auth.go` - Authentication & session management
- `web/templates/admin-login.html` - Login page
- `web/templates/admin-dashboard.html` - Product management dashboard
- `web/static/js/admin.js` - Admin panel JavaScript
- `scripts/migrations/001_create_admin_and_products.sql` - Database schema

### Files Modified
- `cmd/server/main.go` - Added admin routes and API endpoints
- `internal/products/products.go` - Added database CRUD operations
- `go.mod` - Added golang.org/x/crypto/bcrypt dependency

### Database Tables
- `admin_users` - Stores admin credentials
- `products` - Stores product information (migrated from in-memory)

## Features

- Secure authentication with bcrypt password hashing
- Session-based authentication with HTTP-only cookies
- Full CRUD operations for products
- Real-time product management interface
- Responsive design matching your site's style

For detailed documentation, see ADMIN_SETUP.md
