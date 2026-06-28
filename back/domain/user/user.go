package user

import "time"

// Role はユーザーロール値オブジェクト。
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleUser
}

// User はユーザーエンティティ。
type User struct {
	ID           uint
	Email        string
	Name         string
	PasswordHash string
	Role         Role
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser は新規ユーザーを生成する（role は常に user）。
// passwordHash はusecaseでハッシュ化済みの値を渡す。
func NewUser(email, name, passwordHash string) (*User, error) {
	if _, err := NewEmail(email); err != nil {
		return nil, err
	}
	if _, err := NewName(name); err != nil {
		return nil, err
	}
	return &User{
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		Role:         RoleUser,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// IsAdmin は管理者かどうかを返す。
func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }

// CanLogin はログイン可能か（有効アカウントか）を返す。
func (u *User) CanLogin() bool { return u.IsActive }

// ChangeName は名前を変更する（VO検証）。
func (u *User) ChangeName(name string) error {
	vo, err := NewName(name)
	if err != nil {
		return err
	}
	u.Name = vo.Value()
	return nil
}
