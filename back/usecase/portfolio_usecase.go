package usecase

import (
	"context"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/position"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/stockprice"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/trade"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// ModeSummary はmode別の損益サマリー。
type ModeSummary struct {
	TotalPnl   float64 `json:"total_pnl"`
	WeeklyPnl  float64 `json:"weekly_pnl"`
	WinRate    float64 `json:"win_rate"`
	TradeCount int     `json:"trade_count"`
}

// PortfolioSummary はバーチャル/実運用の損益サマリー。
type PortfolioSummary struct {
	Virtual ModeSummary `json:"virtual"`
	Real    ModeSummary `json:"real"`
}

// PortfolioUsecase は損益サマリー・バーチャル保有のユースケース。
type PortfolioUsecase struct {
	tradeRepo      trade.TradeRepository
	stockPriceRepo stockprice.StockPriceRepository
}

func NewPortfolioUsecase(tradeRepo trade.TradeRepository, stockPriceRepo stockprice.StockPriceRepository) *PortfolioUsecase {
	return &PortfolioUsecase{tradeRepo: tradeRepo, stockPriceRepo: stockPriceRepo}
}

// VirtualPositions は未決済バーチャルを銘柄ごとに集計し、stock_prices を結合して
// 現在値・含み益・損益率を付与して返す（共通ポートフォリオ＝adminのtrades）。
func (u *PortfolioUsecase) VirtualPositions(ctx context.Context) ([]*position.Position, error) {
	opens, err := u.tradeRepo.FindOpenVirtualPositions()
	if err != nil {
		return nil, err
	}
	tickers := make([]string, 0, len(opens))
	for _, o := range opens {
		tickers = append(tickers, o.Ticker)
	}
	prices, err := u.stockPriceRepo.FindByTickers(tickers)
	if err != nil {
		return nil, err
	}
	out := make([]*position.Position, 0, len(opens))
	for _, o := range opens {
		p := &position.Position{
			Ticker:   o.Ticker,
			Name:     o.Name,
			Quantity: o.Quantity,
			AvgPrice: o.AvgPrice,
		}
		if sp, ok := prices[o.Ticker]; ok {
			close := sp.Close
			p.CalcPnl(&close)
		} else {
			p.CalcPnl(nil)
		}
		out = append(out, p)
	}
	return out, nil
}

// Summary は admin の trades から virtual/real の損益サマリーを集計する。
func (u *PortfolioUsecase) Summary(ctx context.Context) (*PortfolioSummary, error) {
	weekStart := utils.CurrentWeekStartJST()

	virtual, err := u.tradeRepo.AggregateByAdmin(trade.ModeVirtual, weekStart)
	if err != nil {
		return nil, err
	}
	real, err := u.tradeRepo.AggregateByAdmin(trade.ModeReal, weekStart)
	if err != nil {
		return nil, err
	}
	return &PortfolioSummary{
		Virtual: ModeSummary{TotalPnl: virtual.TotalPnl, WeeklyPnl: virtual.WeeklyPnl, WinRate: virtual.WinRate, TradeCount: virtual.TradeCount},
		Real:    ModeSummary{TotalPnl: real.TotalPnl, WeeklyPnl: real.WeeklyPnl, WinRate: real.WinRate, TradeCount: real.TradeCount},
	}, nil
}
