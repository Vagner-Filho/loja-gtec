#!/bin/bash

echo "Running database migration..."
echo "Please run the following command manually if psql is installed:"
echo ""
echo "  PGPASSWORD=postgres psql -U lojagtec -d lojagtec -f scripts/migrations/001_create_admin_and_products.sql"
echo ""
echo "Or connect to your PostgreSQL database and run the SQL file:"
echo "  psql -U lojagtec -d lojagtec -f scripts/migrations/001_create_admin_and_products.sql"
