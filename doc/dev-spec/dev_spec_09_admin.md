# 実装仕様 - 管理者画面群

**画面ID**: SCR-09〜13
**元仕様**: doc/feature-spec/feature_09_12_admin.md
**パス(フロント)**: `/admin` / `/admin/users` / `/admin/invitations` / `/admin/analysis-settings` / `/admin/analysis-settings/themes` / `/admin/watchlist-candidates`
**対象ロール**: **admin のみ**
**作成日**: 2026-06-27
**バージョン**: 1.0

## 1. 概要・スコープ
- 管理者向け6画面: ①ダッシュボード ②ユーザー管理 ③招待コード管理 ④分析設定 ⑤テーマ管理 ⑥候補承認。
- すべて admin 専用（更新系は `/api/admin/*`＋`RequireAdmin`）。
- 依存: `users`・`invitation_codes`・`analysis_settings`・`analysis_themes`・`watchlist_candidates`・`watchlist`・`analysis_logs`・`trades`。

## 2. 権限・アクセス制御
- 画面: `middleware.ts` の `/admin/*` で `GET /api/auth/me` を呼びrole検証。admin以外は `/`。
- API: 取得・更新ともすべて `/api/admin/*`（`JWTAuth` + `RequireAdmin`）。
- 制約: admin自身は停止・削除のボタンを出さない（SCR-10）。削除は物理削除。

## 3. データモデル / DBアクセス（画面別）
- **SCR-09 ダッシュボード**: `users`件数・有効`invitation_codes`件数・`analysis_logs`最新日時、
  **共通ポートフォリオ（管理者のtrades）の成績サマリー**（勝率・今週損益・累計損益）。
  ※ trades記録はadminのみのため「全ユーザー別成績」は持たず、共通成績を1ブロック表示する。
  ユーザー一覧は名前・ロール・状態のみ（成績列なし）。
- **SCR-10 ユーザー管理**: `users`一覧（停止=`is_active=false`/復活=true）・物理削除（関連データはFK考慮）。
- **SCR-11 招待コード管理**: `invitation_codes` 発行（`TRADE-XXXX-XXXX`生成・`expires_at`）・一覧・無効化（`is_active=false`）。
  - ステータス: `is_active=false`→無効 / `used_by`有→使用済 / `expires_at<NOW()`→期限切 / 他→有効。
- **SCR-12 分析設定**: `analysis_settings`（`is_active=true` の1件をUPSERT）・`theme_ids`(INT[])・`screening`(JSONB)・`style`・`free_prompt`。
- **SCR-12 テーマ管理**: `analysis_themes` CRUD＋`sort_order`並び替え（`name` UNIQUE）。
- **SCR-13 候補承認**: `watchlist_candidates`（`status` pending/approved/rejected）。承認時 `watchlist` を更新。
- データ分離: `trades` のユーザー別集計のみ user_id を使う。他は共通テーブル。

## 4. API仕様（openapi断片・抜粋）

| メソッド | パス | 画面 |
|---------|------|------|
| GET | `/api/admin/users` | 09/10 |
| PATCH | `/api/admin/users/:id` | 10（停止/復活） |
| DELETE | `/api/admin/users/:id` | 10 |
| GET | `/api/admin/invitations` | 09/11 |
| POST | `/api/admin/invitations` | 11（`{expires_days}`→`{code, expires_at}`） |
| DELETE | `/api/admin/invitations/:id` | 11（無効化） |
| GET | `/api/admin/analysis-settings` | 12 |
| PUT | `/api/admin/analysis-settings` | 12 |
| GET | `/api/admin/analysis-themes` | 12 |
| POST | `/api/admin/analysis-themes` | 12 |
| PUT | `/api/admin/analysis-themes/:id` | 12 |
| DELETE | `/api/admin/analysis-themes/:id` | 12 |
| PATCH | `/api/admin/analysis-themes/sort` | 12（`[{id, sort_order}]`） |
| GET | `/api/admin/watchlist-candidates` | 13 |
| PATCH | `/api/admin/watchlist-candidates/:id/approve` | 13 |
| PATCH | `/api/admin/watchlist-candidates/:id/reject` | 13 |
| GET | `/api/analysis/latest` | 09（最終分析日時・protected） |

```yaml
components:
  schemas:
    AdminUser:
      type: object
      properties:
        id: { type: integer }
        email: { type: string }
        name: { type: string }
        role: { type: string, enum: [admin, user] }
        is_active: { type: boolean }
        created_at: { type: string, format: date-time }
    InvitationCreateRequest:
      type: object
      required: [expires_days]
      properties: { expires_days: { type: integer, enum: [3, 7, 14, 30] } }
    AnalysisSettingRequest:
      type: object
      properties:
        theme_ids: { type: array, items: { type: integer } }   # 1つ以上
        screening: { type: object }                            # {min_market_cap, min_volume, max_per}
        style: { type: string, enum: [short_term_trend, short_term_contrarian, short_term_both, mid_term_trend, mid_term_contrarian, mid_term_both] }
        free_prompt: { type: string, maxLength: 1000 }
    ThemeRequest:
      type: object
      required: [name]
      properties:
        name: { type: string, maxLength: 100 }
        description: { type: string, maxLength: 255 }
        is_active: { type: boolean }
```

## 5. バックエンド実装（DDD層別）

