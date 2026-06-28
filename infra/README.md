# trade-analyzer インフラ（AWS CDK / TypeScript）

`doc/dev-spec/infra_requirements.md` の AWS 側構成をコード化した CDK アプリです。

## 構築されるリソース

| リソース | 用途 |
|----------|------|
| Lambda（`ApiFunction`） | Go本体（`back/`）の同期API。`provided.al2023` / 256MB / 30s |
| Lambda（`WorkerFunction`） | 同一Goバイナリの定期バッチ。`provided.al2023` / 512MB / 15分 |
| API Gateway HTTP API | 全ルート（`/api`・`/internal`・`/health`）を `ApiFunction` へプロキシ（payload v2） |
| S3（`LearningBucket`） | 学習CSVのバージョン管理（`learning_versions` のソース）。versioning有効 |
| EventBridge `AnalyzeSchedule` | 平日15:30 JST → `WorkerFunction` を直接invoke（input `{"job":"analyze"}`） |
| EventBridge `WeeklyReportSchedule` | 日曜18:00 JST → `WorkerFunction` を直接invoke（input `{"job":"weekly_report"}`） |
| EventBridge `StockFetchSchedule` | 平日15:00 JST → Python株価取得Lambda（`-c stockFetchLambdaArn=...` 指定時のみ） |

> **Neon は AWS リソースではない**ため本スタックでは管理しません。接続URLはシークレット
> （`NEON_DATABASE_URL`）として Lambda の環境変数に渡されます。
>
> 分析/週次は **EventBridge → Worker Lambda の直接呼び出し（案B）**。`main.go` の `dispatch` が
> 定数input `{"job": "..."}` を判定し、`AnalysisUsecase.RunScheduled` / `ReportUsecase.RunWeekly` を実行します。
> パイプライン本体、および Python 株価取得 Lambda（`lambda/`）は別フェーズです。現状はエントリポイントが
> スタブ（ログのみ・正常終了）で、EventBridge→Lambda の経路は先行して稼働します。

## 前提（デプロイ前に1度だけ）

1. **Node 依存のインストール**
   ```bash
   cd infra && npm install
   ```
2. **Go Lambda のビルド**（`back/lambda-handler.zip` を生成）
   ```bash
   cd ../back && make build-lambda
   ```
3. **Secrets Manager にアプリ用シークレットを作成**（名前は `cdk.json` の `appSecretName`＝`trade-analyzer/app`）
   ```bash
   aws secretsmanager create-secret --name trade-analyzer/app --secret-string '{
     "NEON_DATABASE_URL": "postgresql://...:...@...neon.tech/db?sslmode=require",
     "JWT_SECRET": "<32文字以上のランダム値>",
     "INTERNAL_API_SECRET": "<ランダム値>",
     "ANTHROPIC_API_KEY": "sk-ant-...",
     "LINE_CHANNEL_ACCESS_TOKEN": "...",
     "LINE_USER_ID": "U..."
   }'
   ```
   > LINE/Claude をまだ使わない場合、該当キーは空文字で可（アプリ側がスタブにフォールバックします）。
4. **CDK ブートストラップ**（アカウント／リージョンごとに1度）
   ```bash
   cd ../infra
   export CDK_DEFAULT_REGION=ap-northeast-1
   npx cdk bootstrap
   ```

## デプロイ

```bash
cd infra
export CDK_DEFAULT_REGION=ap-northeast-1
npm run deploy           # = cdk deploy
```

出力された `ApiBaseUrl` をフロントの `NEXT_PUBLIC_API_BASE_URL` に設定します。

### よく使う context 上書き

```bash
# 本番フロントのオリジン（CORS許可）と Claude モデルを指定してデプロイ
# モデル未指定時の既定: 毎日分析=claude-sonnet-4-6 / 週次学習=claude-opus-4-8
npx cdk deploy \
  -c frontendOrigin=https://app.example.com \
  -c anthropicModel=claude-sonnet-4-6 \
  -c anthropicModelWeekly=claude-opus-4-8 \
  -c stockFetchLambdaArn=arn:aws:lambda:ap-northeast-1:123456789012:function:stock-fetch
```

> Claude モデルはジョブ別: `anthropicModel`＝毎日分析（高頻度・コスト重視）、`anthropicModelWeekly`＝
> 週次学習（月4回・高レバレッジ・推論重視）。未指定なら上記の既定が使われます。

## マイグレーション（初回）

Lambda には migrate を組み込んでいません。Neon に対してローカルから実行します。

```bash
cd ../back
NEON_DATABASE_URL="postgresql://...?sslmode=require" go run migrate/migrate.go
```

本番用 seed（管理者＋招待コードのみ）は別途スクリプトを用意してください（`infra_requirements.md` 3章）。

## 片付け

```bash
npx cdk destroy
```

> S3 バケットは `RemovalPolicy.RETAIN`（学習データ保護）のため destroy では削除されません。
> 不要なら手動で空にして削除してください。
