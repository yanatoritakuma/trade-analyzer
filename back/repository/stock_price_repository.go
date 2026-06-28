package repository

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/stockprice"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// StockPriceRepositoryImpl は stockprice.StockPriceRepository のGORM実装。
type StockPriceRepositoryImpl struct {
	db *gorm.DB
}

func NewStockPriceRepositoryImpl(db *gorm.DB) stockprice.StockPriceRepository {
	return &StockPriceRepositoryImpl{db: db}
}

func toStockPriceEntity(m *model.StockPrice) *stockprice.StockPrice {
	return &stockprice.StockPrice{
		ID:           m.ID,
		Ticker:       m.Ticker,
		Date:         m.Date,
		Open:         m.Open,
		High:         m.High,
		Low:          m.Low,
		Close:        m.Close,
		PrevClose:    m.PrevClose,
		ChangeAmount: m.ChangeAmount,
		ChangeRate:   m.ChangeRate,
		Volume:       m.Volume,
	}
}

func (r *StockPriceRepositoryImpl) FindByTicker(ticker string) (*stockprice.StockPrice, error) {
	var m model.StockPrice
	if err := r.db.Where("ticker = ?", ticker).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toStockPriceEntity(&m), nil
}

func (r *StockPriceRepositoryImpl) FindByTickers(tickers []string) (map[string]*stockprice.StockPrice, error) {
	out := map[string]*stockprice.StockPrice{}
	if len(tickers) == 0 {
		return out, nil
	}
	var ms []model.StockPrice
	if err := r.db.Where("ticker IN ?", tickers).Find(&ms).Error; err != nil {
		return nil, err
	}
	for i := range ms {
		out[ms[i].Ticker] = toStockPriceEntity(&ms[i])
	}
	return out, nil
}

func (r *StockPriceRepositoryImpl) Upsert(p *stockprice.StockPrice) error {
	m := model.StockPrice{
		Ticker:       p.Ticker,
		Date:         p.Date,
		Open:         p.Open,
		High:         p.High,
		Low:          p.Low,
		Close:        p.Close,
		PrevClose:    p.PrevClose,
		ChangeAmount: p.ChangeAmount,
		ChangeRate:   p.ChangeRate,
		Volume:       p.Volume,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "ticker"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"date", "open", "high", "low", "close",
			"prev_close", "change_amount", "change_rate", "volume", "updated_at",
		}),
	}).Create(&m).Error
}
