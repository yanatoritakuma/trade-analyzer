package usecase

import (
	"context"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/position"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/user"
)

// PositionUsecase は実運用保有株のユースケース（admin の user_id で操作）。
type PositionUsecase struct {
	positionRepo position.PositionRepository
	userRepo     user.UserRepository
}

func NewPositionUsecase(positionRepo position.PositionRepository, userRepo user.UserRepository) *PositionUsecase {
	return &PositionUsecase{positionRepo: positionRepo, userRepo: userRepo}
}

// adminID は共通ポートフォリオの admin の user_id を返す。
func (u *PositionUsecase) adminID() (uint, error) {
	admin, err := u.userRepo.FindFirstAdmin()
	if err != nil {
		return 0, err
	}
	return admin.ID, nil
}

// ListWithPrice は admin の保有株に現在値・含み益を付与して返す。
func (u *PositionUsecase) ListWithPrice(ctx context.Context) ([]*position.Position, error) {
	adminID, err := u.adminID()
	if err != nil {
		return nil, err
	}
	return u.positionRepo.FindByUserWithPrice(adminID)
}

// Create は保有株を登録する（4桁コードに .T 付与・adminのuser_id）。
func (u *PositionUsecase) Create(ctx context.Context, code string, avgPrice float64, quantity int) (*position.Position, error) {
	adminID, err := u.adminID()
	if err != nil {
		return nil, err
	}
	p := &position.Position{
		UserID:   adminID,
		Ticker:   code + ".T",
		Quantity: quantity,
		AvgPrice: avgPrice,
	}
	if err := u.positionRepo.Save(p); err != nil {
		return nil, domain.NewMessageError(domain.ErrAlreadyExists, "すでに登録されている銘柄です")
	}
	return p, nil
}

// Update は保有株を更新する。
func (u *PositionUsecase) Update(ctx context.Context, id uint, code string, avgPrice float64, quantity int) (*position.Position, error) {
	adminID, err := u.adminID()
	if err != nil {
		return nil, err
	}
	p, err := u.positionRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p.UserID != adminID {
		return nil, domain.ErrForbidden
	}
	p.Ticker = code + ".T"
	p.AvgPrice = avgPrice
	p.Quantity = quantity
	if err := u.positionRepo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete は保有株を削除する。
func (u *PositionUsecase) Delete(ctx context.Context, id uint) error {
	adminID, err := u.adminID()
	if err != nil {
		return err
	}
	p, err := u.positionRepo.FindByID(id)
	if err != nil {
		return err
	}
	if p.UserID != adminID {
		return domain.ErrForbidden
	}
	return u.positionRepo.Delete(id, adminID)
}
