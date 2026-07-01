# OmniRouter Frontend

This directory contains the **Next.js dashboard** for OmniRouter — migrated from
`../OmniRoute/` so the full stack (Go backend + Next.js UI) can be developed and
tested in one place.

---

## Overview

| Service | Tech | Default Port | What it serves |
|---------|------|:---:|----------------|
| **Go backend** (`../backend/`) | Go 1.26, Hexagonal Arch | `8080` | `/api/*` REST + `/healthz` / `/readyz` |
| **Next.js frontend** (`frontend/`) | Next.js 16, TypeScript | `20128` | Dashboard UI + legacy API routes not yet on the Go backend |

The Go backend is the authoritative backend. As migration slices are completed
(Tasks 01–21), more `/api/...` routes move from Next.js to Go. For routes that
have **already been migrated**, set `NEXT_PUBLIC_GO_BACKEND_URL` (see below) so
the dashboard hits the Go backend instead of the Next.js API routes.

---

## Quick Start

### 1. Prerequisites

- Node.js ≥ 22.0.0 (see `.nvmrc` or `package.json` `engines` field)
- Go 1.26+ (for the backend)
- MongoDB (optional — needed for settings, cache, audit slices)
- Redis (optional — needed for cache slice)

### 2. Install frontend dependencies

```bash
cd frontend
npm install
```

Or via Mage from the repo root:

```bash
mage frontend:setup
```

### 3. Configure environment

```bash
cd frontend
cp .env.example .env
```

Edit `.env` — the minimum required values:

```dotenv
# Required — generate strong random values before first use
JWT_SECRET=$(openssl rand -base64 48)
API_KEY_SECRET=$(openssl rand -hex 32)
INITIAL_PASSWORD=changeme

# Optional — point the dashboard at the Go backend for migrated routes
# NEXT_PUBLIC_GO_BACKEND_URL=http://localhost:8080
```

### 4. Run the frontend alone

```bash
cd frontend
npm run dev          # http://localhost:20128
```

Or via Mage:

```bash
mage frontend:dev
```

### 5. Run the Go backend

From the repo root:

```bash
# Without external deps (health + audit + basic routes):
go run ./backend/cmd/server

# With MongoDB + Redis (enables settings, cache):
APP_MONGO_URI=mongodb://localhost:27017 APP_REDIS_ADDR=localhost:6379 \
  go run ./backend/cmd/server
```

---

## Running Both Together

Open two terminals from the repo root:

**Terminal 1 — Go backend (port 8080):**

```bash
APP_MONGO_URI=mongodb://localhost:27017 APP_REDIS_ADDR=localhost:6379 \
  go run ./backend/cmd/server
```

**Terminal 2 — Next.js frontend (port 20128):**

```bash
cd frontend && npm run dev
```

Then open `http://localhost:20128`. The dashboard UI runs from the Next.js app.
API routes that have been migrated to Go are proxied/called directly at
`http://localhost:8080` when `NEXT_PUBLIC_GO_BACKEND_URL` is set.

---

## API Route Split

Routes below are migrated to the Go backend (Tasks 01–03 complete).
For the rest, the Next.js built-in API routes are still active.

| Routes | Status | Served by |
|--------|--------|-----------|
| `GET /healthz`, `GET /readyz` | ✅ Go | Go backend `:8080` |
| `GET /api/audit`, `POST /api/audit` | ✅ Go | Go backend `:8080` |
| `GET /api/cache/stats`, `GET /api/cache/entries`, etc. | ✅ Go | Go backend `:8080` |
| `GET /api/settings`, `PUT /api/settings/{key}`, etc. | ✅ Go | Go backend `:8080` |
| `GET /api/settings/flags`, `PUT /api/settings/flags/{name}` | ✅ Go | Go backend `:8080` |
| Everything else (`/api/v1/chat/completions`, providers, models, …) | 🔲 Pending | Next.js `:20128` |

---

## Available npm Scripts

```bash
npm run dev          # Dev server at http://localhost:20128
npm run build        # Production Next.js build
npm run lint         # ESLint (0 errors expected)
npm run typecheck:core   # TypeScript check (strict)
npm run test:unit    # Unit test suite (Node native runner)
npm run test:vitest  # Vitest suite (MCP, autoCombo, cache)
npm run test:coverage # Coverage gate (60/60/60/60)
npm run check        # lint + test combined
```

Or run any of these via Mage from the repo root:

```bash
mage frontend:dev
mage frontend:build
mage frontend:lint
mage frontend:test
```

---

## Environment Variables

| Variable | Required | Description |
|----------|:--------:|-------------|
| `JWT_SECRET` | Yes | Signs dashboard session cookies (`openssl rand -base64 48`) |
| `API_KEY_SECRET` | Yes | Encrypts API keys at rest in SQLite (`openssl rand -hex 32`) |
| `INITIAL_PASSWORD` | Yes | First-login password (change after first use) |
| `PORT` | No | Override default port `20128` |
| `NEXT_PUBLIC_GO_BACKEND_URL` | No | URL of the Go backend for migrated routes (e.g. `http://localhost:8080`) |
| `APP_LOG_LEVEL` | No | Log level: `debug` / `info` / `warn` / `error` |

See `.env.example` for the full list with descriptions.

---

## Frontend-only vs Full Stack

| Scenario | What to start | Notes |
|----------|---------------|-------|
| UI development only | `npm run dev` | Next.js API routes serve all `/api/*` locally |
| Test migrated Go routes | `npm run dev` + Go backend | Set `NEXT_PUBLIC_GO_BACKEND_URL=http://localhost:8080` |
| Production | Docker Compose (see `../docker-compose.yml`) | Both services in containers |

---

## Tests

```bash
# Unit tests (Node native runner)
npm run test:unit

# Vitest (MCP tools, autoCombo, cache)
npm run test:vitest

# Both
npm run test:all

# Coverage (must stay ≥ 60% statements/lines/functions/branches)
npm run test:coverage
```

---

## Build System

The `mage frontend:*` targets at the repo root (driven by `../node.yaml`) delegate
to the npm scripts above. No Makefiles or shell scripts — everything goes through Mage.
