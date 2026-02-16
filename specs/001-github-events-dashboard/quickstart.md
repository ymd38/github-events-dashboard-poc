# Quickstart: GitHub Events Dashboard

**Date**: 2026-02-15
**Feature**: 001-github-events-dashboard

---

## Prerequisites

- Docker & Docker Compose
- GitHubアカウント
- ngrok（ローカルでWebhookを受信する場合）

---

## 1. リポジトリのクローンと環境変数設定

```bash
git clone <repository-url>
cd github-events-dashboard-poc
cp .env.example .env
```

`.env` を編集し、以下の値を設定する:

```env
# GitHub OAuth App（https://github.com/settings/developers で作成）
GITHUB_CLIENT_ID=your_client_id
GITHUB_CLIENT_SECRET=your_client_secret

# GitHub Webhook Secret（任意の文字列。GitHub Webhook設定時に同じ値を使用）
GITHUB_WEBHOOK_SECRET=your_webhook_secret

# MySQL
MYSQL_HOST=db
MYSQL_PORT=3306
MYSQL_USER=dashboard
MYSQL_PASSWORD=dashboard_password
MYSQL_DATABASE=github_events
MYSQL_ROOT_PASSWORD=root_password

# Backend
BACKEND_PORT=8080
FRONTEND_URL=http://localhost:3000
SESSION_SECRET=your_session_secret

# Frontend
NUXT_PUBLIC_API_BASE=http://localhost:8080
```

---

## 2. GitHub OAuth App の作成

1. https://github.com/settings/developers にアクセス
2. 「New OAuth App」をクリック
3. 以下を入力:
   - **Application name**: GitHub Events Dashboard (Dev)
   - **Homepage URL**: `http://localhost:3000`
   - **Authorization callback URL**: `http://localhost:8080/api/auth/callback`
4. 「Register application」をクリック
5. Client ID と Client Secret を `.env` に設定

---

## 3. 全サービスの起動

```bash
docker-compose up
```

以下のサービスが起動する:

| サービス | URL | 説明 |
|----------|-----|------|
| Frontend | http://localhost:3000 | Nuxt 3 ダッシュボード |
| Backend | http://localhost:8080 | Go API サーバー |
| MySQL | localhost:3306 | データベース |

---

## 4. ヘルスチェック

```bash
curl http://localhost:8080/api/health
```

期待されるレスポンス:

```json
{
  "status": "healthy",
  "checks": {
    "database": "up"
  }
}
```

---

## 5. GitHub Webhook の設定

### ローカル環境の場合（ngrok使用）

```bash
ngrok http 8080
```

ngrokが表示するURLをメモする（例: `https://abc123.ngrok.io`）

### Webhook設定

1. 対象リポジトリの Settings → Webhooks → Add webhook
2. 以下を入力:
   - **Payload URL**: `https://abc123.ngrok.io/api/webhook`（またはデプロイ先URL）
   - **Content type**: `application/json`
   - **Secret**: `.env` の `GITHUB_WEBHOOK_SECRET` と同じ値
   - **Which events**: 「Let me select individual events」を選択
     - ✅ Issues
     - ✅ Pull requests
3. 「Add webhook」をクリック

---

## 6. 動作確認

1. http://localhost:3000 にアクセス
2. GitHubアカウントでログイン
3. 対象リポジトリでIssueを作成
4. ダッシュボードにリアルタイムでイベントが表示されることを確認

---

## 開発時のコマンド

```bash
# 全サービス起動（バックグラウンド）
docker-compose up -d

# ログ確認
docker-compose logs -f backend
docker-compose logs -f frontend

# 全サービス停止
docker-compose down

# DBデータも含めて完全リセット
docker-compose down -v
```

---

## API エンドポイント一覧

| Method | Path | 認証 | 説明 |
|--------|------|------|------|
| POST | `/api/webhook` | Webhook署名 | GitHub Webhookイベント受信 |
| GET | `/api/events` | Session | イベント一覧（ページネーション付き） |
| GET | `/api/events/:id` | Session | イベント詳細 |
| GET | `/api/events/stream` | Session | SSEイベントストリーム |
| GET | `/api/auth/login` | - | GitHub OAuthログイン開始 |
| GET | `/api/auth/callback` | - | GitHub OAuthコールバック |
| POST | `/api/auth/logout` | Session | ログアウト |
| GET | `/api/auth/me` | Session | 現在のユーザー情報 |
| GET | `/api/health` | - | ヘルスチェック |
