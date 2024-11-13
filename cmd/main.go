package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Alhanaqtah/auth/internal/app"
	"github.com/Alhanaqtah/auth/internal/config"
	"github.com/Alhanaqtah/auth/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(cfg.ENV)
	log.Info("starting server", slog.String("port", cfg.HTTPServer.Port))

	app := app.New(cfg, log)

	go app.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	log.Info("shutting down server...")

	app.Stop()
	log.Info("server shutted down successfully")
}
