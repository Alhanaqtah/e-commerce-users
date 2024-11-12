package postgres

import (
	"context"
	"fmt"

	"github.com/Alhanaqtah/auth/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(cfg *config.Postgres) (*pgxpool.Pool, error) {
	const op = "postgres.NewPool"

	pool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s?pool_max_conns=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.MaxConns,
	))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create new pool: %w", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to ping: %w", err)
	}

	return pool, nil
}
