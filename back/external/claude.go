package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	anthropicAPIURL    = "https://api.anthropic.com/v1/messages"
	anthropicVersion   = "2023-06-01"
	defaultClaudeModel = "claude-sonnet-4-6"
	defaultMaxTokens   = 2048
)

// realClaudeClient は Anthropic Messages API を呼び出す ClaudeClient 実装。
type realClaudeClient struct {
	apiKey    string
	model     string
	maxTokens int
	http      *http.Client
}

// NewClaudeClient は ANTHROPIC_API_KEY が設定されていれば実クライアントを、
// 未設定ならスタブを返す。モデルは ANTHROPIC_MODEL→既定(sonnet) の順で解決する。
func NewClaudeClient() ClaudeClient {
	return NewClaudeClientForModel(os.Getenv("ANTHROPIC_MODEL"))
}

// NewClaudeClientForModel は使用モデルを明示指定してクライアントを生成する。
// ジョブごとにモデルを変える用途（例: 毎日分析=Sonnet / 週次学習=Opus）に使う。
// model が空の場合は ANTHROPIC_MODEL→既定(sonnet) の順で解決する。
func NewClaudeClientForModel(model string) ClaudeClient {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if !isConfigured(apiKey) {
		return &stubClaudeClient{}
	}

	if model == "" {
		model = os.Getenv("ANTHROPIC_MODEL")
	}
	if model == "" {
		model = defaultClaudeModel
	}

	maxTokens := defaultMaxTokens
	if v := os.Getenv("ANTHROPIC_MAX_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxTokens = n
		}
	}

	log.Printf("[external] Claude client 有効化（model=%s）", model)
	return &realClaudeClient{
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		http:      &http.Client{Timeout: 60 * time.Second},
	}
}

// claudeRequest は Messages API のリクエストボディ。
type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse は Messages API のレスポンスボディ（必要部分のみ）。
type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Analyze は分析プロンプトを Claude に渡し、生成された本文（テキスト）を返す。
func (c *realClaudeClient) Analyze(prompt string) (string, error) {
	reqBody, err := json.Marshal(claudeRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages:  []claudeMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", fmt.Errorf("Claude Analyze: marshal request failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, anthropicAPIURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("Claude Analyze: new request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("Claude Analyze: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Claude Analyze: read body failed: %w", err)
	}

	var parsed claudeResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("Claude Analyze: unmarshal response failed: %w (body=%s)", err, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", fmt.Errorf("Claude Analyze: API error (%d %s): %s", resp.StatusCode, parsed.Error.Type, parsed.Error.Message)
		}
		return "", fmt.Errorf("Claude Analyze: API returned status %d: %s", resp.StatusCode, string(body))
	}

	// content は複数ブロックになり得るため text タイプを連結する。
	var text string
	for _, b := range parsed.Content {
		if b.Type == "text" {
			text += b.Text
		}
	}
	if text == "" {
		return "", fmt.Errorf("Claude Analyze: empty text in response (body=%s)", string(body))
	}
	return text, nil
}
