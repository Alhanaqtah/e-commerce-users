package services

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrExists             = errors.New("exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCode               = errors.New("invalid code")
	ErrNoActionRequired   = errors.New("no action required")
)

var (
	ErrTokenInvalid        = errors.New("token invalid")
	ErrTokenBlacklisted    = errors.New("token blacklisted")
	ErrTokenExpired        = errors.New("token expired")
	ErrUnexpectedTokenType = errors.New("unexpected token type")
)
