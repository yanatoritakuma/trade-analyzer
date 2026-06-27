package model

import (
	"time"

	"gorm.io/gorm"
)

// InvitationCode は招待コード（invitation_codes）。
type InvitationCode struct {
	gorm.Model
	Code      string `gorm:"uniqueIndex;not null"`
	CreatedBy *uint
	UsedBy    *uint
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	IsActive  bool `gorm:"default:true"`
}
