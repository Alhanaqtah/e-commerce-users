package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/golang-jwt/jwt"
)

var (
	ErrInvalid       = errors.New("invalid")
	ErrClaimNotFound = errors.New("claim not found")
	ErrExpired       = errors.New("expired")
)

func NewAccessToken(
	id string,
	role string,
	version int,
	exp time.Time,
	secret string,
) (string, error) {
	const op = "lib.jwt.NewAccessToken"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     id,
		"role":    role,
		"version": version,
		"type":    "access",
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
	const op = "lib.jwt.NewRefreshToken"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     id,
		"version": version,
		"type":    "refresh",
		"exp":     exp.Unix(),
	})

	tkn, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tkn, nil
}

func FromCtx(ctx context.Context) (map[string]interface{}, error) {
	const op = "lib.jwt.FromCtx"

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return claims, nil
}

func FromString(token, secret string) (jwt.MapClaims, error) {
	const op = "lib.jwt.FromString"

	tkn, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok && ve.Errors == jwt.ValidationErrorExpired {
			return nil, fmt.Errorf("%s: %w", op, ErrExpired)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tkn.Claims.(jwt.MapClaims), nil
}

func GetClaim(claims jwt.MapClaims, claimName string) (string, error) {
	const op = "lib.jwt.GetClaim"

	claim, ok := claims[claimName]
	if !ok {
		return "", fmt.Errorf("%s: %w", op, ErrClaimNotFound)
	}

	switch v := claim.(type) {
	case string:
		return v, nil
	case float64:
		return time.Unix(int64(v), 0).Format(time.RFC3339), nil
	default:
		return "", fmt.Errorf("%s: unsupported claim type %T", op, v)
	}
}
