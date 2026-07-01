# Migration Test Guide

How to test everything implemented so far in the shroute Go backend migration.
Run these in order — prerequisites first, then unit tests, then start the server
for manual API testing, then integration tests if Docker is available.

---

## What Has Been Implemented

| Task | Slice | Routes |
|------|-------|--------|
| 01 | Health, Logs & Audit | `GET /healthz`, `GET /readyz`, `POST /api/audit`, `GET /api/audit` |
| 02 | Cache Management | `GET /api/cache/stats`, `GET /api/cache/entries`, `POST /api/cache/flush` |
| 03 | Settings & Feature Flags | `GET/PUT/DELETE /api/settings/{key}`, `GET /api/settings`, `GET /api/settings/flags`, `GET/PUT /api/settings/flags/{name}` |
| — | Frontend | Next.js dashboard in `frontend/` |

---

## Prerequisites

```bash
# Go 1.26+
go version

# Docker (for integration tests — testcontainers spins MongoDB/Redis automatically)
docker info

# Node.js ≥ 22 (for frontend tests)
node --version

# golangci-lint (for lint check)
$(go env GOPATH)/bin/golangci-lint --version
# If missing: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

---

## 1. Build Check

Verify the entire codebase compiles cleanly.

```bash
cd /path/to/shroute

go build ./...
```

Expected: no output, exit 0.

---

## 2. Unit Tests — All Slices at Once

```bash
go test ./backend/...
```

Expected output (all `ok`, no `FAIL`):

```
ok   github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/handlers
ok   github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/middleware
ok   github.com/vinaycharlie01/shroute/backend/internal/application/audit
ok   github.com/vinaycharlie01/shroute/backend/internal/application/cache
ok   github.com/vinaycharlie01/shroute/backend/internal/application/health
ok   github.com/vinaycharlie01/shroute/backend/internal/application/settings
ok   github.com/vinaycharlie01/shroute/backend/internal/domain/audit
ok   github.com/vinaycharlie01/shroute/backend/internal/domain/health
ok   github.com/vinaycharlie01/shroute/backend/internal/domain/settings
ok   github.com/vinaycharlie01/shroute/backend/internal/infrastructure/config
```

### Run a specific slice's tests

```bash
# Task 01 — Audit
go test ./backend/internal/domain/audit/...
go test ./backend/internal/application/audit/...

# Task 02 — Cache
go test ./backend/internal/application/cache/...

# Task 03 — Settings
go test ./backend/internal/domain/settings/...
go test ./backend/internal/application/settings/...

# All handlers (tasks 01–03 together)
go test ./backend/internal/adapters/inbound/http/handlers/...
```

### Run with verbose output to see each test name

```bash
go test -v ./backend/internal/application/settings/...
```

### Run with race detector

```bash
go test -race ./backend/...
```

### Run with coverage

```bash
go test -cover ./backend/...
```

---

## 3. Lint Check

```bash
$(go env GOPATH)/bin/golangci-lint run --timeout=10m
```

Expected: `0 issues.`

---

## 4. Start the Go Backend

### Without external dependencies (Task 01 health routes only)

```bash
go run ./backend/cmd/server
```

Server starts on `http://localhost:8080`.

### With MongoDB (Task 01 Audit + Task 03 Settings)

```bash
APP_MONGO_URI=mongodb://localhost:27017 go run ./backend/cmd/server
```

### With MongoDB + Redis (all three tasks)

```bash
APP_MONGO_URI=mongodb://localhost:27017 APP_REDIS_ADDR=localhost:6379 \
  go run ./backend/cmd/server
```

You should see log output like:
```
level=INFO msg=starting env=local version=dev
level=INFO msg=listening addr=0.0.0.0:8080
```

---

## 5. Manual API Tests — Task 01: Health & Audit

Keep the server running in one terminal, run these in another.

### Health

```bash
# Liveness probe — always 200 if the process is up
curl -s http://localhost:8080/healthz | python3 -m json.tool

# Readiness probe — 200 when all deps are up, 503 if any are down
curl -s http://localhost:8080/readyz | python3 -m json.tool
```

Expected (`/healthz`):
```json
{"status": "ok", "checks": [{"name": "self", "status": "ok"}]}
```

### Audit (requires MongoDB)

```bash
# Record an audit entry
curl -s -X POST http://localhost:8080/api/audit \
  -H "Content-Type: application/json" \
  -d '{"action":"test.login","actor":"user@example.com","resource":"session","status":"success"}' \
  | python3 -m json.tool

# List audit entries
curl -s "http://localhost:8080/api/audit?limit=10" | python3 -m json.tool
```

Expected POST response:
```json
{"id": "...", "action": "test.login", "actor": "user@example.com", "status": "success", ...}
```

---

