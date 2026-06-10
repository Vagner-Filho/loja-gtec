UPDATE orders SET address_number = '' WHERE address_number IS NULL;
ALTER TABLE orders ALTER COLUMN address_number SET DEFAULT '';
ALTER TABLE orders ALTER COLUMN address_number SET NOT NULL;
