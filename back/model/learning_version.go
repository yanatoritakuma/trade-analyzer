package model

import "gorm.io/gorm"

// LearningVersion は学習CSVのバージョン管理（learning_versions）。全ユーザー共通。
type LearningVersion struct {
	gorm.Model
	Version   int    `gorm:"not null"`
	S3Path    string `gorm:"not null"`
	WeekRange string
	CharCount int
}
