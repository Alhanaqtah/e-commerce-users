package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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

func New(pool *pgxpool.Pool, log *slog.Logger) *UserRepo {
	return &UserRepo{
		db:  pool,
		log: log,
	}
}

func (ur *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "repositories.auth.GetByEmail"

	log := ur.log.With(slog.String("op", op))

	row := ur.db.QueryRow(ctx, `
	SELECT name, surname, birthdate, role, email, created_at
	FROM users
	WHERE email = $1`, email)

	var user models.User
	err := row.Scan(
		&user.Name,
		&user.Surname,
		&user.Birthdate,
		&user.Role,
		&user.Email,
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

func (ur *UserRepo) CreateUser(ctx context.Context, name, surname, birthdate, email string, passHash []byte) error {
	const op = "repositories.auth.CreateUser"

	log := ur.log.With(slog.String("op", op))

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
