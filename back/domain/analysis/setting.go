package analysis

import "time"

// Screening はスクリーニング条件。
type Screening struct {
	MinMarketCap int64   `json:"min_market_cap"`
	MinVolume    int64   `json:"min_volume"`
	MaxPer       float64 `json:"max_per"`
}

// Style は分析スタイル（期間×方向）。
type Style string

// Setting は分析設定エンティティ（is_active=true の1件が有効）。
type Setting struct {
	ID         uint
	ThemeIDs   []int64
	Screening  *Screening
	Style      Style
	FreePrompt string
	IsActive   bool
	CreatedBy  *uint
	UpdatedAt  time.Time
}

// SettingRepository は分析設定永続化のインターフェース。
type SettingRepository interface {
	FindActive() (*Setting, error)
	// Upsert は is_active=true の設定を更新（無ければINSERT）し、常に1件のみactiveを保証する。
	Upsert(s *Setting) (*Setting, error)
}
