package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/middleware"
	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// UserController は認証・ユーザー操作のHTTPハンドラ。
type UserController struct {
	userUsecase *usecase.UserUsecase
}

func NewUserController(userUsecase *usecase.UserUsecase) *UserController {
	return &UserController{userUsecase: userUsecase}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (uc *UserController) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	usr, access, refresh, err := uc.userUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	utils.SetAuthCookies(c, access, refresh)
	c.JSON(http.StatusOK, gin.H{"message": "ログインしました", "user": toUserDTO(usr)})
}

type registerRequest struct {
	InvitationCode string `json:"invitation_code" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email" binding:"required"`
	Password       string `json:"password" binding:"required"`
}

func (uc *UserController) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	usr, err := uc.userUsecase.Register(c.Request.Context(), req.InvitationCode, req.Name, req.Email, req.Password)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "アカウントを作成しました", "user": toUserDTO(usr)})
}

func (uc *UserController) Me(c *gin.Context) {
	userID := middleware.UserID(c)
	usr, err := uc.userUsecase.Me(c.Request.Context(), userID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserDTO(usr))
}

func (uc *UserController) Refresh(c *gin.Context) {
	rt, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}
	access, refresh, err := uc.userUsecase.Refresh(c.Request.Context(), rt)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	utils.SetAuthCookies(c, access, refresh)
	c.JSON(http.StatusOK, gin.H{"message": "再発行しました"})
}

func (uc *UserController) Logout(c *gin.Context) {
	utils.ClearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "ログアウトしました"})
}

type profileUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}

func (uc *UserController) UpdateProfile(c *gin.Context) {
	var req profileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	userID := middleware.UserID(c)
	usr, err := uc.userUsecase.UpdateProfile(c.Request.Context(), userID, req.Name)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserDTO(usr))
}

type passwordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

func (uc *UserController) ChangePassword(c *gin.Context) {
	var req passwordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です"})
		return
	}
	userID := middleware.UserID(c)
	if err := uc.userUsecase.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "パスワードを変更しました"})
}
