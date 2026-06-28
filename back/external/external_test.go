package external

import "testing"

func TestIsConfigured(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"空文字は未設定", "", false},
		{"プレースホルダ(xxxxx)は未設定", "sk-ant-xxxxx", false},
		{"プレースホルダ(xxxxx)は未設定2", "xxxxx", false},
		{"実値は設定済み", "sk-ant-api03-real-value", true},
		{"バケット名は設定済み", "trading-system-learning", true},
	}
	for _, c := range cases {
		if got := isConfigured(c.value); got != c.want {
			t.Errorf("%s: isConfigured(%q)=%v, want %v", c.name, c.value, got, c.want)
		}
	}
}

func TestNewClaudeClient_FallsBackToStub(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "") // 未設定
	if _, ok := NewClaudeClient().(*stubClaudeClient); !ok {
		t.Fatal("APIキー未設定時はスタブを返すべき")
	}

	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-xxxxx") // プレースホルダ
	if _, ok := NewClaudeClient().(*stubClaudeClient); !ok {
		t.Fatal("プレースホルダ時はスタブを返すべき")
	}

	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-real-key")
	if _, ok := NewClaudeClient().(*realClaudeClient); !ok {
		t.Fatal("APIキー設定時は実クライアントを返すべき")
	}
}

func TestNewClaudeClient_ModelDefaults(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-real-key")
	t.Setenv("ANTHROPIC_MODEL", "")
	c, ok := NewClaudeClient().(*realClaudeClient)
	if !ok {
		t.Fatal("実クライアントを返すべき")
	}
	if c.model != defaultClaudeModel {
		t.Errorf("モデル未指定時の既定が不正: got %q, want %q", c.model, defaultClaudeModel)
	}
	if c.maxTokens != defaultMaxTokens {
		t.Errorf("maxTokens 未指定時の既定が不正: got %d, want %d", c.maxTokens, defaultMaxTokens)
	}

	t.Setenv("ANTHROPIC_MODEL", "claude-opus-4-8")
	t.Setenv("ANTHROPIC_MAX_TOKENS", "4096")
	c2 := NewClaudeClient().(*realClaudeClient)
	if c2.model != "claude-opus-4-8" {
		t.Errorf("モデル指定が反映されていない: got %q", c2.model)
	}
	if c2.maxTokens != 4096 {
		t.Errorf("maxTokens 指定が反映されていない: got %d", c2.maxTokens)
	}
}

func TestNewClaudeClientForModel(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-real-key")
	t.Setenv("ANTHROPIC_MODEL", "claude-sonnet-4-6")

	// 明示指定したモデルが ANTHROPIC_MODEL より優先される（週次=Opus 等の用途）
	c, ok := NewClaudeClientForModel("claude-opus-4-8").(*realClaudeClient)
	if !ok {
		t.Fatal("実クライアントを返すべき")
	}
	if c.model != "claude-opus-4-8" {
		t.Errorf("明示モデルが反映されていない: got %q, want claude-opus-4-8", c.model)
	}

	// 空指定時は ANTHROPIC_MODEL にフォールバック
	c2 := NewClaudeClientForModel("").(*realClaudeClient)
	if c2.model != "claude-sonnet-4-6" {
		t.Errorf("空指定時はANTHROPIC_MODELを使うべき: got %q", c2.model)
	}

	// APIキー未設定ならスタブ
	t.Setenv("ANTHROPIC_API_KEY", "")
	if _, ok := NewClaudeClientForModel("claude-opus-4-8").(*stubClaudeClient); !ok {
		t.Fatal("APIキー未設定時はスタブを返すべき")
	}
}

func TestNewLineClient_FallsBackToStub(t *testing.T) {
	// トークンのみ設定（userID未設定）→ スタブ
	t.Setenv("LINE_CHANNEL_ACCESS_TOKEN", "real-token")
	t.Setenv("LINE_USER_ID", "")
	if _, ok := NewLineClient().(*stubLineClient); !ok {
		t.Fatal("userID未設定時はスタブを返すべき")
	}

	// 両方設定 → 実クライアント
	t.Setenv("LINE_USER_ID", "U1234567890")
	if _, ok := NewLineClient().(*realLineClient); !ok {
		t.Fatal("トークンとuserID設定時は実クライアントを返すべき")
	}
}

func TestNewS3Client_FallsBackToStubWhenNoBucket(t *testing.T) {
	t.Setenv("S3_BUCKET_NAME", "")
	if _, ok := NewS3Client().(*stubS3Client); !ok {
		t.Fatal("バケット未設定時はスタブを返すべき")
	}
}
