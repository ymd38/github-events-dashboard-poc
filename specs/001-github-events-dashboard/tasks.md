# Tasks: GitHub Events Dashboard

**Input**: Design documents from `/specs/001-github-events-dashboard/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/openapi.yaml, quickstart.md

**Tests**: テスト関連タスクは含まない（仕様書に明示的なTDD要求なし）。Constitution Checkの「テスト必須」原則に基づき、各Phase完了時にテスト追加を推奨する。

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: プロジェクト初期化と基本構造の作成

- [ ] T001 Create project directory structure per plan.md (backend/, frontend/, db/, docker-compose.yml, .env.example)
- [ ] T002 Initialize Go module in backend/go.mod with module name and Go 1.21+ version
- [ ] T003 [P] Initialize Nuxt 3 project in frontend/ with package.json, nuxt.config.ts, tsconfig.json
- [ ] T004 [P] Create .env.example with all required environment variables (GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, GITHUB_WEBHOOK_SECRET, MYSQL_*, BACKEND_PORT, SESSION_SECRET, NUXT_PUBLIC_API_BASE)

**Checkpoint**: プロジェクトの骨格が完成。各ディレクトリとモジュール初期化が完了。

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 全User Storyの前提となるコアインフラ。このフェーズが完了するまでUser Story作業は開始不可。

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T005 Create backend configuration loader in backend/internal/config/config.go (環境変数読み込み、全設定項目の構造体定義)
- [ ] T006 [P] Create Event model in backend/internal/model/event.go (data-model.mdのeventsテーブルに対応する構造体)
- [ ] T007 [P] Create User model in backend/internal/model/user.go (data-model.mdのusersテーブルに対応する構造体)
- [ ] T008 Create database connection manager in backend/internal/repository/db.go (MySQL接続プール初期化、Ping確認)
- [ ] T009 Create SQL migration files in db/migrations/001_create_events.sql and db/migrations/002_create_users.sql (data-model.mdのCREATE TABLE文)
- [ ] T010 [P] Create structured JSON logger in backend/internal/middleware/logger.go (FR-023: リクエスト・エラー・イベント受信の構造化ログ)
- [ ] T011 [P] Create CORS middleware in backend/internal/middleware/cors.go (フロントエンドからのクロスオリジンリクエスト許可)
- [ ] T012 Setup chi router and server bootstrap in backend/cmd/server/main.go (ルーター初期化、ミドルウェア登録、グレースフルシャットダウン)
- [ ] T013 Create health check handler in backend/internal/handler/health.go (FR-024: GET /api/health、DB接続確認、HealthResponse返却)
- [ ] T014 Create Docker configuration: backend/Dockerfile (Go multi-stage build with Air for hot reload)
- [ ] T015 [P] Create Docker configuration: frontend/Dockerfile (Node.js with nuxt dev)
- [ ] T016 [P] Create Docker configuration: docker-compose.yml (backend, frontend, db services with volume mounts, migration auto-apply)
- [ ] T017 Verify docker-compose up starts all services and health check returns 200 OK

**Checkpoint**: Foundation ready - docker-compose up で全サービスが起動し、GET /api/health が正常応答を返す状態。

---

## Phase 3: User Story 1 - Webhookイベント受信と永続化 (Priority: P1) 🎯 MVP

**Goal**: GitHubからのWebhookイベントを受信し、署名検証後にMySQLに永続化する。重複排除（冪等性）を保証する。

**Independent Test**: curlでWebhookテストペイロード（正当な署名付き）を送信し、DBにイベントが保存されることを確認。不正署名は401で拒否されること。

**Covers**: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-019, FR-023 | SC-005, SC-009

### Implementation for User Story 1

- [ ] T018 [US1] Implement webhook signature verification in backend/internal/handler/webhook.go (FR-002: X-Hub-Signature-256のHMAC-SHA256検証、crypto/hmac + crypto/sha256使用)
- [ ] T019 [US1] Implement event repository in backend/internal/repository/event_repository.go (INSERT with delivery_id UNIQUE制約、重複時はduplicate応答、ListEvents/GetEventByID)
- [ ] T020 [US1] Implement event service in backend/internal/service/event_service.go (ペイロードパース、issues/opened と pull_request/closed+merged=true の振り分け、EventRepository呼び出し)
- [ ] T021 [US1] Implement webhook handler in backend/internal/handler/webhook.go (POST /api/webhook: 署名検証→ペイロードパース→サービス呼び出し→WebhookResponse返却、構造化ログ出力)
- [ ] T022 [US1] Register webhook route in backend/cmd/server/main.go (POST /api/webhook をルーターに登録)
- [ ] T023 [US1] Add idempotency handling in backend/internal/repository/event_repository.go (FR-006: INSERT IGNOREまたはON DUPLICATE KEYでdelivery_id重複時に既存レコードを返す)

**Checkpoint**: POST /api/webhook にcurlでテストペイロードを送信し、正当署名→DB保存、不正署名→401拒否、重複delivery_id→200 duplicate応答を確認。

---

## Phase 4: User Story 2 - ダッシュボードでイベント一覧を閲覧する (Priority: P2)

**Goal**: 認証済みユーザーがダッシュボードでイベント一覧をページネーション付きで閲覧し、フィルタリング・詳細表示ができる。

**Independent Test**: DBにテストデータを投入し、GET /api/events でページネーション付き一覧が返ること。フロントエンドでイベント一覧・詳細・フィルタが動作すること。

**Covers**: FR-007, FR-008, FR-009, FR-010, FR-020 | SC-004

### Implementation for User Story 2

- [ ] T024 [US2] Implement events list API handler in backend/internal/handler/events.go (GET /api/events: page, per_page, event_type クエリパラメータ対応、EventListResponse + Pagination返却)
- [ ] T025 [US2] Implement event detail API handler in backend/internal/handler/events.go (GET /api/events/:id: 単一イベント詳細返却、404ハンドリング)
- [ ] T026 [US2] Register events routes in backend/cmd/server/main.go (GET /api/events, GET /api/events/{id} をルーターに登録)
- [ ] T027 [P] [US2] Install TailwindCSS and Pinia in frontend/ (nuxt.config.ts にモジュール追加、tailwind.config.ts作成)
- [ ] T028 [P] [US2] Create useEvents composable in frontend/composables/useEvents.ts (イベント一覧取得API呼び出し、ページネーション状態管理、フィルタ状態管理)
- [ ] T029 [US2] Create AppHeader component in frontend/components/AppHeader.vue (ヘッダーUI、ユーザー情報表示枠、ログアウトボタン枠)
- [ ] T030 [US2] Create EventList component in frontend/components/EventList.vue (FR-008: イベント種別、リポジトリ名、操作者、発生日時の一覧表示)
- [ ] T031 [P] [US2] Create EventFilter component in frontend/components/EventFilter.vue (FR-009: イベント種別フィルタリングUI)
- [ ] T032 [P] [US2] Create Pagination component in frontend/components/Pagination.vue (FR-020: ページ番号切り替えUI)
- [ ] T033 [US2] Create EventDetail component in frontend/components/EventDetail.vue (FR-010: イベント固有データ、GitHub URLリンク表示)
- [ ] T034 [US2] Create dashboard page in frontend/pages/index.vue (EventList, EventFilter, Pagination, EventDetail を統合、空状態メッセージ対応)
- [ ] T035 [US2] Create app.vue layout in frontend/app.vue (AppHeader + NuxtPage のレイアウト)

**Checkpoint**: docker-compose up 後、http://localhost:3000 でイベント一覧が表示される（認証はまだスキップ）。フィルタ・ページネーション・詳細表示が動作。

---

## Phase 5: User Story 3 - リアルタイムでイベント通知を受け取る (Priority: P3)

**Goal**: ダッシュボードを開いているユーザーに、SSE経由で新着イベントをリアルタイム配信する。自動再接続と切断中イベントの再取得を実装する。

**Independent Test**: ダッシュボードを開いた状態でWebhookテストペイロードを送信し、5秒以内にページリロードなしで新着イベントが表示されることを確認。

**Covers**: FR-011, FR-012, FR-013, FR-018, FR-025 | SC-001, SC-002, SC-003

### Implementation for User Story 3

- [ ] T036 [US3] Implement SSE hub in backend/internal/sse/hub.go (クライアント管理: register/unregister channels、broadcastチャネル、Run goroutine)
- [ ] T037 [US3] Implement SSE handler in backend/internal/handler/sse.go (GET /api/events/stream: http.Flusher使用、text/event-stream応答、クライアント登録/切断処理)
- [ ] T038 [US3] Integrate SSE broadcast into webhook handler in backend/internal/handler/webhook.go (イベント保存成功後にSSE hubへbroadcast)
- [ ] T039 [US3] Register SSE route in backend/cmd/server/main.go (GET /api/events/stream をルーターに登録)
- [ ] T040 [US3] Create useSSE composable in frontend/composables/useSSE.ts (EventSource API使用、自動再接続ロジック、再接続時にAPI再取得 FR-025)
- [ ] T041 [US3] Integrate SSE into dashboard page in frontend/pages/index.vue (useSSE composable接続、新着イベントを一覧先頭に追加、フィルタ適用中の新着イベント処理)
- [ ] T042 [US3] Add new event highlight styling in frontend/components/EventList.vue (FR-013: 新着イベントの視覚的ハイライト、数秒後にフェードアウト)

**Checkpoint**: ダッシュボードを開いた状態でcurlでWebhookペイロードを送信し、5秒以内にリアルタイムで新着イベントがハイライト表示される。接続切断→再接続時に未受信イベントも表示される。

---

## Phase 6: User Story 4 - GitHub OAuthでログインする (Priority: P4)

**Goal**: GitHub OAuthによるログイン認証を実装し、未認証ユーザーのアクセスを制限する。セッション有効期限（24時間）と期限切れ時のSSE切断を実装する。

**Independent Test**: GitHub OAuthフローでログインし、認証済み状態でダッシュボードにアクセスできること。未認証状態ではログイン画面にリダイレクトされること。

**Covers**: FR-014, FR-015, FR-016, FR-017, FR-021, FR-022 | SC-006, SC-007

### Implementation for User Story 4

- [ ] T043 [US4] Implement user repository in backend/internal/repository/user_repository.go (FindByGitHubID、CreateOrUpdate、セッション用のユーザー取得)
- [ ] T044 [US4] Implement OAuth handler in backend/internal/auth/oauth.go (GET /api/auth/login: GitHub認可URLへリダイレクト、state生成)
- [ ] T045 [US4] Implement OAuth callback handler in backend/internal/auth/oauth.go (GET /api/auth/callback: アクセストークン取得→ユーザー情報取得→DB保存→セッション作成→フロントエンドへリダイレクト)
- [ ] T046 [US4] Implement session management with gorilla/sessions in backend/internal/auth/session.go (セッション作成・検証・破棄、MaxAge 24時間、Cookie設定)
- [ ] T047 [US4] Implement auth middleware in backend/internal/middleware/auth.go (FR-015: セッション検証、未認証時401応答、FR-017: 期限切れ検出)
- [ ] T048 [US4] Implement logout handler in backend/internal/auth/oauth.go (POST /api/auth/logout: セッション破棄)
- [ ] T049 [US4] Implement current user handler in backend/internal/auth/oauth.go (GET /api/auth/me: セッションからユーザー情報返却)
- [ ] T050 [US4] Register auth routes and apply auth middleware in backend/cmd/server/main.go (/api/auth/* ルート登録、/api/events* と /api/events/stream に認証ミドルウェア適用)
- [ ] T051 [US4] Add session expiry check to SSE handler in backend/internal/handler/sse.go (FR-021: セッション期限切れ時にsession_expiredイベント送信→接続切断)
- [ ] T052 [P] [US4] Create useAuth composable in frontend/composables/useAuth.ts (ログイン状態管理、/api/auth/me呼び出し、ログアウト処理)
- [ ] T053 [US4] Create login page in frontend/pages/login.vue (GitHubログインボタン、FR-022: エラーメッセージ表示・リトライボタン)
- [ ] T054 [US4] Add auth guard middleware in frontend/middleware/auth.global.ts (未認証時にloginページへリダイレクト)
- [ ] T055 [US4] Integrate auth into AppHeader in frontend/components/AppHeader.vue (ユーザー名・アバター表示、ログアウトボタン)
- [ ] T056 [US4] Handle session_expired SSE event in frontend/composables/useSSE.ts (FR-021: session_expiredイベント受信時にSSE切断→ログイン画面へリダイレクト)

**Checkpoint**: GitHub OAuthでログイン→ダッシュボード表示→ログアウト→ログイン画面リダイレクトの全フローが動作。未認証アクセスは100%ブロック。

---

## Phase 7: User Story 5 - コンテナ化された開発環境で起動する (Priority: P5)

**Goal**: docker-compose up 一発で全サービスが起動し、全機能が動作する開発環境を完成させる。

**Independent Test**: リポジトリをクローンし、.env設定後にdocker-compose upで全サービスが起動し、ヘルスチェックが通ること。

**Covers**: SC-008, SC-010

### Implementation for User Story 5

- [ ] T057 [US5] Finalize docker-compose.yml with all service dependencies, health checks, and restart policies
- [ ] T058 [US5] Add database migration auto-apply to docker-compose.yml (MySQL initdb.d or wait-for-it + migrate)
- [ ] T059 [US5] Create README.md with setup instructions (quickstart.mdの内容をベースに、GitHub OAuth App作成手順、ngrok設定手順を含む)
- [ ] T060 [US5] End-to-end validation: docker-compose down -v → docker-compose up → health check → OAuth login → Webhook送信 → リアルタイム表示の全フロー確認

**Checkpoint**: 新規クローンから docker-compose up のみで全機能が動作する状態。

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: 全User Storyにまたがる改善と品質向上

- [ ] T061 [P] Add structured logging to all handlers in backend/internal/handler/*.go (FR-023: 統一的なJSON形式ログ出力の確認・補完)
- [ ] T062 [P] Add error handling improvements across backend/ (統一的なErrorResponse形式、パニックリカバリミドルウェア)
- [ ] T063 [P] Add empty state UI in frontend/components/EventList.vue (Edge Case: イベント0件時の適切なメッセージ表示)
- [ ] T064 [P] Add loading states to frontend pages (ローディングスピナー、スケルトンUI)
- [ ] T065 Run quickstart.md validation (手順通りに環境構築→全機能動作を確認)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Foundational (Phase 2) completion
- **US2 (Phase 4)**: Depends on Foundational (Phase 2) completion. US1のevent_repositoryを共有するが独立テスト可能
- **US3 (Phase 5)**: Depends on US1 (Phase 3) completion (Webhook受信→SSEブロードキャストの統合が必要)
- **US4 (Phase 6)**: Depends on Foundational (Phase 2) completion. 他のUS完了後に認証ミドルウェアを統合
- **US5 (Phase 7)**: Depends on US1-US4 completion (全機能統合の検証)
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

```
Phase 1: Setup
    ↓
