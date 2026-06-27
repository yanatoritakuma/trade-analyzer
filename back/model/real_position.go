package model

import "gorm.io/gorm"

// RealPosition は実運用の保有株（real_positions）。ユーザーごと。
// (user_id, ticker) でユニーク：同一ユーザーの同一銘柄は1レコードのみ。
type RealPosition struct {
	gorm.Model
	UserID   uint    `gorm:"not null;uniqueIndex:idx_real_positions_user_ticker"`
	Ticker   string  `gorm:"not null;uniqueIndex:idx_real_positions_user_ticker"`
	Name     string
	Quantity int     `gorm:"not null"`
	AvgPrice float64 `gorm:"type:numeric(10,2);not null"`
}
