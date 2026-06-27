package model

import (
	"time"

	"gorm.io/gorm"
)

// LearningLog は週次学習ログ（learning_logs）。全ユーザー共通・管理者のtradesから生成。
type LearningLog struct {
	gorm.Model
	WeekStart   time.Time `gorm:"type:date"`
	WeekEnd     time.Time `gorm:"type:date"`
	TradeCount  int
	WinRate     float64 `gorm:"type:numeric(5,2)"`
	TotalPnl    float64 `gorm:"type:numeric(10,2)"`
	Summary     string
	Lessons     string
	Strategy    string
	RawResponse string
}
