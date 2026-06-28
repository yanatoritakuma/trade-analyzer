package usecase

import (
	"context"
	"time"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/learning"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/trade"
)

// ReportDetail は週次レポート詳細（当週トレード・最大DDを含む）。
type ReportDetail struct {
	Log         *learning.LearningLog
	MaxDrawdown *float64
	Trades      []*trade.Trade
}

// ReportUsecase は週次レポートのユースケース。
type ReportUsecase struct {
	learningRepo learning.LearningLogRepository
	tradeRepo    trade.TradeRepository
}

func NewReportUsecase(learningRepo learning.LearningLogRepository, tradeRepo trade.TradeRepository) *ReportUsecase {
	return &ReportUsecase{learningRepo: learningRepo, tradeRepo: tradeRepo}
}

// List は週次レポート一覧を新しい順で返す。
func (u *ReportUsecase) List(ctx context.Context) ([]*learning.LearningLog, error) {
	return u.learningRepo.FindAll()
}

// Detail は week_start のレポートを取得し、当週トレードと最大DDを付与して返す。
func (u *ReportUsecase) Detail(ctx context.Context, weekStart time.Time) (*ReportDetail, error) {
	log, err := u.learningRepo.FindByWeekStart(weekStart)
	if err != nil {
		return nil, err
	}
	trades, err := u.tradeRepo.FindByAdminBetween(log.WeekStart, log.WeekEnd)
	if err != nil {
		return nil, err
	}
	dd := calcMaxDrawdown(trades)
	return &ReportDetail{Log: log, MaxDrawdown: dd, Trades: trades}, nil
}

// calcMaxDrawdown は決済トレードの累積損益曲線からピーク−ボトムの最大下落幅を算出する。
func calcMaxDrawdown(trades []*trade.Trade) *float64 {
	var cumulative, peak, maxDD float64
	hasClosed := false
	for _, t := range trades {
		if t.ClosedAt == nil || t.ResultPnl == nil {
			continue
		}
		hasClosed = true
		cumulative += *t.ResultPnl
		if cumulative > peak {
			peak = cumulative
		}
		drawdown := peak - cumulative
		if drawdown > maxDD {
			maxDD = drawdown
		}
	}
	if !hasClosed {
		return nil
	}
	return &maxDD
}
