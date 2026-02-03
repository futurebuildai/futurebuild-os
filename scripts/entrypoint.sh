#!/bin/sh
set -e

echo "=== FutureBuild API Startup ==="

# Run database migrations if DATABASE_URL is set
if [ -n "$DATABASE_URL" ]; then
    echo "Running database migrations..."
    # Use postgres:// scheme instead of postgresql:// for golang-migrate
    MIGRATE_URL=$(echo "$DATABASE_URL" | sed 's|^postgresql://|postgres://|')

    # golang-migrate returns 0 on success AND on "no change" (already applied).
    # Non-zero means a genuine failure (syntax error, lock, schema conflict).
    # We must NOT start the API on a failed migration — data corruption risk.
    if /app/migrate -database "$MIGRATE_URL" -path /app/migrations up; then
        echo "Migrations complete."
    else
        EXIT_CODE=$?
        echo "FATAL: Migration failed with exit code $EXIT_CODE"
        echo "Refusing to start API on a potentially corrupted database."
        exit "$EXIT_CODE"
    fi
else
    echo "WARNING: DATABASE_URL not set, skipping migrations"
fi

# Optional: Run integration readiness checks before starting the server.
# Set READINESS_CHECK_ON_STARTUP=true to enable. Logs warnings but does not block startup.
if [ "$READINESS_CHECK_ON_STARTUP" = "true" ]; then
    echo "Running integration readiness checks..."
    if /app/api --readiness-check; then
        echo "Readiness checks passed."
    else
        echo "WARNING: Readiness checks reported failures. Check /api/v1/readiness after startup."
    fi
fi

echo "Starting API server..."
exec /app/api
