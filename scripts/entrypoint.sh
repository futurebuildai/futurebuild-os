#!/bin/sh
set -e

echo "=== FutureBuild API Startup ==="

# Run database migrations if DATABASE_URL is set
if [ -n "$DATABASE_URL" ]; then
    echo "Running database migrations..."
    /app/migrate -database "$DATABASE_URL" -path /app/migrations up
    echo "Migrations complete."
else
    echo "WARNING: DATABASE_URL not set, skipping migrations"
fi

echo "Starting API server..."
exec /app/api
