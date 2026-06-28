# インフラ開発 要件・手順書

**対象**: ISSUE #5（フロント・バックエンド実装）完了後の次フェーズ（インフラ構築）
**作成日**: 2026-06-28
**前提**: アプリ本体（`front/`・`back/`）は実装済み。本書は **AWS/Neon へのデプロイ** と
**外部連携（Claude API・LINE通知・S3・株価取得Lambda）の本実装** に必要な情報をまとめる。

> ISSUE #5 の範囲では外部連携は **スタブ実装**（`back/external/external.go`）とし、
> ローカルは Docker PostgreSQL + seed データで動作確認している。本書の項目を実装することで
> 本番のデータパイプライン（株価取得 → 分析 → 通知 → 学習）が稼働する。

---

## 1. 全体アーキテクチャ

```
                ┌─────────────────────────┐
   ブラウザ ───▶│ Next.js (Vercel等)       │  NEXT_PUBLIC_API_BASE_URL
                └───────────┬─────────────┘
                            │ HTTPS + Cookie(SameSite/Secure)
                            ▼
          ┌──────────────────────────────────┐
          │ API Gateway (HTTP API)            │
          │   └─ Lambda: ApiFunction (back/)   │──┐
          └──────────────────────────────────┘  │ DATABASE接続
              ▲                                   ▼
   X-Internal-Secret(株価取得のみ)        ┌────────────────────┐
              │                          │ Neon (PostgreSQL)  │
   ┌──────────┴─────────────────┐        └────────────────────┘
   │ EventBridge (定期実行)       │                ▲
   │  ├─ 平日15:00 株価取得Lambda │─(yfinance)─▶ /internal/stock-prices
   │  │                          │     ┌──────────────────────────────┐
   │  ├─ 平日15:30 分析実行        │──直接invoke──▶│ Lambda: WorkerFunction │
   │  └─ 日曜18:00 週次レポート     │  {"job":...}  │  (同一Goバイナリ/最大15分) │
   └────────────────────────────┘     └──────────────┬───────────────┘
        (案B: API Gatewayを介さずIAM認証で直接呼び出し)   │
                  ▶ Claude API ─▶ trades/analysis_logs ─▶ LINE通知
                  ▶ Claude API + S3 ─▶ learning_logs/learning_versions
```

| レイヤ | サービス | 役割 |
|--------|---------|------|
| フロント | Vercel 等 | Next.js（App Router）ホスティング |
| API | AWS API Gateway (HTTP API) + Lambda (`ApiFunction`) | Go本体（`back/`）の同期API。`make build-lambda` でデプロイ |
| バッチ | AWS Lambda (`WorkerFunction`) | 同一Goバイナリ。分析・週次をEventBridgeから直接実行（最大15分・案B） |
| 株価取得 | AWS Lambda (Python) | yfinanceで株価取得し `/internal/stock-prices` にPOST |
| 定期実行 | AWS EventBridge | 株価取得・分析・週次レポートのスケジュール起動 |
| DB | Neon (Serverless PostgreSQL) | 本番/ステージング |
| ストレージ | AWS S3 | 学習CSV（`learning_versions`）のバージョン管理 |
| 分析 | Anthropic Claude API | 銘柄分析・週次学習の生成 |
| 通知 | LINE Messaging API | 売買シグナル・候補提案の通知 |

---

## 2. 環境変数・シークレット一覧

`back/.env.example` をベースに、本番は AWS Lambda の環境変数 / Secrets Manager で管理する。

| 変数 | 用途 | 取得先 | 現状 |
|------|------|--------|------|
| `NEON_DATABASE_URL` | Neon接続URL（`DATABASE_URL` を空に） | Neon管理画面 | 要設定 |
| `APP_ENV` | `production` で Cookie に Secure 付与 | — | 本番=production |
| `FRONTEND_ORIGIN` | CORS許可する本番フロントURL | デプロイ先 | 要設定 |
| `JWT_SECRET` | JWT署名鍵（十分長いランダム値） | 生成 | 要差し替え |
| `JWT_ACCESS_EXPIRE_HOURS` / `JWT_REFRESH_EXPIRE_DAYS` | トークン有効期限 | — | 既定24h/30d |
| `ANTHROPIC_API_KEY` | Claude API | Anthropic Console | **未実装（スタブ）** |
| `LINE_CHANNEL_ACCESS_TOKEN` / `LINE_USER_ID` | LINE通知 | LINE Developers | **未実装（スタブ）** |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_REGION` / `S3_BUCKET_NAME` | S3操作 | AWS IAM/S3 | **未実装（スタブ）** |
| `INTERNAL_API_SECRET` | Lambda → Go の内部API認証 | 生成 | 実装済（要差し替え） |

> **Cookie の本番設定**: `APP_ENV=production` のとき `Secure` が付与される（`back/utils/cookie.go`）。
> `SameSite=Strict`・`HttpOnly` は常時。フロントとAPIが **別ドメイン** になる場合は、
> Cookie が送信されるよう **同一サイト構成（同一登録可能ドメインのサブドメイン）** にするか、
> `SameSite=None; Secure` への変更とCORS資格情報設定を検討する（要design判断）。

---

## 3. Neon (PostgreSQL) セットアップ

1. Neon でプロジェクト/ブランチを作成し、接続URL（`postgresql://...?sslmode=require`）を取得。
2. Lambda 環境変数に `NEON_DATABASE_URL` を設定（`DATABASE_URL` は空にする。接続解決順は `back/db/db.go` 参照）。
3. マイグレーション実行（テーブル作成）:
   ```bash
   cd back
   NEON_DATABASE_URL="postgresql://..." go run migrate/migrate.go
   ```
