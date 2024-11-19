package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"e-commerce-users/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(cfg *config.Postgres) (*pgxpool.Pool, error) {
	const op = "postgres.NewPool"

	pool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&pool_max_conns=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.MaxConns,
	))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create new pool: %w", op, err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to ping: %w", op, err)
	}

	if err = migrationUp(cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pool, nil
}

func migrationUp(cfg *config.Postgres) error {
	const op = "postgres.migrationsUp"

	sqlDB, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	))
	if err != nil {
		return fmt.Errorf("%s: failed to open sql DB: %w", op, err)
	}
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("%s: failed to create migration driver: %w", op, err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to create migration instance: %w", op, err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("%s: migration failed: %w", op, err)
	}

	return nil
}
