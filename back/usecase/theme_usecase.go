package usecase

import (
	"context"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/analysis"
)

// ThemeUsecase は分析テーマ管理のユースケース。
type ThemeUsecase struct {
	themeRepo analysis.ThemeRepository
}

func NewThemeUsecase(themeRepo analysis.ThemeRepository) *ThemeUsecase {
	return &ThemeUsecase{themeRepo: themeRepo}
}

// List はテーマ一覧を sort_order 昇順で返す。
func (u *ThemeUsecase) List(ctx context.Context) ([]*analysis.Theme, error) {
	return u.themeRepo.FindAll()
}

// Create はテーマを追加する（名称重複不可）。
func (u *ThemeUsecase) Create(ctx context.Context, name, description string, isActive bool) (*analysis.Theme, error) {
	if name == "" {
		return nil, domain.NewMessageError(domain.ErrInvalidInput, "テーマ名を入力してください")
	}
	exists, err := u.themeRepo.ExistsByName(name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.NewMessageError(domain.ErrAlreadyExists, "すでに存在するテーマ名です")
	}
	t := &analysis.Theme{Name: name, Description: description, IsActive: isActive}
	if err := u.themeRepo.Save(t); err != nil {
		return nil, err
	}
	return t, nil
}

// Update はテーマを編集する。
func (u *ThemeUsecase) Update(ctx context.Context, id uint, name, description string, isActive bool) (*analysis.Theme, error) {
	t, err := u.themeRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, domain.NewMessageError(domain.ErrInvalidInput, "テーマ名を入力してください")
	}
	if name != t.Name {
		exists, err := u.themeRepo.ExistsByName(name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.NewMessageError(domain.ErrAlreadyExists, "すでに存在するテーマ名です")
		}
	}
	t.Name = name
	t.Description = description
	t.IsActive = isActive
	if err := u.themeRepo.Update(t); err != nil {
		return nil, err
	}
	return t, nil
}

// Delete はテーマを削除する。
func (u *ThemeUsecase) Delete(ctx context.Context, id uint) error {
	if _, err := u.themeRepo.FindByID(id); err != nil {
		return err
	}
	return u.themeRepo.Delete(id)
}

// Sort はテーマの並び順を一括更新する。
func (u *ThemeUsecase) Sort(ctx context.Context, items []analysis.ThemeSortItem) error {
	return u.themeRepo.UpdateSortOrders(items)
}
