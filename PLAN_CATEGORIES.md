# Dynamic Product Categories - Implementation Plan

## Overview

Replace the hardcoded four product categories (`bebedouros`, `purificadores`, `refis`, `pecas`) with a normalized `categories` table. Categories can be created, edited, activated/deactivated by admin users. The public store uses a dynamic `<select>` instead of hardcoded category buttons.

---

## 1. Database Schema Changes

### New Table: `categories`

```sql
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,          -- Display name (e.g., "Bebedouros")
    slug TEXT UNIQUE NOT NULL,   -- URL-safe identifier (e.g., "bebedouros")
    allows_compatibility BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Migration Steps

1. Create `categories` table.
2. Insert the 4 existing categories with correct flags:
   - `Bebedouros` / `bebedouros` / `allows_compatibility = FALSE` / `is_active = TRUE`
   - `Purificadores` / `purificadores` / `allows_compatibility = FALSE` / `is_active = TRUE`
   - `Refis` / `refis` / `allows_compatibility = TRUE` / `is_active = TRUE`
   - `Peças` / `pecas` / `allows_compatibility = TRUE` / `is_active = TRUE`
3. Add `category_id INTEGER REFERENCES categories(id)` to `products`.
4. Populate `products.category_id` from the old `products.category` text values via JOIN on `slug`.
5. Drop old `products.category` column.
6. Drop old `idx_products_category` index; create new one on `products(category_id)`.

---

## 2. Backend Changes

### 2.1 `internal/products/products.go`

#### New Structs

```go
type Category struct {
    ID                   int    `json:"id"`
    Name                 string `json:"name"`
    Slug                 string `json:"slug"`
    AllowsCompatibility  bool   `json:"allowsCompatibility"`
    IsActive             bool   `json:"isActive"`
}
```

#### Updated `Product` Struct

```go
type Product struct {
    ID                    int        `json:"id"`
    ProductID             int        `json:"productId"`
    Name                  string     `json:"name"`
    Price                 float64    `json:"price"`
    Image                 string     `json:"image"`
    CategoryID            int        `json:"categoryId"`
    Category              string     `json:"category"`          // slug (for backward compat)
    CategoryName          string     `json:"categoryName"`      // display name
    AllowsCompatibility   bool       `json:"allowsCompatibility"`
    Description           string     `json:"description,omitempty"`
    SKU                   string     `json:"sku,omitempty"`
    IsAvailable           bool       `json:"isAvailable"`
    IsOnOffer             bool       `json:"isOnOffer"`
    OfferPrice            float64    `json:"offerPrice,omitempty"`
    OfferStartDate        *time.Time `json:"offerStartDate,omitempty"`
    OfferEndDate          *time.Time `json:"offerEndDate,omitempty"`
    BrandIDs              []int      `json:"brandIds,omitempty"`
    FitsProductIDs        []int      `json:"fitsProductIds,omitempty"`
}
```

#### New Functions

- `GetAllCategories() ([]Category, error)` — all categories for admin management.
- `GetActiveCategories() ([]Category, error)` — only active categories for public store.
- `GetCategoryBySlug(slug string) (*Category, error)` — resolve slug to category (for store routes).
- `GetCategoryByID(id int) (*Category, error)` — for admin editing.
- `CreateCategory(name string, allowsCompatibility bool) (*Category, error)` — generates slug from name.
- `UpdateCategory(id int, name string, allowsCompatibility bool, isActive bool) error` — edit existing category.
- `ToggleCategoryActive(id int) error` — activate/deactivate.

#### Updated Functions

- `GetAllProducts()` — add `JOIN categories` to populate `Category`, `CategoryName`, `AllowsCompatibility`.
- `GetProductsByCategoryAndBrands(categorySlug string, brandIDs []int)` — filter via `categories.slug`.
- `GetProductByID(id int)` — include category data.
- `CreateProduct(name string, price float64, categoryID int, description, sku string, isAvailable bool, brandIDs, fitsProductIDs []int)` — receives `categoryID` instead of slug string.
- `UpdateProduct(id int, name string, price float64, categoryID int, description, sku string, isAvailable bool, brandIDs, fitsProductIDs []int)` — receives `categoryID`.
- `insertProductCompatibility(tx, productID, categoryID, fitsProductIDs)` — resolve `allows_compatibility` from `categories` table inside the transaction instead of hardcoded check.
- `GetRelatedProducts(excludeProductID int, categorySlug string, limit int)` — filter via `categories.slug`.

#### Removed Functions

- `isPartsCategory()` — no longer needed; logic moves to DB flag.

### 2.2 `internal/offers/offers.go`

#### Updated `Offer` Struct

```go
type Offer struct {
    ID            int        `json:"id"`
    ProductID     int        `json:"productId"`
    Name          string     `json:"name"`
    Price         float64    `json:"price"`
    OfferPrice    float64    `json:"offerPrice"`
    Category      string     `json:"category"`      // slug
    CategoryName  string     `json:"categoryName"`  // display name
    StartDate     *time.Time `json:"startDate,omitempty"`
    EndDate       *time.Time `json:"endDate,omitempty"`
    IsActive      bool       `json:"isActive"`
}
```

#### Updated Queries

All offer queries that `JOIN products` must also `JOIN categories` to populate `Category` (slug) and `CategoryName`.

### 2.3 `cmd/server/main.go`

#### Removed

- `isPartsCategory()` function.
- Hardcoded `formatCategory` template function maps (3 occurrences in admin routes).
- Individual category routes:
  - `/products/bebedouros`
  - `/products/purificadores`
  - `/products/refis`
  - `/products/pecas`

#### New / Updated Routes

**Public Store — Consolidated Category Route:**

```go
http.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
    slug := strings.TrimPrefix(r.URL.Path, "/products/")
    // slug can be "all" or "ofertas" or any category slug
    brandIDs := parseBrandIDs(r.URL.Query().Get("brands"))

    var prods []products.Product
    var err error

    switch slug {
    case "all":
        prods, err = products.GetProductsByCategoryAndBrands("", brandIDs)
    case "ofertas":
        // handle offers (existing logic)
    default:
        prods, err = products.GetProductsByCategoryAndBrands(slug, brandIDs)
    }
    // ... render product-cards.html
})
```

**Public API:**

- `GET /api/categories/options` — returns `<option>` fragment of active categories for the store filter `<select>`.

**Admin Dashboard (`/admin`):**

- Include `Categories` in `adminDashboardData` struct.
- Pass to `admin-dashboard.html` template.

**Admin Product Edit (`/admin/products/{id}/edit`):**

- Include `Categories` in `adminEditData` struct.
- Pass to `admin-edit-form.html` template.

**Admin API — Categories:**

- `GET /api/admin/categories` — returns HTML list fragment for the management page.
- `POST /api/admin/categories` — creates a new category.
- `PUT /api/admin/categories/{id}` — updates category name/compatibility.
- `PUT /api/admin/categories/{id}/toggle` — toggles `is_active`.
- `GET /api/admin/categories/options` — returns `<option>` fragment for product forms.

**Admin API — Products:**

- `POST /api/admin/products` — reads `category_id` (int) from form instead of `category` (text). Validates compatibility via DB.
- `PUT /api/admin/products/{id}` — reads `category_id` (int) from form.

**Product Detail Page (`/produto/{id}`):**

- Use `product.AllowsCompatibility` instead of `product.Category == "refis" || product.Category == "pecas"`.

#### Updated Data Structs

```go
type adminDashboardData struct {
    CanViewOrders bool
    Brands        []products.Brand
    Products      []products.ProductOption
    Categories    []products.Category
}

