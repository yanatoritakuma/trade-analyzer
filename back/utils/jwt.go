package utils

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType はトークン種別（アクセス / リフレッシュ）。
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims はJWTに格納するクレーム。
type Claims struct {
	UserID uint      `json:"user_id"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev_secret_change_me"
	}
	return []byte(secret)
}

func accessExpire() time.Duration {
	h, err := strconv.Atoi(os.Getenv("JWT_ACCESS_EXPIRE_HOURS"))
	if err != nil || h <= 0 {
		h = 24
	}
	return time.Duration(h) * time.Hour
}

func refreshExpire() time.Duration {
	d, err := strconv.Atoi(os.Getenv("JWT_REFRESH_EXPIRE_DAYS"))
	if err != nil || d <= 0 {
		d = 30
	}
	return time.Duration(d) * 24 * time.Hour
}

func generateToken(userID uint, role string, t TokenType, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   t,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

// GenerateAccessToken はアクセストークン（既定24時間）を発行する。
func GenerateAccessToken(userID uint, role string) (string, error) {
	return generateToken(userID, role, AccessToken, accessExpire())
}

// GenerateRefreshToken はリフレッシュトークン（既定30日）を発行する。
func GenerateRefreshToken(userID uint, role string) (string, error) {
	return generateToken(userID, role, RefreshToken, refreshExpire())
}

// ParseToken はトークンを検証してクレームを返す。
func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("予期しない署名方式です")
		}
		return jwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("トークンが無効です")
	}
	return claims, nil
}
