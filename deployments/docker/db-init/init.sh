#!/bin/sh
# deployments/docker/db-init/init.sh

# Wait for PostgreSQL to be fully ready
echo "Waiting for PostgreSQL to be ready..."
until PGPASSWORD=postgres psql -h postgres -U postgres -d podcast_platform -c "SELECT 1" > /dev/null 2>&1; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done
echo "PostgreSQL is up - executing migrations"

# Run migrations
echo "Running migrations..."
migrate -path ./scripts/migrations -database "postgres://postgres:postgres@postgres:5432/podcast_platform?sslmode=disable" -verbose up

# Seed database with initial data
echo "Seeding database..."
PGPASSWORD=postgres psql -h postgres -U postgres -d podcast_platform -f ./scripts/seed/seed_data.sql

echo "Database initialization completed"