4. 初期管理者ユーザー・招待コードの投入は `back/seed/seed.go` を参考に、
   **本番用は最低限「管理者1名 + 招待コード」のみ** を投入するスクリプトを用意すること
   （サンプルのtrades/analysis等は本番に入れない）。
5. マイグレーション運用上の注意は `db_definition.md` / `development_manual.md`（6章）の
   AutoMigrate 制約（カラム削除・型変更は手動ALTER）を参照。

---

## 4. Go本体（back/）の Lambda デプロイ

`back/main.go` は `LAMBDA_TASK_ROOT` 環境変数の有無で Lambda/ローカルを自動切替する（実装済）。

```bash
cd back
make build-lambda     # bootstrap を生成し lambda-handler.zip を作成
```

- API Gateway は **HTTP API (payload v2)** を使用（`ginadapter.NewV2`）。
- ルート: `/api/*`（一般・認証）, `/api/admin/*`（admin）, `/internal/*`（Lambda内部API）, `/health`。
- Lambda の環境変数に 2章の値を設定。VPC は Neon 接続要件に応じて設定（Neonは公開エンドポイントのためVPC不要な構成も可）。

---

## 5. 株価取得Lambda（Python・未実装）

**役割**: ウォッチリスト銘柄の株価を yfinance で取得し、Go の内部APIへPOSTする。
`development_manual.md`（5章）に参照実装あり。`lambda/` ディレクトリに配置する想定。

### 5.1 処理フロー
1. `GET /internal/watchlist`（ヘッダ `X-Internal-Secret`）で対象銘柄を取得。
2. 各 `ticker` の過去120日分 OHLCV を `yfinance` で取得。
3. `POST /internal/stock-prices` に `{fetched_at, stocks:[{ticker, prices:[{date,open,high,low,close,volume}]}]}` を送信。
4. Go側（`usecase/internal_usecase.go`）が末尾2営業日から前日比を算出し `stock_prices` をUPSERT。

### 5.2 内部API契約（実装済み・Go側）
- `GET  /internal/watchlist` → `[{id, ticker, name, mode}]`
- `POST /internal/stock-prices` → `{message, count}`（`StockPricesIngestRequest` は `openapi.yaml` 参照）
- 認証: ヘッダ `X-Internal-Secret: <INTERNAL_API_SECRET>`

### 5.3 必要ライブラリ
`yfinance`, `requests`, `pandas`, `python-dotenv`（`development_manual.md` 5.3）。

---

## 6. 分析パイプライン（Claude / LINE / S3・未実装）

現状 `back/external/external.go` に **インターフェース + スタブ** が用意済み。本番ではスタブを実装に差し替える。

| インターフェース | 実装すべき内容 | 呼び出し元（想定） |
|------------------|---------------|--------------------|
| `external.ClaudeClient.Analyze` | Anthropic Messages API を叩き、銘柄分析JSON（action/confidence/buy_reasons/no_buy_reasons/entry_condition）を返す。最新の学習CSV（S3）とユーザー設定（`analysis_settings`）をプロンプトに含める | 分析実行（平日15:30） |
| `external.LineClient.Notify` | LINE Messaging API（push message）で売買シグナル・候補提案を通知 | 分析後・候補提案後 |
| `external.S3Client.Upload` | 週次学習CSVをS3へPUTし、`learning_versions` に記録 | 週次レポート生成（日曜18:00） |