## 6. Manual API Tests — Task 02: Cache Management

Requires Redis (`APP_REDIS_ADDR=localhost:6379`).

```bash
# Cache stats
curl -s http://localhost:8080/api/cache/stats | python3 -m json.tool

# List cache entries by prefix
curl -s "http://localhost:8080/api/cache/entries?prefix=test&limit=20" | python3 -m json.tool

# Flush cache by prefix
curl -s -X POST "http://localhost:8080/api/cache/flush?prefix=test" | python3 -m json.tool

# Flush entire cache (all=true)
curl -s -X POST "http://localhost:8080/api/cache/flush?all=true" | python3 -m json.tool
```

Expected stats response:
```json
{"hits": 0, "misses": 0, "entries": 0, "memory_bytes": 0}
```

---

## 7. Manual API Tests — Task 03: Settings & Feature Flags

Requires MongoDB (`APP_MONGO_URI=mongodb://localhost:27017`).

### Settings CRUD

```bash
# List all settings (empty at first)
curl -s http://localhost:8080/api/settings | python3 -m json.tool

# Set a setting — known keys only
curl -s -X PUT http://localhost:8080/api/settings/log_level \
  -H "Content-Type: application/json" \
  -d '{"value": "debug"}' \
  | python3 -m json.tool

# Get the setting back
curl -s http://localhost:8080/api/settings/log_level | python3 -m json.tool

# Set another setting
curl -s -X PUT http://localhost:8080/api/settings/theme \
  -H "Content-Type: application/json" \
  -d '{"value": "dark"}' \
  | python3 -m json.tool

# List all settings (should show 2 now)
curl -s http://localhost:8080/api/settings | python3 -m json.tool

# Delete a setting
curl -s -X DELETE http://localhost:8080/api/settings/theme
# Expected: HTTP 204 No Content

# Try an unknown key — should get 400
curl -s -X PUT http://localhost:8080/api/settings/unknown_key \
  -H "Content-Type: application/json" \
  -d '{"value": "x"}' \
  | python3 -m json.tool
# Expected: {"error": "unknown setting key"}
```

**All valid setting keys:**

| Key | Category | Example value |
|-----|----------|---------------|
| `log_level` | `general` | `"debug"` |
| `max_tokens_default` | `general` | `4096` |
| `theme` | `ui` | `"dark"` |
| `require_api_key` | `security` | `true` |
| `allowed_origins` | `security` | `["https://example.com"]` |
| `rate_limit_enabled` | `security` | `false` |
| `default_model` | `routing` | `"gpt-4o"` |
| `fallback_enabled` | `routing` | `true` |

### Feature Flags

```bash
# List all feature flags (seeded automatically at startup)
curl -s http://localhost:8080/api/settings/flags | python3 -m json.tool

# Get a specific flag
curl -s http://localhost:8080/api/settings/flags/CACHE_ENABLED | python3 -m json.tool

# Enable a flag
curl -s -X PUT http://localhost:8080/api/settings/flags/RATE_LIMIT_ENABLED \
  -H "Content-Type: application/json" \
  -d '{"enabled": true}' \
  | python3 -m json.tool

# PII flags must stay false — this sets it true but the seeded default stays false
# (re-seeding on next restart will NOT overwrite operator-set values)
curl -s http://localhost:8080/api/settings/flags/PII_REDACTION_ENABLED | python3 -m json.tool
# Expected: "enabled": false, "default_value": "false"
```

**All seeded feature flags:**

| Flag | Default | Meaning |
|------|---------|---------|
| `PII_REDACTION_ENABLED` | `false` | Opt-in only — never enable by default |
| `PII_RESPONSE_SANITIZATION` | `false` | Opt-in only — never enable by default |
| `RATE_LIMIT_ENABLED` | `false` | Rate limiting on/off |
| `CACHE_ENABLED` | `true` | Redis cache on/off |
| `AUDIT_LOG_ENABLED` | `true` | Audit logging on/off |

---

## 8. Integration Tests (requires Docker)

Testcontainers automatically starts real MongoDB and Redis instances — no manual setup needed.

```bash
# All integration tests
go test -tags=integration ./backend/test/integration/...

# Verbose — see each test name
go test -v -tags=integration ./backend/test/integration/...

# Individual test files
go test -v -tags=integration ./backend/test/integration/ -run TestAuditRepository_AppendAndListAgainstRealMongoDB
go test -v -tags=integration ./backend/test/integration/ -run TestCacheStore_StatsAgainstRealRedis
go test -v -tags=integration ./backend/test/integration/ -run TestCacheStore_ListAgainstRealRedis
go test -v -tags=integration ./backend/test/integration/ -run TestCacheStore_FlushAgainstRealRedis_Prefix
go test -v -tags=integration ./backend/test/integration/ -run TestSettingsRepository_SetGetDeleteAgainstRealMongoDB
go test -v -tags=integration ./backend/test/integration/ -run TestFeatureFlagRepository_SeedDefaultsAndSet
go test -v -tags=integration ./backend/test/integration/ -run TestAdapter_PingAgainstRealMongoDB
go test -v -tags=integration ./backend/test/integration/ -run TestAdapter_PingAgainstRealRedis
```

