package app

import (
	"log/slog"

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

func (a *App) Start() {
	const op = "app.Start"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Debug("initializing server...")

	pg, err := postgres.NewPool(&a.cfg.Postgres)
	if err != nil {
		log.Error("failed to establish connection with postgres", sl.Err(err))
		return
	}

	a.pg = pg

	log.Debug("connection with postgres initialized successfully")

	rds, err := rds.New(&a.cfg.Redis)
	if err != nil {
		log.Error("failed to establish connection with redis", sl.Err(err))
		return
	}

	a.redis = rds

	log.Debug("connection with redis initialized successfully")
}

func (a *App) Stop() {
	const op = "app.Stop"

	log := a.log.With(
		slog.String("op", op),
	)

	err := a.redis.Close()
	log.Error("%s: failed to close connection with redis: %w", op, err)

	a.pg.Close()
}
