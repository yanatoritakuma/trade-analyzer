package model

import "gorm.io/gorm"

// AnalysisTheme は分析テーマ一覧（analysis_themes）。管理者がUI上で管理。
type AnalysisTheme struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	SortOrder   int  `gorm:"default:0;index"`
	IsActive    bool `gorm:"default:true"`
	CreatedBy   *uint
}
