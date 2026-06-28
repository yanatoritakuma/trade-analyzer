package trade

import "time"

// Mode はトレードのモード（virtual/real）。
type Mode string

const (
	ModeVirtual Mode = "virtual"
	ModeReal    Mode = "real"
)

// Action は売買種別（BUY/SELL）。
type Action string

const (
	ActionBuy  Action = "BUY"
	ActionSell Action = "SELL"
)

// Trade はトレード履歴エンティティ。
type Trade struct {
	ID          uint
	UserID      uint
	Ticker      string
	Name        string
	Mode        Mode
	Action      Action
	Price       float64
	Quantity    int
	Confidence  *float64
	Reason      *string
	TargetPrice *float64
	StopLoss    *float64
	ResultPnl   *float64
	ClosedAt    *time.Time
	CreatedAt   time.Time

	// 詳細モーダル用（analysis_logs由来・該当なしはnil）
	BuyReasons     []string
	NoBuyReasons   []string
	EntryCondition *string
}

// Filter はトレード一覧の絞り込み条件。
type Filter struct {
	Mode   Mode // 必須
	Ticker *string
	Action *Action
	From   *time.Time
	To     *time.Time
}

// Summary はフッター集計。
type Summary struct {
	Count    int     `json:"count"`
	WinRate  float64 `json:"win_rate"`
	TotalPnl float64 `json:"total_pnl"`
}
