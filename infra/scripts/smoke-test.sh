#!/usr/bin/env bash
# Smoke test for Entrega 1: checks that ms-auth, ms-video, ms-notify (stubs) respond correctly.
# Run from repo root after: docker compose -f infra/docker-compose.yml up -d --build
set -e

MS_AUTH_URL="${MS_AUTH_URL:-http://localhost:8081}"
MS_VIDEO_URL="${MS_VIDEO_URL:-http://localhost:8082}"
MS_NOTIFY_URL="${MS_NOTIFY_URL:-http://localhost:8083}"

failed=0

check_service() {
  local name="$1"
  local url="$2"
  local expected="$3"
  local body
  body=$(curl -sf --max-time 5 "$url/" 2>/dev/null) || true
  if [[ -z "$body" ]]; then
    echo "FAIL $name: no response from $url/"
    failed=1
    return
  fi
  if [[ "$body" != *"$expected"* ]]; then
    echo "FAIL $name: expected body to contain '$expected', got: $body"
    failed=1
    return
  fi
  echo "OK   $name: $url/ -> $body"
}

check_service "ms-auth"   "$MS_AUTH_URL"   "ms-auth"
check_service "ms-video"  "$MS_VIDEO_URL"  "ms-video"
check_service "ms-notify" "$MS_NOTIFY_URL" "ms-notify"

if [[ $failed -eq 1 ]]; then
  echo "Smoke test failed."
  exit 1
fi
echo "Smoke test passed."
exit 0
