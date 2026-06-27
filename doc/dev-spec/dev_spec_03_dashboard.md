# 実装仕様 - ダッシュボード

**画面ID**: SCR-03
**元仕様**: doc/feature-spec/feature_03_dashboard.md
**パス(フロント)**: `/`
**対象ロール**: user / admin（**閲覧のみ**・表示は管理者の共通ポートフォリオ）
**作成日**: 2026-06-27
**バージョン**: 1.0

## 1. 概要・スコープ
- ログイン後トップ。損益サマリー（バーチャル/実運用）・最新分析シグナル（直近3件）・ウォッチリスト（現在値/前日比）を表示。
- すべて参照系。書き込みなし。一般ユーザーには管理者の共通データを表示。
- 依存: `trades`・`real_positions`・`analysis_logs`・`watchlist`・`stock_prices`。

## 2. 権限・アクセス制御
- 画面: 認証必須（middlewareで未認証は `/login`）。user/admin両方閲覧可。
- API: すべて `protected`（GET・user/admin共通）。

## 3. データモデル / DBアクセス
- **損益サマリー**（バーチャル/実運用別）: `trades`（**admin の trades**、`mode` 別）を集計。
  - 累計損益 = `SUM(result_pnl)`（決済済み）、今週損益 = **今週月曜0:00(JST)以降に決済(`closed_at`)**の `SUM(result_pnl)`、
    勝率 = `result_pnl > 0` の件数 / 決済件数、トレード数 = 件数。
  - 記録先が admin のみ（共通ポートフォリオ）のため、呼び出しユーザーに関わらず admin の集計を返す。
- **最新シグナル**: `analysis_logs` を `analyzed_at DESC LIMIT 3`（`idx_analysis_logs_analyzed_at`）。
  銘柄名は `watchlist` を `ticker` でLEFT JOIN（最新シグナルは現役ウォッチリスト銘柄のため取得可能）。
- **ウォッチリスト**: `watchlist`（`is_active=true`）に `stock_prices` を `ticker` で LEFT JOIN し、
  `close`（現在値）・`change_rate`（前日比）を付与。
- データ分離: `trades`/`real_positions` は user_idを持つが本画面では **admin固定**で集計。共通テーブルは分離不要。

## 4. API仕様（openapi断片）

| メソッド | パス | 認可 | 説明 |
|---------|------|------|------|
| GET | `/api/portfolio/summary` | protected | バーチャル/実運用の損益サマリー |
| GET | `/api/analysis/latest` | protected | 直近3件の分析シグナル |
| GET | `/api/watchlist` | protected | ウォッチリスト（現在値・前日比付き） |

```yaml
components:
  schemas:
    PortfolioSummary:
      type: object
      properties:
        virtual: { $ref: '#/components/schemas/ModeSummary' }
        real:    { $ref: '#/components/schemas/ModeSummary' }
    ModeSummary:
      type: object
      properties:
        total_pnl:    { type: number }   # 累計損益
        weekly_pnl:   { type: number }   # 今週損益
        win_rate:     { type: number }   # 勝率(%)
        trade_count:  { type: integer }  # 累計トレード数
    AnalysisSignal:
      type: object
      properties:
        ticker: { type: string }
        name: { type: string }
        action: { type: string, enum: [BUY, SELL, HOLD] }
        confidence: { type: number }
        analyzed_at: { type: string, format: date-time }
    WatchlistItem:
      type: object
      properties:
        id: { type: integer }
        ticker: { type: string }
        name: { type: string }
        mode: { type: string, enum: [virtual, real, both] }
        close: { type: number, nullable: true }        # 現在値（stock_prices）
        change_rate: { type: number, nullable: true }  # 前日比%（stock_prices）
```

> `/api/portfolio/summary` は **SCR-06 ポートフォリオと共有**。レスポンス形をここで確定し両画面で使う。

## 5. バックエンド実装（DDD層別）

| 層 | ファイル | 追加/変更 | 責務 |
|----|---------|----------|------|
| usecase | `usecase/portfolio_usecase.go` | 新規 | `Summary(ctx)`：adminのtradesからvirtual/real集計 |
| usecase | `usecase/analysis_usecase.go` | 変更 | `Latest(ctx, limit)`：analysis_logs取得 |
| usecase | `usecase/watchlist_usecase.go` | 変更 | `FindAllWithPrice(ctx)`：watchlist+stock_prices |
| repository | `repository/trade_repository.go` | 変更 | `AggregateByAdmin(mode)` 集計クエリ |
| repository | `repository/analysis_repository.go` | 変更 | `FindLatest(limit)` |
| repository | `repository/watchlist_repository.go` | 変更 | `FindAllWithPrice()`（JOIN stock_prices） |
| repository | `repository/stock_price_repository.go` | 新規 | `FindByTicker`（現在値・前日比参照） |
| controller | `controller/portfolio_controller.go` | 新規 | `Summary` |
| controller | `controller/analysis_controller.go` | 変更 | `Latest` |
| controller | `controller/watchlist_controller.go` | 変更 | `GetAll` |
| router | `router/router.go` | 変更 | protectedにGET3本登録 |

- 集計は参照のみのためUoW不要。adminのuser_id特定は `users WHERE role='admin'`（先頭1件）をサブクエリ/JOINで。

## 6. フロントエンド実装（Next.js）
- ページ: `app/page.tsx`（**RSC**）。3APIを `Promise.all` で並列fetch（`cache: 'no-store'`）。
- レイアウト: `components/Navbar`・`components/Sidebar`（共通レイアウトは `app/layout.tsx`）。
- コンポーネント:
  - `VirtualSummaryCard` / `RealSummaryCard`（損益は緑/赤）
  - `LatestSignalList`（BUY緑/SELL赤/HOLDグレー・クリックで `/trades`）
  - `WatchlistSummary`（現在値・前日比・クリックで `/watchlist`）
- ローディング: `loading.tsx` または各カードをMUI Skeleton。データなしは「まだデータがありません」。
- ログアウト: Navbarのボタン → `apiClient.post('/api/auth/logout')` → `/login`。

## 7. バリデーション
- 入力なし（参照画面）。

## 8. 外部連携
- なし（分析・通知はLambda/定期実行側）。

## 9. テスト観点
- usecase: `portfolio.Summary` の集計（勝率・週次・累計）境界値（決済0件で割り算回避）。
- repository: 集計SQL（adminのtradesのみ・mode別）の結合テスト。

## 10. 実装タスク分解
- [ ] `portfolio_usecase.Summary`＋trade集計クエリ
- [ ] `analysis.Latest`・`watchlist.FindAllWithPrice`・`stock_price_repository`
- [ ] controller/router 3本
- [ ] `app/page.tsx` RSC＋カード/リストコンポーネント
- [ ] Skeleton/空表示/エラー表示

## 11. 受け入れ条件
- バーチャル/実運用それぞれの累計損益・今週損益・勝率・トレード数が表示される。
- 最新シグナル3件、ウォッチリスト現在値・前日比が表示される。
- シグナルクリックで `/trades`、ウォッチリストクリックで `/watchlist` に遷移。

## 12. 未確定事項
- なし（「今週損益」は今週月曜0:00(JST)起点の決済損益、と確定）。
