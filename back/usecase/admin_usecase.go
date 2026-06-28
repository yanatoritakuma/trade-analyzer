package usecase

import (
	"context"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/user"
)

// AdminUsecase は管理者によるユーザー管理のユースケース。
type AdminUsecase struct {
	userRepo user.UserRepository
}

func NewAdminUsecase(userRepo user.UserRepository) *AdminUsecase {
	return &AdminUsecase{userRepo: userRepo}
}

// ListUsers はユーザー一覧を返す。
func (u *AdminUsecase) ListUsers(ctx context.Context) ([]*user.User, error) {
	return u.userRepo.FindAll()
}

// SetUserActive はユーザーの停止/復活を切り替える（admin は対象外）。
func (u *AdminUsecase) SetUserActive(ctx context.Context, id uint, isActive bool) (*user.User, error) {
	usr, err := u.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if usr.IsAdmin() {
		return nil, domain.NewMessageError(domain.ErrForbidden, "管理者アカウントは操作できません")
	}
	usr.IsActive = isActive
	if err := u.userRepo.Update(usr); err != nil {
		return nil, err
	}
	return usr, nil
}

// DeleteUser はユーザーを物理削除する（admin は対象外）。
func (u *AdminUsecase) DeleteUser(ctx context.Context, id uint) error {
	usr, err := u.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if usr.IsAdmin() {
		return domain.NewMessageError(domain.ErrForbidden, "管理者アカウントは削除できません")
	}
	return u.userRepo.Delete(id)
}
