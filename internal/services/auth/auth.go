package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"e-commerce-users/internal/models"
	"e-commerce-users/internal/repositories"
	"e-commerce-users/internal/services"

	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, name, surname, birthdate, email string, passHash []byte) error
}

type Service struct {
	usrRepo UserRepo
	log     *slog.Logger
}

type Config struct {
	Repo UserRepo
	Log  *slog.Logger
}

func New(cfg *Config) *Service {
	return &Service{
		usrRepo: cfg.Repo,
		log:     cfg.Log,
	}
}

func (s *Service) SignUp(ctx context.Context, name, surname, birthdate, email, password string) error {
	const op = "services.auth.SignUp"

	log := s.log.With(slog.String("op", op))

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
