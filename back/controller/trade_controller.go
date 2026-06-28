package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/trade"
	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// TradeController はトレード履歴のHTTPハンドラ。
type TradeController struct {
	tradeUsecase *usecase.TradeUsecase
}

func NewTradeController(tradeUsecase *usecase.TradeUsecase) *TradeController {
	return &TradeController{tradeUsecase: tradeUsecase}
}

func (tc *TradeController) GetAll(c *gin.Context) {
	mode := c.Query("mode")
	if mode != "virtual" && mode != "real" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode は virtual または real を指定してください"})
		return
	}
	filter := trade.Filter{Mode: trade.Mode(mode)}

	if ticker := c.Query("ticker"); ticker != "" {
		filter.Ticker = &ticker
	}
	if action := c.Query("action"); action != "" {
		if action != "BUY" && action != "SELL" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "action は BUY または SELL を指定してください"})
			return
		}
		a := trade.Action(action)
		filter.Action = &a
	}
	if from := c.Query("from"); from != "" {
		t, err := time.Parse("2006-01-02", from)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "from の日付形式が不正です"})
			return
		}
		filter.From = &t
	}
	if to := c.Query("to"); to != "" {
		t, err := time.Parse("2006-01-02", to)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "to の日付形式が不正です"})
			return
		}
		filter.To = &t
	}

	items, summary, err := tc.tradeUsecase.List(c.Request.Context(), filter)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	if summary == nil {
		summary = &trade.Summary{}
	}
	c.JSON(http.StatusOK, gin.H{
		"items":   toTradeDTOs(items),
		"summary": summary,
	})
}
