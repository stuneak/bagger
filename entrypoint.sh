#!/bin/sh
set -e

echo "Running database migrations..."
migrate -path /app/db/sqlc/migration -database "$DB_SOURCE" -verbose up

echo "Starting application..."
exec "$@"