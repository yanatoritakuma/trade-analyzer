package position

// Position は実運用保有株エンティティ（real_positions・ユーザーごと）。
// Close/UnrealizedPnl/PnlRate は stock_prices 結合後に算出する参照用フィールド（nilあり）。
type Position struct {
	ID            uint
	UserID        uint
	Ticker        string
	Name          string
	Quantity      int
	AvgPrice      float64
	Close         *float64
	UnrealizedPnl *float64
	PnlRate       *float64
}

// CalcPnl は現在値から含み益・損益率を算出してセットする。
func (p *Position) CalcPnl(close *float64) {
	if close == nil {
		p.Close = nil
		p.UnrealizedPnl = nil
		p.PnlRate = nil
		return
	}
	p.Close = close
	pnl := (*close - p.AvgPrice) * float64(p.Quantity)
	p.UnrealizedPnl = &pnl
	if p.AvgPrice != 0 {
		rate := (*close/p.AvgPrice - 1) * 100
		p.PnlRate = &rate
	}
}
