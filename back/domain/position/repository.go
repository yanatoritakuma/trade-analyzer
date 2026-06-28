package position

// PositionRepository は実運用保有株永続化のインターフェース（ユーザーごと）。
// 全クエリで WHERE user_id = ? を必須とする。
type PositionRepository interface {
	FindByUser(userID uint) ([]*Position, error)
	FindByUserWithPrice(userID uint) ([]*Position, error)
	FindByID(id uint) (*Position, error)
	Save(p *Position) error
	Update(p *Position) error
	Delete(id uint, userID uint) error
}
