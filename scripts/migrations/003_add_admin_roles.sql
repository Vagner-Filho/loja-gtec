-- Add role column to admin users with validation
ALTER TABLE admin_users
    ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('admin', 'product_admin'));

UPDATE admin_users
SET role = 'admin'
WHERE role IS NULL;
