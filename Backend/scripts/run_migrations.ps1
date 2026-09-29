# Applies the embedded database migrations by running the api binary in
# -migrate-only mode. Requires a reachable PostgreSQL (start it first with
#   docker compose -f deployments/docker-compose.yml up -d postgres redis).
#
#   .\scripts\run_migrations.ps1            # local toolchain (go run)
#   .\scripts\run_migrations.ps1 -Docker    # through docker compose
param(
    [switch]$Docker
)
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if ($Docker) {
    docker compose -f deployments/docker-compose.yml run --rm api -migrate-only
}
else {
    go run ./cmd/api -migrate-only
}
exit $LASTEXITCODE
