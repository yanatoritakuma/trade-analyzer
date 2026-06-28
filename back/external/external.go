// Package external は外部サービス（Claude API・LINE通知・S3）のクライアントを提供する。
// ISSUE #5 の範囲ではローカル動作確認のためスタブ実装を提供し、
// 実連携（APIキー・AWS設定が必要）は次のインフラ開発フェーズで差し替える。
package external

import "log"

// ClaudeClient は株価分析を行うClaude APIクライアントのインターフェース。
type ClaudeClient interface {
	// Analyze は銘柄データを分析し、結果（JSON文字列等）を返す。
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

// --- スタブ実装（ローカル動作確認用） ---

type stubClaudeClient struct{}

func NewClaudeClient() ClaudeClient { return &stubClaudeClient{} }

func (c *stubClaudeClient) Analyze(prompt string) (string, error) {
	log.Println("[stub] Claude Analyze called（インフラ開発フェーズで実装）")
	return "{}", nil
}

type stubLineClient struct{}

func NewLineClient() LineClient { return &stubLineClient{} }

func (c *stubLineClient) Notify(message string) error {
	log.Printf("[stub] LINE Notify: %s（インフラ開発フェーズで実装）", message)
	return nil
}

type stubS3Client struct{}

func NewS3Client() S3Client { return &stubS3Client{} }

func (c *stubS3Client) Upload(key string, body []byte) (string, error) {
	log.Printf("[stub] S3 Upload key=%s（インフラ開発フェーズで実装）", key)
	return "s3://stub/" + key, nil
}
