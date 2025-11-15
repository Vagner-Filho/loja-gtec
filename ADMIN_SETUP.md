# Admin Panel Setup Guide

This guide will help you set up and use the admin panel for Loja G-TEC.

## Overview

The admin panel allows you to:
- Authenticate as an administrator
- View all products in the database
- Add new products
- Edit existing products
- Delete products

## Installation Steps

### 1. Install Dependencies

The bcrypt dependency has already been added. If you need to reinstall:

```bash
go get golang.org/x/crypto/bcrypt
```

### 2. Run Database Migration

You need to run the migration to create the necessary database tables:

```bash
PGPASSWORD=postgres psql -U lojagtec -d lojagtec -f scripts/migrations/001_create_admin_and_products.sql
```

Or if you have a different PostgreSQL setup:

```bash
psql -U your_username -d lojagtec -f scripts/migrations/001_create_admin_and_products.sql
```

This migration will:
- Create the `admin_users` table
- Create the `products` table with indexes
- Insert default products from your existing code
- Create a default admin user

### 3. Default Admin Credentials

After running the migration, you can log in with:

- **Username**: `admin`
- **Password**: `admin123`

**IMPORTANT**: Change this password immediately in a production environment!

### 4. Build and Run

Build the frontend:
```bash
tailwind -i ./web/static/css/style.css -o ./web/static/css/dist/style.css
```

Build the backend:
```bash
go build -o lojagtec cmd/server/main.go
```

Run the server:
```bash
./lojagtec
```

## Using the Admin Panel

### Accessing the Admin Panel

1. Navigate to `http://localhost:8080/admin/login`
2. Enter your credentials (default: admin/admin123)
3. You'll be redirected to the admin dashboard at `http://localhost:8080/admin`

### Managing Products

#### Adding a Product

1. Fill in the form at the top of the dashboard:
   - **Product Name**: e.g., "Purificador IBBL Mio Branco"
   - **Price**: e.g., 699.99
   - **Category**: Select from dropdown (bebedouros, purificadores, refis, pecas)
   - **Image URL**: e.g., "/static/images/purificador.jpg"
2. Click "Add Product"
3. The product will appear in the list below

#### Editing a Product

1. Find the product in the list
2. Click the "Edit" button
3. Modify the fields in the modal
4. Click "Update Product"
5. The changes will be saved to the database

#### Deleting a Product

1. Find the product in the list
2. Click the "Delete" button
3. Confirm the deletion
4. The product will be removed from the database

### Logging Out

Click the "Logout" button in the header to end your session.

## Security Features

- **Password Hashing**: All passwords are hashed using bcrypt
- **Session Management**: Sessions use secure, random tokens
- **HTTP-Only Cookies**: Session cookies cannot be accessed via JavaScript
- **Authentication Middleware**: Admin routes require authentication
- **Session Expiration**: Sessions expire after 24 hours

## API Endpoints

The following API endpoints are available (all require authentication):

- `GET /api/admin/products` - Get all products
- `POST /api/admin/products` - Create a new product
- `PUT /api/admin/products/{id}` - Update a product
- `DELETE /api/admin/products/{id}` - Delete a product

## File Structure

```
lojagtec/
├── cmd/server/main.go                    # Main server with admin routes
├── internal/
│   ├── admin/auth.go                     # Authentication & session management
│   ├── products/products.go              # Product CRUD operations
│   └── database/database.go              # Database connection
├── web/
│   ├── templates/
│   │   ├── admin-login.html              # Admin login page
│   │   └── admin-dashboard.html          # Admin dashboard
│   └── static/
│       └── js/
│           └── admin.js                  # Admin panel JavaScript
└── scripts/
    └── migrations/
        └── 001_create_admin_and_products.sql  # Database schema
```

## Troubleshooting

### Cannot connect to database

Make sure your PostgreSQL server is running and the credentials in `configs/config.toml` are correct.

### Products not loading

Check that you've run the database migration to create the products table.

### Login fails

Verify that:
1. The migration has been run
2. The admin user exists in the database
3. You're using the correct credentials

### Session expires immediately

Check that the system time is correct, as sessions use timestamps for expiration.

## Production Considerations

Before deploying to production:

1. **Change the default admin password**
2. **Enable HTTPS** and set `Secure: true` in cookie settings (internal/admin/auth.go:131)
3. **Use environment variables** for sensitive configuration
4. **Implement rate limiting** on login attempts
5. **Add CSRF protection** for forms
6. **Use a production-grade session store** (currently in-memory)
7. **Set up proper logging** for security events

## Adding More Admin Users

Currently, there's no UI to add more admin users. You can add them via SQL:

```sql
-- First, generate a bcrypt hash for the password
-- You can use an online tool or write a small Go program

INSERT INTO admin_users (username, password_hash) 
VALUES ('newadmin', '$2a$10$your_bcrypt_hash_here');
```

Or you can add a utility function in your Go code to create admin users programmatically.

## Support

For issues or questions, refer to the main README.md or check the AGENTS.md file for development guidelines.
