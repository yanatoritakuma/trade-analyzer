package repository

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/analysis"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// AnalysisLogRepositoryImpl は analysis.AnalysisLogRepository のGORM実装。
// analyzed_at は model.CreatedAt（gorm.Model）を使用する。
type AnalysisLogRepositoryImpl struct {
	db *gorm.DB
}

func NewAnalysisLogRepositoryImpl(db *gorm.DB) analysis.AnalysisLogRepository {
	return &AnalysisLogRepositoryImpl{db: db}
}

func (r *AnalysisLogRepositoryImpl) FindLatest(limit int) ([]*analysis.AnalysisLog, error) {
	if limit <= 0 {
		limit = 3
	}
	var rows []struct {
		ID         uint
		Ticker     string
		Name       *string
		Action     string
		Confidence *float64
		AnalyzedAt time.Time
	}
	err := r.db.Table("analysis_logs al").
		Select("al.id, al.ticker, w.name AS name, al.action, al.confidence, al.created_at AS analyzed_at").
		Joins("LEFT JOIN watchlist w ON w.ticker = al.ticker AND w.deleted_at IS NULL").
		Where("al.deleted_at IS NULL").
		Order("al.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*analysis.AnalysisLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, &analysis.AnalysisLog{
			ID:         row.ID,
			Ticker:     row.Ticker,
			Name:       row.Name,
			Action:     analysis.Action(row.Action),
			Confidence: row.Confidence,
			AnalyzedAt: row.AnalyzedAt,
		})
	}
	return out, nil
}

func (r *AnalysisLogRepositoryImpl) LatestAnalyzedAt() (*time.Time, error) {
	var m model.AnalysisLog
	err := r.db.Order("created_at DESC").First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	t := m.CreatedAt
	return &t, nil
}

func (r *AnalysisLogRepositoryImpl) Save(a *analysis.AnalysisLog) error {
	m := model.AnalysisLog{
		Ticker:   a.Ticker,
		Action:   string(a.Action),
		Analysis: datatypes.JSON([]byte("{}")),
	}
	if a.Confidence != nil {
		m.Confidence = *a.Confidence
	}
	if !a.AnalyzedAt.IsZero() {
		m.CreatedAt = a.AnalyzedAt
	}
	if err := r.db.Create(&m).Error; err != nil {
		return err
	}
	a.ID = m.ID
	return nil
}
