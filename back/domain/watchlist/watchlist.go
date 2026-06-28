package watchlist

import "time"

// Mode は監視銘柄のモード値オブジェクト。
type Mode string

const (
	ModeVirtual Mode = "virtual"
	ModeReal    Mode = "real"
	ModeBoth    Mode = "both"
)

func (m Mode) IsValid() bool {
	return m == ModeVirtual || m == ModeReal || m == ModeBoth
}

// Watchlist は監視銘柄エンティティ（全ユーザー共通）。
// Close/ChangeRate は stock_prices をJOINして付与する参照用フィールド（nilあり）。
type Watchlist struct {
	ID         uint
	Ticker     string
	Name       string
	Mode       Mode
	IsActive   bool
	Close      *float64
	ChangeRate *float64
	CreatedAt  time.Time
}
