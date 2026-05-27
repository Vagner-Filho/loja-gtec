#!/bin/bash

echo "Database migrations run automatically when the server starts."
echo ""
echo "To start the server and apply migrations:"
echo ""
echo "  go build -o lojagtec cmd/server/main.go"
echo "  ./lojagtec"
echo ""
echo "To run a specific migration manually:"
echo ""
echo "  psql -U lojagtec -d lojagtec -f scripts/migrations/1_baseline.sql"
