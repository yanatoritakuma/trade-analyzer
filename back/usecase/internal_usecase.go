package usecase

import (
	"context"
	"sort"
	"time"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/stockprice"
)

// OhlcvInput は1日分のOHLCV。
type OhlcvInput struct {
	Date   string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// StockSeriesInput は1銘柄の時系列。
type StockSeriesInput struct {
	Ticker string
	Prices []OhlcvInput
}

// InternalUsecase は内部API（Lambda連携）のユースケース。
type InternalUsecase struct {
	stockPriceRepo stockprice.StockPriceRepository
}

func NewInternalUsecase(stockPriceRepo stockprice.StockPriceRepository) *InternalUsecase {
	return &InternalUsecase{stockPriceRepo: stockPriceRepo}
}

// IngestStockPrices は各銘柄の末尾2営業日から前日比を算出してUPSERTする。
func (u *InternalUsecase) IngestStockPrices(ctx context.Context, stocks []StockSeriesInput) (int, error) {
	count := 0
	for _, s := range stocks {
		if len(s.Prices) == 0 {
			continue
		}
		prices := make([]OhlcvInput, len(s.Prices))
		copy(prices, s.Prices)
		sort.Slice(prices, func(i, j int) bool { return prices[i].Date < prices[j].Date })

		latest := prices[len(prices)-1]
		date, err := time.Parse("2006-01-02", latest.Date)
		if err != nil {
			date = time.Now()
		}

		sp := &stockprice.StockPrice{
			Ticker: s.Ticker,
			Date:   date,
			Open:   latest.Open,
			High:   latest.High,
			Low:    latest.Low,
			Close:  latest.Close,
			Volume: latest.Volume,
		}
		if len(prices) >= 2 {
			prev := prices[len(prices)-2]
			sp.PrevClose = prev.Close
			sp.ChangeAmount = round2(latest.Close - prev.Close)
			if prev.Close != 0 {
				sp.ChangeRate = round2((latest.Close/prev.Close - 1) * 100)
			}
		}
		if err := u.stockPriceRepo.Upsert(sp); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func round2(v float64) float64 {
	return float64(int64(v*100+sign(v)*0.5)) / 100
}

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}
