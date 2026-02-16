# Data Model: GitHub Events Dashboard

**Date**: 2026-02-15
**Feature**: 001-github-events-dashboard
**Source**: [spec.md](./spec.md) Key Entities section

---

## Entity Relationship Diagram

```
┌─────────────────────────────┐
│         events              │
├─────────────────────────────┤
│ id           BIGINT PK AUTO │
│ delivery_id  VARCHAR(36) UQ │
│ event_type   VARCHAR(50)    │
│ action       VARCHAR(50)    │
│ repo_name    VARCHAR(255)   │
│ sender_login VARCHAR(255)   │
│ sender_avatar_url TEXT      │
│ title        VARCHAR(500)   │
│ body         TEXT            │
│ html_url     TEXT            │
│ event_data   JSON           │
│ occurred_at  DATETIME       │
│ received_at  DATETIME       │
│ created_at   DATETIME       │
└─────────────────────────────┘

┌─────────────────────────────┐
│         users               │
├─────────────────────────────┤
│ id           BIGINT PK AUTO │
│ github_id    BIGINT UQ      │
│ login        VARCHAR(255) UQ│
│ display_name VARCHAR(255)   │
│ avatar_url   TEXT            │
│ access_token TEXT            │
│ last_login   DATETIME       │
│ created_at   DATETIME       │
│ updated_at   DATETIME       │
└─────────────────────────────┘
```

**Relationship**: events と users の間に直接的な外部キー関係はない。events.sender_login はGitHubイベントの送信者であり、必ずしもダッシュボードユーザーとは限らない。

---

## Entity: events

GitHubから受信したWebhookイベントを記録するテーブル。

### Fields

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | BIGINT | PK, AUTO_INCREMENT | イベントの一意識別子 |
| `delivery_id` | VARCHAR(36) | UNIQUE, NOT NULL | GitHub delivery ID（冪等性の保証に使用） |
| `event_type` | VARCHAR(50) | NOT NULL, INDEX | イベント種別（`issues`, `pull_request`） |
| `action` | VARCHAR(50) | NOT NULL | アクション（`opened`, `closed`） |
| `repo_name` | VARCHAR(255) | NOT NULL, INDEX | リポジトリのフルネーム（`owner/repo`） |
| `sender_login` | VARCHAR(255) | NOT NULL | イベント実行者のGitHubユーザー名 |
| `sender_avatar_url` | TEXT | NULL | イベント実行者のアバター画像URL |
| `title` | VARCHAR(500) | NULL | Issue/PRのタイトル |
| `body` | TEXT | NULL | Issue/PRの本文（先頭500文字まで保存） |
| `html_url` | TEXT | NOT NULL | GitHubの該当ページURL |
| `event_data` | JSON | NULL | イベント固有の追加データ（生ペイロードのサブセット） |
| `occurred_at` | DATETIME | NOT NULL | イベントがGitHub上で発生した日時 |
| `received_at` | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP | システムがイベントを受信した日時 |
| `created_at` | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP | レコード作成日時 |

### Indexes

| Index Name | Columns | Type | Purpose |
|------------|---------|------|---------|
| `PRIMARY` | `id` | PRIMARY | 主キー |
| `uq_delivery_id` | `delivery_id` | UNIQUE | 冪等性保証（重複排除） |
| `idx_event_type` | `event_type` | INDEX | イベント種別フィルタリング |
| `idx_received_at` | `received_at` | INDEX | 新しい順ソート・ページネーション |
| `idx_repo_name` | `repo_name` | INDEX | 将来の複数リポジトリ対応用 |

### Validation Rules

- `delivery_id`: GitHubから受信した `X-GitHub-Delivery` ヘッダー値。UUID形式
- `event_type`: 初期スコープでは `issues` または `pull_request` のみ許可
- `action`: event_type に応じた有効なアクション値
  - `issues`: `opened`
  - `pull_request`: `closed`（merged=true の場合のみ）
- `html_url`: 有効なURL形式であること
- `occurred_at`: GitHubペイロードから抽出。未来の日時は許可しない

### State Transitions

events テーブルにはライフサイクル状態遷移はない。イベントは一度保存されたら不変（immutable）。削除・更新は行わない。

---

## Entity: users

GitHub OAuthで認証されたダッシュボードユーザーを記録するテーブル。

### Fields

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | BIGINT | PK, AUTO_INCREMENT | ユーザーの一意識別子 |
| `github_id` | BIGINT | UNIQUE, NOT NULL | GitHubユーザーID |
| `login` | VARCHAR(255) | UNIQUE, NOT NULL | GitHubユーザー名 |
| `display_name` | VARCHAR(255) | NULL | GitHubの表示名 |
| `avatar_url` | TEXT | NULL | GitHubプロフィール画像URL |
| `access_token` | TEXT | NOT NULL | GitHub OAuthアクセストークン（暗号化推奨） |
| `last_login` | DATETIME | NOT NULL | 最終ログイン日時 |
| `created_at` | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP | レコード作成日時 |
| `updated_at` | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | レコード更新日時 |

### Indexes

| Index Name | Columns | Type | Purpose |
|------------|---------|------|---------|
| `PRIMARY` | `id` | PRIMARY | 主キー |
| `uq_github_id` | `github_id` | UNIQUE | GitHubユーザーIDの一意性 |
| `uq_login` | `login` | UNIQUE | GitHubユーザー名の一意性 |

### Validation Rules

- `github_id`: GitHubから取得した数値ID。正の整数
- `login`: GitHubユーザー名。英数字とハイフンのみ
- `access_token`: OAuth認証で取得したトークン。本番環境では暗号化して保存すべき

### State Transitions

```
[新規] → OAuth認証成功 → [作成]
[作成] → 再ログイン → [更新] (access_token, last_login を更新)
```

---

## Migration SQL

### 001_create_events.sql

```sql
CREATE TABLE IF NOT EXISTS events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    delivery_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    sender_login VARCHAR(255) NOT NULL,
    sender_avatar_url TEXT,
    title VARCHAR(500),
    body TEXT,
    html_url TEXT NOT NULL,
    event_data JSON,
    occurred_at DATETIME NOT NULL,
    received_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_delivery_id (delivery_id),
    INDEX idx_event_type (event_type),
    INDEX idx_received_at (received_at),
    INDEX idx_repo_name (repo_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 002_create_users.sql

```sql
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    github_id BIGINT NOT NULL,
    login VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    avatar_url TEXT,
    access_token TEXT NOT NULL,
    last_login DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_github_id (github_id),
    UNIQUE KEY uq_login (login)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```
