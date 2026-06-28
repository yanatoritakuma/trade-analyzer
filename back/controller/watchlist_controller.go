package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/watchlist"
	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// WatchlistController はウォッチリストのHTTPハンドラ。
type WatchlistController struct {
	watchlistUsecase *usecase.WatchlistUsecase
}

func NewWatchlistController(watchlistUsecase *usecase.WatchlistUsecase) *WatchlistController {
	return &WatchlistController{watchlistUsecase: watchlistUsecase}
}

func (wc *WatchlistController) GetAll(c *gin.Context) {
	items, err := wc.watchlistUsecase.FindAllWithPrice(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	out := make([]watchlistItemDTO, 0, len(items))
	for _, w := range items {
		out = append(out, toWatchlistItemDTO(w))
	}
	c.JSON(http.StatusOK, out)
}

type watchlistCreateRequest struct {
	Code string `json:"code" binding:"required"`
	Mode string `json:"mode" binding:"required"`
}

func (wc *WatchlistController) Create(c *gin.Context) {
	var req watchlistCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	w, err := wc.watchlistUsecase.Create(c.Request.Context(), req.Code, watchlist.Mode(req.Mode))
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toWatchlistItemDTO(w))
}

func (wc *WatchlistController) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}
	if err := wc.watchlistUsecase.Delete(c.Request.Context(), id); err != nil {
		utils.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
