# Research: GitHub Events Dashboard

**Date**: 2026-02-15
**Feature**: 001-github-events-dashboard
**Status**: Complete (no NEEDS CLARIFICATION remaining)

---

## R-001: Go HTTP ルーターの選定

**Decision**: chi (go-chi/chi v5)

**Rationale**:
- Go標準ライブラリの `net/http` と完全互換（`http.Handler` インターフェース準拠）
- ミドルウェアチェーンが標準的で、認証・ログ・CORSの組み込みが容易
- gorilla/mux はアーカイブ済み（2022年末にメンテナンス停止）のため除外
- 軽量で依存が少なく、PoCに適したサイズ感

**Alternatives considered**:
- **gorilla/mux**: 広く使われていたが、2022年にアーカイブ済み。新規プロジェクトには非推奨
- **gin**: 高機能だが独自のContext型を使用し、標準ライブラリとの互換性が低い
- **net/http (標準のみ)**: Go 1.22+ の新しいルーティングパターンで十分だが、ミドルウェア管理が煩雑

---

## R-002: MySQL ドライバーとマイグレーション

**Decision**: go-sql-driver/mysql + golang-migrate/migrate

**Rationale**:
- `go-sql-driver/mysql` はGo標準の `database/sql` インターフェースに準拠し、最も広く使われるMySQLドライバー
- `golang-migrate/migrate` はSQLファイルベースのマイグレーションをサポートし、docker-compose起動時に自動適用可能
- ORMは使用しない（PoCの規模では `database/sql` + 手書きSQLで十分。不要な抽象化を避ける）

**Alternatives considered**:
- **GORM**: フル機能ORM。PoCの2テーブルには過剰で、SQLの学習・デバッグが困難になる
- **sqlx**: `database/sql` の拡張。構造体マッピングが便利だが、PoCでは標準で十分
- **goose**: マイグレーションツール。golang-migrateと同等だが、コミュニティ規模で劣る

---

## R-003: SSE (Server-Sent Events) 実装パターン

**Decision**: Go標準ライブラリで自前実装（ブロードキャストパターン）

**Rationale**:
- SSEはHTTPレスポンスのストリーミングであり、Go標準の `http.Flusher` インターフェースで実装可能
- クライアント管理（接続/切断）とブロードキャストのハブパターンを自前実装
- 外部ライブラリへの依存を最小化し、PoCとして仕組みを理解しやすくする
- 同時接続数は〜10名のため、パフォーマンス上の懸念なし

**Implementation pattern**:
```
SSEHub:
  - clients map[chan Event]bool  // 接続中クライアントのチャネルマップ
  - register chan (chan Event)    // 新規接続登録
  - unregister chan (chan Event)  // 切断登録
  - broadcast chan Event          // 全クライアントへの配信
  - Run() goroutine              // イベントループ
```

**Alternatives considered**:
- **r3labs/sse**: Go用SSEライブラリ。便利だが、PoCでは学習目的もあり自前実装が望ましい
- **WebSocket (gorilla/websocket)**: 双方向通信が可能だが、本ユースケースはサーバー→クライアントの一方向のみ。SSEの方がシンプル

---

## R-004: GitHub OAuth 2.0 実装

**Decision**: Go標準の `golang.org/x/oauth2` パッケージ + セッション管理は `gorilla/sessions`

**Rationale**:
- `golang.org/x/oauth2` はGoの準標準ライブラリで、GitHub OAuthプロバイダーが組み込み済み
- `gorilla/sessions` はCookieベースのセッション管理で広く使われ、gorilla/muxのアーカイブとは独立して維持されている
- セッション有効期限24時間はCookieのMaxAgeで制御

**OAuth flow**:
1. フロントエンド → `/api/auth/login` → GitHub認可画面へリダイレクト
2. GitHub → `/api/auth/callback` → アクセストークン取得 → ユーザー情報取得 → セッション作成
3. セッションCookieでログイン状態を維持
4. `/api/auth/logout` → セッション破棄

**Alternatives considered**:
- **自前OAuth実装**: HTTPリクエストを直接組み立てる方法。エラーハンドリングが煩雑
- **Auth0/Firebase Auth**: 外部認証サービス。PoCには過剰で、GitHub OAuthに特化する必要がある

---

## R-005: GitHub Webhook 署名検証

**Decision**: Go標準ライブラリの `crypto/hmac` + `crypto/sha256` で自前実装

**Rationale**:
- GitHub Webhookの署名検証は `X-Hub-Signature-256` ヘッダーのHMAC-SHA256検証
- Go標準ライブラリのみで実装可能（外部依存不要）
- リクエストボディ全体をHMAC-SHA256でハッシュし、ヘッダー値と定数時間比較

**Verification flow**:
1. リクエストボディを読み取り
2. Webhook Secretをキーとして HMAC-SHA256 を計算
3. `X-Hub-Signature-256` ヘッダーの値と `hmac.Equal()` で比較
4. 一致しない場合は 401 Unauthorized を返却

---

## R-006: Nuxt 3 フロントエンド構成

**Decision**: Nuxt 3 (SSR mode) + TailwindCSS + Pinia（状態管理）

**Rationale**:
- Nuxt 3はVue 3ベースのフルスタックフレームワークで、SSR/CSR両対応
- TailwindCSSはユーティリティファーストCSSで、ダッシュボードUIの迅速な構築に適する
- Piniaは Vue 3の公式状態管理ライブラリで、イベントデータやフィルタ状態の管理に使用
- SSE接続はcomposable (`useSSE`) として実装し、EventSource APIを使用

**Alternatives considered**:
- **Vuetify**: コンポーネントライブラリ。PoCには過剰で、TailwindCSSの方が柔軟
- **Vuex**: 旧状態管理。Piniaが後継として推奨
- **Next.js (React)**: ユーザーがNuxtを選択済み

---

## R-007: Docker構成とホットリロード

**Decision**: docker-compose with volume mounts + Air (Go) + Nuxt dev server

**Rationale**:
- Go: `cosmtrek/air` でファイル変更検知・自動リビルド
- Nuxt: `nuxt dev` の組み込みHMR（Hot Module Replacement）
- MySQL: 公式イメージ + init scriptでマイグレーション自動適用
- ボリュームマウントでホスト側のソースコード変更をコンテナに即反映

**Alternatives considered**:
- **手動リビルド**: 開発効率が著しく低下
- **Tilt/Skaffold**: Kubernetes向けツール。docker-composeのPoCには過剰

---

## R-008: Webhook冪等性の実装

**Decision**: GitHub delivery ID (`X-GitHub-Delivery` ヘッダー) をユニークキーとしてDBに保存

**Rationale**:
- GitHubは各Webhook配信に一意のdelivery IDを付与する
- eventsテーブルの `delivery_id` カラムにUNIQUE制約を設定
- INSERT時に重複キーエラーが発生した場合は、既に処理済みとして200 OKを返却（冪等）
- MySQL の `INSERT IGNORE` または `ON DUPLICATE KEY` で実装

**Alternatives considered**:
- **アプリケーション層での重複チェック**: SELECT → INSERT のレースコンディションリスク
- **Redis等のキャッシュ**: PoCには過剰な依存追加
