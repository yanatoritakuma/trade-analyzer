package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/invitation"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/user"
	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

// --- モック ---

type mockUserRepo struct {
	user.UserRepository
	findByEmail func(email string) (*user.User, error)
	saved       *user.User
}

func (m *mockUserRepo) FindByEmail(email string) (*user.User, error) { return m.findByEmail(email) }
func (m *mockUserRepo) Save(u *user.User) error                      { u.ID = 1; m.saved = u; return nil }

type mockInviteRepo struct {
	invitation.InvitationCodeRepository
	findByCode func(code string) (*invitation.InvitationCode, error)
	updated    *invitation.InvitationCode
}

func (m *mockInviteRepo) FindByCode(code string) (*invitation.InvitationCode, error) {
	return m.findByCode(code)
}
func (m *mockInviteRepo) Update(i *invitation.InvitationCode) error { m.updated = i; return nil }

type mockUoW struct {
	repos *usecase.Repositories
}

func (m *mockUoW) Do(ctx context.Context, fn func(*usecase.Repositories) error) error {
	return fn(m.repos)
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	hash, _ := utils.HashPassword("password123")
	repo := &mockUserRepo{findByEmail: func(string) (*user.User, error) {
		return &user.User{ID: 1, Email: "a@b.com", Role: user.RoleAdmin, IsActive: true, PasswordHash: hash}, nil
	}}
	uc := usecase.NewUserUsecase(nil, repo)
	usr, at, rt, err := uc.Login(context.Background(), "a@b.com", "password123")
	if err != nil {
		t.Fatalf("想定外エラー: %v", err)
	}
	if usr == nil || at == "" || rt == "" {
		t.Fatal("トークンまたはユーザーが空")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := utils.HashPassword("password123")
	repo := &mockUserRepo{findByEmail: func(string) (*user.User, error) {
		return &user.User{ID: 1, IsActive: true, PasswordHash: hash}, nil
	}}
	uc := usecase.NewUserUsecase(nil, repo)
	if _, _, _, err := uc.Login(context.Background(), "a@b.com", "wrong"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Unauthorized を期待: %v", err)
	}
}

func TestLogin_NotFound(t *testing.T) {
	repo := &mockUserRepo{findByEmail: func(string) (*user.User, error) {
		return nil, domain.ErrNotFound
	}}
	uc := usecase.NewUserUsecase(nil, repo)
	if _, _, _, err := uc.Login(context.Background(), "x@y.com", "password123"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Unauthorized を期待: %v", err)
	}
}

func TestLogin_Disabled(t *testing.T) {
	hash, _ := utils.HashPassword("password123")
	repo := &mockUserRepo{findByEmail: func(string) (*user.User, error) {
		return &user.User{ID: 1, IsActive: false, PasswordHash: hash}, nil
	}}
	uc := usecase.NewUserUsecase(nil, repo)
	if _, _, _, err := uc.Login(context.Background(), "a@b.com", "password123"); !errors.Is(err, domain.ErrAccountDisabled) {
		t.Fatalf("AccountDisabled を期待: %v", err)
	}
}

// --- Register ---

func newRegisterUsecase(invRepo *mockInviteRepo, usrRepo *mockUserRepo) *usecase.UserUsecase {
	uow := &mockUoW{repos: &usecase.Repositories{User: usrRepo, InvitationCode: invRepo}}
	return usecase.NewUserUsecase(uow, usrRepo)
}

func TestRegister_InvalidCode(t *testing.T) {
	invRepo := &mockInviteRepo{findByCode: func(string) (*invitation.InvitationCode, error) {
		return nil, domain.ErrInvalidCode
	}}
	usrRepo := &mockUserRepo{findByEmail: func(string) (*user.User, error) { return nil, domain.ErrNotFound }}
	uc := newRegisterUsecase(invRepo, usrRepo)
	if _, err := uc.Register(context.Background(), "BAD", "名前", "a@b.com", "password123"); !errors.Is(err, domain.ErrInvalidCode) {
		t.Fatalf("ErrInvalidCode を期待: %v", err)
	}
}

func TestRegister_Success(t *testing.T) {
	invRepo := &mockInviteRepo{findByCode: func(string) (*invitation.InvitationCode, error) {
		return &invitation.InvitationCode{ID: 1, IsActive: true, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}}
	usrRepo := &mockUserRepo{findByEmail: func(string) (*user.User, error) { return nil, domain.ErrNotFound }}
	uc := newRegisterUsecase(invRepo, usrRepo)
	if _, err := uc.Register(context.Background(), "TRADE-XXXX-YYYY", "名前", "a@b.com", "password123"); err != nil {
		t.Fatalf("想定外エラー: %v", err)
	}
	if usrRepo.saved == nil {
		t.Fatal("User.Save が呼ばれていない")
	}
	if invRepo.updated == nil || invRepo.updated.UsedBy == nil {
		t.Fatal("招待コードが使用済みに更新されていない")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	invRepo := &mockInviteRepo{findByCode: func(string) (*invitation.InvitationCode, error) {
		return &invitation.InvitationCode{ID: 1, IsActive: true, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}}
	usrRepo := &mockUserRepo{findByEmail: func(string) (*user.User, error) {
		return &user.User{ID: 9}, nil // 既存ユーザーあり
	}}
	uc := newRegisterUsecase(invRepo, usrRepo)
	if _, err := uc.Register(context.Background(), "TRADE-XXXX-YYYY", "名前", "a@b.com", "password123"); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("ErrAlreadyExists を期待: %v", err)
	}
}
