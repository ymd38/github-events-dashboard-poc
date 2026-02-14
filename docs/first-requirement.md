# GitHub Events Dashboard PoC — 初期要件定義

> **プロジェクト名**: github-events-dashboard-poc
> **ドキュメント更新日**: 2026-02-15
> **ステータス**: 初期要件確定 → spec-kit による仕様策定へ進行

---

## 1. プロジェクト概要

### 1.1 目的

GitHubリポジトリのイベント（Issue作成、PRマージなど）をWebhookで受信し、Web UIでリアルタイム通知およびログ表示するダッシュボードサービスを構築する。

### 1.2 位置づけ

- **技術検証（PoC）** を主目的とする
- 一般公開はしないが、**一般公開可能な品質** で設計・実装する
- セキュリティ、認証、コンテナ化など本番運用を見据えた構成とする

---

## 2. 機能要件

### 2.1 対象GitHubイベント（初期スコープ）

最小限のイベントから開始し、段階的に拡張する方針。

| イベント | アクション | 説明 |
|----------|-----------|------|
| Issues | `opened` | Issueが新規作成された |
| Pull Request | `closed` (merged=true) | PRがマージされた |

> **拡張予定**: Issue closed/edited、PR opened/closed、PR Review、Push、Comment 等は後続フェーズで追加

### 2.2 Webhook受信

- GitHub Webhookからのイベントペイロードを受信するエンドポイントを提供する
- **Webhook Secret による署名検証**（`X-Hub-Signature-256`）を必須とする
- 受信したイベントはMySQLに永続化する
- 受信と同時にSSE経由で接続中のクライアントへリアルタイム配信する

### 2.3 ダッシュボード（Web UI）

- イベントログの一覧表示（新しい順）
- リアルタイム通知（SSEによるサーバープッシュで画面を自動更新）
- イベント種別によるフィルタリング
- 各イベントの詳細表示（イベントタイプ、リポジトリ名、アクター、タイムスタンプ等）

### 2.4 認証・アクセス制御

- **GitHub OAuth** によるログイン認証を実装する
- 認証済みユーザーのみダッシュボードにアクセス可能とする
- 将来の一般公開時にそのまま利用できる認証基盤とする

---

## 3. 技術スタック

| レイヤー | 技術 | 備考 |
|----------|------|------|
| **フロントエンド** | Nuxt 3 (Vue 3 + TypeScript) | SSR/CSR対応、ダッシュボードUI |
| **バックエンド** | Go (1.21+) | Webhookハンドラ、REST API、SSEエンドポイント |
| **データベース** | MySQL 8.0 | イベントログの永続化 |
| **リアルタイム通信** | SSE (Server-Sent Events) | Go → ブラウザへの一方向プッシュ |
| **認証** | GitHub OAuth 2.0 | ダッシュボードへのアクセス制御 |
| **コンテナ** | Docker + docker-compose | ローカル開発環境の統一 |

---

## 4. アーキテクチャ概要

```
┌─────────────┐     Webhook      ┌─────────────────┐
│   GitHub     │ ───────────────→ │  Go Backend     │
│  (Webhook)   │   POST /webhook  │  (API Server)   │
└─────────────┘                   │                 │
                                  │  ┌───────────┐  │
                                  │  │ Webhook   │  │
                                  │  │ Handler   │──┼──→ MySQL (永続化)
                                  │  └─────┬─────┘  │
                                  │        │        │
                                  │  ┌─────▼─────┐  │
                                  │  │ SSE       │  │
                                  │  │ Broadcast │  │
                                  │  └─────┬─────┘  │
                                  │        │        │
                                  └────────┼────────┘
                                           │ SSE Stream
                                  ┌────────▼────────┐
                                  │  Nuxt 3         │
                                  │  (Dashboard)    │
                                  │  - イベント一覧  │
                                  │  - リアルタイム   │
                                  │  - フィルタ      │
                                  └─────────────────┘
```

### 4.1 データフロー

1. GitHubがWebhookイベントを `POST /api/webhook` へ送信
2. Goバックエンドが署名検証後、ペイロードをパースしMySQLへ保存
3. 保存と同時にSSEチャネルへイベントをブロードキャスト
4. Nuxtフロントエンドが SSE 接続でイベントを受信し、UIをリアルタイム更新

