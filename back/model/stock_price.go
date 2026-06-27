package model

import (
	"time"

	"gorm.io/gorm"
)

// StockPrice は最新株価スナップショット（stock_prices）。
// 全ユーザー共通・user_idなし。銘柄ごとに最新の現在値・前日比を1行だけ保持し、
// Ticker のユニークキーでUPSERT（上書き）する。
type StockPrice struct {
	gorm.Model
	Ticker       string    `gorm:"uniqueIndex;not null"`        // UPSERTキー（1銘柄1行）
	Date         time.Time `gorm:"type:date;not null"`          // 最新終値の取引日
	Open         float64   `gorm:"type:numeric(10,2);not null"`
	High         float64   `gorm:"type:numeric(10,2);not null"`
	Low          float64   `gorm:"type:numeric(10,2);not null"`
	Close        float64   `gorm:"type:numeric(10,2);not null"` // 現在値（最新終値）
	PrevClose    float64   `gorm:"type:numeric(10,2)"`          // 前営業日終値
	ChangeAmount float64   `gorm:"type:numeric(10,2)"`          // 前日比（円）
	ChangeRate   float64   `gorm:"type:numeric(6,2)"`           // 前日比（%）
	Volume       int64     `gorm:"not null;default:0"`
}
