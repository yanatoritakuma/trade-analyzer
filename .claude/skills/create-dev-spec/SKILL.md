---
name: create-dev-spec
description: >-
  AI株式トレーディングシステムの「実装仕様書（dev-spec）」を画面/機能ごとに生成する。
  doc/feature-spec/ の画面仕様（feature_NN_*.md）を入力に、DDD構成のGoバックエンド・
  Next.jsフロント・DBアクセス・API(openapi断片)・テスト観点・実装タスクまで落とし込んだ
  実装手順書を doc/dev-spec/ 配下に作成する。
  「dev-spec作って」「実装仕様書を作成」「feature_03 を実装仕様に落として」「この画面の
  開発仕様を起こして」「○○画面の実装手順をまとめて」のように、特定の画面/機能を実装に
  落とし込む依頼が来たら必ずこのスキルを使うこと。openapi.yamlの単体作成だけの依頼や、
  feature-spec自体のレビュー・修正依頼には使わない。
---

# create-dev-spec — 実装仕様書ジェネレータ

## このスキルの目的

`doc/feature-spec/` の**画面仕様（feature_NN_*.md）**は「何を作るか（UI・API・挙動）」を
定義しているが、実装者が手を動かすには「**どのファイルに何を書くか**」が足りない。
このスキルは1画面/1機能分を、本リポジトリの設計規約（DDD + Next.js App Router）に沿った
**実装手順書**へ変換し、`doc/dev-spec/` に保存する。

入力単位は**画面/機能ごと**（例: 「feature_03 ダッシュボードのdev-specを作って」）。

## 必ず最初に読むファイル（コンテキスト確定）

仕様は更新され続けるため、**記憶ではなく必ず実ファイルを読む**こと。最低限：

1. 対象の画面仕様 … `doc/feature-spec/feature_NN_*.md`（ユーザー指定の画面）
2. 全体仕様 … `doc/feature-spec/ai_trading_system_spec.md`
   （該当機能の認証・API・分析/通知/学習フロー・売買記録ロジックの章だけでよい）
3. DB定義 … `doc/dev-spec/db_definition.md`（使うテーブル・カラム・制約・データ分離方針）
4. 開発規約 … `doc/dev-spec/development_manual.md`
   （DDD層構成・ディレクトリ・UoW・命名規則・エラー規約・apiClient・middleware）

対象画面が他画面とAPIを共有する場合（例: `/api/portfolio/summary` をダッシュボードと
ポートフォリオが共用）は、関連する devspec も読んで齟齬が出ないようにする。

## 横断ルール（生成物は必ずこれに従う）

feature-spec の最新方針。実装仕様がこれと矛盾していたら**実装仕様側を直す**こと。

- **権限**: 一般ユーザー(`user`)は**閲覧(GET)のみ**。作成・更新・削除は全て **admin専用**
  （`/api/admin/*` ＋ `RequireAdmin`）。画面も書き込みUIは admin のみ表示。
- **認証**: JWTのアクセス＋リフレッシュ**両トークン**をHttpOnly Cookie（`SameSite=Strict`、
  `Secure`は本番のみ）。フロントのログイン/登録パスは `/login`・`/register`。
- **データ分離**: ユーザーごと（`trades`・`real_positions`）は全クエリに `WHERE user_id = ?`。
  共通（`watchlist`・`stock_prices`・`analysis_logs`・`learning_*`・`analysis_*`）は付けない。
  バーチャルtradesの記録先は**adminのみ**（全ユーザー共通の単一ポートフォリオ）。
- **株価**: `stock_prices` は1銘柄1行、`ticker` キーで **UPSERT**。現在値=`close`、前日比=
  `change_amount`/`change_rate` を保持。含み益は参照時に `close − 取得単価` で算出。
- **スケジュール**: 定期分析は**平日15:30のみ**、yfinanceは**過去120日**取得。
- **トランザクション**: 複数リポジトリ更新は **UoW (`uow.Do`)** 内。外部API(Claude/LINE/S3)は
  UoWの**外**。
- **DB操作**: GORM。マイグレーションは `AutoMigrate`。repositoryは**ドメインエンティティ**を返す
  （GORMモデルを直接返さない）。
- **フロント**: fetch標準API（axios不可）、`utils/apiClient.ts` 経由。フォームは
  React Hook Form + Zod。汎用UIは `components/elements/` 配下。

矛盾・未定義に気づいたら、勝手に作話せず**「未確定事項」節に明記**して実装者に判断を委ねる。

## 進め方

1. ユーザーが指定した画面を特定する（IDが曖昧なら `doc/feature-spec/` を一覧して確認）。
2. 上記「必ず読むファイル」を読む。
3. 下の出力テンプレートに沿って、該当画面で**実際に必要な節だけ**を埋める
   （例: 認証画面ならClaude/LINE節は不要、参照専用画面ならバリデーション節は薄くてよい）。
4. `doc/dev-spec/` に保存する。命名は **`dev_spec_NN_<画面名>.md`**
   （feature-spec側の `feature_NN_*.md` と紛れないよう接頭辞 `dev_spec_` を使う）。
   例: `doc/dev-spec/dev_spec_03_dashboard.md`
