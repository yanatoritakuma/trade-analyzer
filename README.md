# trade-analyzer

AI株式トレーディングシステム。

- `front/` … フロントエンド（Next.js 15 / TypeScript）
- `back/` … バックエンド（Go / Gin・DDD構成）
- `doc/` … 仕様書・開発マニュアル（`doc/dev-spec/development_manual.md` ほか）

詳細な開発ガイドは [`doc/dev-spec/development_manual.md`](doc/dev-spec/development_manual.md) を参照してください。

---

## 必要なツール

| ツール | バージョン |
|--------|-----------|
| Docker Desktop | 最新 |
| Node.js | v20 以上 |
| Go | 1.25 以上（ローカルでマイグレーションをホスト実行する場合のみ） |

---

## ローカル開発環境の起動

バックエンド・ローカルDB（PostgreSQL）は Docker、フロントエンドは `npm` で起動する。
ローカルDBは Docker の PostgreSQL コンテナを使う（本番/ステージングは Neon）。

### 1. バックエンド + ローカルDB（PostgreSQL）の起動

```bash
cd back

# 初回のみ: 環境変数ファイルを作成（これが無いとDBコンテナが起動しない）
cp .env.example .env

# ホットリロード（air）でバックエンドとローカルDBを起動
docker compose -f docker-compose.dev.yml up --build
```

- バックエンド: http://localhost:8080
- ローカルDB（PostgreSQL）: `localhost:5432`
- `db`（PostgreSQL）が healthy になってから `app`（Go）が起動する。

> **補足**: `.env` は Git 管理外。`${POSTGRES_USER}` などの値は `back/.env` から読み込まれるため、初回の `cp .env.example .env` を必ず実行すること。

### 2. DBマイグレーション（初回・モデル変更時）

別ターミナルで実行する（テーブルを作成する）。

```bash
cd back
docker compose -f docker-compose.dev.yml exec app go run migrate/migrate.go
```

### 3. フロントエンドの起動

```bash
cd front

# 初回のみ
cp .env.example .env.local
npm install

# API型を openapi.yaml から生成（任意・API変更時に再実行）
npm run generate:api

# 開発サーバー起動（Turbopack）
npm run dev
```

- フロントエンド: http://localhost:3000

---

## 動作確認

| URL / コマンド | 内容 |
|----------------|------|
| http://localhost:3000 | フロントエンド画面 |
| http://localhost:8080/health | バックエンドのヘルスチェック（`{"db":"up","status":"ok"}`） |
| `docker compose -f docker-compose.dev.yml exec db psql -U trade -d trade_analyzer -c '\dt'` | ローカルDBのテーブル一覧 |

### pgAdmin など GUI からローカルDBに接続する場合

| 項目 | 値 |
|------|-----|
| Host | `localhost` |
| Port | `5432` |
| Database | `trade_analyzer` |
| Username | `trade` |
| Password | `trade` |

> `localhost:5432` は PostgreSQL のポート。ブラウザでは開けない（pgAdmin / psql から接続する）。

---

## 停止

```bash
# バックエンド + ローカルDB
cd back
docker compose -f docker-compose.dev.yml down

# DBのデータも削除する場合（テーブル・データを初期化）
docker compose -f docker-compose.dev.yml down -v

# フロントエンド: 起動したターミナルで Ctrl + C
```

---

## DB接続先の切り替え（ローカル / Neon）

アプリは接続先を `DATABASE_URL` → `NEON_DATABASE_URL` の順に解決する（`back/db/db.go`）。

- **ローカル（Docker）**: `back/.env` の `DATABASE_URL` が `db` サービスを指す（デフォルト）。
- **Neon（本番/ステージング）**: `DATABASE_URL` を空にして `NEON_DATABASE_URL` に Neon の接続URLを設定する。
