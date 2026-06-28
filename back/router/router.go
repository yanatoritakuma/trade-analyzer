package router

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/controller"
	"github.com/yanatoritakuma/trade-analyzer/back/middleware"
)

// Controllers はルーティングに必要な全コントローラを束ねる。
type Controllers struct {
	User       *controller.UserController
	Portfolio  *controller.PortfolioController
	Position   *controller.PositionController
	Watchlist  *controller.WatchlistController
	Trade      *controller.TradeController
	Analysis   *controller.AnalysisController
	Report     *controller.ReportController
	Admin      *controller.AdminController
	Invitation *controller.InvitationController
	Theme      *controller.ThemeController
	Candidate  *controller.CandidateController
	Internal   *controller.InternalController
}

// NewRouter はGinエンジンを構築し、全ルートを登録する。
func NewRouter(database *gorm.DB, c *Controllers) *gin.Engine {
	r := gin.Default()

	allowOrigins := []string{"http://localhost:3000"}
	if origin := os.Getenv("FRONTEND_ORIGIN"); origin != "" {
		allowOrigins = append(allowOrigins, origin)
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ヘルスチェック
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := database.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "db": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "up"})
	})

	api := r.Group("/api")

	// ---- 認証（公開） ----
	auth := api.Group("/auth")
	{
		auth.POST("/login", c.User.Login)
		auth.POST("/register", c.User.Register)
		auth.POST("/refresh", c.User.Refresh)
		auth.POST("/logout", c.User.Logout)
		auth.GET("/me", middleware.JWTAuth(), c.User.Me)
	}

	// ---- 認証必須（protected・user/admin共通） ----
	protected := api.Group("")
	protected.Use(middleware.JWTAuth())
	{
		protected.GET("/portfolio/summary", c.Portfolio.Summary)
		protected.GET("/positions", c.Portfolio.Positions)
		protected.GET("/analysis/latest", c.Analysis.Latest)
		protected.GET("/watchlist", c.Watchlist.GetAll)
		protected.GET("/trades", c.Trade.GetAll)
		protected.GET("/reports", c.Report.GetAll)
		protected.GET("/reports/:week", c.Report.GetByWeek)
	}

	// ---- admin専用 ----
	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth(), middleware.RequireAdmin())
	{
		// プロフィール・保有株
		admin.PATCH("/me", c.User.UpdateProfile)
		admin.PUT("/me/password", c.User.ChangePassword)
		admin.POST("/positions", c.Position.Create)
		admin.PUT("/positions/:id", c.Position.Update)
		admin.DELETE("/positions/:id", c.Position.Delete)

		// ウォッチリスト管理
		admin.POST("/watchlist", c.Watchlist.Create)
		admin.DELETE("/watchlist/:id", c.Watchlist.Delete)

		// ユーザー管理
		admin.GET("/users", c.Admin.ListUsers)
		admin.PATCH("/users/:id", c.Admin.SetUserActive)
		admin.DELETE("/users/:id", c.Admin.DeleteUser)

		// 招待コード
		admin.GET("/invitations", c.Invitation.List)
		admin.POST("/invitations", c.Invitation.Create)
		admin.DELETE("/invitations/:id", c.Invitation.Disable)

		// 分析設定・テーマ
		admin.GET("/analysis-settings", c.Analysis.GetSetting)
		admin.PUT("/analysis-settings", c.Analysis.SaveSetting)
		admin.GET("/analysis-themes", c.Theme.List)
		admin.POST("/analysis-themes", c.Theme.Create)
		admin.PATCH("/analysis-themes/sort", c.Theme.Sort)
		admin.PUT("/analysis-themes/:id", c.Theme.Update)
		admin.DELETE("/analysis-themes/:id", c.Theme.Delete)

		// ウォッチリスト候補
		admin.GET("/watchlist-candidates", c.Candidate.List)
		admin.PATCH("/watchlist-candidates/:id/approve", c.Candidate.Approve)
		admin.PATCH("/watchlist-candidates/:id/reject", c.Candidate.Reject)
	}

	// ---- 内部API（Lambda → Go・X-Internal-Secret認証） ----
	internal := r.Group("/internal")
	internal.Use(middleware.InternalAuth())
	{
		internal.GET("/watchlist", c.Internal.GetWatchlist)
		internal.POST("/stock-prices", c.Internal.IngestStockPrices)
	}

	return r
}
