package user

// UserRepository はユーザー永続化のインターフェース。
type UserRepository interface {
	FindByID(id uint) (*User, error)
	FindByEmail(email string) (*User, error)
	FindFirstAdmin() (*User, error)
	FindAll() ([]*User, error)
	Save(u *User) error
	Update(u *User) error
	Delete(id uint) error
}
