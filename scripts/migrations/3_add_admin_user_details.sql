ALTER TABLE admin_users
ADD COLUMN IF NOT EXISTS email TEXT UNIQUE,
ADD COLUMN IF NOT EXISTS cpf TEXT UNIQUE,
ADD COLUMN IF NOT EXISTS phone TEXT;

-- Backfill existing rows with placeholder values so NOT NULL can be enforced.
UPDATE admin_users SET email = CONCAT('user_', id, '@placeholder.local') WHERE email IS NULL;
UPDATE admin_users SET cpf = LPAD(id::TEXT, 11, '0') WHERE cpf IS NULL;
UPDATE admin_users SET phone = '0000000000' WHERE phone IS NULL;

-- Make columns NOT NULL after backfilling.
ALTER TABLE admin_users ALTER COLUMN email SET NOT NULL;
ALTER TABLE admin_users ALTER COLUMN cpf SET NOT NULL;
ALTER TABLE admin_users ALTER COLUMN phone SET NOT NULL;