5. 最後に「[品質チェックリスト](#品質チェックリスト)」を自分で通し、要約を返す。

## 出力テンプレート

画面に関係する節のみ残す。各節は「**どのファイルに何を書くか**」が分かる粒度で書くこと。

```markdown
# 実装仕様 - <画面名>

**画面ID**: SCR-NN
**元仕様**: doc/feature-spec/feature_NN_*.md
**パス(フロント)**: /xxx
**対象ロール**: user / admin（書き込みは admin のみ 等）
**作成日**: <YYYY-MM-DD>
**バージョン**: 1.0

## 1. 概要・スコープ
- この画面で実装する範囲を箇条書き。前提・依存する他機能。

## 2. 権限・アクセス制御
- 画面アクセス可否（middleware.ts の判定・リダイレクト先）
- API側の認可（protected GET / admin専用 のどちらか・RequireAdmin有無）

## 3. データモデル / DBアクセス
- 使用テーブルとカラム（db_definition.md の何を使うか）
- データ分離: WHERE user_id 要否、共通テーブルか、UPSERT/集計の有無
- 代表クエリ（擬似SQLまたはGORM）。集計（勝率・損益・前日比など）はロジックも記す

## 4. API仕様（openapi断片）
- エンドポイント・メソッド・認可区分
- リクエスト/レスポンスのJSONスキーマ（型・必須・例）
- ステータスコード（200/201/400/401/403/404/500）と意味
- openapi.yaml に追記すべき paths/components のドラフト

## 5. バックエンド実装（DDD層別）
各層で **新規/変更するファイル** と責務を表で。development_manual.md のディレクトリに従う。
| 層 | ファイル | 追加/変更 | 責務 |
|----|---------|----------|------|
| domain | domain/xxx/*.go | … | エンティティ・値オブジェクト・repository IF |
| usecase | usecase/xxx_usecase.go | … | 業務フロー（UoW境界・外部API分離） |
| repository | repository/xxx_repository.go | … | GORM実装・ドメイン変換 |
| model | model/xxx.go | … | GORMモデル（gorm tag） |
| controller | controller/xxx_controller.go | … | バインド・HandleError・Cookie |
| router | router/router.go | … | ルート登録（protected / admin） |
- UoWを使う処理は `uow.Do` 内に入れる操作を明記。外部API呼び出しの位置（トランザクション外）も。

## 6. フロントエンド実装（Next.js）
- ページ: app/xxx/page.tsx（RSC/Client の別・データ取得方法）
- コンポーネント: components/elements/ 配下に作る汎用UI、画面固有コンポーネント
- API呼び出し: utils/apiClient.ts の get/post/... と型（types/api.ts 由来）
- フォーム: React Hook Form + Zod スキーマ（バリデーション項目）
- 状態: ローディング/エラー/空表示（Skeleton・Snackbar 等）
- 権限分岐: useAuth() の role による出し分け

## 7. バリデーション
- フロント（Zod）とバック（validator/値オブジェクト）の両方。エラーメッセージは feature-spec準拠。

## 8. 外部連携（該当時のみ）
- Claude API / LINE通知 / S3。プロンプト・通知フォーマット・呼び出し条件。

## 9. テスト観点
- domain: 値オブジェクト・エンティティのバリデーション
- usecase: リポジトリ/UoWをモックにした正常系・異常系
- repository: Neonテストブランチでの結合（必要時）

## 10. 実装タスク分解（チェックリスト）
- [ ] … 着手順に並べた具体タスク

## 11. 受け入れ条件
- 機能が満たすべき観測可能な条件（feature-specの挙動表に対応）

## 12. 未確定事項
- feature-spec/DB定義で未定義・要判断の点（あれば）。なければ「なし」。
```

## 品質チェックリスト（保存前に自己点検）

- [ ] feature-spec の挙動表・バリデーション・API・遷移が、実装仕様に漏れなく対応している
- [ ] 「横断ルール」と矛盾していない（権限/認証/データ分離/15:30/120日/UPSERT 等）
- [ ] バックの各層で「新規/変更ファイル」が具体パスで示されている
- [ ] APIの認可区分（protected GET か admin 専用か）が明示されている
- [ ] フロントのページ・コンポーネント・apiClient・Zodスキーマが具体的
- [ ] 使用テーブルが db_definition.md に実在し、データ分離（user_id要否）が正しい
- [ ] 共有APIを持つ他画面との齟齬がない
- [ ] 推測で埋めた箇所は「未確定事項」に明記している
- [ ] ファイル名が `doc/dev-spec/dev_spec_NN_<画面名>.md` 規約に従っている

## 参考: 画面ID対応表

| ID | 画面 | feature-spec |
|----|------|--------------|
| SCR-01 | ログイン | feature_01_login.md |
| SCR-02 | 新規登録 | feature_02_register.md |
| SCR-03 | ダッシュボード | feature_03_dashboard.md |
| SCR-04 | ウォッチリスト | feature_04_watchlist.md |
| SCR-05 | トレード履歴 | feature_05_trades.md |
| SCR-06 | ポートフォリオ | feature_06_portfolio.md |
| SCR-07 | 週次レポート | feature_07_reports.md |
| SCR-08 | 設定 | feature_08_settings.md |
| SCR-09〜13 | 管理者画面群 | feature_09_12_admin.md |
