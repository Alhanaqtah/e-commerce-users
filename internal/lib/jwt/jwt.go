package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func NewAccessToken(
	id string,
	role string,
	version int,
	exp time.Time,
	secret string,
) (string, error) {
	const op = "jwt.NewAccessToken"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     id,
		"role":    role,
		"version": version,
		"jti":     uuid.New().String(),
		"exp":     exp.Unix(),
	})

	tkn, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tkn, nil
}

func NewRefreshToken(
	id string,
	version int,
	exp time.Time,
	secret string,
) (string, error) {
	const op = "jwt.NewRefreshToken"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     id,
		"version": version,
		"jti":     uuid.New().String(),
		"exp":     exp.Unix(),
	})

	tkn, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tkn, nil
}
