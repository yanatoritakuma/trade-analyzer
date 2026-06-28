package stockprice

import "time"

// StockPrice は最新株価スナップショットエンティティ（全ユーザー共通・1銘柄1行）。
type StockPrice struct {
	ID           uint
	Ticker       string
	Date         time.Time
	Open         float64
	High         float64
	Low          float64
	Close        float64
	PrevClose    float64
	ChangeAmount float64
	ChangeRate   float64
	Volume       int64
}

// StockPriceRepository は株価スナップショット永続化のインターフェース。
type StockPriceRepository interface {
	FindByTicker(ticker string) (*StockPrice, error)
	FindByTickers(tickers []string) (map[string]*StockPrice, error)
	// Upsert は ticker をキーにUPSERT（上書き）する。
	Upsert(p *StockPrice) error
}
