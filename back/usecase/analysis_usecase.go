package usecase

import (
	"context"
	"log"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/analysis"
	"github.com/yanatoritakuma/trade-analyzer/back/external"
)

// AnalysisUsecase は分析シグナル・分析設定のユースケース。
type AnalysisUsecase struct {
	analysisRepo analysis.AnalysisLogRepository
	settingRepo  analysis.SettingRepository
	claudeClient external.ClaudeClient
	lineClient   external.LineClient
}

func NewAnalysisUsecase(
	analysisRepo analysis.AnalysisLogRepository,
	settingRepo analysis.SettingRepository,
	claudeClient external.ClaudeClient,
	lineClient external.LineClient,
) *AnalysisUsecase {
	return &AnalysisUsecase{
		analysisRepo: analysisRepo,
		settingRepo:  settingRepo,
		claudeClient: claudeClient,
		lineClient:   lineClient,
	}
}

// RunScheduled は定期実行（平日15:30 JST・EventBridge→Lambda直接呼び出し）のエントリポイント。
//
// 分析パイプライン本体（stock_prices＋watchlist→Claude分析→analysis_logs/trades保存→LINE通知、
// およびウォッチリスト候補提案）は別フェーズで実装する。現状は配線確認用のスタブ（ログのみ）。
// 戻り値のerrorはEventBridgeのリトライ判定に使われるため、未実装段階では正常終了(nil)を返す。
func (u *AnalysisUsecase) RunScheduled(ctx context.Context) error {
	log.Println("[analysis] RunScheduled 呼び出し（分析パイプライン本体は次フェーズで実装）")
	return nil
}

// Latest は直近の分析シグナルを取得する。
func (u *AnalysisUsecase) Latest(ctx context.Context, limit int) ([]*analysis.AnalysisLog, error) {
	return u.analysisRepo.FindLatest(limit)
}

// GetSetting は有効な分析設定を取得する（無ければ空設定を返す）。
func (u *AnalysisUsecase) GetSetting(ctx context.Context) (*analysis.Setting, error) {
	s, err := u.settingRepo.FindActive()
	if err != nil {
		if err == domain.ErrNotFound {
			return &analysis.Setting{ThemeIDs: []int64{}, IsActive: true}, nil
		}
		return nil, err
	}
	return s, nil
}

// SaveSetting は分析設定をUPSERTする（常に1件のみactive）。
func (u *AnalysisUsecase) SaveSetting(ctx context.Context, s *analysis.Setting) (*analysis.Setting, error) {
	if len(s.ThemeIDs) == 0 {
		return nil, domain.NewMessageError(domain.ErrInvalidInput, "テーマを1つ以上選択してください")
	}
	return u.settingRepo.Upsert(s)
}
