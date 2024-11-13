package app

import (
	"log/slog"
	"os"

	"github.com/Alhanaqtah/auth/internal/config"
	"github.com/Alhanaqtah/auth/pkg/logger/sl"
	"github.com/Alhanaqtah/auth/pkg/postgres"
	rds "github.com/Alhanaqtah/auth/pkg/redis"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	cfg   *config.Config
	log   *slog.Logger
	pg    *pgxpool.Pool
	redis *redis.Client
}

func New(cfg *config.Config, log *slog.Logger) *App {
	return &App{
		cfg: cfg,
		log: log,
	}
}

// Start initializes all necessary connections
func (a *App) Start() {
	const op = "app.Start"

	log := a.log.With(slog.String("op", op))

	log.Debug("initializing server...")

	// Initialize Postgres
	if err := a.initPostgres(); err != nil {
		log.Error("failed to establish connection with postgres", sl.Err(err))
		os.Exit(1)
	}
	log.Debug("connection with postgres initialized successfully")

	// Initialize Redis
	if err := a.initRedis(); err != nil {
		log.Error("failed to establish connection with redis", sl.Err(err))
		os.Exit(1)
	}
	log.Debug("connection with redis initialized successfully")
}

// Stop gracefully closes all connections
func (a *App) Stop() {
	const op = "app.Stop"

	log := a.log.With(slog.String("op", op))

	if err := a.redis.Close(); err != nil {
		log.Error("failed to close connection with redis", sl.Err(err))
	}

	a.pg.Close()
}

// initPostgres initializes the Postgres connection
func (a *App) initPostgres() error {
	pg, err := postgres.NewPool(&a.cfg.Postgres)
	if err != nil {
		return err
	}
	a.pg = pg
	return nil
}

// initRedis initializes the Redis connection
func (a *App) initRedis() error {
	rds, err := rds.New(&a.cfg.Redis)
	if err != nil {
		return err
	}
	a.redis = rds
	return nil
}
