package apphttp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"e-commerce-users/internal/config"
	auth_http "e-commerce-users/internal/delivery/http/auth"
	auth_service "e-commerce-users/internal/services/auth"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type App struct {
	server *http.Server
	log    *slog.Logger
}

func New(
	authSrvc *auth_service.Service,
	log *slog.Logger,
	cfg *config.Config,
) *App {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		return
	})

	r.Route("/api/v1", func(r chi.Router) {
		authHTTPCtrl := auth_http.New(
			&auth_http.Config{
				AuthService: authSrvc,
				Log:         log,
				TknsCfg:     &cfg.Tokens,
			},
		)
		r.Mount("/auth", authHTTPCtrl.Register())
	})

	srv := &http.Server{
		Addr:        cfg.HTTPServer.Host + ":" + cfg.HTTPServer.Port,
		Handler:     r,
		IdleTimeout: cfg.HTTPServer.IdleTimeout,
	}

	return &App{
		server: srv,
		log:    log,
	}
}

func (a *App) Start() error {
	const op = "apphttp.Start"

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("HTTP server error", slog.String("error", err.Error()))
		}
	}()

	return nil
}

func (a *App) Stop() error {
	const op = "apphttp.Stop"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
