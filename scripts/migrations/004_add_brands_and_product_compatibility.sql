-- Brands catalog
CREATE TABLE IF NOT EXISTS brands (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Products - Brands (many-to-many)
CREATE TABLE IF NOT EXISTS product_brands (
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    brand_id INTEGER NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, brand_id)
);

-- Parts - Compatible products (many-to-many)
CREATE TABLE IF NOT EXISTS product_compatibility (
    part_product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    fits_product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    PRIMARY KEY (part_product_id, fits_product_id)
);
