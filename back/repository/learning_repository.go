package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/learning"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// LearningLogRepositoryImpl は learning.LearningLogRepository のGORM実装。
type LearningLogRepositoryImpl struct {
	db *gorm.DB
}

func NewLearningLogRepositoryImpl(db *gorm.DB) learning.LearningLogRepository {
	return &LearningLogRepositoryImpl{db: db}
}

func toLearningEntity(m *model.LearningLog) *learning.LearningLog {
	return &learning.LearningLog{
		ID:         m.ID,
		WeekStart:  m.WeekStart,
		WeekEnd:    m.WeekEnd,
		TradeCount: m.TradeCount,
		WinRate:    m.WinRate,
		TotalPnl:   m.TotalPnl,
		Summary:    m.Summary,
		Lessons:    m.Lessons,
		Strategy:   m.Strategy,
	}
}

func (r *LearningLogRepositoryImpl) FindAll() ([]*learning.LearningLog, error) {
	var ms []model.LearningLog
	if err := r.db.Order("week_start DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]*learning.LearningLog, 0, len(ms))
	for i := range ms {
		out = append(out, toLearningEntity(&ms[i]))
	}
	return out, nil
}

func (r *LearningLogRepositoryImpl) FindByWeekStart(weekStart time.Time) (*learning.LearningLog, error) {
	var m model.LearningLog
	// date型のため日付一致で取得
	if err := r.db.Where("week_start = ?", weekStart.Format("2006-01-02")).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toLearningEntity(&m), nil
}

func (r *LearningLogRepositoryImpl) Save(l *learning.LearningLog) error {
	m := model.LearningLog{
		WeekStart:  l.WeekStart,
		WeekEnd:    l.WeekEnd,
		TradeCount: l.TradeCount,
		WinRate:    l.WinRate,
		TotalPnl:   l.TotalPnl,
		Summary:    l.Summary,
		Lessons:    l.Lessons,
		Strategy:   l.Strategy,
	}
	if err := r.db.Create(&m).Error; err != nil {
		return err
	}
	l.ID = m.ID
	return nil
}
