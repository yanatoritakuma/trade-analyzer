package model

import (
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AnalysisSetting は分析設定（analysis_settings）。全ユーザー共通・管理者が設定。
// is_active = TRUE のレコードが有効な設定（常に1件のみ）。
type AnalysisSetting struct {
	gorm.Model
	ThemeIDs   pq.Int64Array  `gorm:"type:integer[]"`
	Screening  datatypes.JSON // スクリーニング条件（JSONB）
	Style      string
	FreePrompt string
	IsActive   bool `gorm:"default:true"`
	CreatedBy  *uint
}