type adminEditData struct {
    Product         *products.Product
    Brands          []products.Brand
    Products        []products.ProductOption
    Categories      []products.Category
    BrandSelections map[int]bool
    FitSelections   map[int]bool
    ProductImages   []products.ProductImage
}
```

---

## 3. Frontend Changes

### 3.1 Admin Pages

#### `admin-dashboard.html`

- Replace hardcoded `<select>` for category with dynamic `<select>` populated from `{{range .Categories}}`.
- Each `<option value="{{.ID}}">{{.Name}}</option>`.
- Add **"Gerenciar Categorias"** link in the nav or below the product form.
- Add a new section/card for category management, or link to a dedicated page.

#### `admin-edit-form.html`

- Replace hardcoded `<select>` with dynamic one using `{{range .Categories}}` and `selected` based on `Product.CategoryID`.

#### New: `admin-categories.html` (Dedicated Management Page)

Full page listing all categories with:
- Name
- Slug
- "Permite compatibilidade" badge
- Status badge (Ativo / Inativo)
- **Editar** button — opens a modal or inline edit form.
- **Ativar/Desativar** button — toggles `is_active` via HTMX.

#### New: `admin-category-modal.html`

Modal for creating/editing a category:
- Name input
- "Permite compatibilidade" checkbox
- Save / Cancel buttons
- Used both in the categories page and potentially inline.

#### New: `admin-category-options.html`

HTMX fragment returning `<option>` elements for product forms.

#### `admin-product-list.html` & `admin-product-card.html`

- Replace `{{formatCategory .Category}}` with `{{.CategoryName}}`.

#### `admin-offers-list.html`

- Replace `{{.Category}}` with `{{.CategoryName}}`.

### 3.2 Store Pages

#### `index.html`

**Replace category buttons with a `<select>`:**

```html
<div class="flex flex-wrap justify-center gap-2 items-center">
  <label for="category-select" class="text-sm font-medium text-gray-700">Categoria:</label>
  <select id="category-select" onchange="loadProducts()" class="...">
    <option value="all" selected>Todos</option>
    <option value="ofertas">Ofertas</option>
    <!-- dynamic options loaded here -->
  </select>
  <div hx-get="/api/categories/options" hx-trigger="load" hx-target="#category-select" hx-swap="beforeend"></div>
