// Package external は外部サービス（Claude API・LINE通知・S3）のクライアントを提供する。
//
// 各 New*Client コンストラクタは環境変数の設定状況を見て実装を切り替える。
//   - 必要な環境変数が設定済み → 実クライアント（claude.go / line.go / s3.go）
//   - 未設定（ローカル開発・プレースホルダ）→ スタブ実装
//
// これにより本番（Lambda・環境変数あり）では実連携が動き、
// ローカル開発（.env がプレースホルダ）では外部通信せずスタブで動作確認できる。
package external

import (
	"log"
	"strings"
)

// ClaudeClient は株価分析を行うClaude APIクライアントのインターフェース。
type ClaudeClient interface {
	// Analyze は分析プロンプトをClaudeに渡し、生成された本文（JSON文字列等）を返す。
	Analyze(prompt string) (string, error)
}

// LineClient はLINE通知クライアントのインターフェース。
type LineClient interface {
	Notify(message string) error
}

// S3Client は学習CSVを保存するS3クライアントのインターフェース。
type S3Client interface {
	Upload(key string, body []byte) (string, error)
}

// isConfigured は環境変数値が「実際に設定されているか」を判定する。
// 空文字、および .env.example のプレースホルダ（xxxxx を含む値）は未設定とみなす。
func isConfigured(v string) bool {
	if v == "" {
		return false
	}
	if strings.Contains(v, "xxxxx") {
		return false
	}
	return true
}

// --- スタブ実装（ローカル動作確認用） ---

type stubClaudeClient struct{}

func (c *stubClaudeClient) Analyze(prompt string) (string, error) {
	log.Println("[stub] Claude Analyze called（ANTHROPIC_API_KEY 未設定のためスタブ応答）")
	return "{}", nil
}

type stubLineClient struct{}

func (c *stubLineClient) Notify(message string) error {
	log.Printf("[stub] LINE Notify: %s（LINE_CHANNEL_ACCESS_TOKEN/LINE_USER_ID 未設定のためスキップ）", message)
	return nil
}

type stubS3Client struct{}

func (c *stubS3Client) Upload(key string, body []byte) (string, error) {
	log.Printf("[stub] S3 Upload key=%s（S3_BUCKET_NAME 未設定のためスキップ）", key)
	return "s3://stub/" + key, nil
}
