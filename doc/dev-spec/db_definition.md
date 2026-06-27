# AI株式トレーディングシステム DB定義書

**バージョン**: 1.5  
**作成日**: 2026-06-06  
**更新内容**: `stock_prices` を最新株価スナップショット（1銘柄1行・`ticker`でUPSERT）に変更。現在値・前日比を算出済みカラムとして保持（含み益は参照時にユーザー単価と突き合わせて算出）  
**DB**: Neon（Serverless PostgreSQL）  
**ステータス**: Draft

---

## 目次

1. [ER図（テーブル関連図）](#1-er図)
2. [テーブル一覧](#2-テーブル一覧)
3. [テーブル定義詳細](#3-テーブル定義詳細)
4. [インデックス一覧](#4-インデックス一覧)
5. [初期データ（シードデータ）](#5-初期データ)
6. [完全DDL（実行用）](#6-完全ddl)

---

## 1. ER図

```
users
  ├── invitation_codes（created_by, used_by → users.id）
  ├── trades（user_id → users.id）
  ├── real_positions（user_id → users.id）
  ├── analysis_settings（created_by → users.id）
  ├── analysis_themes（created_by → users.id）
  └── watchlist_candidates（decided_by → users.id）

※ watchlist・stock_prices・analysis_logs・learning_logs・learning_versions は全ユーザー共通（user_id なし）
```

### データ分離方針

| 分類 | テーブル | 理由 |
|------|---------|------|
| ユーザーごと | trades・real_positions | 各ユーザーが個別に管理 |
| 全ユーザー共通 | watchlist・**stock_prices**・analysis_logs・**learning_logs**・**learning_versions**・analysis_settings・analysis_themes・watchlist_candidates | 管理者が管理・コスト最適化 |
| 管理者のみ操作 | invitation_codes | 招待コードの発行は管理者のみ |

- ユーザーごとのテーブルはAPIレイヤーで `WHERE user_id = 自分のID` を必ず適用
- `learning_logs`・`learning_versions` は管理者のtradesから生成・全ユーザーで共有
- ユーザーが何人増えても週次レポートは1回のみ生成（コスト固定）

---

## 2. テーブル一覧

| # | テーブル名 | 用途 | データ分離 |
|---|-----------|------|-----------|
| 1 | `users` | ユーザーアカウント | - |
| 2 | `invitation_codes` | 招待コード管理 | 管理者のみ操作 |
| 3 | `watchlist` | 監視銘柄リスト | **全ユーザー共通** |
| 4 | `watchlist_candidates` | AIが提案したウォッチリスト候補 | 全ユーザー共通 |
| 5 | `stock_prices` | 最新株価スナップショット（現在値・前日比・1銘柄1行・UPSERT） | **全ユーザー共通** |
| 6 | `trades` | トレード履歴（バーチャル・実運用共通） | ユーザーごと |
| 7 | `real_positions` | 実運用の保有株 | ユーザーごと |
| 8 | `analysis_logs` | Claude APIの分析結果ログ | **全ユーザー共通** |
| 9 | `learning_logs` | 週次学習ログ | **全ユーザー共通** |
| 10 | `learning_versions` | 学習CSVのバージョン管理 | **全ユーザー共通** |
| 11 | `analysis_settings` | 分析設定（管理者が設定） | 全ユーザー共通 |
| 12 | `analysis_themes` | 分析テーマ一覧（管理者がUI上で管理） | 全ユーザー共通 |

---

## 3. テーブル定義詳細

---

### 3.1 users（ユーザー）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| email | VARCHAR(255) | NOT NULL | - | メールアドレス（ユニーク） |
| name | VARCHAR(100) | NOT NULL | - | 表示名 |
| password_hash | VARCHAR(255) | NOT NULL | - | bcryptハッシュ（平文は保存しない） |
| role | VARCHAR(10) | NOT NULL | `'user'` | ロール：`admin` / `user` |
| is_active | BOOLEAN | NOT NULL | `TRUE` | アカウント有効フラグ |
| created_at | TIMESTAMP | NOT NULL | `NOW()` | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | `NOW()` | 更新日時 |

**制約：**
- `role` CHECK：`IN ('admin', 'user')`
- `email` UNIQUE

---

### 3.2 invitation_codes（招待コード）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| code | VARCHAR(20) | NOT NULL | - | 招待コード（例: `TRADE-XXXX-XXXX`）（ユニーク） |
| created_by | INT | NULL | - | 発行した管理者のuser_id |
| used_by | INT | NULL | - | 使用したユーザーのuser_id |
| expires_at | TIMESTAMP | NOT NULL | - | 有効期限 |
| used_at | TIMESTAMP | NULL | - | 使用日時 |
| is_active | BOOLEAN | NOT NULL | `TRUE` | 有効フラグ（管理者が無効化可能） |
| created_at | TIMESTAMP | NOT NULL | `NOW()` | 作成日時 |

**ステータス判定ロジック：**
```
is_active = false          → 無効化済み
used_by IS NOT NULL        → 使用済み
expires_at < NOW()         → 期限切れ
上記以外                   → 有効
```

**外部キー：**
- `created_by` → `users(id)`
- `used_by` → `users(id)`

---

### 3.3 watchlist（ウォッチリスト）

全ユーザー共通。管理者が管理する。同一銘柄の重複登録を防ぐため `ticker` にUNIQUE制約を付与。Claude APIの分析コストを最小化するため最大3銘柄を全ユーザーで共有する。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| ticker | VARCHAR(10) | NOT NULL | - | 銘柄コード（例: `6522.T`）（ユニーク） |
| name | VARCHAR(100) | NULL | - | 銘柄名（例: アステリスク） |
| mode | VARCHAR(10) | NOT NULL | - | 運用モード：`virtual` / `real` / `both` |
| is_active | BOOLEAN | NOT NULL | `TRUE` | 有効フラグ |
| added_at | TIMESTAMP | NOT NULL | `NOW()` | 追加日時 |

**制約：**
- `mode` CHECK：`IN ('virtual', 'real', 'both')`
- `ticker` UNIQUE：同一銘柄の重複登録を防止
- 最大3銘柄（アプリケーション層で制御）
- 追加・削除は管理者のみ可能

---

### 3.4 watchlist_candidates（ウォッチリスト候補）

AIが毎日15:30の分析時に提案するウォッチリスト候補。管理者が承認/却下する。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| ticker | VARCHAR(10) | NOT NULL | - | 候補銘柄コード（例: `9984.T`） |
| name | VARCHAR(100) | NULL | - | 銘柄名 |
| reason | TEXT | NULL | - | 提案理由（AI生成） |
| replace_ticker | VARCHAR(10) | NULL | - | 置き換え推奨銘柄（任意） |
| confidence | NUMERIC(4,3) | NULL | - | 確信度（0.000〜1.000） |
| status | VARCHAR(10) | NOT NULL | `'pending'` | ステータス：`pending` / `approved` / `rejected` |
| proposed_at | TIMESTAMP | NOT NULL | `NOW()` | 提案日時 |
| decided_at | TIMESTAMP | NULL | - | 承認/却下日時 |
| decided_by | INT | NULL | - | 決定した管理者のuser_id |

**制約：**
- `status` CHECK：`IN ('pending', 'approved', 'rejected')`

**外部キー：**
- `decided_by` → `users(id)`

---

### 3.4.1 stock_prices（最新株価スナップショット）

全ユーザー共通。Lambdaが取得した株価データを Go が受け取り、**銘柄ごとに最新の現在値・前日比を1行だけ保持**するテーブル。ダッシュボード・ポートフォリオの「現在値・前日比・含み益」のデータソースとなる。

日次の履歴は蓄積せず、**1銘柄＝1行**とする。`ticker` にUNIQUE制約を付与し、同一銘柄を再取得した場合は **INSERTではなくUPSERT（あれば更新・なければ挿入）** で最新値に上書きする。

- 現在値・前日比は **算出済みの値をカラムとして保持**する（`close`・`prev_close`・`change_amount`・`change_rate`）。
- 前日比 = 最新終値（`close`）と前営業日終値（`prev_close`）の差。Goが受信した120日分OHLCVの末尾2営業日から計算してUPSERTする。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| ticker | VARCHAR(10) | NOT NULL | - | 銘柄コード（例: `6522.T`）（ユニーク） |
| date | DATE | NOT NULL | - | 最新終値の取引日 |
| open | NUMERIC(10,2) | NOT NULL | - | 最新営業日の始値（円） |
| high | NUMERIC(10,2) | NOT NULL | - | 最新営業日の高値（円） |
| low | NUMERIC(10,2) | NOT NULL | - | 最新営業日の安値（円） |
| close | NUMERIC(10,2) | NOT NULL | - | 現在値＝最新終値（円） |
| prev_close | NUMERIC(10,2) | NULL | - | 前営業日終値（円）※前日比の算出元 |
| change_amount | NUMERIC(10,2) | NULL | - | 前日比（円）＝ `close − prev_close` |
| change_rate | NUMERIC(6,2) | NULL | - | 前日比（%）＝ `(close − prev_close) / prev_close × 100` |
| volume | BIGINT | NOT NULL | `0` | 最新営業日の出来高（株）※INT上限を超えうるためBIGINT |
| updated_at | TIMESTAMP | NOT NULL | `NOW()` | 最終更新（UPSERT）日時 |

**制約：**
- `ticker` UNIQUE：1銘柄1行。再取得時はこのキーでUPSERT（上書き）する

**用途別の参照例：**

| 用途 | 取得方法 |
|------|---------|
| 現在値 | `ticker` で1行取得し `close` |
| 前日比（円・%） | 保持済みの `change_amount` / `change_rate` をそのまま表示 |
| 含み益／損益率（実運用・バーチャル） | `close` − 取得単価（`real_positions.avg_price` / `trades.price`）をユーザーごとに算出 |

> **補足（含み益の扱い）**: 含み益はユーザーの保有数量・取得単価に依存するため、全ユーザー共通の `stock_prices` にはカラムとして持たせず、`close` を取得単価と突き合わせて**参照時に算出**する。`stock_prices` が保持するのは全ユーザー共通の市場データ（現在値・前日比）のみ。

---

### 3.5 trades（トレード履歴）

バーチャル・実運用両方のトレード履歴を1テーブルで管理。`mode`カラムで区別する。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| user_id | INT | NOT NULL | - | ユーザーID |
| ticker | VARCHAR(10) | NOT NULL | - | 銘柄コード |
| mode | VARCHAR(10) | NOT NULL | - | `virtual`（バーチャル）/ `real`（実運用） |
| action | VARCHAR(10) | NOT NULL | - | `BUY` / `SELL` |
| price | NUMERIC(10,2) | NOT NULL | - | 取引単価（円） |
| quantity | INT | NOT NULL | - | 取引数量（株） |
| confidence | NUMERIC(4,3) | NULL | - | AI確信度（0.000〜1.000） |
| reason | TEXT | NULL | - | AI判断根拠 |
| target_price | NUMERIC(10,2) | NULL | - | 目標株価（円） |
| stop_loss | NUMERIC(10,2) | NULL | - | 損切りライン（円） |
| result_pnl | NUMERIC(10,2) | NULL | - | 確定損益（円）※決済後に更新 |
| closed_at | TIMESTAMP | NULL | - | 決済日時 |
| created_at | TIMESTAMP | NOT NULL | `NOW()` | トレード日時 |

**制約：**
- `mode` CHECK：`IN ('virtual', 'real')`
- `action` CHECK：`IN ('BUY', 'SELL')`

**外部キー：**
- `user_id` → `users(id)`

---

### 3.6 real_positions（実運用保有株）

実運用モードでユーザーが実際に保有している株を登録するテーブル。手動で登録・更新する。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| user_id | INT | NOT NULL | - | ユーザーID |
| ticker | VARCHAR(10) | NOT NULL | - | 銘柄コード |
| name | VARCHAR(100) | NULL | - | 銘柄名 |
| quantity | INT | NOT NULL | - | 保有数量（株） |
| avg_price | NUMERIC(10,2) | NOT NULL | - | 平均取得単価（円） |
| updated_at | TIMESTAMP | NOT NULL | `NOW()` | 更新日時 |

**制約：**
- `(user_id, ticker)` UNIQUE：同一ユーザーの同一銘柄は1レコードのみ

**外部キー：**
- `user_id` → `users(id)`

---

### 3.7 analysis_logs（分析結果ログ）

Claude APIの分析結果をすべて記録するログテーブル。管理者設定によりLambdaが一括実行するため**全ユーザー共通**。`user_id`は持たない。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| ticker | VARCHAR(10) | NOT NULL | - | 銘柄コード |
| action | VARCHAR(10) | NULL | - | `BUY` / `SELL` / `HOLD` |
| confidence | NUMERIC(4,3) | NULL | - | 確信度 |
| analysis | JSONB | NULL | - | Claude APIの出力JSON全体 |
| analyzed_at | TIMESTAMP | NOT NULL | `NOW()` | 分析日時 |

**analysisカラムのJSON構造：**
```json
{
  "ticker": "6522.T",
  "action": "BUY",
  "confidence": 0.85,
  "buy_reasons": ["RSI過売り圏", "出来高急増"],
  "no_buy_reasons": ["25日線が下向き"],
  "entry_condition": "翌日陽線確認後",
  "target_price": 1350,
  "stop_loss": 1080,
  "position_size": 0.1,
  "reason": "総合判断コメント"
}
```

---

### 3.8 learning_logs（週次学習ログ）

全ユーザー共通。**管理者のtradesのみ**を集計して生成する。ユーザーが増えても週次レポートは1回のみ生成されるためコスト固定。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| week_start | DATE | NOT NULL | - | 集計開始日（月曜日） |
| week_end | DATE | NOT NULL | - | 集計終了日（金曜日） |
| trade_count | INT | NULL | - | 週間トレード数（管理者のみ） |
| win_rate | NUMERIC(5,2) | NULL | - | 勝率（%） |
| total_pnl | NUMERIC(10,2) | NULL | - | 週間損益（円） |
| summary | TEXT | NULL | - | 週次サマリー（AI生成） |
| lessons | TEXT | NULL | - | 学習・反省（AI生成） |
| strategy | TEXT | NULL | - | 来週の戦略（AI生成） |
| raw_response | TEXT | NULL | - | Claude APIの生レスポンス |
| created_at | TIMESTAMP | NOT NULL | `NOW()` | 生成日時 |

### 3.9 learning_versions（学習CSVバージョン管理）

全ユーザー共通。S3に保存する学習CSVファイルのバージョン履歴を管理する。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| version | INT | NOT NULL | - | バージョン番号（1から始まる連番） |
| s3_path | VARCHAR(255) | NOT NULL | - | S3パス（例: `s3://bucket/learning_v3.csv`） |
| week_range | VARCHAR(50) | NULL | - | 対象期間（例: `2025-12-01〜2026-05-25`） |
| char_count | INT | NULL | - | CSVの文字数（上限3,000文字） |
| created_at | TIMESTAMP | NOT NULL | `NOW()` | 生成日時 |

---

### 3.10 analysis_settings（分析設定）

管理者が設定する全ユーザー共通の分析プロンプト設定。`is_active = TRUE` のレコードが有効な設定。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| theme_ids | INT[] | NULL | - | 選択中のanalysis_themes.idの配列 |
| screening | JSONB | NULL | - | スクリーニング条件 |
| style | VARCHAR(50) | NULL | - | 分析スタイル（例: `short_term_trend`） |
| free_prompt | TEXT | NULL | - | 自由入力プロンプト（最大1,000文字） |
| is_active | BOOLEAN | NOT NULL | `TRUE` | 有効フラグ（常にTRUEのレコードが1件のみ） |
| created_by | INT | NULL | - | 設定した管理者のuser_id |
| created_at | TIMESTAMP | NOT NULL | `NOW()` | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | `NOW()` | 更新日時 |

**screeningカラムのJSON構造：**
```json
{
  "min_market_cap": 10000000000,
  "min_volume": 10000,
  "max_per": 50
}
```

**styleカラムの値：**

| 値 | 意味 |
|----|------|
| `short_term_trend` | 短期・順張り |
| `short_term_contrarian` | 短期・逆張り |
| `mid_term_trend` | 中期・順張り |
| `mid_term_contrarian` | 中期・逆張り |
| `both` | 両方 |

**外部キー：**
- `created_by` → `users(id)`

---

### 3.11 analysis_themes（分析テーマ）

管理者がUI上で管理するテーマ一覧。`is_active = TRUE` のものだけ分析設定画面に表示される。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | SERIAL | NOT NULL | auto | 主キー |
| name | VARCHAR(100) | NOT NULL | - | テーマ名（例: `AI・生成AI`）（ユニーク） |
| description | VARCHAR(255) | NULL | - | 説明（例: `AIソフト・LLM関連銘柄`） |
| sort_order | INT | NOT NULL | `0` | 表示順（小さいほど上に表示） |
| is_active | BOOLEAN | NOT NULL | `TRUE` | 有効フラグ |
| created_by | INT | NULL | - | 作成した管理者のuser_id |
| created_at | TIMESTAMP | NOT NULL | `NOW()` | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | `NOW()` | 更新日時 |

**制約：**
- `name` UNIQUE

**外部キー：**
- `created_by` → `users(id)`

---

## 4. インデックス一覧

| インデックス名 | テーブル | カラム | 目的 |
|-------------|---------|--------|------|
| idx_users_email | users | email | ログイン時のメール検索 |
| idx_invitation_codes_code | invitation_codes | code | 招待コード検索 |
| idx_watchlist_candidates_status | watchlist_candidates | status | pending件数の取得 |
| idx_stock_prices_ticker | stock_prices | ticker | 現在値・前日比の取得（UNIQUE兼用・UPSERTキー） |
| idx_trades_user_id | trades | user_id | ユーザー別履歴取得 |
| idx_trades_ticker | trades | ticker | 銘柄別履歴取得 |
| idx_trades_mode | trades | mode | バーチャル/実運用の切り替え |
| idx_trades_created_at | trades | created_at | 期間指定検索・集計 |
| idx_analysis_logs_ticker | analysis_logs | ticker | 銘柄別ログ取得 |
| idx_analysis_logs_analyzed_at | analysis_logs | analyzed_at | 最新分析取得 |
| idx_analysis_themes_sort_order | analysis_themes | sort_order | テーマ一覧の表示順取得 |

---

## 5. 初期データ（シードデータ）

マイグレーション実行後に投入する初期データ。

### 5.1 管理者ユーザー

```sql
-- パスワードは 'admin_password' をbcrypt(cost=12)でハッシュ化した値に変更すること
INSERT INTO users (email, name, password_hash, role, is_active)
VALUES ('admin@example.com', '管理者', '$2a$12$xxxxxxxxxxxxxxxxxxxxxx', 'admin', TRUE);
```

### 5.2 初期テーマ

```sql
INSERT INTO analysis_themes (name, description, sort_order, is_active, created_by) VALUES
  ('AI・生成AI',      'AIソフト・LLM関連銘柄',         1, TRUE, 1),
  ('半導体',          '製造装置・素材・設計',            2, TRUE, 1),
  ('防衛・宇宙',      '防衛関連・宇宙ビジネス',          3, TRUE, 1),
  ('データセンター',   'サーバー・冷却・電力インフラ',    4, TRUE, 1),
  ('物流・自動化',    '倉庫ロボット・配送効率化',         5, TRUE, 1),
  ('脱炭素・EV',      '再エネ・電池・EV関連',            6, TRUE, 1),
  ('内需・インバウンド','観光・飲食・小売',               7, TRUE, 1);
```

### 5.3 初期分析設定

```sql
INSERT INTO analysis_settings (theme_ids, screening, style, free_prompt, is_active, created_by)
VALUES (
  ARRAY[1, 2],
  '{"min_market_cap": 10000000000, "min_volume": 10000, "max_per": 50}',
  'short_term_trend',
  '出来高の急増を重視してください。',
  TRUE,
  1
);
```

---

## 6. 完全DDL（実行用）

```sql
-- =========================================
-- AI株式トレーディングシステム DB定義
-- =========================================

-- 1. ユーザー
CREATE TABLE users (
  id            SERIAL PRIMARY KEY,
  email         VARCHAR(255) NOT NULL UNIQUE,
  name          VARCHAR(100) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role          VARCHAR(10)  NOT NULL DEFAULT 'user'
                  CHECK (role IN ('admin', 'user')),
  is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- 2. 招待コード
CREATE TABLE invitation_codes (
  id          SERIAL PRIMARY KEY,
  code        VARCHAR(20)  NOT NULL UNIQUE,
  created_by  INT          REFERENCES users(id),
  used_by     INT          REFERENCES users(id),
  expires_at  TIMESTAMP    NOT NULL,
  used_at     TIMESTAMP,
  is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- 3. ウォッチリスト（全ユーザー共通・管理者が管理）
CREATE TABLE watchlist (
  id         SERIAL PRIMARY KEY,
  ticker     VARCHAR(10)  NOT NULL UNIQUE,
  name       VARCHAR(100),
  mode       VARCHAR(10)  NOT NULL
               CHECK (mode IN ('virtual', 'real', 'both')),
  is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
  added_at   TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- 4. ウォッチリスト候補
CREATE TABLE watchlist_candidates (
  id              SERIAL PRIMARY KEY,
  ticker          VARCHAR(10)  NOT NULL,
  name            VARCHAR(100),
  reason          TEXT,
  replace_ticker  VARCHAR(10),
  confidence      NUMERIC(4,3),
  status          VARCHAR(10)  NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'approved', 'rejected')),
  proposed_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
  decided_at      TIMESTAMP,
  decided_by      INT          REFERENCES users(id)
);

-- 4b. 最新株価スナップショット（全ユーザー共通・user_idなし・1銘柄1行）
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
);

-- UPSERT例（同一銘柄は ticker で上書き）
-- INSERT INTO stock_prices (ticker, date, open, high, low, close, prev_close, change_amount, change_rate, volume)
-- VALUES (...)
-- ON CONFLICT (ticker) DO UPDATE SET
--   date = EXCLUDED.date, open = EXCLUDED.open, high = EXCLUDED.high, low = EXCLUDED.low,
--   close = EXCLUDED.close, prev_close = EXCLUDED.prev_close,
--   change_amount = EXCLUDED.change_amount, change_rate = EXCLUDED.change_rate,
--   volume = EXCLUDED.volume, updated_at = NOW();

-- 5. トレード履歴
CREATE TABLE trades (
  id           SERIAL PRIMARY KEY,
  user_id      INT           NOT NULL REFERENCES users(id),
  ticker       VARCHAR(10)   NOT NULL,
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
  result_pnl   NUMERIC(10,2),
  closed_at    TIMESTAMP,
  created_at   TIMESTAMP     NOT NULL DEFAULT NOW()
);

-- 6. 実運用保有株
CREATE TABLE real_positions (
  id         SERIAL PRIMARY KEY,
  user_id    INT           NOT NULL REFERENCES users(id),
  ticker     VARCHAR(10)   NOT NULL,
  name       VARCHAR(100),
  quantity   INT           NOT NULL,
  avg_price  NUMERIC(10,2) NOT NULL,
  updated_at TIMESTAMP     NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, ticker)
);

-- 7. 分析結果ログ（全ユーザー共通・user_idなし）
CREATE TABLE analysis_logs (
  id          SERIAL PRIMARY KEY,
  ticker      VARCHAR(10) NOT NULL,
  action      VARCHAR(10),
  confidence  NUMERIC(4,3),
  analysis    JSONB,
  analyzed_at TIMESTAMP   NOT NULL DEFAULT NOW()
);

-- 8. 週次学習ログ（全ユーザー共通・管理者のtradesから生成）
CREATE TABLE learning_logs (
  id           SERIAL PRIMARY KEY,
  week_start   DATE         NOT NULL,
  week_end     DATE         NOT NULL,
  trade_count  INT,
  win_rate     NUMERIC(5,2),
  total_pnl    NUMERIC(10,2),
  summary      TEXT,
  lessons      TEXT,
  strategy     TEXT,
  raw_response TEXT,
  created_at   TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- 9. 学習CSVバージョン管理（全ユーザー共通）
CREATE TABLE learning_versions (
  id          SERIAL PRIMARY KEY,
  version     INT          NOT NULL,
  s3_path     VARCHAR(255) NOT NULL,
  week_range  VARCHAR(50),
  char_count  INT,
  created_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- 10. 分析設定
CREATE TABLE analysis_settings (
  id          SERIAL PRIMARY KEY,
  theme_ids   INT[],
  screening   JSONB,
  style       VARCHAR(50),
  free_prompt TEXT,
  is_active   BOOLEAN   NOT NULL DEFAULT TRUE,
  created_by  INT       REFERENCES users(id),
  created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 11. 分析テーマ
CREATE TABLE analysis_themes (
  id          SERIAL PRIMARY KEY,
  name        VARCHAR(100) NOT NULL UNIQUE,
  description VARCHAR(255),
  sort_order  INT          NOT NULL DEFAULT 0,
  is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
  created_by  INT          REFERENCES users(id),
  created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- =========================================
-- インデックス
-- =========================================
CREATE INDEX idx_users_email                 ON users (email);
CREATE INDEX idx_invitation_codes_code       ON invitation_codes (code);
CREATE INDEX idx_watchlist_candidates_status ON watchlist_candidates (status);
-- stock_prices は ticker のUNIQUE制約が索引を兼ねる（UPSERT・参照のキー）
CREATE INDEX idx_trades_user_id              ON trades (user_id);
CREATE INDEX idx_trades_ticker               ON trades (ticker);
CREATE INDEX idx_trades_mode                 ON trades (mode);
CREATE INDEX idx_trades_created_at           ON trades (created_at);
CREATE INDEX idx_analysis_logs_ticker        ON analysis_logs (ticker);
CREATE INDEX idx_analysis_logs_analyzed_at   ON analysis_logs (analyzed_at);
CREATE INDEX idx_analysis_themes_sort_order  ON analysis_themes (sort_order);

-- =========================================
-- 初期データ（シードデータ）
-- =========================================

-- 管理者ユーザー（パスワードハッシュは実際の値に変更すること）
INSERT INTO users (email, name, password_hash, role, is_active)
VALUES ('admin@example.com', '管理者', '$2a$12$xxxxxxxxxxxxxxxxxxxxxx', 'admin', TRUE);

-- 初期テーマ
INSERT INTO analysis_themes (name, description, sort_order, is_active, created_by) VALUES
  ('AI・生成AI',       'AIソフト・LLM関連銘柄',          1, TRUE, 1),
  ('半導体',           '製造装置・素材・設計',             2, TRUE, 1),
  ('防衛・宇宙',       '防衛関連・宇宙ビジネス',           3, TRUE, 1),
  ('データセンター',    'サーバー・冷却・電力インフラ',     4, TRUE, 1),
  ('物流・自動化',     '倉庫ロボット・配送効率化',          5, TRUE, 1),
  ('脱炭素・EV',       '再エネ・電池・EV関連',             6, TRUE, 1),
  ('内需・インバウンド','観光・飲食・小売',                 7, TRUE, 1);

-- 初期分析設定
INSERT INTO analysis_settings (theme_ids, screening, style, free_prompt, is_active, created_by)
VALUES (
  ARRAY[1, 2],
  '{"min_market_cap": 10000000000, "min_volume": 10000, "max_per": 50}',
  'short_term_trend',
  '出来高の急増を重視してください。',
  TRUE,
  1
);
```
