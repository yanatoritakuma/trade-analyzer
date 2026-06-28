package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/watchlist"
	"github.com/yanatoritakuma/trade-analyzer/back/middleware"
	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// CandidateController はウォッチリスト候補承認のHTTPハンドラ。
type CandidateController struct {
	candidateUsecase *usecase.CandidateUsecase
}

func NewCandidateController(candidateUsecase *usecase.CandidateUsecase) *CandidateController {
	return &CandidateController{candidateUsecase: candidateUsecase}
}

func (cc *CandidateController) List(c *gin.Context) {
	cands, err := cc.candidateUsecase.List(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	out := make([]candidateDTO, 0, len(cands))
	for _, cand := range cands {
		out = append(out, toCandidateDTO(cand))
	}
	c.JSON(http.StatusOK, out)
}

type candidateApproveRequest struct {
	Mode string `json:"mode"`
}

func (cc *CandidateController) Approve(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}
	var req candidateApproveRequest
	_ = c.ShouldBindJSON(&req) // body任意
	adminID := middleware.UserID(c)
	cand, err := cc.candidateUsecase.Approve(c.Request.Context(), id, adminID, watchlist.Mode(req.Mode))
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCandidateDTO(cand))
}

func (cc *CandidateController) Reject(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}
	adminID := middleware.UserID(c)
	cand, err := cc.candidateUsecase.Reject(c.Request.Context(), id, adminID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCandidateDTO(cand))
}
