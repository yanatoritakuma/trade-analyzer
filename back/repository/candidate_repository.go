package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/watchlist"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// CandidateRepositoryImpl は watchlist.CandidateRepository のGORM実装。
type CandidateRepositoryImpl struct {
	db *gorm.DB
}

func NewCandidateRepositoryImpl(db *gorm.DB) watchlist.CandidateRepository {
	return &CandidateRepositoryImpl{db: db}
}

func toCandidateEntity(m *model.WatchlistCandidate) *watchlist.Candidate {
	return &watchlist.Candidate{
		ID:            m.ID,
		Ticker:        m.Ticker,
		Name:          m.Name,
		Reason:        m.Reason,
		ReplaceTicker: m.ReplaceTicker,
		Confidence:    m.Confidence,
		Status:        watchlist.CandidateStatus(m.Status),
		ProposedAt:    m.ProposedAt,
		DecidedAt:     m.DecidedAt,
		DecidedBy:     m.DecidedBy,
	}
}

func (r *CandidateRepositoryImpl) FindAll() ([]*watchlist.Candidate, error) {
	var ms []model.WatchlistCandidate
	if err := r.db.Order("proposed_at DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]*watchlist.Candidate, 0, len(ms))
	for i := range ms {
		out = append(out, toCandidateEntity(&ms[i]))
	}
	return out, nil
}

func (r *CandidateRepositoryImpl) FindByID(id uint) (*watchlist.Candidate, error) {
	var m model.WatchlistCandidate
	if err := r.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toCandidateEntity(&m), nil
}

func (r *CandidateRepositoryImpl) Save(c *watchlist.Candidate) error {
	m := model.WatchlistCandidate{
		Ticker:        c.Ticker,
		Name:          c.Name,
		Reason:        c.Reason,
		ReplaceTicker: c.ReplaceTicker,
		Confidence:    c.Confidence,
		Status:        string(c.Status),
	}
	if err := r.db.Create(&m).Error; err != nil {
		return err
	}
	c.ID = m.ID
	c.ProposedAt = m.ProposedAt
	return nil
}

func (r *CandidateRepositoryImpl) Update(c *watchlist.Candidate) error {
	return r.db.Model(&model.WatchlistCandidate{}).Where("id = ?", c.ID).Updates(map[string]interface{}{
		"status":     string(c.Status),
		"decided_at": c.DecidedAt,
		"decided_by": c.DecidedBy,
	}).Error
}
