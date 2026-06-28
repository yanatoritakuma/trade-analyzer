package invitation

import "time"

// InvitationCode は招待コードエンティティ。
type InvitationCode struct {
	ID        uint
	Code      string
	CreatedBy *uint
	UsedBy    *uint
	ExpiresAt time.Time
	UsedAt    *time.Time
	IsActive  bool
	CreatedAt time.Time
}

// IsValid は有効な招待コードか（未使用・有効・期限内）を返す。
func (i *InvitationCode) IsValid() bool {
	return i.IsActive && i.UsedBy == nil && i.ExpiresAt.After(time.Now())
}

// IsUsed は使用済みかを返す。
func (i *InvitationCode) IsUsed() bool { return i.UsedBy != nil }

// IsExpired は期限切れかを返す。
func (i *InvitationCode) IsExpired() bool { return i.ExpiresAt.Before(time.Now()) }

// MarkAsUsed は使用済みにマークする。
func (i *InvitationCode) MarkAsUsed(userID uint) {
	now := time.Now()
	i.UsedBy = &userID
	i.UsedAt = &now
}

// Status は表示用ステータスを返す（valid/used/expired/disabled）。
func (i *InvitationCode) Status() string {
	switch {
	case !i.IsActive:
		return "disabled"
	case i.IsUsed():
		return "used"
	case i.IsExpired():
		return "expired"
	default:
		return "valid"
	}
}
