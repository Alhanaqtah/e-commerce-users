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
	db *pgxpool.Pool
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
	SELECT u.id, u.name, u.surname, u.birthdate, u.role, u.is_active, lc.email, lc.pass_hash, version, u.created_at
	FROM
		users u
	JOIN
		local_credentials lc
	on
		u.id = lc.user_id
	WHERE lc.email = $1`, email)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.Birthdate,
		&user.Role,
		&user.IsActive,
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
	SELECT u.id, u.name, u.surname, u.birthdate, u.role, u.is_active, lc.email, lc.pass_hash, version, u.created_at
	FROM
		users u
	JOIN
		local_credentials lc
	on
		u.id = lc.user_id
	WHERE u.id = $1`, id)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.Birthdate,
		&user.Role,
		&user.IsActive,
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

	tx, err := ur.db.Begin(ctx)
	if err != nil {
		log.Error("failed to begin transaction", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
	INSERT INTO users (name, surname, birthdate)
	VALUES ($1, $2, $3) RETURNING id
	`, name, surname, birthdate)

	var userID string
	err = row.Scan(&userID)
	if err != nil {
		log.Error("failed to create user: inserting into users table", slog.String("email", email), sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(ctx, `
	INSERT INTO local_credentials (user_id, email, pass_hash)
	VALUES ($1, $2, $3)
	`, userID, email, passHash)
	if err != nil {
		log.Error("failed to create user: inserting into users table", slog.String("email", email), sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("failed to commit transaction", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (ur *UserRepo) ActivateUser(ctx context.Context, email string) error {
	const op = "repositories.auth.ActivateUser"

	log := http_lib.GetCtxLogger(ctx)
	log = log.With(slog.String("op", op))

	var id string
	row := ur.db.QueryRow(ctx, `SELECT user_id FROM local_credentials WHERE email = $1`, email)

	if err := row.Scan(&id); err != nil {
		log.Error("failed to scan query result: get user id by email", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err := ur.db.Exec(ctx, `UPDATE users SET is_active = true WHERE id = $1`, id)
	if err != nil {
		log.Error("failed to activate user", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