### 6.1 分析実行（平日15:30・新規 usecase が必要）
1. `analysis_settings`（is_active）と最新学習CSV（S3）を読み込む。
2. `stock_prices` + `watchlist` を対象に `ClaudeClient.Analyze` を実行。
3. 結果を `analysis_logs` に保存。BUY/SELL は `trades`（**管理者のuser_id**）に記録（`development_manual.md` 4.4 例②のUoW構成）。
4. シグナルを `LineClient.Notify` で通知。
5. ウォッチリスト入れ替え候補があれば `watchlist_candidates` に保存し通知（承認はUI=SCR-13）。

> **注意**: Claude/LINE/S3 は **トランザクション外** で呼ぶ（`development_manual.md` 4.4 のルール）。
> DB保存成功後にのみ通知する。

### 6.2 週次レポート生成（日曜18:00・新規 usecase が必要）
1. 当週（月〜日）の管理者 `trades` を集計。
2. `ClaudeClient.Analyze`（学習用プロンプト）で `summary`/`lessons`/`strategy` を生成。
3. `learning_logs` に保存。学習CSVを更新して `S3Client.Upload` → `learning_versions` に記録。
4. 閲覧はSCR-07（実装済み・`/api/reports`）。

---

## 7. EventBridge スケジュール

| ルール | スケジュール（JST） | ターゲット | 呼び出し方式 |
|--------|---------------------|-----------|-------------|
| 株価取得 | 平日 15:00 | Python株価取得Lambda | 直接invoke（別フェーズ） |
| 分析実行 | 平日 15:30 | Worker Lambda（`{"job":"analyze"}`） | **直接invoke（採用）** |
| 週次レポート | 日曜 18:00 | Worker Lambda（`{"job":"weekly_report"}`） | **直接invoke（採用）** |

> **採用方式（案B）**: 分析実行・週次レポートは **EventBridge → Worker Lambda の直接呼び出し** とする。
> API Gateway / API Destination を介さないため、①EventBridge Connection 由来の Secrets Manager
> シークレット（$0.40/月）が不要、②API Gatewayの30秒制限を受けず最大15分まで実行可能、という利点がある。
> 認証は IAM で担保され `X-Internal-Secret` は不要。`main.go` の `dispatch` が定数input `{"job": "..."}` を
> 判定し、`AnalysisUsecase.RunScheduled` / `ReportUsecase.RunWeekly` を実行する（パイプライン本体は別フェーズ）。
> 詳細は `infra_architecture.md` を参照。
>
> なお HTTP の `/internal/analyze`・`/internal/weekly-report` は**定期実行には不要**だが、手動トリガー用に
> `InternalAuth` 付きで追加することは任意で可能。

---

## 8. フロントエンド デプロイ

- `front/` を Vercel 等にデプロイ。環境変数 `NEXT_PUBLIC_API_BASE_URL`（本番APIのURL）・`NEXT_PUBLIC_APP_NAME` を設定。
- `front/src/middleware.ts` が `/admin`・`/settings` を role 検証（`/api/auth/me`）でガードする。
  本番では API と Cookie ドメインが疎通することを確認（2章のCookie注意点）。

---

## 9. 残課題チェックリスト（インフラフェーズ）

- [ ] Neon プロジェクト作成・`NEON_DATABASE_URL` 設定・`migrate` 実行
- [ ] 本番用 seed（管理者 + 招待コードのみ）スクリプト作成
- [ ] `JWT_SECRET` / `INTERNAL_API_SECRET` を本番用ランダム値に差し替え
- [~] Go Lambda デプロイ（API Gateway HTTP API）※CDK化済（`infra/`）。`make build-lambda`＋`cdk deploy`。実デプロイは要AWS認証
- [ ] Python 株価取得Lambda 実装・デプロイ（`lambda/`）※CDK側は配線済（`-c stockFetchLambdaArn=...`で有効化）
- [x] `external.ClaudeClient` 本実装（Anthropic Messages API）※env未設定時はスタブにフォールバック
- [x] `external.LineClient` 本実装（LINE Messaging API push）※env未設定時はスタブにフォールバック
- [x] `external.S3Client` 本実装（S3 PUT・aws-sdk-go-v2）※`NewS3Client`実装済。分析/週次usecase配線は次フェーズ
- [~] 分析実行 / 週次レポート生成 ※エントリポイント（`RunScheduled`/`RunWeekly`）とEventBridge直接invoke・dispatch配線は実装済。パイプライン本体（Claude/集計/S3）は次フェーズ
- [~] EventBridge スケジュール3本設定 ※CDK化済（`infra/`）。分析/週次は**Worker Lambda直接invoke（案B）**、株価取得はLambda ARN指定で有効化
- [ ] フロント本番デプロイ・Cookie/CORS の本番疎通確認
- [ ] 本番 Cookie 設定（`APP_ENV=production`・別ドメイン時の SameSite 方針確定）
