# 実装仕様 - トレード履歴

**画面ID**: SCR-05
**元仕様**: doc/feature-spec/feature_05_trades.md
**パス(フロント)**: `/trades`
**対象ロール**: user / admin（**閲覧のみ**・表示は管理者の共通ポートフォリオ）
**作成日**: 2026-06-27
**バージョン**: 1.0

## 1. 概要・スコープ
- バーチャル/実運用のトレード履歴をタブ切替・フィルタ（銘柄・期間・BUY/SELL）で一覧表示。
- 行クリックで詳細モーダル（根拠・買う/買わない理由・目標/損切・信頼度）。
- 依存: `trades`（adminの共通データ）。総合コメントは `trades.reason`、買う/買わない理由は `analysis_logs.analysis`(JSONB) から取得。

## 2. 権限・アクセス制御
- 画面: 認証必須・user/admin閲覧可。
- API: `GET /api/trades`（protected）。

## 3. データモデル / DBアクセス
- `trades`: `ticker`・`name`・`mode`・`action`・`price`・`quantity`・`confidence`・`reason`・`target_price`・`stop_loss`・`result_pnl`・`closed_at`・`created_at`。
  - 銘柄名は `trades.name`（記録時に `watchlist.name` から保存済み）をそのまま使用。JOINは不要。
- 詳細モーダルの「買う/買わない理由」は `analysis_logs` を `ticker = ?` かつ `analyzed_at::date = trades.created_at::date` で1件引き、`analysis`(JSONB) の `buy_reasons`/`no_buy_reasons`/`entry_condition` を表示。
- 表示は **admin の trades**（共通ポートフォリオ）に固定。
- フィルタ → WHERE 句:
  - `mode = ?`（必須・virtual/real）
  - `ticker = ?`（任意）
  - `action = ?`（任意・BUY/SELL）
  - `created_at BETWEEN ? AND ?`（任意）
  - 並び: `created_at DESC`（`idx_trades_created_at`）
- フッター集計: 件数・勝率（`result_pnl>0`/決済件数）・累計損益（`SUM(result_pnl)`）。

## 4. API仕様（openapi断片）

`GET /api/trades` — protected

| クエリ | 型 | 必須 | 説明 |
|--------|----|------|------|
| mode | string(virtual/real) | ○ | タブ |
| ticker | string | - | 銘柄絞り込み |
| action | string(BUY/SELL) | - | 売買絞り込み |
| from / to | date(YYYY-MM-DD) | - | 期間 |

```yaml
components:
  schemas:
    Trade:
      type: object
      properties:
        id: { type: integer }
        ticker: { type: string }
        name: { type: string, nullable: true }
        mode: { type: string, enum: [virtual, real] }
        action: { type: string, enum: [BUY, SELL] }
        price: { type: number }
        quantity: { type: integer }
        confidence: { type: number, nullable: true }
        reason: { type: string, nullable: true }
        target_price: { type: number, nullable: true }
        stop_loss: { type: number, nullable: true }
        result_pnl: { type: number, nullable: true }
        closed_at: { type: string, format: date-time, nullable: true }
        created_at: { type: string, format: date-time }
        # 詳細モーダル用（analysis_logsをLEFT JOINして同梱・該当なしはnull）
        buy_reasons: { type: array, items: { type: string }, nullable: true }
        no_buy_reasons: { type: array, items: { type: string }, nullable: true }
        entry_condition: { type: string, nullable: true }
    TradeListResponse:
      type: object
      properties:
        items: { type: array, items: { $ref: '#/components/schemas/Trade' } }
        summary:
          type: object
          properties:
            count: { type: integer }
            win_rate: { type: number }
            total_pnl: { type: number }
```

## 5. バックエンド実装（DDD層別）

| 層 | ファイル | 追加/変更 | 責務 |
|----|---------|----------|------|
| domain | `domain/trade/trade.go` | 既存 | エンティティ・`Action`/`Mode` VO |
| domain | `domain/trade/repository.go` | 変更 | `FindByFilter(filter)`・集計 |
| usecase | `usecase/trade_usecase.go` | 変更 | `List(ctx, filter)`：items＋summary |
| repository | `repository/trade_repository.go` | 変更 | 動的WHERE（admin固定）・集計・`analysis_logs` LEFT JOIN |
| controller | `controller/trade_controller.go` | 変更 | `GetAll`：クエリbind→usecase |
| router | `router/router.go` | 変更 | `GET /api/trades`（protected） |

- 集計はSQLで（勝率の分母=決済件数=`closed_at IS NOT NULL`）。adminのuser_idはJOIN/サブクエリで限定。
- 詳細用の理由は `LEFT JOIN analysis_logs al ON al.ticker = trades.ticker AND al.analyzed_at::date = trades.created_at::date` で取得し、`al.analysis->'buy_reasons'` 等を展開（該当なしはnull）。

## 6. フロントエンド実装（Next.js）
- ページ: `app/trades/page.tsx`（Client）。タブ（バーチャル/実運用）で `mode` を切替えて再取得。
- フィルタ: 銘柄セレクト（ウォッチリスト由来）・期間（date-fns＋MUI）・BUY/SELLセレクト。
- テーブル: `components/elements/tableBox`。損益は緑/赤・未確定(`result_pnl=null`)はグレー。
- 行クリック: 詳細モーダル（`components/elements/modalBox`）に `reason`・目標/損切・信頼度。
- ローディング: Skeleton。空: 「トレード履歴がありません」。

## 7. バリデーション
- 入力はフィルタのみ（日付の from<=to を軽くチェック）。

## 8. 外部連携
- なし。

## 9. テスト観点
- usecase: フィルタ組合せ（mode必須・任意条件の有無）・summary集計（決済0件で勝率0）。
- repository: 動的WHEREのSQL生成・admin限定。

## 10. 実装タスク分解
- [ ] repository `FindByFilter`＋summary
- [ ] usecase `List`
- [ ] controller/router
- [ ] `app/trades/page.tsx`（タブ・フィルタ・テーブル・詳細モーダル）

## 11. 受け入れ条件
- mode切替・各フィルタで結果が絞り込まれる。
- 行クリックで詳細モーダルが開く。
- フッターに件数・勝率・累計損益が出る。

## 12. 未確定事項
- なし（銘柄名は `trades.name`、買う/買わない理由は `analysis_logs.analysis` JSONB から取得、と確定）。
