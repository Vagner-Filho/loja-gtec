-- Migration 002: Update existing products to have is_available = true
-- This ensures all existing products are marked as available

UPDATE products SET is_available = true WHERE is_available IS NULL;

-- Add a default value for future inserts
ALTER TABLE products ALTER COLUMN is_available SET DEFAULT true;