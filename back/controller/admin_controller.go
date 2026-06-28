package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// AdminController は管理者によるユーザー管理のHTTPハンドラ。
type AdminController struct {
	adminUsecase *usecase.AdminUsecase
}

func NewAdminController(adminUsecase *usecase.AdminUsecase) *AdminController {
	return &AdminController{adminUsecase: adminUsecase}
}

func (ac *AdminController) ListUsers(c *gin.Context) {
	users, err := ac.adminUsecase.ListUsers(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	out := make([]adminUserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, toAdminUserDTO(u))
	}
	c.JSON(http.StatusOK, out)
}

type userStatusUpdateRequest struct {
	IsActive *bool `json:"is_active" binding:"required"`
}

func (ac *AdminController) SetUserActive(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}
	var req userStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.IsActive == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	usr, err := ac.adminUsecase.SetUserActive(c.Request.Context(), id, *req.IsActive)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAdminUserDTO(usr))
}

func (ac *AdminController) DeleteUser(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}
	if err := ac.adminUsecase.DeleteUser(c.Request.Context(), id); err != nil {
		utils.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
