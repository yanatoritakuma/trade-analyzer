package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword は平文パスワードをbcryptでハッシュ化する。
func HashPassword(raw string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// ComparePassword はbcryptハッシュと平文を照合する。一致すれば nil を返す。
func ComparePassword(hash, raw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw))
}
