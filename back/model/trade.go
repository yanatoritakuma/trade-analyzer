package model

import (
	"time"

	"gorm.io/gorm"
)

// Trade はトレード履歴（trades）。バーチャル・実運用共通。ユーザーごと。
type Trade struct {
	gorm.Model
	UserID      uint   `gorm:"not null;index"`
	Ticker      string `gorm:"not null;index"`
	Name        string
	Mode        string  `gorm:"not null;check:mode IN ('virtual','real');index"`
	Action      string  `gorm:"not null;check:action IN ('BUY','SELL')"`
	Price       float64 `gorm:"type:numeric(10,2);not null"`
	Quantity    int     `gorm:"not null"`
	Confidence  float64 `gorm:"type:numeric(4,3)"`
	Reason      string
	TargetPrice float64 `gorm:"type:numeric(10,2)"`
	StopLoss    float64 `gorm:"type:numeric(10,2)"`
	ResultPnl   float64 `gorm:"type:numeric(10,2)"`
	ClosedAt    *time.Time
}
