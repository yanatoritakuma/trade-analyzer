# 実装仕様 - 週次レポート

**画面ID**: SCR-07
**元仕様**: doc/feature-spec/feature_07_reports.md
**パス(フロント)**: `/reports` / `/reports/:week`
**対象ロール**: user / admin（**閲覧のみ**・全ユーザー共通レポート）
**作成日**: 2026-06-27
**バージョン**: 1.0

## 1. 概要・スコープ
- AIが毎週日曜18:00に生成した週次学習レポート（管理者のtradesから生成）を一覧・詳細表示。
- 一覧: 期間・勝率・損益・トレード数。詳細: サマリー4指標＋AI学習メモ（有効/失敗/来週）＋当週トレード一覧。
- 依存: `learning_logs`（共通・user_id無し）。当週トレード一覧は `trades`（admin）。

## 2. 権限・アクセス制御
- 画面: 認証必須・user/admin閲覧可。
- API: GET（protected）。

## 3. データモデル / DBアクセス
- `learning_logs`: `week_start`・`week_end`・`trade_count`・`win_rate`・`total_pnl`・`summary`・`lessons`・`strategy`。
- 一覧: `ORDER BY week_start DESC`。
- 詳細: `:week` は **`week_start`（YYYY-MM-DD・月曜）** で1件取得。
- 最大ドローダウンは `learning_logs` に列を**追加せず**、`report_usecase.Detail` が当週 `trades`（admin・決済）の
  累積損益曲線から算出して返す（ピーク−ボトムの最大下落幅）。
- 当週トレード一覧: `trades`（admin・`created_at BETWEEN week_start AND week_end`）。

## 4. API仕様（openapi断片）

| メソッド | パス | 認可 | 説明 |
|---------|------|------|------|
| GET | `/api/reports` | protected | レポート一覧 |
| GET | `/api/reports/:week` | protected | レポート詳細（week=week_start・YYYY-MM-DD） |

```yaml
components:
  schemas:
    ReportSummary:
      type: object
      properties:
        week_start: { type: string, format: date }
        week_end: { type: string, format: date }
        trade_count: { type: integer }
        win_rate: { type: number }
        total_pnl: { type: number }
    ReportDetail:
      allOf:
        - $ref: '#/components/schemas/ReportSummary'
        - type: object
          properties:
            max_drawdown: { type: number, nullable: true }
            summary: { type: string }
            lessons: { type: string }
            strategy: { type: string }
            trades: { type: array, items: { $ref: '#/components/schemas/Trade' } }
```

## 5. バックエンド実装（DDD層別）

| 層 | ファイル | 追加/変更 | 責務 |
|----|---------|----------|------|
| domain | `domain/learning/learning_log.go` | 既存 | エンティティ |
| domain | `domain/learning/repository.go` | 変更 | `FindAll()`・`FindByWeekStart(date)` |
| usecase | `usecase/report_usecase.go` | 変更 | `List(ctx)`・`Detail(ctx, weekStart)`（当週trades結合・DD算出） |
| repository | `repository/learning_repository.go` | 変更 | 上記取得 |
| repository | `repository/trade_repository.go` | 既存 | `FindByAdminBetween(start,end)` |
| controller | `controller/report_controller.go` | 変更 | `GetAll`・`GetByWeek` |
| router | `router/router.go` | 変更 | protected GET2本 |

> レポートの**生成**は本画面の範囲外（`/internal/weekly-report` ＋ `report_usecase` の生成系で別途実装）。本仕様は**閲覧**のみ。

## 6. フロントエンド実装（Next.js）
- 一覧: `app/reports/page.tsx`（RSC）。`apiClient.get('/api/reports')`。カード（期間・勝率・損益・件数・[詳細を見る]）。
- 詳細: `app/reports/[week]/page.tsx`（RSC）。`apiClient.get('/api/reports/{week}')`。
  - 4指標カード（勝率・損益・取引数・最大DD）
  - AI学習メモ（✅有効/❌失敗/📌来週）= `lessons`/`summary`/`strategy` をセクション表示
  - 当週トレード表（SCR-05のテーブルコンポーネント流用）
- 空: 「まだレポートがありません。毎週日曜18:00に自動生成されます」。

## 7. バリデーション
- `:week` は `YYYY-MM-DD` 形式チェック（不正は404）。

## 8. 外部連携
- なし（生成側がClaude/S3/LINEを使用）。

## 9. テスト観点
- usecase: `Detail` で当週tradesが正しく結合される / 該当週なしで404。
- repository: `FindByWeekStart` の一致。

## 10. 実装タスク分解
- [ ] repository `FindAll`/`FindByWeekStart`/`FindByAdminBetween`
- [ ] usecase `List`/`Detail`（DD算出）
- [ ] controller/router
- [ ] `app/reports/page.tsx`・`app/reports/[week]/page.tsx`

## 11. 受け入れ条件
- 週次レポートが新しい順に一覧表示される。
- 詳細でAI学習メモと当週トレード一覧が表示される。

## 12. 未確定事項
- なし（最大ドローダウンは `report_usecase.Detail` で当週tradesから算出、と確定）。
