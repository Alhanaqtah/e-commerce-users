package main

import (
	"log/slog"

	"github.com/Alhanaqtah/auth/internal/config"
	"github.com/Alhanaqtah/auth/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(cfg.ENV)

	log.Info("starting server", slog.String("port", cfg.HTTPServer.Port))
}
