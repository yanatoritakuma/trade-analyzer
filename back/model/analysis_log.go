package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AnalysisLog はClaude APIの分析結果ログ（analysis_logs）。全ユーザー共通・user_idなし。
type AnalysisLog struct {
	gorm.Model
	Ticker     string `gorm:"not null;index"`
	Action     string
	Confidence float64        `gorm:"type:numeric(4,3)"`
	Analysis   datatypes.JSON // JSONBカラム
}