What each integration test verifies:

| Test | What it checks |
|------|---------------|
| `TestAdapter_PingAgainstRealMongoDB` | MongoDB connection + ping |
| `TestAdapter_PingAgainstRealRedis` | Redis connection + ping |
| `TestAuditRepository_AppendAndListAgainstRealMongoDB` | Append audit entries, List with cursor, pagination |
| `TestCacheStore_StatsAgainstRealRedis` | Stats from empty cache |
| `TestCacheStore_ListAgainstRealRedis` | SET keys, List by prefix, limit |
| `TestCacheStore_FlushAgainstRealRedis_Prefix` | Flush keys matching prefix, verify others survive |
| `TestSettingsRepository_SetGetDeleteAgainstRealMongoDB` | Set → Get → List → Delete → Get returns ErrNotFound |
| `TestFeatureFlagRepository_SeedDefaultsAndSet` | PII flag default=false; `$setOnInsert` never overwrites operator value |

---

## 9. Via Mage Targets

All the above can be run through the Mage build system from the repo root.

```bash
# Install mage once
go install github.com/magefile/mage@latest

mage test          # unit tests (go test ./backend/...)
mage race          # unit tests with -race
mage coverage      # unit tests with coverage report
mage integration   # integration tests (-tags=integration, requires Docker)
mage lint          # golangci-lint
mage vet           # go vet
mage build         # compile binary
mage gen           # regenerate Swagger docs (swag init)
```

---

## 10. Frontend Tests

```bash
cd frontend

# Install dependencies (first time)
npm install

# Unit tests (Node native runner — covers most of the app)
npm run test:unit

# Vitest suite (MCP tools, autoCombo, cache)
npm run test:vitest

# Both suites
npm run test:all

# Coverage (must stay ≥ 60% statements/lines/functions/branches)
npm run test:coverage

# TypeScript check
npm run typecheck:core

# Lint
npm run lint
```

---

## 11. Full Stack — Run Both Services Together

Open two terminals.

**Terminal 1 — Go backend:**
```bash
APP_MONGO_URI=mongodb://localhost:27017 APP_REDIS_ADDR=localhost:6379 \
  go run ./backend/cmd/server
# Listening on http://localhost:8080
```

**Terminal 2 — Next.js frontend:**
```bash
cd frontend
cp .env.example .env
# Edit .env: set JWT_SECRET, API_KEY_SECRET, INITIAL_PASSWORD
npm install
npm run dev
# Dashboard at http://localhost:20128
```

Open `http://localhost:20128` in your browser. The dashboard UI loads from Next.js. Migrated API routes (health, audit, cache, settings) can be tested against the Go backend at port 8080.

---

## 12. Quick Smoke Test Script

Run this to verify the Go backend is healthy end-to-end in one go.

```bash
#!/bin/bash
set -e
BASE=http://localhost:8080

echo "=== Health ==="
curl -sf "$BASE/healthz" | python3 -m json.tool

echo "=== Readyz ==="
curl -sf "$BASE/readyz" | python3 -m json.tool

echo "=== Set log_level ==="
curl -sf -X PUT "$BASE/api/settings/log_level" \
  -H "Content-Type: application/json" \
  -d '{"value":"info"}' | python3 -m json.tool

echo "=== Get log_level ==="
curl -sf "$BASE/api/settings/log_level" | python3 -m json.tool

echo "=== List flags ==="
curl -sf "$BASE/api/settings/flags" | python3 -m json.tool

echo "=== Cache stats ==="
curl -sf "$BASE/api/cache/stats" | python3 -m json.tool

echo "=== All checks passed ==="
```

Save as `migrationdocstest/smoke.sh`, make executable with `chmod +x migrationdocstest/smoke.sh`, and run with the backend already started.

---

## 13. What Is NOT Yet Testable (Pending Tasks 04–21)

The following features are not yet implemented in the Go backend and are still served by the Next.js app in `frontend/`:

- API Keys (`/api/keys`)
- Models & Mappings (`/api/models`)
- Usage & Quota (`/api/usage`)
- Providers (`/api/providers`)
- Chat completions (`/api/v1/chat/completions`)
- All other `/api/v1/...` routes
- OAuth flows, MCP server, A2A protocol, Agent Bridge

These continue to work via the Next.js frontend's built-in API routes.
