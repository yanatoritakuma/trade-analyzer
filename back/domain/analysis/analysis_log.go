package analysis

import "time"

// Action は分析シグナルのアクション。
type Action string

const (
	ActionBuy  Action = "BUY"
	ActionSell Action = "SELL"
	ActionHold Action = "HOLD"
)

// AnalysisLog は分析結果ログエンティティ（全ユーザー共通）。
type AnalysisLog struct {
	ID         uint
	Ticker     string
	Name       *string // watchlistからJOINで付与
	Action     Action
	Confidence *float64
	AnalyzedAt time.Time
}

// AnalysisLogRepository は分析ログ永続化のインターフェース。
type AnalysisLogRepository interface {
	FindLatest(limit int) ([]*AnalysisLog, error)
	LatestAnalyzedAt() (*time.Time, error)
	Save(a *AnalysisLog) error
}