---

## 5. リポジトリ構成（モノレポ）

```
github-events-dashboard-poc/
├── docker-compose.yml          # 全サービスの起動定義
├── frontend/                   # Nuxt 3 アプリケーション
│   ├── Dockerfile
│   ├── nuxt.config.ts
│   ├── pages/
│   ├── components/
│   ├── composables/
│   └── ...
├── backend/                    # Go API サーバー
│   ├── Dockerfile
│   ├── go.mod
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── handler/            # HTTPハンドラ (webhook, api, sse)
│   │   ├── model/              # データモデル
│   │   ├── repository/         # DB操作
│   │   ├── service/            # ビジネスロジック
│   │   └── auth/               # GitHub OAuth
│   └── ...
├── db/                         # DB関連
│   └── migrations/             # マイグレーションSQL
├── docs/                       # ドキュメント
│   └── first-requirement.md
└── specs/                      # spec-kit 仕様書
```

---

## 6. Docker構成

**モノレポ + docker-compose** 構成。`docker-compose up` で全サービスが起動する。

### 6.1 サービス一覧

| サービス名 | イメージ | ポート (ホスト:コンテナ) | 説明 |
|-----------|---------|------------------------|------|
| `frontend` | Nuxt 3 (Node 20) | `3000:3000` | ダッシュボードUI |
| `backend` | Go 1.21+ | `8080:8080` | API / Webhook / SSE |
| `db` | MySQL 8.0 | `3306:3306` | イベントデータ永続化 |

### 6.2 要件

- `docker-compose up` 一発で全サービスが起動すること
- 各サービスは独立したDockerfileを持つこと
- ホットリロード対応（開発時）
- 環境変数は `.env` ファイルで管理（`.env.example` をリポジトリに含める）

---

## 7. 環境変数（想定）

```env
# GitHub OAuth
GITHUB_CLIENT_ID=xxx
GITHUB_CLIENT_SECRET=xxx

# GitHub Webhook
GITHUB_WEBHOOK_SECRET=xxx

# MySQL
MYSQL_HOST=db
MYSQL_PORT=3306
MYSQL_USER=dashboard
MYSQL_PASSWORD=xxx
MYSQL_DATABASE=github_events

# Backend
BACKEND_PORT=8080
FRONTEND_URL=http://localhost:3000

# Frontend
NUXT_PUBLIC_API_BASE=http://localhost:8080
```

---

## 8. 非機能要件

| 項目 | 要件 |
|------|------|
| **開発環境** | Docker + docker-compose でローカル完結。OS非依存 |
| **コード品質** | Go: golangci-lint、TypeScript: ESLint + Prettier |
| **テスト** | 単体テスト必須（Go: testing、Nuxt: Vitest） |
| **セキュリティ** | Webhook署名検証、GitHub OAuth、CORS設定、環境変数による秘匿情報管理 |
| **ログ** | 構造化ログ（JSON形式）をGoバックエンドで出力 |
| **エラーハンドリング** | Webhook受信失敗時のリトライは行わない（GitHub側が再送する前提） |

---

## 9. スコープ外（初期フェーズ）

以下は初期スコープに含めない。後続フェーズで検討する。

- Issue closed/edited、PR opened/closed/review 等の追加イベント
- 複数リポジトリの管理UI
- イベントの検索・集計・ダッシュボードウィジェット
- メール/Slack等の外部通知連携
- CI/CDパイプライン構築
- 本番環境へのデプロイ（クラウドインフラ）
- パフォーマンスチューニング・負荷テスト

---

## 10. 開発の進め方

1. **本ドキュメント確定** → spec-kit `/speckit.specify` で正式仕様書を生成
2. `/speckit.plan` で技術設計・実装計画を策定
3. `/speckit.tasks` でタスク分解
4. `/speckit.implement` で実装開始

---

## 11. 用語集

| 用語 | 説明 |
|------|------|
| **Webhook** | GitHubが外部URLへHTTP POSTでイベントを通知する仕組み |
| **SSE** | Server-Sent Events。サーバーからクライアントへの一方向リアルタイム通信 |
| **GitHub OAuth** | GitHubアカウントを利用したOAuth 2.0認証フロー |
| **PoC** | Proof of Concept（技術検証） |
