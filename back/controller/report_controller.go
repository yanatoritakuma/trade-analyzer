package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// ReportController は週次レポートのHTTPハンドラ。
type ReportController struct {
	reportUsecase *usecase.ReportUsecase
}

func NewReportController(reportUsecase *usecase.ReportUsecase) *ReportController {
	return &ReportController{reportUsecase: reportUsecase}
}

func (rc *ReportController) GetAll(c *gin.Context) {
	logs, err := rc.reportUsecase.List(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	out := make([]reportSummaryDTO, 0, len(logs))
	for _, l := range logs {
		out = append(out, toReportSummaryDTO(l))
	}
	c.JSON(http.StatusOK, out)
}

func (rc *ReportController) GetByWeek(c *gin.Context) {
	weekStr := c.Param("week")
	weekStart, err := time.Parse("2006-01-02", weekStr)
	if err != nil {
		utils.HandleError(c, domain.ErrNotFound)
		return
	}
	detail, err := rc.reportUsecase.Detail(c.Request.Context(), weekStart)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	resp := toReportSummaryDTO(detail.Log)
	c.JSON(http.StatusOK, gin.H{
		"week_start":   resp.WeekStart,
		"week_end":     resp.WeekEnd,
		"trade_count":  resp.TradeCount,
		"win_rate":     resp.WinRate,
		"total_pnl":    resp.TotalPnl,
		"max_drawdown": detail.MaxDrawdown,
		"summary":      detail.Log.Summary,
		"lessons":      detail.Log.Lessons,
		"strategy":     detail.Log.Strategy,
		"trades":       toTradeDTOs(detail.Trades),
	})
}
