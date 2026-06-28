package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// PortfolioController は損益サマリー・保有株のHTTPハンドラ。
type PortfolioController struct {
	portfolioUsecase *usecase.PortfolioUsecase
	positionUsecase  *usecase.PositionUsecase
}

func NewPortfolioController(portfolioUsecase *usecase.PortfolioUsecase, positionUsecase *usecase.PositionUsecase) *PortfolioController {
	return &PortfolioController{portfolioUsecase: portfolioUsecase, positionUsecase: positionUsecase}
}

func (pc *PortfolioController) Summary(c *gin.Context) {
	s, err := pc.portfolioUsecase.Summary(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPortfolioSummaryDTO(s))
}

// Positions は保有ポジションを返す。?mode=virtual で未決済バーチャル保有（現在値・含み益付き）、
// 既定（mode未指定 or real）で実運用保有株（real_positions）を返す。
func (pc *PortfolioController) Positions(c *gin.Context) {
	if c.Query("mode") == "virtual" {
		ps, err := pc.portfolioUsecase.VirtualPositions(c.Request.Context())
		if err != nil {
			utils.HandleError(c, err)
			return
		}
		out := make([]positionDTO, 0, len(ps))
		for _, p := range ps {
			out = append(out, toPositionDTO(p))
		}
		c.JSON(http.StatusOK, out)
		return
	}

	ps, err := pc.positionUsecase.ListWithPrice(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	out := make([]positionDTO, 0, len(ps))
	for _, p := range ps {
		out = append(out, toPositionDTO(p))
	}
	c.JSON(http.StatusOK, out)
}
