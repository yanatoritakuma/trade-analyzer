package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/watchlist"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// WatchlistRepositoryImpl は watchlist.WatchlistRepository のGORM実装。
type WatchlistRepositoryImpl struct {
	db *gorm.DB
}

func NewWatchlistRepositoryImpl(db *gorm.DB) watchlist.WatchlistRepository {
	return &WatchlistRepositoryImpl{db: db}
}

func toWatchlistEntity(m *model.Watchlist) *watchlist.Watchlist {
	return &watchlist.Watchlist{
		ID:        m.ID,
		Ticker:    m.Ticker,
		Name:      m.Name,
		Mode:      watchlist.Mode(m.Mode),
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
	}
}

func (r *WatchlistRepositoryImpl) FindAll() ([]*watchlist.Watchlist, error) {
	var ms []model.Watchlist
	if err := r.db.Where("is_active = ?", true).Order("created_at ASC").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]*watchlist.Watchlist, 0, len(ms))
	for i := range ms {
		out = append(out, toWatchlistEntity(&ms[i]))
	}
	return out, nil
}

// watchlistPriceRow はJOIN結果のスキャン用。
type watchlistPriceRow struct {
	ID         uint
	Ticker     string
	Name       string
	Mode       string
	IsActive   bool
	Close      *float64
	ChangeRate *float64
}

func (r *WatchlistRepositoryImpl) FindAllWithPrice() ([]*watchlist.Watchlist, error) {
	var rows []watchlistPriceRow
	err := r.db.Table("watchlist").
		Select("watchlist.id, watchlist.ticker, watchlist.name, watchlist.mode, watchlist.is_active, sp.close AS close, sp.change_rate AS change_rate").
		Joins("LEFT JOIN stock_prices sp ON sp.ticker = watchlist.ticker AND sp.deleted_at IS NULL").
		Where("watchlist.is_active = ? AND watchlist.deleted_at IS NULL", true).
		Order("watchlist.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*watchlist.Watchlist, 0, len(rows))
	for _, row := range rows {
		out = append(out, &watchlist.Watchlist{
			ID:         row.ID,
			Ticker:     row.Ticker,
			Name:       row.Name,
			Mode:       watchlist.Mode(row.Mode),
			IsActive:   row.IsActive,
			Close:      row.Close,
			ChangeRate: row.ChangeRate,
		})
	}
	return out, nil
}

func (r *WatchlistRepositoryImpl) FindByID(id uint) (*watchlist.Watchlist, error) {
	var m model.Watchlist
	if err := r.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toWatchlistEntity(&m), nil
}

func (r *WatchlistRepositoryImpl) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&model.Watchlist{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

func (r *WatchlistRepositoryImpl) ExistsByTicker(ticker string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Watchlist{}).Where("ticker = ?", ticker).Count(&count).Error
	return count > 0, err
}

func (r *WatchlistRepositoryImpl) Save(w *watchlist.Watchlist) error {
	m := model.Watchlist{
		Ticker:   w.Ticker,
		Name:     w.Name,
		Mode:     string(w.Mode),
		IsActive: w.IsActive,
	}
	if err := r.db.Create(&m).Error; err != nil {
		return err
	}
	w.ID = m.ID
	w.CreatedAt = m.CreatedAt
	return nil
}

func (r *WatchlistRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&model.Watchlist{}, id).Error
}

func (r *WatchlistRepositoryImpl) DeleteByTicker(ticker string) error {
	return r.db.Where("ticker = ?", ticker).Delete(&model.Watchlist{}).Error
}
