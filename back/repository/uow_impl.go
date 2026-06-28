package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
)

// UnitOfWorkImpl はGORMのTransactionを使ってUoWを実装する。
type UnitOfWorkImpl struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) usecase.UnitOfWork {
	return &UnitOfWorkImpl{db: db}
}

func (u *UnitOfWorkImpl) Do(ctx context.Context, fn func(*usecase.Repositories) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repos := &usecase.Repositories{
			User:               NewUserRepositoryImpl(tx),
			InvitationCode:     NewInvitationCodeRepositoryImpl(tx),
			Watchlist:          NewWatchlistRepositoryImpl(tx),
			WatchlistCandidate: NewCandidateRepositoryImpl(tx),
			Trade:              NewTradeRepositoryImpl(tx),
			Position:           NewPositionRepositoryImpl(tx),
			StockPrice:         NewStockPriceRepositoryImpl(tx),
			AnalysisLog:        NewAnalysisLogRepositoryImpl(tx),
			AnalysisTheme:      NewThemeRepositoryImpl(tx),
			AnalysisSetting:    NewSettingRepositoryImpl(tx),
			LearningLog:        NewLearningLogRepositoryImpl(tx),
		}
		return fn(repos)
	})
}
