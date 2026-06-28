# AI株式トレーディングシステム 開発マニュアル

**バージョン**: 4.1  
**作成日**: 2026-06-06  
**更新内容**: ①フロントのログイン/登録パスを `/login`・`/register` に統一（middleware・ディレクトリ） ②JWTのアクセス＋リフレッシュ両トークンをHttpOnly Cookieで発行する実装例（SameSite=Strict・Secure環境切替・refresh/logout）を追加 ③`/settings` をadmin専用に ④`APP_ENV` を追加。前版で `stock_prices`（最新株価スナップショット・`ticker`でUPSERT）を追加  
**参考リポジトリ**: https://github.com/yanatoritakuma/budget  
**ステータス**: Draft

---

## 目次

1. [プロジェクト構成](#1-プロジェクト構成)
2. [技術スタック・ライブラリ一覧](#2-技術スタックライブラリ一覧)
3. [フロントエンド開発ガイド（Next.js）](#3-フロントエンド開発ガイド)
4. [バックエンド開発ガイド（Go / DDD）](#4-バックエンド開発ガイド)
5. [株価取得開発ガイド（Python / Lambda）](#5-株価取得開発ガイド)
6. [データベースガイド（Neon PostgreSQL）](#6-データベースガイド)
7. [環境変数管理](#7-環境変数管理)
8. [開発フロー・命名規則](#8-開発フロー命名規則)
8a. [テスト方針](#8a-テスト方針)
8b. [エラーハンドリング規則](#8b-エラーハンドリング規則)
8c. [CORS設定](#8c-cors設定)
8d. [.gitignore](#8d-gitignore)
9. [ディレクトリ構成](#9-ディレクトリ構成)
10. [ローカル開発環境セットアップ](#10-ローカル開発環境セットアップ)
11. [CI/CD・自動デプロイ](#11-cicd自動デプロイ)

---

## 1. プロジェクト構成

budgetリポジトリと同様に `front/` と `back/` を分けたモノレポ構成。

```
trading-system/
  ├── front/           # Next.js（TypeScript）
  ├── back/            # Go / Gin（DDD設計）
  ├── lambda/          # Python（株価取得専用）
  └── docs/            # 仕様書・devspec
      ├── spec.md
      ├── development_manual.md
      ├── feature_01_login.md
      └── ...
```

---

## 2. 技術スタック・ライブラリ一覧

### 2.1 フロントエンド（Next.js）

budgetリポジトリの `front/package.json` を参考に構成。axiosは使用せず **fetch標準API** を使用する。

| カテゴリ | ライブラリ | バージョン | 用途 |
|---------|-----------|---------|------|
| フレームワーク | Next.js | 15系 | App Router・Turbopack使用 |
| 言語 | TypeScript | 5系 | 型安全な開発 |
| UIコンポーネント | @mui/material | v7系 | UIコンポーネント全般 |
| MUIアイコン | @mui/icons-material | v7系 | アイコン |
| MUIスタイル | @emotion/react / @emotion/styled | v11系 | MUI依存スタイル |
| HTTPクライアント | fetch（標準API） | - | APIリクエスト（axiosは使わない） |
| スタイル | SCSS（sass） | v1系 | グローバルスタイル・モジュール |
| 日付処理 | date-fns | v4系 | 日付フォーマット |
| チャート | Recharts | v2系 | 株価・損益グラフ |
| フォーム管理 | React Hook Form | v7系 | フォーム状態管理 |
| バリデーション | Zod + @hookform/resolvers | v3系 | スキーマ定義・バリデーション |
| API型生成 | openapi-typescript | v7系 | openapi.yamlからTS型を自動生成 |
| コードフォーマット | Prettier | v3系 | コード整形 |
| Linter | ESLint | v9系 | 静的解析 |

### 2.2 バックエンド（Go）

budgetリポジトリの `back/go.mod` と構成を参考に。

| カテゴリ | ライブラリ | 用途 |
|---------|-----------|------|
| フレームワーク | github.com/gin-gonic/gin | HTTPルーティング・ミドルウェア |
| ORM | gorm.io/gorm | DBアクセス |
| GORMドライバ | gorm.io/driver/postgres | PostgreSQL接続 |
| JWT | github.com/golang-jwt/jwt/v5 | JWTトークン発行・検証 |
| パスワード | golang.org/x/crypto/bcrypt | パスワードハッシュ化 |
| 環境変数 | github.com/joho/godotenv | .envファイル読み込み |
| AWS Lambda | github.com/aws/aws-lambda-go | Lambda対応（budgetと同様） |
| Lambda Gin Adapter | github.com/awslabs/aws-lambda-go-api-proxy/gin | GinをLambdaで動かすアダプタ |
| AWS SDK | github.com/aws/aws-sdk-go-v2 | S3操作（学習CSV） |
| バリデーション | github.com/go-playground/validator/v10 | リクエストバリデーション |
| OpenAPI生成 | oapi-codegen | openapi.yamlからサーバコード自動生成 |
| ホットリロード | air（github.com/air-verse/air） | 開発時の自動再起動 |

### 2.3 Python / Lambda

| カテゴリ | ライブラリ | 用途 |
|---------|-----------|------|
| 株価取得 | yfinance | Yahoo Financeから株価取得 |
| HTTPクライアント | requests | GoのAPIエンドポイントにPOST |
| 数値計算 | pandas | データ整形 |
| 環境変数 | os（標準） | 環境変数読み込み |

### 2.4 インフラ

| カテゴリ | サービス | 用途 |
|---------|---------|------|
| DB | Neon（Serverless PostgreSQL） | データ永続化 |
| ストレージ | AWS S3 | 学習CSVのバージョン管理 |
| 定期実行 | AWS EventBridge | Lambda定期起動 |
| サーバーレス | AWS Lambda | Go本体 + Python株価取得 |

---

## 3. フロントエンド開発ガイド

### 3.1 ディレクトリ構成

budgetリポジトリの `front/src/` 構成を踏襲。

```
front/
  ├── public/
  ├── src/
  │   ├── app/                          # App Router ページ
  │   │   ├── layout.tsx                # ルートレイアウト
  │   │   ├── page.tsx                  # ダッシュボード（/）
  │   │   ├── globals.css
  │   │   ├── style.scss
  │   │   ├── login/
  │   │   │   └── page.tsx              # ログイン（/login）
  │   │   ├── register/
  │   │   │   └── page.tsx              # 新規登録（/register）
  │   │   ├── watchlist/
  │   │   │   └── page.tsx
  │   │   ├── trades/
  │   │   │   └── page.tsx
  │   │   ├── portfolio/
  │   │   │   └── page.tsx
  │   │   ├── reports/
  │   │   │   ├── page.tsx
  │   │   │   └── [week]/
  │   │   │       └── page.tsx
  │   │   ├── settings/
  │   │   │   └── page.tsx
  │   │   ├── admin/
  │   │   │   ├── page.tsx
  │   │   │   ├── users/
  │   │   │   │   └── page.tsx
  │   │   │   ├── invitations/
  │   │   │   │   └── page.tsx
  │   │   │   └── analysis-settings/
  │   │   │       └── page.tsx
  │   │   └── api/                      # Next.js Route Handlers（必要に応じて）
  │   ├── components/
  │   │   └── elements/                 # budgetと同様に elements/ 配下に汎用UI
  │   │       ├── buttonBox/            # ボタン系コンポーネント
  │   │       ├── textBox/              # 入力欄系コンポーネント
  │   │       ├── selectBox/            # セレクト系コンポーネント
  │   │       ├── modalBox/             # モーダル系コンポーネント
  │   │       ├── tableBox/             # テーブル系コンポーネント
  │   │       └── chartBox/             # グラフ系コンポーネント
  │   ├── types/
  │   │   └── api.ts                    # openapi-typescriptで自動生成
  │   └── utils/                        # 共通ユーティリティ
  │       ├── apiClient.ts              # fetchベースのAPIクライアント
  │       └── format.ts                 # 日付・金額フォーマット
  ├── .prettierrc
  ├── eslint.config.mjs
  ├── next.config.ts
  ├── tsconfig.json
  └── package.json
```

### 3.2 API型の自動生成

budgetリポジトリと同様に **openapi-typescript** を使用してバックエンドのopenapi.yamlからTypeScript型を自動生成する。

```bash
# back/openapi.yaml から型を生成
npm run generate:api
# → src/types/api.ts が生成される
```

```json
// package.json scripts
{
  "generate:api": "openapi-typescript ../back/openapi.yaml -o src/types/api.ts"
}
```

### 3.3 APIクライアント設定（fetch標準API）

axiosは使用しない。fetch標準APIをラップした共通クライアントを `src/utils/apiClient.ts` に定義する。

```typescript
// src/utils/apiClient.ts
const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080';

type FetchOptions = RequestInit & {
  params?: Record<string, string>;
};

async function request<T>(path: string, options: FetchOptions = {}): Promise<T> {
  const { params, ...init } = options;

  // クエリパラメータの付与
  const url = new URL(`${BASE_URL}${path}`);
  if (params) {
    Object.entries(params).forEach(([k, v]) => url.searchParams.set(k, v));
  }

  const res = await fetch(url.toString(), {
    ...init,
    credentials: 'include',           // HttpOnly Cookieを送受信
    headers: {
      'Content-Type': 'application/json',
      ...init.headers,
    },
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(error.error ?? `HTTP ${res.status}`);
  }

  // 204 No Content など body なしのレスポンスに対応
  const text = await res.text();
  return text ? JSON.parse(text) : ({} as T);
}

export const apiClient = {
  get: <T>(path: string, params?: Record<string, string>) =>
    request<T>(path, { method: 'GET', params }),

  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(body) }),

  put: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),

  patch: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'PATCH', body: JSON.stringify(body) }),

  delete: <T>(path: string) =>
    request<T>(path, { method: 'DELETE' }),
};
```

**使用例：**

```typescript
// クライアントコンポーネントでの使用例
import { apiClient } from '@/utils/apiClient';
import type { paths } from '@/types/api';  // openapi-typescript生成型

type LoginResponse = paths['/api/auth/login']['post']['responses']['200']['content']['application/json'];

const handleLogin = async (data: LoginFormValues) => {
  const result = await apiClient.post<LoginResponse>('/api/auth/login', data);
  // ...
};
```

### 3.4 SCSSの使い方

- グローバルスタイルは `app/globals.css` と `app/style.scss`
- コンポーネント固有スタイルは `*.module.css` または `*.module.scss`

```scss
// app/style.scss（グローバル共通変数など）
$primary-color: #1976d2;
$success-color: #2e7d32;
$error-color:   #c62828;
```

### 3.5 コンポーネント設計方針

budgetの `components/elements/` 構成に倣い、汎用UIコンポーネントを `elements/` に集約する。

```typescript
// components/elements/buttonBox/LoadingButton.tsx
// MUIのButtonをラップしてローディング状態を管理する汎用ボタン
type Props = {
  label: string;
  loading: boolean;
  onClick: () => void;
  variant?: 'contained' | 'outlined' | 'text';
};

export const LoadingButton = ({ label, loading, onClick, variant = 'contained' }: Props) => {
  return (
    <Button variant={variant} onClick={onClick} disabled={loading}>
      {loading ? <CircularProgress size={20} /> : label}
    </Button>
  );
};
```

### 3.6 フォームの実装パターン（RHF + Zod）

```typescript
// Zodスキーマ定義
const loginSchema = z.object({
  email: z.string().min(1, 'メールアドレスを入力してください').email('正しいメールアドレスを入力してください'),
  password: z.string().min(1, 'パスワードを入力してください'),
});

type LoginFormValues = z.infer<typeof loginSchema>;

// フォームコンポーネント（'use client'）
const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<LoginFormValues>({
  resolver: zodResolver(loginSchema),
});
```

### 3.7 middleware.ts（アクセス制御）

```typescript
// front/src/middleware.ts
import { NextRequest, NextResponse } from 'next/server';

export async function middleware(request: NextRequest) {
  const token = request.cookies.get('access_token');
  const { pathname } = request.nextUrl;

  const publicPaths = ['/login', '/register'];
  const adminPaths = ['/admin', '/settings'];   // /settings も書き込み画面のためadmin専用
  const isAdminPath = adminPaths.some(p => pathname.startsWith(p));

  // 未認証 → ログインへリダイレクト
  if (!token && !publicPaths.includes(pathname)) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  // ログイン済みで公開パスにアクセス → ダッシュボードへリダイレクト
  if (token && publicPaths.includes(pathname)) {
    return NextResponse.redirect(new URL('/', request.url));
  }

  // admin専用パス（/admin/*・/settings）へのアクセス → roleをAPIで検証
  if (token && isAdminPath) {
    const res = await fetch(
      `${process.env.NEXT_PUBLIC_API_BASE_URL}/api/auth/me`,
      {
        headers: { Cookie: `access_token=${token.value}` },
      }
    );
    if (!res.ok) {
      return NextResponse.redirect(new URL('/login', request.url));
    }
    const user = await res.json();
    if (user.role !== 'admin') {
      // admin以外は閲覧専用のダッシュボードへ
      return NextResponse.redirect(new URL('/', request.url));
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};
```

---

## 4. バックエンド開発ガイド（DDD設計）

### 4.1 DDD層構造

budgetリポジトリの実際の構成を参考にした層設計。

| ディレクトリ | 層 | 役割 |
|------------|---|------|
| `domain/` | ドメイン層 | エンティティ・値オブジェクト・リポジトリIF（依存なし） |
| `usecase/` | ユースケース層 | 業務フロー・アプリケーションロジック |
| `repository/` | インフラ層 | リポジトリ実装（GORMでDB操作） |
| `controller/` | プレゼンテーション層 | HTTPリクエスト処理 |
| `router/` | プレゼンテーション層 | ルーティング定義 |
| `model/` | インフラ層 | GORMモデル（DBマッピング専用） |
| `internal/api/` | インフラ層 | openapi.yamlから自動生成したサーバコード |
| `migrate/` | インフラ層 | DBマイグレーション |
| `utils/` | 共通 | JWT・ハッシュ・共通処理 |
| `db/` | インフラ層 | DB接続 |

### 4.2 ディレクトリ構成

budgetリポジトリの実際の構成を踏襲。

```
back/
  ├── main.go                          # DI・ルーティング初期化・Lambda対応
  ├── .air.toml                        # airホットリロード設定
  ├── .env                             # 環境変数（gitignore）
  ├── .env.example                     # 環境変数サンプル
  ├── Makefile                         # よく使うコマンド集
  ├── openapi.yaml                     # OpenAPI仕様（フロントと共有）
  ├── Dockerfile
  ├── docker-compose.yml
  │
  ├── domain/                          # ドメイン層（外部依存なし）
  │   ├── user/
  │   │   ├── user.go                  # Userエンティティ
  │   │   ├── value_object.go          # Email・Password・Name値オブジェクト
  │   │   └── repository.go            # UserRepositoryインターフェース
  │   ├── trade/
  │   │   ├── trade.go                 # Tradeエンティティ
  │   │   ├── value_object.go          # Action・Mode値オブジェクト
  │   │   └── repository.go
  │   ├── watchlist/
  │   │   ├── watchlist.go
  │   │   ├── candidate.go             # ウォッチリスト候補エンティティ
  │   │   └── repository.go
  │   ├── position/
  │   │   ├── position.go
  │   │   └── repository.go
  │   ├── analysis/
  │   │   ├── analysis_log.go
  │   │   ├── theme.go                 # 分析テーマエンティティ
  │   │   └── repository.go
  │   ├── learning/
  │   │   ├── learning_log.go
  │   │   ├── learning_version.go
  │   │   └── repository.go
  │   └── invitation/
  │       ├── invitation_code.go
  │       └── repository.go
  │
  ├── usecase/                         # ユースケース層
  │   ├── uow.go                       # Unit of Workインターフェース（budgetと同様）
  │   ├── user_usecase.go              # 認証・ユーザー管理
  │   ├── trade_usecase.go             # トレード記録・取得
  │   ├── watchlist_usecase.go         # ウォッチリスト管理
  │   ├── candidate_usecase.go         # 候補銘柄承認・却下
  │   ├── position_usecase.go          # 保有株管理
  │   ├── analysis_usecase.go          # Claude API分析・候補提案・LINE通知
  │   ├── theme_usecase.go             # 分析テーマ管理
  │   ├── report_usecase.go            # 週次レポート・CSV更新
  │   ├── admin_usecase.go             # 管理者操作
  │   └── invitation_usecase.go        # 招待コード管理
  │
  ├── repository/                      # インフラ層（リポジトリ実装）
  │   ├── uow_impl.go                  # Unit of Work実装（budgetと同様）
  │   ├── user_repository.go
  │   ├── trade_repository.go
  │   ├── watchlist_repository.go
  │   ├── candidate_repository.go      # 候補銘柄リポジトリ実装
  │   ├── position_repository.go
  │   ├── analysis_repository.go
  │   ├── theme_repository.go          # 分析テーマリポジトリ実装
  │   ├── learning_repository.go
  │   └── invitation_repository.go
  │
  ├── controller/                      # プレゼンテーション層（HTTPハンドラ）
  │   ├── user_controller.go           # 認証・ユーザー操作
  │   ├── trade_controller.go
  │   ├── watchlist_controller.go
  │   ├── candidate_controller.go      # 候補銘柄承認・却下
  │   ├── position_controller.go
  │   ├── analysis_controller.go
  │   ├── theme_controller.go          # 分析テーマCRUD
  │   ├── report_controller.go
  │   ├── admin_controller.go
  │   └── internal_controller.go       # Lambda受取（/internal/*）
  │
  ├── router/
  │   └── router.go                    # ルーティング定義
  │
  ├── model/                           # GORMモデル（DB専用・ドメインエンティティとは別）
  │   ├── user.go
  │   ├── trade.go
  │   ├── watchlist.go
  │   ├── stock_price.go               # 最新株価スナップショットGORMモデル
  │   ├── position.go
  │   ├── analysis_log.go
  │   ├── analysis_theme.go            # 分析テーマGORMモデル
  │   ├── watchlist_candidate.go       # ウォッチリスト候補GORMモデル
  │   ├── learning_log.go
  │   ├── learning_version.go
  │   ├── invitation_code.go
  │   └── analysis_setting.go
  │
  ├── internal/
  │   └── api/
  │       ├── server.gen.go            # oapi-codegenで自動生成
  │       └── types.gen.go             # oapi-codegenで自動生成
  │
  ├── migrate/
  │   └── migrate.go                   # AutoMigrateでテーブル作成
  │
  ├── db/
  │   └── db.go                        # Neon DB接続
  │
  ├── utils/
  │   ├── jwt.go                       # JWT発行・検証
  │   ├── password.go                  # bcryptハッシュ化
  │   └── response.go                  # 統一レスポンス形式
  │
  ├── go.mod
  └── go.sum
```

### 4.3 domainの実装例（budgetを踏襲）

```go
// domain/user/user.go
package user

import "time"

// UserID はUserエンティティの識別子（値オブジェクト）
type UserID uint

// Role はユーザーロールを表す値オブジェクト。
// stringで管理するとタイポリスクがあるため専用型を定義する。
type Role string

const (
  RoleAdmin Role = "admin"
  RoleUser  Role = "user"
)

func (r Role) IsValid() bool {
  return r == RoleAdmin || r == RoleUser
}

type User struct {
  ID        UserID
  Email     *Email
  Password  Password
  Name      Name
  Role      Role
  IsActive  bool
  CreatedAt time.Time
  UpdatedAt time.Time
}

// NewUser は新しいUserドメインエンティティを生成します。
func NewUser(email, password, name string) (*User, error) {
  voEmail, err := NewEmail(email)
  if err != nil {
    return nil, err
  }
  voPassword, err := NewPassword(password)
  if err != nil {
    return nil, err
  }
  voName, err := NewName(name)
  if err != nil {
    return nil, err
  }
  return &User{
    Email:     voEmail,
    Password:  voPassword,
    Name:      voName,
    Role:      RoleUser,     // 登録時は常にuserロール
    IsActive:  true,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
  }, nil
}

// IsAdmin checks if the user has admin rights.
func (u *User) IsAdmin() bool {
  return u.Role == RoleAdmin
}

// CanLogin checks if the user is active.
func (u *User) CanLogin() bool {
  return u.IsActive
}
```

```go
// domain/user/value_object.go
package user

import (
  "errors"
  "regexp"
)

type Email struct{ value string }

func NewEmail(email string) (*Email, error) {
  re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
  if !re.MatchString(email) {
    return nil, errors.New("メールアドレスの形式が正しくありません")
  }
  return &Email{value: email}, nil
}

func (e *Email) Value() string { return e.value }

type Name struct{ value string }

func NewName(name string) (Name, error) {
  if name == "" {
    return Name{}, errors.New("名前を入力してください")
  }
  if len([]rune(name)) > 50 {
    return Name{}, errors.New("名前は50文字以内で入力してください")
  }
  return Name{value: name}, nil
}

func (n Name) Value() string { return n.value }

type Password struct{ value string }

func NewPassword(raw string) (Password, error) {
  if len(raw) < 8 {
    return Password{}, errors.New("パスワードは8文字以上必要です")
  }
  // bcryptハッシュ化はusecaseで実施
  return Password{value: raw}, nil
}

func (p Password) Value() string { return p.value }
```

```go
// domain/user/repository.go
package user

type UserRepository interface {
  FindByID(id UserID) (*User, error)
  FindByEmail(email string) (*User, error)
  Save(user *User) error
  Update(user *User) error
  Delete(id UserID) error
  FindAll() ([]*User, error)
}
```

### 4.4 Unit of Work（トランザクション設計）

#### DDDにおけるトランザクションの原則

```
ドメイン層    → トランザクションを知らない（GORMもDBも知らない）
ユースケース層 → トランザクション境界を「宣言」する
インフラ層    → トランザクションの「実装」を担う（UoWImpl）
```

**トランザクション境界はユースケースが決め、実装はインフラ層（UoW）が持つ。**  
リポジトリ個別にトランザクションを持たせると、複数リポジトリをまたぐ操作の原子性が保証できないため採用しない。

#### usecase/uow.go（インターフェース定義）

```go
// usecase/uow.go
package usecase

import "context"

// UnitOfWork はトランザクション境界を宣言するインターフェース。
// ユースケース層はこのIFに依存し、実装（GORM）を知らない。
type UnitOfWork interface {
  Do(ctx context.Context, fn func(repos *Repositories) error) error
}

// Repositories はトランザクション内で使用できる全リポジトリを保持する。
// UoW.Do() の中でのみ使用し、全DB操作が同一トランザクションに含まれることを保証する。
type Repositories struct {
  User               user.UserRepository
  InvitationCode     invitation.InvitationCodeRepository
  Watchlist           watchlist.WatchlistRepository        // 全ユーザー共通
  WatchlistCandidate watchlist.CandidateRepository        // 全ユーザー共通
  Trade              trade.TradeRepository                // ユーザーごと
  Position           position.PositionRepository          // ユーザーごと
  AnalysisLog        analysis.AnalysisLogRepository       // 全ユーザー共通
  AnalysisTheme      analysis.ThemeRepository             // 全ユーザー共通
  LearningLog        learning.LearningLogRepository       // 全ユーザー共通（管理者のtradesから生成）
  LearningVersion    learning.LearningVersionRepository   // 全ユーザー共通
  AnalysisSetting    setting.AnalysisSettingRepository    // 全ユーザー共通
}
```

#### repository/uow_impl.go（実装）

```go
// repository/uow_impl.go
package repository

// UnitOfWorkImpl はGORMのTransactionを使ってUoWを実装する。
// domain層・usecase層はこの実装を知らない。
type UnitOfWorkImpl struct {
  db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) usecase.UnitOfWork {
  return &UnitOfWorkImpl{db: db}
}

func (u *UnitOfWorkImpl) Do(ctx context.Context, fn func(*usecase.Repositories) error) error {
  return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    // トランザクション内のtxを使って全Repositoryを構築
    // → 全DB操作が同一トランザクションに含まれる
    repos := &usecase.Repositories{
      User:               NewUserRepositoryImpl(tx),
      InvitationCode:     NewInvitationCodeRepositoryImpl(tx),
      Watchlist:           NewWatchlistRepositoryImpl(tx),
      WatchlistCandidate: NewCandidateRepositoryImpl(tx),
      Trade:              NewTradeRepositoryImpl(tx),
      Position:           NewPositionRepositoryImpl(tx),
      AnalysisLog:        NewAnalysisLogRepositoryImpl(tx),
      AnalysisTheme:      NewThemeRepositoryImpl(tx),
      LearningLog:        NewLearningLogRepositoryImpl(tx),
      LearningVersion:    NewLearningVersionRepositoryImpl(tx),
      AnalysisSetting:    NewAnalysisSettingRepositoryImpl(tx),
    }
    return fn(repos)
    // fnがerrorを返す → 自動ロールバック
    // fnがnilを返す  → 自動コミット
  })
}
```

#### ユースケースでの使用例①：招待コード登録（複数リポジトリをまたぐ）

```go
// usecase/user_usecase.go
func (u *UserUsecase) Register(ctx context.Context, code, email, password, name string) error {
  // UoW.Do の中で全DB操作を行う → 原子性が保証される
  return u.uow.Do(ctx, func(repos *usecase.Repositories) error {

    // 招待コード検証
    invitation, err := repos.InvitationCode.FindByCode(code)
    if err != nil {
      return errors.New("無効な招待コードです")
    }
    if !invitation.IsValid() {
      return errors.New("招待コードの有効期限が切れています")
    }

    // ユーザー作成
    newUser, err := user.NewUser(email, password, name)
    if err != nil {
      return err
    }
    if err := repos.User.Save(newUser); err != nil {
      return err
    }

    // 招待コードを使用済みにマーク
    invitation.MarkAsUsed(newUser.ID)
    if err := repos.InvitationCode.Update(invitation); err != nil {
      return err
      // ↑ ここで失敗 → User保存もInvitationCode更新も両方ロールバック
    }

    return nil // → 両方コミット
  })
}
```

#### ユースケースでの使用例②：分析結果保存（外部APIとの分離）

外部API（Claude・LINE・S3）はトランザクション外に置く。DB保存成功後にのみ通知を実行する。

```go
// usecase/analysis_usecase.go
func (u *AnalysisUsecase) Analyze(ctx context.Context, stockData StockData) error {

  // ① 外部API呼び出し（トランザクション外）
  // 失敗してもDBに影響しない
  result, err := u.claudeClient.Analyze(stockData)
  if err != nil {
    return err
  }

  // ② DB保存（トランザクション内で一括コミット）
  // AnalysisLogとTradeを同時に保存 → どちらかが失敗すれば両方ロールバック
  err = u.uow.Do(ctx, func(repos *usecase.Repositories) error {
    if err := repos.AnalysisLog.Save(result); err != nil {
      return err
    }
    if result.Action != "HOLD" {
      if err := repos.Trade.Save(result); err != nil {
        return err
      }
    }
    return nil
  })
  if err != nil {
    return err // DB保存失敗 → LINE通知しない
  }

  // ③ LINE通知（DB保存成功後のみ実行）
  if result.Action != "HOLD" {
    u.lineClient.Notify(result)
  }

  return nil
}
```

#### トランザクション設計のルール

| ルール | 内容 |
|--------|------|
| リポジトリはトランザクションを持たない | 個別トランザクションは複数リポジトリ間の原子性を壊す |
| 全DB操作はuow.Do内で行う | トランザクション外でrepos.X.Save()を呼ばない |
| 外部APIはuow.Doの外に置く | Claude API・LINE・S3はトランザクションに含めない |
| contextを必ず伝播させる | タイムアウト・キャンセルが正しく機能するようにする |
| fnがerrorを返せば自動ロールバック | GORMのTransactionが自動で処理する |

### 4.5 main.goのDI構成（budgetを踏襲・Lambda対応）

```go
// back/main.go
var ginLambda *ginadapter.GinLambdaV2

func setupRouter() *gin.Engine {
  dbInstance := db.NewDB()

  // UoW（全リポジトリのトランザクションを管理）
  uow := repository.NewUnitOfWork(dbInstance)

  // 読み取り専用のRepository（トランザクション不要な参照系）
  // ※ 書き込みはuow.Do内でreposを使う
  userRepo      := repository.NewUserRepositoryImpl(dbInstance)
  tradeRepo     := repository.NewTradeRepositoryImpl(dbInstance)
  watchlistRepo  := repository.NewWatchlistRepositoryImpl(dbInstance)
  analysisRepo  := repository.NewAnalysisLogRepositoryImpl(dbInstance)

  // External Services（外部API・トランザクション外）
  claudeClient := external.NewClaudeClient()
  lineClient   := external.NewLineClient()
  s3Client     := external.NewS3Client()

  // Usecases（uowを注入・書き込みはuow.Do経由）
  userUsecase     := usecase.NewUserUsecase(uow, userRepo)
  tradeUsecase    := usecase.NewTradeUsecase(uow, tradeRepo)
  analysisUsecase := usecase.NewAnalysisUsecase(uow, analysisRepo, claudeClient, lineClient, s3Client)
  reportUsecase   := usecase.NewReportUsecase(uow, tradeRepo, s3Client, claudeClient, lineClient)
  adminUsecase    := usecase.NewAdminUsecase(uow, userRepo)

  // Controllers
  userController     := controller.NewUserController(userUsecase)
  tradeController    := controller.NewTradeController(tradeUsecase)
  analysisController := controller.NewAnalysisController(analysisUsecase)
  reportController   := controller.NewReportController(reportUsecase)
  adminController    := controller.NewAdminController(adminUsecase)
  internalController := controller.NewInternalController(analysisUsecase, reportUsecase)

  return router.NewRouter(
    userController,
    tradeController,
    analysisController,
    reportController,
    adminController,
    internalController,
  )
}

// scheduledEvent はEventBridgeルールが渡す定数input（{"job": "..."}）。
type scheduledEvent struct {
  Job string `json:"job"`
}

// dispatch はイベント種別で処理を振り分ける。
//   - EventBridge定期実行（{"job":"analyze"|"weekly_report"}）→ 対応usecaseを直接実行
//   - それ以外（API Gateway HTTP API イベント）→ ginにプロキシ
// 同一バイナリを ApiFunction / WorkerFunction の2つのLambdaにデプロイし、本関数が振り分ける（案B）。
func dispatch(ctx context.Context, raw json.RawMessage) (interface{}, error) {
  var ev scheduledEvent
  if err := json.Unmarshal(raw, &ev); err == nil && ev.Job != "" {
    switch ev.Job {
    case "analyze":
      return nil, analysisUsecaseRef.RunScheduled(ctx)       // 平日15:30
    case "weekly_report":
      return nil, reportUsecaseRef.RunWeekly(ctx)            // 日曜18:00
    default:
      return nil, fmt.Errorf("dispatch: 未知のjob %q", ev.Job)
    }
  }
  var req events.APIGatewayV2HTTPRequest
  if err := json.Unmarshal(raw, &req); err != nil {
    return nil, err
  }
  return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
  r := setupRouter()
  if _, ok := os.LookupEnv("LAMBDA_TASK_ROOT"); ok {
    ginLambda = ginadapter.NewV2(r)
    lambda.Start(dispatch) // API Gateway と EventBridge定期実行の両方を1つのdispatchで処理
  } else {
    log.Fatal(r.Run(":8080"))
  }
}
```

> **定期実行（案B）**: 分析（平日15:30）・週次レポート（日曜18:00）は EventBridge が **Worker Lambda を
> 直接invoke** する（API Gateway/HTTPを介さない）。`{"job":"..."}` を `dispatch` が判定して
> `RunScheduled`/`RunWeekly` を呼ぶ。利点は Secrets Manager コスト削減（Connection不要）と
> API Gatewayの30秒制限回避（最大15分）。詳細は `doc/dev-spec/infra_architecture.md`。

### 4.6 OpenAPI仕様とコード生成

budgetリポジトリと同様に **openapi.yaml** を単一の真実のソースとして管理する。

```
back/openapi.yaml
  ↓ oapi-codegen（バックエンド）
back/internal/api/server.gen.go   # サーバインターフェース自動生成
back/internal/api/types.gen.go    # リクエスト・レスポンス型自動生成

  ↓ openapi-typescript（フロントエンド）
front/src/types/api.ts            # TypeScript型自動生成
```

```makefile
# back/Makefile
generate-go:
	oapi-codegen -generate types -o internal/api/types.gen.go -package api openapi.yaml
	oapi-codegen -generate gin-server -o internal/api/server.gen.go -package api openapi.yaml

migrate:
	go run migrate/migrate.go
```

### 4.7 エラーレスポンス規則

```go
// utils/response.go
type ErrorResponse struct {
  Error string `json:"error"`
}

// HTTPステータスコード規則
// 200 OK             - 正常
// 201 Created        - 作成成功
// 400 Bad Request    - バリデーションエラー
// 401 Unauthorized   - 未認証
// 403 Forbidden      - 権限なし
// 404 Not Found      - リソースなし
// 500 Internal Error - サーバーエラー
```

### 4.8 airホットリロード設定

```toml
# back/.air.toml
[build]
  cmd = "go build -o ./tmp/main ."
  bin = "./tmp/main"
  include_ext = ["go"]
  exclude_dir = ["tmp", "vendor"]

[log]
  time = true
```

---

## 5. 株価取得開発ガイド

### 5.1 ディレクトリ構成

```
lambda/
  ├── fetch_price.py      # メイン処理（株価取得 → GoにPOST）
  ├── requirements.txt
  └── .env
```

### 5.2 処理フロー

```python
# fetch_price.py
import yfinance as yf
import requests
import os
import json
from datetime import datetime, timezone, timedelta

GO_API_BASE_URL     = os.environ['GO_API_BASE_URL']
INTERNAL_API_SECRET = os.environ['INTERNAL_API_SECRET']

def now_jst() -> str:
    """現在時刻をJST（UTC+9）のISO8601形式で返す"""
    jst = timezone(timedelta(hours=9))
    return datetime.now(jst).isoformat()

def format_stock_data(ticker: str, df) -> dict:
    """yfinanceのDataFrameをGoのAPIが受け取れるJSON形式に変換する"""
    prices = []
    for date, row in df.iterrows():
        prices.append({
            'date':   date.strftime('%Y-%m-%d'),
            'open':   round(float(row['Open']),   2),
            'high':   round(float(row['High']),   2),
            'low':    round(float(row['Low']),    2),
            'close':  round(float(row['Close']),  2),
            'volume': int(row['Volume']),
        })
    return {'ticker': ticker, 'prices': prices}

def handler(event, context):
    # 1. GoのAPIからウォッチリスト取得
    headers = {'X-Internal-Secret': INTERNAL_API_SECRET}
    resp = requests.get(f'{GO_API_BASE_URL}/internal/watchlist', headers=headers)
    resp.raise_for_status()
    watchlist = resp.json()

    # 2. yfinanceで各銘柄の過去120日分OHLCVを取得
    stocks = []
    for item in watchlist:
        ticker = item['ticker']
        try:
            df = yf.Ticker(ticker).history(period='120d')
            if df.empty:
                print(f'WARNING: {ticker} のデータが取得できませんでした')
                continue
            stocks.append(format_stock_data(ticker, df))
        except Exception as e:
            print(f'ERROR: {ticker} の取得に失敗しました: {e}')
            continue

    if not stocks:
        print('取得できた銘柄がありません。処理を終了します。')
        return {'statusCode': 200, 'body': 'no stocks'}

    # 3. GoのAPIにPOST（以降の処理はGoに委譲）
    payload = {'fetched_at': now_jst(), 'stocks': stocks}
    result = requests.post(
        f'{GO_API_BASE_URL}/internal/stock-prices',
        json=payload,
        headers=headers,
    )
    result.raise_for_status()
    return {'statusCode': 200, 'body': result.text}

# ローカルテスト用
if __name__ == '__main__':
    from dotenv import load_dotenv
    load_dotenv()
    print(handler({}, {}))
```

### 5.3 requirements.txt

```
yfinance==0.2.40
requests==2.32.0
pandas==2.2.0
python-dotenv==1.0.0
```

---

## 6. データベースガイド

### 6.1 マイグレーション方針

budgetリポジトリと同様に `migrate/` ディレクトリにGORMの `AutoMigrate` を使用。

```go
// back/migrate/migrate.go
func Migrate(db *gorm.DB) {
  db.AutoMigrate(
    &model.User{},
    &model.InvitationCode{},
    &model.Watchlist{},
    &model.WatchlistCandidate{},   // AIが提案した候補銘柄
    &model.StockPrice{},           // 最新株価スナップショット（現在値・前日比のソース・UPSERT）
    &model.Trade{},
    &model.AnalysisLog{},
    &model.RealPosition{},
    &model.LearningLog{},
    &model.LearningVersion{},
    &model.AnalysisSetting{},
    &model.AnalysisTheme{},        // 管理者がUI上で管理するテーマ
  )
}
```

#### ⚠️ AutoMigrateの制約と注意事項

AutoMigrateは便利だが以下の操作は**自動では行わない**。本番環境での操作には注意が必要。

| 操作 | AutoMigrateの挙動 | 対処 |
|------|-----------------|------|
| カラム追加 | ✅ 自動で追加される | そのままOK |
| カラム削除 | ❌ 自動で削除されない | 手動でALTER TABLE |
| カラムの型変更 | ❌ 自動で変更されない | 手動でALTER TABLE |
| インデックス追加 | ✅ 自動で追加される | そのままOK |
| テーブル名変更 | ❌ 古いテーブルが残る | 手動で対応 |

**本番環境でのマイグレーション手順：**
```bash
# 1. 変更内容を事前にNeon DBのブランチ機能で検証する
# 2. ステージング環境でAutoMigrateを実行して確認
# 3. カラム削除・型変更が必要な場合は手動SQLを別途実行
# 4. 本番環境でAutoMigrateを実行
docker compose exec app make migrate
```

**将来的な移行先（規模拡大時）：** golang-migrate などの明示的なマイグレーションツールへの移行を検討する。

### 6.2 GORMモデル定義規則（model/）

GORMモデルはDBマッピング専用。ドメインエンティティとは別に定義。

```go
// model/user.go
type User struct {
  gorm.Model
  Email        string `gorm:"uniqueIndex;not null"`
  Name         string `gorm:"not null"`
  PasswordHash string `gorm:"not null"`
  Role         string `gorm:"default:user"`
  IsActive     bool   `gorm:"default:true"`
}

// model/analysis_log.go
// 全ユーザー共通のためUserIDフィールドなし
type AnalysisLog struct {
  gorm.Model
  Ticker     string         `gorm:"not null;index"`
  Action     string
  Confidence float64
  Analysis   datatypes.JSON // JSONBカラム（gorm.io/datatypes）
}

// model/watchlist.go
// 全ユーザー共通・管理者が管理・user_idなし
type Watchlist struct {
  gorm.Model
  Ticker    string `gorm:"uniqueIndex;not null"`
  Name      string
  Mode      string `gorm:"not null;check:mode IN ('virtual','real','both')"`
  IsActive  bool   `gorm:"default:true"`
}

// model/stock_price.go
// 全ユーザー共通・user_idなし。銘柄ごとに最新の現在値・前日比を1行だけ保持する。
// 日次履歴は蓄積せず、Ticker のユニークキーでUPSERT（上書き）する。
// 現在値・前日比の参照元（含み益は参照時にユーザー単価と突き合わせて算出）。
type StockPrice struct {
  gorm.Model
  Ticker       string    `gorm:"uniqueIndex;not null"`        // UPSERTキー（1銘柄1行）
  Date         time.Time `gorm:"type:date;not null"`          // 最新終値の取引日
  Open         float64   `gorm:"type:numeric(10,2);not null"`
  High         float64   `gorm:"type:numeric(10,2);not null"`
  Low          float64   `gorm:"type:numeric(10,2);not null"`
  Close        float64   `gorm:"type:numeric(10,2);not null"` // 現在値（最新終値）
  PrevClose    float64   `gorm:"type:numeric(10,2)"`          // 前営業日終値
  ChangeAmount float64   `gorm:"type:numeric(10,2)"`          // 前日比（円）
  ChangeRate   float64   `gorm:"type:numeric(6,2)"`           // 前日比（%）
  Volume       int64     `gorm:"not null;default:0"`
}

// UPSERT例（repository層・gorm clauseを使用）
//   import "gorm.io/gorm/clause"
//   db.Clauses(clause.OnConflict{
//     Columns:   []clause.Column{{Name: "ticker"}},
//     DoUpdates: clause.AssignmentColumns([]string{
//       "date", "open", "high", "low", "close",
//       "prev_close", "change_amount", "change_rate", "volume", "updated_at",
//     }),
//   }).Create(&price)

// model/learning_log.go
// 全ユーザー共通・管理者のtradesから生成・user_idなし
type LearningLog struct {
  gorm.Model
  WeekStart   time.Time
  WeekEnd     time.Time
  TradeCount  int
  WinRate     float64
  TotalPnl    float64
  Summary     string
  Lessons     string
  Strategy    string
  RawResponse string
}

// model/learning_version.go
// 全ユーザー共通・user_idなし
type LearningVersion struct {
  gorm.Model
  Version   int    `gorm:"not null"`
  S3Path    string `gorm:"not null"`
  WeekRange string
  CharCount int
}

// model/analysis_theme.go
type AnalysisTheme struct {
  gorm.Model
  Name        string `gorm:"uniqueIndex;not null"`
  Description string
  SortOrder   int    `gorm:"default:0;index"`
  IsActive    bool   `gorm:"default:true"`
  CreatedBy   *uint
}

// model/watchlist_candidate.go
type WatchlistCandidate struct {
  gorm.Model
  Ticker        string     `gorm:"not null"`
  Name          string
  Reason        string
  ReplaceTicker string
  Confidence    float64
  Status        string     `gorm:"default:pending;check:status IN ('pending','approved','rejected');index"`
  ProposedAt    time.Time  `gorm:"autoCreateTime"`
  DecidedAt     *time.Time
  DecidedBy     *uint
}
```

### 6.3 データアクセス規則

- ユーザーごとのテーブルは全クエリに `WHERE user_id = ?` を必ず付与（データ完全分離）
- 以下のテーブルは `user_id` を持たないため `WHERE user_id = ?` は不要

| テーブル | 理由 |
|---------|------|
| `watchlist` | 全ユーザー共通・管理者が管理 |
| `stock_prices` | 市場データ・全ユーザー共通 |
| `analysis_logs` | Lambda一括実行・全ユーザー共通 |
| `learning_logs` | 管理者のtradesから生成・全ユーザー共通 |
| `learning_versions` | 全ユーザー共通のCSV |

- repositoryはドメインエンティティを返す（GORMモデルをそのまま返さない）

```go
// repository/trade_repository.go（ユーザーごと）
func (r *TradeRepositoryImpl) FindByUserID(userID uint) ([]*trade.Trade, error) {
  var models []model.Trade
  if err := r.db.Where("user_id = ?", userID).Find(&models).Error; err != nil {
    return nil, err
  }
  return toTradeEntities(models), nil
}

// repository/watchlist_repository.go（全ユーザー共通）
func (r *WatchlistRepositoryImpl) FindAll() ([]*watchlist.Watchlist, error) {
  var models []model.Watchlist
  if err := r.db.Where("is_active = ?", true).Find(&models).Error; err != nil {
    return nil, err
  }
  return toWatchlistEntities(models), nil
}

// repository/learning_log_repository.go（全ユーザー共通）
func (r *LearningLogRepositoryImpl) FindLatest(limit int) ([]*learning.LearningLog, error) {
  var models []model.LearningLog
  if err := r.db.Order("week_start DESC").Limit(limit).Find(&models).Error; err != nil {
    return nil, err
  }
  return toLearningLogEntities(models), nil
}

// repository/weekly_report.go（週次レポート集計：管理者のtradesのみ）
func (r *TradeRepositoryImpl) FindByAdminForWeeklyReport(weekStart, weekEnd time.Time) ([]*trade.Trade, error) {
  var models []model.Trade
  err := r.db.
    Joins("JOIN users ON users.id = trades.user_id").
    Where("users.role = ? AND trades.created_at BETWEEN ? AND ?", "admin", weekStart, weekEnd).
    Find(&models).Error
  if err != nil {
    return nil, err
  }
  return toTradeEntities(models), nil
}
```
}
```

---

## 7. 環境変数管理

### 7.1 バックエンド（back/.env）

```env
# DB
NEON_DATABASE_URL=postgresql://user:pass@host/dbname?sslmode=require

# アプリ環境（production のとき Cookie に Secure を付与）
APP_ENV=development

# JWT
JWT_SECRET=your_jwt_secret_key
JWT_ACCESS_EXPIRE_HOURS=24
JWT_REFRESH_EXPIRE_DAYS=30

# Claude API
ANTHROPIC_API_KEY=sk-ant-xxxxx

# LINE
LINE_CHANNEL_ACCESS_TOKEN=xxxxx
LINE_USER_ID=xxxxx

# AWS S3
AWS_ACCESS_KEY_ID=xxxxx
AWS_SECRET_ACCESS_KEY=xxxxx
AWS_REGION=ap-northeast-1
S3_BUCKET_NAME=trading-system-learning

# 内部API認証（Lambda → Go）
INTERNAL_API_SECRET=xxxxx
```

### 7.2 フロントエンド（front/.env.local）

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=AI Trading System
```

### 7.3 Lambda（lambda/.env）

```env
GO_API_BASE_URL=https://api.your-domain.com
INTERNAL_API_SECRET=xxxxx
```

---

## 8. 開発フロー・命名規則

### 8.1 ブランチ戦略

```
main          # 本番環境
  └── develop # 開発統合ブランチ
       ├── feature/auth
       ├── feature/watchlist
       └── feature/analysis
```

### 8.2 命名規則

#### TypeScript / Next.js

| 対象 | 規則 | 例 |
|------|------|-----|
| コンポーネントファイル | PascalCase | `LoginForm.tsx` |
| コンポーネントフォルダ | camelCase | `buttonBox/` |
| 型・インターフェース | PascalCase | `type TradeItem` |
| 定数 | UPPER_SNAKE_CASE | `API_BASE_URL` |
| 関数・変数 | camelCase | `handleSubmit` |
| SCSSモジュール | camelCase | `styles.loginForm` |

#### Go

| 対象 | 規則 | 例 |
|------|------|-----|
| パッケージ | lowercase | `user`, `trade` |
| エンティティ・構造体 | PascalCase | `User`, `Trade` |
| 関数（公開） | PascalCase | `NewUser`, `FindByID` |
| 関数（非公開） | camelCase | `toTradeEntity` |
| ファイル名 | snake_case | `user_repository.go` |

#### Python

| 対象 | 規則 | 例 |
|------|------|-----|
| 関数・変数 | snake_case | `fetch_prices` |
| 定数 | UPPER_SNAKE_CASE | `GO_API_BASE_URL` |

### 8.3 コミットメッセージ規則

```
feat: ログイン機能を実装
fix: JWT検証エラーを修正
refactor: TradeRepositoryをリファクタリング
docs: devspecにポートフォリオ画面を追加
chore: 依存パッケージを更新
```

---

## 8a. テスト方針

### 8a.1 テストの基本方針

DDD設計の恩恵を活かし、**ユースケース層のユニットテストを最優先**とする。リポジトリがインターフェースに依存しているためモックに差し替えが容易。

| 層 | テスト種別 | 方針 |
|---|-----------|------|
| domain（エンティティ・値オブジェクト） | ユニットテスト | バリデーションロジックを重点的にテスト |
| usecase | ユニットテスト | リポジトリをモックに差し替えてテスト |
| repository | インテグレーションテスト | Neon DBのテスト用ブランチを使用 |
| controller | E2Eテスト | 優先度低・必要に応じて追加 |

### 8a.2 Goのユニットテスト（usecase）

リポジトリをモックに差し替えて、外部依存なしでテストする。

```go
// usecase/user_usecase_test.go
package usecase_test

import (
  "context"
  "testing"
  "errors"
)

// UserRepositoryのモック実装
type mockUserRepository struct {
  findByEmailFn func(email string) (*user.User, error)
  saveFn        func(u *user.User) error
}

func (m *mockUserRepository) FindByEmail(email string) (*user.User, error) {
  return m.findByEmailFn(email)
}
func (m *mockUserRepository) Save(u *user.User) error {
  return m.saveFn(u)
}
// 他のメソッドは省略（テストで使わないものはpanicで実装）

// InvitationCodeRepositoryのモック
type mockInvitationCodeRepository struct {
  findByCodeFn func(code string) (*invitation.InvitationCode, error)
  updateFn     func(inv *invitation.InvitationCode) error
}

// UnitOfWorkのモック（トランザクションをスキップして直接fnを実行）
type mockUnitOfWork struct{}

func (m *mockUnitOfWork) Do(ctx context.Context, fn func(*usecase.Repositories) error) error {
  return fn(&usecase.Repositories{
    User:           &mockUserRepository{ /* ... */ },
    InvitationCode: &mockInvitationCodeRepository{ /* ... */ },
  })
}

func TestRegister_InvalidInvitationCode(t *testing.T) {
  uow := &mockUnitOfWork{}
  uc  := usecase.NewUserUsecase(uow, nil)

  err := uc.Register(context.Background(), "INVALID-CODE", "test@example.com", "password123", "テスト")

  if err == nil {
    t.Fatal("エラーが返るべきなのにnilが返った")
  }
  if err.Error() != "無効な招待コードです" {
    t.Fatalf("予期しないエラーメッセージ: %s", err.Error())
  }
}

func TestRegister_Success(t *testing.T) {
  // 正常系のテスト
}
```

### 8a.3 domainのユニットテスト（値オブジェクト）

```go
// domain/user/value_object_test.go
package user_test

func TestNewEmail_Invalid(t *testing.T) {
  cases := []struct{ input string }{
    {"not-an-email"},
    {"missing@"},
    {"@nodomain.com"},
  }
  for _, c := range cases {
    _, err := user.NewEmail(c.input)
    if err == nil {
      t.Errorf("入力 %q はエラーになるべき", c.input)
    }
  }
}

func TestNewEmail_Valid(t *testing.T) {
  _, err := user.NewEmail("user@example.com")
  if err != nil {
    t.Fatalf("有効なメールアドレスでエラー: %v", err)
  }
}

func TestNewPassword_TooShort(t *testing.T) {
  _, err := user.NewPassword("short")
  if err == nil {
    t.Fatal("8文字未満はエラーになるべき")
  }
}
```

### 8a.4 テスト実行コマンド（Makefile追加）

```makefile
# back/Makefile（追加）
test:
	go test ./... -v

test-unit:
	go test ./domain/... ./usecase/... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
```

---

## 8b. エラーハンドリング規則

### 8b.1 Goのエラーラッピング規則

```go
// ✅ 正しい書き方：%w でラップしてスタックトレースを保持
func (u *UserUsecase) Register(ctx context.Context, ...) error {
  user, err := repos.User.FindByEmail(email)
  if err != nil {
    return fmt.Errorf("Register: FindByEmail failed: %w", err)
  }
  return nil
}

// ❌ 避ける書き方：エラーを握りつぶす
if err != nil {
  return errors.New("エラーが発生しました") // 元のエラー情報が失われる
}
```

### 8b.2 エラー種別の定義

ドメイン固有のエラーは `domain/` で定義し、HTTPステータスコードへのマッピングはcontroller層で行う。

```go
// domain/errors.go
package domain

import "errors"

var (
  ErrNotFound         = errors.New("リソースが見つかりません")
  ErrUnauthorized     = errors.New("認証が必要です")
  ErrForbidden        = errors.New("権限がありません")
  ErrInvalidInput     = errors.New("入力値が不正です")
  ErrAlreadyExists    = errors.New("すでに存在しています")
  ErrInvalidCode      = errors.New("無効な招待コードです")
  ErrExpiredCode      = errors.New("招待コードの有効期限が切れています")
  ErrAccountDisabled  = errors.New("このアカウントは無効です")
)
```

```go
// controller層でドメインエラー → HTTPステータスにマッピング
// utils/response.go
func HandleError(c *gin.Context, err error) {
  switch {
  case errors.Is(err, domain.ErrNotFound):
    c.JSON(404, gin.H{"error": err.Error()})
  case errors.Is(err, domain.ErrUnauthorized):
    c.JSON(401, gin.H{"error": err.Error()})
  case errors.Is(err, domain.ErrForbidden):
    c.JSON(403, gin.H{"error": err.Error()})
  case errors.Is(err, domain.ErrInvalidInput),
       errors.Is(err, domain.ErrAlreadyExists),
       errors.Is(err, domain.ErrInvalidCode),
       errors.Is(err, domain.ErrExpiredCode):
    c.JSON(400, gin.H{"error": err.Error()})
  case errors.Is(err, domain.ErrAccountDisabled):
    c.JSON(403, gin.H{"error": err.Error()})
  default:
    // 想定外エラーはログに記録して500を返す
    log.Printf("unexpected error: %+v", err)
    c.JSON(500, gin.H{"error": "サーバーエラーが発生しました"})
  }
}
```

アクセストークン・リフレッシュトークンの**両方**をHttpOnly Cookieに発行する（spec 4.10）。`SameSite=Strict`・`HttpOnly` は必須、`Secure` は本番のみ有効（ローカルhttpではCookieが付与されないため環境変数で切り替える）。Ginの `c.SetCookie` は `SameSite` を別途 `c.SetSameSite` で指定する。

```go
// utils/cookie.go : トークンCookieの発行を共通化
const (
  accessTokenMaxAge  = 24 * 60 * 60        // 24時間（秒）
  refreshTokenMaxAge = 30 * 24 * 60 * 60   // 30日（秒）
)

func SetAuthCookies(c *gin.Context, accessToken, refreshToken string) {
  secure := os.Getenv("APP_ENV") == "production" // 本番のみSecure
  c.SetSameSite(http.SameSiteStrictMode)         // SameSite=Strict（CSRF対策）
  // 引数: name, value, maxAge, path, domain, secure, httpOnly
  c.SetCookie("access_token",  accessToken,  accessTokenMaxAge,  "/", "", secure, true)
  c.SetCookie("refresh_token", refreshToken, refreshTokenMaxAge, "/", "", secure, true)
}

func ClearAuthCookies(c *gin.Context) {
  secure := os.Getenv("APP_ENV") == "production"
  c.SetSameSite(http.SameSiteStrictMode)
  c.SetCookie("access_token",  "", -1, "/", "", secure, true)
  c.SetCookie("refresh_token", "", -1, "/", "", secure, true)
}
```

```go
// controller での使用例
func (uc *UserController) Login(c *gin.Context) {
  var req dto.LoginRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": "リクエストが不正です"})
    return
  }
  // usecaseはアクセストークンとリフレッシュトークンの両方を返す
  user, accessToken, refreshToken, err := uc.userUsecase.Login(c.Request.Context(), req.Email, req.Password)
  if err != nil {
    utils.HandleError(c, err) // ← 共通エラーハンドラに委譲
    return
  }
  utils.SetAuthCookies(c, accessToken, refreshToken) // 両トークンをHttpOnly Cookieへ
  c.JSON(200, dto.ToLoginResponse(user))
}

// POST /api/auth/refresh : refresh_tokenから新しいアクセストークンを再発行
func (uc *UserController) Refresh(c *gin.Context) {
  rt, err := c.Cookie("refresh_token")
  if err != nil {
    c.JSON(401, gin.H{"error": "認証が必要です"})
    return
  }
  accessToken, refreshToken, err := uc.userUsecase.Refresh(c.Request.Context(), rt)
  if err != nil {
    utils.HandleError(c, err)
    return
  }
  utils.SetAuthCookies(c, accessToken, refreshToken) // ローテーション（リフレッシュトークンも再発行）
  c.JSON(200, gin.H{"message": "再発行しました"})
}

// POST /api/auth/logout : 両トークンを削除
func (uc *UserController) Logout(c *gin.Context) {
  utils.ClearAuthCookies(c)
  c.JSON(200, gin.H{"message": "ログアウトしました"})
}
```

### 8b.3 エラーハンドリングルール

| ルール | 内容 |
|--------|------|
| `%w` でラップ | `fmt.Errorf("context: %w", err)` でエラーを伝播させる |
| ドメインエラーは `domain/errors.go` で定義 | HTTPステータスをドメイン層に持ち込まない |
| controller層でマッピング | `utils.HandleError` で一元的にHTTPレスポンスに変換 |
| 想定外エラーはログ記録 | `log.Printf` でスタックトレースを記録してから500を返す |
| エラーを握りつぶさない | `_ = someFunc()` は原則禁止 |

---

## 8c. CORS設定

### 8c.1 概要

フロントエンド（localhost:3000）からバックエンド（localhost:8080）へのクロスオリジンリクエストを許可するために、Ginにgo-corsミドルウェアを設定する。

```go
// router/router.go
import "github.com/gin-contrib/cors"

func NewRouter(...) *gin.Engine {
  r := gin.Default()

  r.Use(cors.New(cors.Config{
    AllowOrigins: []string{
      "http://localhost:3000",                        // ローカル開発
      os.Getenv("FRONTEND_ORIGIN"),                   // 本番URL（環境変数から取得）
    },
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,   // HttpOnly Cookieの送受信に必須
    MaxAge:           12 * time.Hour,
  }))

  // ルート定義...
}
```

### 8c.2 環境変数追加

```env
# back/.env に追加
FRONTEND_ORIGIN=https://your-frontend-domain.com
```

### 8c.3 go.modへの追加

```bash
go get github.com/gin-contrib/cors
```

---

## 8d. .gitignore

### 8d.1 ルート（trading-system/.gitignore）

```gitignore
# 環境変数（絶対にコミットしない）
**/.env
**/.env.local
**/.env.*.local

# IDE
.idea/
.vscode/
*.swp
*.swo
```

### 8d.2 バックエンド（back/.gitignore）

```gitignore
# ビルド成果物
tmp/
bootstrap
lambda-handler.zip

# 環境変数
.env

# テストカバレッジ
coverage.out
coverage.html
```

### 8d.3 フロントエンド（front/.gitignore）

```gitignore
# Next.js
.next/
out/

# 依存関係
node_modules/

# 環境変数
.env.local
.env.*.local

# 自動生成ファイル（生成コマンドで再生成できるためコミット不要）
src/types/api.ts

# ビルド成果物
dist/
```

### 8d.4 Lambda（lambda/.gitignore）

```gitignore
# 環境変数
.env

# Python
__pycache__/
*.pyc
*.pyo
.pytest_cache/
package/
lambda-python.zip

# 仮想環境
venv/
.venv/
```

---

## 9. ディレクトリ構成（全体）

```
  ├── front/
  │   ├── public/
  │   ├── src/
  │   │   ├── app/
  │   │   ├── components/
  │   │   │   └── elements/
  │   │   ├── types/
  │   │   │   └── api.ts          # 自動生成
  │   │   └── utils/
  │   ├── .prettierrc
  │   ├── eslint.config.mjs
  │   ├── next.config.ts
  │   ├── tsconfig.json
  │   └── package.json
  │
  ├── back/
  │   ├── main.go                  # DI・Lambda対応
  │   ├── openapi.yaml             # API仕様（フロントと共有）
  │   ├── Makefile
  │   ├── Dockerfile
  │   ├── docker-compose.yml
  │   ├── .air.toml
  │   ├── .env / .env.example
  │   ├── domain/                  # ドメイン層
  │   ├── usecase/                 # ユースケース層
  │   ├── repository/              # リポジトリ実装
  │   ├── controller/              # HTTPハンドラ
  │   ├── router/                  # ルーティング
  │   ├── model/                   # GORMモデル
  │   ├── internal/api/            # oapi-codegen自動生成
  │   ├── migrate/                 # マイグレーション
  │   ├── db/                      # DB接続
  │   ├── utils/                   # 共通処理
  │   ├── go.mod
  │   └── go.sum
  │
  ├── lambda/
  │   ├── fetch_price.py
  │   ├── requirements.txt
  │   └── .env
  │
  └── docs/
      ├── spec.md
      ├── development_manual.md
      ├── feature_01_login.md
      └── ...
```

---

## 10. ローカル開発環境セットアップ

### 10.1 概要

budgetリポジトリと同様の構成を踏襲する。**ローカル開発では `docker compose` にローカル用のPostgreSQLコンテナ（`db`）を同梱し、ローカル完結で動作確認できる構成**とする（ISSUE #2 の方針）。本番・ステージングではNeon（クラウド）を使用し、接続先は環境変数 `DATABASE_URL` / `NEON_DATABASE_URL` で切り替える。

| サービス | 起動方法 | ポート |
|---------|---------|--------|
| フロントエンド（Next.js） | ローカル（npm run dev） | 3000 |
| バックエンド（Go / Gin） | Docker（budgetと同様） | 8080 |
| DB（ローカル） | Docker（postgres:16-alpine の `db` サービス） | 5432 |
| DB（本番/ステージング） | Neon（クラウド・常時起動） | - |
| Python Lambda | ローカル手動実行 | - |

> **DB接続の切り替え**: アプリは `DATABASE_URL` を優先し、未設定時に `NEON_DATABASE_URL` を使用する。ローカルは `db` サービスを指す `DATABASE_URL`（`.env.example` 参照）、Neon利用時は `DATABASE_URL` を空にして `NEON_DATABASE_URL` を設定する。

---

### 10.2 必要なツール

```bash
# Docker Desktop
docker --version
docker compose version

# Node.js（v20以上）
node -v

# Python（3.11以上）※ Lambdaローカルテスト用
python3 --version

# oapi-codegen（OpenAPI→Goコード生成）
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Claude Code CLI
npm install -g @anthropic-ai/claude-code
```

---

### 10.3 バックエンドのDocker構成

budgetの `back/Dockerfile` と `back/docker-compose.yml` を参考に構成。**ローカル開発用にPostgreSQLコンテナ（`db`）を同梱する**（ISSUE #2 の方針）。`app` は `db` の healthcheck 完了後に起動し、`DATABASE_URL` で `db` サービスに接続する。Neon利用時は `db` を起動せず `DATABASE_URL` / `NEON_DATABASE_URL` をNeonのURLに差し替える。実際の構成は `back/docker-compose.yml`・`back/docker-compose.dev.yml` を参照。

#### back/Dockerfile

budgetのDockerfileをそのまま踏襲（マルチステージビルド）。

```dockerfile
# back/Dockerfile
FROM golang:1.22 AS builder

RUN mkdir /app
WORKDIR /app

ENV GO111MODULE=on

COPY . .

RUN go mod tidy
RUN go build -o main .

FROM alpine AS dev
WORKDIR /app
COPY --from=builder /app/main /app/main
EXPOSE 8080

CMD ["./main"]
```

#### back/docker-compose.yml

budgetのdocker-compose.ymlからDBコンテナを除いた構成。

```yaml
# back/docker-compose.yml
version: "3.8"
services:
  app:
    build: .
    ports:
      - "8080:8080"
    env_file: ./.env
    environment:
      # Neon DBの接続URLは .env から読み込む
      DATABASE_URL: ${NEON_DATABASE_URL}
    networks:
      - backend
    volumes:
      # ホットリロード用にソースをマウント（airを使う場合）
      - .:/app
      - /app/tmp

networks:
  backend:
```

#### back/docker-compose.dev.yml（airホットリロード開発用）

```yaml
# back/docker-compose.dev.yml
version: "3.8"
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    env_file: ./.env
    environment:
      DATABASE_URL: ${NEON_DATABASE_URL}
    volumes:
      - .:/app
      - /app/tmp
    networks:
      - backend

networks:
  backend:
```

#### back/Dockerfile.dev（airホットリロード開発用）

```dockerfile
# back/Dockerfile.dev
FROM golang:1.22

RUN mkdir /app
WORKDIR /app

ENV GO111MODULE=on

# airのインストール
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
```

---

### 10.4 フロントエンドのDocker構成（任意）

budgetのフロントはDockerfileなし（ローカルで`npm run dev`のみ）。同様にローカル起動を基本とする。

```bash
# front/ はDockerを使わずローカルで起動
cd front
npm install
npm run generate:api   # openapi.yaml → src/types/api.ts 生成
npm run dev            # Turbopack起動
# → http://localhost:3000
```

---

### 10.5 起動手順

#### ① 環境変数の設定

```bash
cd back
cp .env.example .env
# .env を編集してNeon DBのURLなどを設定
```

```env
# back/.env
NEON_DATABASE_URL=postgresql://user:pass@host/dbname?sslmode=require
APP_ENV=development
JWT_SECRET=your_jwt_secret_key
JWT_ACCESS_EXPIRE_HOURS=24
JWT_REFRESH_EXPIRE_DAYS=30
ANTHROPIC_API_KEY=sk-ant-xxxxx
LINE_CHANNEL_ACCESS_TOKEN=xxxxx
LINE_USER_ID=xxxxx
AWS_ACCESS_KEY_ID=xxxxx
AWS_SECRET_ACCESS_KEY=xxxxx
AWS_REGION=ap-northeast-1
S3_BUCKET_NAME=trading-system-learning
INTERNAL_API_SECRET=xxxxx
```

#### ② バックエンド起動（Docker）

```bash
# 通常起動（本番に近い環境）
cd back
docker compose up --build

# airホットリロード起動（開発推奨）
docker compose -f docker-compose.dev.yml up --build
```

#### ③ DBマイグレーション

```bash
# コンテナ起動後にマイグレーション実行
docker compose exec app go run migrate/migrate.go

# または Makefile で
docker compose exec app make migrate
```

#### ④ フロントエンド起動

```bash
cd front
npm install
npm run generate:api
npm run dev
# → http://localhost:3000
```

#### ⑤ Python Lambda（ローカルテスト）

```bash
cd lambda
pip install -r requirements.txt
# .env を設定してローカルのGoバックエンドに向ける
GO_API_BASE_URL=http://localhost:8080 python fetch_price.py
```

---

### 10.6 よく使うMakefileコマンド（back/Makefile）

budgetリポジトリのMakefileを踏襲・拡張。

```makefile
# back/Makefile

# openapi.yaml → Goコード自動生成（budgetと同様）
generate-go:
	oapi-codegen -generate types -o internal/api/types.gen.go -package api openapi.yaml
	oapi-codegen -generate gin-server -o internal/api/server.gen.go -package api openapi.yaml

# DBマイグレーション実行
migrate:
	go run migrate/migrate.go

# Docker通常起動
up:
	docker compose up --build

# Dockerホットリロード起動（開発）
dev:
	docker compose -f docker-compose.dev.yml up --build

# Docker停止
down:
	docker compose down

# AWS Lambda向けGoバイナリビルド（budgetと同様）
build-lambda:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap -ldflags="-s -w" .
	rm -f lambda-handler.zip
	zip lambda-handler.zip bootstrap
	rm bootstrap
```

---

### 10.7 開発時の起動順序まとめ

```
1. Neon DB（クラウド・常時起動・操作不要）

2. バックエンド起動
   cd back
   docker compose -f docker-compose.dev.yml up --build
   → http://localhost:8080

3. マイグレーション（初回のみ）
   docker compose exec app make migrate

4. フロントエンド起動
   cd front && npm run generate:api && npm run dev
   → http://localhost:3000

5. Python Lambda（動作確認時のみ）
   cd lambda && python fetch_price.py
```

---

## 11. CI/CD・自動デプロイ

### 11.1 概要

budgetリポジトリの `.github/workflows/deploy-lambda.yml` を参考に、このシステム用に2つのワークフローを構成する。

| ワークフロー | トリガー | 対象 |
|------------|---------|------|
| `deploy-backend.yml` | `main`ブランチへのpush（`back/**`変更時） | GoバックエンドをAWS Lambdaにデプロイ |
| `deploy-lambda-python.yml` | `main`ブランチへのpush（`lambda/**`変更時） | Python株価取得をAWS Lambdaにデプロイ |

### 11.2 AWSの事前設定（OIDC認証）

budgetと同様に **GitHub Actions OIDC** を使用してAWSに認証する（アクセスキーをSecretsに保存しない安全な方式）。

#### GitHub Secretsに登録する値

| Secret名 | 内容 |
|---------|------|
| `AWS_ROLE_LAMBDA_DEPLOY` | IAMロールのARN（例: `arn:aws:iam::123456789:role/github-actions-role`） |
| `AWS_REGION` | AWSリージョン（例: `ap-northeast-1`） |
| `LAMBDA_FUNCTION_NAME_BACKEND` | GoバックエンドのLambda関数名 |
| `LAMBDA_FUNCTION_NAME_PYTHON` | Python株価取得のLambda関数名 |

### 11.3 Goバックエンド デプロイワークフロー

budgetの `deploy-lambda.yml` をそのまま踏襲。

```yaml
# .github/workflows/deploy-backend.yml
name: Deploy Backend to AWS Lambda

on:
  push:
    branches:
      - main
    paths:
      - 'back/**'      # back/ 配下の変更時のみ実行

permissions:
  id-token: write      # OIDC認証に必要
  contents: read

jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Configure AWS Credentials（OIDC）
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_LAMBDA_DEPLOY }}
          aws-region: ${{ secrets.AWS_REGION }}

      - name: Build and Zip Go application for Lambda
        working-directory: ./back
        run: |
          make build-lambda
          # → back/lambda-handler.zip が生成される

      - name: Update Lambda function code
        run: |
          aws lambda update-function-code \
            --function-name ${{ secrets.LAMBDA_FUNCTION_NAME_BACKEND }} \
            --zip-file fileb://back/lambda-handler.zip
```

**デプロイの流れ：**
```
mainブランチにpush（back/**変更）
  ↓
GitHub Actions起動
  ↓
OIDCでAWSに認証（アクセスキー不要）
  ↓
make build-lambda
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build → bootstrap
  zip lambda-handler.zip bootstrap
  ↓
aws lambda update-function-code
  ↓
Lambda関数が新バイナリで更新される
```

### 11.4 Python株価取得 デプロイワークフロー

Python Lambdaはbudgetにはない今回のシステム固有のワークフロー。

```yaml
# .github/workflows/deploy-lambda-python.yml
name: Deploy Python Lambda（株価取得）

on:
  push:
    branches:
      - main
    paths:
      - 'lambda/**'    # lambda/ 配下の変更時のみ実行

permissions:
  id-token: write
  contents: read

jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'

      - name: Configure AWS Credentials（OIDC）
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_LAMBDA_DEPLOY }}
          aws-region: ${{ secrets.AWS_REGION }}

      - name: Install dependencies and Zip
        working-directory: ./lambda
        run: |
          pip install -r requirements.txt -t ./package
          cp fetch_price.py ./package/
          cd package && zip -r ../lambda-python.zip .

      - name: Update Lambda function code
        run: |
          aws lambda update-function-code \
            --function-name ${{ secrets.LAMBDA_FUNCTION_NAME_PYTHON }} \
            --zip-file fileb://lambda/lambda-python.zip
```

**デプロイの流れ：**
```
mainブランチにpush（lambda/**変更）
  ↓
GitHub Actions起動
  ↓
OIDCでAWSに認証
  ↓
pip install -r requirements.txt -t ./package
  yfinance・requests・pandas を package/ に展開
  ↓
fetch_price.py を package/ にコピー
zip -r lambda-python.zip package/
  ↓
aws lambda update-function-code
  ↓
Lambda関数が更新される
```

### 11.5 ファイル配置

```
trading-system/
  ├── .github/
  │   └── workflows/
  │       ├── deploy-backend.yml         # GoバックエンドのLambdaデプロイ
  │       └── deploy-lambda-python.yml   # Python株価取得のLambdaデプロイ
  ├── back/
  │   ├── Makefile                       # build-lambdaコマンド含む
  │   └── ...
  └── lambda/
      ├── fetch_price.py
      └── requirements.txt
```

### 11.6 IAMロール設定（AWSコンソールで実施）

GitHub ActionsがOIDCでAWSに認証するために必要なIAMロールの設定。

```json
// IAMロールの信頼ポリシー
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::アカウントID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
          "token.actions.githubusercontent.com:sub": "repo:あなたのGitHubユーザー名/trading-system:ref:refs/heads/main"
        }
      }
    }
  ]
}

// アタッチするポリシー（最小権限）
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:UpdateFunctionCode"
      ],
      "Resource": [
        "arn:aws:lambda:ap-northeast-1:アカウントID:function:バックエンド関数名",
        "arn:aws:lambda:ap-northeast-1:アカウントID:function:Python関数名"
      ]
    }
  ]
}
```

### 11.7 デプロイ確認手順

```bash
# mainブランチへのpush後、GitHub Actionsの実行を確認
# https://github.com/あなたのユーザー名/trading-system/actions

# AWSコンソールでLambdaの更新日時を確認
aws lambda get-function --function-name バックエンド関数名 \
  --query 'Configuration.LastModified'
```

