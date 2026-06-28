package repository

import (
	"encoding/json"
	"errors"

	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/analysis"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// SettingRepositoryImpl は analysis.SettingRepository のGORM実装。
type SettingRepositoryImpl struct {
	db *gorm.DB
}

func NewSettingRepositoryImpl(db *gorm.DB) analysis.SettingRepository {
	return &SettingRepositoryImpl{db: db}
}

func toSettingEntity(m *model.AnalysisSetting) *analysis.Setting {
	s := &analysis.Setting{
		ID:         m.ID,
		ThemeIDs:   []int64(m.ThemeIDs),
		Style:      analysis.Style(m.Style),
		FreePrompt: m.FreePrompt,
		IsActive:   m.IsActive,
		CreatedBy:  m.CreatedBy,
		UpdatedAt:  m.UpdatedAt,
	}
	if len(m.Screening) > 0 {
		var sc analysis.Screening
		if err := json.Unmarshal(m.Screening, &sc); err == nil {
			s.Screening = &sc
		}
	}
	if s.ThemeIDs == nil {
		s.ThemeIDs = []int64{}
	}
	return s
}

func (r *SettingRepositoryImpl) FindActive() (*analysis.Setting, error) {
	var m model.AnalysisSetting
	if err := r.db.Where("is_active = ?", true).Order("id DESC").First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toSettingEntity(&m), nil
}

func (r *SettingRepositoryImpl) Upsert(s *analysis.Setting) (*analysis.Setting, error) {
	var screeningJSON datatypes.JSON
	if s.Screening != nil {
		b, err := json.Marshal(s.Screening)
		if err != nil {
			return nil, err
		}
		screeningJSON = datatypes.JSON(b)
	}

	var existing model.AnalysisSetting
	err := r.db.Where("is_active = ?", true).Order("id DESC").First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		m := model.AnalysisSetting{
			ThemeIDs:   pq.Int64Array(s.ThemeIDs),
			Screening:  screeningJSON,
			Style:      string(s.Style),
			FreePrompt: s.FreePrompt,
			IsActive:   true,
			CreatedBy:  s.CreatedBy,
		}
		if err := r.db.Create(&m).Error; err != nil {
			return nil, err
		}
		return toSettingEntity(&m), nil
	}

	// 既存activeを更新（単一active保証）
	updates := map[string]interface{}{
		"theme_ids":   pq.Int64Array(s.ThemeIDs),
		"screening":   screeningJSON,
		"style":       string(s.Style),
		"free_prompt": s.FreePrompt,
		"is_active":   true,
	}
	if err := r.db.Model(&model.AnalysisSetting{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	var reloaded model.AnalysisSetting
	if err := r.db.First(&reloaded, existing.ID).Error; err != nil {
		return nil, err
	}
	return toSettingEntity(&reloaded), nil
}