Phase 2: Foundational
    ↓
    ├── Phase 3: US1 (Webhook受信・永続化) ← MVP
    │       ↓
    │   Phase 5: US3 (リアルタイム配信) ← US1のWebhook処理に依存
    │
    ├── Phase 4: US2 (ダッシュボード表示) ← US1と並行可能
    │
    └── Phase 6: US4 (GitHub OAuth認証) ← US1/US2と並行可能
            ↓
        Phase 7: US5 (開発環境統合)
            ↓
        Phase 8: Polish
```

### Within Each User Story

- Models before services
- Services before handlers
- Backend before frontend (API first)
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

**Phase 1 内**:
- T003 (Nuxt初期化) と T004 (.env.example) は並行可能

**Phase 2 内**:
- T006 (Event model) と T007 (User model) は並行可能
- T010 (Logger) と T011 (CORS) は並行可能
- T014 (backend Dockerfile) と T015 (frontend Dockerfile) と T016 (docker-compose) は並行可能

**Phase 4 (US2) 内**:
- T027 (TailwindCSS/Pinia) と T028 (useEvents) は並行可能
- T031 (EventFilter) と T032 (Pagination) は並行可能

**Phase 6 (US4) 内**:
- T052 (useAuth) はバックエンドの認証実装と並行可能

**Cross-Story並行**:
- US1 (Phase 3) と US2 (Phase 4) は Phase 2 完了後に並行開始可能
- US4 (Phase 6) は Phase 2 完了後にUS1/US2と並行開始可能

---

## Parallel Example: User Story 1

```bash
# Backend implementation (sequential within story):
Task T018: "Webhook signature verification in backend/internal/handler/webhook.go"
Task T019: "Event repository in backend/internal/repository/event_repository.go"
Task T020: "Event service in backend/internal/service/event_service.go"
Task T021: "Webhook handler completion in backend/internal/handler/webhook.go"
Task T022: "Register webhook route in backend/cmd/server/main.go"
Task T023: "Idempotency handling in backend/internal/repository/event_repository.go"
```

## Parallel Example: User Story 2

```bash
# Backend (sequential):
Task T024: "Events list API in backend/internal/handler/events.go"
Task T025: "Event detail API in backend/internal/handler/events.go"
Task T026: "Register events routes in backend/cmd/server/main.go"

