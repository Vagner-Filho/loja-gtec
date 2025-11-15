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

## What's New (File Upload Update)

### Image Upload Feature
- Products now support **file upload** instead of manual file path entry
- Upload images directly from your computer (max 5MB)
- **Image preview** before submission
- Automatic **unique filename generation** to avoid conflicts
- Images stored in `/web/static/images/uploads/`
- **Automatic cleanup** when products are deleted or images replaced

### Routes Created
- `/admin/login` - Admin login page (GET/POST)
- `/admin` - Admin dashboard (requires auth)
- `/admin/logout` - Logout endpoint
- `/api/admin/products` - Product API (GET, POST with multipart/form-data)
- `/api/admin/products/{id}` - Product API (PUT with multipart/form-data, DELETE)

### Files Created
- `internal/admin/auth.go` - Authentication & session management
- `web/templates/admin-login.html` - Login page
- `web/templates/admin-dashboard.html` - Product management dashboard with file upload
- `web/static/js/admin.js` - Admin panel JavaScript with FormData handling
- `web/static/images/uploads/` - Directory for uploaded images
- `scripts/migrations/001_create_admin_and_products.sql` - Database schema

### Files Modified
- `cmd/server/main.go` - Added file upload handling, multipart/form-data parsing, image validation
- `internal/products/products.go` - Database CRUD operations
- `go.mod` - Added golang.org/x/crypto/bcrypt dependency

### Database Tables
- `admin_users` - Stores admin credentials
- `products` - Stores product information (with image paths)

## How to Use

### Adding a Product with Image Upload

1. Go to admin dashboard
2. Fill in product details (name, price, category)
3. Click "Product Image" and select an image file
4. See preview of selected image
5. Click "Add Product"
6. Image is uploaded and product is created in one request

### Editing a Product

1. Click "Edit" on any product
2. Current image is displayed
3. Modify any fields
4. **Optional**: Upload a new image (or leave empty to keep current)
5. Click "Update Product"
6. Old image is automatically deleted if replaced

### Technical Details

**Single Request Approach:**
- One multipart/form-data request contains all product data + image file
- Atomic operation: either everything succeeds or nothing does
- No orphaned files from failed product creation
- Simpler error handling

**File Validation:**
- Max size: 5MB
- Type: Images only (checked via Content-Type header)
- Unique filenames using crypto/rand

**Image Cleanup:**
- Deleting a product removes its uploaded image
- Replacing an image removes the old one
- Only images in `/uploads/` are automatically cleaned

## Features

- Secure authentication with bcrypt password hashing
- Session-based authentication with HTTP-only cookies
- Full CRUD operations for products
- **File upload with validation and preview**
- Real-time product management interface
- Responsive design matching your site's style
- Automatic image cleanup

For detailed documentation, see ADMIN_SETUP.md
