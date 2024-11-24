package app

import (
	"log/slog"
	"os"

	apphttp "e-commerce-users/internal/app/http"
	"e-commerce-users/internal/config"
	user_repo "e-commerce-users/internal/repositories/user"
	auth_service "e-commerce-users/internal/services/auth"
	"e-commerce-users/pkg/logger/sl"
	"e-commerce-users/pkg/postgres"
	rds "e-commerce-users/pkg/redis"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	cfg     *config.Config
	log     *slog.Logger
	strg    *pgxpool.Pool
	cache   *redis.Client
	httpSrv *apphttp.App
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

	// Initialize Postgres
	if err := a.initStorage(); err != nil {
		log.Error("failed to establish connection with storage", sl.Err(err))
		os.Exit(1)
	}

	log.Debug("connection with storage initialized successfully")

	// Initialize Redis
	if err := a.initCache(); err != nil {
		log.Error("failed to establish connection with cache", sl.Err(err))
		os.Exit(1)
	}

	log.Debug("connection with cache initialized successfully")

	authSrvc := auth_service.New(
		&auth_service.Config{
			Repo:    user_repo.New(a.strg),
			TknsCfg: &a.cfg.Tokens,
		},
	)

	httpServer := apphttp.New(
		authSrvc,
		log,
		a.cfg,
	)

	if err := httpServer.Start(); err != nil {
		log.Error("failed to start http server", sl.Err(err))
		os.Exit(1)
	}

	a.httpSrv = httpServer
}

// Stop gracefully closes all connections
func (a *App) Stop() {
	const op = "app.Stop"

	if err := a.httpSrv.Stop(); err != nil {
		a.log.Error("failed to stop server gracefully", sl.Err(err))
	}

	log := a.log.With(slog.String("op", op))

	if a.strg != nil {
		a.strg.Close()
	}

	if a.cache != nil {
		if err := a.cache.Close(); err != nil {
			log.Error("failed to close connection with cache", sl.Err(err))
		}
	}
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
