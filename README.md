# GitHub Events Dashboard

GitHub Webhook イベントをリアルタイムで表示するダッシュボードアプリケーション。

## Tech Stack

- **Backend**: Go 1.21+ (chi router, go-sql-driver/mysql, gorilla/sessions, golang.org/x/oauth2)
- **Frontend**: Nuxt 3 (Vue 3 + TypeScript, TailwindCSS, Pinia)
- **Database**: MySQL 8.0
- **Real-time**: Server-Sent Events (SSE)
- **Auth**: GitHub OAuth 2.0
- **Container**: Docker + docker-compose

## Quick Start

### 1. Prerequisites

- Docker & Docker Compose
- GitHub OAuth App ([Settings > Developer settings > OAuth Apps](https://github.com/settings/developers))
  - **Homepage URL**: `http://localhost:3000`
  - **Authorization callback URL**: `http://localhost:8080/api/auth/callback`

### 2. Environment Setup

```bash
cp .env.example .env
```

Edit `.env` with your values:

```
GITHUB_CLIENT_ID=<your_oauth_app_client_id>
GITHUB_CLIENT_SECRET=<your_oauth_app_client_secret>
GITHUB_WEBHOOK_SECRET=<your_webhook_secret>
SESSION_SECRET=<random_string>
```

### 3. Start Services

```bash
docker-compose up
```

This starts:
- **MySQL** on port 3306 (auto-applies migrations)
- **Backend** on port 8080 (with hot-reload via Air)
- **Frontend** on port 3000 (Nuxt dev server)

### 4. Verify

```bash
curl http://localhost:8080/api/health
```

Expected: `{"status":"healthy","checks":{"database":"up"}}`

### 5. Configure GitHub Webhook

1. Go to your repository **Settings > Webhooks > Add webhook**
2. **Payload URL**: Use ngrok or similar to expose `http://localhost:8080/api/webhook`
3. **Content type**: `application/json`
4. **Secret**: Same as `GITHUB_WEBHOOK_SECRET` in `.env`
5. **Events**: Select "Issues" and "Pull requests"

#### Using ngrok for local development

```bash
ngrok http 8080
```

Use the generated HTTPS URL as the Payload URL.

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/health` | No | Health check |
| POST | `/api/webhook` | Signature | GitHub webhook receiver |
| GET | `/api/auth/login` | No | Start OAuth flow |
| GET | `/api/auth/callback` | No | OAuth callback |
| GET | `/api/auth/me` | No | Current user info |
| POST | `/api/auth/logout` | No | Logout |
| GET | `/api/events` | Yes | List events (paginated) |
| GET | `/api/events/{id}` | Yes | Event detail |
| GET | `/api/events/stream` | Yes | SSE event stream |

### Query Parameters for `/api/events`

- `page` (default: 1)
- `per_page` (default: 20, max: 100)
- `event_type` (optional: `issues` or `pull_request`)

## Project Structure

```
├── backend/
│   ├── cmd/server/main.go          # Entry point
│   ├── internal/
│   │   ├── auth/                   # OAuth & session management
│   │   ├── config/                 # Configuration loader
│   │   ├── handler/                # HTTP handlers
│   │   ├── middleware/             # HTTP middleware
│   │   ├── model/                  # Data models
│   │   ├── repository/            # Database operations
│   │   ├── service/               # Business logic
│   │   └── sse/                   # SSE hub
│   ├── Dockerfile
│   └── .air.toml
├── frontend/
│   ├── components/                 # Vue components
│   ├── composables/               # Vue composables
│   ├── middleware/                 # Nuxt middleware
│   ├── pages/                     # Nuxt pages
│   ├── nuxt.config.ts
│   └── Dockerfile
├── db/migrations/                  # SQL migrations
├── docker-compose.yml
└── .env.example
```