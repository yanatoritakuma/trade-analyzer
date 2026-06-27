package router

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewRouter はGinエンジンを構築する。
// 環境構築フェーズではヘルスチェックのみを公開し、機能ごとのルートは
// dev-spec の実装時に追加していく。
func NewRouter(database *gorm.DB) *gin.Engine {
	r := gin.Default()

	// 許可オリジン：ローカル開発に加え、設定されていれば本番URLも許可する
	allowOrigins := []string{"http://localhost:3000"}
	if origin := os.Getenv("FRONTEND_ORIGIN"); origin != "" {
		allowOrigins = append(allowOrigins, origin)
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true, // HttpOnly Cookieの送受信に必須
		MaxAge:           12 * time.Hour,
	}))

	// ヘルスチェック：DBへの疎通も確認する
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := database.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "db": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "up"})
	})

	return r
}
