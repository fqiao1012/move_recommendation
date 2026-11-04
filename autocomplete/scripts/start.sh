
#!/usr/bin/env bash
set -euo pipefail

# Resolve project root
ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

echo "Starting SQL Server..."
docker compose up -d movie-db

echo "Waiting for SQL Server to become healthy..."
until docker compose exec -T movie-db sh -c \
  '/opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -b -Q "SELECT 1" -o /dev/null' \
  >/dev/null 2>&1; do
    if [ "$(docker compose ps -q movie-db)" = "" ]; then
      echo "SQL Server container not found."
      exit 1
    fi
    sleep 2
done

echo "Building and starting autocomplete..."
docker compose up -d --build autocomplete

echo "Autocomplete service status:"
docker compose ps

echo "Test API:"
echo "curl 'http://localhost:8080/v1/autocomplete?q=star&limit=5'"
