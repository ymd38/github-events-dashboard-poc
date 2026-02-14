# Implementation Plan: GitHub Events Dashboard

**Branch**: `001-github-events-dashboard` | **Date**: 2026-02-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-github-events-dashboard/spec.md`

## Summary

GitHubリポジトリのイベント（Issue作成・PRマージ）をWebhookで受信し、MySQLに永続化した上で、SSE経由でNuxt 3ダッシュボードにリアルタイム配信するサービス。GoバックエンドがWebhookハンドラ・REST API・SSEエンドポイント・GitHub OAuth認証を提供し、docker-composeで全サービスをローカル起動する。

## Technical Context

**Language/Version**: Go 1.21+（バックエンド）、Node.js 20+ / TypeScript（フロントエンド）  
**Primary Dependencies**: Go標準ライブラリ + gorilla/mux or chi（ルーター）、Nuxt 3 (Vue 3)、TailwindCSS  
**Storage**: MySQL 8.0（イベント・ユーザーデータの永続化）  
**Testing**: Go: testing + testify、Nuxt: Vitest + Vue Test Utils  
**Target Platform**: Docker コンテナ（ローカル開発環境）  
**Project Type**: Web application（frontend + backend + database）  
**Performance Goals**: Webhook受信からダッシュボード表示まで5秒以内  
**Constraints**: ローカルDocker環境で完結、〜10名同時接続  
**Scale/Scope**: 単一リポジトリ、2イベント種別（Issue opened, PR merged）、〜10ユーザー

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution はプレースホルダー状態のため、明示的なゲート違反なし。以下の原則を本プロジェクトに適用する:

- **テスト必須**: Go は testing + testify、Nuxt は Vitest で単体テストを記述
- **シンプルさ優先**: PoC として最小限の構成。不要な抽象化を避ける
- **観測性**: 構造化ログ（JSON）+ ヘルスチェックエンドポイント
- **セキュリティ**: Webhook署名検証、GitHub OAuth、環境変数による秘匿情報管理

**Gate Result**: PASS（違反なし）

## Project Structure

### Documentation (this feature)

```text
specs/001-github-events-dashboard/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (OpenAPI)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
github-events-dashboard-poc/
├── docker-compose.yml
├── .env.example
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── auth/           # GitHub OAuth ハンドラ・ミドルウェア
│   │   ├── config/         # 環境変数読み込み・設定
│   │   ├── handler/        # HTTP ハンドラ (webhook, events, sse, health)
│   │   ├── middleware/     # ログ・認証・CORS ミドルウェア
│   │   ├── model/          # データモデル (Event, User)
│   │   ├── repository/     # DB 操作 (MySQL)
│   │   ├── service/        # ビジネスロジック
│   │   └── sse/            # SSE ブロードキャスト管理
│   └── tests/
│       ├── integration/
│       └── unit/
├── frontend/
│   ├── Dockerfile
│   ├── nuxt.config.ts
│   ├── package.json
│   ├── app.vue
│   ├── pages/
│   │   ├── index.vue       # ダッシュボード（イベント一覧）
│   │   └── login.vue       # ログイン画面
│   ├── components/
│   │   ├── EventList.vue
│   │   ├── EventDetail.vue
│   │   ├── EventFilter.vue
│   │   ├── Pagination.vue
│   │   └── AppHeader.vue
│   ├── composables/
│   │   ├── useEvents.ts    # イベントAPI呼び出し
│   │   ├── useSSE.ts       # SSE接続管理・自動再接続
│   │   └── useAuth.ts      # 認証状態管理
│   ├── server/
│   │   └── api/            # Nuxt server routes (proxy)
│   └── tests/
│       └── unit/
├── db/
│   └── migrations/
│       ├── 001_create_events.sql
│       └── 002_create_users.sql
├── docs/
│   └── first-requirement.md
└── specs/
```

**Structure Decision**: Web application 構成（frontend + backend）を採用。モノレポでdocker-compose管理。Go バックエンドは `cmd/server` + `internal/` のクリーンアーキテクチャ風レイアウト。Nuxt 3 フロントエンドは pages/components/composables の標準構成。

## Complexity Tracking

> Constitution にゲート違反なし。追加の正当化は不要。
