package app

import (
	"log/slog"

	"github.com/Alhanaqtah/auth/internal/config"
	"github.com/Alhanaqtah/auth/pkg/logger/sl"
	"github.com/Alhanaqtah/auth/pkg/postgres"
)

type App struct {
	cfg *config.Config
	log *slog.Logger
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

	db, err := postgres.NewPool(&a.cfg.Postgres)
	if err != nil {
		log.Error("failed to establish connection with postgres", sl.Err(err))
		return
	}

	log.Debug("connection with postgres initialized successfully")

	_ = db

	// Cache conn
}
