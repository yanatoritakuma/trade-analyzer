# 実装仕様 - ポートフォリオ

**画面ID**: SCR-06
**元仕様**: doc/feature-spec/feature_06_portfolio.md
**パス(フロント)**: `/portfolio`
**対象ロール**: user / admin（**閲覧のみ**・表示は管理者の共通ポートフォリオ）
**作成日**: 2026-06-27
**バージョン**: 1.0

## 1. 概要・スコープ
- バーチャル/実運用の保有状況と損益推移グラフ（折れ線）・サマリー・保有ポジション表を表示。
- 含み益＝最新終値（`stock_prices.close`）− 取得単価。損益推移は `trades` の累計損益から算出。
- 依存: `trades`・`real_positions`・`stock_prices`。

## 2. 権限・アクセス制御
- 画面: 認証必須・user/admin閲覧可。
- API: 全てGET（protected）。

## 3. データモデル / DBアクセス
- **サマリー**: `/api/portfolio/summary`（SCR-03と共有・virtual/real別の累計損益・勝率）。
- **損益推移グラフ**: `trades`（admin・mode別）の決済（`closed_at`）を日付順に累積した時系列。
  - X=日付、Y=累計損益。期間切替（1/3/6ヶ月）はfromで絞る。
  - ※ `stock_prices` は最新スナップショットのみで株価ヒストリは持たないため、グラフは
    **株価ではなく確定損益の累積**で描く（feature_06準拠）。
- **保有ポジション**:
  - バーチャル: `trades`(mode=virtual) の未決済（`closed_at IS NULL`）を銘柄ごとに集計
    （数量合計・加重平均取得単価）。
  - 実運用: `real_positions`（adminの登録）。
  - 各行に `stock_prices.close` を結合 → 含み益=`(close - avg_price) * quantity`、損益率=`(close/avg_price - 1)*100`。

## 4. API仕様（openapi断片）

| メソッド | パス | 認可 | 説明 |
|---------|------|------|------|
| GET | `/api/portfolio/summary` | protected | 損益サマリー（SCR-03共有） |
| GET | `/api/positions` | protected | 実運用保有株（現在値・含み益付き） |
| GET | `/api/trades?mode=virtual` | protected | バーチャル保有/損益推移の元データ |

```yaml
components:
  schemas:
    Position:
      type: object
      properties:
        id: { type: integer }
        ticker: { type: string }
        name: { type: string, nullable: true }
        quantity: { type: integer }
        avg_price: { type: number }
        close: { type: number, nullable: true }
        unrealized_pnl: { type: number, nullable: true }   # 含み益
        pnl_rate: { type: number, nullable: true }         # 損益率%
    PnlPoint:
      type: object
      properties:
        date: { type: string, format: date }
        cumulative_pnl: { type: number }
```

> 損益推移の時系列は**専用APIを新設せず**、`/api/trades?mode=...` の決済トレード（`closed_at` あり）を
> フロントで日付昇順に累積して描画する（`PnlPoint[]` はクライアント生成）。新規エンドポイントは作らない。

## 5. バックエンド実装（DDD層別）

| 層 | ファイル | 追加/変更 | 責務 |
|----|---------|----------|------|
| usecase | `usecase/portfolio_usecase.go` | 変更 | `Summary`（SCR-03で実装済）・`Positions(ctx)` |
| usecase | `usecase/position_usecase.go` | 変更 | `ListWithPrice(ctx)`：real_positions+stock_prices |
| repository | `repository/position_repository.go` | 変更 | `FindByAdminWithPrice()` |
| repository | `repository/trade_repository.go` | 変更 | `FindOpenVirtualPositions()`（未決済集計）・`FindClosedForTimeseries(mode, from)` |
| repository | `repository/stock_price_repository.go` | 既存 | `FindByTickers([]ticker)` |
| controller | `controller/portfolio_controller.go` | 変更 | `Summary`・`Positions` |
| controller | `controller/position_controller.go` | 変更 | `GetAll` |
| router | `router/router.go` | 変更 | protected GET登録 |

- 含み益はクエリ取得後にusecaseで算出（`stock_prices` が無い銘柄はnull）。

## 6. フロントエンド実装（Next.js）
- ページ: `app/portfolio/page.tsx`（Client。タブ・期間切替の状態を持つ）。
- グラフ: `components/elements/chartBox` に Recharts `LineChart`（X=日付・Y=累計損益）。期間ボタン1/3/6ヶ月。
- サマリーカード: 累計損益・勝率。
- 保有ポジション表: `tableBox`。含み益（緑/赤）・損益率。行クリックで `/trades?ticker=...` へ。
- ローディング: グラフ/表をSkeleton。空: 「現在保有中のポジションはありません」。

## 7. バリデーション
- 入力なし（参照画面）。

## 8. 外部連携
- なし。

## 9. テスト観点
- usecase: 未決済バーチャルの加重平均取得単価・含み益算出（stock_prices欠損時null）。
- repository: `FindOpenVirtualPositions` の銘柄別集計・実運用JOIN。

## 10. 実装タスク分解
- [ ] repository: 未決済集計・実運用+価格JOIN・時系列元データ
- [ ] usecase: `Positions`・（必要なら）timeseries
- [ ] controller/router
- [ ] `app/portfolio/page.tsx`（グラフ・タブ・期間・ポジション表）
- [ ] 損益推移はフロントで `/api/trades` の決済データから累積（専用APIなし）

## 11. 受け入れ条件
- タブで virtual/real を切替えられ、損益推移グラフが期間切替で更新される。
- 保有ポジションの含み益・損益率が最新終値ベースで表示される。
- ポジション行クリックで該当銘柄のトレード履歴へ遷移する。

## 12. 未確定事項
- なし（損益推移は専用APIを設けず `/api/trades` 決済データのフロント累積、と確定）。
- 保有ポジション表の銘柄名は実運用=`real_positions.name`、バーチャル=`trades.name` を使用。
