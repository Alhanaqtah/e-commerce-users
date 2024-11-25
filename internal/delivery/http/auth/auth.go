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
	Logout(ctx context.Context, accessToken, refreshToken string) error
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
	r.Post("/logout", c.logout)

	return r
}

func (c *Controller) signUp(w http.ResponseWriter, r *http.Request) {
	const op = "http.auth.signUp"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var creds signUpCredentials
	if err := render.DecodeJSON(r.Body, &creds); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http_lib.ErrUnprocessableEntity(w, r)
		return
	}

	defer r.Body.Close()

	if err := c.valdtr.Struct(creds); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http_lib.ErrInvalid(w, r, err)
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
			http_lib.ErrConflict(w, r, "user already exists")
			return
		}

		http_lib.ErrInternal(w, r)
		return
	}

	log.Info("user signed up succesfully", slog.String("email", creds.Email))

	render.Status(r, http.StatusCreated)
	render.Render(w, r, http_lib.RespOk("User signed up succesfully"))
}

func (c *Controller) signIn(w http.ResponseWriter, r *http.Request) {
	const op = "http.auth.signUp"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var creds signInCredentials
	if err := render.DecodeJSON(r.Body, &creds); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http_lib.ErrUnprocessableEntity(w, r)
		return
	}

	defer r.Body.Close()

	if err := c.valdtr.Struct(creds); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http_lib.ErrInvalid(w, r, err)
		return
	}

	assessToken, refreshToken, err := c.as.SignIn(r.Context(), creds.Email, creds.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			http_lib.ErrUnauthorized(w, r, "Invalid credentials")
			return
		}

		http_lib.ErrInternal(w, r)
		return
	}

	log.Info("user logged in successfully", slog.String("email", creds.Email))

	render.Status(r, http.StatusOK)
	render.Render(w, r, tokens{
		AccessToken:  assessToken,
		RefreshToken: refreshToken,
	})
}

func (c *Controller) refresh(w http.ResponseWriter, r *http.Request) {
	const op = "http.auth.refresh"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var rfrshReq refreshRequest
	if err := render.DecodeJSON(r.Body, &rfrshReq); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http_lib.ErrUnprocessableEntity(w, r)
		return
	}

	defer r.Body.Close()

	if err := c.valdtr.Struct(rfrshReq); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http_lib.ErrInvalid(w, r, err)
		return
	}

	accTkn, rfrshTkn, err := c.as.Refresh(r.Context(), rfrshReq.RefreshToken)
	if err != nil {
		if errors.Is(err, services.ErrTokenBlacklisted) {
			http_lib.ErrUnauthorized(w, r, "Token revoked")
			return
		}
		if errors.Is(err, services.ErrTokenExpired) {
			http_lib.ErrUnauthorized(w, r, "Token expired")
			return
		}
		if errors.Is(err, services.ErrUnexpectedTokenType) {
			http_lib.ErrUnauthorized(w, r, "Unexpected token type: expected refresh token")
			return
		}

		http_lib.ErrInternal(w, r)
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, tokens{
		AccessToken:  accTkn,
		RefreshToken: rfrshTkn,
	})
}

func (c *Controller) logout(w http.ResponseWriter, r *http.Request) {
	const op = "http.auth.logout"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var tkns tokens
	if err := render.DecodeJSON(r.Body, &tkns); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http_lib.ErrUnprocessableEntity(w, r)
		return
	}

	if err := c.as.Logout(r.Context(), tkns.AccessToken, tkns.RefreshToken); err != nil {
		if errors.Is(err, services.ErrTokenInvalid) {
			http_lib.ErrUnauthorized(w, r, "Invalid token")
			return
		}

		if errors.Is(err, services.ErrTokenBlacklisted) {
			http_lib.ErrUnauthorized(w, r, "Token revoked")
			return
		}

		http_lib.ErrInternal(w, r)
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, http_lib.RespOk("User logged out succesfully"))
}

func (t tokens) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
