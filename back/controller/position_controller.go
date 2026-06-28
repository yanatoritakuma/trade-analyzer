package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// PositionController は実運用保有株CRUD（admin）のHTTPハンドラ。
type PositionController struct {
	positionUsecase *usecase.PositionUsecase
}

func NewPositionController(positionUsecase *usecase.PositionUsecase) *PositionController {
	return &PositionController{positionUsecase: positionUsecase}
}

type positionRequest struct {
	Code     string  `json:"code" binding:"required"`
	AvgPrice float64 `json:"avg_price" binding:"required"`
	Quantity int     `json:"quantity" binding:"required"`
}

func (pc *PositionController) Create(c *gin.Context) {
	var req positionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	p, err := pc.positionUsecase.Create(c.Request.Context(), req.Code, req.AvgPrice, req.Quantity)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toPositionDTO(p))
}

func (pc *PositionController) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}
	var req positionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	p, err := pc.positionUsecase.Update(c.Request.Context(), id, req.Code, req.AvgPrice, req.Quantity)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPositionDTO(p))
}

func (pc *PositionController) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}
	if err := pc.positionUsecase.Delete(c.Request.Context(), id); err != nil {
		utils.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// parseIDParam は :id パスパラメータを uint に変換する共通関数。
func parseIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
