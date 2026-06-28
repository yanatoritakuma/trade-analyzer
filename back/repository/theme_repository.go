package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/analysis"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// ThemeRepositoryImpl は analysis.ThemeRepository のGORM実装。
type ThemeRepositoryImpl struct {
	db *gorm.DB
}

func NewThemeRepositoryImpl(db *gorm.DB) analysis.ThemeRepository {
	return &ThemeRepositoryImpl{db: db}
}

func toThemeEntity(m *model.AnalysisTheme) *analysis.Theme {
	return &analysis.Theme{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (r *ThemeRepositoryImpl) FindAll() ([]*analysis.Theme, error) {
	var ms []model.AnalysisTheme
	if err := r.db.Order("sort_order ASC, id ASC").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]*analysis.Theme, 0, len(ms))
	for i := range ms {
		out = append(out, toThemeEntity(&ms[i]))
	}
	return out, nil
}

func (r *ThemeRepositoryImpl) FindByID(id uint) (*analysis.Theme, error) {
	var m model.AnalysisTheme
	if err := r.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toThemeEntity(&m), nil
}

func (r *ThemeRepositoryImpl) ExistsByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&model.AnalysisTheme{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *ThemeRepositoryImpl) Save(t *analysis.Theme) error {
	m := model.AnalysisTheme{
		Name:        t.Name,
		Description: t.Description,
		SortOrder:   t.SortOrder,
		IsActive:    t.IsActive,
		CreatedBy:   t.CreatedBy,
	}
	if err := r.db.Create(&m).Error; err != nil {
		return err
	}
	t.ID = m.ID
	t.CreatedAt = m.CreatedAt
	t.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *ThemeRepositoryImpl) Update(t *analysis.Theme) error {
	return r.db.Model(&model.AnalysisTheme{}).Where("id = ?", t.ID).Updates(map[string]interface{}{
		"name":        t.Name,
		"description": t.Description,
		"is_active":   t.IsActive,
	}).Error
}

func (r *ThemeRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&model.AnalysisTheme{}, id).Error
}

func (r *ThemeRepositoryImpl) UpdateSortOrders(items []analysis.ThemeSortItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Model(&model.AnalysisTheme{}).
				Where("id = ?", item.ID).
				Update("sort_order", item.SortOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
