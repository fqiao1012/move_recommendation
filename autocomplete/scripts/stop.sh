
#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

echo "Stopping autocomplete service..."
docker compose stop autocomplete

echo "Autocomplete service stopped."
