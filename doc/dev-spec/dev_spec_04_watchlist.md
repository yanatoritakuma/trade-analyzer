# 実装仕様 - ウォッチリスト

**画面ID**: SCR-04
**元仕様**: doc/feature-spec/feature_04_watchlist.md
**パス(フロント)**: `/watchlist`（閲覧・全員） / `/admin/watchlist`（管理・admin）
**対象ロール**: 閲覧=user/admin、追加・削除=**admin のみ**
**作成日**: 2026-06-27
**バージョン**: 1.0

## 1. 概要・スコープ
- 全ユーザー共通のウォッチリスト（最大3銘柄）を表示。追加・削除は管理者のみ。
- 銘柄コードは4桁数字入力 → 末尾 `.T` を付与して保存。
- 依存: `watchlist`・`stock_prices`（現在値/前日比表示）。

## 2. 権限・アクセス制御
- `/watchlist`: 認証必須・user/admin閲覧可（書き込みUIなし）。
- `/admin/watchlist`: middlewareでadmin判定（`/admin/*`）。
- API: GETは `protected`、POST/DELETEは `/api/admin/*`＋`RequireAdmin`。

## 3. データモデル / DBアクセス
- `watchlist`: `ticker` UNIQUE・`mode`(`virtual`/`real`/`both`)・`is_active`。
- 表示時は `stock_prices` を `ticker` でLEFT JOINし現在値/前日比を付与。
- 制約（アプリ層）: `is_active=true` の件数が **3未満のときのみ追加可**。`ticker` 重複不可。
- データ分離: 共通テーブル（user_id無し）。

## 4. API仕様（openapi断片）

| メソッド | パス | 認可 | 説明 |
|---------|------|------|------|
| GET | `/api/watchlist` | protected | ウォッチリスト取得（現在値付き） |
| POST | `/api/admin/watchlist` | admin | 銘柄追加 |
| DELETE | `/api/admin/watchlist/:id` | admin | 銘柄削除 |

```yaml
components:
  schemas:
    WatchlistCreateRequest:
      type: object
      required: [code, mode]
      properties:
        code: { type: string, pattern: '^[0-9]{4}$' }   # 4桁。サーバで `.T` 付与
        mode: { type: string, enum: [virtual, real, both] }
```

- エラー: 「最大3銘柄まで登録できます」(400) / 「すでに登録されている銘柄です」(400)。

## 5. バックエンド実装（DDD層別）

| 層 | ファイル | 追加/変更 | 責務 |
|----|---------|----------|------|
| domain | `domain/watchlist/watchlist.go` | 既存 | エンティティ・`mode` VO・`.T`付与は usecase |
| domain | `domain/watchlist/repository.go` | 変更 | `FindAll`・`CountActive`・`ExistsByTicker`・`Save`・`Delete` |
| usecase | `usecase/watchlist_usecase.go` | 変更 | `Create`（上限3・重複チェック）・`Delete`・`FindAllWithPrice` |
| repository | `repository/watchlist_repository.go` | 変更 | 上記クエリ実装 |
| model | `model/watchlist.go` | 既存 | GORMモデル |
| controller | `controller/watchlist_controller.go` | 変更 | `GetAll`（protected）・`Create`/`Delete`（admin） |
| router | `router/router.go` | 変更 | GETはprotected、POST/DELETEはadminグループ |

**usecase Create**: `CountActive() >= 3` → `ErrInvalidInput("最大3銘柄まで…")`／`ExistsByTicker(code+".T")` → `ErrAlreadyExists`／OKなら `Save`。単一テーブルのみなのでUoW不要（必要なら使用可）。

## 6. フロントエンド実装（Next.js）
- 閲覧: `app/watchlist/page.tsx`。`apiClient.get('/api/watchlist')`。カード一覧（現在値・前日比・モード）。
  「銘柄の追加・変更は管理者が行います」を注記。
- 管理: `app/admin/watchlist/page.tsx`。`[+ 銘柄を追加]` ＋ 各行 `[削除]`。
- 追加ダイアログ: `components/elements/modalBox` ＋ RHF+Zod。
  ```ts
  const addSchema = z.object({
    code: z.string().regex(/^[0-9]{4}$/, '銘柄コードは4桁の数字で入力してください'),
    mode: z.enum(['virtual','real','both'], { message: 'モードを選択してください' }),
  });
  ```
- 追加成功/削除後は一覧再取得。削除は確認ダイアログ。
- role分岐: `const isAdmin = useAuth().user?.role === 'admin'`（閲覧画面では追加/削除UIを出さない）。

## 7. バリデーション
| フィールド | フロント | バック |
|-----------|---------|--------|
| code | 4桁数字 | 4桁数字→`.T`付与・重複 |
| mode | enum必須 | `mode` CHECK |
| 件数上限 | （送信前に警告可） | `CountActive < 3` を必須 |

## 8. 外部連携
- なし（Lambda は `/internal/watchlist` で別途このテーブルを参照）。

## 9. テスト観点
- usecase: `Create` 上限到達（3件で拒否）/ 重複拒否 / 正常追加 / `Delete`。
- repository: `FindAllWithPrice` のJOIN結果（stock_prices無し銘柄はnull）。

## 10. 実装タスク分解
- [ ] repository: `CountActive`/`ExistsByTicker`/`FindAllWithPrice`
- [ ] usecase `Create`/`Delete`/`FindAllWithPrice`
- [ ] controller/router（GET=protected, POST/DELETE=admin）
- [ ] `app/watchlist/page.tsx`（閲覧）
- [ ] `app/admin/watchlist/page.tsx`（管理・ダイアログ）
- [ ] usecaseユニットテスト

## 11. 受け入れ条件
- 全ユーザーが同一のウォッチリストを閲覧でき、現在値・前日比が出る。
- 管理者のみ追加（最大3・重複不可・`.T`付与）・削除ができる。
- 一般ユーザーに追加/削除UIが表示されない。

## 12. 未確定事項
- なし。
