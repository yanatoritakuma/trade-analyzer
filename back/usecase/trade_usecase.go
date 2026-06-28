package usecase

import (
	"context"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/trade"
)

// TradeUsecase はトレード履歴のユースケース。
type TradeUsecase struct {
	tradeRepo trade.TradeRepository
}

func NewTradeUsecase(tradeRepo trade.TradeRepository) *TradeUsecase {
	return &TradeUsecase{tradeRepo: tradeRepo}
}

// List は admin の trades をフィルタ付きで取得し、items と summary を返す。
func (u *TradeUsecase) List(ctx context.Context, f trade.Filter) ([]*trade.Trade, *trade.Summary, error) {
	return u.tradeRepo.FindByFilter(f)
}
