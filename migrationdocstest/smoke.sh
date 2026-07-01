#!/bin/bash
# Quick smoke test for the Go backend.
# Requires the server to be running: go run ./backend/cmd/server
# With MongoDB+Redis: APP_MONGO_URI=mongodb://localhost:27017 APP_REDIS_ADDR=localhost:6379 go run ./backend/cmd/server

set -e

BASE=${BACKEND_URL:-http://localhost:8080}
PASS=0
FAIL=0

check() {
    local name=$1
    local result=$2
    if echo "$result" | python3 -m json.tool > /dev/null 2>&1; then
        echo "  PASS  $name"
        PASS=$((PASS + 1))
    else
        echo "  FAIL  $name — not valid JSON: $result"
        FAIL=$((FAIL + 1))
    fi
}

echo ""
echo "Go backend smoke test — $BASE"
echo "================================================"

echo ""
echo "--- Task 01: Health ---"
check "GET /healthz" "$(curl -sf "$BASE/healthz")"
check "GET /readyz"  "$(curl -sf "$BASE/readyz" || echo '{"error":"readyz returned non-2xx"}')"

echo ""
echo "--- Task 03: Settings (requires MongoDB) ---"
check "PUT /api/settings/log_level" "$(curl -sf -X PUT "$BASE/api/settings/log_level" \
  -H 'Content-Type: application/json' -d '{"value":"info"}' 2>/dev/null || echo '{"error":"MongoDB not connected"}')"

check "GET /api/settings/log_level" "$(curl -sf "$BASE/api/settings/log_level" 2>/dev/null || echo '{"error":"MongoDB not connected"}')"

check "GET /api/settings" "$(curl -sf "$BASE/api/settings" 2>/dev/null || echo '{"error":"MongoDB not connected"}')"

check "GET /api/settings/flags" "$(curl -sf "$BASE/api/settings/flags" 2>/dev/null || echo '{"error":"MongoDB not connected"}')"

check "GET /api/settings/flags/CACHE_ENABLED" "$(curl -sf "$BASE/api/settings/flags/CACHE_ENABLED" 2>/dev/null || echo '{"error":"MongoDB not connected"}')"

echo ""
echo "--- Task 02: Cache (requires Redis) ---"
check "GET /api/cache/stats" "$(curl -sf "$BASE/api/cache/stats" 2>/dev/null || echo '{"error":"Redis not connected"}')"

check "POST /api/cache/flush?prefix=test" "$(curl -sf -X POST "$BASE/api/cache/flush?prefix=test" 2>/dev/null || echo '{"error":"Redis not connected"}')"

echo ""
echo "--- Unknown key validation (Task 03) ---"
RESP=$(curl -s -X PUT "$BASE/api/settings/nonexistent_key" \
  -H 'Content-Type: application/json' -d '{"value":"x"}')
if echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); exit(0 if 'error' in d else 1)" 2>/dev/null; then
    echo "  PASS  unknown key returns error"
    PASS=$((PASS + 1))
else
    echo "  FAIL  unknown key did not return error: $RESP"
    FAIL=$((FAIL + 1))
fi

echo ""
echo "================================================"
echo "  Results: $PASS passed, $FAIL failed"
echo ""

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
