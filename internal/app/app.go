package app

import (
	"log/slog"

	"github.com/Alhanaqtah/auth/internal/config"
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

	// DB conn

	// Cache conn
}
