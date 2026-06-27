package model

import (
	"time"

	"gorm.io/gorm"
)

// WatchlistCandidate はAIが提案したウォッチリスト候補（watchlist_candidates）。
type WatchlistCandidate struct {
	gorm.Model
	Ticker        string `gorm:"not null"`
	Name          string
	Reason        string
	ReplaceTicker string
	Confidence    float64   `gorm:"type:numeric(4,3)"`
	Status        string    `gorm:"default:pending;check:status IN ('pending','approved','rejected');index"`
	ProposedAt    time.Time `gorm:"autoCreateTime"`
	DecidedAt     *time.Time
	DecidedBy     *uint
}
