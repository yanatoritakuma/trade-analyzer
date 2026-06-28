package trade

import "time"

// ModeAggregate は損益サマリーのmode別集計値。
type ModeAggregate struct {
	TotalPnl   float64
	WeeklyPnl  float64
	WinRate    float64
	TradeCount int
}

// OpenPosition は未決済バーチャル保有の銘柄別集計。
type OpenPosition struct {
	Ticker   string
	Name     string
	Quantity int
	AvgPrice float64
}

// TradeRepository はトレード永続化のインターフェース。
// 共通ポートフォリオのため参照は admin の trades に固定する。
type TradeRepository interface {
	// FindByFilter は admin の trades をフィルタ付きで取得し、集計も返す。
	FindByFilter(f Filter) ([]*Trade, *Summary, error)
	// AggregateByAdmin は mode別の損益サマリーを返す（weekStart 以降を週次損益とする）。
	AggregateByAdmin(mode Mode, weekStart time.Time) (*ModeAggregate, error)
	// FindOpenVirtualPositions は未決済バーチャルの銘柄別集計を返す。
	FindOpenVirtualPositions() ([]*OpenPosition, error)
	// FindClosedForTimeseries は決済済みトレードを日付昇順で返す（損益推移用）。
	FindClosedForTimeseries(mode Mode, from *time.Time) ([]*Trade, error)
	// FindByAdminBetween は当週レポート用に admin の trades を期間取得する。
	FindByAdminBetween(start, end time.Time) ([]*Trade, error)
	// Save はトレードを保存する（seed・分析側で使用）。
	Save(t *Trade) error
}
