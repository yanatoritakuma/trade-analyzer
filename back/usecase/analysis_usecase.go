package usecase

import (
	"context"

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
