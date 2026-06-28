package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// InternalController は内部API（Lambda → Go）のHTTPハンドラ。
type InternalController struct {
	watchlistUsecase *usecase.WatchlistUsecase
	internalUsecase  *usecase.InternalUsecase
}

func NewInternalController(watchlistUsecase *usecase.WatchlistUsecase, internalUsecase *usecase.InternalUsecase) *InternalController {
	return &InternalController{watchlistUsecase: watchlistUsecase, internalUsecase: internalUsecase}
}

type internalWatchlistItemDTO struct {
	ID     uint    `json:"id"`
	Ticker string  `json:"ticker"`
	Name   *string `json:"name"`
	Mode   string  `json:"mode"`
}

func (ic *InternalController) GetWatchlist(c *gin.Context) {
	items, err := ic.watchlistUsecase.FindAllForInternal(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	out := make([]internalWatchlistItemDTO, 0, len(items))
	for _, w := range items {
		var name *string
		if w.Name != "" {
			n := w.Name
			name = &n
		}
		out = append(out, internalWatchlistItemDTO{ID: w.ID, Ticker: w.Ticker, Name: name, Mode: string(w.Mode)})
	}
	c.JSON(http.StatusOK, out)
}

type ohlcvRequest struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

type stockSeriesRequest struct {
	Ticker string         `json:"ticker"`
	Prices []ohlcvRequest `json:"prices"`
}

type stockPricesIngestRequest struct {
	FetchedAt string               `json:"fetched_at"`
	Stocks    []stockSeriesRequest `json:"stocks"`
}

func (ic *InternalController) IngestStockPrices(c *gin.Context) {
	var req stockPricesIngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	stocks := make([]usecase.StockSeriesInput, 0, len(req.Stocks))
	for _, s := range req.Stocks {
		prices := make([]usecase.OhlcvInput, 0, len(s.Prices))
		for _, p := range s.Prices {
			prices = append(prices, usecase.OhlcvInput{
				Date: p.Date, Open: p.Open, High: p.High, Low: p.Low, Close: p.Close, Volume: p.Volume,
			})
		}
		stocks = append(stocks, usecase.StockSeriesInput{Ticker: s.Ticker, Prices: prices})
	}
	count, err := ic.internalUsecase.IngestStockPrices(c.Request.Context(), stocks)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "取り込みました", "count": count})
}
