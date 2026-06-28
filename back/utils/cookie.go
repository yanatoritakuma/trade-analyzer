package utils

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	accessTokenMaxAge  = 24 * 60 * 60      // 24時間（秒）
	refreshTokenMaxAge = 30 * 24 * 60 * 60 // 30日（秒）
)

// SetAuthCookies はアクセス/リフレッシュトークンをHttpOnly Cookieで発行する。
// SameSite=Strict・HttpOnly は必須、Secure は本番（APP_ENV=production）のみ付与する。
func SetAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	secure := os.Getenv("APP_ENV") == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("access_token", accessToken, accessTokenMaxAge, "/", "", secure, true)
	c.SetCookie("refresh_token", refreshToken, refreshTokenMaxAge, "/", "", secure, true)
}

// ClearAuthCookies は認証Cookieを削除する（ログアウト）。
func ClearAuthCookies(c *gin.Context) {
	secure := os.Getenv("APP_ENV") == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("access_token", "", -1, "/", "", secure, true)
	c.SetCookie("refresh_token", "", -1, "/", "", secure, true)
}
