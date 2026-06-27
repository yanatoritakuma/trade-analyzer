package model

import "gorm.io/gorm"

// Watchlist は監視銘柄リスト（watchlist）。全ユーザー共通・管理者が管理。
type Watchlist struct {
	gorm.Model
	Ticker   string `gorm:"uniqueIndex;not null"`
	Name     string
	Mode     string `gorm:"not null;check:mode IN ('virtual','real','both')"`
	IsActive bool   `gorm:"default:true"`
}

// TableName はテーブル名を明示する（db_definition では単数形 watchlist）。
func (Watchlist) TableName() string {
	return "watchlist"
}
