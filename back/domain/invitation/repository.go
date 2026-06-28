package invitation

// InvitationCodeRepository は招待コード永続化のインターフェース。
type InvitationCodeRepository interface {
	FindByCode(code string) (*InvitationCode, error)
	FindByID(id uint) (*InvitationCode, error)
	FindAll() ([]*InvitationCode, error)
	CountValid() (int64, error)
	Save(i *InvitationCode) error
	Update(i *InvitationCode) error
}
