# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GitHub Events Dashboard — a real-time webhook event viewer that captures GitHub Issues and Pull Request events. Go backend + Nuxt 3 frontend + MySQL, with SSE for real-time streaming.

## Development Commands

### Docker Compose (primary development method)

```bash
make up              # Start all services (backend:8080, frontend:3000, MySQL:3306)
make build           # Build and start all services
make down            # Stop all services
make clean           # Stop and remove volumes (resets DB)
make restart         # Restart all services
make logs-backend    # Follow backend logs
make logs-frontend   # Follow frontend logs
make backend-shell   # Shell into backend container
make db-shell        # MySQL CLI access
make ngrok           # Expose backend for webhook testing
```

### Frontend (inside container or locally)

```bash
cd frontend
npm install
npm run dev          # Dev server with hot reload
npm run build        # Production build
npm run test         # Run Vitest tests
```

### Backend

Backend uses Air for hot reload inside Docker (configured in `.air.toml`). Manual build:

```bash
cd backend
go build -o ./tmp/main ./cmd/server/main.go
go test ./...        # Run all tests
go test ./internal/handler/  # Run tests for a specific package
```

## Architecture

### Backend (Go 1.24, chi router)

Layered architecture in `backend/internal/`:

- **cmd/server/main.go** — Entry point. Initializes DB, router, middleware, graceful shutdown.
- **handler/** — HTTP handlers: `webhook.go` (receives GitHub events), `events.go` (list/detail), `sse.go` (event stream), `health.go`
- **service/event_service.go** — Business logic. Parses webhook payloads, filters events (Issues opened, PRs merged only), truncates body to 500 chars.
- **repository/** — MySQL data access. `event_repository.go`, `user_repository.go`, `db.go`.
- **auth/** — GitHub OAuth 2.0 flow (`oauth.go`) and session management (`session.go`).
- **crypto/crypto.go** — AES-256-GCM encryption for storing OAuth tokens.
- **middleware/** — CORS, auth (session validation), recovery, logger.
- **sse/hub.go** — SSE client hub for broadcasting events to connected frontends.
- **model/** — Data structs for events and users.

### Frontend (Nuxt 3, Vue 3 + TypeScript, TailwindCSS, Pinia)

- **pages/** — `index.vue` (dashboard), `login.vue`
- **composables/** — `useEvents.ts` (event state + pagination), `useAuth.ts` (auth state), `useSSE.ts` (SSE connection with exponential backoff reconnect 3s→30s)
- **components/** — `EventList`, `EventDetail`, `EventFilter`, `Pagination`, `AppHeader`
- **middleware/auth.global.ts** — Route guard: unauthenticated → `/login`, authenticated away from `/login`

### Key Data Flows

1. **Webhook ingestion**: GitHub POST → `/api/webhook` → HMAC-SHA256 verify → EventService parses → DB insert → SSE broadcast to all clients
2. **Auth**: Login redirect → GitHub OAuth → callback exchanges code → encrypts token → creates session cookie
3. **Real-time**: Frontend connects to `/api/events/stream` (SSE) → receives `new_event` messages → prepends to event list

### API Routes (defined in main.go)

Public: `/api/health`, `/api/webhook` (HMAC), `/api/auth/{login,callback,me,logout}`
Protected (session required): `/api/events`, `/api/events/{id}`, `/api/events/stream`

### Database (MySQL 8.0)

Migrations in `db/migrations/`. Two tables:
- **events** — `delivery_id` is UNIQUE for idempotency. Indexed on `event_type`, `received_at`, `repo_name`.
- **users** — `github_id` is UNIQUE. `access_token` stored encrypted.

## Environment Setup

Copy `.env.example` to `.env` and configure GitHub OAuth credentials, webhook secret, session secret, and token encryption key (64 hex chars for AES-256).
