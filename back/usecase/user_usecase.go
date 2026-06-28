package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/user"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// UserUsecase は認証・ユーザー管理のユースケース。
type UserUsecase struct {
	uow      UnitOfWork
	userRepo user.UserRepository
}

func NewUserUsecase(uow UnitOfWork, userRepo user.UserRepository) *UserUsecase {
	return &UserUsecase{uow: uow, userRepo: userRepo}
}

// Login はメール＋パスワードで認証し、アクセス/リフレッシュトークンを返す。
func (u *UserUsecase) Login(ctx context.Context, email, password string) (*user.User, string, string, error) {
	usr, err := u.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, "", "", domain.ErrUnauthorized
		}
		return nil, "", "", fmt.Errorf("Login: FindByEmail failed: %w", err)
	}
	if !usr.CanLogin() {
		return nil, "", "", domain.ErrAccountDisabled
	}
	if err := utils.ComparePassword(usr.PasswordHash, password); err != nil {
		return nil, "", "", domain.ErrUnauthorized
	}
	accessToken, err := utils.GenerateAccessToken(usr.ID, string(usr.Role))
	if err != nil {
		return nil, "", "", fmt.Errorf("Login: GenerateAccessToken failed: %w", err)
	}
	refreshToken, err := utils.GenerateRefreshToken(usr.ID, string(usr.Role))
	if err != nil {
		return nil, "", "", fmt.Errorf("Login: GenerateRefreshToken failed: %w", err)
	}
	return usr, accessToken, refreshToken, nil
}

// Register は招待コードを検証してユーザーを作成する（UoWで一括・自動ログインなし）。
// 作成したユーザーを返す（トークンは発行しない）。
func (u *UserUsecase) Register(ctx context.Context, code, name, email, password string) (*user.User, error) {
	hashed, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("Register: HashPassword failed: %w", err)
	}
	var created *user.User
	err = u.uow.Do(ctx, func(repos *Repositories) error {
		inv, err := repos.InvitationCode.FindByCode(code)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidCode) || errors.Is(err, domain.ErrNotFound) {
				return domain.ErrInvalidCode
			}
			return err
		}
		switch {
		case !inv.IsActive:
			return domain.ErrInvalidCode
		case inv.IsUsed():
			return domain.ErrUsedCode
		case inv.IsExpired():
			return domain.ErrExpiredCode
		}

		if _, err := repos.User.FindByEmail(email); err == nil {
			return domain.NewMessageError(domain.ErrAlreadyExists, "このメールアドレスはすでに使用されています")
		} else if !errors.Is(err, domain.ErrNotFound) {
			return err
		}

		newUser, err := user.NewUser(email, name, hashed)
		if err != nil {
			return domain.NewMessageError(domain.ErrInvalidInput, err.Error())
		}
		if err := repos.User.Save(newUser); err != nil {
			return err
		}
		inv.MarkAsUsed(newUser.ID)
		if err := repos.InvitationCode.Update(inv); err != nil {
			return err
		}
		created = newUser
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Refresh はリフレッシュトークンを検証して新しいトークン対を発行する（ローテーション）。
func (u *UserUsecase) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := utils.ParseToken(refreshToken)
	if err != nil || claims.Type != utils.RefreshToken {
		return "", "", domain.ErrUnauthorized
	}
	usr, err := u.userRepo.FindByID(claims.UserID)
	if err != nil {
		return "", "", domain.ErrUnauthorized
	}
	if !usr.CanLogin() {
		return "", "", domain.ErrAccountDisabled
	}
	access, err := utils.GenerateAccessToken(usr.ID, string(usr.Role))
	if err != nil {
		return "", "", fmt.Errorf("Refresh: GenerateAccessToken failed: %w", err)
	}
	refresh, err := utils.GenerateRefreshToken(usr.ID, string(usr.Role))
	if err != nil {
		return "", "", fmt.Errorf("Refresh: GenerateRefreshToken failed: %w", err)
	}
	return access, refresh, nil
}

// Me はユーザー情報を取得する。
func (u *UserUsecase) Me(ctx context.Context, userID uint) (*user.User, error) {
	return u.userRepo.FindByID(userID)
}

// UpdateProfile は名前を更新する。
func (u *UserUsecase) UpdateProfile(ctx context.Context, userID uint, name string) (*user.User, error) {
	usr, err := u.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if err := usr.ChangeName(name); err != nil {
		return nil, domain.NewMessageError(domain.ErrInvalidInput, err.Error())
	}
	if err := u.userRepo.Update(usr); err != nil {
		return nil, err
	}
	return usr, nil
}

// ChangePassword は現在のパスワードを照合して新パスワードに変更する。
func (u *UserUsecase) ChangePassword(ctx context.Context, userID uint, currentPw, newPw string) error {
	usr, err := u.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if err := utils.ComparePassword(usr.PasswordHash, currentPw); err != nil {
		return domain.NewMessageError(domain.ErrInvalidInput, "現在のパスワードが正しくありません")
	}
	if _, err := user.NewPassword(newPw); err != nil {
		return domain.NewMessageError(domain.ErrInvalidInput, err.Error())
	}
	hashed, err := utils.HashPassword(newPw)
	if err != nil {
		return fmt.Errorf("ChangePassword: HashPassword failed: %w", err)
	}
	usr.PasswordHash = hashed
	return u.userRepo.Update(usr)
}
