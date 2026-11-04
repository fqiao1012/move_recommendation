
#!/usr/bin/env bash
set -euo pipefail

# Resolve project root from this script location.
ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

cd "$ROOT_DIR"

echo "Waiting for SQL Server..."

until docker compose exec -T movie-db sh -c \
  '/opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -b -Q "SELECT 1" -o /dev/null' \
  >/dev/null 2>&1; do
    sleep 2
done

echo "Loading movie data..."

docker compose exec -T movie-db sh -c \
  'exec /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -b -i /seed/init.sql'

echo "Movie database initialized successfully!"
