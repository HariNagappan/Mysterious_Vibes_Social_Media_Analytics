#!/usr/bin/env bash
# Applies the embedded database migrations by running the api binary in
# -migrate-only mode.
#
#   ./scripts/run_migrations.sh            # local toolchain (go run)
#   ./scripts/run_migrations.sh --docker   # through docker compose
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ "${1:-}" == "--docker" ]]; then
  docker compose -f deployments/docker-compose.yml run --rm api -migrate-only
else
  go run ./cmd/api -migrate-only
fi
