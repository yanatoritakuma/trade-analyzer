package analysis

import "time"

// Theme は分析テーマエンティティ。
type Theme struct {
	ID          uint
	Name        string
	Description string
	SortOrder   int
	IsActive    bool
	CreatedBy   *uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ThemeSortItem は並び替え用のID・順序ペア。
type ThemeSortItem struct {
	ID        uint
	SortOrder int
}

// ThemeRepository は分析テーマ永続化のインターフェース。
type ThemeRepository interface {
	FindAll() ([]*Theme, error)
	FindByID(id uint) (*Theme, error)
	ExistsByName(name string) (bool, error)
	Save(t *Theme) error
	Update(t *Theme) error
	Delete(id uint) error
	UpdateSortOrders(items []ThemeSortItem) error
}
