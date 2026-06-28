package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/analysis"
	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// ThemeController は分析テーマ管理のHTTPハンドラ。
type ThemeController struct {
	themeUsecase *usecase.ThemeUsecase
}

func NewThemeController(themeUsecase *usecase.ThemeUsecase) *ThemeController {
	return &ThemeController{themeUsecase: themeUsecase}
}

func (tc *ThemeController) List(c *gin.Context) {
	themes, err := tc.themeUsecase.List(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	out := make([]themeDTO, 0, len(themes))
	for _, t := range themes {
		out = append(out, toThemeDTO(t))
	}
	c.JSON(http.StatusOK, out)
}

type themeRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

func (req themeRequest) active() bool {
	if req.IsActive == nil {
		return true
	}
	return *req.IsActive
}

func (tc *ThemeController) Create(c *gin.Context) {
	var req themeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "テーマ名を入力してください"})
		return
	}
	t, err := tc.themeUsecase.Create(c.Request.Context(), req.Name, req.Description, req.active())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toThemeDTO(t))
}

func (tc *ThemeController) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}
	var req themeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "テーマ名を入力してください"})
		return
	}
	t, err := tc.themeUsecase.Update(c.Request.Context(), id, req.Name, req.Description, req.active())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toThemeDTO(t))
}

func (tc *ThemeController) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}
	if err := tc.themeUsecase.Delete(c.Request.Context(), id); err != nil {
		utils.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type themeSortItemRequest struct {
	ID        uint `json:"id" binding:"required"`
	SortOrder int  `json:"sort_order"`
}

func (tc *ThemeController) Sort(c *gin.Context) {
	var req []themeSortItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	items := make([]analysis.ThemeSortItem, 0, len(req))
	for _, r := range req {
		items = append(items, analysis.ThemeSortItem{ID: r.ID, SortOrder: r.SortOrder})
	}
	if err := tc.themeUsecase.Sort(c.Request.Context(), items); err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "並び替えました"})
}
