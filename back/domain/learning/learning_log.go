package learning

import "time"

// LearningLog は週次学習ログエンティティ（全ユーザー共通）。
type LearningLog struct {
	ID         uint
	WeekStart  time.Time
	WeekEnd    time.Time
	TradeCount int
	WinRate    float64
	TotalPnl   float64
	Summary    string
	Lessons    string
	Strategy   string
}

// LearningLogRepository は週次学習ログ永続化のインターフェース。
type LearningLogRepository interface {
	FindAll() ([]*LearningLog, error)
	FindByWeekStart(weekStart time.Time) (*LearningLog, error)
	Save(l *LearningLog) error
}
