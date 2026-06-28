package domain

import "errors"

// ドメイン共通エラー。controller層でHTTPステータスにマッピングする。
var (
	ErrNotFound        = errors.New("リソースが見つかりません")
	ErrUnauthorized    = errors.New("メールアドレスまたはパスワードが正しくありません")
	ErrForbidden       = errors.New("権限がありません")
	ErrInvalidInput    = errors.New("入力値が不正です")
	ErrAlreadyExists   = errors.New("すでに存在しています")
	ErrInvalidCode     = errors.New("無効な招待コードです")
	ErrExpiredCode     = errors.New("招待コードの有効期限が切れています")
	ErrUsedCode        = errors.New("この招待コードはすでに使用されています")
	ErrAccountDisabled = errors.New("このアカウントは無効です。管理者にお問い合わせください")
)

// MessageError はドメインエラーに任意のメッセージを付与しつつ、
// errors.Is で種別判定できるようにするラッパー。
type MessageError struct {
	base error
	msg  string
}

func (e *MessageError) Error() string { return e.msg }
func (e *MessageError) Unwrap() error { return e.base }

// NewMessageError は種別 base と表示メッセージ msg を持つエラーを生成する。
func NewMessageError(base error, msg string) error {
	return &MessageError{base: base, msg: msg}
}
