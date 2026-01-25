-- Add role column to admin users
ALTER TABLE admin_users
    ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'admin';

UPDATE admin_users
SET role = 'admin'
WHERE role IS NULL;