# Frontend (parallel after T027):
Task T027: "TailwindCSS + Pinia setup"
Task T028: "useEvents composable" (parallel with T027)
Task T031: "EventFilter component" (parallel with T032)
Task T032: "Pagination component" (parallel with T031)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Webhook受信・永続化)
4. **STOP and VALIDATE**: curlでWebhookテストペイロードを送信し、DB保存・署名検証・冪等性を確認
5. MVP完了 - イベントログとしての最小限の価値を提供

### Incremental Delivery

1. Setup + Foundational → Foundation ready (docker-compose up で起動)
2. Add US1 → Webhook受信・永続化が動作 → **MVP!**
3. Add US2 → ダッシュボードでイベント閲覧可能 → Demo可能
4. Add US3 → リアルタイム更新が動作 → ユーザー体験向上
5. Add US4 → 認証が動作 → セキュリティ確保
6. Add US5 → 開発環境完成 → 他の開発者がすぐに参加可能
7. Polish → 品質向上

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Webhook) → User Story 3 (SSE)
   - Developer B: User Story 2 (Dashboard UI)
   - Developer C: User Story 4 (OAuth)
3. All converge for User Story 5 (Integration) and Polish

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- 全タスクにファイルパスを明記し、LLMが追加コンテキストなしで実行可能な粒度で記述
