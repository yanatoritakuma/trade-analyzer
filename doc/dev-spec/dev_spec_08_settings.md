# 実装仕様 - 設定

**画面ID**: SCR-08
**元仕様**: doc/feature-spec/feature_08_settings.md
**パス(フロント)**: `/settings`
**対象ロール**: **admin のみ**（書き込み画面のため。一般ユーザーはアクセス不可）
**作成日**: 2026-06-27
**バージョン**: 1.0

## 1. 概要・スコープ
- 管理者のプロフィール更新・パスワード変更・実運用保有株（`real_positions`）のCRUD。
- 一般ユーザーは閲覧のみ方針のため本画面に入れない（middlewareで `/settings` をadmin判定）。
- 依存: `users`・`real_positions`・`stock_prices`（保有株の現在値表示は任意）。

## 2. 権限・アクセス制御
- 画面: `middleware.ts` の `adminPaths` に `/settings` を含め、admin以外は `/` へリダイレクト。
- API: 取得系（`/api/auth/me`・`/api/positions`）はprotected、更新系はすべて `/api/admin/*`＋`RequireAdmin`。

## 3. データモデル / DBアクセス
- `users`: `name`（更新可）・`email`（表示のみ・変更不可）・`password_hash`。
- `real_positions`: `(user_id, ticker)` UNIQUE・`quantity`・`avg_price`。**adminのuser_idで操作**。
- パスワード変更: 現在PWをbcrypt照合 → 新PWを再ハッシュ保存。

## 4. API仕様（openapi断片）

| メソッド | パス | 認可 | 説明 |
|---------|------|------|------|
| GET | `/api/auth/me` | protected | プロフィール取得 |
| PATCH | `/api/admin/me` | admin | 名前更新 |
| PUT | `/api/admin/me/password` | admin | パスワード変更 |
| GET | `/api/positions` | protected | 保有株一覧 |
| POST | `/api/admin/positions` | admin | 保有株登録 |
| PUT | `/api/admin/positions/:id` | admin | 保有株更新 |
| DELETE | `/api/admin/positions/:id` | admin | 保有株削除 |

```yaml
components:
  schemas:
    ProfileUpdateRequest:
      type: object
      required: [name]
      properties: { name: { type: string, maxLength: 50 } }
    PasswordChangeRequest:
      type: object
      required: [current_password, new_password]
      properties:
        current_password: { type: string }
        new_password: { type: string, minLength: 8 }
    PositionRequest:
      type: object
      required: [code, avg_price, quantity]
      properties:
        code: { type: string, pattern: '^[0-9]{4}$' }  # `.T`付与
        avg_price: { type: number, minimum: 1 }
        quantity: { type: integer, minimum: 1 }
```

- パスワード現在値誤り → 「現在のパスワードが正しくありません」(400)。

## 5. バックエンド実装（DDD層別）

| 層 | ファイル | 追加/変更 | 責務 |
|----|---------|----------|------|
| domain | `domain/user/user.go` | 変更 | `ChangeName`・`ChangePassword`（VO検証） |
| domain | `domain/position/position.go` | 既存 | エンティティ |
| domain | `domain/position/repository.go` | 変更 | `FindByUser`・`Save`・`Update`・`Delete` |
| usecase | `usecase/user_usecase.go` | 変更 | `UpdateProfile`・`ChangePassword`（現PW照合） |
| usecase | `usecase/position_usecase.go` | 変更 | `Create`/`Update`/`Delete`/`List`（admin user_id付与） |
| repository | `repository/position_repository.go` | 変更 | CRUD（`WHERE user_id = ?` 必須） |
| utils | `utils/password.go` | 既存 | bcrypt生成/照合 |
| controller | `controller/user_controller.go` | 変更 | `Me`・`UpdateProfile`・`ChangePassword` |
| controller | `controller/position_controller.go` | 変更 | CRUD |
| router | `router/router.go` | 変更 | GET=protected、更新系=admin |

- `real_positions` はユーザーごとテーブル → 全クエリ `WHERE user_id = ?`（adminのID）を必須。
- 単一テーブル操作のためUoWは任意（パスワード変更も単一）。

## 6. フロントエンド実装（Next.js）
- ページ: `app/settings/page.tsx`（Client・admin前提）。3セクション（アコーディオン）。
- プロフィール: 名前（編集可）・メール（表示のみ）。RHF+Zod（name必須・最大50）。`PATCH /api/admin/me`。
- パスワード変更: 現在/新規/確認。Zod（新規8文字以上・英数字・確認一致）。`PUT /api/admin/me/password`。成功でフォームリセット。
- 保有株: 一覧（`GET /api/positions`）＋ [編集][削除] ＋ [+ 保有株を追加]。ダイアログでCRUD。
  ```ts
  const positionSchema = z.object({
    code: z.string().regex(/^[0-9]{4}$/, '銘柄コードは4桁の数字で入力してください'),
    avgPrice: z.coerce.number().min(1, '1以上の数値を入力してください'),
    quantity: z.coerce.number().int().min(1, '1以上の整数を入力してください'),
  });
  ```
- 保存成功/失敗は Snackbar（緑/赤）。削除は確認ダイアログ。

## 7. バリデーション
| フィールド | フロント | バック |
|-----------|---------|--------|
| name | 必須・最大50 | `Name` VO |
| current/new password | 必須・新8文字英数字・確認一致 | 現PW照合・`Password` VO |
| code/avg_price/quantity | 4桁/1以上/1以上整数 | 同等＋`.T`付与 |

## 8. 外部連携
- なし。

## 9. テスト観点
- usecase: `ChangePassword` 現PW不一致で拒否・正常変更。`Position` CRUD（user_id付与・UNIQUE衝突）。
- domain: `Name`/`Password` VO。

## 10. 実装タスク分解
- [ ] user_usecase `UpdateProfile`/`ChangePassword`
- [ ] position_usecase CRUD（admin user_id）
- [ ] controller/router（GET=protected・更新=admin）
- [ ] middleware に `/settings` を admin判定追加（development_manual反映済を確認）
- [ ] `app/settings/page.tsx`（3セクション・ダイアログ・Snackbar）

## 11. 受け入れ条件
- 管理者が名前変更・パスワード変更・保有株CRUDを行える。
- 現在PW誤りで変更が拒否される。
- 一般ユーザーは `/settings` にアクセスすると `/` にリダイレクトされる。

## 12. 未確定事項
- なし（設定画面の保有株一覧は feature_08 準拠で**単価・数量のみ**表示。現在値・含み益はポートフォリオ画面=SCR-06で表示、と確定）。
