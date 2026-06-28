package usecase

import (
	"context"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/analysis"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/invitation"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/learning"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/position"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/stockprice"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/trade"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/user"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/watchlist"
)

// UnitOfWork はトランザクション境界を宣言するインターフェース。
// ユースケース層はこのIFに依存し、実装（GORM）を知らない。
type UnitOfWork interface {
	Do(ctx context.Context, fn func(repos *Repositories) error) error
}

// Repositories はトランザクション内で使用できる全リポジトリを保持する。
type Repositories struct {
	User               user.UserRepository
	InvitationCode     invitation.InvitationCodeRepository
	Watchlist          watchlist.WatchlistRepository
	WatchlistCandidate watchlist.CandidateRepository
	Trade              trade.TradeRepository
	Position           position.PositionRepository
	StockPrice         stockprice.StockPriceRepository
	AnalysisLog        analysis.AnalysisLogRepository
	AnalysisTheme      analysis.ThemeRepository
	AnalysisSetting    analysis.SettingRepository
	LearningLog        learning.LearningLogRepository
}
