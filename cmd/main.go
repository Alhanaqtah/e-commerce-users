package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"e-commerce-users/internal/app"
	"e-commerce-users/internal/config"
	"e-commerce-users/pkg/logger"
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
