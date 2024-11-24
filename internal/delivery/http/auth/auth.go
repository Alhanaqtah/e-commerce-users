package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"e-commerce-users/internal/config"
	http_lib "e-commerce-users/internal/lib/http"
	"e-commerce-users/internal/services"
	"e-commerce-users/pkg/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type AuthService interface {
	SignUp(ctx context.Context, name, surname, birthdate, email, password string) error
	SignIn(ctx context.Context, name, password string) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
}

type Controller struct {
	as     AuthService
	tCfg   *config.Tokens
	valdtr *validator.Validate
}

type Config struct {
	AuthService AuthService
	TknsCfg     *config.Tokens
}

type signUpCredentials struct {
	Name      string    `json:"name" validate:"required"`
	Surname   string    `json:"surname" validate:"required"`
	Birthdate time.Time `json:"birthdate" validate:"required"`
	Email     string    `json:"email" validate:"required,email"`
	Password  string    `json:"password" validate:"required"`
}

type signInCredentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func New(cfg *Config) *Controller {
	return &Controller{
		as:     cfg.AuthService,
		tCfg:   cfg.TknsCfg,
		valdtr: validator.New(),
	}
}

func (c *Controller) Register() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/sign-up", c.signUp)
	r.Post("/sign-in", c.signIn)
	r.Post("/refresh", c.refresh)

	return r
}

func (c *Controller) signUp(w http.ResponseWriter, r *http.Request) {
	const op = "controllers.auth.signUp"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var creds signUpCredentials
	defer r.Body.Close()
	if err := render.DecodeJSON(r.Body, &creds); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http.Error(w, "unprocessable entity", http.StatusUnprocessableEntity)
		return
	}

	if err := c.valdtr.Struct(creds); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http.Error(w, "some fields are invalid", http.StatusBadRequest)
		return
	}

	if err := c.as.SignUp(r.Context(),
		creds.Name,
		creds.Surname,
		creds.Birthdate.Format("2006-01-02"),
		creds.Email,
		creds.Password,
	); err != nil {
		if errors.Is(err, services.ErrExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}

		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusCreated)
}

func (c *Controller) signIn(w http.ResponseWriter, r *http.Request) {
	const op = "controllers.auth.signUp"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var creds signInCredentials
	defer r.Body.Close()
	if err := render.DecodeJSON(r.Body, &creds); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http.Error(w, "unprocessable entity", http.StatusUnprocessableEntity)
		return
	}

	if err := c.valdtr.Struct(creds); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	assessToken, refreshToken, err := c.as.SignIn(r.Context(), creds.Email, creds.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, tokens{
		AccessToken:  assessToken,
		RefreshToken: refreshToken,
	})
}

func (c *Controller) refresh(w http.ResponseWriter, r *http.Request) {
	const op = "controllers.auth.refresh"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var rfrshReq refreshRequest
	defer r.Body.Close()
	if err := render.DecodeJSON(r.Body, &rfrshReq); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http.Error(w, "unprocessable entity", http.StatusUnprocessableEntity)
		return
	}

	if err := c.valdtr.Struct(rfrshReq); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http.Error(w, "some fiels are invalid", http.StatusBadRequest)
		return
	}

	accTkn, rfrshTkn, err := c.as.Refresh(r.Context(), rfrshReq.RefreshToken)
	if err != nil {
		if errors.Is(err, services.ErrTokenBlacklisted) {
			http.Error(w, "token is blacklisted", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, services.ErrTokenExpired) {
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}

		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, tokens{
		AccessToken:  accTkn,
		RefreshToken: rfrshTkn,
	})
}

func (t tokens) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
