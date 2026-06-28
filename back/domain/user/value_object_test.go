package user_test

import (
	"testing"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/user"
)

func TestNewEmail_Invalid(t *testing.T) {
	cases := []string{"not-an-email", "missing@", "@nodomain.com", ""}
	for _, c := range cases {
		if _, err := user.NewEmail(c); err == nil {
			t.Errorf("入力 %q はエラーになるべき", c)
		}
	}
}

func TestNewEmail_Valid(t *testing.T) {
	if _, err := user.NewEmail("user@example.com"); err != nil {
		t.Fatalf("有効なメールアドレスでエラー: %v", err)
	}
}

func TestNewPassword_TooShort(t *testing.T) {
	if _, err := user.NewPassword("short"); err == nil {
		t.Fatal("8文字未満はエラーになるべき")
	}
}

func TestNewPassword_OK(t *testing.T) {
	if _, err := user.NewPassword("password123"); err != nil {
		t.Fatalf("8文字以上でエラー: %v", err)
	}
}

func TestNewName_TooLong(t *testing.T) {
	long := make([]rune, 51)
	for i := range long {
		long[i] = 'あ'
	}
	if _, err := user.NewName(string(long)); err == nil {
		t.Fatal("50文字超はエラーになるべき")
	}
}
