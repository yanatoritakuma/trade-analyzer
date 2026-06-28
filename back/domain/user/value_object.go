package user

import (
	"errors"
	"regexp"
)

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email はメールアドレス値オブジェクト。
type Email struct{ value string }

func NewEmail(email string) (*Email, error) {
	if !emailRe.MatchString(email) {
		return nil, errors.New("メールアドレスの形式が正しくありません")
	}
	return &Email{value: email}, nil
}

func (e *Email) Value() string { return e.value }

// Name は表示名値オブジェクト（最大50文字）。
type Name struct{ value string }

func NewName(name string) (Name, error) {
	if name == "" {
		return Name{}, errors.New("名前を入力してください")
	}
	if len([]rune(name)) > 50 {
		return Name{}, errors.New("名前は50文字以内で入力してください")
	}
	return Name{value: name}, nil
}

func (n Name) Value() string { return n.value }

// Password は平文パスワード値オブジェクト（最小8文字）。ハッシュ化はusecaseで行う。
type Password struct{ value string }

func NewPassword(raw string) (Password, error) {
	if len(raw) < 8 {
		return Password{}, errors.New("パスワードは8文字以上で入力してください")
	}
	return Password{value: raw}, nil
}

func (p Password) Value() string { return p.value }
