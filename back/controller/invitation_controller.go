package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/middleware"
	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// InvitationController は招待コード管理のHTTPハンドラ。
type InvitationController struct {
	invitationUsecase *usecase.InvitationUsecase
}

func NewInvitationController(invitationUsecase *usecase.InvitationUsecase) *InvitationController {
	return &InvitationController{invitationUsecase: invitationUsecase}
}

func (ic *InvitationController) List(c *gin.Context) {
	invs, err := ic.invitationUsecase.List(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	out := make([]invitationDTO, 0, len(invs))
	for _, i := range invs {
		out = append(out, toInvitationDTO(i))
	}
	c.JSON(http.StatusOK, out)
}

type invitationCreateRequest struct {
	ExpiresDays int `json:"expires_days" binding:"required"`
}

func (ic *InvitationController) Create(c *gin.Context) {
	var req invitationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	adminID := middleware.UserID(c)
	inv, err := ic.invitationUsecase.Create(c.Request.Context(), req.ExpiresDays, adminID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": inv.Code, "expires_at": inv.ExpiresAt})
}

func (ic *InvitationController) Disable(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}
	if err := ic.invitationUsecase.Disable(c.Request.Context(), id); err != nil {
		utils.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
