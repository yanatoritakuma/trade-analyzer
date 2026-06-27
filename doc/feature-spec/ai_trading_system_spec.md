# AI株式トレーディングシステム
## 要件定義・機能仕様書

**バージョン**: 2.6  
**作成日**: 2026-06-06  
**更新内容**: ①一般ユーザーを閲覧のみ（全書き込みはadmin専用・trades記録はadminのみ）に変更 ②バーチャル売買の約定単価・数量ロジック（最新終値・単元100株・FIFO決済）を5.5に新設 ③ウォッチリスト書き込みをadmin専用・パス統一 ④`/internal/watchlist` を追加 ⑤実行時刻を平日15:30のみに統一 ⑥取得日数を120日に統一 ⑦learning_versions旧DDL（user_id付き）を削除 ⑧positions/プロフィール系APIをadmin配下で整理 ⑨崩れていた5.4の節構成を5.5として再編  
**ステータス**: Draft

---

## 目次

1. [プロジェクト概要](#1-プロジェクト概要)
2. [システムアーキテクチャ](#2-システムアーキテクチャ)
3. [技術スタック](#3-技術スタック)
4. [認証・認可仕様](#4-認証認可仕様)
5. [機能要件](#5-機能要件)
6. [データベース設計](#6-データベース設計)
7. [API仕様（Claude API・LINE API）](#7-api仕様)
8. [Go / Gin バックエンドAPI仕様](#8-go--gin-バックエンドapi仕様)
9. [Lambda関数仕様（Python）](#9-lambda関数仕様)
10. [通知仕様](#10-通知仕様)
11. [学習サイクル仕様](#11-学習サイクル仕様)
12. [非機能要件](#12-非機能要件)
13. [コスト試算](#13-コスト試算)
14. [フェーズ計画](#14-フェーズ計画)
15. [リスクと制約](#15-リスクと制約)

---

## 1. プロジェクト概要

### 1.1 目的

Claude AIを活用して日本株市場を自動分析し、売買シグナルをLINE通知で提供するシステムを構築する。バーチャルモードと実運用モードを並行稼働させ、週次の学習サイクルによって分析精度を継続的に向上させることを目的とする。

### 1.2 背景・課題

- 毎日の株価チェックと売買判断に時間と労力がかかる
- 感情に左右されず、ルールベースで売買判断を行いたい
- 過去のトレード結果を活かして判断精度を向上させたい
- 自動発注は行わず、最終判断は人間が行う安全設計にしたい

### 1.3 スコープ

| 対象 | 内容 |
|------|------|
| 対象市場 | 東京証券取引所（日本株） |
| 対象銘柄 | ウォッチリスト登録銘柄（初期は手動設定） |
| 売買実行 | **通知のみ（自動発注は行わない）** |
| 運用モード | バーチャル（シミュレーション） / 実運用（通知ベース） |

---

## 2. システムアーキテクチャ

### 2.1 全体構成図

```
【定期実行】平日 15:30 JST（引け後・1日1回）
                    ↓
          AWS EventBridge（スケジューラ）
                    ↓
          AWS Lambda（Python）
          yfinanceで株価データ取得
                    ↓ HTTP POST（株価データ付き）
          Go / Gin（/internal/stock-prices）
          受け取ったデータをそのままClaude APIに分析依頼
                    ↓
     ┌──────────────────────────┐
     ↓                          ↓
Neon PostgreSQL            LINE Messaging API
（株価・分析結果・履歴を保存）（BUY/SELLシグナル通知）
```

### 2.2 週次学習サイクル構成図

```
毎週日曜 18:00
     ↓
AWS EventBridge
     ↓
AWS Lambda（Python）
  週次レポートトリガー用途のみ
     ↓ HTTP POST
Go / Gin（/internal/weekly-report）
     ↓
Neon DBからtradesを直近26週集計（SQL）
     ↓
S3から learning_vN.csv 読み込み
     ↓
Claude APIでCSV自己最適化
     ↓
learning_v(N+1).csv をS3に保存
     ↓
LINE通知（週次サマリー）
```

---

## 3. 技術スタック

### 3.1 レイヤー構成

```
┌─────────────────────────────────────┐
│  Next.js（フロントエンド）            │
│  - ダッシュボード表示                 │
│  - ウォッチリスト管理                 │
│  - トレード履歴・損益閲覧             │
│  - 週次レポート表示                   │
└──────────────┬──────────────────────┘
               ↓ REST API
┌──────────────────────────────────────┐
│  Go / Gin（バックエンド）             │
│  - 認証・認可                        │
│  - 株価データ受取 → Claude API分析   │
│  - 分析結果・トレード履歴をDB保存     │
│  - ポートフォリオ管理                 │
│  - S3のCSV読み書き                   │
│  - LINE通知送信                      │
│  - フロントへのAPI提供               │
└──────────────────────────────────────┘
               ↑ HTTP POST（株価データ付き）
┌──────────────────────────────────────┐
│  Python / Lambda（株価取得専用）      │
│  - 定期実行トリガー（EventBridge）   │
│  - yfinanceで株価データ取得          │
│  - GoのAPIエンドポイントにPOST       │
│  ※ DB操作・分析・通知は行わない      │
└──────────────────────────────────────┘
```

### 3.2 技術一覧

| レイヤー | 採用技術 | 選定理由 | 費用 |
|---------|---------|---------|------|
| フロントエンド | Next.js（TypeScript） | UIダッシュボード構築 | 無料 |
| バックエンド | Go / Gin | 高速・型安全・DB操作・AI分析・通知 | 無料 |
| 株価取得 | Python 3.11 / Lambda | yfinanceがPython専用ライブラリ | 無料枠内 |
| 定期実行 | AWS EventBridge | Lambda連携・Cron設定 | 無料枠内 |
| 株価取得ライブラリ | yfinance | 日本株対応・無料・簡単 | 無料 |
| AI分析エンジン | Claude API（claude-sonnet-4-20250514） | 高精度分析・JSON出力 | 月$3〜5 |
| データベース | Neon（Serverless PostgreSQL） | 無料枠十分・SQL集計可能 | 無料 |
| 学習CSVストレージ | AWS S3 | バージョン管理・低コスト | ほぼ無料 |
| 通知 | LINE Messaging API | 普及率高く確認しやすい | 無料 |
| 開発環境 | Claude Code（Proプラン） | AIアシスト開発・開発期間のみ | $20/月 |

### 3.1 Neon無料枠の内訳

| リソース | 無料枠 | 今回の利用想定 |
|---------|--------|-------------|
| ストレージ | 0.5GB/プロジェクト | 数年分のデータでも余裕 |
| コンピュート | 100 CU-hours/月 | 1日3回のバッチ処理で十分 |
| スケールtoゼロ | 5分後に自動停止 | Lambdaバッチ処理と相性◎ |

---

## 4. 認証・認可仕様

### 4.1 概要

- ログインしていないユーザーはダッシュボードを含む全画面・全APIにアクセスできない
- アカウント作成は**招待コード制**（管理者が発行したコードを知っている者のみ登録可能）
- ユーザーは**管理者（admin）**と**一般ユーザー（user）**の2ロールに分かれる
- データは**ユーザーごとに完全分離**（他ユーザーのデータは参照・操作不可）
- 認証にはJWT（JSON Web Token）を採用し、Go / Gin バックエンドで発行・検証する

### 4.2 ロール定義

| ロール | 説明 | 付与タイミング |
|--------|------|-------------|
| `admin` | 全機能 + 管理者画面にアクセス可能（作成・更新・削除すべて可） | 初期データ投入（DB直接） |
| `user` | **閲覧のみ**。各画面の表示（GET）はできるが、作成・更新・削除は一切不可 | 招待コードで登録時に自動付与 |

#### 権限ポリシー（重要）

本システムは**全ユーザー共通の単一ポートフォリオ**を管理者が運用し、一般ユーザーはそれを閲覧する構成とする。

| 操作種別 | `admin` | `user` |
|---------|---------|--------|
| 参照（GET：ダッシュボード・トレード・ポートフォリオ・レポート・ウォッチリスト閲覧など） | ✅ | ✅ |
| 作成・更新・削除（POST/PUT/PATCH/DELETE：ウォッチリスト管理・保有株管理・プロフィール/パスワード変更・分析設定・候補承認など） | ✅ | ❌ |
| 管理者画面（/admin/*） | ✅ | ❌ |

- 一般ユーザーが閲覧する各画面のデータは、**管理者のデータ（全ユーザー共通の仮想ポートフォリオ）**を表示する。
- バックエンドでは **GET以外のメソッドは原則 `RequireAdmin` を必須**とする（認証系の `/api/auth/*` を除く）。
- フロントは `user` の場合、追加・編集・削除等の操作UI（ボタン・フォーム）を非表示にする。

### 4.3 招待コードフロー

```
【招待コード発行】
管理者 → 管理者画面（/admin/invitations）
       → 「招待コード発行」ボタン
       → POST /api/admin/invitations（Go / Gin）
       → ランダムな招待コード生成（例: TRADE-XXXX-XXXX）
       → invitation_codesテーブルに保存（有効期限: 7日）
       → 管理者が招待コードを対象者に連絡（メール・LINEなど手動）

【招待コードで登録】
新規ユーザー → /register（招待コード入力フォーム）
             → POST /api/auth/register（招待コード + メール + パスワード + 名前）
             → 招待コードの有効性を検証（期限・使用済みチェック）
             → NG → エラー返却（「無効な招待コードです」）
             → OK → usersテーブルに保存（role: "user"）
                  → 招待コードを使用済みにマーク
                  → 登録完了
```

### 4.4 認証フロー

```
【ログイン】
ユーザー → /login（メール・パスワード入力）
         → POST /api/auth/login（Go / Gin）
         → パスワード照合（bcrypt）
         → JWTアクセストークン発行（有効期限: 24時間）
         → JWTリフレッシュトークン発行（有効期限: 30日）
         → トークンをHttpOnly Cookieにセット
         → roleに応じてリダイレクト
           admin → /admin
           user  → /（ダッシュボード）

【認証済みリクエスト】
ユーザー → Next.js（各画面）
         → APIリクエスト（Cookieにトークン付与）
         → Go / Gin ミドルウェアでJWT検証
         → 検証OK → role確認 → 処理続行
         → 検証NG → 401 Unauthorized → /loginにリダイレクト

【トークンリフレッシュ】
アクセストークン期限切れ
         → POST /api/auth/refresh
         → リフレッシュトークンで新アクセストークン発行
         → 透過的に再試行

【ログアウト】
ユーザー → POST /api/auth/logout
         → CookieのトークンをクリアDelete
```

### 4.5 画面アクセス制御（Next.js）

| 画面 | パス | 認証 | ロール | 備考 |
|------|------|------|--------|------|
| ログイン | `/login` | 不要 | 全員 | - |
| 新規登録 | `/register` | 不要 | 全員（招待コード必須） | - |
| ダッシュボード | `/` | **必要** | user / admin | userは閲覧のみ |
| ウォッチリスト | `/watchlist` | **必要** | user / admin | userは閲覧のみ |
| トレード履歴 | `/trades` | **必要** | user / admin | userは閲覧のみ |
| ポートフォリオ | `/portfolio` | **必要** | user / admin | userは閲覧のみ |
| 週次レポート | `/reports` | **必要** | user / admin | userは閲覧のみ |
| 設定 | `/settings` | **必要** | **admin のみ** | 保有株・プロフィール管理は書き込みのためadmin専用 |
| 管理者ダッシュボード | `/admin` | **必要** | **admin のみ** | - |
| ユーザー管理 | `/admin/users` | **必要** | **admin のみ** | - |
| 招待コード管理 | `/admin/invitations` | **必要** | **admin のみ** | - |
| ウォッチリスト管理 | `/admin/watchlist` | **必要** | **admin のみ** | 銘柄の追加・削除 |
| 分析設定 | `/admin/analysis-settings` | **必要** | **admin のみ** | - |
| テーマ管理 | `/admin/analysis-settings/themes` | **必要** | **admin のみ** | - |
| ウォッチリスト候補承認 | `/admin/watchlist-candidates` | **必要** | **admin のみ** | - |

Next.jsの `middleware.ts` でトークンの有無・roleを検証し、権限外アクセスは適切な画面にリダイレクトする。一般ユーザー（user）に表示する画面は**すべて閲覧のみ**で、追加・編集・削除のUIは表示しない。

### 4.6 APIアクセス制御（Go / Gin）

```go
// 認証不要
router.POST("/api/auth/register", authHandler.Register)
router.POST("/api/auth/login",    authHandler.Login)
router.POST("/api/auth/refresh",  authHandler.Refresh)

// 認証必要・閲覧（GET）専用：user / admin 共通
// 一般ユーザーが利用できるのはこの参照系のみ
protected := router.Group("/api")
protected.Use(middleware.JWTAuth())
{
    protected.GET("/watchlist",          watchlistHandler.GetAll)   // 閲覧のみ
    protected.GET("/trades",             tradeHandler.GetAll)
    protected.GET("/positions",          positionHandler.GetAll)
    protected.GET("/portfolio/summary",  portfolioHandler.Summary)
    protected.GET("/reports",            reportHandler.GetAll)
    protected.GET("/reports/:week",      reportHandler.GetByWeek)
    protected.GET("/analysis/latest",    analysisHandler.Latest)
    protected.GET("/auth/me",            authHandler.Me)
    // 参照系（GET）のみ。作成・更新・削除は下記 admin グループに集約する。
}

// 管理者専用：作成・更新・削除はすべてここ（RequireAdmin必須）
admin := router.Group("/api/admin")
admin.Use(middleware.JWTAuth(), middleware.RequireAdmin())
{
    // ウォッチリスト管理（admin専用・パスは /api/admin/watchlist）
    admin.POST("/watchlist",                     watchlistHandler.Create)
    admin.DELETE("/watchlist/:id",               watchlistHandler.Delete)
    // 保有株管理（実運用・admin専用）
    admin.POST("/positions",                     positionHandler.Create)
    admin.PUT("/positions/:id",                  positionHandler.Update)
    admin.DELETE("/positions/:id",               positionHandler.Delete)
    // プロフィール・パスワード（admin専用：一般ユーザーは変更不可）
    admin.PATCH("/me",                           authHandler.UpdateProfile)
    admin.PUT("/me/password",                    authHandler.ChangePassword)
    // ユーザー・招待コード管理
    admin.GET("/users",                          adminHandler.GetUsers)
    admin.PATCH("/users/:id",                    adminHandler.UpdateUser)
    admin.DELETE("/users/:id",                   adminHandler.DeleteUser)
    admin.GET("/invitations",                    adminHandler.GetInvitations)
    admin.POST("/invitations",                   adminHandler.CreateInvitation)
    admin.DELETE("/invitations/:id",             adminHandler.DeleteInvitation)
    // 分析設定・テーマ
    admin.GET("/analysis-settings",              analysisHandler.GetSettings)
    admin.PUT("/analysis-settings",              analysisHandler.UpdateSettings)
    admin.GET("/analysis-themes",                themeHandler.GetAll)
    admin.POST("/analysis-themes",               themeHandler.Create)
    admin.PUT("/analysis-themes/:id",            themeHandler.Update)
    admin.DELETE("/analysis-themes/:id",         themeHandler.Delete)
    admin.PATCH("/analysis-themes/sort",         themeHandler.Sort)
    // 候補承認
    admin.GET("/watchlist-candidates",           candidateHandler.GetAll)
    admin.PATCH("/watchlist-candidates/:id/approve", candidateHandler.Approve)
    admin.PATCH("/watchlist-candidates/:id/reject",  candidateHandler.Reject)
}
```

> プロフィール・パスワード変更（`PATCH /api/admin/me`・`PUT /api/admin/me/password`）も、本仕様の「一般ユーザーは閲覧のみ」方針に従い admin 専用とする。一般ユーザーは自身のプロフィールも変更しない（必要時は管理者が `/api/admin/users/:id` で対応）。

### 4.7 認証・管理エンドポイント仕様

#### 認証系

| メソッド | パス | 説明 | 認証要否 |
|---------|------|------|---------|
| POST | `/api/auth/register` | 招待コードで新規登録 | 不要 |
| POST | `/api/auth/login` | ログイン・トークン発行 | 不要 |
| POST | `/api/auth/logout` | ログアウト・Cookie削除 | 必要 |
| POST | `/api/auth/refresh` | アクセストークン再発行 | 不要 |
| GET  | `/api/auth/me` | ログイン中ユーザー情報取得 | 必要 |

#### 管理者系

| メソッド | パス | 説明 |
|---------|------|------|
| POST   | `/api/admin/watchlist` | ウォッチリストに銘柄追加（最大3銘柄） |
| DELETE | `/api/admin/watchlist/:id` | ウォッチリストから銘柄削除 |
| POST   | `/api/admin/positions` | 実運用保有株を登録 |
| PUT    | `/api/admin/positions/:id` | 実運用保有株を更新 |
| DELETE | `/api/admin/positions/:id` | 実運用保有株を削除 |
| PATCH  | `/api/admin/me` | プロフィール更新（名前） |
| PUT    | `/api/admin/me/password` | パスワード変更 |
| GET    | `/api/admin/users` | ユーザー一覧取得 |
| PATCH  | `/api/admin/users/:id` | ユーザー情報更新（停止・復活など） |
| DELETE | `/api/admin/users/:id` | ユーザー削除 |
| GET    | `/api/admin/invitations` | 招待コード一覧取得 |
| POST   | `/api/admin/invitations` | 招待コード発行 |
| DELETE | `/api/admin/invitations/:id` | 招待コード無効化 |
| GET    | `/api/admin/analysis-settings` | 現在の分析設定取得 |
| PUT    | `/api/admin/analysis-settings` | 分析設定を保存（次回定期実行時に反映） |
| GET    | `/api/admin/analysis-themes` | テーマ一覧取得 |
| POST   | `/api/admin/analysis-themes` | テーマ追加 |
| PUT    | `/api/admin/analysis-themes/:id` | テーマ編集 |
| DELETE | `/api/admin/analysis-themes/:id` | テーマ削除 |
| PATCH  | `/api/admin/analysis-themes/sort` | テーマ並び替え |
| GET    | `/api/admin/watchlist-candidates` | 候補銘柄一覧取得 |
| PATCH  | `/api/admin/watchlist-candidates/:id/approve` | 候補銘柄を承認（ウォッチリストに追加） |
| PATCH  | `/api/admin/watchlist-candidates/:id/reject` | 候補銘柄を却下 |

#### POST /api/auth/register リクエスト

```json
{
  "invitation_code": "TRADE-XXXX-XXXX",
  "email": "user@example.com",
  "password": "password123",
  "name": "山田太郎"
}
```

#### POST /api/auth/login リクエスト／レスポンス

```json
// リクエスト
{ "email": "user@example.com", "password": "password123" }

// レスポンス
{
  "message": "ログインしました",
  "user": { "id": 1, "email": "user@example.com", "name": "山田太郎", "role": "user" }
}
```

※ アクセストークン・リフレッシュトークンはレスポンスボディではなく **HttpOnly Cookie** にセットする（XSS対策）。

### 4.8 管理者画面の機能

#### ユーザー管理（/admin/users）

| 機能 | 内容 |
|------|------|
| ユーザー一覧 | 名前・メール・ロール・登録日・ステータスを表示 |
| ユーザー停止 | is_active を false に変更（ログイン不可になる） |
| ユーザー復活 | is_active を true に戻す |
| ユーザー削除 | アカウントと紐づくデータを削除 |

#### 招待コード管理（/admin/invitations）

| 機能 | 内容 |
|------|------|
| 招待コード発行 | ランダムコード生成・有効期限設定（デフォルト7日） |
| 招待コード一覧 | コード・発行日・有効期限・使用状況を表示 |
| 招待コード無効化 | 発行済みコードを手動で無効化 |

### 4.9 パスワードバリデーション

| ルール | 内容 |
|--------|------|
| 最小文字数 | 8文字以上 |
| 必須文字種 | 英字・数字を各1文字以上含む |
| ハッシュ化 | bcrypt（コスト係数: 12） |
| 保存形式 | ハッシュのみ保存（平文は保存しない） |

### 4.10 JWTトークン仕様

| 項目 | アクセストークン | リフレッシュトークン |
|------|----------------|-------------------|
| 有効期限 | 24時間 | 30日 |
| 保存場所 | HttpOnly Cookie | HttpOnly Cookie |
| Cookieフラグ | Secure, SameSite=Strict | Secure, SameSite=Strict |
| Payload | user_id, email, role, exp | user_id, exp |

---

## 5. 機能要件

### 5.1 株価データ収集

| 項目 | 仕様 |
|------|------|
| データソース | yfinance（Yahoo Finance非公式ラッパー） |
| 実行タイミング | **平日 15:30（JST）のみ** |
| 対象銘柄数 | **最大3銘柄**（watchlistに登録された銘柄） |
| ティッカー形式 | 末尾に`.T`を付与（例：`6522.T`） |
| 取得データ | 始値・高値・安値・終値・出来高（OHLCV）、**過去120日分** |
| 派生指標 | RSI（14日）、MACD、移動平均（5日・25日・75日）、ボリンジャーバンド、一目均衡表、出来高移動平均、ATR |

### 5.2 AI分析エンジン

#### 5.2.1 Claude API呼び出し仕様

| 項目 | 値 |
|------|-----|
| モデル | `claude-sonnet-4-20250514` |
| max_tokens | 1,500 |
| 月額上限 | $5（Anthropicダッシュボードで設定） |
| 月間コスト見込み | 約$1.2（3銘柄×1回×20日） |

#### 5.2.2 システムプロンプト（毎回共通・変わらない）

AIの人格・ルール・前提知識を定義する。毎回同じ内容を渡す。

```
【役割定義】
あなたは日本株の短期トレードアナリストです。

【過去の学習データ】
S3の learning_vN.csv の内容をそのまま注入
（有効パターン・無効パターン・相場環境別戦略など）

【分析設定】（管理者がUIから設定した内容）
- テーマ: AI・半導体
- スクリーニング条件: 時価総額100億以上・出来高1万株以上
- 分析スタイル: 短期・順張り
- 自由プロンプト: 「材料株を重視してください」

【リスク管理ルール】
- 損切りラインの設定は必須
- 確信度0.7未満は必ず HOLD にすること
- ポジションサイズは資産の10%以内
```

#### 5.2.3 ユーザープロンプト（毎回変わる・その日のデータ）

毎回の具体的な分析依頼とデータを渡す。

```
以下の銘柄を分析してください。

【銘柄情報】
銘柄: アステリスク (6522.T)
現在値: 1,180円

【過去120日のOHLCVデータ】
2025-12-01, 1050, 1080, 1020, 1060, 30000
2025-12-02, 1060, 1100, 1050, 1090, 45000
...（120日分）

【テクニカル指標】
RSI(14): 28.5
MACD: -12.3 / シグナル: -8.1
移動平均 5日: 1,160 / 25日: 1,150 / 75日: 1,100
ボリンジャーバンド 上限: 1,280 / 下限: 1,080
ATR: 42.3

以下の形式でJSON出力してください。
```

#### 5.2.4 出力フォーマット（JSON）

```json
{
  "ticker": "6522.T",
  "action": "BUY",
  "confidence": 0.85,
  "buy_reasons": [
    "RSI28.5と過売り圏からの反発局面",
    "出来高が直近平均の1.8倍に急増",
    "ボリンジャーバンド下限でサポート確認"
  ],
  "no_buy_reasons": [
    "25日移動平均線がまだ下向き",
    "市場全体が軟調"
  ],
  "entry_condition": "翌日始値が1,150円以上で陽線確認後にエントリー",
  "target_price": 1350,
  "stop_loss": 1080,
  "position_size": 0.1,
  "reason": "RSI過売り圏からの反発と出来高急増が重なり短期反発を期待"
}
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| ticker | string | 銘柄コード（例：`6522.T`） |
| action | string | `BUY` / `SELL` / `HOLD` |
| confidence | float | 確信度（0.0〜1.0） |
| buy_reasons | array | 買う理由（複数） |
| no_buy_reasons | array | 買わない理由（複数） |
| entry_condition | string | エントリーの具体的条件 |
| target_price | number | 目標株価 |
| stop_loss | number | 損切り価格 |
| position_size | float | 推奨ポジションサイズ（資産比率） |
| reason | string | 総合判断コメント（日本語） |

### 5.3 分析設定機能

#### 5.3.1 概要

管理者がUIから定期実行（EventBridge）のClaude API分析プロンプトを設定する機能。設定した内容は次回以降の定期実行（**平日15:30**）に自動的に反映される。全ユーザー共通の設定。

テーマは**DBで管理**し、管理者がUIから自由に追加・削除・並び替えができる。

#### 5.3.2 設定項目

| 項目 | 内容 | 管理方法 |
|------|------|---------|
| テーマ | 注目セクター・テーマを複数選択 | **DBで管理・管理者がUI上で追加・削除可能** |
| スクリーニング条件 | 銘柄絞り込みの数値条件 | 設定画面から入力 |
| 分析スタイル | トレードの方向性・期間 | 設定画面から選択 |
| 自由プロンプト | 上記以外の任意指示 | 自由テキスト入力 |

#### 5.3.3 設定の反映タイミング

```
管理者 → /admin/analysis-settings（設定画面）
       → テーマ・条件・スタイル・自由プロンプトを設定
       → 「保存」ボタン
       → PUT /api/admin/analysis-settings（Go / Gin）
       → analysis_settingsテーブルに保存

次回の定期実行時（平日 15:30）
       → Goがanalysis_settingsとanalysis_themesを読み込む
       → 設定内容をシステムプロンプトに注入
       → 分析実行
```

**設定を保存した時点では分析は実行されない。次の定期実行（15:30）で自動反映される。**

#### 5.3.4 テーマ管理（analysis_themes）

テーマ一覧はDBで管理する。管理者はUIから自由に追加・削除・並び替えができる。

| 操作 | 内容 |
|------|------|
| テーマ追加 | テーマ名・説明・並び順を入力して追加 |
| テーマ削除 | 不要なテーマを削除（使用中の設定からも除外される） |
| 並び替え | ドラッグ&ドロップで表示順を変更 |
| 有効/無効 | テーマを一時的に非表示にする（削除せず保持） |

**初期データ（マイグレーション時に投入）：**

| テーマ名 | 説明 | 並び順 |
|---------|------|--------|
| AI・生成AI | AIソフト・LLM関連銘柄 | 1 |
| 半導体 | 製造装置・素材・設計 | 2 |
| 防衛・宇宙 | 防衛関連・宇宙ビジネス | 3 |
| データセンター | サーバー・冷却・電力インフラ | 4 |
| 物流・自動化 | 倉庫ロボット・配送効率化 | 5 |
| 脱炭素・EV | 再エネ・電池・EV関連 | 6 |
| 内需・インバウンド | 観光・飲食・小売 | 7 |

### 5.4 ウォッチリスト候補提案機能（AI自動取得）

#### 5.4.1 概要

毎日15:30の定期分析と同時にClaude APIが**ウォッチリスト候補銘柄**を提案する。管理者が承認するとウォッチリストに追加される（最大3銘柄を維持）。

#### 5.4.2 候補提案フロー

```
毎日 15:30 定期実行
  ↓
① 現在のウォッチリスト銘柄を分析（既存フロー）
  ↓
② Claude APIが同時に候補銘柄を提案
   - 分析設定（テーマ・スクリーニング条件）に合致する銘柄
   - 最大3件を提案（現在のウォッチリストと重複なし）
  ↓
③ 提案内容をwatchlist_candidatesテーブルに保存
  ↓
④ 管理者にLINE通知（候補銘柄あり）
  ↓
⑤ 管理者が /admin/watchlist-candidates 画面で確認
  ↓
⑥ 「承認」→ ウォッチリストに追加（既存銘柄を置き換え提案）
   「却下」→ 候補をキャンセル
```

#### 5.4.3 候補提案プロンプト（ユーザープロンプトに追加）

```
【ウォッチリスト候補の提案】
現在のウォッチリスト: 6522.T, 3993.T, 7203.T
上記の分析設定（テーマ・スクリーニング条件）に基づき、
現在のウォッチリスト以外で注目すべき銘柄を最大3件提案してください。

出力形式:
{
  "candidates": [
    {
      "ticker": "XXXX.T",
      "name": "銘柄名",
      "reason": "提案理由",
      "replace_suggestion": "6522.T",  // 既存のどの銘柄と置き換えを推奨するか（任意）
      "confidence": 0.80
    }
  ]
}
```

#### 5.4.4 置き換え提案ルール

| 条件 | 挙動 |
|------|------|
| ウォッチリストが3銘柄未満 | 空きスロットに追加提案 |
| ウォッチリストが3銘柄満杯 | 最も成績の低い銘柄との置き換えを提案 |
| 管理者が承認 | 対象銘柄を削除し候補銘柄を追加 |
| 管理者が却下 | 既存ウォッチリストを維持 |
| 通知タイミング | 候補提案時にLINEで通知 |

### 5.5 運用モードと売買記録ロジック

#### 5.5.1 記録対象ユーザー（全ユーザー共通の単一ポートフォリオ）

定期分析（平日15:30）は全ユーザー共通で**1回のみ**実行され、その結果生成されるBUY/SELLトレードは**管理者（admin）の `trades` にのみ記録**する（`user_id = admin`）。

一般ユーザー（role: `user`）は**閲覧のみ**で、トレード・保有株・各種設定を作成・更新・削除できない（[4.2 ロール定義・権限ポリシー](#42-ロール定義)参照）。ダッシュボード・トレード履歴・ポートフォリオ・週次レポートには**管理者のデータ（全ユーザー共通の仮想ポートフォリオ）を表示**する。

これにより「分析1回／全ユーザー共通」と「`trades` は `user_id` ごと」の不整合は解消される（記録先は常にadmin1人）。

#### 5.5.2 バーチャルモード

| 項目 | 仕様 |
|------|------|
| 初期仮想資金 | 1,000,000円（固定） |
| 売買記録 | AIのBUY/SELLシグナルに基づき管理者のtradesへ自動記録 |
| 損益計算 | 含み益＝最新終値（stock_prices.close）− 取得単価 |
| ポジション管理 | 同一銘柄の複数エントリー対応（FIFOで決済） |

#### 5.5.3 約定単価・数量の決定ロジック（バーチャル）

AI出力は `position_size`（資産比率）・`target_price`・`stop_loss` を返すが、約定単価・株数は返さないため、Go側で以下のルールにより確定する。

**【約定単価 `price`】**
- BUY・SELL とも、**分析実行時点（15:30）の最新終値 `stock_prices.close` を約定単価**とする。
- `entry_condition`（例:「翌日始値が1,150円以上で…」）はLINE通知に載せる**参考情報**であり、記録単価には用いない。

**【数量 `quantity`】**
- 日本株の売買単位に合わせ **1単元 = 100株** とする。
- 取得株数 = `floor( (1,000,000 × position_size) ÷ price ÷ 100 ) × 100`
- 算出結果が **0株（1単元に満たない）の場合はBUYをスキップ**（trades記録・通知とも行わない）。

**【BUY記録】**
- `trades` に `mode='virtual'`, `action='BUY'`, `ticker`, `name`, `price`, `quantity`, `confidence`, `reason`, `target_price`, `stop_loss` を保存（`user_id = admin`, `closed_at = NULL`）。
- `name` は分析対象の `watchlist.name` から取得して**記録時に保存**する（ウォッチリストから外れても履歴で銘柄名を表示できるようにするため）。

**【SELL記録・損益確定】**
- 同一銘柄の未決済BUYポジション（`closed_at IS NULL`）を**古い順（FIFO）に決済**する。
- 決済数量分について `result_pnl = (SELL price − BUY price) × 決済数量` を計算し、対象BUY行に `result_pnl`・`closed_at` を更新。
- SELL自体も `trades` に `action='SELL'` で記録する。
- 保有ポジションが無い銘柄のSELLシグナルは**記録しない**（バーチャルでは通知も行わない。実運用の通知条件は5.5.5参照）。

#### 5.5.4 実運用モード

| 項目 | 仕様 |
|------|------|
| 保有株登録 | real_positionsテーブルに手動登録（**管理者のみ**） |
| 売買実行 | **通知のみ・自動発注は行わない** |
| 損益計算 | 登録された取得単価（avg_price）と最新終値をもとに計算 |
| 更新タイミング | 売買後に管理者が手動更新 |

#### 5.5.5 両モード並行運用

- バーチャルと実運用は独立して管理する。
- SELLシグナルは、バーチャルは「保有ポジションあり」、実運用は「real_positionsに登録あり」を条件に通知する。
- 週次レポートで両モードの成績を比較し、乖離が大きい場合はAIが原因を分析する。

### 5.6 LINE通知

#### 5.6.1 通知トリガー条件

| トリガー | 条件 |
|---------|------|
| 買いシグナル | action = `BUY` かつ confidence ≥ 0.7 |
| 売りシグナル（バーチャル） | action = `SELL` かつ保有ポジションあり |
| 売りシグナル（実運用） | action = `SELL` かつ real_positionsに登録あり |
| ウォッチリスト候補提案 | AIが候補銘柄を提案した場合（承認待ち） |
| 週次レポート | 毎週日曜 18:00 |
| エラー通知 | Lambda実行エラー時 |

#### 5.6.2 通知しない条件

- action = `HOLD` の場合
- confidence < 0.7 の場合
- 祝日・市場休場日

### 5.7 週次レポート・学習機能

| 項目 | 仕様 |
|------|------|
| 実行タイミング | 毎週日曜 18:00 JST |
| 集計期間 | 直前の月曜〜金曜 |
| 集計内容 | トレード数・勝率・損益・最大ドローダウン |
| AI分析 | 成功・失敗パターンの分析、来週の戦略提案 |
| 保存先 | learning_logsテーブル |
| 学習反映 | 次回以降の分析プロンプトに直近4週分を注入 |

---

## 6. データベース設計

### 6.1 テーブル一覧

| テーブル名 | 用途 |
|-----------|------|
| users | ユーザーアカウント |
| invitation_codes | 招待コード管理 |
| watchlist | 監視銘柄リスト（全ユーザー共通・管理者が管理） |
| watchlist_candidates | AIが提案したウォッチリスト候補 |
| stock_prices | 最新株価スナップショット（全ユーザー共通・1銘柄1行・UPSERT・現在値/前日比のソース） |
| trades | トレード履歴（バーチャル・実運用共通） |
| analysis_logs | Claude APIの分析結果ログ（全ユーザー共通） |
| real_positions | 実運用の保有株 |
| learning_logs | 週次学習ログ（全ユーザー共通・管理者のtradesから生成） |
| learning_versions | 学習CSVのバージョン管理（全ユーザー共通） |
| analysis_settings | 分析設定（管理者が設定・全ユーザー共通） |
| analysis_themes | 分析テーマ一覧（管理者がUI上で管理） |

### 6.2 DDL

```sql
-- ユーザー
CREATE TABLE users (
  id            SERIAL PRIMARY KEY,
  email         VARCHAR(255) NOT NULL UNIQUE,
  name          VARCHAR(100) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role          VARCHAR(10)  NOT NULL DEFAULT 'user'
                  CHECK (role IN ('admin', 'user')),
  is_active     BOOLEAN      DEFAULT TRUE,
  created_at    TIMESTAMP    DEFAULT NOW(),
  updated_at    TIMESTAMP    DEFAULT NOW()
);

-- 招待コード
CREATE TABLE invitation_codes (
  id          SERIAL PRIMARY KEY,
  code        VARCHAR(20)  NOT NULL UNIQUE,    -- 例: TRADE-XXXX-XXXX
  created_by  INT          REFERENCES users(id),
  used_by     INT          REFERENCES users(id),
  expires_at  TIMESTAMP    NOT NULL,
  used_at     TIMESTAMP,
  is_active   BOOLEAN      DEFAULT TRUE,
  created_at  TIMESTAMP    DEFAULT NOW()
);

-- ウォッチリスト（全ユーザー共通・管理者が管理）
CREATE TABLE watchlist (
  id         SERIAL PRIMARY KEY,
  ticker     VARCHAR(10)  NOT NULL UNIQUE,       -- 例: 6522.T（重複不可）
  name       VARCHAR(100),
  mode       VARCHAR(10)  NOT NULL
               CHECK (mode IN ('virtual', 'real', 'both')),
  is_active  BOOLEAN      DEFAULT TRUE,
  added_at   TIMESTAMP    DEFAULT NOW()
);

-- 最新株価スナップショット（全ユーザー共通・1銘柄1行・現在値/前日比のソース）
CREATE TABLE stock_prices (
  id            SERIAL PRIMARY KEY,
  ticker        VARCHAR(10)   NOT NULL UNIQUE,    -- UPSERTキー（1銘柄1行）
  date          DATE          NOT NULL,           -- 最新終値の取引日
  open          NUMERIC(10,2) NOT NULL,
  high          NUMERIC(10,2) NOT NULL,
  low           NUMERIC(10,2) NOT NULL,
  close         NUMERIC(10,2) NOT NULL,           -- 現在値（最新終値）
  prev_close    NUMERIC(10,2),                    -- 前営業日終値
  change_amount NUMERIC(10,2),                    -- 前日比（円）
  change_rate   NUMERIC(6,2),                     -- 前日比（%）
  volume        BIGINT        NOT NULL DEFAULT 0,
  updated_at    TIMESTAMP     NOT NULL DEFAULT NOW()
);                                                -- 同一銘柄は ON CONFLICT (ticker) でUPSERT

-- トレード履歴
CREATE TABLE trades (
  id           SERIAL PRIMARY KEY,
  user_id      INT           NOT NULL REFERENCES users(id),
  ticker       VARCHAR(10)   NOT NULL,
  name         VARCHAR(100),                   -- 記録時にwatchlist.nameから非正規化保存（履歴表示用）
  mode         VARCHAR(10)   NOT NULL
                 CHECK (mode IN ('virtual', 'real')),
  action       VARCHAR(10)   NOT NULL
                 CHECK (action IN ('BUY', 'SELL')),
  price        NUMERIC(10,2) NOT NULL,
  quantity     INT           NOT NULL,
  confidence   NUMERIC(4,3),
  reason       TEXT,
  target_price NUMERIC(10,2),
  stop_loss    NUMERIC(10,2),
  result_pnl   NUMERIC(10,2),                  -- 決済後に更新
  closed_at    TIMESTAMP,                      -- 決済日時
  created_at   TIMESTAMP     DEFAULT NOW()
);

-- 分析結果ログ（全ユーザー共通・管理者設定によるLambda一括実行の結果）
CREATE TABLE analysis_logs (
  id          SERIAL PRIMARY KEY,
  ticker      VARCHAR(10) NOT NULL,
  action      VARCHAR(10),
  confidence  NUMERIC(4,3),
  analysis    JSONB,                           -- Claude APIの生出力
  analyzed_at TIMESTAMP   NOT NULL DEFAULT NOW()
);

-- 保有株（実運用モード）
CREATE TABLE real_positions (
  id         SERIAL PRIMARY KEY,
  user_id    INT           NOT NULL REFERENCES users(id),
  ticker     VARCHAR(10)   NOT NULL,
  name       VARCHAR(100),
  quantity   INT           NOT NULL,
  avg_price  NUMERIC(10,2) NOT NULL,
  updated_at TIMESTAMP     DEFAULT NOW(),
  UNIQUE (user_id, ticker)
);

-- 週次学習ログ（全ユーザー共通・管理者のtradesのみを集計）
CREATE TABLE learning_logs (
  id           SERIAL PRIMARY KEY,
  week_start   DATE        NOT NULL,
  week_end     DATE        NOT NULL,
  trade_count  INT,
  win_rate     NUMERIC(5,2),
  total_pnl    NUMERIC(10,2),
  summary      TEXT,                          -- 週次サマリー
  lessons      TEXT,                          -- 学習・反省
  strategy     TEXT,                          -- 来週の戦略
  raw_response TEXT,                          -- Claude APIの生テキスト
  created_at   TIMESTAMP   DEFAULT NOW()
);

-- 学習CSVバージョン管理（全ユーザー共通）
CREATE TABLE learning_versions (
  id          SERIAL PRIMARY KEY,
  version     INT          NOT NULL,
  s3_path     VARCHAR(255) NOT NULL,          -- 例: s3://bucket/learning_v3.csv
  week_range  VARCHAR(50),                    -- 例: 2025-12-01〜2026-05-25
  char_count  INT,
  created_at  TIMESTAMP    DEFAULT NOW()
);

-- 分析設定（管理者が設定・全ユーザー共通）
CREATE TABLE analysis_settings (
  id             SERIAL PRIMARY KEY,
  theme_ids      INT[],                       -- analysis_themesのIDを参照
  screening      JSONB,                       -- 例: {"min_market_cap": 10000000000, "min_volume": 10000}
  style          VARCHAR(50),                 -- 例: "short_term_contrarian"
  free_prompt    TEXT,                        -- 自由入力プロンプト
  is_active      BOOLEAN   DEFAULT TRUE,      -- 現在有効な設定
  created_by     INT       REFERENCES users(id),
  created_at     TIMESTAMP DEFAULT NOW(),
  updated_at     TIMESTAMP DEFAULT NOW()
);

-- 分析テーマ一覧（管理者がUI上で管理・追加・削除・並び替え可能）
CREATE TABLE analysis_themes (
  id          SERIAL PRIMARY KEY,
  name        VARCHAR(100) NOT NULL UNIQUE,   -- 例: "AI・生成AI"
  description VARCHAR(255),                   -- 例: "AIソフト・LLM関連銘柄"
  sort_order  INT          NOT NULL DEFAULT 0,-- 表示順
  is_active   BOOLEAN      DEFAULT TRUE,      -- 有効/無効
  created_by  INT          REFERENCES users(id),
  created_at  TIMESTAMP    DEFAULT NOW(),
  updated_at  TIMESTAMP    DEFAULT NOW()
);

-- ウォッチリスト候補（AIが提案・管理者が承認/却下）
CREATE TABLE watchlist_candidates (
  id                 SERIAL PRIMARY KEY,
  ticker             VARCHAR(10)  NOT NULL,   -- 例: 6522.T
  name               VARCHAR(100),
  reason             TEXT,                    -- 提案理由
  replace_ticker     VARCHAR(10),             -- 置き換え推奨銘柄（任意）
  confidence         NUMERIC(4,3),            -- 確信度
  status             VARCHAR(10)  NOT NULL DEFAULT 'pending'
                       CHECK (status IN ('pending', 'approved', 'rejected')),
  proposed_at        TIMESTAMP    DEFAULT NOW(),
  decided_at         TIMESTAMP,               -- 承認/却下日時
  decided_by         INT          REFERENCES users(id)
);
```

### 6.3 インデックス

```sql
CREATE INDEX idx_users_email                 ON users (email);
CREATE INDEX idx_invitation_codes_code       ON invitation_codes (code);
CREATE INDEX idx_watchlist_candidates_status ON watchlist_candidates (status);
-- stock_prices は ticker のUNIQUE制約が索引を兼ねる（UPSERT・参照キー）
CREATE INDEX idx_trades_user_id              ON trades (user_id);
CREATE INDEX idx_trades_ticker               ON trades (ticker);
CREATE INDEX idx_trades_mode                 ON trades (mode);
CREATE INDEX idx_trades_created_at           ON trades (created_at);
CREATE INDEX idx_analysis_logs_ticker        ON analysis_logs (ticker);
CREATE INDEX idx_analysis_logs_analyzed_at   ON analysis_logs (analyzed_at);
CREATE INDEX idx_analysis_themes_sort_order  ON analysis_themes (sort_order);
```

---

## 7. API仕様

### 7.1 Claude API呼び出し仕様

| 項目 | 値 |
|------|-----|
| エンドポイント | `https://api.anthropic.com/v1/messages` |
| モデル | `claude-sonnet-4-20250514` |
| max_tokens | 1,500（分析JSON＋ウォッチリスト候補提案を含むため・5.2.1と統一） |
| 月額上限 | $5（Anthropicダッシュボードで設定） |

#### リクエスト構造

```python
{
  "model": "claude-sonnet-4-20250514",
  "max_tokens": 1500,
  "system": "<システムプロンプト（学習ログ含む）>",
  "messages": [
    {
      "role": "user",
      "content": "<銘柄・テクニカルデータ・JSON出力指示>"
    }
  ]
}
```

### 7.2 LINE Messaging API仕様

| 項目 | 値 |
|------|-----|
| 使用API | LINE Messaging API（Push Message） |
| 認証 | Channel Access Token |
| 送信先 | 個人のLINE User ID |
| 形式 | テキストメッセージ |

---

## 8. Go / Gin バックエンドAPI仕様

### 8.1 エンドポイント一覧

#### フロントエンド向け・参照系（認証必要・user / admin 共通）

一般ユーザーが利用できるのはこの **GET（閲覧）のみ**。

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/api/watchlist` | ウォッチリスト取得（閲覧） |
| GET | `/api/trades` | トレード履歴取得 |
| GET | `/api/trades?mode=virtual` | バーチャル履歴取得 |
| GET | `/api/trades?mode=real` | 実運用履歴取得 |
| GET | `/api/positions` | 実運用保有株取得 |
| GET | `/api/portfolio/summary` | 損益サマリー取得 |
| GET | `/api/reports` | 週次レポート一覧取得 |
| GET | `/api/reports/:week` | 週次レポート詳細取得 |
| GET | `/api/analysis/latest` | 最新分析結果取得 |
| GET | `/api/auth/me` | ログイン中ユーザー情報取得 |

#### フロントエンド向け・更新系（**admin のみ**）

ウォッチリスト・保有株・プロフィール/パスワードの作成・更新・削除は `/api/admin/*` に集約し、`RequireAdmin` を必須とする（[4.6](#46-apiアクセス制御go--gin)・[4.7](#47-認証管理エンドポイント仕様)参照）。

| メソッド | パス | 説明 |
|---------|------|------|
| POST   | `/api/admin/watchlist` | 銘柄追加（最大3銘柄） |
| DELETE | `/api/admin/watchlist/:id` | 銘柄削除 |
| POST   | `/api/admin/positions` | 保有株登録 |
| PUT    | `/api/admin/positions/:id` | 保有株更新 |
| DELETE | `/api/admin/positions/:id` | 保有株削除 |
| PATCH  | `/api/admin/me` | プロフィール更新 |
| PUT    | `/api/admin/me/password` | パスワード変更 |

#### Lambda（Python）からの内部向け（認証不要・`X-Internal-Secret` で認証）

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/internal/watchlist` | ウォッチリスト取得（Lambdaが株価取得対象を取得） |
| POST | `/internal/stock-prices` | 株価データ受取・stock_prices保存・Claude分析起動 |
| POST | `/internal/weekly-report` | 週次レポート・CSV更新起動 |

### 8.2 内部エンドポイント詳細

#### POST /internal/stock-prices

LambdaからGoへ株価データを渡すエンドポイント。受信後にGoが**銘柄ごとの最新値（現在値・前日比）を `stock_prices` に `ticker` キーでUPSERT保存（1銘柄1行を上書き）**し、続けて分析・通知まで一連の処理を行う。

```json
// リクエスト（Lambda → Go）
{
  "fetched_at": "2026-06-01T06:30:00Z",
  "stocks": [
    {
      "ticker": "6522.T",
      "prices": [
        { "date": "2026-05-30", "open": 1180, "high": 1220, "low": 1160, "close": 1200, "volume": 52000 }
      ]
    }
  ]
}

// レスポンス
{ "message": "ok", "analyzed": 5, "signals": 2 }
```

**Go側の処理フロー（受信後）:**
```
① 末尾2営業日から現在値・前日比を算出し stock_prices にUPSERT（ticker キーで1銘柄1行を上書き）
② メモリ上でテクニカル指標を計算
③ S3から最新 learning_vN.csv を読み込む
④ Claude APIに分析依頼（株価データ + 学習CSV）
⑤ 分析結果をanalysis_logsに保存・tradesに記録（Neon DB）
⑥ BUY/SELLシグナルをLINE通知
※ stock_prices は現在値・前日比のデータソース（含み益は参照時にユーザー単価と突き合わせて算出）
```

#### POST /internal/weekly-report

```json
// リクエスト（Lambda → Go）
{ "triggered_at": "2026-06-01T09:00:00Z" }

// レスポンス
{ "message": "ok", "csv_version": 4 }
```

### 8.3 ディレクトリ構成（Go）

```
backend/
  ├── main.go
  ├── handler/
  │   ├── auth.go
  │   ├── watchlist.go
  │   ├── trades.go
  │   ├── positions.go
  │   ├── portfolio.go
  │   ├── reports.go
  │   └── internal.go          # Lambda受取・内部処理
  ├── service/
  │   ├── analyzer.go          # Claude API分析
  │   ├── notifier.go          # LINE通知
  │   ├── learning.go          # CSV更新・S3操作
  │   └── weekly_report.go     # 週次レポート生成
  ├── repository/
  │   ├── watchlist.go
  │   ├── trades.go
  │   └── positions.go
  ├── model/
  │   ├── watchlist.go
  │   ├── trade.go
  │   ├── position.go
  │   └── stock_price.go       # 受取データの構造体定義（stock_pricesにUPSERT保存）
  ├── middleware/
  │   ├── jwt_auth.go
  │   └── require_admin.go
  └── db/
      └── neon.go
```

---

## 9. Lambda関数仕様（Python）

### 9.1 役割

**PythonのLambdaは株価取得のみを担当する。**
DB保存・AI分析・通知・CSV更新はすべてGoが行う。

```
Lambda（Python）の責務:
  ① ウォッチリスト銘柄をGoから取得
  ② yfinanceで株価データを取得
  ③ GoのAPIエンドポイントにPOST
  以上
```

### 9.2 ファイル構成

```
lambda/
  ├── fetch_price.py      # メイン（株価取得 → GoにPOST）
  └── requirements.txt    # yfinance, requests
```

### 9.3 fetch_price.py フロー

```
1. GET /internal/watchlist をGoのAPIに呼び出し、対象銘柄リストを取得（X-Internal-Secret付き）
2. yfinanceで各銘柄の過去120日分OHLCVを取得
3. POST /internal/stock-prices にデータを渡す
4. 完了（以降の処理はGoに委譲）
```

### 9.4 EventBridgeスケジュール設定

| 処理 | Cronスケジュール（UTC） | JST換算 |
|------|----------------------|---------|
| 株価取得（引け後） | `30 6 * * MON-FRI` | **平日 15:30のみ** |
| 週次レポートトリガー | `0 9 * * SUN` | 日曜 18:00 |

### 9.5 環境変数

| 変数名 | 説明 |
|--------|------|
| `GO_API_BASE_URL` | GoバックエンドのベースURL |
| `INTERNAL_API_SECRET` | 内部エンドポイント認証用シークレット |

---

## 10. 通知仕様

### 8.1 買いシグナル通知

```
🟢 買いシグナル
━━━━━━━━━━━━━━━━
銘柄: アステリスク (6522)
現在値: ¥1,180
目標値: ¥1,350
損切り: ¥1,050
━━━━━━━━━━━━━━━━
根拠: RSI過売り圏からの反発、
      出来高1.8倍。サポートで下げ止まり。
信頼度: 85%
━━━━━━━━━━━━━━━━
[バーチャル] 自動記録済み
```

### 8.2 売りシグナル通知（保有株あり）

```
🔴 売りシグナル（保有株あり）
━━━━━━━━━━━━━━━━
銘柄: ○○○ (XXXX)
現在値: ¥2,400
取得単価: ¥1,900
含み益: +¥50,000 (+26.3%)
━━━━━━━━━━━━━━━━
根拠: 目標値到達、RSI過買い圏
信頼度: 80%
━━━━━━━━━━━━━━━━
推奨: 半決済または全決済を検討
```

### 8.3 週次レポート通知

```
📊 週次トレードレポート
（5/26〜5/30）
━━━━━━━━━━━━━━━━
【バーチャル成績】
トレード数: 8件
勝率: 62.5%
週間損益: +¥42,800
累計損益: +¥187,000

【実運用】
（実績は手入力で更新）
━━━━━━━━━━━━━━━━
【AIの学習メモ】
✅ 出来高急増銘柄のエントリーが有効
❌ 下落トレンド中の逆張りは損失要因
📌 来週: トレンドフォロー重視
```

### 8.4 エラー通知

```
⚠️ システムエラー
処理: 定期分析（15:30）
エラー: yfinance接続タイムアウト
時刻: 2026-05-31 15:32 JST
```

---

## 11. 学習サイクル仕様

### 11.1 概要

AIの学習は**自己最適化するCSVファイル**をベースに行う。毎週日曜18:00に**管理者のトレード結果（直近26週）**をもとにCSVをバージョンアップさせ、全ユーザー共通の学習データとして毎日の分析プロンプトに注入する。

ユーザーが増えても週次レポートは1回のみ生成されるため**コストはユーザー数に依存しない。**

### 11.2 学習サイクルフロー

```
毎週日曜 18:00 Lambda実行
    ↓
① Neon DBのtradesテーブルから
  「管理者のtrades」の直近26週分を集計（SQL）
  WHERE user_id IN (SELECT id FROM users WHERE role = 'admin')
    ↓
② S3から現在の learning_vN.csv を読み込む
    ↓
③ Claude APIに以下を渡す
   - 管理者の直近26週のトレード集計データ
   - 現在のlearning.csvの内容
   - 更新指示プロンプト（3,000文字制限）
    ↓
④ ClaudeがCSVを自己最適化して返却
   （有効パターン強調・無効パターン削除・戦略を凝縮）
    ↓
⑤ learning_v(N+1).csv としてS3に保存（全ユーザー共通）
    ↓
⑥ 最新バージョンのパスをNeon DBに記録（learning_versions）
    ↓
⑦ LINE通知（週次サマリー）
```

### 11.3 毎日の分析フロー（Go側の処理）

```
POST /internal/stock-prices を受信（Lambdaから）
    ↓
① 末尾2営業日から現在値・前日比を算出し stock_prices にUPSERT
   （ticker キーで1銘柄1行を上書き・現在値/前日比のソース）
    ↓
② メモリ上でテクニカル指標を計算
   （RSI・MACD・移動平均・ボリンジャーバンド）
    ↓
③ S3から最新の learning_vN.csv を読み込む
    ↓
④ プロンプトを構築
   ┌─────────────────────────────┐
   │ システムプロンプト            │
   │  + learning.csvの内容（注入）│
   ├─────────────────────────────┤
   │ ユーザープロンプト            │
   │  + 銘柄データ                │
   │  + テクニカル指標            │
   └─────────────────────────────┘
    ↓
⑤ Claude APIに送信 → 投資判断JSON取得
    ↓
⑥ 分析結果をanalysis_logsに保存・tradesに記録（Neon DB）
    ↓
⑦ BUY/SELLシグナルをLINE通知
※ stock_prices は現在値・前日比のデータソース（含み益は参照時にユーザー単価と突き合わせて算出）
```

### 11.4 CSVの構造

```csv
updated_at,version,week_range,char_count,content
2026-06-01,3,2025-12-01〜2026-05-25,2850,"
【有効パターン（勝率70%以上）】
- RSI30以下 + 出来高1.5倍以上 → 勝率78%
- 朝9時エントリー → 夕方より平均+3.2%有利
- 信頼度0.8以上のみ参戦 → 勝率82%

【無効パターン（廃止）】
- 下落トレンド中の逆張り → 勝率34%で廃止
- 決算週のエントリー → 不確実性高く見送り

【相場環境別戦略】
- 上昇相場: 積極参戦（信頼度0.7以上）
- 下落相場: 見送り推奨
- レンジ相場: 信頼度0.85以上のみ

【損切りルール】
- 取得価格-8%で必ず損切り
- 信頼度0.7未満のシグナルは見送り

【来週の重点事項】
- 出来高確認を最優先
- 上昇トレンド銘柄のみ参戦
"
```

### 11.5 CSV更新プロンプト仕様

```
Claudeへの更新指示:

「以下の制約でlearning.csvを更新してください：

【制約】
- 最大文字数: 3,000文字以内（厳守）
- 対象期間: 直近26週（半年）のデータのみ反映
- 古い情報は圧縮または削除してよい

【更新方針】
- 勝率70%以上のパターンは強調・具体化する
- 勝率50%未満のパターンは削除する
- より具体的な数値（勝率・平均損益）を含める
- 戦略は実践的・簡潔に記述する
- 有効な情報を凝縮し、より良い戦略に進化させる

【今週のトレード結果】
（集計データをここに挿入）

【現在のlearning.csv】
（既存CSVの内容をここに挿入）
」
```

### 11.6 バージョン管理

| 項目 | 内容 |
|------|------|
| 保存先 | AWS S3 |
| ファイル名 | `learning_v{N}.csv` |
| 最新版の参照 | Neon DBの `learning_versions` テーブルで管理 |
| 過去版の保持 | S3に全バージョン保管（振り返り・比較用） |
| 文字数上限 | **3,000文字**（約750トークン） |
| 対象期間 | 直近26週（半年分）のトレード結果 |

テーブル定義は [6.2 DDL](#62-ddl) の `learning_versions` を正とする（**全ユーザー共通・`user_id` なし**）。

```sql
-- バージョン管理テーブル（全ユーザー共通・user_idなし）
CREATE TABLE learning_versions (
  id          SERIAL PRIMARY KEY,
  version     INT          NOT NULL,
  s3_path     VARCHAR(255) NOT NULL,   -- 例: s3://bucket/learning_v3.csv（全ユーザー共通）
  week_range  VARCHAR(50),             -- 例: 2025-12-01〜2026-05-25
  char_count  INT,
  created_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);
```

### 11.7 学習の進化イメージ

| バージョン | 学習状態 |
|-----------|---------|
| v1〜v4（1ヶ月） | 基本的な勝敗パターンの記録開始 |
| v5〜v13（3ヶ月） | 有効パターンが明確化・無効パターンが削除される |
| v14〜v26（半年） | 相場環境別の戦略が確立・損切りルールが最適化 |
| v27〜（1年〜） | 銘柄特性・季節性まで反映した洗練された戦略 |

続けるほど**不要な情報が削られ・有効な情報が凝縮されていく自己最適化プロンプト**に進化する。

---

## 12. 非機能要件

### 12.1 信頼性

| 項目 | 要件 |
|------|------|
| Lambda実行失敗時 | エラーをLINE通知・次回実行時にリトライ |
| yfinance障害時 | エラーログを記録・通知して処理スキップ |
| Neon接続失敗時 | 3回リトライ後にエラー通知 |
| データ欠損 | 欠損銘柄をスキップして他銘柄の処理を継続 |

### 12.2 セキュリティ

| 項目 | 対策 |
|------|------|
| APIキー管理 | AWS Lambda環境変数に保存（ハードコード禁止） |
| DB接続情報 | 環境変数で管理・SSL接続を使用 |
| LINE Token | 環境変数で管理 |

### 12.3 保守性

- 銘柄の追加・削除はwatchlistテーブルの更新のみ
- 分析ロジックはプロンプトの修正で対応可能
- ログはanalysis_logsに全件保存

---

## 13. コスト試算

### 13.1 開発期間中

| サービス | 費用 |
|---------|------|
| Claude Code（Proプラン） | $20/月 |
| Claude API | $3〜5/月 |
| AWS Lambda | 無料枠内 |
| Neon PostgreSQL | 無料 |
| LINE Messaging API | 無料 |
| **合計** | **約$25/月（約3,750円）** |

### 13.2 運用期間中（開発完了後）

| サービス | 費用 |
|---------|------|
| Claude Code | 不要（解約） |
| Claude API | $3〜5/月（上限$5設定） |
| AWS Lambda | 無料枠内 |
| Neon PostgreSQL | 無料 |
| LINE Messaging API | 無料 |
| **合計** | **約$5/月（約750円）** |

### 13.3 Claude API上限設定

Anthropicダッシュボードにて月額上限を **$5** に設定し、青天井課金を防止する。

---

## 14. フェーズ計画

### Phase 1：基本構成の構築（目安：2〜3週）

| タスク | 内容 |
|--------|------|
| Lambda環境構築 | AWS Lambda + EventBridge設定 |
| 株価取得 | yfinanceによるOHLCV取得実装 |
| Claude API連携 | 分析プロンプト・JSON出力実装 |
| LINE通知 | BUY/SELLシグナル通知実装 |
| 動作確認 | 単銘柄でのE2Eテスト |

### Phase 2：DB連携・認証・フロント基盤（目安：2〜3週）

| タスク | 内容 |
|--------|------|
| Neon DB構築 | 全テーブル作成・接続設定 |
| 認証API | JWT・招待コード登録・ログイン実装 |
| フロント基盤 | Next.js App Router・MUI・middleware設定 |
| ログイン・登録画面 | SCR-01・SCR-02実装 |
| 分析ログ保存 | analysis_logsへの保存実装 |
| バーチャル売買 | tradesテーブルへの自動記録 |

### Phase 3：ダッシュボード・管理者機能（目安：2〜3週）

| タスク | 内容 |
|--------|------|
| ダッシュボード | SCR-03実装（損益サマリー・シグナル一覧） |
| ウォッチリスト | SCR-04実装（手動追加・削除） |
| 管理者画面基盤 | SCR-09〜11実装（ユーザー管理・招待コード） |
| 分析設定 | SCR-12実装（テーマDB管理・設定保存） |
| テーマ管理 | analysis_themesのCRUD実装 |

### Phase 4：候補提案・学習ループ（目安：1〜2週）

| タスク | 内容 |
|--------|------|
| 候補提案 | watchlist_candidates生成・LINE通知実装 |
| 候補承認画面 | SCR-13実装（承認・却下フロー） |
| 週次レポート | 週次CSV更新・学習ループ実装 |
| 実運用モード | real_positions・売りシグナル通知実装 |

### Phase 5：残画面・精度向上（継続）

| タスク | 内容 |
|--------|------|
| トレード履歴 | SCR-05実装 |
| ポートフォリオ | SCR-06実装（グラフ含む） |
| 週次レポート画面 | SCR-07実装 |
| 設定画面 | SCR-08実装 |
| 精度向上 | テクニカル指標拡充・バックテスト |

---

## 15. リスクと制約

### 15.1 技術的リスク

| リスク | 影響 | 対策 |
|--------|------|------|
| yfinance仕様変更 | データ取得不能 | エラー通知・代替ソースへの切り替え準備 |
| Claude API障害 | 分析不能 | エラーログ記録・次回実行で補完 |
| Neon無料枠超過 | DB停止 | 定期的なデータ量確認・不要ログ削除 |

### 15.2 運用上の制約

| 制約 | 内容 |
|------|------|
| 自動発注 | 実装しない（人間が最終判断） |
| リアルタイムデータ | yfinanceは厳密なリアルタイムではない（数分遅延の可能性） |
| 投資判断の責任 | AIの分析はあくまで参考情報。投資判断・損益の責任はユーザーが負う |
| yfinance商用利用 | 個人利用の範囲で使用する |

### 15.3 投資リスク免責

本システムはAIによる分析補助ツールであり、投資利益を保証するものではない。すべての投資判断と損益の責任はユーザー自身が負うものとする。

---

*本仕様書は開発進行に伴い随時更新する。*
