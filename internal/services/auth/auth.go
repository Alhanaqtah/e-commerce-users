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
	CreateUser(ctx context.Context, name, surname, birthdate, email string, passHash []byte) error
}

type Service struct {
	usrRepo UserRepo
	TknsCfg *config.Tokens
}

type Config struct {
	Repo    UserRepo
	TknsCfg *config.Tokens
}

func New(cfg *Config) *Service {
	return &Service{
		usrRepo: cfg.Repo,
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
			return "", "", fmt.Errorf("%s: %w", op, services.ErrNotFound)
		}

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Debug("invalid password")
			return "", "", fmt.Errorf("%s: %w", op, services.ErrInvalidCredentials)
		}

		log.Debug("failed to compare password hash", sl.Err(err))
		return "", "", err
	}

	accessTknExp := time.Now().Add(s.TknsCfg.AccessTTL)
	refreshTknExp := time.Now().Add(s.TknsCfg.RefreshTTL)

	accessToken, err := jwt_lib.NewAccessToken(
		user.ID,
		user.Role,
		user.Version,
		accessTknExp,
		s.TknsCfg.Secret,
	)
	if err != nil {
		log.Error("failed to generate access token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %s", op, err)
	}

	refreshToken, err := jwt_lib.NewRefreshToken(
		user.ID,
		user.Version,
		refreshTknExp,
		s.TknsCfg.Secret,
	)
	if err != nil {
		log.Error("failed to generate refresh token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %s", op, err)
	}

	return accessToken, refreshToken, nil
}
