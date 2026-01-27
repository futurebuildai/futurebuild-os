#!/bin/sh

echo "=== FutureBuild API Startup ==="

# Run database migrations if DATABASE_URL is set
if [ -n "$DATABASE_URL" ]; then
    echo "Running database migrations..."
    # Use postgres:// scheme instead of postgresql:// for golang-migrate
    MIGRATE_URL=$(echo "$DATABASE_URL" | sed 's|^postgresql://|postgres://|')
    if /app/migrate -database "$MIGRATE_URL" -path /app/migrations up; then
        echo "Migrations complete."
    else
        EXIT_CODE=$?
        echo "WARNING: Migration exited with code $EXIT_CODE"
        echo "This may be expected if migrations were already applied."
    fi
else
    echo "WARNING: DATABASE_URL not set, skipping migrations"
fi

echo "Starting API server..."
exec /app/api
