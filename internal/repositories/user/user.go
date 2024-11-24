package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	http_lib "e-commerce-users/internal/lib/http"
	"e-commerce-users/internal/models"
	"e-commerce-users/internal/repositories"
	"e-commerce-users/pkg/logger/sl"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func New(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		db: pool,
	}
}

func (ur *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "repositories.auth.GetByEmail"

	log := http_lib.GetCtxLogger(ctx)
	log = log.With(slog.String("op", op))

	row := ur.db.QueryRow(ctx, `
	SELECT id, name, surname, birthdate, role, email, pass_hash, version, created_at
	FROM users
	WHERE email = $1`, email)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.Birthdate,
		&user.Role,
		&user.Email,
		&user.PassHash,
		&user.Version,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Info("user not found", slog.String("email", email))
			return nil, fmt.Errorf("%s: %w", op, repositories.ErrNotFound)
		}

		log.Error("failed to get user by email", slog.String("email", email), sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, err
}

func (ur *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	const op = "repositories.auth.GetByEmail"

	log := http_lib.GetCtxLogger(ctx)
	log = log.With(slog.String("op", op))

	row := ur.db.QueryRow(ctx, `
	SELECT id, name, surname, birthdate, role, email, pass_hash, version, created_at
	FROM users
	WHERE id = $1`, id)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.Birthdate,
		&user.Role,
		&user.Email,
		&user.PassHash,
		&user.Version,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Info("user not found", slog.String("id", id))
			return nil, fmt.Errorf("%s: %w", op, repositories.ErrNotFound)
		}

		log.Error("failed to get user by id", slog.String("id", id), sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, err
}

func (ur *UserRepo) CreateUser(ctx context.Context, name, surname, birthdate, email string, passHash []byte) error {
	const op = "repositories.auth.CreateUser"

	log := http_lib.GetCtxLogger(ctx)
	log = log.With(slog.String("op", op))

	_, err := ur.db.Exec(ctx, `
	INSERT INTO users (name, surname, birthdate, email, pass_hash)
	VALUES ($1, $2, $3, $4, $5)
	`, name, surname, birthdate, email, passHash)

	if err != nil {
		log.Error("failed to create user", slog.String("email", email), sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
