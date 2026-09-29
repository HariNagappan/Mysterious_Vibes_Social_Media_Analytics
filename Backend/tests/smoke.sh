#!/usr/bin/env bash
# Black-box smoke test for a locally running PulseGraph stack.
set -euo pipefail

BASE="${BASE:-http://localhost:8080}"

check() { # name url expected_status
  local name="$1" url="$2" expected="$3" status
  status=$(curl -s -o /dev/null -w "%{http_code}" "$url")
  if [[ "$status" == "$expected" ]]; then
    echo "PASS  $name ($status)"
  else
    echo "FAIL  $name (expected $expected, got $status)"
    exit 1
  fi
}

check "api liveness"                 "$BASE/healthz"                200
check "api readiness"                "$BASE/readyz"                 200
check "openapi spec served"          "$BASE/swagger/openapi.yaml"   200
check "metrics exposed"              "$BASE/metrics"                200
check "unauthenticated API rejected" "$BASE/api/v1/topics/1/dashboard" 401

echo "smoke: all checks passed"
