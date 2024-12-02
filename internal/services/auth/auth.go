package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"e-commerce-users/internal/config"
	http_lib "e-commerce-users/internal/lib/http"
	jwt_lib "e-commerce-users/internal/lib/jwt"
	"e-commerce-users/internal/models"
	"e-commerce-users/internal/repositories"
	"e-commerce-users/internal/services"
	"e-commerce-users/pkg/logger/sl"

	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	CreateUser(ctx context.Context, name, surname, birthdate, email string, passHash []byte) error
}

type Cache interface {
	IsBlacklisted(ctx context.Context, token string) (bool, error)
	AddToBlacklist(ctx context.Context, token string, ttl time.Duration) error
}

type Service struct {
	usrRepo UserRepo
	cache   Cache
	TknsCfg *config.Tokens
}

type Config struct {
	Repo    UserRepo
	Cache   Cache
	TknsCfg *config.Tokens
}

func New(cfg *Config) *Service {
	return &Service{
		usrRepo: cfg.Repo,
		cache:   cfg.Cache,
		TknsCfg: cfg.TknsCfg,
	}
}

func (s *Service) SignUp(ctx context.Context, name, surname, birthdate, email, password string) error {
	const op = "services.auth.SignUp"

	log := http_lib.GetCtxLogger(ctx)
	log = log.With(slog.String("op", op))

	user, err := s.usrRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return fmt.Errorf("%s: %w", op, err)
	}

	if user != nil {
		log.Debug("user already exists", slog.String("email", email))
		return fmt.Errorf("%s: %w", op, services.ErrExists)
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password")
		return err
	}

	if err := s.usrRepo.CreateUser(ctx,
		name,
		surname,
		birthdate,
		email,
		passHash,
	); err != nil {
		if errors.Is(err, repositories.ErrExists) {
			return fmt.Errorf("%s: %w", op, services.ErrExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) SignIn(ctx context.Context, email, password string) (string, string, error) {
	const op = "services.auth.SignIn"

	log := http_lib.GetCtxLogger(ctx)
	log = log.With(slog.String("op", op))

	user, err := s.usrRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			log.Warn("email not found", slog.String("email", email))
			return "", "", fmt.Errorf("%s: %w", op, services.ErrNotFound)
		}

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Warn("invalid password", slog.String("email", email))
			return "", "", fmt.Errorf("%s: %w", op, services.ErrInvalidCredentials)
		}

		log.Debug("failed to compare password hash", sl.Err(err))
		return "", "", err
	}

	accessToken, err := jwt_lib.NewAccessToken(
		user.ID,
		user.Role,
		user.Version,
		time.Now().Add(s.TknsCfg.AccessTTL),
		s.TknsCfg.Secret,
	)
	if err != nil {
		log.Error("failed to generate access token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %s", op, err)
	}

	refreshToken, err := jwt_lib.NewRefreshToken(
		user.ID,
		user.Version,
		time.Now().Add(s.TknsCfg.RefreshTTL),
		s.TknsCfg.Secret,
	)
	if err != nil {
		log.Error("failed to generate refresh token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %s", op, err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	const op = "services.auth.Refresh"

	log := http_lib.GetCtxLogger(ctx)
	log = log.With(slog.String("op", op))

	blacklisted, err := s.cache.IsBlacklisted(ctx, refreshToken)
	if err != nil {
		log.Error("failed to check if token blacklisted", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if blacklisted {
		log.Warn("token is blacklisted")
		return "", "", fmt.Errorf("%s: %w", op, services.ErrTokenBlacklisted)
	}

	claims, err := jwt_lib.FromString(refreshToken, s.TknsCfg.Secret)
	if err != nil {
		if errors.Is(err, jwt_lib.ErrExpired) {
			log.Error("token expired", sl.Err(err))
			return "", "", fmt.Errorf("%s: %w", op, services.ErrTokenExpired)
		}

		log.Error("failed to extract token claims", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	tknType, err := jwt_lib.GetClaim(claims, "type")
	if err != nil {
		log.Error("failed to get user ID from claims", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if tknType != "refresh" {
		log.Warn("unexpected token type: expected 'refresh'")
		return "", "", fmt.Errorf("%s: %w", op, services.ErrUnexpectedTokenType)
	}

	userID, err := jwt_lib.GetClaim(claims, "sub")
	if err != nil {
		log.Error("failed to get user ID from claims", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	user, err := s.usrRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error("failed to get user by id", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	accTkn, err := jwt_lib.NewAccessToken(
		user.ID,
		user.Role,
		user.Version,
		time.Now().Add(s.TknsCfg.AccessTTL),
		s.TknsCfg.Secret,
	)
	if err != nil {
		log.Error("failed to generate access token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	rfrshTkn, err := jwt_lib.NewRefreshToken(
		user.ID,
		user.Version,
		time.Now().Add(s.TknsCfg.RefreshTTL),
		s.TknsCfg.Secret,
	)
	if err != nil {
		log.Error("failed to generate refresh token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	expStr, err := jwt_lib.GetClaim(claims, "exp")
	if err != nil {
		log.Error("failed to get expiration time from claims", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	expTime, err := time.Parse(time.RFC3339, expStr)
	if err != nil {
		log.Error("failed to parse expiration time", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err := s.cache.AddToBlacklist(ctx, refreshToken, time.Until(expTime)); err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accTkn, rfrshTkn, nil
}

func (s *Service) Logout(ctx context.Context, accessToken, refreshToken string) error {
	const op = "services.auth.Logout"

	log := http_lib.GetCtxLogger(ctx)
	log = log.With(slog.String("op", op))

	// Check access token in blacklist
	blacklisted, err := s.cache.IsBlacklisted(ctx, accessToken)
	if err != nil {
		log.Error("failed to check if access token blacklisted", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if blacklisted {
		log.Warn("acces token blacklisted")
		return fmt.Errorf("%s: %w", op, services.ErrTokenBlacklisted)
	}

	// Check refresh token in blacklist
	blacklisted, err = s.cache.IsBlacklisted(ctx, refreshToken)
	if err != nil {
		log.Error("failed to check if refresh token blacklisted", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if blacklisted {
		log.Warn("refresh token blacklisted")
		return fmt.Errorf("%s: %w", op, services.ErrTokenBlacklisted)
	}

	accClaims, err := jwt_lib.FromString(accessToken, s.TknsCfg.Secret)
	if err != nil && !errors.Is(err, jwt_lib.ErrExpired) {
		if errors.Is(err, jwt_lib.ErrInvalid) {
			log.Error("token invalid", sl.Err(err))
			return fmt.Errorf("%s: %w", op, services.ErrTokenInvalid)
		}

		log.Error("failed to extract token claims", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	accExpStr, err := jwt_lib.GetClaim(accClaims, "exp")
	if err != nil {
		log.Error("failed to get expiration time from claims", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	accExpTime, err := time.Parse(time.RFC3339, accExpStr)
	if err != nil {
		log.Error("failed to parse expiration time", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	rfrshClaims, err := jwt_lib.FromString(refreshToken, s.TknsCfg.Secret)
	if err != nil {
		if errors.Is(err, jwt_lib.ErrExpired) {
			log.Warn("token expired", sl.Err(err))
			return fmt.Errorf("%s: %w", op, services.ErrTokenExpired)
		}
		if errors.Is(err, jwt_lib.ErrInvalid) {
			log.Error("token invalid", sl.Err(err))
			return fmt.Errorf("%s: %w", op, services.ErrTokenInvalid)
		}

		log.Error("failed to extract token claims", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	rfrshExpStr, err := jwt_lib.GetClaim(rfrshClaims, "exp")
	if err != nil {
		log.Error("failed to get expiration time from claims", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	rfrshExpTime, err := time.Parse(time.RFC3339, rfrshExpStr)
	if err != nil {
		log.Error("failed to parse expiration time", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.cache.AddToBlacklist(ctx, refreshToken, time.Until(rfrshExpTime)); err != nil {
		log.Error("failed to blacklist refresh token", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.cache.AddToBlacklist(ctx, accessToken, time.Until(accExpTime)); err != nil {
		log.Error("failed to blacklist access token", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