| 層 | ファイル | 追加/変更 | 責務 |
|----|---------|----------|------|
| usecase | `usecase/admin_usecase.go` | 変更 | ユーザー一覧/成績・停止/復活/削除 |
| usecase | `usecase/invitation_usecase.go` | 変更 | コード生成・一覧・無効化 |
| usecase | `usecase/analysis_usecase.go` | 変更 | 設定取得・保存（UPSERTで単一active） |
| usecase | `usecase/theme_usecase.go` | 変更 | テーマCRUD・並び替え |
| usecase | `usecase/candidate_usecase.go` | 変更 | 承認（**UoW**: watchlist更新＋candidate更新）・却下 |
| repository | `repository/user_repository.go` | 変更 | 一覧・更新・削除 |
| repository | `repository/invitation_repository.go` | 変更 | 生成・一覧・更新 |
| repository | `repository/analysis_repository.go` | 変更 | 設定取得/UPSERT |
| repository | `repository/theme_repository.go` | 変更 | CRUD・`sort_order`一括更新 |
| repository | `repository/candidate_repository.go` | 変更 | 取得・status更新 |
| repository | `repository/watchlist_repository.go` | 既存 | 承認時の削除/追加 |
| controller | `controller/admin_controller.go` 等 | 変更 | 各ハンドラ |
| router | `router/router.go` | 変更 | adminグループに全登録 |

**候補承認（SCR-13・UoW必須）**:
1. 候補取得（status=pending）
2. `replace_ticker` 指定あり → `watchlist` から該当削除
3. 候補を `watchlist` に追加（`ticker`・`name` は候補から、`mode` は**承認ダイアログで選択した値・既定 `both`**）
4. `watchlist_candidates.status='approved'`・`decided_by`/`decided_at` 更新
   → これらを `uow.Do` 内で一括（途中失敗で全ロールバック）
- **上限3チェック**: replace無し承認で既に3銘柄なら拒否（`ErrInvalidInput`）。

**招待コード生成**: `TRADE-` + ランダム英数字4-4（重複時リトライ）・`expires_at = NOW()+days`。

**分析設定保存**: 既存 `is_active=true` を更新（無ければINSERT）。常に1件のみactiveを保証。

## 6. フロントエンド実装（Next.js）
- 共通: `components/AdminSidebar`。各ページは admin前提（middleware保護）。
- **/admin** (`app/admin/page.tsx`): サマリーカード（ユーザー数・有効招待数・最終分析）＋共通ポートフォリオ成績サマリー（勝率・今週・累計）＋ユーザー一覧（名前・ロール・状態のみ）。
- **/admin/users** (`app/admin/users/page.tsx`): 一覧＋[停止/復活][削除]（確認ダイアログ）。admin自身は操作ボタン非表示。
- **/admin/invitations** (`app/admin/invitations/page.tsx`): 発行ダイアログ（有効期限セレクト3/7/14/30日）→発行後コードをクリップボードコピー＋Snackbar。一覧にステータス色分け。
- **/admin/analysis-settings** (`app/admin/analysis-settings/page.tsx`): テーマ複数選択（チェックボックス）・スクリーニング数値・スタイル（期間×方向→`style`へマップ）・自由プロンプト・プレビュー・[保存する]→「次回15:30から反映」Snackbar。
  - **スタイルのマッピング**（UIの期間×方向 → `style` 値）:

    | 期間＼方向 | 順張り | 逆張り | 両方 |
    |-----------|--------|--------|------|
    | 短期 | `short_term_trend` | `short_term_contrarian` | `short_term_both` |
    | 中期 | `mid_term_trend` | `mid_term_contrarian` | `mid_term_both` |
- **/admin/analysis-settings/themes** (`.../themes/page.tsx`): テーマ一覧（D&D並び替え→`PATCH /sort`）・追加/編集ダイアログ・削除。
- **/admin/watchlist-candidates** (`app/admin/watchlist-candidates/page.tsx`): 承認待ちカード（[承認]＝確認ダイアログ/[却下]＝即時）＋承認済み履歴テーブル。
- フォームは RHF+Zod。テーマ選択は1つ以上必須。自由プロンプト最大1000文字。

## 7. バリデーション
- 招待: 有効期限 enum。
- 分析設定: テーマ1つ以上 / スクリーニング 0以上 / 自由プロンプト1000文字以内。
- テーマ: 名称必須・最大100・重複不可（`name` UNIQUE→重複時「すでに存在するテーマ名です」）。

## 8. 外部連携
- 候補提案・分析設定の**反映先**は定期実行（Lambda→`/internal/stock-prices`→Go分析）。本画面は設定保存のみで分析は実行しない。
- 候補提案時のLINE通知は分析側（`analysis_usecase`）が担当。本画面は通知しない。

## 9. テスト観点
- usecase: 候補承認（replace有/無・上限3拒否・UoWロールバック）、分析設定UPSERT（単一active維持）、テーマ重複拒否・並び替え。
- usecase: ユーザー停止/復活/削除（admin自身は対象外）、招待コード重複回避。

## 10. 実装タスク分解
- [ ] admin/invitation/analysis/theme/candidate の各 usecase・repository
- [ ] 候補承認を UoW で実装（watchlist連動・上限3）
- [ ] 分析設定 UPSERT（single active）
- [ ] テーマ CRUD＋sort一括更新
- [ ] controller/router（adminグループ）
- [ ] 各 `app/admin/**/page.tsx`（D&D・ダイアログ・Snackbar・クリップボード）

## 11. 受け入れ条件
- 管理者がユーザー停止/復活/削除、招待コード発行/無効化、分析設定保存、テーマCRUD/並び替え、候補承認/却下を行える。
- 候補承認で `watchlist` が正しく置き換わり（最大3維持）、候補 status が更新される。
- 分析設定保存後、activeレコードが常に1件に保たれる。
- admin以外は全 `/admin/*` にアクセスできない。

## 12. 未確定事項
- なし（候補承認時の `watchlist.mode` は承認ダイアログで選択・既定 `both`、SCR-09は共通ポートフォリオ成績サマリー＋ユーザー一覧（成績列なし）、と確定）。
