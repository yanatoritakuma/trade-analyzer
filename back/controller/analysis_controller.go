package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/analysis"
	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// AnalysisController は分析シグナル・分析設定のHTTPハンドラ。
type AnalysisController struct {
	analysisUsecase *usecase.AnalysisUsecase
}

func NewAnalysisController(analysisUsecase *usecase.AnalysisUsecase) *AnalysisController {
	return &AnalysisController{analysisUsecase: analysisUsecase}
}

func (ac *AnalysisController) Latest(c *gin.Context) {
	limit := 3
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	logs, err := ac.analysisUsecase.Latest(c.Request.Context(), limit)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	out := make([]analysisSignalDTO, 0, len(logs))
	for _, a := range logs {
		out = append(out, toAnalysisSignalDTO(a))
	}
	c.JSON(http.StatusOK, out)
}

func (ac *AnalysisController) GetSetting(c *gin.Context) {
	s, err := ac.analysisUsecase.GetSetting(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAnalysisSettingDTO(s))
}

type analysisSettingRequest struct {
	ThemeIDs   []int64              `json:"theme_ids" binding:"required"`
	Screening  *analysis.Screening  `json:"screening"`
	Style      string               `json:"style"`
	FreePrompt string               `json:"free_prompt"`
}

func (ac *AnalysisController) SaveSetting(c *gin.Context) {
	var req analysisSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	s := &analysis.Setting{
		ThemeIDs:   req.ThemeIDs,
		Screening:  req.Screening,
		Style:      analysis.Style(req.Style),
		FreePrompt: req.FreePrompt,
		IsActive:   true,
	}
	saved, err := ac.analysisUsecase.SaveSetting(c.Request.Context(), s)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAnalysisSettingDTO(saved))
}
