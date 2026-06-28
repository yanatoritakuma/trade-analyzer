// Package middleware はGinのHTTPミドルウェア（認証・認可）を提供する。
package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// コンテキストキー。
const (
	CtxUserID = "user_id"
	CtxRole   = "role"
)

// JWTAuth は access_token Cookie を検証し、user_id / role をコンテキストに格納する。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
			return
		}
		claims, err := utils.ParseToken(token)
		if err != nil || claims.Type != utils.AccessToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// RequireAdmin は role=admin を検証する。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get(CtxRole)
		if !ok || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "権限がありません"})
			return
		}
		c.Next()
	}
}

// InternalAuth は X-Internal-Secret ヘッダで内部API（Lambda）を認証する。
func InternalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := os.Getenv("INTERNAL_API_SECRET")
		if secret == "" || c.GetHeader("X-Internal-Secret") != secret {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
			return
		}
		c.Next()
	}
}

// UserID はコンテキストから user_id を取得する。
func UserID(c *gin.Context) uint {
	if v, ok := c.Get(CtxUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}
