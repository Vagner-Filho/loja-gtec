# Admin Panel Setup Guide

This guide will help you set up and use the admin panel for Loja G-TEC.

## Overview

The admin panel allows you to:
- Authenticate as an administrator
- View all products in the database
- Add new products with image uploads
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
   - **Product Image**: Click to select an image file from your computer
2. You'll see a preview of the selected image
3. Click "Add Product"
4. The image will be uploaded and the product will be created

#### Editing a Product

1. Find the product in the list
2. Click the "Edit" button
3. The current product image is displayed
4. Modify any fields you want to change
5. Optionally upload a new image (leave empty to keep the current image)
6. Click "Update Product"
7. The changes will be saved to the database

#### Deleting a Product

1. Find the product in the list
2. Click the "Delete" button
3. Confirm the deletion
4. The product and its uploaded image will be removed

### Image Upload Details

- **Supported Formats**: All common image formats (JPEG, PNG, GIF, WebP, etc.)
- **Maximum File Size**: 5MB
- **Storage Location**: Uploaded images are stored in `/web/static/images/uploads/`
- **Filename Generation**: Random unique filenames are generated to avoid conflicts
- **Image Cleanup**: When a product is deleted or its image is replaced, the old image file is automatically removed

### Logging Out

Click the "Logout" button in the header to end your session.

## Security Features

- **Password Hashing**: All passwords are hashed using bcrypt
- **Session Management**: Sessions use secure, random tokens
- **HTTP-Only Cookies**: Session cookies cannot be accessed via JavaScript
- **Authentication Middleware**: Admin routes require authentication
- **Session Expiration**: Sessions expire after 24 hours
- **File Upload Validation**: 
  - File size limits (5MB max)
  - File type validation (images only)
  - Secure filename generation

## API Endpoints

The following API endpoints are available (all require authentication):

- `GET /api/admin/products` - Get all products (JSON response)
- `POST /api/admin/products` - Create a new product (multipart/form-data with file upload)
- `PUT /api/admin/products/{id}` - Update a product (multipart/form-data, image optional)
- `DELETE /api/admin/products/{id}` - Delete a product (also removes uploaded image)

## File Structure

```
lojagtec/
├── cmd/server/main.go                    # Main server with admin routes & file upload
├── internal/
│   ├── admin/auth.go                     # Authentication & session management
│   ├── products/products.go              # Product CRUD operations
│   └── database/database.go              # Database connection
├── web/
│   ├── templates/
│   │   ├── admin-login.html              # Admin login page
│   │   └── admin-dashboard.html          # Admin dashboard with file upload
│   └── static/
│       ├── js/
│       │   └── admin.js                  # Admin panel JavaScript with FormData
│       └── images/
│           └── uploads/                  # Directory for uploaded product images
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

### Image upload fails

Check that:
1. The `/web/static/images/uploads/` directory exists and is writable
2. The image file is under 5MB
3. The file is a valid image format
4. You have sufficient disk space

### Uploaded images not displaying

Verify that:
1. The `/static/` route is properly serving files
2. The uploads directory has correct permissions
3. The image path in the database starts with `/static/images/uploads/`

## Production Considerations

Before deploying to production:

1. **Change the default admin password**
2. **Enable HTTPS** and set `Secure: true` in cookie settings (internal/admin/auth.go:131)
3. **Use environment variables** for sensitive configuration
4. **Implement rate limiting** on login attempts
5. **Add CSRF protection** for forms
6. **Use a production-grade session store** (currently in-memory)
7. **Set up proper logging** for security events
8. **Configure proper file upload limits** based on your needs
9. **Implement image optimization** (resize, compress) for uploaded images
10. **Set up backup strategy** for uploaded images
11. **Consider using a CDN** for serving static assets and uploaded images

## Image Upload Best Practices

### For Development:
- The current implementation stores images in `/web/static/images/uploads/`
- Images are served directly by the Go server

### For Production:
Consider these improvements:
1. **Image Optimization**: Resize and compress images on upload
2. **Cloud Storage**: Use S3, Google Cloud Storage, or similar for uploaded images
3. **CDN**: Serve images through a CDN for better performance
4. **Virus Scanning**: Scan uploaded files for malware
5. **Backup**: Regular backups of uploaded images
6. **Cleanup Jobs**: Remove orphaned images that are no longer referenced

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
