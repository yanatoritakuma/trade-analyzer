package model

import "gorm.io/gorm"

// User はユーザーアカウント（users）。
type User struct {
	gorm.Model
	Email        string `gorm:"uniqueIndex;not null"`
	Name         string `gorm:"not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"default:user;check:role IN ('admin','user')"`
	IsActive     bool   `gorm:"default:true"`
}
