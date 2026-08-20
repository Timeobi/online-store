package service

import (
	"errors"
	"time"

	"github.com/Timeobi/go-ecommerce/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

// Claims — данные, которые мы "зашиваем" внутрь JWT-токена.
type Claims struct {
	UserID int        `json:"user_id"`
	Role   model.Role `json:"role"`
	jwt.RegisteredClaims
}

// generateToken создаёт подписанный JWT-токен для пользователя.
func generateToken(user *model.User, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken проверяет подпись токена и возвращает распакованные claims.
func ParseToken(tokenString, secret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
