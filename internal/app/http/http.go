package apphttp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"e-commerce-users/internal/config"
	auth_http "e-commerce-users/internal/delivery/http/auth"
	users_http "e-commerce-users/internal/delivery/http/users"
	http_lib "e-commerce-users/internal/lib/http"
	auth_service "e-commerce-users/internal/services/auth"
	users_service "e-commerce-users/internal/services/users"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type App struct {
	server *http.Server
	log    *slog.Logger
}

func New(
	authSrvc *auth_service.Service,
	usrSrvc *users_service.Service,
	log *slog.Logger,
	cfg *config.Config,
) *App {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(http_lib.TraceID)
	r.Use(http_lib.Logging(log))
	r.Use(middleware.Recoverer)

	r.Post("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api/v1", func(r chi.Router) {
		authCtrl := auth_http.New(
			&auth_http.Config{
				AuthService: authSrvc,
				TknsCfg:     &cfg.Tokens,
			},
		)
		r.Mount("/auth", authCtrl.Register())

		usersCtrl := users_http.New(
			&users_http.Config{
				UsrSrvc: usrSrvc,
				TknsCfg: cfg.Tokens,
			},
		)
		r.Mount("/users", usersCtrl.Register())
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
