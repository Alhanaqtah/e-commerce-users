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
	strg  *pgxpool.Pool
	cache *redis.Client
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
	if err := a.initStorage(); err != nil {
		log.Error("failed to establish connection with storage", sl.Err(err))
		return
	}

	log.Debug("connection with storage initialized successfully")

	// Initialize Redis
	if err := a.initCache(); err != nil {
		log.Error("failed to establish connection with cache", sl.Err(err))
		return
	}

	log.Debug("connection with cache initialized successfully")
}

// Stop gracefully closes all connections
func (a *App) Stop() {
	const op = "app.Stop"

	log := a.log.With(slog.String("op", op))

	if err := a.cache.Close(); err != nil {
		log.Error("failed to close connection with cache", sl.Err(err))
	}

	a.strg.Close()
}

// initStorage initializes connection with storage
func (a *App) initStorage() error {
	storage, err := postgres.NewPool(&a.cfg.Postgres)
	if err != nil {
		return err
	}

	a.strg = storage

	return nil
}

// initCache initializes connection with cache
func (a *App) initCache() error {
	cache, err := rds.New(&a.cfg.Redis)
	if err != nil {
		return err
	}

	a.cache = cache

	return nil
}
