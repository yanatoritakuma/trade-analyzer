# 実装仕様 - 新規登録画面

**画面ID**: SCR-02
**元仕様**: doc/feature-spec/feature_02_register.md
**パス(フロント)**: `/register`
**対象ロール**: 未認証ユーザー（招待コード保持者）
**作成日**: 2026-06-27
**バージョン**: 1.0

## 1. 概要・スコープ
- 招待コード＋名前＋メール＋パスワードでアカウント作成（role は常に `user`）。
- 招待コードの有効性検証と「使用済みマーク」を**同一トランザクション**で行う。
- 成功時は `/login` へ遷移し成功メッセージを表示。
- 依存: `users`・`invitation_codes`。

## 2. 権限・アクセス制御
- 画面: 公開（`publicPaths` に `/register`）。token保持時は `/` へリダイレクト。
- API: `POST /api/auth/register` は**認証不要**。

## 3. データモデル / DBアクセス
- `invitation_codes`: `code`・`expires_at`・`used_by`・`is_active`。
  - 有効判定: `is_active = true` かつ `used_by IS NULL` かつ `expires_at >= NOW()`。
- `users`: `email` UNIQUE。重複時は `ErrAlreadyExists`。
- データ分離: 不要。

## 4. API仕様（openapi断片）

`POST /api/auth/register` — 認証不要

```yaml
paths:
  /api/auth/register:
    post:
      requestBody:
        required: true
        content: { application/json: { schema: { $ref: '#/components/schemas/RegisterRequest' } } }
      responses:
        '201': { description: 作成成功, content: { application/json: { schema: { $ref: '#/components/schemas/LoginResponse' } } } }
        '400': { description: 招待コード無効/期限切れ/使用済み・メール重複・入力不正 }
components:
  schemas:
    RegisterRequest:
      type: object
      required: [invitation_code, name, email, password]
      properties:
        invitation_code: { type: string }
        name: { type: string, maxLength: 50 }
        email: { type: string, format: email }
        password: { type: string, minLength: 8 }
```

- エラーメッセージ: 「無効な招待コードです」「招待コードの有効期限が切れています」「この招待コードはすでに使用されています」「このメールアドレスはすでに使用されています」を出し分ける。

## 5. バックエンド実装（DDD層別）

| 層 | ファイル | 追加/変更 | 責務 |
|----|---------|----------|------|
| domain | `domain/user/user.go` | 既存 | `NewUser`（role=user固定・VO検証） |
| domain | `domain/invitation/invitation_code.go` | 既存 | `IsValid()`・`MarkAsUsed(userID)` |
| domain | `domain/invitation/repository.go` | 既存 | `FindByCode`・`Update` |
| domain | `domain/errors.go` | 変更 | `ErrInvalidCode`・`ErrExpiredCode`・`ErrAlreadyExists` |
| usecase | `usecase/user_usecase.go` | 変更 | `Register(ctx, code, email, pw, name)` を **UoW** で実装 |
| usecase | `usecase/uow.go` | 既存 | `Repositories{User, InvitationCode}` |
| repository | `repository/uow_impl.go` | 既存 | tx内リポジトリ束ね |
| controller | `controller/user_controller.go` | 変更 | `Register`：bind→usecase→201 |
| router | `router/router.go` | 変更 | `POST /api/auth/register`（認証不要） |

**usecase Register（development_manual 4.4の例に準拠・UoW内）**:
1. `InvitationCode.FindByCode(code)` → 無ければ `ErrInvalidCode`
2. `IsValid()` false → 期限切れ/使用済みを区別して返す
3. `User.FindByEmail` で重複チェック → 重複なら `ErrAlreadyExists`
4. `user.NewUser(...)` → `User.Save`
5. `invitation.MarkAsUsed(newUser.ID)` → `InvitationCode.Update`
6. fn が nil を返せば両方コミット（どれか失敗で全ロールバック）

> 登録時は自動ログインしない（feature_02準拠）。トークン発行なし。

## 6. フロントエンド実装（Next.js）
- ページ: `app/register/page.tsx`（`'use client'`）。
- Zodスキーマ（feature_02実装メモ準拠）:
  ```ts
  const registerSchema = z.object({
    invitationCode: z.string().min(1, '招待コードを入力してください'),
    name: z.string().min(1, 'お名前を入力してください').max(50, 'お名前は50文字以内で入力してください'),
    email: z.string().min(1, 'メールアドレスを入力してください').email('正しいメールアドレスを入力してください'),
    password: z.string().min(8, 'パスワードは8文字以上で入力してください')
      .regex(/^(?=.*[a-zA-Z])(?=.*[0-9])/, 'パスワードは英字と数字を含めてください'),
    passwordConfirm: z.string().min(1, 'パスワード（確認）を入力してください'),
  }).refine(d => d.password === d.passwordConfirm, { message: 'パスワードが一致しません', path: ['passwordConfirm'] });
  ```
- 招待コード入力は大文字変換・オートフォーカス。
- 送信: `apiClient.post('/api/auth/register', { invitation_code, name, email, password })`。
- 成功: `router.push('/login?registered=1')` で `/login` 側に成功メッセージ表示。

## 7. バリデーション
- フロント: 上記Zod。バック: `domain/user` の各VO（`Name`最大50・`Password`最小8）＋招待コード/メール重複は usecase。

## 8. 外部連携
- なし。

## 9. テスト観点
- usecase: 招待コード無効 / 期限切れ / 使用済み / メール重複 / 正常（UoWモックで `User.Save` と `InvitationCode.Update` 両方呼ばれる）。
- domain: `InvitationCode.IsValid()`・`NewUser` のVO検証。

## 10. 実装タスク分解
- [ ] `domain/invitation` の `IsValid`/`MarkAsUsed` 実装
- [ ] `user_usecase.Register` を UoW で実装
- [ ] `user_controller.Register` ＋ router 登録
- [ ] `app/register/page.tsx`（RHF+Zod・成功遷移）
- [ ] `/login` の成功メッセージ表示（クエリ `registered=1`）
- [ ] usecaseユニットテスト

## 11. 受け入れ条件
- 有効な招待コードで登録でき、コードが使用済みになる。
- 無効/期限切れ/使用済みコード・メール重複・パスワード不一致で適切なメッセージ。
- 登録途中でDB失敗時、ユーザー作成と招待コード更新が両方ロールバックされる。

## 12. 未確定事項
- なし。
