# 実装仕様 - ログイン画面

**画面ID**: SCR-01
**元仕様**: doc/feature-spec/feature_01_login.md
**パス(フロント)**: `/login`
**対象ロール**: 未認証の全員（admin / user 共通）
**作成日**: 2026-06-27
**バージョン**: 1.0

## 1. 概要・スコープ
- メール＋パスワードでログインし、JWT（アクセス＋リフレッシュ）をHttpOnly Cookieで受け取る。
- 成功時は role に応じてリダイレクト（user→`/`、admin→`/admin`）。
- ログイン済みで `/login` に来た場合は `/` へ戻す（middleware）。
- 依存: `users` テーブル、JWT/bcrypt/cookieユーティリティ。

## 2. 権限・アクセス制御
- 画面: 公開（`middleware.ts` の `publicPaths` に `/login`）。token保持時は `/` へリダイレクト。
- API: `POST /api/auth/login` は**認証不要**グループ。

## 3. データモデル / DBアクセス
- 使用テーブル: `users`（`email`・`password_hash`・`role`・`is_active`）。
- クエリ: `SELECT * FROM users WHERE email = ?`（`idx_users_email` 利用）。
- 認可判定: `is_active = false` ならログイン不可（`ErrAccountDisabled`）。
- データ分離: 不要（自分の認証情報のみ）。

## 4. API仕様（openapi断片）

`POST /api/auth/login` — 認証不要

```yaml
paths:
  /api/auth/login:
    post:
      requestBody:
        required: true
        content: { application/json: { schema: { $ref: '#/components/schemas/LoginRequest' } } }
      responses:
        '200':
          description: ログイン成功（Set-Cookieでaccess_token/refresh_tokenを返す）
          content: { application/json: { schema: { $ref: '#/components/schemas/LoginResponse' } } }
        '400': { description: リクエスト不正 }
        '401': { description: メール/パスワード不一致 }
        '403': { description: アカウント停止中 }
components:
  schemas:
    LoginRequest:
      type: object
      required: [email, password]
      properties:
        email: { type: string, format: email }
        password: { type: string }
    LoginResponse:
      type: object
      properties:
        message: { type: string }
        user:
          type: object
          properties:
            id: { type: integer }
            email: { type: string }
            name: { type: string }
            role: { type: string, enum: [admin, user] }
```

- 認証失敗メッセージは「メールアドレスまたはパスワードが正しくありません」で統一（どちらが誤りか特定させない）。
- トークンはレスポンスボディに含めず**Cookieのみ**。

## 5. バックエンド実装（DDD層別）

| 層 | ファイル | 追加/変更 | 責務 |
|----|---------|----------|------|
| domain | `domain/user/user.go` | 既存 | `CanLogin()`（is_active判定）・`IsAdmin()` |
| domain | `domain/user/value_object.go` | 既存 | `Email` VO |
| domain | `domain/user/repository.go` | 既存 | `FindByEmail` |
| domain | `domain/errors.go` | 変更 | `ErrUnauthorized`・`ErrAccountDisabled` |
| usecase | `usecase/user_usecase.go` | 変更 | `Login(ctx, email, pw) (user, accessToken, refreshToken, error)` |
| repository | `repository/user_repository.go` | 既存 | `FindByEmail`（ドメイン変換） |
| utils | `utils/password.go` | 既存 | bcrypt照合 `CompareHashAndPassword` |
| utils | `utils/jwt.go` | 変更 | `GenerateAccessToken`(24h)・`GenerateRefreshToken`(30d) |
| utils | `utils/cookie.go` | 新規 | `SetAuthCookies`（development_manual 8b.2参照） |
| controller | `controller/user_controller.go` | 変更 | `Login`：bind→usecase→`SetAuthCookies`→200 |
| router | `router/router.go` | 変更 | `router.POST("/api/auth/login", ...)`（認証不要） |

**usecase Login の流れ**（UoW不要・参照のみ）:
1. `FindByEmail` → 見つからなければ `ErrUnauthorized`
2. `CanLogin()` が false → `ErrAccountDisabled`
3. bcrypt照合失敗 → `ErrUnauthorized`
4. アクセス/リフレッシュトークン生成して返す

## 6. フロントエンド実装（Next.js）
- ページ: `app/login/page.tsx`（`'use client'`）。
- フォーム: React Hook Form + Zod。
  ```ts
  const loginSchema = z.object({
    email: z.string().min(1, 'メールアドレスを入力してください').email('正しいメールアドレスを入力してください'),
    password: z.string().min(1, 'パスワードを入力してください'),
  });
  ```
- API: `apiClient.post<LoginResponse>('/api/auth/login', values)`。
- 成功: `result.user.role === 'admin' ? router.push('/admin') : router.push('/')`。
- パスワード表示トグル（`@mui/icons-material` の Visibility）。
- ローディング: 送信中はボタンを「ログイン中...」＋disabled（`components/elements/buttonBox/LoadingButton`）。
- エラー: `catch` で `error.message` をフォーム上部に表示。停止アカウントは403メッセージ表示。

## 7. バリデーション
| フィールド | フロント(Zod) | バック |
|-----------|--------------|--------|
| email | 必須・メール形式 | `Email` VO（usecaseでは形式チェック省略可、認証失敗に集約） |
| password | 必須・1文字以上 | bcrypt照合のみ（複雑性チェックはしない） |

## 8. 外部連携
- なし。

## 9. テスト観点
- domain: `NewEmail` 正常/異常。
- usecase: `Login` 成功 / メール無し（Unauthorized）/ パスワード不一致（Unauthorized）/ 停止中（AccountDisabled）。リポジトリ・bcryptはモック。
- controller(任意): Cookieが2本セットされること。

## 10. 実装タスク分解
- [ ] `utils/cookie.go`（SetAuthCookies/ClearAuthCookies）
- [ ] `utils/jwt.go` にリフレッシュトークン生成を追加
- [ ] `user_usecase.Login` を access/refresh 返却に変更
- [ ] `user_controller.Login` で両Cookieセット
- [ ] router 登録
- [ ] `app/login/page.tsx` 実装（RHF+Zod・role分岐リダイレクト）
- [ ] usecaseユニットテスト

## 11. 受け入れ条件
- 正しい資格情報で user は `/`、admin は `/admin` へ遷移する。
- Cookieに `access_token`・`refresh_token`（HttpOnly/SameSite=Strict）がセットされる。
- 誤資格情報・停止アカウントで適切なメッセージが表示され、トークンは発行されない。
- 二重送信が防止される。

## 12. 未確定事項
- なし。
