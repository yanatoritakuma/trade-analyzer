package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/invitation"
)

const inviteCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// InvitationUsecase は招待コード管理のユースケース。
type InvitationUsecase struct {
	invitationRepo invitation.InvitationCodeRepository
}

func NewInvitationUsecase(invitationRepo invitation.InvitationCodeRepository) *InvitationUsecase {
	return &InvitationUsecase{invitationRepo: invitationRepo}
}

// List は招待コード一覧を返す。
func (u *InvitationUsecase) List(ctx context.Context) ([]*invitation.InvitationCode, error) {
	return u.invitationRepo.FindAll()
}

// CountValid は有効な招待コード数を返す。
func (u *InvitationUsecase) CountValid(ctx context.Context) (int64, error) {
	return u.invitationRepo.CountValid()
}

// Create は有効期限を指定して招待コードを発行する（TRADE-XXXX-XXXX）。
func (u *InvitationUsecase) Create(ctx context.Context, expiresDays int, createdBy uint) (*invitation.InvitationCode, error) {
	switch expiresDays {
	case 3, 7, 14, 30:
	default:
		return nil, domain.NewMessageError(domain.ErrInvalidInput, "有効期限が不正です")
	}

	var code string
	for attempt := 0; attempt < 10; attempt++ {
		candidate, err := generateInviteCode()
		if err != nil {
			return nil, err
		}
		if _, err := u.invitationRepo.FindByCode(candidate); err != nil {
			if err == domain.ErrInvalidCode || err == domain.ErrNotFound {
				code = candidate
				break
			}
			return nil, err
		}
	}
	if code == "" {
		return nil, fmt.Errorf("招待コードの生成に失敗しました")
	}

	inv := &invitation.InvitationCode{
		Code:      code,
		CreatedBy: &createdBy,
		ExpiresAt: time.Now().Add(time.Duration(expiresDays) * 24 * time.Hour),
		IsActive:  true,
	}
	if err := u.invitationRepo.Save(inv); err != nil {
		return nil, err
	}
	return inv, nil
}

// Disable は招待コードを無効化する。
func (u *InvitationUsecase) Disable(ctx context.Context, id uint) error {
	inv, err := u.invitationRepo.FindByID(id)
	if err != nil {
		return err
	}
	inv.IsActive = false
	return u.invitationRepo.Update(inv)
}

func generateInviteCode() (string, error) {
	block := func() (string, error) {
		b := make([]byte, 4)
		for i := range b {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCharset))))
			if err != nil {
				return "", err
			}
			b[i] = inviteCharset[n.Int64()]
		}
		return string(b), nil
	}
	first, err := block()
	if err != nil {
		return "", err
	}
	second, err := block()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("TRADE-%s-%s", first, second), nil
}
