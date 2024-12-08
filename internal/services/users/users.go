package users

import (
	"context"
	"fmt"
	"log/slog"

	http_lib "e-commerce-users/internal/lib/http"
	"e-commerce-users/internal/models"
	"e-commerce-users/pkg/logger/sl"
)

type UserRepo interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
}

type Service struct {
	usrRepo UserRepo
}

type Config struct {
	Repo UserRepo
}

func New(cfg *Config) *Service {
	return &Service{
		usrRepo: cfg.Repo,
	}
}

func (s *Service) GetProfile(ctx context.Context, id string) (*models.User, error) {
	const op = "services.users.GetProfile"

	log := http_lib.GetCtxLogger(ctx)
	log = log.With(slog.String("op", op))

	user, err := s.usrRepo.GetByID(ctx, id)
	if err != nil {
		log.Error("failed to get user by id", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
