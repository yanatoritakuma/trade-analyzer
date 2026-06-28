package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/position"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// PositionRepositoryImpl は position.PositionRepository のGORM実装（ユーザーごと）。
type PositionRepositoryImpl struct {
	db *gorm.DB
}

func NewPositionRepositoryImpl(db *gorm.DB) position.PositionRepository {
	return &PositionRepositoryImpl{db: db}
}

func toPositionEntity(m *model.RealPosition) *position.Position {
	return &position.Position{
		ID:       m.ID,
		UserID:   m.UserID,
		Ticker:   m.Ticker,
		Name:     m.Name,
		Quantity: m.Quantity,
		AvgPrice: m.AvgPrice,
	}
}

func (r *PositionRepositoryImpl) FindByUser(userID uint) ([]*position.Position, error) {
	var ms []model.RealPosition
	if err := r.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]*position.Position, 0, len(ms))
	for i := range ms {
		out = append(out, toPositionEntity(&ms[i]))
	}
	return out, nil
}

func (r *PositionRepositoryImpl) FindByUserWithPrice(userID uint) ([]*position.Position, error) {
	var rows []struct {
		ID       uint
		Ticker   string
		Name     string
		Quantity int
		AvgPrice float64
		Close    *float64
	}
	err := r.db.Table("real_positions rp").
		Select("rp.id, rp.ticker, rp.name, rp.quantity, rp.avg_price, sp.close AS close").
		Joins("LEFT JOIN stock_prices sp ON sp.ticker = rp.ticker AND sp.deleted_at IS NULL").
		Where("rp.user_id = ? AND rp.deleted_at IS NULL", userID).
		Order("rp.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*position.Position, 0, len(rows))
	for _, row := range rows {
		p := &position.Position{
			ID:       row.ID,
			UserID:   userID,
			Ticker:   row.Ticker,
			Name:     row.Name,
			Quantity: row.Quantity,
			AvgPrice: row.AvgPrice,
		}
		p.CalcPnl(row.Close)
		out = append(out, p)
	}
	return out, nil
}

func (r *PositionRepositoryImpl) FindByID(id uint) (*position.Position, error) {
	var m model.RealPosition
	if err := r.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toPositionEntity(&m), nil
}

func (r *PositionRepositoryImpl) Save(p *position.Position) error {
	m := model.RealPosition{
		UserID:   p.UserID,
		Ticker:   p.Ticker,
		Name:     p.Name,
		Quantity: p.Quantity,
		AvgPrice: p.AvgPrice,
	}
	if err := r.db.Create(&m).Error; err != nil {
		return err
	}
	p.ID = m.ID
	return nil
}

func (r *PositionRepositoryImpl) Update(p *position.Position) error {
	return r.db.Model(&model.RealPosition{}).
		Where("id = ? AND user_id = ?", p.ID, p.UserID).
		Updates(map[string]interface{}{
			"ticker":    p.Ticker,
			"name":      p.Name,
			"quantity":  p.Quantity,
			"avg_price": p.AvgPrice,
		}).Error
}

func (r *PositionRepositoryImpl) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.RealPosition{}).Error
}
