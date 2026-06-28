package usecase

import (
	"context"
	stdlog "log"
	"time"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/learning"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/trade"
	"github.com/yanatoritakuma/trade-analyzer/back/external"
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
	// claudeClient は週次学習（RunWeekly）でsummary/lessons/strategyを生成する。
	// 週次は低頻度・高レバレッジのため、毎日分析より高性能なモデル（例: Opus 4.8）を割り当てる。
	claudeClient external.ClaudeClient
}

func NewReportUsecase(
	learningRepo learning.LearningLogRepository,
	tradeRepo trade.TradeRepository,
	claudeClient external.ClaudeClient,
) *ReportUsecase {
	return &ReportUsecase{learningRepo: learningRepo, tradeRepo: tradeRepo, claudeClient: claudeClient}
}

// RunWeekly は週次レポート生成（日曜18:00 JST・EventBridge→Lambda直接呼び出し）のエントリポイント。
//
// 当週トレード集計→Claude学習プロンプト→learning_logs保存→学習CSV更新→S3アップロード→
// learning_versions記録 は別フェーズで実装する。現状は配線確認用のスタブ（ログのみ）。
// 戻り値のerrorはEventBridgeのリトライ判定に使われるため、未実装段階では正常終了(nil)を返す。
func (u *ReportUsecase) RunWeekly(ctx context.Context) error {
	stdlog.Println("[report] RunWeekly 呼び出し（週次生成パイプライン本体は次フェーズで実装）")
	return nil
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