</div>
```

**Refactor JavaScript:**

- Remove `updateActiveFilter()` (was button-based).
- Remove `updateProductFilterButtons()` button loop.
- Add `loadProducts()`:
  ```js
  function loadProducts() {
    const category = document.getElementById('category-select').value;
    const brandParam = getBrandQueryString();
    const url = category === 'ofertas'
      ? `/products/ofertas${brandParam}`
      : `/products/${category}${brandParam}`;
    htmx.ajax('GET', url, { target: '#products-container', swap: 'innerHTML' });
  }
  ```
- Update `toggleBrandFilter()` and `clearBrandFilters()` to call `loadProducts()` at the end instead of `activeBtn.click()`.
- Remove `let currentCategory` tracking via buttons; read from select instead.

#### `product.html`

- Meta tags, breadcrumb, badge: replace `{{.Product.Category}}` with `{{.Product.CategoryName}}`.
- Compatibility section: replace `or (eq .Product.Category "refis") (eq .Product.Category "pecas")` with `{{.Product.AllowsCompatibility}}`.

---

## 4. Migration Script

This change was incorporated into the baseline migration (`scripts/migrations/1_baseline.sql`). For future schema changes, add a new numbered migration file (e.g., `2_add_feature.sql`) and restart the server — the migration runner will apply it automatically.

If you ever need to recreate the categories migration from scratch on a legacy database, the steps are:

```sql
BEGIN;

CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    allows_compatibility BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO categories (name, slug, allows_compatibility, is_active) VALUES
  ('Bebedouros', 'bebedouros', FALSE, TRUE),
  ('Purificadores', 'purificadores', FALSE, TRUE),
  ('Refis', 'refis', TRUE, TRUE),
  ('Peças', 'pecas', TRUE, TRUE);

ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id INTEGER REFERENCES categories(id);

UPDATE products SET category_id = c.id
FROM categories c WHERE products.category = c.slug;

ALTER TABLE products ALTER COLUMN category_id SET NOT NULL;
ALTER TABLE products DROP COLUMN IF EXISTS category;

DROP INDEX IF EXISTS idx_products_category;
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);

COMMIT;
```

---

## 5. Implementation Order

1. **Database:** Run migration script.
2. **`internal/products/products.go`:**
   - Add `Category` struct.
   - Update `Product` struct.
   - Implement category CRUD functions.
   - Update all product queries to JOIN categories.
   - Update `CreateProduct` / `UpdateProduct` signatures.
   - Replace `isPartsCategory` with DB flag lookup.
3. **`internal/offers/offers.go`:** Update queries and `Offer` struct.
4. **`cmd/server/main.go`:**
   - Remove hardcoded routes and functions.
   - Add consolidated `/products/` route.
   - Add admin category API routes.
   - Update admin dashboard and edit handlers.
   - Update product detail handler.
5. **Templates (Admin):**
   - Update `admin-dashboard.html` category select.
   - Update `admin-edit-form.html` category select.
   - Create `admin-categories.html`.
   - Create `admin-category-modal.html`.
   - Create `admin-category-options.html`.
   - Update product list/card templates.
   - Update offers list template.
6. **Templates (Store):**
   - Update `index.html` category filter to `<select>`.
   - Update `product.html` to use `CategoryName` and `AllowsCompatibility`.
7. **Test:** Build, run, verify admin CRUD, store filtering, product detail compatibility, offers display.

---

## 6. Behavior Summary

| Scenario | Behavior |
|---|---|
| Category is **active** | Appears in store `<select>`, products in it appear when filtered. |
| Category is **inactive** | Hidden from store `<select>`, but products still appear in "Todos" and search. |
| Editing a product in an inactive category | Admin dropdown shows all categories; inactive ones marked "(Inativo)". |
| Creating a category | Admin specifies name and whether it allows compatibility. Slug auto-generated. |
| Editing a category | Name and compatibility flag can be changed. Slug stays immutable (to preserve URLs). |
| Compatibility check | Uses `categories.allows_compatibility` flag; no hardcoded slugs. |

---

## Labels (Português)

| English Concept | Portuguese Label |
|---|---|
| Category | Categoria |
| Categories | Categorias |
| Add Category | Adicionar Categoria |
| Manage Categories | Gerenciar Categorias |
| Allows Compatibility | Permite compatibilidade |
| Active | Ativo |
| Inactive | Inativo |
| (Inactive) | (Inativo) |
| All | Todos |
| Offers | Ofertas |
| Select a category | Selecione uma categoria |
