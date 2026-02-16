.PHONY: up down build restart logs logs-backend logs-frontend logs-db ps clean db-shell backend-shell frontend-shell health ngrok

# === Docker Compose ===

## 全サービスをバックグラウンドで起動
up:
	docker compose up -d

## 全サービスをビルドして起動
build:
	docker compose up --build -d

## 全サービスを停止
down:
	docker compose down

## 全サービスを停止しボリュームも削除（DB初期化）
clean:
	docker compose down -v

## 全サービスを再起動
restart:
	docker compose restart

## 全サービスのログを表示（フォロー）
logs:
	docker compose logs -f

## バックエンドのログを表示
logs-backend:
	docker compose logs -f backend

## フロントエンドのログを表示
logs-frontend:
	docker compose logs -f frontend

## DBのログを表示
logs-db:
	docker compose logs -f db

## サービス一覧を表示
ps:
	docker compose ps

# === Shell ===

## バックエンドコンテナにシェル接続
backend-shell:
	docker compose exec backend sh

## フロントエンドコンテナにシェル接続
frontend-shell:
	docker compose exec frontend sh

## MySQLクライアントに接続
db-shell:
	docker compose exec db mysql -u$${MYSQL_USER:-dashboard} -p$${MYSQL_PASSWORD:-dashboard_password} $${MYSQL_DATABASE:-github_events}

# === Tunnel ===

## ngrokでバックエンドを公開（Webhook受信用）
ngrok:
	@command -v ngrok >/dev/null 2>&1 || { echo "Error: ngrok is not installed. See https://ngrok.com/download"; exit 1; }
	ngrok http 8080

# === Health Check ===

## バックエンドのヘルスチェック
health:
	@curl -sf http://localhost:8080/api/health && echo " OK" || echo " FAIL"

# === Help ===

## コマンド一覧を表示
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Docker:"
	@echo "  up              全サービスをバックグラウンドで起動"
	@echo "  build           全サービスをビルドして起動"
	@echo "  down            全サービスを停止"
	@echo "  clean           全サービスを停止しボリュームも削除（DB初期化）"
	@echo "  restart         全サービスを再起動"
	@echo "  logs            全サービスのログを表示"
	@echo "  logs-backend    バックエンドのログを表示"
	@echo "  logs-frontend   フロントエンドのログを表示"
	@echo "  logs-db         DBのログを表示"
	@echo "  ps              サービス一覧を表示"
	@echo ""
	@echo "Shell:"
	@echo "  backend-shell   バックエンドコンテナにシェル接続"
	@echo "  frontend-shell  フロントエンドコンテナにシェル接続"
	@echo "  db-shell        MySQLクライアントに接続"
	@echo ""
	@echo "Tunnel:"
	@echo "  ngrok           ngrokでバックエンドを公開（Webhook受信用）"
	@echo ""
	@echo "Other:"
	@echo "  health          バックエンドのヘルスチェック"
	@echo "  help            このヘルプを表示"
