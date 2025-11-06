#!/bin/sh
set -e

echo "Starting PM Tool Backend..."

# Wait for database to be ready
echo "Waiting for database to be ready..."
max_attempts=30
attempt=0

until [ $attempt -ge $max_attempts ]
do
  if wget --no-verbose --tries=1 --spider http://db:5432 2>/dev/null || nc -z ${DB_HOST:-db} ${DB_PORT:-5432}; then
    echo "Database is ready!"
    break
  fi

  attempt=$((attempt+1))
  echo "Database is not ready yet (attempt $attempt/$max_attempts)..."
  sleep 2
done

if [ $attempt -ge $max_attempts ]; then
  echo "Warning: Could not connect to database after $max_attempts attempts"
  echo "Continuing anyway - migrations may fail if database is not ready"
fi

# Run database migrations
echo "Running database migrations..."
if ./main migrate up; then
  echo "Migrations completed successfully"
else
  echo "Warning: Migrations failed or no migrations to run"
fi

# Start the application
echo "Starting application server..."
exec ./main serve
