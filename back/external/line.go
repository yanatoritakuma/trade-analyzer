package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const linePushURL = "https://api.line.me/v2/bot/message/push"

// realLineClient は LINE Messaging API（push message）で通知する LineClient 実装。
type realLineClient struct {
	accessToken string
	userID      string
	http        *http.Client
}

// NewLineClient は LINE_CHANNEL_ACCESS_TOKEN と LINE_USER_ID が両方設定されていれば
// 実クライアントを、未設定ならスタブを返す。
func NewLineClient() LineClient {
	token := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	userID := os.Getenv("LINE_USER_ID")
	if !isConfigured(token) || !isConfigured(userID) {
		return &stubLineClient{}
	}

	log.Println("[external] LINE client 有効化")
	return &realLineClient{
		accessToken: token,
		userID:      userID,
		http:        &http.Client{Timeout: 15 * time.Second},
	}
}

type linePushRequest struct {
	To       string            `json:"to"`
	Messages []lineTextMessage `json:"messages"`
}

type lineTextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Notify は LINE のユーザーへ push message でテキストを送信する。
func (c *realLineClient) Notify(message string) error {
	reqBody, err := json.Marshal(linePushRequest{
		To:       c.userID,
		Messages: []lineTextMessage{{Type: "text", Text: message}},
	})
	if err != nil {
		return fmt.Errorf("LINE Notify: marshal request failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, linePushURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("LINE Notify: new request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("LINE Notify: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LINE Notify: API returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
